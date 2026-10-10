// Package stats turns stored focus sessions into the monthly statistics view.
package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// maxStreakLookback bounds the streak query; a streak longer than this is shown as this many days.
const maxStreakLookback = 400 * 24 * time.Hour

// Service implements ports.StatsReader.
type Service struct {
	sessions ports.FocusSessionRepository
	clock    ports.Clock
}

var _ ports.StatsReader = (*Service)(nil)

// New creates the statistics Service.
func New(sessions ports.FocusSessionRepository, clock ports.Clock) *Service {
	return &Service{sessions: sessions, clock: clock}
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
