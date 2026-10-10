// Package google is the Google OAuth adapter for auth.Provider.
package google

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/auth"
)

// Provider implements auth.Provider for Google.
type Provider struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

var _ auth.Provider = (*Provider)(nil)

// New creates the Google provider.
func New(clientID, clientSecret, redirectURL string) *Provider {
	return &Provider{clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL}
}

// Name implements auth.Provider.
func (p *Provider) Name() string { return "google" }

// AuthURL implements auth.Provider. TODO(stage 4).
func (p *Provider) AuthURL(state string) string { return "" }

// Exchange implements auth.Provider. TODO(stage 4).
func (p *Provider) Exchange(ctx context.Context, code string) (auth.Identity, error) {
	return auth.Identity{}, apperr.ErrNotImplemented
}
