package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const Currency = "usd"
const MinimumAmount = 50
const MaximumAmount = 99_999_999
const integration = "transfer-service"

type Stripe struct {
	key           string
	webhookSecret string
	returnURL     string
	client        *http.Client
}

func NewStripe(key, webhookSecret, returnURL string) (*Stripe, error) {
	destination, err := url.Parse(returnURL)
	if !strings.HasPrefix(key, "sk_test_") || !strings.HasPrefix(webhookSecret, "whsec_") ||
		err != nil || destination.Hostname() == "" || destination.User != nil || destination.RawQuery != "" || destination.Fragment != "" ||
		(destination.Scheme != "https" && !(destination.Scheme == "http" && destination.Hostname() == "localhost")) {
		return nil, errors.New("Stripe требует тестовый ключ, секрет вебхука и корректный URL возврата")
	}
	return &Stripe{key, webhookSecret, returnURL, &http.Client{Timeout: 5 * time.Second}}, nil
}

func (s *Stripe) Checkout(ctx context.Context, id uuid.UUID, amount int64) (string, error) {
	var session struct {
		URL string `json:"url"`
	}
	err := s.post(ctx, "/checkout/sessions", "booking-"+id.String(), url.Values{
		"mode":                                          {"payment"},
		"allowed_payment_method_types[]":                {"card"},
		"line_items[0][quantity]":                       {"1"},
		"line_items[0][price_data][currency]":           {Currency},
		"line_items[0][price_data][unit_amount]":        {strconv.FormatInt(amount, 10)},
		"line_items[0][price_data][product_data][name]": {"Бронирование поездки (тест)"},
		"metadata[integration]":                         {integration},
		"client_reference_id":                           {id.String()},
		"success_url":                                   {s.returnURL + "?result=success"},
		"cancel_url":                                    {s.returnURL + "?result=cancel"},
	}, &session)
	if err == nil && !strings.HasPrefix(session.URL, "https://checkout.stripe.com/") {
		err = errors.New("Stripe не вернул URL Checkout")
	}
	return session.URL, err
}

type stripeError struct {
	Code   string `json:"code"`
	Status int    `json:"-"`
}

func (e *stripeError) Error() string {
	return fmt.Sprintf("Stripe HTTP %d (%s)", e.Status, e.Code)
}

func (s *Stripe) post(ctx context.Context, path, key string, form url.Values, result any) error {
	return s.request(ctx, http.MethodPost, path, key, strings.NewReader(form.Encode()), result)
}

func (s *Stripe) request(ctx context.Context, method, path, key string, body io.Reader, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, "https://api.stripe.com/v1"+path, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.key, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	req.Header.Set("Stripe-Version", "2026-09-30.endive")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		var body struct {
			Error stripeError `json:"error"`
		}
		_ = decoder.Decode(&body)
		body.Error.Status = resp.StatusCode
		return &body.Error
	}
	return decoder.Decode(result)
}

type CheckoutEvent struct {
	Reference string            `json:"client_reference_id"`
	Intent    string            `json:"payment_intent"`
	Amount    int64             `json:"amount_total"`
	Currency  string            `json:"currency"`
	Status    string            `json:"payment_status"`
	Metadata  map[string]string `json:"metadata"`
}

func (s *Stripe) PaidCheckout(payload []byte, signature string, now time.Time) (*CheckoutEvent, error) {
	if !validSignature(payload, signature, s.webhookSecret, now) {
		return nil, errors.New("неверная подпись Stripe")
	}
	var event struct {
		Type string `json:"type"`
		Live bool   `json:"livemode"`
		Data struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	if event.Live || event.Type != "checkout.session.completed" {
		return nil, nil
	}
	var checkout CheckoutEvent
	if err := json.Unmarshal(event.Data.Object, &checkout); err != nil {
		return nil, err
	}
	if checkout.Status != "paid" || checkout.Metadata["integration"] != integration {
		return nil, nil
	}
	if !strings.HasPrefix(checkout.Intent, "pi_") || checkout.Currency != Currency || checkout.Amount < MinimumAmount {
		return nil, errors.New("неверные данные оплаты Stripe")
	}
	return &checkout, nil
}

func validSignature(payload []byte, signature, secret string, now time.Time) bool {
	var timestamp string
	var signatures []string
	for _, part := range strings.Split(signature, ",") {
		key, value, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}
	unix, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || unix < now.Unix()-300 || unix > now.Unix()+300 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	expected := mac.Sum(nil)
	for _, encoded := range signatures {
		digest, err := hex.DecodeString(encoded)
		if err == nil && hmac.Equal(digest, expected) {
			return true
		}
	}
	return false
}
