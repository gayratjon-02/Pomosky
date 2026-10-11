package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

// Home serves GET /.
func (c *Controllers) Home(w http.ResponseWriter, r *http.Request) {
	if c.isAdminHost(r) { // the operator host has no landing page
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	c.render(w, r, http.StatusOK, "home", views.PageData{
		Title:       "Work together under a live sky",
		Description: "Pomosky is a pomodoro room where the sky behind you is the real weather and sun at the host's location.",
		Scripts:     []string{"/static/landing.js"},
	})
}
