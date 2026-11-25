-- +goose Up
ALTER TABLE bookings ADD COLUMN arrived_at TIMESTAMPTZ, ADD COLUMN paid_amount_cents BIGINT;
UPDATE bookings b SET paid_amount_cents = (t.price_per_seat * b.seats * 100)::bigint
FROM trips t WHERE t.id = b.trip_id AND b.stripe_payment_intent IS NOT NULL;

-- +goose Down
ALTER TABLE bookings DROP COLUMN arrived_at, DROP COLUMN paid_amount_cents;
