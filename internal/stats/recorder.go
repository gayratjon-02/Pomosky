package stats

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// Recorder implements ports.FocusRecorder. RecordFocus only queues the session;
// a background worker writes it, so a slow database never stalls a room.
type Recorder struct {
	repo ports.FocusSessionRepository
	log  *slog.Logger
	ch   chan models.FocusSession
	done chan struct{}
	once sync.Once
}

var _ ports.FocusRecorder = (*Recorder)(nil)

// NewRecorder starts the worker. Call Close to flush and stop it.
func NewRecorder(repo ports.FocusSessionRepository, log *slog.Logger) *Recorder {
	r := &Recorder{repo: repo, log: log, ch: make(chan models.FocusSession, 256), done: make(chan struct{})}
	go r.run()
	return r
}

// RecordFocus implements ports.FocusRecorder. When the queue is full the
// session is dropped and logged instead of blocking.
func (r *Recorder) RecordFocus(s models.FocusSession) {
	select {
	case r.ch <- s:
	default:
		r.log.Error("focus session dropped, queue is full", "room", s.RoomID)
	}
}

// Close writes what is still queued and stops the worker.
func (r *Recorder) Close() {
	r.once.Do(func() { close(r.ch) })
	<-r.done
}

func (r *Recorder) run() {
	defer close(r.done)
	for s := range r.ch {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var err error
		if s.Anonymous {
			err = r.repo.CreateAnonymous(ctx, s.StartedAt, s.Focused)
		} else {
			err = r.repo.Create(ctx, s)
		}
		if err != nil {
			r.log.Error("store focus session", "room", s.RoomID, "err", err)
		}
		cancel()
	}
}
