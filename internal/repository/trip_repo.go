package repository

import (
	"context"
	"fmt"
	"transfer-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
	query := `SELECT id, driver_id, from_city, to_city, departure_at, seats_total, seats_available, price_per_seat 
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
		err := rows.Scan(&t.ID, &t.DriverID, &t.FromCity, &t.ToCity, &t.DepartureAt, &t.SeatsTotal, &t.SeatsAvailable, &t.PricePerSeat)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &t)
	}
	return trips, nil
}
