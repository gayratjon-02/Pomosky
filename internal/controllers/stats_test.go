package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

func TestDemoStatsFixtureShape(t *testing.T) {
	c := newTestControllers(t)
	rec := httptest.NewRecorder()
	c.DemoStats(rec, httptest.NewRequest("GET", "/api/stats/demo?month=2026-02", nil))
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var got statsJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Month != "2026-02" || len(got.Days) != 28 {
		t.Fatalf("month %q with %d days", got.Month, len(got.Days))
	}
	var total int64
	var sessions, active int
	for _, d := range got.Days {
		total += d.Seconds
		sessions += d.Sessions
		if d.Seconds > 0 {
			active++
		}
	}
	if total != got.TotalSeconds || sessions != got.Sessions || active != got.ActiveDays {
		t.Fatalf("totals do not add up: %+v", got)
	}
}

func TestDemoStatsFutureMonthIsEmpty(t *testing.T) {
	m := fixtureMonth(time.Now().AddDate(0, 3, 0), time.Now())
	if m.TotalSeconds != 0 || m.ActiveDays != 0 {
		t.Fatalf("future month has data: %+v", m)
	}
}

func TestDemoStatsRejectsBadMonth(t *testing.T) {
	c := newTestControllers(t)
	rec := httptest.NewRecorder()
	c.DemoStats(rec, httptest.NewRequest("GET", "/api/stats/demo?month=nope", nil))
	if rec.Code != 400 {
		t.Fatalf("status %d", rec.Code)
	}
}

type fakeStats struct {
	rows   []models.LeaderRow
	err    error
	period string
	viewer models.UserID
}

func (f *fakeStats) Month(context.Context, models.UserID, string, string) (models.MonthStats, error) {
	return models.MonthStats{}, nil
}
func (f *fakeStats) Community(context.Context, string) (models.CommunityStats, error) {
	return models.CommunityStats{}, nil
}
func (f *fakeStats) Leaderboard(_ context.Context, period, _ string, viewer models.UserID) ([]models.LeaderRow, error) {
	f.period, f.viewer = period, viewer
	return f.rows, f.err
}

func TestLeaderboardAPIHidesIdentifiers(t *testing.T) {
	c := newTestControllers(t)
	fs := &fakeStats{rows: []models.LeaderRow{{Rank: 1, Nickname: "Ayla", AvatarURL: "https://x/a.png", Seconds: 3600, Sessions: 2, You: true}}}
	c.deps.Stats = fs

	req := httptest.NewRequest("GET", "/api/stats/leaderboard?period=week&tz=UTC", nil)
	req = req.WithContext(reqctx.WithUser(req.Context(), &models.User{ID: "secret-id", Nickname: "Ayla"}))
	rec := httptest.NewRecorder()
	c.LeaderboardAPI(rec, req)

	if rec.Code != 200 || fs.period != "week" || fs.viewer != "secret-id" {
		t.Fatalf("status %d, period %q, viewer %q", rec.Code, fs.period, fs.viewer)
	}
	body := rec.Body.String()
	if strings.Contains(body, "secret-id") || strings.Contains(body, "email") {
		t.Fatalf("the response leaks an identifier: %s", body)
	}
	var got struct {
		Entries []leaderJSON `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got.Entries) != 1 || got.Entries[0].Name != "Ayla" || !got.Entries[0].You {
		t.Fatalf("body = %s", body)
	}

	fs.err = apperr.ErrInvalid
	rec = httptest.NewRecorder()
	c.LeaderboardAPI(rec, httptest.NewRequest("GET", "/api/stats/leaderboard?period=nope", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad period: %d", rec.Code)
	}
}
