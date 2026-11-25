package models

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	DriverName        string     `json:"driver_name"`
	DriverRating      *float64   `json:"driver_rating"`
	DriverRatingCount int        `json:"driver_rating_count"`
	ID                uuid.UUID  `json:"id"`
	DriverID          uuid.UUID  `json:"driver_id"`
	FromCity          string     `json:"from_city"`
	ToCity            string     `json:"to_city"`
	DepartureAt       time.Time  `json:"departure_at"`
	SeatsTotal        int        `json:"seats_total"`
	SeatsAvailable    int        `json:"seats_available"`
	PricePerSeat      float64    `json:"price_per_seat"`
	CompletedAt       *time.Time `json:"completed_at"`
}

type TripCreateRequest struct {
	FromCity     string    `json:"from_city" validate:"required"`
	ToCity       string    `json:"to_city" validate:"required"`
	DepartureAt  time.Time `json:"departure_at" validate:"required"`
	SeatsTotal   int       `json:"seats_total" validate:"required,gt=0"`
	PricePerSeat float64   `json:"price_per_seat" validate:"required,gte=0"`
}

type DriverBalance struct {
	Currency       string      `json:"currency"`
	LockedCents    int64       `json:"locked_cents"`
	AvailableCents int64       `json:"available_cents"`
	ReadyTripIDs   []uuid.UUID `json:"ready_trip_ids"`
}
