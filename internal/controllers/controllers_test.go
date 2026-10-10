package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

func newTestControllers(t *testing.T) *Controllers {
	t.Helper()
	v, err := views.Default()
	if err != nil {
		t.Fatal(err)
	}
	return New(Deps{Views: v})
}

func TestHomeAndHealth(t *testing.T) {
	c := newTestControllers(t)

	rec := httptest.NewRecorder()
	c.Home(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Pomosky") {
		t.Fatalf("home: %d %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	c.Health(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("health: %d %q", rec.Code, rec.Body.String())
	}
}

func TestDevReloadScriptOnlyInDev(t *testing.T) {
	v, err := views.Default()
	if err != nil {
		t.Fatal(err)
	}
	const script = "/__dev/reload.js"

	for _, dev := range []bool{false, true} {
		c := New(Deps{Views: v, Config: config.Config{Dev: dev}})
		rec := httptest.NewRecorder()
		c.Home(rec, httptest.NewRequest("GET", "/", nil))
		if got := strings.Contains(rec.Body.String(), script); got != dev {
			t.Errorf("Dev=%v: page contains reload script = %v", dev, got)
		}
	}
}

func TestRoomShowRejectsInvalidID(t *testing.T) {
	c := newTestControllers(t)
	req := httptest.NewRequest("GET", "/r/INVALID", nil)
	req.SetPathValue("id", "INVALID")
	rec := httptest.NewRecorder()
	c.RoomShow(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
