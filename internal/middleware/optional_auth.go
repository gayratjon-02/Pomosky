package middleware

import (
	"context"
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/models"
)

// SessionResolver turns a session cookie value into a user.
// auth.Service implements it.
type SessionResolver interface {
	UserFromSession(ctx context.Context, sessionID string) (*models.User, error)
}

// OptionalAuth attaches the logged-in user to the request context when a valid
// session cookie is present. It NEVER blocks a request: no cookie, an invalid
// session or a nil resolver all simply mean "anonymous".
func OptionalAuth(resolver SessionResolver) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if resolver != nil {
				if c, err := r.Cookie(auth.SessionCookieName); err == nil && c.Value != "" {
					if u, err := resolver.UserFromSession(r.Context(), c.Value); err == nil && u != nil {
						r = r.WithContext(WithUser(r.Context(), u))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
