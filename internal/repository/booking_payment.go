package repository

import (
	"context"
	"errors"

	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrBookingNotFound = errors.New("бронь не найдена")
var ErrBookingNotPayable = errors.New("бронь недоступна для оплаты")

const bookingAmount = "(t.price_per_seat * b.seats * 100)::bigint"
const bookingUnexpired = "COALESCE(b.payment_opened_at, b.created_at) + $2::bigint * interval '1 second' > clock_timestamp()"

func (r *BookingRepository) CheckoutAmount(ctx context.Context, bookingID, passengerID uuid.UUID) (int64, error) {
	var amount int64
	var payable bool
	err := r.pool.QueryRow(ctx, `
		SELECT `+bookingAmount+`, b.status = 'pending' AND t.completed_at IS NULL AND `+bookingUnexpired+`
		FROM bookings b JOIN trips t ON t.id = b.trip_id
		WHERE b.id = $1 AND b.passenger_id = $3`,
		bookingID, int64(models.BookingHold.Seconds()), passengerID).Scan(&amount, &payable)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrBookingNotFound
	}
	if err == nil && !payable {
		return 0, ErrBookingNotPayable
	}
	return amount, err
}

// ConfirmPayment applies a payment or durably queues its refund in the same transaction.
func (r *BookingRepository) ConfirmPayment(ctx context.Context, bookingID uuid.UUID, intent string, amount int64) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	confirmed, err := applyPayment(ctx, tx, bookingID, intent, amount)
	if err != nil {
		return false, err
	}
	if !confirmed {
		if _, err := tx.Exec(ctx, `
			INSERT INTO stripe_refunds (payment_intent, booking_id, amount_cents)
			VALUES ($1, $2, $3) ON CONFLICT (payment_intent) DO NOTHING`, intent, bookingID, amount); err != nil {
			return false, err
		}
	}
	return confirmed, tx.Commit(ctx)
}

func applyPayment(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, intent string, amount int64) (bool, error) {
	// Serialize with the owner's completion while allowing seat updates on the trip.
	var tripID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT t.id FROM trips t JOIN bookings b ON b.trip_id = t.id
		WHERE b.id = $1 FOR KEY SHARE OF t`, bookingID).Scan(&tripID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var paidIntent string
	err = tx.QueryRow(ctx, `SELECT COALESCE(stripe_payment_intent, '') FROM bookings WHERE id = $1 FOR UPDATE`, bookingID).Scan(&paidIntent)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if paidIntent != "" {
		return paidIntent == intent, nil
	}

	result, err := tx.Exec(ctx, `
		UPDATE bookings b SET status = 'confirmed', stripe_payment_intent = $3, paid_amount_cents = $4
		FROM trips t WHERE t.id = b.trip_id AND b.id = $1
		AND b.status = 'pending' AND t.completed_at IS NULL AND `+bookingUnexpired+` AND `+bookingAmount+` = $4
		AND NOT EXISTS (SELECT 1 FROM stripe_refunds WHERE payment_intent = $3)`,
		bookingID, int64(models.BookingHold.Seconds()), intent, amount)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 1, nil
}
