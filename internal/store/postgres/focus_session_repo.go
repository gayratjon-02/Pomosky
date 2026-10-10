package postgres

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// FocusSessionRepo implements ports.FocusSessionRepository.
type FocusSessionRepo struct{ db *DB }

var _ ports.FocusSessionRepository = (*FocusSessionRepo)(nil)

// NewFocusSessionRepo creates the repository.
func NewFocusSessionRepo(db *DB) *FocusSessionRepo { return &FocusSessionRepo{db: db} }

// TODO(stage 4): implement with sqlc-generated queries.

func (r *FocusSessionRepo) Create(ctx context.Context, s models.FocusSession) error {
	return apperr.ErrNotImplemented
}

func (r *FocusSessionRepo) ListByRoom(ctx context.Context, id room.ID, limit int) ([]models.FocusSession, error) {
	return nil, apperr.ErrNotImplemented
}
