package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/models"
	"transfer-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Мок для броней
type mockBookingRepo struct {
	errToReturn error
}

func (m *mockBookingRepo) Create(ctx context.Context, booking *models.Booking) error {
	return m.errToReturn
}

func (m *mockBookingRepo) GetByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]*models.Booking, error) {
	return []*models.Booking{{ID: uuid.New(), Status: "confirmed"}}, nil
}

func (m *mockBookingRepo) Cancel(ctx context.Context, bookingID uuid.UUID, passengerID uuid.UUID) error {
	return m.errToReturn
}

func TestCreateBooking_Success(t *testing.T) {
	mockRepo := &mockBookingRepo{}
	handler := NewBookingHandler(mockRepo)

	reqBody := `{"trip_id": "` + uuid.New().String() + `", "seats": 2}`
	req, _ := http.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString(reqBody))

	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "passenger")
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestCreateBooking_NotEnoughSeats(t *testing.T) {
	// Говорим моку вернуть ошибку нехватки мест
	mockRepo := &mockBookingRepo{errToReturn: repository.ErrNotEnoughSeats}
	handler := NewBookingHandler(mockRepo)

	reqBody := `{"trip_id": "` + uuid.New().String() + `", "seats": 2}`
	req, _ := http.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString(reqBody))

	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "passenger")
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Недостаточно свободных мест")
}

func TestCancelBooking_Success(t *testing.T) {
	mockRepo := &mockBookingRepo{}
	handler := NewBookingHandler(mockRepo)

	bookingID := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, "/bookings/"+bookingID.String(), nil)

	// Имитируем chi.URLParam (извлечение ID из URL)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", bookingID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

	// Добавляем данные пользователя
	ctx = context.WithValue(ctx, middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Cancel(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code) // 204 при успехе
}

func (m *mockBookingRepo) ConfirmArrival(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *mockBookingRepo) DriverBookings(context.Context, uuid.UUID) ([]models.DriverBooking, error) {
	return nil, nil
}
func (m *mockBookingRepo) DecideRequest(context.Context, uuid.UUID, uuid.UUID, bool) error {
	return nil
}
func (m *mockBookingRepo) Rate(context.Context, uuid.UUID, uuid.UUID, bool, int, string) error {
	return nil
}
