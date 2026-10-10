package admin

import (
	"context"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

type fakeRepo struct {
	since      map[string][]time.Time // calls by kind
	days       []models.AdminDay
	usersCalls [][3]int
}

func (f *fakeRepo) Totals(context.Context) (models.AdminTotals, error) {
	return models.AdminTotals{Users: 7, RegisteredSeconds: 3600, AnonymousSeconds: 1800}, nil
}
func (f *fakeRepo) NewUsersSince(_ context.Context, since time.Time) (int, error) {
	f.since["new"] = append(f.since["new"], since)
	return len(f.since["new"]), nil
}
func (f *fakeRepo) ActiveHostsSince(_ context.Context, since time.Time) (int, error) {
	f.since["hosts"] = append(f.since["hosts"], since)
	return 10 + len(f.since["hosts"]), nil
}
func (f *fakeRepo) Daily(context.Context, time.Time, time.Time, string) ([]models.AdminDay, error) {
	return f.days, nil
}
func (f *fakeRepo) Users(_ context.Context, mostFocused bool, limit, offset int) ([]models.AdminUser, int, error) {
	m := 0
	if mostFocused {
		m = 1
	}
	f.usersCalls = append(f.usersCalls, [3]int{m, limit, offset})
	return nil, 0, nil
}
func (f *fakeRepo) DatabaseSize(context.Context) (int64, error) { return 4096, nil }

type fixedLive struct{}

func (fixedLive) LiveSummary() models.LiveSummary {
	return models.LiveSummary{Rooms: 2, Clients: 5, Focus: 1}
}

type fakeCounters struct{}

func (fakeCounters) Stats() weather.ServiceStats {
	return weather.ServiceStats{Calls: 3, Hits: 9, Errors: 1, Busy: 2}
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func TestOverviewWindowsAndThirtyDays(t *testing.T) {
	repo := &fakeRepo{since: map[string][]time.Time{}, days: []models.AdminDay{{Date: "2026-10-10", NewUsers: 2, ActiveHosts: 3}}}
	svc := New(repo, fixedLive{}, fakeCounters{}, fixedClock{time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)})

	o, err := svc.Overview(context.Background(), "UTC")
	if err != nil {
		t.Fatal(err)
	}
	// Windows: today, the last 7 days and the last 30 days, in that order, for both counters.
	wantFrom := []string{"2026-10-10", "2026-10-04", "2026-09-11"}
	for kind, froms := range repo.since {
		for i, f := range froms {
			if got := f.Format("2006-01-02"); got != wantFrom[i] {
				t.Errorf("%s window %d starts %s, want %s", kind, i, got, wantFrom[i])
			}
		}
	}
	if o.NewUsersToday != 1 || o.NewUsersWeek != 2 || o.NewUsersMonth != 3 || o.HostsToday != 11 || o.HostsMonth != 13 {
		t.Fatalf("%+v", o)
	}
	if len(o.Days) != 30 || o.Days[0].Date != "2026-09-11" || o.Days[29].Date != "2026-10-10" || o.Days[29].NewUsers != 2 || o.Days[28].NewUsers != 0 {
		t.Fatalf("days: %d, first %s, last %+v", len(o.Days), o.Days[0].Date, o.Days[29])
	}
	if o.Live.Clients != 5 || o.Totals.Users != 7 || o.System.DatabaseBytes != 4096 || o.System.Goroutines < 1 || o.System.GoVersion == "" {
		t.Fatalf("live/totals/system: %+v", o)
	}
	if o.System.WeatherCalls != 3 || o.System.WeatherHits != 9 || o.System.WeatherErrors != 1 || o.System.WeatherBusy != 2 {
		t.Fatalf("weather counters: %+v", o.System)
	}
}

func TestOverviewWorksWithoutLiveOrWeather(t *testing.T) {
	repo := &fakeRepo{since: map[string][]time.Time{}}
	svc := New(repo, nil, nil, fixedClock{time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)})
	if _, err := svc.Overview(context.Background(), "Not/AZone"); err != nil {
		t.Fatal(err)
	}
}

func TestUsersPaging(t *testing.T) {
	repo := &fakeRepo{since: map[string][]time.Time{}}
	svc := New(repo, nil, nil, fixedClock{time.Now()})
	_, _, _ = svc.Users(context.Background(), true, 3, 25)
	_, _, _ = svc.Users(context.Background(), false, 0, 1000)
	if repo.usersCalls[0] != [3]int{1, 25, 50} || repo.usersCalls[1] != [3]int{0, 25, 0} {
		t.Fatalf("calls = %v", repo.usersCalls)
	}
}
