package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

// Account serves GET /account: profile and statistics. Anonymous visitors are
// sent to the login page.
func (c *Controllers) Account(w http.ResponseWriter, r *http.Request) {
	if reqctx.User(r.Context()) == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	c.render(w, r, http.StatusOK, "account", views.PageData{
		Title:   "Your account",
		Scripts: []string{"/static/account.js"},
		Data:    views.AccountPage{StatsURL: "/api/stats"},
	})
}

// AccountDelete serves POST /account/delete. The form must carry confirm=delete
// so a stray click cannot erase an account.
func (c *Controllers) AccountDelete(w http.ResponseWriter, r *http.Request) {
	user := reqctx.User(r.Context())
	if user == nil || c.deps.Accounts == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.PostFormValue("confirm") != "delete" {
		http.Error(w, "type delete to confirm", http.StatusBadRequest)
		return
	}
	if err := c.deps.Accounts.DeleteAccount(r.Context(), user.ID); err != nil {
		c.deps.Log.Error("delete account", "err", err)
		http.Error(w, "could not delete the account", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, httpx.ExpiredSessionCookie(c.secure()))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
