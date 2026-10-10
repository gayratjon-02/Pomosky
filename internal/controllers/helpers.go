package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

func (c *Controllers) render(w http.ResponseWriter, status int, page string, data views.PageData) {
	data.DevReload = c.deps.Config.Dev
	c.deps.Views.Render(w, status, page, data)
}

// NotFound renders the 404 page. It is also the catch-all route.
func (c *Controllers) NotFound(w http.ResponseWriter, r *http.Request) {
	c.render(w, http.StatusNotFound, "error", views.PageData{
		Title: "Page not found",
		Data:  "We could not find what you were looking for.",
	})
}
