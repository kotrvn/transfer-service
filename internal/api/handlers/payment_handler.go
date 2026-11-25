package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"transfer-service/internal/api/middleware"
	"transfer-service/internal/payment"
	"transfer-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaymentRepo interface {
	CheckoutAmount(context.Context, uuid.UUID, uuid.UUID) (int64, error)
	ConfirmPayment(context.Context, uuid.UUID, string, int64) (bool, error)
}

type PaymentHandler struct {
	repo   PaymentRepo
	stripe *payment.Stripe
}

func NewPaymentHandler(repo PaymentRepo, stripe *payment.Stripe) *PaymentHandler {
	return &PaymentHandler{repo, stripe}
}

type checkoutResponse struct {
	URL string `json:"url"`
}

// Checkout godoc
// @Summary Начать тестовую оплату брони
// @Description Создаёт Stripe Checkout для своей действующей брони в статусе pending. Цена рассчитывается на сервере в USD.
// @Tags bookings
// @Produce json
// @Security Bearer
// @Param id path string true "ID брони (UUID)"
// @Success 200 {object} checkoutResponse
// @Failure 400,401,403,404,409,500,502,503 {string} string
// @Router /bookings/{id}/pay [post]
func (h *PaymentHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "passenger" {
		http.Error(w, "Оплата доступна пассажирам", http.StatusForbidden)
		return
	}
	if h.stripe == nil {
		http.Error(w, "Тестовая оплата отключена", http.StatusServiceUnavailable)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный ID брони", http.StatusBadRequest)
		return
	}
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	passengerID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Неверный ID пассажира", http.StatusUnauthorized)
		return
	}
	amount, err := h.repo.CheckoutAmount(r.Context(), id, passengerID)
	switch {
	case errors.Is(err, repository.ErrBookingNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, repository.ErrBookingNotPayable):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		log.Printf("Расчёт оплаты брони %s: %v", id, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	case amount < payment.MinimumAmount || amount > payment.MaximumAmount:
		http.Error(w, "Сумма вне диапазона тестовой оплаты", http.StatusConflict)
	default:
		url, err := h.stripe.Checkout(r.Context(), id, amount)
		if err != nil {
			log.Printf("Stripe Checkout для брони %s: %v", id, err)
			http.Error(w, "Не удалось открыть Stripe Checkout", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(checkoutResponse{URL: url})
	}
}

// Webhook godoc
// @Summary Принять событие тестовой оплаты Stripe
// @Description Обрабатывает подписанное событие Checkout и подтверждает действующую бронь или сохраняет возврат в очереди.
// @Tags bookings
// @Accept json
// @Param Stripe-Signature header string true "Подпись Stripe"
// @Success 204 "No Content"
// @Failure 400,500,503 {string} string
// @Router /webhooks/stripe [post]
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.stripe == nil {
		http.Error(w, "Тестовая оплата отключена", http.StatusServiceUnavailable)
		return
	}
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Неверное тело события", http.StatusBadRequest)
		return
	}
	checkout, err := h.stripe.PaidCheckout(payload, r.Header.Get("Stripe-Signature"), time.Now())
	if err != nil {
		http.Error(w, "Неверное событие Stripe", http.StatusBadRequest)
		return
	}
	if checkout == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	id, err := uuid.Parse(checkout.Reference)
	if err != nil {
		http.Error(w, "Неверный ID брони", http.StatusBadRequest)
		return
	}
	_, err = h.repo.ConfirmPayment(r.Context(), id, checkout.Intent, checkout.Amount)
	if err != nil {
		log.Printf("Подтверждение оплаты брони %s: %v", id, err)
		http.Error(w, "Ошибка подтверждения оплаты", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
