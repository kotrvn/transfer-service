package repository

import (
	"context"
	"errors"
	"fmt"
	"transfer-service/internal/models"
	"transfer-service/internal/payment"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTripNotDeparted = errors.New("поездка ещё не отправилась")

var ErrPassengerArrivalRequired = errors.New("ожидаем подтверждения от пассажиров")

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (r *TripRepository) Create(ctx context.Context, trip *models.Trip) error {
	query := `INSERT INTO trips (id, driver_id, from_city, to_city, departure_at, seats_total, seats_available, price_per_seat) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		trip.ID, trip.DriverID, trip.FromCity, trip.ToCity,
		trip.DepartureAt, trip.SeatsTotal, trip.SeatsAvailable, trip.PricePerSeat,
	)
	return err
}

func (r *TripRepository) List(ctx context.Context, fromCity, toCity string) ([]*models.Trip, error) {
	query := `
	SELECT t.id, t.driver_id, t.from_city, t.to_city, t.departure_at, t.seats_total,
		t.seats_available, t.price_per_seat, t.completed_at, u.name, rating.average, rating.count
	FROM trips t JOIN users u ON u.id = t.driver_id LEFT JOIN LATERAL (
		SELECT AVG(b.driver_rating)::float8 AS average, COUNT(b.driver_rating) AS count
		FROM trips history JOIN bookings b ON b.trip_id = history.id
		WHERE history.driver_id = t.driver_id AND b.driver_rating IS NOT NULL
	) rating ON true WHERE 1=1`
	var args []interface{}
	argID := 1

	if fromCity != "" {
		query += fmt.Sprintf(" AND from_city ILIKE $%d", argID)
		args = append(args, "%"+fromCity+"%")
		argID++
	}
	if toCity != "" {
		query += fmt.Sprintf(" AND to_city ILIKE $%d", argID)
		args = append(args, "%"+toCity+"%")
		argID++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	// ... остальной код сканирования rows (он не меняется)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var t models.Trip
		err := rows.Scan(&t.ID, &t.DriverID, &t.FromCity, &t.ToCity, &t.DepartureAt, &t.SeatsTotal, &t.SeatsAvailable, &t.PricePerSeat, &t.CompletedAt, &t.DriverName, &t.DriverRating, &t.DriverRatingCount)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &t)
	}
	return trips, rows.Err()
}

func (r *TripRepository) Complete(ctx context.Context, tripID, driverID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var departed, completed bool
	err = tx.QueryRow(ctx, `
		SELECT departure_at <= clock_timestamp(), completed_at IS NOT NULL
		FROM trips WHERE id = $1 AND driver_id = $2 FOR UPDATE`, tripID, driverID).Scan(&departed, &completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTripNotFound
	}
	if err != nil || completed {
		return err
	}
	if !departed {
		return ErrTripNotDeparted
	}
	var ready bool
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) > 0 AND COUNT(*) FILTER (WHERE arrived_at IS NULL) = 0
		FROM bookings WHERE trip_id = $1 AND status = 'confirmed'`, tripID).Scan(&ready); err != nil {
		return err
	}
	if !ready {
		return ErrPassengerArrivalRequired
	}
	if _, err := tx.Exec(ctx, "UPDATE trips SET completed_at = clock_timestamp() WHERE id = $1", tripID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *TripRepository) DriverBalance(ctx context.Context, driverID uuid.UUID) (models.DriverBalance, error) {
	balance := models.DriverBalance{Currency: payment.Currency}
	err := r.pool.QueryRow(ctx, `
		WITH earnings AS (
			SELECT t.id, t.completed_at, t.departure_at,
				COALESCE(SUM(b.paid_amount_cents), 0) AS amount,
				COUNT(b.id) > 0 AND COUNT(b.id) FILTER (WHERE b.arrived_at IS NULL) = 0 AS all_arrived
			FROM trips t LEFT JOIN bookings b ON b.trip_id = t.id AND b.status = 'confirmed'
			WHERE t.driver_id = $1 GROUP BY t.id
		)
		SELECT COALESCE(SUM(amount) FILTER (WHERE completed_at IS NULL), 0)::bigint,
			COALESCE(SUM(amount) FILTER (WHERE completed_at IS NOT NULL), 0)::bigint,
			COALESCE(array_agg(id) FILTER (WHERE completed_at IS NULL AND departure_at <= clock_timestamp() AND all_arrived), ARRAY[]::uuid[])
		FROM earnings`, driverID).Scan(&balance.LockedCents, &balance.AvailableCents, &balance.ReadyTripIDs)
	return balance, err
}
