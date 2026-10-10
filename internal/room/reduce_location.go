package room

import (
	"math"
	"strings"
	"time"
)

const maxTZLength = 64

// round2 rounds to two decimals (~1 km). The core enforces the privacy rule
// that exact coordinates are never kept, whatever the caller sends.
func round2(v float64) float64 {
	r := math.Round(v*100) / 100
	if r == 0 {
		return 0 // avoid negative zero
	}
	return r
}

// setLocation validates the location and stores it rounded.
func setLocation(s State, e LocationSet) (State, error) {
	l := e.Location
	// Written as a positive range check so NaN and Inf are rejected too.
	if !(l.LatRounded >= -90 && l.LatRounded <= 90) || !(l.LonRounded >= -180 && l.LonRounded <= 180) {
		return s, ErrInvalidEvent
	}
	city := strings.TrimSpace(l.City)
	tz := strings.TrimSpace(l.TZ)
	if !validLabel(city) || len(tz) > maxTZLength {
		return s, ErrInvalidEvent
	}
	s.Location = Location{
		City:       city,
		LatRounded: round2(l.LatRounded),
		LonRounded: round2(l.LonRounded),
		TZ:         tz,
	}
	return s, nil
}

// hostRecovered restores the timer after a server restart, using the timer the
// reconnecting host still has. It only applies to a room whose timer is idle:
// if the server kept its own timer, the host's copy is ignored. Only the
// documented fields are copied, so nothing else can be smuggled in.
func hostRecovered(s State, e HostRecovered, now time.Time) (State, error) {
	if s.Timer.Phase != PhaseIdle {
		return s, nil
	}
	t := e.Timer
	if !validLabel(t.Label) {
		return s, ErrInvalidEvent
	}
	switch t.Phase {
	case PhaseIdle:
		return s, nil
	case PhaseFocus, PhaseBreak:
		if !validPhaseDuration(t.Duration) || t.StartedAt.IsZero() || t.StartedAt.After(now) {
			return s, ErrInvalidEvent
		}
		s.Timer = Timer{Phase: t.Phase, StartedAt: t.StartedAt, Duration: t.Duration, Label: t.Label}
	case PhasePaused:
		if t.PausedFrom != PhaseFocus && t.PausedFrom != PhaseBreak {
			return s, ErrInvalidEvent
		}
		if !validPhaseDuration(t.Duration) || t.PausedRemaining <= 0 || t.PausedRemaining > t.Duration {
			return s, ErrInvalidEvent
		}
		s.Timer = Timer{
			Phase:           PhasePaused,
			PausedFrom:      t.PausedFrom,
			Duration:        t.Duration,
			PausedRemaining: t.PausedRemaining,
			Label:           t.Label,
		}
	default:
		return s, ErrInvalidEvent
	}
	return s, nil
}
