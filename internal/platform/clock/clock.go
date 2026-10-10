// Package clock provides ports.Clock implementations.
package clock

import "time"

// Real is the system clock.
type Real struct{}

// Now returns the current time.
func (Real) Now() time.Time { return time.Now() }

// Fixed is a clock frozen at T, for tests.
type Fixed struct{ T time.Time }

// Now returns the frozen time.
func (f Fixed) Now() time.Time { return f.T }
