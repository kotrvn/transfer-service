package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrArrivalUnavailable = errors.New("подтвердить прибытие можно после отправления по подтверждённой брони")

func (r *BookingRepository) ConfirmArrival(ctx context.Context, bookingID, passengerID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Водитель завершает поездку под FOR UPDATE на той же строке.
	var tripID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT t.id FROM trips t JOIN bookings b ON b.trip_id = t.id
		WHERE b.id = $1 AND b.passenger_id = $2 FOR KEY SHARE OF t`, bookingID, passengerID).Scan(&tripID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrBookingNotFound
	}
	if err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `
		UPDATE bookings b SET arrived_at = COALESCE(b.arrived_at, clock_timestamp())
		FROM trips t WHERE t.id = b.trip_id AND b.id = $1
		AND (b.arrived_at IS NOT NULL OR (b.status = 'confirmed' AND t.completed_at IS NULL AND t.departure_at <= clock_timestamp()))`, bookingID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrArrivalUnavailable
	}
	return tx.Commit(ctx)
}
