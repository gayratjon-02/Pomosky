package controllers

import (
	"net/http"
	"strings"
)

// Robots serves GET /robots.txt. Search engines may index the landing and
// privacy pages. Rooms (private links), accounts, the admin page and the APIs
// are kept out of search results.
func (c *Controllers) Robots(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString("User-agent: *\nAllow: /$\nAllow: /privacy\n")
	for _, p := range []string{"/r/", "/ws/", "/api/", "/admin", "/account", "/login", "/auth/", "/__dev/"} {
		b.WriteString("Disallow: " + p + "\n")
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(b.String()))
}
