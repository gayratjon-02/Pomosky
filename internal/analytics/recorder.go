// Package analytics records anonymous usage numbers (peak viewers, session
// counts). It never stores personal data.
package analytics

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Recorder receives anonymous metrics.
type Recorder interface {
	RecordPeak(ctx context.Context, id room.ID, viewers int) error
}

// Noop discards everything. It is the default until real metrics are wired.
type Noop struct{}

// RecordPeak implements Recorder.
func (Noop) RecordPeak(context.Context, room.ID, int) error { return nil }
