package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/clock"
	"github.com/gayratjon-02/Pomosky/internal/room"
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

// ---- rooms

type fakeLive struct{ rooms map[room.ID]room.State }

func (f *fakeLive) Register(s room.State, _ models.UserID) {
	if f.rooms == nil {
		f.rooms = map[room.ID]room.State{}
	}
	f.rooms[s.ID] = s
}

func (f *fakeLive) Has(id room.ID) bool { _, ok := f.rooms[id]; return ok }

func (f *fakeLive) Location(id room.ID) (room.Location, bool) {
	s, ok := f.rooms[id]
	return s.Location, ok
}

type fakeIDs struct{ roomIDs []room.ID }

func (f *fakeIDs) NewRoomID() room.ID {
	id := f.roomIDs[0]
	if len(f.roomIDs) > 1 {
		f.roomIDs = f.roomIDs[1:]
	}
	return id
}
func (f *fakeIDs) NewToken() string     { return "secret-token" }
func (f *fakeIDs) NewSessionID() string { return "session" }

type fakeHasher struct{}

func (fakeHasher) Hash(token string) string { return "hash:" + token }

func newRoomControllers(t *testing.T, live *fakeLive, ids ...room.ID) *Controllers {
	t.Helper()
	v, err := views.Default()
	if err != nil {
		t.Fatal(err)
	}
	return New(Deps{
		Views:      v,
		Config:     config.Config{BaseURL: "https://pomosky.example/"},
		Live:       live,
		HostTokens: fakeHasher{},
		Clock:      clock.Fixed{T: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)},
		IDs:        &fakeIDs{roomIDs: ids},
	})
}

func TestRoomCreateRegistersARoomAndRedirectsToTheHostPanel(t *testing.T) {
	live := &fakeLive{}
	c := newRoomControllers(t, live, "abcd2345")

	rec := httptest.NewRecorder()
	c.RoomCreate(rec, httptest.NewRequest("POST", "/r/new", nil))

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/r/abcd2345/host#token=secret-token" {
		t.Fatalf("Location = %q", loc)
	}
	st, ok := live.rooms["abcd2345"]
	if !ok {
		t.Fatal("the room was not registered")
	}
	if st.HostTokenHash != "hash:secret-token" {
		t.Fatalf("host token hash = %q, want the hashed form (never the raw token)", st.HostTokenHash)
	}
}

func TestRoomCreateSkipsTakenIDs(t *testing.T) {
	live := &fakeLive{}
	live.Register(room.New("abcd2345", "x", time.Now()), "")
	c := newRoomControllers(t, live, "abcd2345", "efgh6723")

	rec := httptest.NewRecorder()
	c.RoomCreate(rec, httptest.NewRequest("POST", "/r/new", nil))

	if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, "/r/efgh6723/host") {
		t.Fatalf("Location = %q, want the second candidate", loc)
	}
}

func TestRoomCreateNeedsItsDependencies(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestControllers(t).RoomCreate(rec, httptest.NewRequest("POST", "/r/new", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", rec.Code)
	}
}

func TestRoomPages(t *testing.T) {
	live := &fakeLive{}
	live.Register(room.New("abcd2345", "hash", time.Now()), "")
	c := newRoomControllers(t, live)

	cases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		url     string
		id      string
		want    int
		contain string
	}{
		{"viewer page", c.RoomShow, "/r/abcd2345", "abcd2345", http.StatusOK, "/static/room.js"},
		{"record mode", c.RoomShow, "/r/abcd2345?mode=record", "abcd2345", http.StatusOK, "/static/record.js"},
		{"host panel", c.HostPanel, "/r/abcd2345/host", "abcd2345", http.StatusOK, "/static/host.js"},
		{"viewer page of a missing room", c.RoomShow, "/r/zzzzzzzz", "zzzzzzzz", http.StatusNotFound, "not live"},
		{"viewer page with an invalid id", c.RoomShow, "/r/BAD", "BAD", http.StatusNotFound, "not live"},
		{"host panel of a missing room", c.HostPanel, "/r/zzzzzzzz/host", "zzzzzzzz", http.StatusNotFound, "not live"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			req.SetPathValue("id", tc.id)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != tc.want || !strings.Contains(rec.Body.String(), tc.contain) {
				t.Fatalf("%d, want %d containing %q; body: %.200s", rec.Code, tc.want, tc.contain, rec.Body.String())
			}
		})
	}

	// The canonical link uses the configured base URL without a double slash.
	req := httptest.NewRequest("GET", "/r/abcd2345", nil)
	req.SetPathValue("id", "abcd2345")
	rec := httptest.NewRecorder()
	c.RoomShow(rec, req)
	if !strings.Contains(rec.Body.String(), `href="https://pomosky.example/r/abcd2345"`) {
		t.Fatal("canonical link is wrong")
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

func TestPagesDefaultToTheShareImage(t *testing.T) {
	c := newTestControllers(t)
	c.deps.Config = config.Config{BaseURL: "https://pomosky.com/"}
	rec := httptest.NewRecorder()
	c.Home(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(rec.Body.String(), `property="og:image" content="https://pomosky.com/static/og.png"`) {
		t.Fatalf("no default share image in:\n%s", rec.Body.String())
	}
}
