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
const bookingUnexpired = "b.created_at + $2::bigint * interval '1 second' > clock_timestamp()"

func (r *BookingRepository) CheckoutAmount(ctx context.Context, bookingID, passengerID uuid.UUID) (int64, error) {
	var amount int64
	var payable bool
	err := r.pool.QueryRow(ctx, `
		SELECT `+bookingAmount+`, b.status = 'pending' AND `+bookingUnexpired+`
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

// ConfirmPayment returns false when the charge needs to be refunded.
func (r *BookingRepository) ConfirmPayment(ctx context.Context, bookingID uuid.UUID, intent string, amount int64) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

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
		UPDATE bookings b SET status = 'confirmed', stripe_payment_intent = $3
		FROM trips t WHERE t.id = b.trip_id AND b.id = $1
		AND b.status = 'pending' AND `+bookingUnexpired+` AND `+bookingAmount+` = $4`,
		bookingID, int64(models.BookingHold.Seconds()), intent, amount)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 1, tx.Commit(ctx)
}
