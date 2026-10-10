// Package stats turns stored focus sessions into the monthly statistics view.
package stats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// maxStreakLookback bounds the streak query; a streak longer than this is shown as this many days.
const maxStreakLookback = 400 * 24 * time.Hour

// communityCacheTTL keeps the public totals from hitting the database on every page view.
const communityCacheTTL = time.Minute

// Service implements ports.StatsReader.
type Service struct {
	sessions ports.FocusSessionRepository
	clock    ports.Clock

	mu        sync.Mutex
	community map[string]communityEntry // by time zone
	boards    map[string]boardEntry     // by period and time zone
}

type boardEntry struct {
	entries []models.LeaderEntry
	at      time.Time
}

type communityEntry struct {
	stats models.CommunityStats
	at    time.Time
}

var _ ports.StatsReader = (*Service)(nil)

// New creates the statistics Service.
func New(sessions ports.FocusSessionRepository, clock ports.Clock) *Service {
	return &Service{sessions: sessions, clock: clock, community: map[string]communityEntry{}, boards: map[string]boardEntry{}}
}

// Month implements ports.StatsReader.
func (s *Service) Month(ctx context.Context, user models.UserID, month, tz string) (models.MonthStats, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}
	first, err := time.ParseInLocation("2006-01", month, loc)
	if err != nil {
		return models.MonthStats{}, fmt.Errorf("%w: month %q", apperr.ErrInvalid, month)
	}
	next := first.AddDate(0, 1, 0)
	prev := first.AddDate(0, -1, 0)
	now := s.clock.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	out := models.MonthStats{Month: month}
	if first.After(today) {
		out.Days = emptyDays(first, next)
		return out, nil
	}

	rows, err := s.sessions.DailyTotals(ctx, user, first, next, loc.String())
	if err != nil {
		return models.MonthStats{}, err
	}
	byDate := make(map[string]models.DayStat, len(rows))
	for _, r := range rows {
		byDate[r.Date] = r
	}
	for d := first; d.Before(next); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		day, ok := byDate[key]
		if !ok {
			day = models.DayStat{Date: key}
		}
		out.Days = append(out.Days, day)
		out.TotalSeconds += day.Seconds
		out.Sessions += day.Sessions
		if day.Seconds > 0 {
			out.ActiveDays++
		}
	}

	prevRows, err := s.sessions.DailyTotals(ctx, user, prev, first, loc.String())
	if err != nil {
		return models.MonthStats{}, err
	}
	for _, r := range prevRows {
		out.PrevTotalSeconds += r.Seconds
	}

	ref := next.AddDate(0, 0, -1) // last day of the month
	if today.Before(ref) {
		ref = today
	}
	active, err := s.sessions.ActiveDays(ctx, user, ref.Add(-maxStreakLookback), loc.String())
	if err != nil {
		return models.MonthStats{}, err
	}
	out.StreakDays = streak(active, ref, ref.Equal(today))
	return out, nil
}

func emptyDays(first, next time.Time) []models.DayStat {
	var days []models.DayStat
	for d := first; d.Before(next); d = d.AddDate(0, 0, 1) {
		days = append(days, models.DayStat{Date: d.Format("2006-01-02")})
	}
	return days
}

// streak counts consecutive active days ending at ref. When ref is today and
// today has no focus yet, the streak is still alive and counts from yesterday.
func streak(active []string, ref time.Time, refIsToday bool) int {
	set := make(map[string]bool, len(active))
	for _, d := range active {
		set[d] = true
	}
	d := ref
	if refIsToday && !set[d.Format("2006-01-02")] {
		d = d.AddDate(0, 0, -1)
	}
	n := 0
	for set[d.Format("2006-01-02")] {
		n++
		d = d.AddDate(0, 0, -1)
	}
	return n
}

// Community implements ports.StatsReader: today, the last 7 days, this month
// and all time for everyone, plus the last 30 days day by day. The result is
// cached for a minute per time zone.
func (s *Service) Community(ctx context.Context, tz string) (models.CommunityStats, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}
	key := loc.String()
	now := s.clock.Now()

	s.mu.Lock()
	entry, ok := s.community[key]
	s.mu.Unlock()
	if ok && now.Sub(entry.at) < communityCacheTTL {
		return entry.stats, nil
	}

	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)
	weekStart := today.AddDate(0, 0, -6)
	monthStart := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
	chartStart := today.AddDate(0, 0, -29)
	epoch := time.Unix(0, 0)

	var out models.CommunityStats
	for _, w := range []struct {
		dst      *models.Period
		from, to time.Time
	}{
		{&out.Today, today, tomorrow},
		{&out.Week, weekStart, tomorrow},
		{&out.Month, monthStart, tomorrow},
		{&out.All, epoch, tomorrow},
	} {
		p, err := s.sessions.CommunityBetween(ctx, w.from, w.to)
		if err != nil {
			return models.CommunityStats{}, err
		}
		*w.dst = p
	}

	rows, err := s.sessions.CommunityDaily(ctx, chartStart, tomorrow, key)
	if err != nil {
		return models.CommunityStats{}, err
	}
	byDate := make(map[string]models.DayStat, len(rows))
	for _, r := range rows {
		byDate[r.Date] = r
	}
	for d := chartStart; d.Before(tomorrow); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		day, ok := byDate[date]
		if !ok {
			day = models.DayStat{Date: date}
		}
		out.Days = append(out.Days, day)
	}

	s.mu.Lock()
	s.community[key] = communityEntry{stats: out, at: now}
	s.mu.Unlock()
	return out, nil
}

// LeaderboardSize is how many users the leaderboard lists.
const LeaderboardSize = 50

// Leaderboard implements ports.StatsReader: the most focused users of a period
// who allow it, best first. The list is cached for a minute per period and time
// zone; the viewer's own row is marked afterwards, so the cache stays shared.
func (s *Service) Leaderboard(ctx context.Context, period, tz string, viewer models.UserID) ([]models.LeaderRow, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}
	now := s.clock.Now()
	from, to, ok := periodRange(period, now, loc)
	if !ok {
		return nil, fmt.Errorf("%w: period %q", apperr.ErrInvalid, period)
	}

	key := period + "|" + loc.String()
	s.mu.Lock()
	entry, cached := s.boards[key]
	s.mu.Unlock()
	if !cached || now.Sub(entry.at) >= communityCacheTTL {
		entries, err := s.sessions.Leaderboard(ctx, from, to, LeaderboardSize)
		if err != nil {
			return nil, err
		}
		entry = boardEntry{entries: entries, at: now}
		s.mu.Lock()
		s.boards[key] = entry
		s.mu.Unlock()
	}

	rows := make([]models.LeaderRow, 0, len(entry.entries))
	for i, e := range entry.entries {
		rows = append(rows, models.LeaderRow{
			Rank: i + 1, Nickname: e.Nickname, AvatarURL: e.AvatarURL, Seconds: e.Seconds, Sessions: e.Sessions,
			You: viewer != "" && e.UserID == viewer,
		})
	}
	return rows, nil
}

// periodRange returns the time range of today, the last 7 days, this calendar
// month or all time, in the given zone.
func periodRange(period string, now time.Time, loc *time.Location) (from, to time.Time, ok bool) {
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	to = today.AddDate(0, 0, 1)
	switch period {
	case "today":
		return today, to, true
	case "week":
		return today.AddDate(0, 0, -6), to, true
	case "month":
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc), to, true
	case "all":
		return time.Unix(0, 0), to, true
	}
	return time.Time{}, time.Time{}, false
}
