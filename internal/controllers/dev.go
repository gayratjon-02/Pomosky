package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

// SkyLab serves GET /__dev/sky: a development-only playground to try the sky
// with sliders for time, place and weather. It is registered only in dev mode.
func (c *Controllers) SkyLab(w http.ResponseWriter, r *http.Request) {
	c.render(w, r, http.StatusOK, "skylab", views.PageData{
		Title:   "Sky lab",
		Scripts: []string{"/static/skylab.js"},
	})
}
