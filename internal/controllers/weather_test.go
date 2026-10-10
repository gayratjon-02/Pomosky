package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

type fakeWeather struct {
	snap weather.Snapshot
	err  error
	lat  float64
	lon  float64
}

func (f *fakeWeather) Current(_ context.Context, lat, lon float64) (weather.Snapshot, error) {
	f.lat, f.lon = lat, lon
	return f.snap, f.err
}

func weatherSetup(t *testing.T, w *fakeWeather, loc room.Location) *Controllers {
	t.Helper()
	live := &fakeLive{}
	st := room.New("abcd2345", "h", time.Now())
	st.Location = loc
	live.Register(st, "")
	c := newTestControllers(t)
	c.deps.Live = live
	c.deps.Weather = w
	c.deps.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	return c
}

func getWeather(c *Controllers, query string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c.WeatherAPI(rec, httptest.NewRequest("GET", "/api/weather"+query, nil))
	return rec
}

func TestWeatherAPIReturnsTheHostPlaceWeather(t *testing.T) {
	fw := &fakeWeather{snap: weather.Snapshot{CloudCoverPct: 40, Code: 63, WindSpeedKMH: 12, FetchedAt: time.UnixMilli(5000)}}
	c := weatherSetup(t, fw, room.Location{City: "Seoul", LatRounded: 37.57, LonRounded: 126.98, TZ: "Asia/Seoul"})

	rec := getWeather(c, "?room=abcd2345")
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") == "" {
		t.Fatalf("status %d, cache %q", rec.Code, rec.Header().Get("Cache-Control"))
	}
	var got weatherJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.CloudCoverPct != 80 || got.PrecipitationMMH <= 0 || got.WindSpeedKMH != 12 || got.FetchedAtMS != 5000 {
		t.Fatalf("body = %+v", got)
	}
	if fw.lat != 37.57 || fw.lon != 126.98 {
		t.Fatalf("looked up %v,%v", fw.lat, fw.lon)
	}
}

func TestWeatherAPIErrors(t *testing.T) {
	loc := room.Location{City: "Seoul", LatRounded: 37.57, LonRounded: 126.98, TZ: "Asia/Seoul"}
	cases := []struct {
		name  string
		c     *Controllers
		query string
		want  int
	}{
		{"missing room", weatherSetup(t, &fakeWeather{}, loc), "", http.StatusBadRequest},
		{"unknown room", weatherSetup(t, &fakeWeather{}, loc), "?room=zzzzzzzz", http.StatusNotFound},
		{"no place set yet", weatherSetup(t, &fakeWeather{}, room.Location{}), "?room=abcd2345", http.StatusNotFound},
		{"upstream down", weatherSetup(t, &fakeWeather{err: errors.New("down")}, loc), "?room=abcd2345", http.StatusBadGateway},
		{"weather off", newTestControllers(t), "?room=abcd2345", http.StatusNotImplemented},
	}
	for _, tc := range cases {
		if rec := getWeather(tc.c, tc.query); rec.Code != tc.want {
			t.Errorf("%s: status %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}

type fakePlaces struct {
	places []weather.Place
	err    error
	query  string
}

func (f *fakePlaces) Search(_ context.Context, q string, _ int) ([]weather.Place, error) {
	f.query = q
	return f.places, f.err
}

func TestPlacesAPI(t *testing.T) {
	c := newTestControllers(t)
	c.deps.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	fp := &fakePlaces{places: []weather.Place{{Name: "Tashkent", Country: "Uzbekistan", Lat: 41.26, Lon: 69.21, Timezone: "Asia/Tashkent"}}}
	c.deps.Places = fp

	get := func(q string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c.PlacesAPI(rec, httptest.NewRequest("GET", "/api/places?q="+q, nil))
		return rec
	}
	rec := get("+Tashkent+")
	var body struct {
		Places []placeJSON `json:"places"`
	}
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &body) != nil || len(body.Places) != 1 || body.Places[0].TZ != "Asia/Tashkent" || fp.query != "Tashkent" {
		t.Fatalf("%d %s (query %q)", rec.Code, rec.Body.String(), fp.query)
	}
	if rec := get("a"); rec.Code != http.StatusBadRequest {
		t.Fatalf("too short: %d", rec.Code)
	}
	fp.err = errors.New("down")
	if rec := get("Seoul"); rec.Code != http.StatusBadGateway {
		t.Fatalf("upstream error: %d", rec.Code)
	}
	c.deps.Places = nil
	if rec := get("Seoul"); rec.Code != http.StatusNotImplemented {
		t.Fatalf("disabled: %d", rec.Code)
	}
}
