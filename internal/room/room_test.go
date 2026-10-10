package room

import (
	"testing"
	"time"
)

func TestTimerRemaining(t *testing.T) {
	start := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	focus := Timer{Phase: PhaseFocus, StartedAt: start, Duration: 25 * time.Minute}

	cases := []struct {
		name  string
		timer Timer
		now   time.Time
		want  time.Duration
	}{
		{"idle", Timer{Phase: PhaseIdle}, start, 0},
		{"focus mid-way", focus, start.Add(10 * time.Minute), 15 * time.Minute},
		{"focus finished", focus, start.Add(30 * time.Minute), 0},
		{"paused", Timer{Phase: PhasePaused, PausedRemaining: 7 * time.Minute}, start, 7 * time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.timer.Remaining(tc.now); got != tc.want {
				t.Fatalf("Remaining() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIDValid(t *testing.T) {
	cases := map[ID]bool{
		"abcd2345": true,
		"ABCD2345": false,
		"abc":      false,
		"abcd1890": false,
	}
	for id, want := range cases {
		if got := id.Valid(); got != want {
			t.Errorf("ID(%q).Valid() = %v, want %v", id, got, want)
		}
	}
}
