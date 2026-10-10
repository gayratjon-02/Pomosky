package room

import (
	"strings"
	"testing"
	"time"
)

func TestStartTimer(t *testing.T) {
	s := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute, Label: "write the reducer"})

	if s.Timer.Phase != PhaseFocus || !s.Timer.StartedAt.Equal(t0) || s.Timer.Duration != 25*time.Minute {
		t.Fatalf("unexpected timer after start: %+v", s.Timer)
	}
	if s.Timer.Label != "write the reducer" {
		t.Fatalf("label = %q", s.Timer.Label)
	}
	if got := s.Timer.Remaining(t0.Add(10 * time.Minute)); got != 15*time.Minute {
		t.Fatalf("remaining after 10m = %v, want 15m", got)
	}
	if !s.UpdatedAt.Equal(t0) {
		t.Fatalf("UpdatedAt = %v, want %v", s.UpdatedAt, t0)
	}
}

func TestStartTimerRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		ev   TimerStart
	}{
		{"idle phase", TimerStart{Phase: PhaseIdle, Duration: time.Minute}},
		{"paused phase", TimerStart{Phase: PhasePaused, Duration: time.Minute}},
		{"unknown phase", TimerStart{Phase: "party", Duration: time.Minute}},
		{"zero duration", TimerStart{Phase: PhaseFocus, Duration: 0}},
		{"negative duration", TimerStart{Phase: PhaseFocus, Duration: -time.Minute}},
		{"too short", TimerStart{Phase: PhaseFocus, Duration: time.Millisecond}},
		{"too long", TimerStart{Phase: PhaseBreak, Duration: MaxPhaseDuration + time.Second}},
		{"label too long", TimerStart{Phase: PhaseFocus, Duration: time.Minute, Label: strings.Repeat("x", MaxLabelLength+1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantErr(t, newTestState(), t0, tc.ev, ErrInvalidEvent)
		})
	}
}

func TestStartTimerReplacesRunningPhase(t *testing.T) {
	s := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute})
	s = mustReduce(t, s, t0.Add(5*time.Minute), TimerStart{Phase: PhaseBreak, Duration: 5 * time.Minute})

	if s.Timer.Phase != PhaseBreak || !s.Timer.StartedAt.Equal(t0.Add(5*time.Minute)) {
		t.Fatalf("break did not replace focus: %+v", s.Timer)
	}
}

func TestPauseAndResumeKeepRemainingTime(t *testing.T) {
	for _, phase := range []Phase{PhaseFocus, PhaseBreak} {
		t.Run(string(phase), func(t *testing.T) {
			s := mustReduce(t, newTestState(), t0, TimerStart{Phase: phase, Duration: 25 * time.Minute, Label: "deep work"})

			s = mustReduce(t, s, t0.Add(10*time.Minute), TimerPause{})
			if s.Timer.Phase != PhasePaused || s.Timer.PausedFrom != phase || s.Timer.PausedRemaining != 15*time.Minute {
				t.Fatalf("unexpected paused timer: %+v", s.Timer)
			}
			// Paused time does not run down, however long the pause lasts.
			if got := s.Timer.Remaining(t0.Add(3 * time.Hour)); got != 15*time.Minute {
				t.Fatalf("remaining while paused = %v, want 15m", got)
			}

			resumeAt := t0.Add(20 * time.Minute)
			s = mustReduce(t, s, resumeAt, TimerResume{})
			if s.Timer.Phase != phase || s.Timer.PausedFrom != "" || s.Timer.PausedRemaining != 0 {
				t.Fatalf("unexpected resumed timer: %+v", s.Timer)
			}
			if s.Timer.Label != "deep work" {
				t.Fatalf("label lost on resume: %q", s.Timer.Label)
			}
			if got := s.Timer.Remaining(resumeAt); got != 15*time.Minute {
				t.Fatalf("remaining right after resume = %v, want 15m", got)
			}
			if got := s.Timer.Remaining(resumeAt.Add(15 * time.Minute)); got != 0 {
				t.Fatalf("remaining at the end = %v, want 0", got)
			}
		})
	}
}

func TestInvalidTimerTransitions(t *testing.T) {
	running := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute})
	paused := mustReduce(t, running, t0.Add(time.Minute), TimerPause{})

	cases := []struct {
		name  string
		state State
		now   time.Time
		ev    Event
	}{
		{"pause when idle", newTestState(), t0, TimerPause{}},
		{"pause when already paused", paused, t0.Add(2 * time.Minute), TimerPause{}},
		{"pause after the phase finished", running, t0.Add(30 * time.Minute), TimerPause{}},
		{"resume when idle", newTestState(), t0, TimerResume{}},
		{"resume when running", running, t0.Add(time.Minute), TimerResume{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantErr(t, tc.state, tc.now, tc.ev, ErrInvalidTransition)
		})
	}
}

func TestResetTimer(t *testing.T) {
	running := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute, Label: "x"})
	paused := mustReduce(t, running, t0.Add(time.Minute), TimerPause{})

	for name, from := range map[string]State{"idle": newTestState(), "running": running, "paused": paused} {
		t.Run(name, func(t *testing.T) {
			s := mustReduce(t, from, t0.Add(2*time.Minute), TimerReset{})
			if s.Timer != (Timer{Phase: PhaseIdle}) {
				t.Fatalf("timer after reset = %+v, want idle and empty", s.Timer)
			}
		})
	}
}

func TestReduceDoesNotMutateItsInput(t *testing.T) {
	before := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute})
	snapshot := before.Timer

	_ = mustReduce(t, before, t0.Add(time.Minute), TimerPause{})
	_ = mustReduce(t, before, t0.Add(time.Minute), TimerReset{})

	if before.Timer != snapshot {
		t.Fatalf("Reduce mutated its input: %+v != %+v", before.Timer, snapshot)
	}
}

func TestUnknownEvent(t *testing.T) {
	wantErr(t, newTestState(), t0, unknownEvent{}, ErrUnknownEvent)
}

type unknownEvent struct{}

func (unknownEvent) EventName() string { return "unknown" }
