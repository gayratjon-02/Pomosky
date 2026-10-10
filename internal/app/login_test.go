package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Needs a real database: set TEST_DATABASE_URL (the tables are emptied).
func TestLoginAccountStatsAndDeleteFlow(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	a, err := New(config.Config{Env: "development", StaticDir: t.TempDir(), DatabaseURL: dbURL, BaseURL: "http://pomosky.test"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err := a.db.SQL().Exec("TRUNCATE users CASCADE"); err != nil {
		t.Fatal(err)
	}
	h := a.Handler()

	// Anonymous visitors see the example statistics behind a login prompt; their own API says 401.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/account", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Example data") || !strings.Contains(rec.Body.String(), "/api/stats/demo") {
		t.Fatalf("anonymous /account = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "Delete my account") {
		t.Fatal("the preview must not offer account deletion")
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/stats", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous /api/stats = %d", rec.Code)
	}

	// Sign a user in through the service the app is wired with.
	svc := a.sessions.(*auth.Service)
	login, err := svc.SignIn(context.Background(), models.Identity{Provider: "google", Subject: "g1", Email: "ayla@example.com", Name: "Ayla"})
	if err != nil {
		t.Fatal(err)
	}
	session := httpx.NewSessionCookie(login.Token, false, login.Expires)

	rec = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/account", nil)
	req.AddCookie(session)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Ayla") {
		t.Fatalf("/account = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/stats?month=2026-10&tz=Asia/Seoul", nil)
	req.AddCookie(session)
	h.ServeHTTP(rec, req)
	var stats struct {
		Month string `json:"month"`
		Days  []any  `json:"days"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &stats) != nil || stats.Month != "2026-10" || len(stats.Days) != 31 {
		t.Fatalf("/api/stats = %d %s", rec.Code, rec.Body.String())
	}

	// State-changing requests need the CSRF token.
	csrfCookie := csrfFor(t, h)
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		form.Set("csrf", csrfCookie.Value)
		req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(session)
		req.AddCookie(csrfCookie)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	// A logged-in host's room is stored with them as owner, and focus time counts.
	rec = post("/r/new", url.Values{})
	loc := rec.Header().Get("Location")
	if rec.Code != http.StatusSeeOther || !strings.HasPrefix(loc, "/r/") {
		t.Fatalf("create room = %d %q", rec.Code, loc)
	}
	roomID := room.ID(strings.SplitN(strings.TrimPrefix(loc, "/r/"), "/", 2)[0])
	rh, ok := a.hub.Room(roomID)
	if !ok {
		t.Fatal("room is not live")
	}
	now := time.Now()
	start := now.Add(-40 * time.Minute)
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 60 * time.Minute}, start); err != nil {
		t.Fatal(err)
	}
	if err := rh.Apply(room.TimerReset{}, now.Add(-10*time.Minute)); err != nil { // 30 min focused
		t.Fatal(err)
	}
	a.recorder.Close() // flush the queue
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/stats?month="+now.UTC().Format("2006-01")+"&tz=UTC", nil)
	req.AddCookie(session)
	h.ServeHTTP(rec, req)
	var month struct {
		TotalSeconds int64 `json:"totalSeconds"`
		Sessions     int   `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &month); err != nil || month.TotalSeconds != 1800 || month.Sessions != 1 {
		t.Fatalf("stats after a focus session: %d %s", rec.Code, rec.Body.String())
	}

	if rec := post("/account/delete", url.Values{"confirm": {"nope"}}); rec.Code != http.StatusBadRequest {
		t.Fatalf("delete without confirmation = %d", rec.Code)
	}
	if rec := post("/account/delete", url.Values{"confirm": {"delete"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("delete = %d", rec.Code)
	}
	if _, err := svc.UserFromSession(context.Background(), login.Token); err == nil {
		t.Fatal("session still valid after the account was deleted")
	}
}

func csrfFor(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	for _, c := range rec.Result().Cookies() {
		if c.Name == "pomosky_csrf" {
			return c
		}
	}
	t.Fatal("no csrf cookie")
	return nil
}

// The Google button shares a form with the required email field, so it must
// skip browser validation or it can never be clicked with an empty email.
func TestGoogleButtonSkipsEmailValidation(t *testing.T) {
	a, err := New(config.Config{Env: "development", StaticDir: t.TempDir(), OAuth: config.OAuthConfig{GoogleClientID: "id", GoogleClientSecret: "secret"}, DatabaseURL: os.Getenv("TEST_DATABASE_URL")})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if a.sessions == nil {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/login", nil))
	body := rec.Body.String()
	i := strings.Index(body, `value="google"`)
	if i < 0 || !strings.Contains(body[i:i+60], "formnovalidate") {
		t.Fatalf("the Google button lacks formnovalidate")
	}
}

// The admin page and its API are visible to operators only.
func TestAdminAccessFlow(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	a, err := New(config.Config{Env: "development", StaticDir: t.TempDir(), DatabaseURL: dbURL, BaseURL: "http://pomosky.test", AdminEmails: []string{"boss@example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err := a.db.SQL().Exec("TRUNCATE users CASCADE"); err != nil {
		t.Fatal(err)
	}
	h := a.Handler()
	svc := a.sessions.(*auth.Service)
	get := func(path string, login *models.Login) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		if login != nil {
			req.AddCookie(httpx.NewSessionCookie(login.Token, false, login.Expires))
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	boss, _ := svc.SignIn(context.Background(), models.Identity{Provider: "google", Subject: "g-boss", Email: "boss@example.com", Name: "Boss"})
	other, _ := svc.SignIn(context.Background(), models.Identity{Provider: "google", Subject: "g-other", Email: "other@example.com", Name: "Other"})

	for _, path := range []string{"/admin", "/api/admin/overview", "/api/admin/users"} {
		if rec := get(path, nil); rec.Code != http.StatusNotFound {
			t.Errorf("anonymous %s = %d", path, rec.Code)
		}
		if rec := get(path, &other); rec.Code != http.StatusNotFound {
			t.Errorf("non-operator %s = %d", path, rec.Code)
		}
		if rec := get(path, &boss); rec.Code != http.StatusOK {
			t.Errorf("operator %s = %d", path, rec.Code)
		}
	}
	var o struct {
		Totals struct {
			Users int `json:"users"`
		} `json:"totals"`
		Days []any `json:"days"`
	}
	rec := get("/api/admin/overview?tz=UTC", &boss)
	if err := json.Unmarshal(rec.Body.Bytes(), &o); err != nil || o.Totals.Users != 2 || len(o.Days) != 30 {
		t.Fatalf("overview: %s", rec.Body.String())
	}
	var u struct {
		Users []struct {
			Email string `json:"email"`
		} `json:"users"`
		Total int `json:"total"`
	}
	rec = get("/api/admin/users?sort=newest", &boss)
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil || u.Total != 2 || len(u.Users) != 2 {
		t.Fatalf("users: %s", rec.Body.String())
	}
}
