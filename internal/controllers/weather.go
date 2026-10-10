package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
)

// WeatherAPI serves GET /api/weather?room={id}: the cached snapshot for the
// location of a room. TODO(stage 6).
func (c *Controllers) WeatherAPI(w http.ResponseWriter, r *http.Request) {
	httpx.NotImplemented(w)
}
