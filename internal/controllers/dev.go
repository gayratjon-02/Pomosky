package controllers

import (
	"net/http"
	"strconv"
	"strings"

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

// MobileLab serves GET /__dev/mobile?path=/account&w=390&h=844: a page shown in
// a phone-sized frame, to check the layout at that width. Development only.
func (c *Controllers) MobileLab(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\\r\n") {
		path = "/"
	}
	size := func(name string, def, lo, hi int) int {
		n, err := strconv.Atoi(r.URL.Query().Get(name))
		if err != nil || n < lo || n > hi {
			return def
		}
		return n
	}
	c.render(w, r, http.StatusOK, "mobilelab", views.PageData{
		Title: "Mobile preview",
		Data:  views.MobileLab{Path: path, Width: size("w", 390, 280, 1200), Height: size("h", 844, 400, 2000)},
	})
}
