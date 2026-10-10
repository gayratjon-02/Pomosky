// Package openmeteo is the Open-Meteo adapter for weather.Provider.
package openmeteo

import (
	"context"
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

// Client fetches current conditions from Open-Meteo.
type Client struct {
	baseURL string
	http    *http.Client
}

var _ weather.Provider = (*Client)(nil)

// New creates a Client. A nil hc uses http.DefaultClient.
func New(baseURL string, hc *http.Client) *Client {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &Client{baseURL: baseURL, http: hc}
}

// Current implements weather.Provider.
//
// TODO(stage 6): verify parameter names, units, limits and terms in the
// official Open-Meteo documentation, then implement the request and the
// mapping in mapper.go.
func (c *Client) Current(ctx context.Context, lat, lon float64) (weather.Snapshot, error) {
	return weather.Snapshot{}, apperr.ErrNotImplemented
}
