// Package openmeteo is the Open-Meteo adapter for weather.Provider.
//
// Endpoint, parameter names, units and the WMO code table come from
// https://open-meteo.com/en/docs (checked 2026-10-10). The free API is for
// non-commercial use only and its data is CC-BY 4.0: pages that show the
// weather must credit Open-Meteo.com.
package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/weather"
)

const currentVars = "temperature_2m,cloud_cover,precipitation,rain,snowfall,weather_code,wind_speed_10m,wind_direction_10m,is_day"

// Client fetches current conditions from Open-Meteo.
type Client struct {
	baseURL string
	http    *http.Client
}

var _ weather.Provider = (*Client)(nil)

// New creates a Client. A nil hc gets a client with a 10 second timeout.
func New(baseURL string, hc *http.Client) *Client {
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{baseURL: baseURL, http: hc}
}

// Current implements weather.Provider.
func (c *Client) Current(ctx context.Context, lat, lon float64) (weather.Snapshot, error) {
	q := url.Values{
		"latitude":   {strconv.FormatFloat(lat, 'f', 2, 64)},
		"longitude":  {strconv.FormatFloat(lon, 'f', 2, 64)},
		"current":    {currentVars},
		"timeformat": {"unixtime"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/forecast?"+q.Encode(), nil)
	if err != nil {
		return weather.Snapshot{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Pomosky/1.0 (+https://pomosky.com)")
	resp, err := c.http.Do(req)
	if err != nil {
		return weather.Snapshot{}, fmt.Errorf("openmeteo: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return weather.Snapshot{}, fmt.Errorf("openmeteo: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return weather.Snapshot{}, fmt.Errorf("openmeteo: status %d", resp.StatusCode)
	}
	return parseCurrent(body)
}

func parseCurrent(body []byte) (weather.Snapshot, error) {
	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return weather.Snapshot{}, fmt.Errorf("openmeteo: bad response: %w", err)
	}
	return r.snapshot()
}
