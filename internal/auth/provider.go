package auth

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

// Provider is an OAuth-style login provider (Google, ...).
type Provider interface {
	Name() string
	// AuthURL is where the browser is sent to sign in; state comes back unchanged.
	AuthURL(state string) string
	// Exchange trades the callback code for the identity of the user.
	Exchange(ctx context.Context, code string) (models.Identity, error)
}
