package auth

import "context"

// Identity is who a login provider says the user is.
type Identity struct {
	Provider string
	Subject  string
	Email    string
	Name     string
}

// Provider is an OAuth-style login provider (Google, ...).
type Provider interface {
	Name() string
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (Identity, error)
}
