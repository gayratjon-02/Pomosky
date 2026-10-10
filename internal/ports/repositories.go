package ports

import (
	"context"
	"net/http"
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
	UpdateNickname(ctx context.Context, id models.UserID, nickname string) error
	SetLeaderboardVisible(ctx context.Context, id models.UserID, visible bool) error
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
	// Leaderboard returns the most focused users who allow it, for sessions that started in [from, to), best first.
	Leaderboard(ctx context.Context, from, to time.Time, limit int) ([]models.LeaderEntry, error)
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

// AdminRepository serves the operator's admin page: totals, daily series and the user list.
type AdminRepository interface {
	Totals(ctx context.Context) (models.AdminTotals, error)
	NewUsersSince(ctx context.Context, since time.Time) (int, error)
	ActiveHostsSince(ctx context.Context, since time.Time) (int, error)
	// Daily returns every day in [from, to) with the admin metrics, days taken in tz.
	Daily(ctx context.Context, from, to time.Time, tz string) ([]models.AdminDay, error)
	// Users lists users, newest first or most focused first, with the total count.
	Users(ctx context.Context, mostFocused bool, limit, offset int) ([]models.AdminUser, int, error)
	DatabaseSize(ctx context.Context) (int64, error)
}

// VisitRepository stores anonymous usage events and reads them back for the admin page.
type VisitRepository interface {
	Create(ctx context.Context, v models.Visit) error
	// Traffic summarises the UTC days in [from, to).
	Traffic(ctx context.Context, from, to time.Time) (models.TrafficOverview, error)
}

// VisitTracker records one usage event of an HTTP request. It never fails the
// request and stores nothing that identifies a person. analytics.Collector implements it.
type VisitTracker interface {
	Track(r *http.Request, kind string)
}

// TrafficReader serves the traffic section of the admin page.
type TrafficReader interface {
	Traffic(ctx context.Context) (models.TrafficOverview, error)
}
