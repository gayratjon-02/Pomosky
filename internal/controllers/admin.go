package controllers

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

const adminPageSize = 25

// adminUser returns the logged-in user when they are an operator: their
// verified email is in ADMIN_EMAILS. Both login methods verify the address
// (Google reports email_verified, a magic link proves the inbox).
func (c *Controllers) adminUser(r *http.Request) *models.User {
	u := reqctx.User(r.Context())
	if u == nil || u.Email == nil || len(c.deps.Config.AdminEmails) == 0 {
		return nil
	}
	email := strings.ToLower(*u.Email)
	for _, a := range c.deps.Config.AdminEmails {
		if a == email {
			return u
		}
	}
	return nil
}

// isAdminHost reports whether the request came in on ADMIN_HOST (port ignored).
func (c *Controllers) isAdminHost(r *http.Request) bool {
	want := c.deps.Config.AdminHost
	if want == "" {
		return false
	}
	host := strings.ToLower(r.Host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host == want
}

// onAdminHost is true when the admin may be served for this request: always
// when ADMIN_HOST is not configured, otherwise only on that host.
func (c *Controllers) onAdminHost(r *http.Request) bool {
	return c.deps.Config.AdminHost == "" || c.isAdminHost(r)
}

// adminGate answers 404 for everybody who is not an operator, so the page does
// not reveal that it exists. It returns false when the request was answered.
func (c *Controllers) adminGate(w http.ResponseWriter, r *http.Request, json bool) bool {
	if u := c.adminUser(r); u != nil && c.onAdminHost(r) {
		c.deps.Log.Info("admin access", "user", u.ID, "path", r.URL.Path)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		return true
	}
	if json {
		http.NotFound(w, r)
	} else {
		c.NotFound(w, r)
	}
	return false
}

// Admin serves GET /admin.
func (c *Controllers) Admin(w http.ResponseWriter, r *http.Request) {
	if !c.adminGate(w, r, false) {
		return
	}
	c.render(w, r, http.StatusOK, "admin", views.PageData{
		Title:   "Admin",
		Scripts: []string{"/static/admin.js"},
	})
}

type adminDayJSON struct {
	Date               string `json:"date"`
	NewUsers           int    `json:"newUsers"`
	RegisteredSeconds  int64  `json:"registeredSeconds"`
	RegisteredSessions int    `json:"registeredSessions"`
	AnonymousSeconds   int64  `json:"anonymousSeconds"`
	AnonymousSessions  int    `json:"anonymousSessions"`
	ActiveHosts        int    `json:"activeHosts"`
}

// adminAnonymousJSON is one focus phase of someone without an account: when and how long, nothing else.
type adminAnonymousJSON struct {
	StartedAtMS int64 `json:"startedAtMs"`
	Seconds     int64 `json:"seconds"`
}

type adminOverviewJSON struct {
	Live struct {
		Rooms   int `json:"rooms"`
		Clients int `json:"clients"`
		Focus   int `json:"focus"`
		Break   int `json:"break"`
		Paused  int `json:"paused"`
		Idle    int `json:"idle"`
	} `json:"live"`
	Totals struct {
		Users              int   `json:"users"`
		Rooms              int   `json:"rooms"`
		RegisteredSeconds  int64 `json:"registeredSeconds"`
		RegisteredSessions int   `json:"registeredSessions"`
		AnonymousSeconds   int64 `json:"anonymousSeconds"`
		AnonymousSessions  int   `json:"anonymousSessions"`
		HiddenUsers        int   `json:"hiddenUsers"`
	} `json:"totals"`
	NewUsers struct {
		Today int `json:"today"`
		Week  int `json:"week"`
		Month int `json:"month"`
	} `json:"newUsers"`
	ActiveHosts struct {
		Today int `json:"today"`
		Week  int `json:"week"`
		Month int `json:"month"`
	} `json:"activeHosts"`
	Days            []adminDayJSON       `json:"days"`
	RecentAnonymous []adminAnonymousJSON `json:"recentAnonymous"`
	System          struct {
		StartedAtMS   int64  `json:"startedAtMs"`
		GoVersion     string `json:"goVersion"`
		Goroutines    int    `json:"goroutines"`
		HeapBytes     uint64 `json:"heapBytes"`
		DatabaseBytes int64  `json:"databaseBytes"`
		WeatherCalls  int64  `json:"weatherCalls"`
		WeatherHits   int64  `json:"weatherHits"`
		WeatherErrors int64  `json:"weatherErrors"`
		WeatherBusy   int64  `json:"weatherBusy"`
	} `json:"system"`
}

func toAdminOverviewJSON(o models.AdminOverview) adminOverviewJSON {
	var j adminOverviewJSON
	j.Live.Rooms, j.Live.Clients, j.Live.Focus, j.Live.Break, j.Live.Paused, j.Live.Idle = o.Live.Rooms, o.Live.Clients, o.Live.Focus, o.Live.Break, o.Live.Paused, o.Live.Idle
	t := o.Totals
	j.Totals.Users, j.Totals.Rooms, j.Totals.RegisteredSeconds, j.Totals.RegisteredSessions = t.Users, t.Rooms, t.RegisteredSeconds, t.RegisteredSessions
	j.Totals.AnonymousSeconds, j.Totals.AnonymousSessions, j.Totals.HiddenUsers = t.AnonymousSeconds, t.AnonymousSessions, t.HiddenUsers
	j.NewUsers.Today, j.NewUsers.Week, j.NewUsers.Month = o.NewUsersToday, o.NewUsersWeek, o.NewUsersMonth
	j.ActiveHosts.Today, j.ActiveHosts.Week, j.ActiveHosts.Month = o.HostsToday, o.HostsWeek, o.HostsMonth
	j.Days = make([]adminDayJSON, 0, len(o.Days))
	for _, d := range o.Days {
		j.Days = append(j.Days, adminDayJSON{
			Date: d.Date, NewUsers: d.NewUsers, RegisteredSeconds: d.RegisteredSeconds, RegisteredSessions: d.RegisteredSessions,
			AnonymousSeconds: d.AnonymousSeconds, AnonymousSessions: d.AnonymousSessions, ActiveHosts: d.ActiveHosts,
		})
	}
	j.RecentAnonymous = make([]adminAnonymousJSON, 0, len(o.RecentAnonymous))
	for _, a := range o.RecentAnonymous {
		j.RecentAnonymous = append(j.RecentAnonymous, adminAnonymousJSON{StartedAtMS: a.StartedAt.UnixMilli(), Seconds: int64(a.Focused.Seconds())})
	}
	s := o.System
	j.System.StartedAtMS, j.System.GoVersion, j.System.Goroutines, j.System.HeapBytes = s.StartedAt.UnixMilli(), s.GoVersion, s.Goroutines, s.HeapBytes
	j.System.DatabaseBytes, j.System.WeatherCalls, j.System.WeatherHits, j.System.WeatherErrors, j.System.WeatherBusy = s.DatabaseBytes, s.WeatherCalls, s.WeatherHits, s.WeatherErrors, s.WeatherBusy
	return j
}

// AdminOverviewAPI serves GET /api/admin/overview?tz=Area/City.
func (c *Controllers) AdminOverviewAPI(w http.ResponseWriter, r *http.Request) {
	if !c.adminGate(w, r, true) {
		return
	}
	if c.deps.Admin == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	o, err := c.deps.Admin.Overview(r.Context(), r.URL.Query().Get("tz"))
	if err != nil {
		c.deps.Log.Error("admin overview", "err", err)
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toAdminOverviewJSON(o))
}

type adminUserJSON struct {
	Email             string `json:"email"`
	Provider          string `json:"provider"`
	Name              string `json:"name"`
	AvatarURL         string `json:"avatarUrl,omitempty"`
	JoinedAtMS        int64  `json:"joinedAtMs"`
	ShowOnLeaderboard bool   `json:"showOnLeaderboard"`
	FocusSeconds      int64  `json:"focusSeconds"`
	Sessions          int    `json:"sessions"`
	LastFocusAtMS     int64  `json:"lastFocusAtMs,omitempty"`
}

// AdminUsersAPI serves GET /api/admin/users?sort=newest|focus&page=1.
func (c *Controllers) AdminUsersAPI(w http.ResponseWriter, r *http.Request) {
	if !c.adminGate(w, r, true) {
		return
	}
	if c.deps.Admin == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	users, total, err := c.deps.Admin.Users(r.Context(), r.URL.Query().Get("sort") == "focus", page, adminPageSize)
	if err != nil {
		c.deps.Log.Error("admin users", "err", err)
		httpx.WriteError(w, err)
		return
	}
	out := make([]adminUserJSON, 0, len(users))
	for _, u := range users {
		j := adminUserJSON{
			Email: u.Email, Provider: u.Provider, Name: u.Nickname, AvatarURL: u.AvatarURL, JoinedAtMS: u.JoinedAt.UnixMilli(),
			ShowOnLeaderboard: u.ShowOnLeaderboard, FocusSeconds: u.FocusSeconds, Sessions: u.Sessions,
		}
		if u.LastFocusAt != nil {
			j.LastFocusAtMS = u.LastFocusAt.UnixMilli()
		}
		out = append(out, j)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": out, "total": total, "pageSize": adminPageSize, "serverTimeMs": time.Now().UnixMilli()})
}
