package auth

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

type fakeUsers struct{ byID map[models.UserID]models.User }

func (f *fakeUsers) Create(_ context.Context, u models.User) error {
	for _, e := range f.byID {
		if e.Provider == u.Provider && e.ProviderSubject == u.ProviderSubject {
			return apperr.ErrConflict
		}
	}
	f.byID[u.ID] = u
	return nil
}
func (f *fakeUsers) ByID(_ context.Context, id models.UserID) (models.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return models.User{}, apperr.ErrNotFound
}
func (f *fakeUsers) ByEmail(_ context.Context, email string) (models.User, error) {
	for _, u := range f.byID {
		if u.Email != nil && *u.Email == email {
			return u, nil
		}
	}
	return models.User{}, apperr.ErrNotFound
}
func (f *fakeUsers) ByIdentity(_ context.Context, p, s string) (models.User, error) {
	for _, u := range f.byID {
		if u.Provider == p && u.ProviderSubject == s {
			return u, nil
		}
	}
	return models.User{}, apperr.ErrNotFound
}
func (f *fakeUsers) Delete(_ context.Context, id models.UserID) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeUsers) UpdateAvatar(_ context.Context, id models.UserID, url string) error {
	u := f.byID[id]
	u.AvatarURL = url
	f.byID[id] = u
	return nil
}

type fakeSessions struct{ byID map[string]models.Session }

func (f *fakeSessions) Create(_ context.Context, s models.Session) error {
	f.byID[s.ID] = s
	return nil
}
func (f *fakeSessions) ByID(_ context.Context, id string) (models.Session, error) {
	if s, ok := f.byID[id]; ok {
		return s, nil
	}
	return models.Session{}, apperr.ErrNotFound
}
func (f *fakeSessions) Delete(_ context.Context, id string) error               { delete(f.byID, id); return nil }
func (f *fakeSessions) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

type fakeTokens struct {
	byHash map[string]struct {
		email   string
		expires time.Time
	}
}

func (f *fakeTokens) Create(_ context.Context, h, email string, exp time.Time) error {
	f.byHash[h] = struct {
		email   string
		expires time.Time
	}{email, exp}
	return nil
}
func (f *fakeTokens) Consume(_ context.Context, h string, now time.Time) (string, error) {
	t, ok := f.byHash[h]
	if !ok || !t.expires.After(now) {
		return "", apperr.ErrNotFound
	}
	delete(f.byHash, h)
	return t.email, nil
}

func (f *fakeTokens) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

type fakeMailer struct{ to, body string }

func (m *fakeMailer) Send(_ context.Context, to, _, body string) error {
	m.to, m.body = to, body
	return nil
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type seqIDs struct{ n int }

func (s *seqIDs) NewRoomID() room.ID { return "aaaaaaaa" }
func (s *seqIDs) NewToken() string {
	s.n++
	return "token" + string(rune('a'+s.n)) + strings.Repeat("x", 20)
}
func (s *seqIDs) NewSessionID() string { return s.NewToken() }

func newTestService() (*Service, *fakeSessions, *fakeMailer, *fakeClock) {
	clk := &fakeClock{now: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)}
	sessions := &fakeSessions{byID: map[string]models.Session{}}
	mailer := &fakeMailer{}
	svc := NewService(Deps{
		Users:    &fakeUsers{byID: map[models.UserID]models.User{}},
		Sessions: sessions,
		Tokens: &fakeTokens{byHash: map[string]struct {
			email   string
			expires time.Time
		}{}},
		Mailer: mailer, Clock: clk, IDs: &seqIDs{},
	}, Options{BaseURL: "https://pomosky.test"})
	return svc, sessions, mailer, clk
}

func TestSignInCreatesUserOnceAndResolvesSession(t *testing.T) {
	svc, sessions, _, _ := newTestService()
	ctx := context.Background()
	id := models.Identity{Provider: "google", Subject: "g1", Email: "Ayla@Example.com", Name: "Ayla"}

	first, err := svc.SignIn(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.SignIn(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if first.User.ID != second.User.ID {
		t.Fatal("a second login created a second user")
	}
	if first.User.Email == nil || *first.User.Email != "ayla@example.com" || first.User.Nickname != "Ayla" {
		t.Fatalf("user = %+v", first.User)
	}
	if _, stored := sessions.byID[first.Token]; stored {
		t.Fatal("the raw session token must not be stored")
	}
	u, err := svc.UserFromSession(ctx, first.Token)
	if err != nil || u.ID != first.User.ID {
		t.Fatalf("UserFromSession: %+v %v", u, err)
	}
	if _, err := svc.UserFromSession(ctx, "unknown"); err == nil {
		t.Fatal("unknown token accepted")
	}
}

func TestSessionExpiresAndSignOut(t *testing.T) {
	svc, _, _, clk := newTestService()
	ctx := context.Background()
	login, _ := svc.SignIn(ctx, models.Identity{Provider: "google", Subject: "g1"})

	if err := svc.SignOut(ctx, login.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UserFromSession(ctx, login.Token); err == nil {
		t.Fatal("session usable after sign out")
	}

	login, _ = svc.SignIn(ctx, models.Identity{Provider: "google", Subject: "g1"})
	clk.now = clk.now.Add(31 * 24 * time.Hour)
	if _, err := svc.UserFromSession(ctx, login.Token); err == nil {
		t.Fatal("expired session accepted")
	}
}

func TestMagicLinkFlow(t *testing.T) {
	svc, _, mailer, clk := newTestService()
	ctx := context.Background()

	if err := svc.SendMagicLink(ctx, "not an email"); !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("bad address: %v", err)
	}
	if err := svc.SendMagicLink(ctx, "  Ayla@Example.com "); err != nil {
		t.Fatal(err)
	}
	if mailer.to != "ayla@example.com" {
		t.Fatalf("mail to %q", mailer.to)
	}
	var link string
	for _, line := range strings.Split(mailer.body, "\n") {
		if strings.HasPrefix(line, "https://pomosky.test/auth/callback") {
			link = line
		}
	}
	u, err := url.Parse(link)
	if err != nil || u.Query().Get("token") == "" {
		t.Fatalf("link = %q", link)
	}
	token := u.Query().Get("token")

	login, err := svc.CompleteMagicLink(ctx, token)
	if err != nil || login.User.Provider != "magiclink" || login.User.Nickname != "ayla" {
		t.Fatalf("CompleteMagicLink: %+v %v", login.User, err)
	}
	if _, err := svc.CompleteMagicLink(ctx, token); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("link worked twice: %v", err)
	}

	_ = svc.SendMagicLink(ctx, "b@example.com")
	clk.now = clk.now.Add(16 * time.Minute)
	for _, line := range strings.Split(mailer.body, "\n") {
		if strings.HasPrefix(line, "https://") {
			link = line
		}
	}
	u, _ = url.Parse(link)
	if _, err := svc.CompleteMagicLink(ctx, u.Query().Get("token")); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("expired link accepted: %v", err)
	}
}

func TestMagicLinkDisabledWithoutMailer(t *testing.T) {
	svc := NewService(Deps{}, Options{})
	if svc.MagicLinkEnabled() {
		t.Fatal("enabled without a mailer")
	}
	if err := svc.SendMagicLink(context.Background(), "a@example.com"); !errors.Is(err, apperr.ErrNotImplemented) {
		t.Fatalf("err = %v", err)
	}
}

func TestAvatarIsStoredFilteredAndRefreshed(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()
	id := models.Identity{Provider: "google", Subject: "g1", Email: "a@example.com", Name: "Ayla", Picture: "https://lh3.googleusercontent.com/a/abc=s96-c"}

	first, err := svc.SignIn(ctx, id)
	if err != nil || first.User.AvatarURL != id.Picture {
		t.Fatalf("first login: %+v %v", first.User, err)
	}
	// The picture changed at the provider: the next login refreshes it.
	id.Picture = "https://lh3.googleusercontent.com/a/new=s96-c"
	second, _ := svc.SignIn(ctx, id)
	if second.User.AvatarURL != id.Picture {
		t.Fatalf("avatar not refreshed: %q", second.User.AvatarURL)
	}
	// Only https URLs are kept.
	for _, bad := range []string{"http://example.com/a.png", "javascript:alert(1)", "data:image/png;base64,AAAA", "/relative.png", strings.Repeat("a", 600)} {
		if got := safeAvatarURL(bad); got != "" {
			t.Errorf("safeAvatarURL(%q) = %q, want empty", bad[:min(len(bad), 30)], got)
		}
	}
}
