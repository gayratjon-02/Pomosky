package stats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
)

type fakeRepo struct {
	totals         map[string]models.DayStat // by date
	communityCalls int
	boardCalls     int
	board          []models.LeaderEntry
	lastFrom       time.Time
	lastTo         time.Time
	lastLimit      int
}

func (f *fakeRepo) Create(context.Context, models.FocusSession) error { return nil }

func (f *fakeRepo) Leaderboard(_ context.Context, from, to time.Time, limit int) ([]models.LeaderEntry, error) {
	f.boardCalls++
	f.lastFrom, f.lastTo, f.lastLimit = from, to, limit
	return f.board, nil
}

func (f *fakeRepo) CreateAnonymous(context.Context, time.Time, time.Duration) error { return nil }

// CommunityBetween sums the fixture days that fall inside [from, to), as if they were everyone's.
func (f *fakeRepo) CommunityBetween(_ context.Context, from, to time.Time) (models.Period, error) {
	f.communityCalls++
	var p models.Period
	for date, d := range f.totals {
		t, _ := time.Parse("2006-01-02", date)
		if !t.Before(from) && t.Before(to) {
			p.Seconds += d.Seconds
			p.Sessions += d.Sessions
		}
	}
	return p, nil
}

func (f *fakeRepo) CommunityDaily(_ context.Context, from, to time.Time, _ string) ([]models.DayStat, error) {
	var out []models.DayStat
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		if s, ok := f.totals[d.Format("2006-01-02")]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeRepo) DailyTotals(_ context.Context, _ models.UserID, from, to time.Time, _ string) ([]models.DayStat, error) {
	var out []models.DayStat
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		if s, ok := f.totals[d.Format("2006-01-02")]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeRepo) ActiveDays(_ context.Context, _ models.UserID, since time.Time, _ string) ([]string, error) {
	var out []string
	for d := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC); !d.Before(since); d = d.AddDate(0, 0, -1) {
		if s, ok := f.totals[d.Format("2006-01-02")]; ok && s.Seconds > 0 {
			out = append(out, d.Format("2006-01-02"))
		}
	}
	return out, nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func day(date string, seconds int64, sessions int) models.DayStat {
	return models.DayStat{Date: date, Seconds: seconds, Sessions: sessions}
}

func newSvc(totals ...models.DayStat) *Service {
	repo := &fakeRepo{totals: map[string]models.DayStat{}}
	for _, d := range totals {
		repo.totals[d.Date] = d
	}
	return New(repo, fixedClock{time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)})
}

func TestMonthTotalsAndFilledDays(t *testing.T) {
	svc := newSvc(
		day("2026-10-01", 3600, 2), day("2026-10-02", 1800, 1), day("2026-10-09", 600, 1),
		day("2026-09-15", 7200, 3),
	)
	m, err := svc.Month(context.Background(), "u", "2026-10", "UTC")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Days) != 31 || m.Days[0].Date != "2026-10-01" || m.Days[30].Date != "2026-10-31" {
		t.Fatalf("days = %d", len(m.Days))
	}
	if m.TotalSeconds != 6000 || m.Sessions != 4 || m.ActiveDays != 3 || m.PrevTotalSeconds != 7200 {
		t.Fatalf("stats = %+v", m)
	}
}

func TestStreak(t *testing.T) {
	// Today is 10 Oct. Active on 7, 8 and 9: today is empty so the streak is still alive.
	m, _ := newSvc(day("2026-10-07", 60, 1), day("2026-10-08", 60, 1), day("2026-10-09", 60, 1)).Month(context.Background(), "u", "2026-10", "UTC")
	if m.StreakDays != 3 {
		t.Fatalf("streak = %d, want 3", m.StreakDays)
	}
	// Active today as well.
	m, _ = newSvc(day("2026-10-09", 60, 1), day("2026-10-10", 60, 1)).Month(context.Background(), "u", "2026-10", "UTC")
	if m.StreakDays != 2 {
		t.Fatalf("streak = %d, want 2", m.StreakDays)
	}
	// A gap yesterday ends it.
	m, _ = newSvc(day("2026-10-07", 60, 1)).Month(context.Background(), "u", "2026-10", "UTC")
	if m.StreakDays != 0 {
		t.Fatalf("streak = %d, want 0", m.StreakDays)
	}
	// A past month shows the streak ending on its last day: 29 and 30 Sept, not 1 Oct.
	m, _ = newSvc(day("2026-09-29", 60, 1), day("2026-09-30", 60, 1), day("2026-10-01", 60, 1)).Month(context.Background(), "u", "2026-09", "UTC")
	if m.StreakDays != 2 {
		t.Fatalf("past month streak = %d, want 2", m.StreakDays)
	}
}

func TestFutureMonthAndErrors(t *testing.T) {
	svc := newSvc(day("2026-10-01", 60, 1))
	m, err := svc.Month(context.Background(), "u", "2027-01", "UTC")
	if err != nil || len(m.Days) != 31 || m.TotalSeconds != 0 {
		t.Fatalf("future month: %+v %v", m, err)
	}
	if _, err := svc.Month(context.Background(), "u", "nope", "UTC"); !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("bad month: %v", err)
	}
	if _, err := svc.Month(context.Background(), "u", "2026-10", "Not/AZone"); err != nil {
		t.Fatalf("a bad zone should fall back to UTC: %v", err)
	}
}

func TestCommunityWindowsAndChart(t *testing.T) {
	// Now is 2026-10-10. Today 10 Oct, week = 4..10 Oct, month = 1..10 Oct.
	svc := newSvc(
		day("2026-10-10", 3600, 2), // today
		day("2026-10-05", 1800, 1), // this week
		day("2026-10-02", 600, 1),  // this month, not this week
		day("2026-09-20", 7200, 3), // earlier
	)
	c, err := svc.Community(context.Background(), "UTC")
	if err != nil {
		t.Fatal(err)
	}
	if c.Today.Seconds != 3600 || c.Today.Sessions != 2 {
		t.Fatalf("today %+v", c.Today)
	}
	if c.Week.Seconds != 5400 || c.Month.Seconds != 6000 || c.All.Seconds != 13200 || c.All.Sessions != 7 {
		t.Fatalf("week %+v month %+v all %+v", c.Week, c.Month, c.All)
	}
	if len(c.Days) != 30 || c.Days[29].Date != "2026-10-10" || c.Days[0].Date != "2026-09-11" || c.Days[29].Seconds != 3600 {
		t.Fatalf("chart: %d days, first %s last %s", len(c.Days), c.Days[0].Date, c.Days[29].Date)
	}
}

func TestCommunityIsCached(t *testing.T) {
	repo := &fakeRepo{totals: map[string]models.DayStat{"2026-10-10": day("2026-10-10", 60, 1)}}
	clk := &movingClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)}
	svc := New(repo, clk)

	for i := 0; i < 3; i++ {
		if _, err := svc.Community(context.Background(), "UTC"); err != nil {
			t.Fatal(err)
		}
	}
	if repo.communityCalls != 4 { // four windows, queried once
		t.Fatalf("database queried %d times for 3 requests, want 4 (one refresh)", repo.communityCalls)
	}
	clk.t = clk.t.Add(2 * time.Minute)
	_, _ = svc.Community(context.Background(), "UTC")
	if repo.communityCalls != 8 {
		t.Fatalf("no refresh after the cache expired: %d", repo.communityCalls)
	}
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestLeaderboardRanksMarksTheViewerAndCaches(t *testing.T) {
	repo := &fakeRepo{totals: map[string]models.DayStat{}, board: []models.LeaderEntry{
		{UserID: "u2", Nickname: "Bek", Seconds: 7200, Sessions: 3},
		{UserID: "u1", Nickname: "Ayla", AvatarURL: "https://x/a.png", Seconds: 3600, Sessions: 2},
	}}
	clk := &movingClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)}
	svc := New(repo, clk)

	rows, err := svc.Leaderboard(context.Background(), "week", "UTC", "u1")
	if err != nil || len(rows) != 2 {
		t.Fatalf("%+v %v", rows, err)
	}
	if rows[0].Rank != 1 || rows[0].Nickname != "Bek" || rows[0].You || rows[1].Rank != 2 || !rows[1].You {
		t.Fatalf("rows = %+v", rows)
	}
	if repo.lastLimit != LeaderboardSize || !repo.lastFrom.Equal(time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)) || !repo.lastTo.Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("range %v - %v limit %d", repo.lastFrom, repo.lastTo, repo.lastLimit)
	}

	// Another viewer shares the cached list but sees their own mark.
	other, _ := svc.Leaderboard(context.Background(), "week", "UTC", "u2")
	if repo.boardCalls != 1 || !other[0].You || other[1].You {
		t.Fatalf("cache or marks wrong: calls=%d %+v", repo.boardCalls, other)
	}
	anon, _ := svc.Leaderboard(context.Background(), "week", "UTC", "")
	if anon[0].You || anon[1].You {
		t.Fatal("an anonymous visitor has no own row")
	}
}

func TestLeaderboardPeriodsAndErrors(t *testing.T) {
	repo := &fakeRepo{totals: map[string]models.DayStat{}}
	svc := New(repo, &movingClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)})
	want := map[string][2]string{
		"today": {"2026-10-10", "2026-10-11"},
		"month": {"2026-10-01", "2026-10-11"},
		"all":   {"1970-01-01", "2026-10-11"},
	}
	for period, r := range want {
		if _, err := svc.Leaderboard(context.Background(), period, "UTC", ""); err != nil {
			t.Fatal(err)
		}
		if repo.lastFrom.Format("2006-01-02") != r[0] || repo.lastTo.Format("2006-01-02") != r[1] {
			t.Errorf("%s: %s - %s", period, repo.lastFrom.Format("2006-01-02"), repo.lastTo.Format("2006-01-02"))
		}
	}
	if _, err := svc.Leaderboard(context.Background(), "decade", "UTC", ""); !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("bad period: %v", err)
	}
}
