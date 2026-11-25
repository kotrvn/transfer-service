package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/models"
	"transfer-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookingRepo interface {
	Create(ctx context.Context, booking *models.Booking) error
	GetByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]*models.Booking, error)
	Cancel(ctx context.Context, bookingID uuid.UUID, passengerID uuid.UUID) error
}

type BookingHandler struct {
	repo            BookingRepo
	PaymentsEnabled bool
}

func NewBookingHandler(repo BookingRepo) *BookingHandler {
	return &BookingHandler{repo: repo}
}

// Create godoc
// @Summary Забронировать поездку
// @Description Создает новую бронь на поездку. Доступно только для пассажиров.
// @Tags bookings
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.BookingCreateRequest true "Данные бронирования"
// @Success 201 {object} models.Booking
// @Failure 400,401,403,404,500 {string} string
// @Router /bookings [post]
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем роль
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "passenger" {
		http.Error(w, "Только пассажиры могут бронировать", http.StatusForbidden)
		return
	}

	// 2. Читаем тело запроса
	var req models.BookingCreateRequest
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Достаем ID пассажира из контекста
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	passengerID, _ := uuid.Parse(userIDStr)

	// 4. Создаем объект брони
	booking := &models.Booking{
		ID:          uuid.New(),
		TripID:      req.TripID,
		PassengerID: passengerID,
		Seats:       req.Seats,
		Status:      "pending",
	}

	// 5. Сохраняем (с обработкой кастомных ошибок)
	err := h.repo.Create(r.Context(), booking)
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			http.Error(w, "Поездка не найдена", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrNotEnoughSeats) {
			http.Error(w, "Недостаточно свободных мест", http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// ListMy godoc
// @Summary Мои бронирования
// @Description Возвращает список всех броней текущего авторизованного пассажира.
// @Tags bookings
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Booking
// @Failure 401,500 {string} string
// @Router /bookings/my [get]
func (h *BookingHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	passengerID, _ := uuid.Parse(userIDStr)

	bookings, err := h.repo.GetByPassengerID(r.Context(), passengerID)
	if err != nil {
		http.Error(w, "Ошибка получения броней", http.StatusInternalServerError)
		return
	}

	if !h.PaymentsEnabled {
		for _, booking := range bookings {
			booking.PaymentAvailable = false
		}
	}

	if bookings == nil {
		bookings = []*models.Booking{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

// Cancel godoc
// @Summary Отменить бронь
// @Description Отменяет бронь владельца и возвращает места. Для оплаченной через Stripe брони возвращает 400.
// @Tags bookings
// @Produce json
// @Security Bearer
// @Param id path string true "ID брони (UUID)"
// @Success 204 "No Content"
// @Failure 400,401,403,404 {string} string
// @Router /bookings/{id} [delete]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	// Достаем ID брони из URL (например, /bookings/123e4567-e89b-12d3...)
	bookingIDStr := chi.URLParam(r, "id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		http.Error(w, "Неверный ID брони", http.StatusBadRequest)
		return
	}

	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	passengerID, _ := uuid.Parse(userIDStr)

	err = h.repo.Cancel(r.Context(), bookingID, passengerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content — успешно удалено
}
