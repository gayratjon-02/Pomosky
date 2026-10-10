package models

import "time"

// UserID identifies a registered user.
type UserID string

// User is an optional registered account. Anonymous visitors have no User.
type User struct {
	ID                UserID
	Email             *string
	Provider          string // "google", "magiclink"
	ProviderSubject   string
	Nickname          string
	AvatarURL         string // profile picture from the login provider, may be empty
	ShowOnLeaderboard bool   // appear on the public leaderboard (default true)
	CreatedAt         time.Time
}

// Identity is who a login provider says the user is.
type Identity struct {
	Provider string // "google", "magiclink"
	Subject  string // stable id at the provider (the lowercase email for magic links)
	Email    string
	Name     string
	Picture  string // profile picture URL, may be empty
}

// Login is the result of a successful sign-in: the secret session token for
// the cookie and the user it belongs to.
type Login struct {
	Token   string
	Expires time.Time
	User    User
}
