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

	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
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

	// Anonymous visitors: the account page redirects, the API says 401.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/account", nil))
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Fatalf("anonymous /account = %d %s", rec.Code, rec.Header().Get("Location"))
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
