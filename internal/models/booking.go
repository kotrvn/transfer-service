package models

import (
	"time"

	"github.com/google/uuid"
)

const BookingHold = 15 * time.Minute

type Booking struct {
	ID               uuid.UUID `json:"id"`
	TripID           uuid.UUID `json:"trip_id"`
	PassengerID      uuid.UUID `json:"passenger_id"`
	Seats            int       `json:"seats"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	Paid             bool      `json:"paid"`
	PaymentAvailable bool      `json:"payment_available"`
}

type BookingCreateRequest struct {
	TripID uuid.UUID `json:"trip_id" validate:"required"`
	Seats  int       `json:"seats" validate:"required,gt=0"`
}
