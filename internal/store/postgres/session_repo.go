package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// SessionRepo implements ports.SessionRepository.
type SessionRepo struct{ db *DB }

var _ ports.SessionRepository = (*SessionRepo)(nil)

// NewSessionRepo creates the repository.
func NewSessionRepo(db *DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s models.Session) error {
	return mapErr(r.db.q.CreateSession(ctx, sqlcgen.CreateSessionParams{
		ID: s.ID, UserID: string(s.UserID), ExpiresAt: s.ExpiresAt,
	}))
}

func (r *SessionRepo) ByID(ctx context.Context, id string) (models.Session, error) {
	row, err := r.db.q.GetSession(ctx, id)
	if err != nil {
		return models.Session{}, mapErr(err)
	}
	return models.Session{ID: row.ID, UserID: models.UserID(row.UserID), CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt}, nil
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	return mapErr(r.db.q.DeleteSession(ctx, id))
}

func (r *SessionRepo) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	n, err := r.db.q.DeleteExpiredSessions(ctx, now)
	return n, mapErr(err)
}
