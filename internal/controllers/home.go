package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

// Home serves GET /.
func (c *Controllers) Home(w http.ResponseWriter, r *http.Request) {
	c.render(w, http.StatusOK, "home", views.PageData{
		Title:       "Work together under a live sky",
		Description: "Pomosky is a pomodoro room where the sky behind you is the real weather and sun at the host's location.",
	})
}
