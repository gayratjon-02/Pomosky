package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// LoginTokenRepo implements ports.LoginTokenRepository.
type LoginTokenRepo struct{ db *DB }

var _ ports.LoginTokenRepository = (*LoginTokenRepo)(nil)

// NewLoginTokenRepo creates the repository.
func NewLoginTokenRepo(db *DB) *LoginTokenRepo { return &LoginTokenRepo{db: db} }

func (r *LoginTokenRepo) Create(ctx context.Context, tokenHash, email string, expires time.Time) error {
	return mapErr(r.db.q.CreateLoginToken(ctx, sqlcgen.CreateLoginTokenParams{TokenHash: tokenHash, Email: email, ExpiresAt: expires}))
}

func (r *LoginTokenRepo) Consume(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	email, err := r.db.q.ConsumeLoginToken(ctx, sqlcgen.ConsumeLoginTokenParams{TokenHash: tokenHash, ExpiresAt: now})
	return email, mapErr(err)
}
