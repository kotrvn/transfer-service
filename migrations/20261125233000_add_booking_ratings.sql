-- +goose Up
ALTER TABLE bookings ADD COLUMN driver_rating SMALLINT CHECK (driver_rating BETWEEN 1 AND 5),
  ADD COLUMN passenger_rating SMALLINT CHECK (passenger_rating BETWEEN 1 AND 5),
  ADD COLUMN payment_opened_at TIMESTAMPTZ;
CREATE INDEX bookings_passenger_ratings_idx ON bookings (passenger_id) WHERE passenger_rating IS NOT NULL;
CREATE INDEX bookings_trip_id_idx ON bookings (trip_id);

-- +goose Down
UPDATE bookings SET status = 'cancelled' WHERE status = 'rejected';
UPDATE bookings SET status = 'pending' WHERE status = 'requested';
DROP INDEX bookings_trip_id_idx;
DROP INDEX bookings_passenger_ratings_idx;
ALTER TABLE bookings DROP COLUMN driver_rating, DROP COLUMN passenger_rating, DROP COLUMN payment_opened_at;
