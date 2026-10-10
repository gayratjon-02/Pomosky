package models

import "time"

// UserStats are the aggregated statistics of a registered user.
type UserStats struct {
	UserID         UserID
	TotalFocus     time.Duration
	StreakDays     int
	LastActiveDate time.Time
}
