package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"transfer-service/internal/models"
	"transfer-service/internal/payment"

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
	                VALUES ($1, $2, $3, $4, $5) RETURNING created_at`
	err = tx.QueryRow(ctx, queryInsert, booking.ID, booking.TripID, booking.PassengerID, booking.Seats, booking.Status).Scan(&booking.CreatedAt)
	if err != nil {
		return err
	}

	// 6. Фиксируем изменения (Commit)
	return tx.Commit(ctx)
}

// Получение броней конкретного пассажира
func (r *BookingRepository) GetByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]*models.Booking, error) {
	query := `SELECT b.id, b.trip_id, b.passenger_id, b.seats, b.status, b.created_at, b.stripe_payment_intent IS NOT NULL, b.status = 'pending' AND ` + bookingUnexpired + ` AND ` + bookingAmount + ` BETWEEN $3 AND $4 FROM bookings b JOIN trips t ON t.id = b.trip_id WHERE b.passenger_id = $1`

	rows, err := r.pool.Query(ctx, query, passengerID, int64(models.BookingHold.Seconds()), payment.MinimumAmount, payment.MaximumAmount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		var b models.Booking
		err := rows.Scan(&b.ID, &b.TripID, &b.PassengerID, &b.Seats, &b.Status, &b.CreatedAt, &b.Paid, &b.PaymentAvailable)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, rows.Err()
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
	var paid bool
	queryBooking := `SELECT trip_id, seats, status, stripe_payment_intent IS NOT NULL FROM bookings WHERE id = $1 AND passenger_id = $2 FOR UPDATE`
	err = tx.QueryRow(ctx, queryBooking, bookingID, passengerID).Scan(&tripID, &seats, &status, &paid)
	if err != nil {
		return errors.New("бронь не найдена или не принадлежит вам")
	}

	if paid {
		return errors.New("оплаченную бронь нельзя отменить")
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

func (r *BookingRepository) CleanupExpired(ctx context.Context, expireAfter time.Duration) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Передаем время как параметр, чтобы легко менять в тестах
	intervalStr := fmt.Sprintf("%d seconds", int(expireAfter.Seconds()))
	query := `
		UPDATE bookings 
		SET status = 'cancelled' 
		WHERE status = 'pending' AND created_at < NOW() - $1::interval
		RETURNING trip_id, seats
	`
	rows, err := tx.Query(ctx, query, intervalStr)
	if err != nil {
		return 0, err
	}

	type cancelled struct {
		tripID uuid.UUID
		seats  int
	}
	var items []cancelled

	for rows.Next() {
		var c cancelled
		if err := rows.Scan(&c.tripID, &c.seats); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, c)
	}
	rows.Close()

	for _, c := range items {
		_, err := tx.Exec(ctx, `UPDATE trips SET seats_available = seats_available + $1 WHERE id = $2`, c.seats, c.tripID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(items), nil
}
