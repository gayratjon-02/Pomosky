package middleware

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

type ctxKey int

const (
	userKey ctxKey = iota
	requestIDKey
)

// WithUser stores the logged-in user in the context.
func WithUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// UserFromContext returns the logged-in user, or nil for anonymous visitors.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userKey).(*models.User)
	return u
}

// RequestIDFromContext returns the request ID set by RequestID.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
