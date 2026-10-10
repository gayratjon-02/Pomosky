package controllers

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

const (
	oauthStateCookie = "pomosky_oauth_state"
	oauthStateTTL    = 10 * time.Minute
	googleProvider   = "google"
	magicProvider    = "magiclink"
)

func (c *Controllers) secure() bool { return c.deps.Config.IsProduction() }

func (c *Controllers) loginPage(w http.ResponseWriter, r *http.Request, status int, flash, confirmToken string) {
	page := views.LoginPage{ConfirmToken: confirmToken}
	if c.deps.Accounts != nil {
		page.Google = c.deps.Accounts.ProviderEnabled(googleProvider)
		page.MagicLink = c.deps.Accounts.MagicLinkEnabled()
	}
	c.render(w, r, status, "login", views.PageData{
		Title:   "Log in",
		Flash:   flash,
		Scripts: []string{"/static/ambient.js"},
		Data:    page,
	})
}

// startSession sets the session cookie and sends the user to their account.
func (c *Controllers) startSession(w http.ResponseWriter, r *http.Request, login models.Login) {
	http.SetCookie(w, httpx.NewSessionCookie(login.Token, c.secure(), login.Expires))
	c.track(r, models.VisitLoginDone)
	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

// Login serves GET /login.
func (c *Controllers) Login(w http.ResponseWriter, r *http.Request) {
	if reqctx.User(r.Context()) != nil {
		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}
	c.loginPage(w, r, http.StatusOK, "", "")
}

// LoginSubmit serves POST /login: it starts the Google flow or emails a magic link.
func (c *Controllers) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if c.deps.Accounts == nil {
		c.loginPage(w, r, http.StatusServiceUnavailable, "Login is not available right now. You can still use Pomosky without an account.", "")
		return
	}
	switch r.PostFormValue("provider") {
	case googleProvider:
		url, state, err := c.deps.Accounts.ProviderStart(googleProvider)
		if err != nil {
			c.loginPage(w, r, http.StatusNotFound, "Google sign-in is not available.", "")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name: oauthStateCookie, Value: state, Path: "/", MaxAge: int(oauthStateTTL.Seconds()),
			HttpOnly: true, Secure: c.secure(), SameSite: http.SameSiteLaxMode,
		})
		c.track(r, models.VisitLoginStarted)
		http.Redirect(w, r, url, http.StatusSeeOther)
	case magicProvider:
		err := c.deps.Accounts.SendMagicLink(r.Context(), r.PostFormValue("email"))
		switch {
		case errors.Is(err, apperr.ErrInvalid):
			c.loginPage(w, r, http.StatusBadRequest, "That does not look like an email address.", "")
		case errors.Is(err, apperr.ErrNotImplemented):
			c.loginPage(w, r, http.StatusServiceUnavailable, "Email sign-in is not available right now.", "")
		case err != nil:
			c.deps.Log.Error("send magic link", "err", err)
			c.loginPage(w, r, http.StatusInternalServerError, "We could not send the email. Please try again later.", "")
		default:
			c.track(r, models.VisitLoginStarted)
			// The same message whether or not the address has an account.
			c.loginPage(w, r, http.StatusOK, "If that address is valid, a sign-in link is on its way. It works once and expires in 15 minutes.", "")
		}
	default:
		c.loginPage(w, r, http.StatusBadRequest, "Choose how you want to log in.", "")
	}
}

// AuthCallback serves GET /auth/callback: the page behind the emailed link.
// It only shows a confirmation button: mail scanners often open links, and the
// one-time token must not be spent by a scanner (the button posts it, see AuthConfirm).
func (c *Controllers) AuthCallback(w http.ResponseWriter, r *http.Request) {
	if c.deps.Accounts == nil {
		c.loginPage(w, r, http.StatusServiceUnavailable, "Login is not available right now.", "")
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		c.loginPage(w, r, http.StatusBadRequest, "That sign-in link is not valid.", "")
		return
	}
	c.loginPage(w, r, http.StatusOK, "", token)
}

// GoogleCallback serves GET /auth/google/callback: Google returns here with
// code and state. The state must match the cookie set when the flow started.
func (c *Controllers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if c.deps.Accounts == nil || !c.deps.Accounts.ProviderEnabled(googleProvider) {
		c.loginPage(w, r, http.StatusServiceUnavailable, "Google sign-in is not available.", "")
		return
	}
	q := r.URL.Query()
	cookie, err := r.Cookie(oauthStateCookie)
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: c.secure(), SameSite: http.SameSiteLaxMode})
	if err != nil || cookie.Value == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(q.Get("state"))) != 1 {
		c.loginPage(w, r, http.StatusBadRequest, "The sign-in session expired. Please try again.", "")
		return
	}
	login, err := c.deps.Accounts.CompleteProvider(r.Context(), googleProvider, q.Get("code"))
	if err != nil {
		c.loginPage(w, r, http.StatusUnauthorized, "Google sign-in failed. Please try again.", "")
		return
	}
	c.startSession(w, r, login)
}

// AuthConfirm serves POST /auth/callback: it spends the magic-link token.
func (c *Controllers) AuthConfirm(w http.ResponseWriter, r *http.Request) {
	if c.deps.Accounts == nil {
		c.loginPage(w, r, http.StatusServiceUnavailable, "Login is not available right now.", "")
		return
	}
	login, err := c.deps.Accounts.CompleteMagicLink(r.Context(), r.PostFormValue("token"))
	if err != nil {
		c.loginPage(w, r, http.StatusUnauthorized, "That sign-in link has expired or was already used. Request a new one.", "")
		return
	}
	c.startSession(w, r, login)
}

// Logout serves POST /logout.
func (c *Controllers) Logout(w http.ResponseWriter, r *http.Request) {
	if c.deps.Accounts != nil {
		if cookie, err := r.Cookie(httpx.SessionCookieName); err == nil {
			if err := c.deps.Accounts.SignOut(r.Context(), cookie.Value); err != nil {
				c.deps.Log.Error("sign out", "err", err)
			}
		}
	}
	http.SetCookie(w, httpx.ExpiredSessionCookie(c.secure()))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
