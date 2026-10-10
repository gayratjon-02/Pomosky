// Package auth implements the optional login: sessions, email magic links and
// OAuth providers. Anonymous use never depends on it.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// Provider name used for email sign-in.
const MagicLinkProvider = "magiclink"

const maxNicknameRunes = 24

// Options configures the Service.
type Options struct {
	BaseURL    string        // public base URL, used in the emailed link
	SessionTTL time.Duration // default 30 days
	LinkTTL    time.Duration // default 15 minutes
}

// Deps are the ports the Service needs. Mailer may be nil (magic links off).
type Deps struct {
	Users     ports.UserRepository
	Sessions  ports.SessionRepository
	Tokens    ports.LoginTokenRepository
	Mailer    ports.Mailer
	Clock     ports.Clock
	IDs       ports.IDGenerator
	Providers []Provider
}

// Service signs users in and out and resolves sessions.
type Service struct {
	d         Deps
	opts      Options
	providers map[string]Provider
}

var (
	_ ports.Accounts = (*Service)(nil)
	_ interface {
		UserFromSession(context.Context, string) (*models.User, error)
	} = (*Service)(nil)
)

// NewService creates the auth Service.
func NewService(d Deps, opts Options) *Service {
	if opts.SessionTTL <= 0 {
		opts.SessionTTL = 30 * 24 * time.Hour
	}
	if opts.LinkTTL <= 0 {
		opts.LinkTTL = 15 * time.Minute
	}
	providers := map[string]Provider{}
	for _, p := range d.Providers {
		providers[p.Name()] = p
	}
	return &Service{d: d, opts: opts, providers: providers}
}

// hash is what is stored in the database: a leaked table cannot be replayed
// as session cookies or sign-in links.
func hash(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// UserFromSession returns the user of a session cookie value, or an error if
// the session is unknown or expired. Callers treat any error as "anonymous".
func (s *Service) UserFromSession(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, apperr.ErrUnauthorized
	}
	sess, err := s.d.Sessions.ByID(ctx, hash(token))
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}
	if !sess.ExpiresAt.After(s.d.Clock.Now()) {
		_ = s.d.Sessions.Delete(ctx, sess.ID)
		return nil, apperr.ErrUnauthorized
	}
	u, err := s.d.Users.ByID(ctx, sess.UserID)
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}
	return &u, nil
}

// SignIn creates a session for the identity, creating the user on first login.
func (s *Service) SignIn(ctx context.Context, id models.Identity) (models.Login, error) {
	if id.Provider == "" || id.Subject == "" {
		return models.Login{}, apperr.ErrInvalid
	}
	user, err := s.d.Users.ByIdentity(ctx, id.Provider, id.Subject)
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		user, err = s.createUser(ctx, id)
		if err != nil {
			return models.Login{}, err
		}
	case err != nil:
		return models.Login{}, err
	}

	token := s.d.IDs.NewSessionID()
	expires := s.d.Clock.Now().Add(s.opts.SessionTTL)
	if err := s.d.Sessions.Create(ctx, models.Session{ID: hash(token), UserID: user.ID, ExpiresAt: expires}); err != nil {
		return models.Login{}, err
	}
	return models.Login{Token: token, Expires: expires, User: user}, nil
}

func (s *Service) createUser(ctx context.Context, id models.Identity) (models.User, error) {
	u := models.User{
		ID:              models.UserID(s.d.IDs.NewToken()),
		Provider:        id.Provider,
		ProviderSubject: id.Subject,
		Nickname:        nicknameFor(id),
		CreatedAt:       s.d.Clock.Now(),
	}
	if id.Email != "" {
		email := strings.ToLower(id.Email)
		u.Email = &email
	}
	if err := s.d.Users.Create(ctx, u); err != nil {
		// The same address already exists under another provider: that is a
		// different login method for the same person. Refusing is the safe
		// default (silently merging accounts needs verified emails on both sides).
		return models.User{}, err
	}
	return u, nil
}

func nicknameFor(id models.Identity) string {
	name := strings.TrimSpace(id.Name)
	if name == "" {
		name, _, _ = strings.Cut(id.Email, "@")
	}
	if name == "" {
		name = "Focuser"
	}
	if utf8.RuneCountInString(name) > maxNicknameRunes {
		name = string([]rune(name)[:maxNicknameRunes])
	}
	return name
}

// SignOut deletes a session. An unknown session is not an error.
func (s *Service) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.d.Sessions.Delete(ctx, hash(token))
}

// DeleteAccount permanently deletes the user (rooms, focus history and sessions cascade).
func (s *Service) DeleteAccount(ctx context.Context, id models.UserID) error {
	return s.d.Users.Delete(ctx, id)
}

// MagicLinkEnabled reports whether emails can be sent.
func (s *Service) MagicLinkEnabled() bool { return s.d.Mailer != nil && s.d.Tokens != nil }

// SendMagicLink emails a one-time sign-in link to the address.
func (s *Service) SendMagicLink(ctx context.Context, email string) error {
	if !s.MagicLinkEnabled() {
		return apperr.ErrNotImplemented
	}
	addr, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	token := s.d.IDs.NewToken()
	if err := s.d.Tokens.Create(ctx, hash(token), addr, s.d.Clock.Now().Add(s.opts.LinkTTL)); err != nil {
		return err
	}
	link := strings.TrimRight(s.opts.BaseURL, "/") + "/auth/callback?token=" + url.QueryEscape(token)
	body := fmt.Sprintf("Sign in to Pomosky:\n\n%s\n\nThe link works once and expires in %d minutes. If you did not ask for it, ignore this email.\n",
		link, int(s.opts.LinkTTL.Minutes()))
	return s.d.Mailer.Send(ctx, addr, "Your Pomosky sign-in link", body)
}

// CompleteMagicLink consumes the token and signs the address in.
func (s *Service) CompleteMagicLink(ctx context.Context, token string) (models.Login, error) {
	if !s.MagicLinkEnabled() || token == "" {
		return models.Login{}, apperr.ErrUnauthorized
	}
	email, err := s.d.Tokens.Consume(ctx, hash(token), s.d.Clock.Now())
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return models.Login{}, apperr.ErrUnauthorized
		}
		return models.Login{}, err
	}
	return s.SignIn(ctx, models.Identity{Provider: MagicLinkProvider, Subject: email, Email: email})
}

// ProviderEnabled reports whether an OAuth provider is configured.
func (s *Service) ProviderEnabled(name string) bool {
	_, ok := s.providers[name]
	return ok
}

// ProviderStart returns the provider URL and a fresh state value.
func (s *Service) ProviderStart(name string) (string, string, error) {
	p, ok := s.providers[name]
	if !ok {
		return "", "", apperr.ErrNotFound
	}
	state := s.d.IDs.NewToken()
	return p.AuthURL(state), state, nil
}

// CompleteProvider exchanges the callback code and signs the user in.
func (s *Service) CompleteProvider(ctx context.Context, name, code string) (models.Login, error) {
	p, ok := s.providers[name]
	if !ok || code == "" {
		return models.Login{}, apperr.ErrUnauthorized
	}
	id, err := p.Exchange(ctx, code)
	if err != nil {
		return models.Login{}, apperr.ErrUnauthorized
	}
	return s.SignIn(ctx, id)
}

func normalizeEmail(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 254 {
		return "", apperr.ErrInvalid
	}
	a, err := mail.ParseAddress(raw)
	if err != nil || a.Address != raw || !strings.Contains(a.Address[strings.LastIndex(a.Address, "@"):], ".") {
		return "", apperr.ErrInvalid
	}
	return strings.ToLower(a.Address), nil
}
