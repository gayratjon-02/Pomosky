package weather

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProvider struct {
	calls int
	err   error
}

func (f *fakeProvider) Current(context.Context, float64, float64) (Snapshot, error) {
	f.calls++
	if f.err != nil {
		return Snapshot{}, f.err
	}
	return Snapshot{CloudCoverPct: 40}, nil
}

func TestServiceCachesAndServesStale(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	p := &fakeProvider{}
	s := NewService(p, 10*time.Minute, clock)

	for i := 0; i < 3; i++ {
		if _, err := s.Current(context.Background(), 37.5665, 126.978); err != nil {
			t.Fatal(err)
		}
	}
	if p.calls != 1 {
		t.Fatalf("provider called %d times within TTL, want 1", p.calls)
	}

	now = now.Add(11 * time.Minute)
	p.err = errors.New("upstream down")
	snap, err := s.Current(context.Background(), 37.5665, 126.978)
	if err != nil || snap.CloudCoverPct != 40 {
		t.Fatalf("expected stale snapshot, got %+v, %v", snap, err)
	}
}
