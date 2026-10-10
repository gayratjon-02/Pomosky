package middleware

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/ratelimit"
)

// RateLimit rejects requests with 429 once a client exceeds the limiter.
// trustProxy must be true only behind a proxy you control (see httpx.ClientIP).
func RateLimit(l *ratelimit.Limiter, trustProxy bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(httpx.ClientIP(r, trustProxy)) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
