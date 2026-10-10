package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

// RequestID assigns every request an ID (or keeps the one set by the proxy),
// exposes it in the X-Request-ID response header and in the context.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" || len(id) > 64 {
				b := make([]byte, 8)
				_, _ = rand.Read(b)
				id = hex.EncodeToString(b)
			}
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(reqctx.WithRequestID(r.Context(), id)))
		})
	}
}
