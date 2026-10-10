package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "handler") }),
		mk("a"), mk("b"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if got := strings.Join(order, ","); got != "a,b,handler" {
		t.Fatalf("order = %s", got)
	}
}

func TestRecoverReturns500(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }), Recover(log))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

type failingResolver struct{}

func (failingResolver) UserFromSession(context.Context, string) (*models.User, error) {
	return nil, context.Canceled
}

func TestOptionalAuthNeverBlocks(t *testing.T) {
	reached := false
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		if UserFromContext(r.Context()) != nil {
			t.Error("user should be anonymous")
		}
	}), OptionalAuth(failingResolver{}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "pomosky_session", Value: "bad"})
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !reached {
		t.Fatal("OptionalAuth blocked the request")
	}
}

func csrfRoundTrip(t *testing.T, mutate func(r *http.Request, token string)) int {
	t.Helper()
	var seen string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = reqctx.CSRF(r.Context()) }), CSRF(false))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != CSRFCookieName || seen != cookies[0].Value {
		t.Fatalf("GET should set the csrf cookie and expose the token: %v %q", cookies, seen)
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(""))
	req.Host = "example.com"
	req.AddCookie(cookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mutate(req, cookies[0].Value)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestCSRFAcceptsMatchingFormToken(t *testing.T) {
	code := csrfRoundTrip(t, func(r *http.Request, token string) {
		r.Body = io.NopCloser(strings.NewReader("csrf=" + token))
	})
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
}

func TestCSRFAcceptsHeaderToken(t *testing.T) {
	code := csrfRoundTrip(t, func(r *http.Request, token string) { r.Header.Set(CSRFHeader, token) })
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
}

func TestCSRFRejectsMissingOrWrongToken(t *testing.T) {
	if code := csrfRoundTrip(t, func(*http.Request, string) {}); code != http.StatusForbidden {
		t.Fatalf("missing token: status = %d", code)
	}
	code := csrfRoundTrip(t, func(r *http.Request, _ string) { r.Header.Set(CSRFHeader, "wrong") })
	if code != http.StatusForbidden {
		t.Fatalf("wrong token: status = %d", code)
	}
}

func TestCSRFRejectsForeignOrigin(t *testing.T) {
	code := csrfRoundTrip(t, func(r *http.Request, token string) {
		r.Header.Set(CSRFHeader, token)
		r.Header.Set("Origin", "https://evil.example")
	})
	if code != http.StatusForbidden {
		t.Fatalf("status = %d", code)
	}
	code = csrfRoundTrip(t, func(r *http.Request, token string) {
		r.Header.Set(CSRFHeader, token)
		r.Header.Set("Origin", "https://example.com")
	})
	if code != http.StatusOK {
		t.Fatalf("same-origin status = %d", code)
	}
}

type fakeTracker struct{ kinds, paths []string }

func (f *fakeTracker) Track(r *http.Request, kind string) {
	f.kinds = append(f.kinds, kind)
	f.paths = append(f.paths, r.URL.Path)
}

func TestTrackPageviewsCountsOnlyHTMLPages(t *testing.T) {
	tr := &fakeTracker{}
	h := TrackPageviews(tr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
		case "/api/stats":
			w.Header().Set("Content-Type", "application/json")
		default:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
	}))
	for _, c := range []struct{ method, path string }{
		{"GET", "/"}, {"GET", "/privacy"}, {"GET", "/missing"}, {"GET", "/api/stats"},
		{"GET", "/static/site.js"}, {"POST", "/login"}, {"GET", "/healthz"},
	} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(c.method, c.path, nil))
	}
	if len(tr.paths) != 2 || tr.paths[0] != "/" || tr.paths[1] != "/privacy" || tr.kinds[0] != "pageview" {
		t.Fatalf("tracked %v", tr.paths)
	}
}

func TestTrackPageviewsWithoutTrackerIsTransparent(t *testing.T) {
	called := false
	h := TrackPageviews(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Fatal("handler not called")
	}
}
