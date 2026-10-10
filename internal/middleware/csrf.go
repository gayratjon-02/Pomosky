package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

// Names used by the double-submit CSRF protection.
const (
	CSRFCookieName = "pomosky_csrf"
	CSRFFormField  = "csrf"
	CSRFHeader     = "X-CSRF-Token"
)

const maxFormBytes = 64 << 10

// CSRF protects state-changing requests with the double-submit cookie pattern:
// every visitor gets a random token in an HttpOnly cookie, forms carry the same
// value in a hidden field (see reqctx.CSRF), and a POST must send both and they
// must match. A cross-site page can make the browser send the cookie but cannot
// read it, so it cannot build a matching form. A mismatching Origin header is
// rejected as an additional check.
func CSRF(secure bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if c, err := r.Cookie(CSRFCookieName); err == nil && validCSRFToken(c.Value) {
				token = c.Value
			} else {
				token = newCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name: CSRFCookieName, Value: token, Path: "/",
					HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
				})
			}
			r = r.WithContext(reqctx.WithCSRF(r.Context(), token))

			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			if origin := r.Header.Get("Origin"); origin != "" && origin != "null" {
				if u, err := url.Parse(origin); err != nil || u.Host != r.Host {
					http.Error(w, "cross-origin request rejected", http.StatusForbidden)
					return
				}
			}
			submitted := r.Header.Get(CSRFHeader)
			if submitted == "" {
				r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
				if err := r.ParseForm(); err != nil {
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				submitted = r.PostForm.Get(CSRFFormField)
			}
			if subtle.ConstantTimeCompare([]byte(submitted), []byte(token)) != 1 {
				http.Error(w, "invalid CSRF token, reload the page and try again", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func newCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: no randomness available: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func validCSRFToken(s string) bool {
	return len(s) == 43 // 32 bytes, unpadded base64url
}
