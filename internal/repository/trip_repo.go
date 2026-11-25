package repository

import (
	"context"
	"errors"
	"fmt"
	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTripNotDeparted = errors.New("поездка ещё не отправилась")

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
	query := `SELECT id, driver_id, from_city, to_city, departure_at, seats_total, seats_available, price_per_seat, completed_at
	          FROM trips WHERE 1=1`
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
		err := rows.Scan(&t.ID, &t.DriverID, &t.FromCity, &t.ToCity, &t.DepartureAt, &t.SeatsTotal, &t.SeatsAvailable, &t.PricePerSeat, &t.CompletedAt)
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
	if _, err := tx.Exec(ctx, "UPDATE trips SET completed_at = clock_timestamp() WHERE id = $1", tripID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
