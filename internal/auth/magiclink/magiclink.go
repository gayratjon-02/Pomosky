// Package magiclink implements passwordless email login.
package magiclink

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// Service issues and verifies one-time login links.
type Service struct {
	mailer ports.Mailer
	ids    ports.IDGenerator
}

// New creates the magic-link Service.
func New(m ports.Mailer, ids ports.IDGenerator) *Service {
	return &Service{mailer: m, ids: ids}
}

// Issue emails a one-time login link to the address. TODO(stage 4).
func (s *Service) Issue(ctx context.Context, email string) error {
	return apperr.ErrNotImplemented
}

// Verify consumes a one-time token and returns the identity. TODO(stage 4).
func (s *Service) Verify(ctx context.Context, token string) (auth.Identity, error) {
	return auth.Identity{}, apperr.ErrNotImplemented
}
