package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

// Account serves GET /account: the statistics page. A logged-in user sees
// their own numbers. Everyone else sees the same page filled with example data
// behind a prompt to log in, so the feature is visible before signing up.
func (c *Controllers) Account(w http.ResponseWriter, r *http.Request) {
	page := views.AccountPage{StatsURL: "/api/stats"}
	title := "Your statistics"
	if reqctx.User(r.Context()) == nil {
		page = views.AccountPage{StatsURL: "/api/stats/demo", Preview: true}
		title = "Statistics"
	}
	c.render(w, r, http.StatusOK, "account", views.PageData{
		Title:   title,
		Scripts: []string{"/static/account.js"},
		Data:    page,
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
