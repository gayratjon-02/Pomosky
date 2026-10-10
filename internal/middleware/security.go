package middleware

import "net/http"

// DefaultCSP is the Content-Security-Policy used by the server.
// TODO(stage 5/6): tighten once the frontend bundle and inline needs are known.
const DefaultCSP = "default-src 'self'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data: https://*.googleusercontent.com; connect-src 'self' ws: wss:; frame-ancestors 'none'; base-uri 'self'"

// SecurityHeaders sets conservative security headers on every response.
func SecurityHeaders(csp string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(self), camera=(), microphone=()")
			next.ServeHTTP(w, r)
		})
	}
}
