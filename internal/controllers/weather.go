package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

// The JSON shape of GET /api/weather (documented in docs/api.md).
type weatherJSON struct {
	CloudCoverPct    int     `json:"cloudCoverPct"`
	PrecipitationMMH float64 `json:"precipitationMmh"`
	SnowfallCMH      float64 `json:"snowfallCmh"`
	WindSpeedKMH     float64 `json:"windSpeedKmh"`
	WindDirDeg       int     `json:"windDirDeg"`
	FogDensity       float64 `json:"fogDensity"`
	Thunder          bool    `json:"thunder"`
	IsDay            bool    `json:"isDay"`
	TemperatureC     float64 `json:"temperatureC"`
	Code             int     `json:"code"`
	ObservedAtMS     int64   `json:"observedAtMs"`
	FetchedAtMS      int64   `json:"fetchedAtMs"`
}

func toWeatherJSON(v weather.View) weatherJSON {
	return weatherJSON{
		CloudCoverPct: v.CloudCoverPct, PrecipitationMMH: v.PrecipitationMMH, SnowfallCMH: v.SnowfallCMH,
		WindSpeedKMH: v.WindSpeedKMH, WindDirDeg: v.WindDirDeg, FogDensity: v.FogDensity, Thunder: v.Thunder,
		IsDay: v.IsDay, TemperatureC: v.TemperatureC, Code: v.Code,
		ObservedAtMS: v.ObservedAt.UnixMilli(), FetchedAtMS: v.FetchedAt.UnixMilli(),
	}
}

// WeatherAPI serves GET /api/weather?room={id}: the cached weather at the
// host's place of a live room. Viewers never send their own location.
func (c *Controllers) WeatherAPI(w http.ResponseWriter, r *http.Request) {
	if c.deps.Weather == nil || c.deps.Live == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	id := room.ID(r.URL.Query().Get("room"))
	if !id.Valid() {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	loc, ok := c.deps.Live.Location(id)
	if !ok || (loc.LatRounded == 0 && loc.LonRounded == 0 && loc.TZ == "") {
		httpx.WriteError(w, apperr.ErrNotFound) // no such room, or the host has not set a place yet
		return
	}
	snap, err := c.deps.Weather.Current(r.Context(), loc.LatRounded, loc.LonRounded)
	if err != nil {
		c.deps.Log.Warn("weather lookup failed", "room", id, "err", err)
		httpx.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "weather is not available right now"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	httpx.WriteJSON(w, http.StatusOK, toWeatherJSON(snap.View()))
}
