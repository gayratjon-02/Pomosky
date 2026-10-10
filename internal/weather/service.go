package weather

import (
	"context"
	"math"
	"sync"
	"time"
)

// Provider fetches a fresh snapshot from an upstream weather API.
type Provider interface {
	Current(ctx context.Context, lat, lon float64) (Snapshot, error)
}

// Service caches snapshots so that every viewer of a room shares one
// upstream request per TTL. If the upstream fails, a stale snapshot is served.
type Service struct {
	provider Provider
	ttl      time.Duration
	now      func() time.Time

	mu    sync.Mutex
	cache map[cacheKey]cacheEntry
}

type cacheKey struct{ lat, lon float64 }

type cacheEntry struct {
	snap Snapshot
	at   time.Time
}

// NewService creates a caching Service.
func NewService(p Provider, ttl time.Duration, now func() time.Time) *Service {
	return &Service{provider: p, ttl: ttl, now: now, cache: map[cacheKey]cacheEntry{}}
}

// Current returns the (possibly cached) snapshot for the given location.
func (s *Service) Current(ctx context.Context, lat, lon float64) (Snapshot, error) {
	key := cacheKey{lat: round2(lat), lon: round2(lon)}
	now := s.now()

	s.mu.Lock()
	entry, ok := s.cache[key]
	s.mu.Unlock()
	if ok && now.Sub(entry.at) < s.ttl {
		return entry.snap, nil
	}

	snap, err := s.provider.Current(ctx, key.lat, key.lon)
	if err != nil {
		if ok {
			return entry.snap, nil // serve stale data rather than nothing
		}
		return Snapshot{}, err
	}
	snap.FetchedAt = now

	s.mu.Lock()
	s.cache[key] = cacheEntry{snap: snap, at: now}
	s.mu.Unlock()
	return snap, nil
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
