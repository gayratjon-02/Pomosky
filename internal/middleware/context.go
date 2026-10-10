package middleware

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

// WithUser stores the logged-in user in the context.
func WithUser(ctx context.Context, u *models.User) context.Context { return reqctx.WithUser(ctx, u) }

// UserFromContext returns the logged-in user, or nil for anonymous visitors.
func UserFromContext(ctx context.Context) *models.User { return reqctx.User(ctx) }

// RequestIDFromContext returns the request ID set by RequestID.
func RequestIDFromContext(ctx context.Context) string { return reqctx.RequestID(ctx) }
