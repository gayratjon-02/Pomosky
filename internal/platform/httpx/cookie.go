package httpx

import (
	"net/http"
	"time"
)

// SessionCookieName is the name of the login session cookie.
const SessionCookieName = "pomosky_session"

// NewSessionCookie builds the session cookie with safe defaults.
func NewSessionCookie(value string, secure bool, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ExpiredSessionCookie returns a cookie that clears the session on the client.
func ExpiredSessionCookie(secure bool) *http.Cookie {
	c := NewSessionCookie("", secure, time.Unix(0, 0))
	c.MaxAge = -1
	return c
}
