-- +goose Up
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES users(id),
    from_city VARCHAR(100) NOT NULL,
    to_city VARCHAR(100) NOT NULL,
    departure_at TIMESTAMP WITH TIME ZONE NOT NULL,
    seats_total INTEGER NOT NULL,
    seats_available INTEGER NOT NULL,
    price_per_seat NUMERIC(10, 2) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS trips;