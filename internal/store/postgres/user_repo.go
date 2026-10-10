package postgres

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// UserRepo implements ports.UserRepository.
type UserRepo struct{ db *DB }

var _ ports.UserRepository = (*UserRepo)(nil)

// NewUserRepo creates the repository.
func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

// TODO(stage 4): implement with sqlc-generated queries.

func (r *UserRepo) Create(ctx context.Context, u models.User) error {
	return apperr.ErrNotImplemented
}

func (r *UserRepo) ByID(ctx context.Context, id models.UserID) (models.User, error) {
	return models.User{}, apperr.ErrNotImplemented
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (models.User, error) {
	return models.User{}, apperr.ErrNotImplemented
}

func (r *UserRepo) Delete(ctx context.Context, id models.UserID) error {
	return apperr.ErrNotImplemented
}
