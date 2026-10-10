package middleware

import (
	"net/http"
	"strings"
)

// VisitTracker records one usage event of a request (analytics.Collector implements it).
type VisitTracker interface {
	Track(r *http.Request, kind string)
}

// untracked paths are assets, APIs, sockets and probes: they are not page views.
var untracked = []string{"/static/", "/api/", "/ws/", "/__dev/", "/healthz", "/robots.txt", "/favicon"}

// TrackPageviews records a "pageview" for every successful HTML page. A nil
// tracker turns the middleware into a no-op.
func TrackPageviews(t VisitTracker) Middleware {
	return func(next http.Handler) http.Handler {
		if t == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || isUntracked(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.status == http.StatusOK && strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
				t.Track(r, "pageview")
			}
		})
	}
}

func isUntracked(p string) bool {
	for _, u := range untracked {
		if strings.HasPrefix(p, u) {
			return true
		}
	}
	return false
}
