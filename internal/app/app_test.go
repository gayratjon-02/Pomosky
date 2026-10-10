package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/config"
)

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
