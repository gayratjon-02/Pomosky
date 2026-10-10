package models

import "time"

// UserID identifies a registered user.
type UserID string

// User is an optional registered account. Anonymous visitors have no User.
type User struct {
	ID              UserID
	Email           *string
	Provider        string // "google", "magiclink"
	ProviderSubject string
	Nickname        string
	CreatedAt       time.Time
}
