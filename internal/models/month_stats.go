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
