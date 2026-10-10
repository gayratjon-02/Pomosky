package ports

import "time"

// Clock provides the current time. Production uses platform/clock.Real,
// tests use platform/clock.Fixed.
type Clock interface {
	Now() time.Time
}
