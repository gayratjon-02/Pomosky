package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

// The JSON shape of GET /api/stats (documented in docs/api.md).
type statsJSON struct {
	Month            string    `json:"month"`
	TotalSeconds     int64     `json:"totalSeconds"`
	Sessions         int       `json:"sessions"`
	ActiveDays       int       `json:"activeDays"`
	StreakDays       int       `json:"streakDays"`
	PrevTotalSeconds int64     `json:"prevTotalSeconds"`
	Days             []dayJSON `json:"days"`
}

type dayJSON struct {
	Date     string `json:"date"`
	Seconds  int64  `json:"seconds"`
	Sessions int    `json:"sessions"`
}

func toStatsJSON(m models.MonthStats) statsJSON {
	days := make([]dayJSON, 0, len(m.Days))
	for _, d := range m.Days {
		days = append(days, dayJSON{Date: d.Date, Seconds: d.Seconds, Sessions: d.Sessions})
	}
	return statsJSON{
		Month: m.Month, TotalSeconds: m.TotalSeconds, Sessions: m.Sessions, ActiveDays: m.ActiveDays,
		StreakDays: m.StreakDays, PrevTotalSeconds: m.PrevTotalSeconds, Days: days,
	}
}

// parseMonth reads ?month=YYYY-MM and falls back to the month of now.
func parseMonth(r *http.Request, now time.Time) (time.Time, bool) {
	raw := r.URL.Query().Get("month")
	if raw == "" {
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), true
	}
	t, err := time.ParseInLocation("2006-01", raw, now.Location())
	return t, err == nil
}

func writeStats(w http.ResponseWriter, m models.MonthStats) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(toStatsJSON(m))
}

// StatsAPI serves GET /api/stats?month=YYYY-MM&tz=Area/City: the monthly focus
// statistics of the logged-in user.
func (c *Controllers) StatsAPI(w http.ResponseWriter, r *http.Request) {
	user := reqctx.User(r.Context())
	if user == nil {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	if c.deps.Stats == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	month, ok := parseMonth(r, time.Now())
	if !ok {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	stats, err := c.deps.Stats.Month(r.Context(), user.ID, month.Format("2006-01"), r.URL.Query().Get("tz"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	writeStats(w, stats)
}
