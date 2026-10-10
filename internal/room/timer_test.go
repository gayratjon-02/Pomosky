package room

import (
	"testing"
	"time"
)

func TestFocusedTime(t *testing.T) {
	start := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	run := Timer{Phase: PhaseFocus, StartedAt: start, Duration: 25 * time.Minute}
	if d, ok := FocusedTime(run, start.Add(10*time.Minute)); !ok || d != 10*time.Minute {
		t.Fatalf("running: %v %v", d, ok)
	}
	if d, ok := FocusedTime(run, start.Add(time.Hour)); !ok || d != 25*time.Minute {
		t.Fatalf("finished: %v %v", d, ok)
	}
	paused := Timer{Phase: PhasePaused, PausedFrom: PhaseFocus, Duration: 25 * time.Minute, PausedRemaining: 20 * time.Minute}
	if d, ok := FocusedTime(paused, start); !ok || d != 5*time.Minute {
		t.Fatalf("paused: %v %v", d, ok)
	}
	pausedBreak := Timer{Phase: PhasePaused, PausedFrom: PhaseBreak, Duration: 5 * time.Minute, PausedRemaining: time.Minute}
	if _, ok := FocusedTime(pausedBreak, start); ok {
		t.Fatal("a paused break is not focus")
	}
	if _, ok := FocusedTime(Timer{Phase: PhaseBreak, StartedAt: start, Duration: time.Minute}, start); ok {
		t.Fatal("a break is not focus")
	}
	if _, ok := FocusedTime(Timer{Phase: PhaseIdle}, start); ok {
		t.Fatal("idle is not focus")
	}
}
