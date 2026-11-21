package repository

import (
	"context"
	"errors"
	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Создаем кастомные ошибки, чтобы хендлер понимал, что именно пошло не так
var (
	ErrTripNotFound   = errors.New("поездка не найдена")
	ErrNotEnoughSeats = errors.New("недостаточно свободных мест")
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	// 1. Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Если что-то пойдет не так, транзакция откатится (Rollback)
	defer tx.Rollback(ctx)

	// 2. Блокируем строку поездки, чтобы другие запросы ждали
	var seatsAvailable int
	queryTrip := `SELECT seats_available FROM trips WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryTrip, booking.TripID).Scan(&seatsAvailable)
	if err != nil {
		return ErrTripNotFound // Если строка не найдена
	}

	// 3. Проверяем, хватает ли мест
	if seatsAvailable < booking.Seats {
		return ErrNotEnoughSeats
	}

	// 4. Уменьшаем количество свободных мест
	queryUpdate := `UPDATE trips SET seats_available = seats_available - $1 WHERE id = $2`
	_, err = tx.Exec(ctx, queryUpdate, booking.Seats, booking.TripID)
	if err != nil {
		return err
	}

	// 5. Создаем бронь
	queryInsert := `INSERT INTO bookings (id, trip_id, passenger_id, seats, status) 
	                VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, queryInsert, booking.ID, booking.TripID, booking.PassengerID, booking.Seats, booking.Status)
	if err != nil {
		return err
	}

	// 6. Фиксируем изменения (Commit)
	return tx.Commit(ctx)
}

// Получение броней конкретного пассажира
func (r *BookingRepository) GetByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]*models.Booking, error) {
	query := `SELECT id, trip_id, passenger_id, seats, status FROM bookings WHERE passenger_id = $1`

	rows, err := r.pool.Query(ctx, query, passengerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		var b models.Booking
		err := rows.Scan(&b.ID, &b.TripID, &b.PassengerID, &b.Seats, &b.Status)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

// Отмена брони (с возвратом мест)
func (r *BookingRepository) Cancel(ctx context.Context, bookingID uuid.UUID, passengerID uuid.UUID) error {
	// 1. Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 2. Ищем бронь и блокируем её (FOR UPDATE)
	var tripID uuid.UUID
	var seats int
	var status string
	queryBooking := `SELECT trip_id, seats, status FROM bookings WHERE id = $1 AND passenger_id = $2 FOR UPDATE`
	err = tx.QueryRow(ctx, queryBooking, bookingID, passengerID).Scan(&tripID, &seats, &status)
	if err != nil {
		return errors.New("бронь не найдена или не принадлежит вам")
	}

	if status == "cancelled" {
		return errors.New("бронь уже отменена")
	}

	// 3. Меняем статус брони
	queryUpdateBooking := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateBooking, bookingID)
	if err != nil {
		return err
	}

	// 4. Возвращаем места в поездку
	queryUpdateTrip := `UPDATE trips SET seats_available = seats_available + $1 WHERE id = $2`
	_, err = tx.Exec(ctx, queryUpdateTrip, seats, tripID)
	if err != nil {
		return err
	}

	// 5. Фиксируем
	return tx.Commit(ctx)
}
