package auth

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// Service resolves sessions and signs users in and out.
type Service struct {
	users    ports.UserRepository
	sessions ports.SessionRepository
	clock    ports.Clock
	ids      ports.IDGenerator
}

// NewService creates the auth Service.
func NewService(u ports.UserRepository, s ports.SessionRepository, c ports.Clock, ids ports.IDGenerator) *Service {
	return &Service{users: u, sessions: s, clock: c, ids: ids}
}

// UserFromSession returns the user of a session cookie value, or an error if
// the session is unknown or expired. Callers treat any error as "anonymous".
//
// TODO(stage 4): implement.
func (s *Service) UserFromSession(ctx context.Context, sessionID string) (*models.User, error) {
	return nil, apperr.ErrNotImplemented
}

// SignIn creates a session for the identity, creating the user if needed.
//
// TODO(stage 4): implement.
func (s *Service) SignIn(ctx context.Context, id Identity) (models.Session, error) {
	return models.Session{}, apperr.ErrNotImplemented
}

// SignOut deletes a session.
//
// TODO(stage 4): implement.
func (s *Service) SignOut(ctx context.Context, sessionID string) error {
	return apperr.ErrNotImplemented
}
