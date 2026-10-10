package middleware

import (
	"log/slog"
	"net/http"
)

// Recover turns panics into 500 responses and logs them.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic in handler", "panic", rec, "path", r.URL.Path,
						"request_id", RequestIDFromContext(r.Context()))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
