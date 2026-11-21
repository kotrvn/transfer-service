-- +goose Up
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    trip_id UUID NOT NULL REFERENCES trips(id),
    passenger_id UUID NOT NULL REFERENCES users(id),
    seats INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed'
);

-- +goose Down
DROP TABLE IF EXISTS bookings;