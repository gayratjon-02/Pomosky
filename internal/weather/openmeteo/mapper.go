package openmeteo

import (
	"errors"
	"math"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/weather"
)

// response is the part of the Open-Meteo reply that is used. "current" values
// of precipitation, rain and snowfall are sums over the preceding interval
// (900 s for 15-minute data), so they are scaled to per-hour rates here.
type response struct {
	Current *struct {
		Time        int64   `json:"time"`
		Interval    int     `json:"interval"`
		Temperature float64 `json:"temperature_2m"`
		CloudCover  float64 `json:"cloud_cover"`
		Precip      float64 `json:"precipitation"`
		Rain        float64 `json:"rain"`
		Snowfall    float64 `json:"snowfall"`
		WeatherCode int     `json:"weather_code"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WindDir     float64 `json:"wind_direction_10m"`
		IsDay       int     `json:"is_day"`
	} `json:"current"`
}

func (r response) snapshot() (weather.Snapshot, error) {
	c := r.Current
	if c == nil || c.Time == 0 {
		return weather.Snapshot{}, errors.New("openmeteo: response has no current conditions")
	}
	interval := c.Interval
	if interval <= 0 {
		interval = 3600
	}
	perHour := 3600 / float64(interval)
	// Rain and showers are part of "precipitation"; take the larger of the two
	// so a model that fills only one of them is not read as dry.
	rainMMH := math.Max(c.Precip, c.Rain) * perHour
	return weather.Snapshot{
		ObservedAt:       time.Unix(c.Time, 0).UTC(),
		CloudCoverPct:    int(math.Round(c.CloudCover)),
		PrecipitationMMH: rainMMH,
		SnowfallCMH:      c.Snowfall * perHour,
		WindSpeedKMH:     c.WindSpeed,
		WindDirDeg:       int(math.Round(c.WindDir)) % 360,
		Code:             c.WeatherCode,
		IsDay:            c.IsDay == 1,
		TemperatureC:     c.Temperature,
	}, nil
}
