package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Мок для поездок
func (m *mockTripRepo) Complete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type mockTripRepo struct {
	trips []*models.Trip
}

func (m *mockTripRepo) Create(ctx context.Context, trip *models.Trip) error {
	m.trips = append(m.trips, trip)
	return nil
}

func (m *mockTripRepo) List(ctx context.Context, fromCity, toCity string) ([]*models.Trip, error) {
	return m.trips, nil
}

func TestCreateTrip_Success(t *testing.T) {
	mockRepo := &mockTripRepo{}
	handler := NewTripHandler(mockRepo)

	reqBody := `{"from_city": "Miami", "to_city": "Warsaw", "departure_at": "2026-11-20T10:00:00Z", "seats_total": 4, "price_per_seat": 100}`
	req, _ := http.NewRequest(http.MethodPost, "/trips", bytes.NewBufferString(reqBody))

	// Имитируем middleware: кладем роль и ID в контекст
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "driver")
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdTrip models.Trip
	json.NewDecoder(rr.Body).Decode(&createdTrip)
	assert.Equal(t, "Miami", createdTrip.FromCity)
	assert.Equal(t, 4, createdTrip.SeatsAvailable) // Проверяем, что места заполнились
}

func TestCreateTrip_ForbiddenForPassenger(t *testing.T) {
	mockRepo := &mockTripRepo{}
	handler := NewTripHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodPost, "/trips", bytes.NewBufferString(`{}`))
	// Кладем роль "passenger", хотя ручка только для водителей
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "passenger")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Только водители")
}

func TestListTrips_Success(t *testing.T) {
	// Заранее готовим данные
	mockRepo := &mockTripRepo{
		trips: []*models.Trip{{ID: uuid.New(), FromCity: "Paris", ToCity: "Rome", DepartureAt: time.Now()}},
	}
	handler := NewTripHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodGet, "/trips?from=Paris", nil)
	rr := httptest.NewRecorder()
	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Paris")
}

func (m *mockTripRepo) DriverBalance(context.Context, uuid.UUID) (models.DriverBalance, error) {
	return models.DriverBalance{}, nil
}
