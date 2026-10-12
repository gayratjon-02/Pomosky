// Package admin builds the data of the operator's admin page.
package admin

import (
	"context"
	"runtime"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

// WeatherCounters reports how the weather cache is doing (weather.Service implements it).
type WeatherCounters interface {
	Stats() weather.ServiceStats
}

// recentAnonymousLimit is how many of the latest anonymous focus phases the overview lists.
const recentAnonymousLimit = 40

// Service implements ports.AdminReader.
type Service struct {
	repo    ports.AdminRepository
	live    ports.LiveInspector // may be nil
	weather WeatherCounters     // may be nil
	clock   ports.Clock
	started time.Time
}

var _ ports.AdminReader = (*Service)(nil)

// New creates the Service. live and weatherCounters may be nil.
func New(repo ports.AdminRepository, live ports.LiveInspector, weatherCounters WeatherCounters, clock ports.Clock) *Service {
	return &Service{repo: repo, live: live, weather: weatherCounters, clock: clock, started: clock.Now()}
}

// Overview implements ports.AdminReader.
func (s *Service) Overview(ctx context.Context, tz string) (models.AdminOverview, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}
	now := s.clock.Now()
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)
	weekStart := today.AddDate(0, 0, -6)
	monthStart := today.AddDate(0, 0, -29)

	var out models.AdminOverview
	if out.Totals, err = s.repo.Totals(ctx); err != nil {
		return out, err
	}
	for _, q := range []struct {
		dst  *int
		from time.Time
		fn   func(context.Context, time.Time) (int, error)
	}{
		{&out.NewUsersToday, today, s.repo.NewUsersSince},
		{&out.NewUsersWeek, weekStart, s.repo.NewUsersSince},
		{&out.NewUsersMonth, monthStart, s.repo.NewUsersSince},
		{&out.HostsToday, today, s.repo.ActiveHostsSince},
		{&out.HostsWeek, weekStart, s.repo.ActiveHostsSince},
		{&out.HostsMonth, monthStart, s.repo.ActiveHostsSince},
	} {
		if *q.dst, err = q.fn(ctx, q.from); err != nil {
			return out, err
		}
	}

	rows, err := s.repo.Daily(ctx, monthStart, tomorrow, loc.String())
	if err != nil {
		return out, err
	}
	byDate := make(map[string]models.AdminDay, len(rows))
	for _, r := range rows {
		byDate[r.Date] = r
	}
	for d := monthStart; d.Before(tomorrow); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		day, ok := byDate[date]
		if !ok {
			day = models.AdminDay{Date: date}
		}
		out.Days = append(out.Days, day)
	}

	if out.RecentAnonymous, err = s.repo.RecentAnonymous(ctx, recentAnonymousLimit); err != nil {
		return out, err
	}

	if s.live != nil {
		out.Live = s.live.LiveSummary()
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	out.System = models.SystemInfo{
		StartedAt: s.started, GoVersion: runtime.Version(), Goroutines: runtime.NumGoroutine(), HeapBytes: mem.HeapAlloc,
	}
	if size, err := s.repo.DatabaseSize(ctx); err == nil {
		out.System.DatabaseBytes = size
	}
	if s.weather != nil {
		w := s.weather.Stats()
		out.System.WeatherCalls, out.System.WeatherHits, out.System.WeatherErrors, out.System.WeatherBusy = w.Calls, w.Hits, w.Errors, w.Busy
	}
	return out, nil
}

// Users implements ports.AdminReader.
func (s *Service) Users(ctx context.Context, mostFocused bool, page, pageSize int) ([]models.AdminUser, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	return s.repo.Users(ctx, mostFocused, pageSize, (page-1)*pageSize)
}
