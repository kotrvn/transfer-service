-- +goose Up
ALTER TABLE bookings ADD COLUMN stripe_payment_intent TEXT UNIQUE;

-- +goose Down
ALTER TABLE bookings DROP COLUMN stripe_payment_intent;
