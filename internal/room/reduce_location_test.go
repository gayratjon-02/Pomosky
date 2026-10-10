package room

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestSetLocationRoundsAndTrims(t *testing.T) {
	cases := []struct {
		name string
		in   Location
		want Location
	}{
		{
			"seoul is rounded to two decimals",
			Location{City: "  Seoul ", LatRounded: 37.56649, LonRounded: 126.97796, TZ: " Asia/Seoul "},
			Location{City: "Seoul", LatRounded: 37.57, LonRounded: 126.98, TZ: "Asia/Seoul"},
		},
		{
			"negative coordinates",
			Location{City: "Buenos Aires", LatRounded: -34.60372, LonRounded: -58.38159},
			Location{City: "Buenos Aires", LatRounded: -34.6, LonRounded: -58.38},
		},
		{
			"negative zero becomes zero",
			Location{LatRounded: -0.004, LonRounded: 0.004},
			Location{},
		},
		{
			"boundaries are valid",
			Location{LatRounded: 90, LonRounded: -180},
			Location{LatRounded: 90, LonRounded: -180},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := mustReduce(t, newTestState(), t0, LocationSet{Location: tc.in})
			if s.Location != tc.want {
				t.Fatalf("Location = %+v, want %+v", s.Location, tc.want)
			}
		})
	}
}

func TestSetLocationRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		in   Location
	}{
		{"latitude too high", Location{LatRounded: 90.01}},
		{"latitude too low", Location{LatRounded: -90.01}},
		{"longitude too high", Location{LonRounded: 180.01}},
		{"longitude too low", Location{LonRounded: -180.01}},
		{"NaN latitude", Location{LatRounded: math.NaN()}},
		{"infinite longitude", Location{LonRounded: math.Inf(1)}},
		{"city too long", Location{City: strings.Repeat("c", MaxLabelLength+1)}},
		{"time zone too long", Location{TZ: strings.Repeat("z", maxTZLength+1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantErr(t, newTestState(), t0, LocationSet{Location: tc.in}, ErrInvalidEvent)
		})
	}
}

func TestHostRecoveredRestoresTimer(t *testing.T) {
	startedAt := t0.Add(-10 * time.Minute)

	t.Run("running focus", func(t *testing.T) {
		s := mustReduce(t, newTestState(), t0, HostRecovered{Timer: Timer{
			Phase: PhaseFocus, StartedAt: startedAt, Duration: 25 * time.Minute, Label: "api",
			PausedRemaining: time.Hour, PausedFrom: PhaseBreak, // must be dropped
		}})
		if s.Timer.Phase != PhaseFocus || s.Timer.Label != "api" {
			t.Fatalf("unexpected timer: %+v", s.Timer)
		}
		if got := s.Timer.Remaining(t0); got != 15*time.Minute {
			t.Fatalf("remaining = %v, want 15m", got)
		}
		if s.Timer.PausedRemaining != 0 || s.Timer.PausedFrom != "" {
			t.Fatalf("unrelated fields were copied: %+v", s.Timer)
		}
	})

	t.Run("paused", func(t *testing.T) {
		s := mustReduce(t, newTestState(), t0, HostRecovered{Timer: Timer{
			Phase: PhasePaused, PausedFrom: PhaseBreak, Duration: 5 * time.Minute, PausedRemaining: 2 * time.Minute,
		}})
		if s.Timer.Phase != PhasePaused || s.Timer.PausedFrom != PhaseBreak || s.Timer.Remaining(t0) != 2*time.Minute {
			t.Fatalf("unexpected timer: %+v", s.Timer)
		}
	})

	t.Run("already finished is still restored", func(t *testing.T) {
		s := mustReduce(t, newTestState(), t0, HostRecovered{Timer: Timer{
			Phase: PhaseFocus, StartedAt: t0.Add(-time.Hour), Duration: 25 * time.Minute,
		}})
		if s.Timer.Phase != PhaseFocus || s.Timer.Remaining(t0) != 0 {
			t.Fatalf("unexpected timer: %+v", s.Timer)
		}
	})

	t.Run("idle recovery changes nothing", func(t *testing.T) {
		s := mustReduce(t, newTestState(), t0, HostRecovered{Timer: Timer{Phase: PhaseIdle}})
		if s.Timer != (Timer{Phase: PhaseIdle}) {
			t.Fatalf("unexpected timer: %+v", s.Timer)
		}
	})
}

func TestHostRecoveredIsIgnoredWhenServerHasATimer(t *testing.T) {
	running := mustReduce(t, newTestState(), t0, TimerStart{Phase: PhaseFocus, Duration: 25 * time.Minute, Label: "server copy"})

	s := mustReduce(t, running, t0.Add(time.Minute), HostRecovered{Timer: Timer{
		Phase: PhaseBreak, StartedAt: t0, Duration: time.Minute, Label: "host copy",
	}})

	if s.Timer != running.Timer {
		t.Fatalf("server timer was overwritten: %+v", s.Timer)
	}
}

func TestHostRecoveredRejectsInvalidTimers(t *testing.T) {
	good := Timer{Phase: PhaseFocus, StartedAt: t0.Add(-time.Minute), Duration: 25 * time.Minute}
	paused := Timer{Phase: PhasePaused, PausedFrom: PhaseFocus, Duration: 25 * time.Minute, PausedRemaining: time.Minute}

	mut := func(base Timer, f func(*Timer)) Timer { f(&base); return base }

	cases := []struct {
		name  string
		timer Timer
	}{
		{"unknown phase", mut(good, func(x *Timer) { x.Phase = "party" })},
		{"zero start time", mut(good, func(x *Timer) { x.StartedAt = time.Time{} })},
		{"start in the future", mut(good, func(x *Timer) { x.StartedAt = t0.Add(time.Second) })},
		{"duration too short", mut(good, func(x *Timer) { x.Duration = time.Millisecond })},
		{"duration too long", mut(good, func(x *Timer) { x.Duration = MaxPhaseDuration + time.Second })},
		{"label too long", mut(good, func(x *Timer) { x.Label = strings.Repeat("x", MaxLabelLength+1) })},
		{"paused without a phase to resume", mut(paused, func(x *Timer) { x.PausedFrom = PhaseIdle })},
		{"paused with nothing remaining", mut(paused, func(x *Timer) { x.PausedRemaining = 0 })},
		{"paused with more than the duration", mut(paused, func(x *Timer) { x.PausedRemaining = x.Duration + time.Second })},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantErr(t, newTestState(), t0, HostRecovered{Timer: tc.timer}, ErrInvalidEvent)
		})
	}
}
