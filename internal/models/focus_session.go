package models

import (
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// FocusSession is a finished focus phase, written once when it ends.
type FocusSession struct {
	ID          int64
	RoomID      room.ID
	StartedAt   time.Time
	EndedAt     time.Time
	PeakViewers int
}
