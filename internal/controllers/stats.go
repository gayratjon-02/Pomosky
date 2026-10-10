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

type periodJSON struct {
	Seconds  int64 `json:"seconds"`
	Sessions int   `json:"sessions"`
}

type communityJSON struct {
	Today periodJSON `json:"today"`
	Week  periodJSON `json:"week"`
	Month periodJSON `json:"month"`
	All   periodJSON `json:"all"`
	Days  []dayJSON  `json:"days"`
}

func toCommunityJSON(c models.CommunityStats) communityJSON {
	p := func(x models.Period) periodJSON { return periodJSON{Seconds: x.Seconds, Sessions: x.Sessions} }
	days := make([]dayJSON, 0, len(c.Days))
	for _, d := range c.Days {
		days = append(days, dayJSON{Date: d.Date, Seconds: d.Seconds, Sessions: d.Sessions})
	}
	return communityJSON{Today: p(c.Today), Week: p(c.Week), Month: p(c.Month), All: p(c.All), Days: days}
}

// CommunityAPI serves GET /api/stats/community?tz=Area/City: how much everyone
// has focused on Pomosky, public and cached.
func (c *Controllers) CommunityAPI(w http.ResponseWriter, r *http.Request) {
	if c.deps.Stats == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	stats, err := c.deps.Stats.Community(r.Context(), r.URL.Query().Get("tz"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	httpx.WriteJSON(w, http.StatusOK, toCommunityJSON(stats))
}

type leaderJSON struct {
	Rank      int    `json:"rank"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Seconds   int64  `json:"seconds"`
	Sessions  int    `json:"sessions"`
	You       bool   `json:"you,omitempty"`
}

// LeaderboardAPI serves GET /api/stats/leaderboard?period=today|week|month|all&tz=Area/City:
// the most focused users who allow it, best first. Public. It never contains user ids or emails.
func (c *Controllers) LeaderboardAPI(w http.ResponseWriter, r *http.Request) {
	if c.deps.Stats == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	period := r.URL.Query().Get("period")
	var viewer models.UserID
	if u := reqctx.User(r.Context()); u != nil {
		viewer = u.ID
	}
	rows, err := c.deps.Stats.Leaderboard(r.Context(), period, r.URL.Query().Get("tz"), viewer)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	out := make([]leaderJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, leaderJSON{Rank: row.Rank, Name: row.Nickname, AvatarURL: row.AvatarURL, Seconds: row.Seconds, Sessions: row.Sessions, You: row.You})
	}
	// The viewer's own row differs per user, so shared caches must not keep it.
	w.Header().Set("Cache-Control", "private, max-age=30")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"period": period, "entries": out})
}
