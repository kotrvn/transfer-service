package repository

import (
	"context"
	"errors"

	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrBookingActionUnavailable = errors.New("действие недоступно для этой брони")

const visibleBookingStatus = `CASE
	WHEN b.status = 'confirmed' AND t.completed_at IS NOT NULL THEN 'completed'
	WHEN b.status IN ('requested', 'pending') AND t.completed_at IS NOT NULL THEN 'expired'
	ELSE b.status END`

func (r *BookingRepository) DriverBookings(ctx context.Context, driverID uuid.UUID) ([]models.DriverBooking, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id, b.trip_id, b.seats, u.name, rating.average, rating.count, b.passenger_rating, b.passenger_review,
			`+visibleBookingStatus+`, b.status = 'requested' AND t.completed_at IS NULL AND `+bookingUnexpired+`
		FROM trips t JOIN bookings b ON b.trip_id = t.id JOIN users u ON u.id = b.passenger_id
		LEFT JOIN LATERAL (
			SELECT AVG(passenger_rating)::float8 AS average, COUNT(passenger_rating) AS count
			FROM bookings WHERE passenger_id = b.passenger_id AND passenger_rating IS NOT NULL
		) rating ON true
		WHERE t.driver_id = $1 ORDER BY b.created_at DESC, b.id`, driverID, int64(models.BookingHold.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.DriverBooking, 0)
	for rows.Next() {
		var b models.DriverBooking
		if err := rows.Scan(&b.ID, &b.TripID, &b.Seats, &b.PassengerName, &b.PassengerRating, &b.PassengerRatingCount,
			&b.RatingGiven, &b.ReviewGiven, &b.Status, &b.DecisionAvailable); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

// Rate accepts one score per side of a completed booking.
func (r *BookingRepository) Rate(ctx context.Context, id, actor uuid.UUID, driver bool, score int, review string) error {
	column, reviewColumn, owner := "driver_rating", "driver_review", "b.passenger_id"
	if driver {
		column, reviewColumn, owner = "passenger_rating", "passenger_review", "t.driver_id"
	}
	result, err := r.pool.Exec(ctx, `
		UPDATE bookings b SET `+column+` = $3, `+reviewColumn+` = NULLIF($4, '') FROM trips t
		WHERE t.id = b.trip_id AND b.id = $1 AND `+owner+` = $2
		AND b.status = 'confirmed' AND t.completed_at IS NOT NULL AND b.`+column+` IS NULL`, id, actor, score, review)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrBookingActionUnavailable
	}
	return nil
}

func (r *BookingRepository) DecideRequest(ctx context.Context, id, driverID uuid.UUID, approve bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var tripID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT t.id FROM trips t JOIN bookings b ON b.trip_id = t.id
		WHERE b.id = $1 AND t.driver_id = $2 FOR KEY SHARE OF t`, id, driverID).Scan(&tripID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrBookingNotFound
	}
	if err != nil {
		return err
	}
	var seats int
	err = tx.QueryRow(ctx, `
		UPDATE bookings b SET
			status = CASE WHEN $3 THEN 'pending' ELSE 'rejected' END,
			payment_opened_at = CASE WHEN $3 THEN clock_timestamp() END
		FROM trips t WHERE t.id = b.trip_id AND b.id = $1
		AND b.status = 'requested' AND t.completed_at IS NULL AND `+bookingUnexpired+`
		RETURNING b.seats`, id, int64(models.BookingHold.Seconds()), approve).Scan(&seats)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrBookingActionUnavailable
	}
	if err != nil {
		return err
	}
	if !approve {
		if _, err := tx.Exec(ctx, "UPDATE trips SET seats_available = seats_available + $1 WHERE id = $2", seats, tripID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
