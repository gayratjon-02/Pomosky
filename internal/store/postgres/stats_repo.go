package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// StatsRepo implements ports.StatsRepository.
type StatsRepo struct{ db *DB }

var _ ports.StatsRepository = (*StatsRepo)(nil)

// NewStatsRepo creates the repository.
func NewStatsRepo(db *DB) *StatsRepo { return &StatsRepo{db: db} }

// TODO(stage 4): implement with sqlc-generated queries.

func (r *StatsRepo) Get(ctx context.Context, id models.UserID) (models.UserStats, error) {
	return models.UserStats{}, apperr.ErrNotImplemented
}

func (r *StatsRepo) AddFocus(ctx context.Context, id models.UserID, d time.Duration, day time.Time) error {
	return apperr.ErrNotImplemented
}
