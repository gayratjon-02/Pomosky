package weather

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// ErrBusy is returned when the upstream request budget is used up and there is
// no cached snapshot to fall back on.
var ErrBusy = errors.New("weather: too many upstream requests")

// DefaultMaxFetchesPerMinute keeps the service far below Open-Meteo's free
// limit (600 calls per minute) even if many rooms pick different places.
const DefaultMaxFetchesPerMinute = 60

const upstreamTimeout = 10 * time.Second

// Provider fetches a fresh snapshot from an upstream weather API.
type Provider interface {
	Current(ctx context.Context, lat, lon float64) (Snapshot, error)
}

// Service caches snapshots so that every viewer of a room shares one
// upstream request per TTL. Concurrent requests for the same place share one
// upstream call, the number of upstream calls per minute is capped, and if the
// upstream fails a stale snapshot is served.
type Service struct {
	provider  Provider
	ttl       time.Duration
	now       func() time.Time
	maxPerMin int

	calls, hits, errs, busy atomic.Int64 // counters for the admin page

	mu       sync.Mutex
	cache    map[cacheKey]cacheEntry
	inflight map[cacheKey]*call
	fetches  []time.Time // start times of the upstream calls in the last minute
}

type cacheKey struct{ lat, lon float64 }

type cacheEntry struct {
	snap Snapshot
	at   time.Time
}

type call struct {
	done chan struct{}
	snap Snapshot
	err  error
}

// NewService creates a caching Service.
func NewService(p Provider, ttl time.Duration, now func() time.Time) *Service {
	return &Service{
		provider: p, ttl: ttl, now: now, maxPerMin: DefaultMaxFetchesPerMinute,
		cache: map[cacheKey]cacheEntry{}, inflight: map[cacheKey]*call{},
	}
}

// SetMaxFetchesPerMinute changes the upstream budget (call before serving).
func (s *Service) SetMaxFetchesPerMinute(n int) { s.maxPerMin = n }

// Current returns the (possibly cached) snapshot for the given location.
func (s *Service) Current(ctx context.Context, lat, lon float64) (Snapshot, error) {
	key := cacheKey{lat: round2(lat), lon: round2(lon)}
	now := s.now()

	s.mu.Lock()
	entry, cached := s.cache[key]
	if cached && now.Sub(entry.at) < s.ttl {
		s.mu.Unlock()
		s.hits.Add(1)
		return entry.snap, nil
	}
	if c, ok := s.inflight[key]; ok { // someone is already asking for this place
		s.mu.Unlock()
		return s.wait(ctx, c, entry, cached)
	}
	if !s.allowFetchLocked(now) {
		s.mu.Unlock()
		s.busy.Add(1)
		if cached {
			return entry.snap, nil
		}
		return Snapshot{}, ErrBusy
	}
	c := &call{done: make(chan struct{})}
	s.inflight[key] = c
	s.mu.Unlock()

	// The upstream call does not depend on this request's context: the other
	// waiters would fail together with it.
	fctx, cancel := context.WithTimeout(context.Background(), upstreamTimeout)
	defer cancel()
	s.calls.Add(1)
	snap, err := s.provider.Current(fctx, key.lat, key.lon)
	if err != nil {
		s.errs.Add(1)
	}
	if err == nil {
		snap.FetchedAt = now
	}

	s.mu.Lock()
	delete(s.inflight, key)
	if err == nil {
		s.cache[key] = cacheEntry{snap: snap, at: now}
	}
	c.snap, c.err = snap, err
	close(c.done)
	s.mu.Unlock()

	if err != nil {
		if cached {
			return entry.snap, nil // serve stale data rather than nothing
		}
		return Snapshot{}, err
	}
	return snap, nil
}

func (s *Service) wait(ctx context.Context, c *call, stale cacheEntry, hasStale bool) (Snapshot, error) {
	select {
	case <-c.done:
		if c.err != nil {
			if hasStale {
				return stale.snap, nil
			}
			return Snapshot{}, c.err
		}
		return c.snap, nil
	case <-ctx.Done():
		return Snapshot{}, ctx.Err()
	}
}

// allowFetchLocked counts an upstream call against the per-minute budget.
func (s *Service) allowFetchLocked(now time.Time) bool {
	cutoff := now.Add(-time.Minute)
	kept := s.fetches[:0]
	for _, t := range s.fetches {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	s.fetches = kept
	if s.maxPerMin > 0 && len(s.fetches) >= s.maxPerMin {
		return false
	}
	s.fetches = append(s.fetches, now)
	return true
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// ServiceStats are counters since the server started.
type ServiceStats struct {
	Calls  int64 // upstream requests made
	Hits   int64 // answered from the cache
	Errors int64 // upstream requests that failed
	Busy   int64 // refused because the per-minute budget was used up
}

// Stats returns the counters.
func (s *Service) Stats() ServiceStats {
	return ServiceStats{Calls: s.calls.Load(), Hits: s.hits.Load(), Errors: s.errs.Load(), Busy: s.busy.Load()}
}
