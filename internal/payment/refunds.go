package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"transfer-service/internal/models"
)

type stripeRefund struct {
	Status string `json:"status"`
	Amount int64  `json:"amount"`
}

// ReconcileRefund checks Stripe before creating or retrying a refund. A new
// attempt key avoids reusing a cached 500 after Stripe shows no active refund.
func (s *Stripe) ReconcileRefund(ctx context.Context, job models.RefundJob) (string, error) {
	var list struct {
		Data    []stripeRefund `json:"data"`
		HasMore bool           `json:"has_more"`
	}
	query := url.Values{"payment_intent": {job.Intent}, "limit": {"100"}}
	if err := s.request(ctx, http.MethodGet, "/refunds?"+query.Encode(), "", nil, &list); err != nil {
		return "", err
	}
	if list.HasMore {
		return "", errors.New("слишком много возвратов для автоматической сверки")
	}

	var returned int64
	var pending, requiresAction, failed bool
	for _, refund := range list.Data {
		switch refund.Status {
		case "succeeded":
			returned += refund.Amount
		case "requires_action":
			requiresAction = true
		case "pending":
			pending = true
		case "failed", "canceled":
			failed = true
		default:
			return "", fmt.Errorf("неизвестный статус возврата %q", refund.Status)
		}
	}
	switch {
	case returned >= job.Amount:
		return "succeeded", nil
	case requiresAction:
		return "requires_action", nil
	case pending:
		return "pending", nil
	case failed:
		return "failed", nil
	}

	var refund stripeRefund
	key := fmt.Sprintf("late-booking-%s-%d", job.Intent, job.Attempts)
	err := s.post(ctx, "/refunds", key, url.Values{"payment_intent": {job.Intent}}, &refund)
	var failure *stripeError
	if errors.As(err, &failure) && failure.Code == "charge_already_refunded" {
		return "succeeded", nil
	}
	if err != nil {
		return "", err
	}
	switch refund.Status {
	case "succeeded", "pending", "requires_action", "failed":
		return refund.Status, nil
	case "canceled":
		return "failed", nil
	default:
		return "", fmt.Errorf("неизвестный статус возврата %q", refund.Status)
	}
}
