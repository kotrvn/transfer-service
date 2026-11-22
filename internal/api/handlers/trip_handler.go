package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/models"

	"github.com/google/uuid"
)

type TripRepo interface {
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
