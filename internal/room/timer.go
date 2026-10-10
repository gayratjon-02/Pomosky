package room

import "time"

// Timer is the server-authoritative pomodoro timer.
//
// Clients never tick over the network: they compute the remaining time from
// StartedAt and Duration using a server clock offset (see web/src/core).
type Timer struct {
	Phase           Phase
	StartedAt       time.Time     // server time at which the running phase started
	Duration        time.Duration // length of the current phase
	PausedRemaining time.Duration // remaining time while Phase == PhasePaused
	PausedFrom      Phase         // phase to return to on resume (only while paused)
	Label           string        // what the host is working on
}

// Remaining returns the time left on the timer at now.
func (t Timer) Remaining(now time.Time) time.Duration {
	switch t.Phase {
	case PhasePaused:
		return t.PausedRemaining
	case PhaseFocus, PhaseBreak:
		left := t.StartedAt.Add(t.Duration).Sub(now)
		if left < 0 {
			return 0
		}
		return left
	default:
		return 0
	}
}

// FocusedTime returns how long the timer has been focusing at now, pauses not
// counted, for a focus phase that is running or paused. The second result is
// false for idle timers and for break phases.
func FocusedTime(t Timer, now time.Time) (time.Duration, bool) {
	var left time.Duration
	switch {
	case t.Phase == PhaseFocus:
		left = t.Remaining(now)
	case t.Phase == PhasePaused && t.PausedFrom == PhaseFocus:
		left = t.PausedRemaining
	default:
		return 0, false
	}
	done := t.Duration - left
	if done < 0 {
		done = 0
	}
	return done, true
}
