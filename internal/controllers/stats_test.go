package controllers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
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
