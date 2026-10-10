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

// Identity is who a login provider says the user is.
type Identity struct {
	Provider string // "google", "magiclink"
	Subject  string // stable id at the provider (the lowercase email for magic links)
	Email    string
	Name     string
}

// Login is the result of a successful sign-in: the secret session token for
// the cookie and the user it belongs to.
type Login struct {
	Token   string
	Expires time.Time
	User    User
}
