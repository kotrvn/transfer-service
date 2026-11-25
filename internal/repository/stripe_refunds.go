package repository

import (
	"context"
	"errors"
	"time"

	"transfer-service/internal/models"

	"github.com/jackc/pgx/v5"
)

// ClaimRefund gives a worker a 60-second lease and a fencing token in Attempts.
func (r *BookingRepository) ClaimRefund(ctx context.Context) (*models.RefundJob, error) {
	var job models.RefundJob
	err := r.pool.QueryRow(ctx, `
		WITH due AS (
			SELECT payment_intent FROM stripe_refunds
			WHERE status IN ('queued', 'pending', 'requires_action') AND next_attempt_at <= now()
			ORDER BY next_attempt_at, payment_intent LIMIT 1 FOR UPDATE SKIP LOCKED
		)
		UPDATE stripe_refunds r SET attempts = attempts + 1, next_attempt_at = now() + interval '60 seconds', updated_at = now()
		FROM due WHERE r.payment_intent = due.payment_intent
		RETURNING r.payment_intent, r.amount_cents, r.status, r.attempts`).Scan(&job.Intent, &job.Amount, &job.Status, &job.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *BookingRepository) UpdateRefund(ctx context.Context, job models.RefundJob, status, lastError string, delay time.Duration) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE stripe_refunds SET status = $3, last_error = NULLIF($4, ''),
			next_attempt_at = now() + $5::double precision * interval '1 second', updated_at = now()
		WHERE payment_intent = $1 AND attempts = $2`, job.Intent, job.Attempts, status, lastError, delay.Seconds())
	return err
}
