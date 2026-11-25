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

type TripRepo interface {
	DriverBalance(context.Context, uuid.UUID) (models.DriverBalance, error)
	Complete(ctx context.Context, tripID, driverID uuid.UUID) error
	Create(ctx context.Context, trip *models.Trip) error
	List(ctx context.Context, fromCity, toCity string) ([]*models.Trip, error)
}

type TripHandler struct {
	repo TripRepo
}

func NewTripHandler(repo TripRepo) *TripHandler {
	return &TripHandler{repo: repo}
}

// Create godoc
// @Summary Создать поездку
// @Description Добавляет новую поездку. Доступно только для водителей.
// @Tags trips
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.TripCreateRequest true "Данные поездки"
// @Success 201 {object} models.Trip
// @Failure 400,401,403,500 {string} string
// @Router /trips [post]
func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем роль (берем из контекста, куда её положил middleware)
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "driver" {
		http.Error(w, "Только водители могут создавать поездки", http.StatusForbidden)
		return
	}

	// 2. Читаем тело
	var req models.TripCreateRequest
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Достаем ID водителя из токена
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	driverID, _ := uuid.Parse(userIDStr)

	// 4. Создаем поездку
	trip := &models.Trip{
		ID:             uuid.New(),
		DriverID:       driverID,
		FromCity:       req.FromCity,
		ToCity:         req.ToCity,
		DepartureAt:    req.DepartureAt,
		SeatsTotal:     req.SeatsTotal,
		SeatsAvailable: req.SeatsTotal, // Все места свободны изначально
		PricePerSeat:   req.PricePerSeat,
	}

	if err := h.repo.Create(r.Context(), trip); err != nil {
		http.Error(w, "Ошибка сохранения поездки", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trip)
}

// List godoc
// @Summary Получить список поездок
// @Description Возвращает список поездок с возможностью фильтрации по городам.
// @Tags trips
// @Produce json
// @Param from query string false "Город отправления (например, Berlin)"
// @Param to query string false "Город прибытия (например, Warsaw)"
// @Success 200 {array} models.Trip
// @Failure 500 {string} string
// @Router /trips [get]
func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	// Читаем query-параметры из URL (например, ?from=Berlin&to=Warsaw)
	fromCity := r.URL.Query().Get("from")
	toCity := r.URL.Query().Get("to")

	trips, err := h.repo.List(r.Context(), fromCity, toCity)
	if err != nil {
		http.Error(w, "Ошибка получения поездок", http.StatusInternalServerError)
		return
	}

	if trips == nil {
		trips = []*models.Trip{} // Чтобы возвращался [] а не null
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

// Complete godoc
// @Summary Завершить поездку
// @Description Водитель-владелец отмечает поездку завершённой после времени отправления и подтверждения прибытия всеми пассажирами с подтверждённой бронью. Повторный запрос сохраняет результат.
// @Tags trips
// @Security Bearer
// @Param id path string true "ID поездки (UUID)"
// @Success 204 "No Content"
// @Failure 400,401,403,404,409,500 {string} string
// @Router /trips/{id}/complete [post]
func (h *TripHandler) Complete(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "driver" {
		http.Error(w, "Только водитель может завершить поездку", http.StatusForbidden)
		return
	}
	tripID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный ID поездки", http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	driverID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Неверный ID водителя", http.StatusUnauthorized)
		return
	}
	err = h.repo.Complete(r.Context(), tripID, driverID)
	switch {
	case errors.Is(err, repository.ErrTripNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, repository.ErrTripNotDeparted), errors.Is(err, repository.ErrPassengerArrivalRequired):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		http.Error(w, "Не удалось завершить поездку", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// Balance godoc
// @Summary Получить внутренний тестовый баланс водителя
// @Tags trips
// @Produce json
// @Security Bearer
// @Success 200 {object} models.DriverBalance
// @Failure 401,403,500 {string} string
// @Router /trips/balance [get]
func (h *TripHandler) Balance(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "driver" {
		http.Error(w, "Баланс доступен водителям", http.StatusForbidden)
		return
	}
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	driverID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Неверный ID водителя", http.StatusUnauthorized)
		return
	}
	balance, err := h.repo.DriverBalance(r.Context(), driverID)
	if err != nil {
		http.Error(w, "Не удалось получить баланс", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(balance)
}
