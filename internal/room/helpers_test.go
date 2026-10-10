package room

import (
	"errors"
	"testing"
	"time"
)

// t0 is a fixed reference time used by the table tests.
var t0 = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

func newTestState() State { return New("abcd2345", "hash", t0) }

// mustReduce applies ev and fails the test on error.
func mustReduce(t *testing.T, s State, now time.Time, ev Event) State {
	t.Helper()
	next, err := Reduce(s, ev, now)
	if err != nil {
		t.Fatalf("Reduce(%s) failed: %v", ev.EventName(), err)
	}
	return next
}

// wantErr asserts that Reduce fails with target and leaves the state unchanged.
func wantErr(t *testing.T, s State, now time.Time, ev Event, target error) {
	t.Helper()
	next, err := Reduce(s, ev, now)
	if !errors.Is(err, target) {
		t.Fatalf("Reduce(%s) error = %v, want %v", ev.EventName(), err, target)
	}
	if next.Timer != s.Timer || next.Viewers != s.Viewers || next.Location != s.Location {
		t.Fatalf("Reduce(%s) changed the state despite the error", ev.EventName())
	}
}
