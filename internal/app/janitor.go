package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// staleRoomAge is how old a stored room without any focus history must be
// before the janitor deletes it.
const staleRoomAge = 7 * 24 * time.Hour

// janitor removes expired rows (sessions, sign-in links) and stale empty rooms.
type janitor struct {
	log      *slog.Logger
	clock    ports.Clock
	sessions ports.SessionRepository
	tokens   ports.LoginTokenRepository
	rooms    interface {
		DeleteStale(ctx context.Context, before time.Time) (int64, error)
	}
}

func (j janitor) run(ctx context.Context, interval time.Duration) {
	j.once(ctx)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			j.once(ctx)
		}
	}
}

func (j janitor) once(ctx context.Context) {
	now := j.clock.Now()
	if n, err := j.sessions.DeleteExpired(ctx, now); err != nil {
		j.log.Error("cleanup: sessions", "err", err)
	} else if n > 0 {
		j.log.Info("cleanup: expired sessions removed", "count", n)
	}
	if n, err := j.tokens.DeleteExpired(ctx, now); err != nil {
		j.log.Error("cleanup: login tokens", "err", err)
	} else if n > 0 {
		j.log.Info("cleanup: expired sign-in links removed", "count", n)
	}
	if n, err := j.rooms.DeleteStale(ctx, now.Add(-staleRoomAge)); err != nil {
		j.log.Error("cleanup: rooms", "err", err)
	} else if n > 0 {
		j.log.Info("cleanup: stale rooms removed", "count", n)
	}
}
