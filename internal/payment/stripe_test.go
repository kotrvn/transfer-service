package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"transfer-service/internal/models"
)

func TestPaidCheckout(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	stripe, err := NewStripe("sk_test_example", "whsec_example", "http://localhost:3000/bookings/payment-return")
	if err != nil {
		t.Fatal(err)
	}
	paid := `{"type":"checkout.session.completed","livemode":false,"data":{"object":{"client_reference_id":"test","payment_intent":"pi_test","amount_total":2468,"currency":"usd","payment_status":"paid","metadata":{"integration":"transfer-service"}}}}`
	sign := func(payload string, stamp time.Time) string {
		mac := hmac.New(sha256.New, []byte(stripe.webhookSecret))
		fmt.Fprintf(mac, "%d.%s", stamp.Unix(), payload)
		return fmt.Sprintf("t=%d,v1=0000,v1=%x", stamp.Unix(), mac.Sum(nil))
	}
	event, err := stripe.PaidCheckout([]byte(paid), sign(paid, now), now)
	if err != nil || event == nil || event.Amount != 2468 || event.Intent != "pi_test" {
		t.Fatalf("paid checkout: %+v, %v", event, err)
	}
	if _, err := stripe.PaidCheckout([]byte(paid+" "), sign(paid, now), now); err == nil {
		t.Error("accepted altered payload")
	}
	if _, err := stripe.PaidCheckout([]byte(paid), sign(paid, now.Add(-6*time.Minute)), now); err == nil {
		t.Error("accepted expired signature")
	}
	for _, payload := range []string{
		strings.Replace(paid, `"livemode":false`, `"livemode":true`, 1),
		strings.Replace(paid, `"paid"`, `"unpaid"`, 1),
		strings.Replace(paid, `"transfer-service"`, `"other-integration"`, 1),
	} {
		if event, err := stripe.PaidCheckout([]byte(payload), sign(payload, now), now); err != nil || event != nil {
			t.Errorf("ignored checkout: %+v, %v", event, err)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCheckoutAndRefundRequests(t *testing.T) {
	stripe, err := NewStripe("sk_test_example", "whsec_example", "http://localhost:3000/bookings/payment-return")
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	stripe.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if key, _, ok := req.BasicAuth(); !ok || key != "sk_test_example" {
			t.Error("expected Stripe test authentication")
		}
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		switch req.URL.Path {
		case "/v1/checkout/sessions":
			if req.Header.Get("Idempotency-Key") != "booking-"+id.String() ||
				req.Header.Get("Stripe-Version") != "2026-09-30.endive" ||
				req.Form.Get("allowed_payment_method_types[]") != "card" ||
				req.Form.Get("client_reference_id") != id.String() ||
				req.Form.Get("line_items[0][price_data][unit_amount]") != "2468" ||
				req.Form.Get("line_items[0][price_data][currency]") != "usd" {
				t.Errorf("unexpected Checkout request: %v", req.Form)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"url":"https://checkout.stripe.com/c/pay/cs_test"}`))}, nil
		case "/v1/refunds":
			if req.Method == http.MethodGet {
				if req.URL.Query().Get("payment_intent") != "pi_test" || req.Header.Get("Idempotency-Key") != "" {
					t.Error("unexpected Stripe reconciliation request")
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"data":[]}`))}, nil
			}
			if req.Header.Get("Idempotency-Key") != "late-booking-pi_test-2" || req.Form.Get("payment_intent") != "pi_test" {
				t.Errorf("unexpected refund request: %v", req.Form)
			}
			return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"charge_already_refunded"}}`))}, nil
		default:
			t.Fatalf("unexpected Stripe path: %s", req.URL.Path)
			return nil, nil
		}
	})
	if checkout, err := stripe.Checkout(context.Background(), id, 2468); err != nil || checkout == "" {
		t.Fatalf("checkout: %q, %v", checkout, err)
	}
	if status, err := stripe.ReconcileRefund(context.Background(), models.RefundJob{Intent: "pi_test", Amount: 2468, Attempts: 2}); err != nil || status != "succeeded" {
		t.Fatalf("refund retry: %v", err)
	}
}

func TestReconcileExistingRefund(t *testing.T) {
	for _, status := range []string{"pending", "failed", "succeeded", "requires_action"} {
		t.Run(status, func(t *testing.T) {
			stripe, err := NewStripe("sk_test_example", "whsec_example", "http://localhost:3000/bookings/payment-return")
			if err != nil {
				t.Fatal(err)
			}
			stripe.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatal("a known refund must be reconciled without issuing another")
				}
				body := fmt.Sprintf(`{"data":[{"status":%q,"amount":500}]}`, status)
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			got, err := stripe.ReconcileRefund(context.Background(), models.RefundJob{Intent: "pi_test", Amount: 500, Attempts: 2})
			if err != nil || got != status {
				t.Fatalf("refund status %q, error %v", got, err)
			}
		})
	}
}
