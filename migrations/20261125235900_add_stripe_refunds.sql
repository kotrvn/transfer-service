-- +goose Up
CREATE TABLE stripe_refunds (
    payment_intent TEXT PRIMARY KEY,
    booking_id UUID NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'pending', 'requires_action', 'succeeded', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX stripe_refunds_due_idx ON stripe_refunds (next_attempt_at)
    WHERE status IN ('queued', 'pending', 'requires_action');

-- +goose Down
DROP TABLE stripe_refunds;
