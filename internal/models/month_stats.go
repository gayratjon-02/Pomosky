package models

// DayStat is the focus time of one calendar day (in the user's time zone).
type DayStat struct {
	Date     string // YYYY-MM-DD
	Seconds  int64
	Sessions int
}

// MonthStats is what the monthly statistics view shows.
type MonthStats struct {
	Month            string // YYYY-MM
	TotalSeconds     int64
	Sessions         int
	ActiveDays       int
	StreakDays       int       // consecutive active days ending today (or at the end of a past month)
	PrevTotalSeconds int64     // the previous month, for the comparison
	Days             []DayStat // every day of the month, in order, zero days included
}

// Period is the focus time and number of sessions of a stretch of time.
type Period struct {
	Seconds  int64
	Sessions int
}

// CommunityStats is the focus time of everyone on Pomosky (logged-in and
// anonymous hosts together), shown to every visitor.
type CommunityStats struct {
	Today Period
	Week  Period // the last 7 days including today
	Month Period // the current calendar month
	All   Period
	Days  []DayStat // the last 30 days, every day present
}

// LeaderEntry is one row of the leaderboard. UserID is for the server only
// (to mark the viewer's own row); it is never sent to browsers.
type LeaderEntry struct {
	UserID    UserID
	Nickname  string
	AvatarURL string
	Seconds   int64
	Sessions  int
}

// LeaderRow is a leaderboard line as shown to visitors: no user id, just rank,
// public name, picture and the totals. You marks the viewer's own line.
type LeaderRow struct {
	Rank      int
	Nickname  string
	AvatarURL string
	Seconds   int64
	Sessions  int
	You       bool
}
