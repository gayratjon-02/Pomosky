package models

import (
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// FocusSession is a focus phase that has ended, written once when it ends.
type FocusSession struct {
	ID          int64
	RoomID      room.ID
	StartedAt   time.Time
	EndedAt     time.Time
	Focused     time.Duration // time actually spent focusing; pauses are not counted
	PeakViewers int
}
