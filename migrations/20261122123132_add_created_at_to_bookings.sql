-- +goose Up
ALTER TABLE bookings ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL;

-- +goose Down
ALTER TABLE bookings DROP COLUMN created_at;
