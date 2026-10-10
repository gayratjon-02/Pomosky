// Package reqctx carries per-request values (logged-in user, CSRF token,
// request ID) in the request context. Middleware sets them, controllers read them.
package reqctx

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

type key int

const (
	userKey key = iota
	csrfKey
	requestIDKey
)

// WithUser stores the logged-in user.
func WithUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// User returns the logged-in user, or nil for anonymous visitors.
func User(ctx context.Context) *models.User {
	u, _ := ctx.Value(userKey).(*models.User)
	return u
}

// WithCSRF stores the CSRF token of the request.
func WithCSRF(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfKey, token)
}

// CSRF returns the CSRF token to embed in forms.
func CSRF(ctx context.Context) string {
	t, _ := ctx.Value(csrfKey).(string)
	return t
}

// WithRequestID stores the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request ID.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
