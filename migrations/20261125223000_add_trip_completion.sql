-- +goose Up
ALTER TABLE trips ADD COLUMN completed_at TIMESTAMPTZ;

-- +goose Down
UPDATE bookings SET status = 'cancelled' WHERE status = 'expired';
ALTER TABLE trips DROP COLUMN completed_at;
