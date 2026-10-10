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
