-- +goose Up
ALTER TABLE bookings
  ADD COLUMN driver_review TEXT CHECK (char_length(driver_review) <= 1000),
  ADD COLUMN passenger_review TEXT CHECK (char_length(passenger_review) <= 1000);

-- +goose Down
ALTER TABLE bookings DROP COLUMN driver_review, DROP COLUMN passenger_review;
