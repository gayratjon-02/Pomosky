package middleware

import (
	"net/http"
	"strings"
)

// DefaultCSP is the Content-Security-Policy used by the server.
// TODO(stage 5/6): tighten once the frontend bundle and inline needs are known.
const DefaultCSP = "default-src 'self'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data: https://*.googleusercontent.com; connect-src 'self' ws: wss:; frame-ancestors 'none'; base-uri 'self'"

// DevCSP is DefaultCSP for development: pages may be framed by the same site,
// which the mobile preview (/__dev/mobile) needs. Never used in production.
var DevCSP = strings.Replace(DefaultCSP, "frame-ancestors 'none'", "frame-ancestors 'self'", 1)

// SecurityHeaders sets conservative security headers on every response.
// With frameSameOrigin the site may frame itself (development preview only).
func SecurityHeaders(csp string, frameSameOrigin ...bool) Middleware {
	frame := "DENY"
	if len(frameSameOrigin) > 0 && frameSameOrigin[0] {
		frame = "SAMEORIGIN"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", frame)
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(self), camera=(), microphone=()")
			next.ServeHTTP(w, r)
		})
	}
}
