package ports

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// RoomRepository persists room records. Implementations return
// apperr.ErrNotFound when a record does not exist.
type RoomRepository interface {
	Create(ctx context.Context, r models.Room) error
	ByID(ctx context.Context, id room.ID) (models.Room, error)
	SetOwner(ctx context.Context, id room.ID, owner models.UserID) error
	UpdateLocation(ctx context.Context, id room.ID, loc room.Location) error
}

// UserRepository persists optional registered users.
type UserRepository interface {
	Create(ctx context.Context, u models.User) error
	ByID(ctx context.Context, id models.UserID) (models.User, error)
	ByEmail(ctx context.Context, email string) (models.User, error)
	Delete(ctx context.Context, id models.UserID) error
}

// SessionRepository persists login sessions.
type SessionRepository interface {
	Create(ctx context.Context, s models.Session) error
	ByID(ctx context.Context, id string) (models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

// FocusSessionRepository persists finished focus phases.
type FocusSessionRepository interface {
	Create(ctx context.Context, s models.FocusSession) error
	ListByRoom(ctx context.Context, id room.ID, limit int) ([]models.FocusSession, error)
}

// StatsRepository persists per-user statistics.
type StatsRepository interface {
	Get(ctx context.Context, id models.UserID) (models.UserStats, error)
	AddFocus(ctx context.Context, id models.UserID, d time.Duration, day time.Time) error
}
