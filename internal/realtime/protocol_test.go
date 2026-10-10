package realtime

import (
	"errors"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

func TestTimerViewRoundTrip(t *testing.T) {
	start := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	cases := map[string]room.Timer{
		"idle":    {Phase: room.PhaseIdle},
		"running": {Phase: room.PhaseFocus, StartedAt: start, Duration: 25 * time.Minute, Label: "api"},
		"paused": {
			Phase: room.PhasePaused, PausedFrom: room.PhaseBreak,
			Duration: 5 * time.Minute, PausedRemaining: 90 * time.Second,
		},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := FromTimerView(ToTimerView(in))
			if err != nil {
				t.Fatal(err)
			}
			if out != in {
				t.Fatalf("round trip changed the timer:\n got %+v\nwant %+v", out, in)
			}
		})
	}
}

func TestFromTimerViewRejectsOutOfRangeValues(t *testing.T) {
	bad := []TimerView{
		{Phase: "focus", DurationMS: -1},
		{Phase: "focus", DurationMS: maxPhaseMS + 1},
		{Phase: "paused", PausedRemainingMS: -5},
		{Phase: "paused", PausedRemainingMS: maxPhaseMS + 1},
	}
	for _, v := range bad {
		if _, err := FromTimerView(v); !errors.Is(err, ErrBadTimer) {
			t.Errorf("FromTimerView(%+v) error = %v, want ErrBadTimer", v, err)
		}
	}
}

func TestToTimerStart(t *testing.T) {
	ev, err := ToTimerStart(TimerStartData{Phase: "focus", DurationMS: 1_500_000, Label: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Phase != room.PhaseFocus || ev.Duration != 25*time.Minute || ev.Label != "x" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if _, err := ToTimerStart(TimerStartData{Phase: "focus", DurationMS: -1}); !errors.Is(err, ErrBadTimer) {
		t.Fatalf("negative duration error = %v", err)
	}
	if _, err := ToTimerStart(TimerStartData{Phase: "focus", DurationMS: 9_223_372_036_854_775_807}); !errors.Is(err, ErrBadTimer) {
		t.Fatalf("overflowing duration error = %v", err)
	}
}
