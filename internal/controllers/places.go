package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

const (
	maxPlaceQueryRunes = 60
	placeResults       = 6
)

// PlaceSearcher is what the controllers need from the geocoding layer.
type PlaceSearcher = weather.Geocoder

type placeJSON struct {
	Name    string  `json:"name"`
	Region  string  `json:"region,omitempty"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	TZ      string  `json:"tz"`
}

// PlacesAPI serves GET /api/places?q=...: place search for the host's "sky
// place" box. The browser never talks to the geocoding service directly.
func (c *Controllers) PlacesAPI(w http.ResponseWriter, r *http.Request) {
	if c.deps.Places == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) < 2 || utf8.RuneCountInString(q) > maxPlaceQueryRunes {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	places, err := c.deps.Places.Search(ctx, q, placeResults)
	if err != nil {
		c.deps.Log.Warn("place search failed", "err", err)
		httpx.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "place search is not available right now"})
		return
	}
	out := make([]placeJSON, 0, len(places))
	for _, p := range places {
		out = append(out, placeJSON{Name: p.Name, Region: p.Region, Country: p.Country, Lat: p.Lat, Lon: p.Lon, TZ: p.Timezone})
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"places": out})
}
