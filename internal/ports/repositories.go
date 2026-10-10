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
	// DeleteStale removes rooms created before the cutoff that have no focus history.
	DeleteStale(ctx context.Context, before time.Time) (int64, error)
}

// UserRepository persists optional registered users.
type UserRepository interface {
	Create(ctx context.Context, u models.User) error
	ByID(ctx context.Context, id models.UserID) (models.User, error)
	ByEmail(ctx context.Context, email string) (models.User, error)
	ByIdentity(ctx context.Context, provider, subject string) (models.User, error)
	Delete(ctx context.Context, id models.UserID) error
	UpdateAvatar(ctx context.Context, id models.UserID, url string) error
}

// SessionRepository persists login sessions.
type SessionRepository interface {
	Create(ctx context.Context, s models.Session) error
	ByID(ctx context.Context, id string) (models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

// FocusSessionRepository persists finished focus phases and reads them back
// as per-day totals for the statistics view.
type FocusSessionRepository interface {
	Create(ctx context.Context, s models.FocusSession) error
	// CreateAnonymous stores focus time of a room without an owner: when it started and how long, nothing else.
	CreateAnonymous(ctx context.Context, startedAt time.Time, focused time.Duration) error
	// CommunityBetween sums everyone's focus time for sessions that started in [from, to).
	CommunityBetween(ctx context.Context, from, to time.Time) (models.Period, error)
	// CommunityDaily returns one entry per day with any focus time, for everyone, days taken in tz.
	CommunityDaily(ctx context.Context, from, to time.Time, tz string) ([]models.DayStat, error)
	// DailyTotals returns one entry per day that has focus time, for sessions
	// that started in [from, to). Days are taken in the time zone tz (IANA name).
	DailyTotals(ctx context.Context, owner models.UserID, from, to time.Time, tz string) ([]models.DayStat, error)
	// ActiveDays returns the dates (YYYY-MM-DD) with any focus time since a moment, newest first.
	ActiveDays(ctx context.Context, owner models.UserID, since time.Time, tz string) ([]string, error)
}

// LoginTokenRepository stores one-time email sign-in tokens (hashed).
type LoginTokenRepository interface {
	Create(ctx context.Context, tokenHash, email string, expires time.Time) error
	// Consume deletes the token and returns its email. It returns
	// apperr.ErrNotFound when the token is unknown, used or expired.
	Consume(ctx context.Context, tokenHash string, now time.Time) (string, error)
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}
