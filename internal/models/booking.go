package models

import "github.com/google/uuid"

type Booking struct {
	ID          uuid.UUID `json:"id"`
	TripID      uuid.UUID `json:"trip_id"`
	PassengerID uuid.UUID `json:"passenger_id"`
	Seats       int       `json:"seats"`
	Status      string    `json:"status"`
}

type BookingCreateRequest struct {
	TripID uuid.UUID `json:"trip_id" validate:"required"`
	Seats  int       `json:"seats" validate:"required,gt=0"`
}
