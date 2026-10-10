// Package google is the Google OAuth adapter for auth.Provider (authorization
// code flow). The endpoints come from Google's OpenID Connect discovery
// document; TODO(verify) them against https://accounts.google.com/.well-known/openid-configuration
// before going live.
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/models"
)

const (
	defaultAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultTokenURL    = "https://oauth2.googleapis.com/token"
	defaultUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
)

// Provider implements auth.Provider for Google.
type Provider struct {
	clientID     string
	clientSecret string
	redirectURL  string

	authURL     string
	tokenURL    string
	userInfoURL string
	http        *http.Client
}

var _ auth.Provider = (*Provider)(nil)

// New creates the Google provider. redirectURL must be registered in the
// Google console, e.g. https://pomosky.com/auth/callback?provider=google.
func New(clientID, clientSecret, redirectURL string) *Provider {
	return &Provider{
		clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL,
		authURL: defaultAuthURL, tokenURL: defaultTokenURL, userInfoURL: defaultUserInfoURL,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name implements auth.Provider.
func (p *Provider) Name() string { return "google" }

// AuthURL implements auth.Provider.
func (p *Provider) AuthURL(state string) string {
	q := url.Values{
		"client_id":     {p.clientID},
		"redirect_uri":  {p.redirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"prompt":        {"select_account"},
	}
	return p.authURL + "?" + q.Encode()
}

// Exchange implements auth.Provider. It only accepts verified email addresses.
func (p *Provider) Exchange(ctx context.Context, code string) (models.Identity, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"redirect_uri":  {p.redirectURL},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return models.Identity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := p.doJSON(req, &tok); err != nil {
		return models.Identity{}, fmt.Errorf("google: token exchange: %w", err)
	}
	if tok.AccessToken == "" {
		return models.Identity{}, errors.New("google: no access token")
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return models.Identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	var info struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := p.doJSON(req, &info); err != nil {
		return models.Identity{}, fmt.Errorf("google: userinfo: %w", err)
	}
	if info.Sub == "" || info.Email == "" || !info.EmailVerified {
		return models.Identity{}, errors.New("google: account without a verified email")
	}
	return models.Identity{Provider: "google", Subject: info.Sub, Email: info.Email, Name: info.Name}, nil
}

func (p *Provider) doJSON(req *http.Request, out any) error {
	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}
