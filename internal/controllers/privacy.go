package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

// Privacy serves GET /privacy.
func (c *Controllers) Privacy(w http.ResponseWriter, r *http.Request) {
	c.render(w, http.StatusOK, "privacy", views.PageData{Title: "Privacy"})
}
