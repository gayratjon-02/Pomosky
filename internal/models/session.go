package models

import "time"

// Session is a login session referenced by the session cookie.
type Session struct {
	ID        string
	UserID    UserID
	CreatedAt time.Time
	ExpiresAt time.Time
}
