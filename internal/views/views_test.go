package views

import (
	"github.com/gayratjon-02/Pomosky/internal/models"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

func TestEmbeddedTemplatesRender(t *testing.T) {
	r, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range []string{"home", "room", "host", "record", "login", "account", "privacy", "error", "admin"} {
		if !r.Has(page) {
			t.Errorf("missing page template %q", page)
			continue
		}
		rec := httptest.NewRecorder()
		data := any(RoomPage{ID: "abc", ShareURL: "http://x/r/abc"})
		switch page {
		case "account":
			data = AccountPage{StatsURL: "/api/stats"}
		case "login":
			data = LoginPage{Google: true, MagicLink: true}
		}
		r.Render(rec, http.StatusOK, page, PageData{Title: "Test", Data: data, User: &models.User{Nickname: "Ayla", ShowOnLeaderboard: true}})
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", page, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<title>") {
			t.Errorf("%s: no <title> in output", page)
		}
	}
}

func TestHeadCarriesShareTags(t *testing.T) {
	r, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	r.Render(rec, http.StatusOK, "privacy", PageData{Title: "Work together", Description: "Join this room", CanonicalURL: "https://pomosky.com/r/abc", OGImage: "https://pomosky.com/static/og.png"})
	body := rec.Body.String()
	for _, want := range []string{
		`property="og:title" content="Work together"`, `property="og:description" content="Join this room"`,
		`property="og:url" content="https://pomosky.com/r/abc"`, `property="og:image" content="https://pomosky.com/static/og.png"`,
		`name="twitter:card" content="summary_large_image"`, `rel="icon"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("head lacks %s", want)
		}
	}
}

func TestAssetFunctionUsesTheResolver(t *testing.T) {
	r, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	render := func() string {
		rec := httptest.NewRecorder()
		r.Render(rec, http.StatusOK, "privacy", PageData{Title: "T", Scripts: []string{"/static/landing.js"}})
		return rec.Body.String()
	}

	// Without a resolver the plain names are used.
	body := render()
	for _, want := range []string{`href="/static/base.css"`, `src="/static/site.js"`, `src="/static/landing.js"`} {
		if !strings.Contains(body, want) {
			t.Errorf("default output lacks %s", want)
		}
	}

	r.SetAssetResolver(func(name string) string {
		return "/static/" + strings.TrimSuffix(strings.TrimPrefix(name, "/static/"), path.Ext(name)) + ".HASH123" + path.Ext(name)
	})
	body = render()
	for _, want := range []string{`href="/static/base.HASH123.css"`, `src="/static/site.HASH123.js"`, `src="/static/landing.HASH123.js"`} {
		if !strings.Contains(body, want) {
			t.Errorf("resolved output lacks %s", want)
		}
	}
}
