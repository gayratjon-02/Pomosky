package room

import (
	"time"
	"unicode/utf8"
)

// Limits for a single timer phase and its label.
const (
	MinPhaseDuration = time.Second
	MaxPhaseDuration = 12 * time.Hour
	MaxLabelLength   = 80 // runes
)

func validPhaseDuration(d time.Duration) bool {
	return d >= MinPhaseDuration && d <= MaxPhaseDuration
}

func validLabel(label string) bool {
	return utf8.RuneCountInString(label) <= MaxLabelLength
}

// canStart reports whether a new phase may begin: the timer is idle, or the
// running phase has run to its end. A running or paused phase must be paused,
// resumed or reset first, so pressing "start" twice can never restart the clock.
func canStart(t Timer, now time.Time) bool {
	switch t.Phase {
	case PhaseIdle:
		return true
	case PhaseFocus, PhaseBreak:
		return t.Remaining(now) == 0
	default: // paused
		return false
	}
}

// startTimer starts a new focus or break phase when the timer is idle or the
// previous phase has ended.
func startTimer(s State, e TimerStart, now time.Time) (State, error) {
	if e.Phase != PhaseFocus && e.Phase != PhaseBreak {
		return s, ErrInvalidEvent
	}
	if !validPhaseDuration(e.Duration) || !validLabel(e.Label) {
		return s, ErrInvalidEvent
	}
	if !canStart(s.Timer, now) {
		return s, ErrInvalidTransition
	}
	s.Timer = Timer{
		Phase:     e.Phase,
		StartedAt: now,
		Duration:  e.Duration,
		Label:     e.Label,
	}
	return s, nil
}

// pauseTimer pauses a running phase that has not finished yet.
func pauseTimer(s State, now time.Time) (State, error) {
	t := s.Timer
	if t.Phase != PhaseFocus && t.Phase != PhaseBreak {
		return s, ErrInvalidTransition
	}
	left := t.Remaining(now)
	if left <= 0 {
		return s, ErrInvalidTransition // already finished, nothing to pause
	}
	s.Timer = Timer{
		Phase:           PhasePaused,
		PausedFrom:      t.Phase,
		Duration:        t.Duration,
		PausedRemaining: left,
		Label:           t.Label,
	}
	return s, nil
}

// resumeTimer continues a paused phase with exactly the remaining time it had.
func resumeTimer(s State, now time.Time) (State, error) {
	t := s.Timer
	if t.Phase != PhasePaused {
		return s, ErrInvalidTransition
	}
	elapsed := t.Duration - t.PausedRemaining
	s.Timer = Timer{
		Phase:     t.PausedFrom,
		StartedAt: now.Add(-elapsed),
		Duration:  t.Duration,
		Label:     t.Label,
	}
	return s, nil
}

// resetTimer returns the timer to idle and clears the label.
func resetTimer(s State) (State, error) {
	s.Timer = Timer{Phase: PhaseIdle}
	return s, nil
}
