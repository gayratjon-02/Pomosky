package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// SessionRepo implements ports.SessionRepository.
type SessionRepo struct{ db *DB }

var _ ports.SessionRepository = (*SessionRepo)(nil)

// NewSessionRepo creates the repository.
func NewSessionRepo(db *DB) *SessionRepo { return &SessionRepo{db: db} }

// TODO(stage 4): implement with sqlc-generated queries.

func (r *SessionRepo) Create(ctx context.Context, s models.Session) error {
	return apperr.ErrNotImplemented
}

func (r *SessionRepo) ByID(ctx context.Context, id string) (models.Session, error) {
	return models.Session{}, apperr.ErrNotImplemented
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	return apperr.ErrNotImplemented
}

func (r *SessionRepo) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	return 0, apperr.ErrNotImplemented
}
