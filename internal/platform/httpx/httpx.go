// Package httpx contains small HTTP helpers shared by controllers and middleware.
package httpx

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
)

// WriteJSON writes v as a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// StatusFor maps an application error to an HTTP status code.
func StatusFor(err error) int {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperr.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, apperr.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperr.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, apperr.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, apperr.ErrNotImplemented):
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// WriteError writes err as a JSON error response.
func WriteError(w http.ResponseWriter, err error) {
	status := StatusFor(err)
	msg := http.StatusText(status)
	if status != http.StatusInternalServerError {
		msg = err.Error()
	}
	WriteJSON(w, status, map[string]string{"error": msg})
}

// NotImplemented answers 501 for endpoints that are still stubs.
func NotImplemented(w http.ResponseWriter) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// ClientIP returns the client address. X-Forwarded-For is honoured only when
// trustProxy is true, i.e. when the server runs behind a proxy you control
// (Caddy). Otherwise the header can be spoofed by clients.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			return strings.TrimSpace(first)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
