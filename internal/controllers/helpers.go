package controllers

import (
	"net/http"
	"strings"

	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

// render fills the per-request page data (user, CSRF token) and renders the page.
func (c *Controllers) render(w http.ResponseWriter, r *http.Request, status int, page string, data views.PageData) {
	data.DevReload = c.deps.Config.Dev
	data.CSRFToken = reqctx.CSRF(r.Context())
	if data.OGImage == "" {
		data.OGImage = strings.TrimRight(c.deps.Config.BaseURL, "/") + "/static/og.png"
	}
	if cookie, err := r.Cookie(views.MoodCookie); err == nil && views.ValidMood(cookie.Value) {
		data.Mood = cookie.Value // so the page is painted in the chosen mood from the first frame
	}
	if data.User == nil {
		data.User = reqctx.User(r.Context())
	}
	c.deps.Views.Render(w, status, page, data)
}

// NotFound renders the 404 page. It is also the catch-all route.
func (c *Controllers) NotFound(w http.ResponseWriter, r *http.Request) {
	c.render(w, r, http.StatusNotFound, "error", views.PageData{
		Title:   "Page not found",
		Scripts: []string{"/static/ambient.js"},
		Data:    "We could not find what you were looking for.",
	})
}
