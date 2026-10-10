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
	totals map[string]models.DayStat // by date
}

func (f *fakeRepo) Create(context.Context, models.FocusSession) error { return nil }

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
