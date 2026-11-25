package models

import (
	"time"

	"github.com/google/uuid"
)

const BookingHold = 15 * time.Minute

type Booking struct {
	DriverReview     *string    `json:"driver_review"`
	DriverRating     *int       `json:"driver_rating"`
	ArrivedAt        *time.Time `json:"arrived_at"`
	ArrivalAvailable bool       `json:"arrival_available"`
	ID               uuid.UUID  `json:"id"`
	TripID           uuid.UUID  `json:"trip_id"`
	PassengerID      uuid.UUID  `json:"passenger_id"`
	Seats            int        `json:"seats"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	Paid             bool       `json:"paid"`
	PaymentAvailable bool       `json:"payment_available"`
}

type BookingCreateRequest struct {
	TripID uuid.UUID `json:"trip_id" validate:"required"`
	Seats  int       `json:"seats" validate:"required,gt=0"`
}

type BookingRatingRequest struct {
	Review string `json:"review" validate:"max=1000"`
	Score  int    `json:"score" validate:"required,min=1,max=5"`
}

type DriverBooking struct {
	Seats                int       `json:"seats"`
	ReviewGiven          *string   `json:"review_given"`
	ID                   uuid.UUID `json:"id"`
	TripID               uuid.UUID `json:"trip_id"`
	PassengerName        string    `json:"passenger_name"`
	PassengerRating      *float64  `json:"passenger_rating"`
	PassengerRatingCount int       `json:"passenger_rating_count"`
	RatingGiven          *int      `json:"rating_given"`
	Status               string    `json:"status"`
	DecisionAvailable    bool      `json:"decision_available"`
}
