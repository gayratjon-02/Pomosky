// Package ratelimit implements an in-memory token-bucket limiter keyed by string
// (client IP, user ID, ...). Suitable for a single server instance.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter allows RatePerSec events per second per key with the given burst.
type Limiter struct {
	rate  float64
	burst float64
	now   func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// New creates a Limiter. now is injected for tests; pass time.Now in production.
func New(ratePerSec float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: ratePerSec, burst: float64(burst), now: now, buckets: map[string]*bucket{}}
}

// Allow reports whether one more event for key is allowed right now.
func (l *Limiter) Allow(key string) bool {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
