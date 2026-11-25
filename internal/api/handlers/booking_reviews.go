package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/models"

	"transfer-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DriverBookings godoc
// @Summary Заявки и оценки пассажиров для собственных поездок
// @Tags bookings
// @Produce json
// @Security Bearer
// @Success 200 {array} models.DriverBooking
// @Failure 401,403,500 {string} string
// @Router /bookings/driver [get]
func (h *BookingHandler) DriverBookings(w http.ResponseWriter, r *http.Request) {
	actor, ok := bookingActor(w, r, "driver")
	if !ok {
		return
	}
	bookings, err := h.repo.DriverBookings(r.Context(), actor)
	if err != nil {
		http.Error(w, "Ошибка получения заявок", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(bookings)
}

type bookingDecision struct {
	Decision string `json:"decision" validate:"required,oneof=approve reject"`
}

// Decide godoc
// @Summary Принять или отклонить заявку на свою поездку
// @Tags bookings
// @Accept json
// @Security Bearer
// @Param id path string true "ID брони"
// @Param request body bookingDecision true "approve или reject"
// @Success 204 "No Content"
// @Failure 400,401,403,404,409,500 {string} string
// @Router /bookings/{id}/decision [post]
func (h *BookingHandler) Decide(w http.ResponseWriter, r *http.Request) {
	actor, ok := bookingActor(w, r, "driver")
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный ID брони", http.StatusBadRequest)
		return
	}
	var req bookingDecision
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeBookingAction(w, h.repo.DecideRequest(r.Context(), id, actor, req.Decision == "approve"))
}

// Rate godoc
// @Summary Оценить другую сторону завершённой брони
// @Tags bookings
// @Accept json
// @Security Bearer
// @Param id path string true "ID брони"
// @Param request body models.BookingRatingRequest true "Оценка от 1 до 5 и необязательный отзыв до 1000 символов"
// @Success 204 "No Content"
// @Failure 400,401,403,409,500 {string} string
// @Router /bookings/{id}/rating [post]
func (h *BookingHandler) Rate(w http.ResponseWriter, r *http.Request) {
	actor, ok := bookingActor(w, r, "")
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный ID брони", http.StatusBadRequest)
		return
	}
	var req models.BookingRatingRequest
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	writeBookingAction(w, h.repo.Rate(r.Context(), id, actor, role == "driver", req.Score, strings.TrimSpace(req.Review)))
}

func bookingActor(w http.ResponseWriter, r *http.Request, requiredRole string) (uuid.UUID, bool) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if (requiredRole != "" && role != requiredRole) || (role != "driver" && role != "passenger") {
		http.Error(w, "Недостаточно прав", http.StatusForbidden)
		return uuid.Nil, false
	}
	raw, _ := r.Context().Value(middleware.UserIDKey).(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		http.Error(w, "Неверный ID пользователя", http.StatusUnauthorized)
		return uuid.Nil, false
	}
	return id, true
}

func writeBookingAction(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrBookingNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, repository.ErrBookingActionUnavailable):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		http.Error(w, "Не удалось выполнить действие", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
