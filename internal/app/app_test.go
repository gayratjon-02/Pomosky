package app

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/config"
)

func newDevApp(t *testing.T) *App {
	t.Helper()
	a, err := New(config.Config{
		Env:         "development",
		Dev:         true,
		DevViewsDir: "../views",
		StaticDir:   t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestDevReloadEndpointsOnlyWhenEnabled(t *testing.T) {
	prod, err := New(config.Config{Env: "development", StaticDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"/__dev/reload.js", "/__dev/events"} {
		rec := httptest.NewRecorder()
		prod.Handler().ServeHTTP(rec, httptest.NewRequest("GET", p, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("dev disabled: GET %s = %d, want 404", p, rec.Code)
		}
	}

	dev := newDevApp(t)
	rec := httptest.NewRecorder()
	dev.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/__dev/reload.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("dev enabled: reload.js = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	dev.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(rec.Body.String(), "/__dev/reload.js") {
		t.Fatal("dev enabled: the page does not include the reload script")
	}
}

// The event stream must work through the full middleware stack (logging wraps
// the ResponseWriter and has to keep it flushable).
func TestDevEventStreamThroughMiddleware(t *testing.T) {
	srv := httptest.NewServer(newDevApp(t).Handler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/__dev/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "retry:") {
		t.Fatalf("first stream line = %q, err = %v", line, err)
	}
}

func TestRoutesSmoke(t *testing.T) {
	a, err := New(config.Config{Env: "development", StaticDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	h := a.Handler()

	cases := []struct {
		method, path string
		want         int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/healthz", http.StatusOK},
		{"GET", "/privacy", http.StatusOK},
		{"GET", "/nope", http.StatusNotFound},
		{"GET", "/r/INVALID", http.StatusNotFound},
		{"GET", "/r/abcd2345", http.StatusNotImplemented},
		{"POST", "/r/new", http.StatusNotImplemented},
		{"GET", "/ws/abcd2345", http.StatusNotImplemented},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, tc.want)
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s %s: security headers missing", tc.method, tc.path)
		}
	}
}
