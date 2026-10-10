package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	l := New(1, 2, func() time.Time { return now })

	if !l.Allow("ip") || !l.Allow("ip") {
		t.Fatal("burst of 2 should be allowed")
	}
	if l.Allow("ip") {
		t.Fatal("third immediate event should be rejected")
	}
	if !l.Allow("other") {
		t.Fatal("keys are independent")
	}
	now = now.Add(1100 * time.Millisecond)
	if !l.Allow("ip") {
		t.Fatal("a token should have refilled after ~1s")
	}
}
