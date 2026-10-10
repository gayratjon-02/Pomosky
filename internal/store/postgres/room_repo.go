package postgres

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// RoomRepo implements ports.RoomRepository.
type RoomRepo struct{ db *DB }

var _ ports.RoomRepository = (*RoomRepo)(nil)

// NewRoomRepo creates the repository.
func NewRoomRepo(db *DB) *RoomRepo { return &RoomRepo{db: db} }

// TODO(stage 4): implement with sqlc-generated queries.

func (r *RoomRepo) Create(ctx context.Context, rec models.Room) error {
	return apperr.ErrNotImplemented
}

func (r *RoomRepo) ByID(ctx context.Context, id room.ID) (models.Room, error) {
	return models.Room{}, apperr.ErrNotImplemented
}

func (r *RoomRepo) SetOwner(ctx context.Context, id room.ID, owner models.UserID) error {
	return apperr.ErrNotImplemented
}

func (r *RoomRepo) UpdateLocation(ctx context.Context, id room.ID, loc room.Location) error {
	return apperr.ErrNotImplemented
}
