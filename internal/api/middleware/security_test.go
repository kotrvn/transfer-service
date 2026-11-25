package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWTAuthRejectsInvalidClaims(t *testing.T) {
	secret := strings.Repeat("a", 32)
	handler := JWTAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, test := range []struct {
		name   string
		method jwt.SigningMethod
		sub    string
		role   string
		exp    any
		want   int
	}{
		{"valid", jwt.SigningMethodHS256, uuid.NewString(), "passenger", time.Now().Add(time.Hour).Unix(), 204},
		{"wrong algorithm", jwt.SigningMethodHS384, uuid.NewString(), "passenger", time.Now().Add(time.Hour).Unix(), 401},
		{"missing expiration", jwt.SigningMethodHS256, uuid.NewString(), "passenger", nil, 401},
		{"invalid role", jwt.SigningMethodHS256, uuid.NewString(), "other", time.Now().Add(time.Hour).Unix(), 401},
		{"invalid subject", jwt.SigningMethodHS256, "invalid", "driver", time.Now().Add(time.Hour).Unix(), 401},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := jwt.MapClaims{"sub": test.sub, "role": test.role}
			if test.exp != nil {
				claims["exp"] = test.exp
			}
			raw, err := jwt.NewWithClaims(test.method, claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+raw)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != test.want {
				t.Fatalf("got %d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestLimitAuthIgnoresForwardingHeaders(t *testing.T) {
	handler := LimitAuth(1, 2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, test := range []struct {
		peer, forwarded string
		want            int
	}{
		{"192.0.2.1:1111", "192.0.2.10", 204},
		{"192.0.2.1:2222", "192.0.2.20", 429},
		{"192.0.2.2:1111", "192.0.2.10", 204},
		{"192.0.2.3:1111", "192.0.2.10", 429},
	} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = test.peer
		req.Header.Set("X-Forwarded-For", test.forwarded)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != test.want {
			t.Fatalf("got %d, want %d", response.Code, test.want)
		}
	}
}
