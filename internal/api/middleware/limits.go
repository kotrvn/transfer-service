package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const JSONBodyLimit = 16 << 10

func LimitJSONBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks/stripe" {
			r.Body = http.MaxBytesReader(w, r.Body, JSONBodyLimit)
		}
		next.ServeHTTP(w, r)
	})
}

// LimitAuth bounds work per peer and across the process. A local reverse proxy
// shares the same peer budget. Client-supplied forwarding headers are ignored.
func LimitAuth(perPeer, total int, interval time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	var resetAt time.Time
	var count int
	var active int
	peers := make(map[string]int)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			mu.Lock()
			now := time.Now()
			if !now.Before(resetAt) {
				resetAt = now.Add(interval)
				count = 0
				clear(peers)
			}
			allowed := active < 4 && count < total && peers[host] < perPeer
			if allowed {
				count++
				peers[host]++
				active++
			}
			retryAfter := max(1, int((resetAt.Sub(now)+time.Second-1)/time.Second))
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "Слишком много запросов. Попробуйте позже.", http.StatusTooManyRequests)
				return
			}
			defer func() {
				mu.Lock()
				active--
				mu.Unlock()
			}()
			next.ServeHTTP(w, r)
		})
	}
}
