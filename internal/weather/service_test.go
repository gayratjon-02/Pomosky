package weather

import (
	"context"
	"errors"
	"sync"
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

func TestServiceSharesConcurrentRequests(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	release := make(chan struct{})
	p := &blockingProvider{release: release}
	s := NewService(p, time.Minute, func() time.Time { return now })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Current(context.Background(), 1, 2); err != nil {
				t.Error(err)
			}
		}()
	}
	for p.callCount() == 0 {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond) // let the others queue behind the first call
	close(release)
	wg.Wait()
	if p.callCount() != 1 {
		t.Fatalf("provider called %d times for 8 concurrent requests", p.callCount())
	}
}

func TestServiceCapsUpstreamCalls(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	p := &fakeProvider{}
	s := NewService(p, time.Hour, func() time.Time { return now })
	s.SetMaxFetchesPerMinute(2)

	for i, lat := range []float64{10, 20} {
		if _, err := s.Current(context.Background(), lat, 0); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if _, err := s.Current(context.Background(), 30, 0); !errors.Is(err, ErrBusy) {
		t.Fatalf("over budget and uncached: %v", err)
	}
	if _, err := s.Current(context.Background(), 10, 0); err != nil { // cached places keep working
		t.Fatalf("cached place: %v", err)
	}
	now = now.Add(61 * time.Second) // the budget refills
	if _, err := s.Current(context.Background(), 30, 0); err != nil {
		t.Fatalf("after a minute: %v", err)
	}
}

type blockingProvider struct {
	release chan struct{}
	mu      sync.Mutex
	calls   int
}

func (b *blockingProvider) Current(context.Context, float64, float64) (Snapshot, error) {
	b.mu.Lock()
	b.calls++
	b.mu.Unlock()
	<-b.release
	return Snapshot{CloudCoverPct: 10}, nil
}

func (b *blockingProvider) callCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}
