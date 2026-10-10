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

// DefaultGeocodingURL is Open-Meteo's geocoding API (same terms as the forecast API).
const DefaultGeocodingURL = "https://geocoding-api.open-meteo.com"

// Geocoder searches places by name with Open-Meteo's geocoding API.
type Geocoder struct {
	baseURL string
	http    *http.Client
}

var _ weather.Geocoder = (*Geocoder)(nil)

// NewGeocoder creates a Geocoder. A nil hc gets a client with a 10 second timeout.
func NewGeocoder(baseURL string, hc *http.Client) *Geocoder {
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Geocoder{baseURL: baseURL, http: hc}
}

// Search implements weather.Geocoder.
func (g *Geocoder) Search(ctx context.Context, query string, limit int) ([]weather.Place, error) {
	q := url.Values{
		"name":     {query},
		"count":    {strconv.Itoa(limit)},
		"language": {"en"},
		"format":   {"json"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/v1/search?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Pomosky/1.0 (+https://pomosky.com)")
	resp, err := g.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openmeteo geocoding: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("openmeteo geocoding: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openmeteo geocoding: status %d", resp.StatusCode)
	}
	return parsePlaces(body)
}

func parsePlaces(body []byte) ([]weather.Place, error) {
	var r struct {
		Results []struct {
			Name      string  `json:"name"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Timezone  string  `json:"timezone"`
			Country   string  `json:"country"`
			Admin1    string  `json:"admin1"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("openmeteo geocoding: bad response: %w", err)
	}
	places := make([]weather.Place, 0, len(r.Results))
	for _, x := range r.Results {
		if x.Name == "" || x.Timezone == "" { // a place without a time zone cannot be used for the sky
			continue
		}
		places = append(places, weather.Place{Name: x.Name, Region: x.Admin1, Country: x.Country, Lat: x.Latitude, Lon: x.Longitude, Timezone: x.Timezone})
	}
	return places, nil
}
