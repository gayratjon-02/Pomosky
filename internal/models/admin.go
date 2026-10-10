package models

import "time"

// AdminTotals are the all-time numbers of the admin page.
type AdminTotals struct {
	Users              int
	Rooms              int // rooms stored for logged-in hosts
	RegisteredSeconds  int64
	RegisteredSessions int
	AnonymousSeconds   int64
	AnonymousSessions  int
	HiddenUsers        int // users who switched the leaderboard off
}

// AdminDay is one day of the admin charts.
type AdminDay struct {
	Date               string
	NewUsers           int
	RegisteredSeconds  int64
	RegisteredSessions int
	AnonymousSeconds   int64
	AnonymousSessions  int
	ActiveHosts        int // distinct logged-in hosts who focused that day
}

// AdminUser is a row of the user list. It exists only for the operator.
type AdminUser struct {
	Email             string
	Provider          string
	Nickname          string
	AvatarURL         string
	JoinedAt          time.Time
	ShowOnLeaderboard bool
	FocusSeconds      int64
	Sessions          int
	LastFocusAt       *time.Time
}

// LiveSummary is what is happening right now on this server. Counts only: no
// room links, labels or names.
type LiveSummary struct {
	Rooms   int
	Clients int // connected browsers
	Focus   int // rooms in a focus phase
	Break   int
	Paused  int
	Idle    int
}

// SystemInfo describes the running process.
type SystemInfo struct {
	StartedAt     time.Time
	GoVersion     string
	Goroutines    int
	HeapBytes     uint64
	DatabaseBytes int64
	WeatherCalls  int64 // upstream calls made
	WeatherHits   int64 // answered from the cache
	WeatherErrors int64
	WeatherBusy   int64 // refused because of the per-minute budget
}

// AdminOverview is the whole first screen of the admin page.
type AdminOverview struct {
	Live          LiveSummary
	Totals        AdminTotals
	NewUsersToday int
	NewUsersWeek  int
	NewUsersMonth int
	HostsToday    int
	HostsWeek     int
	HostsMonth    int
	Days          []AdminDay // the last 30 days
	System        SystemInfo
}
