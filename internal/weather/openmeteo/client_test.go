package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/seoul_current.json")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseRecordedResponse(t *testing.T) {
	s, err := parseCurrent(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if s.CloudCoverPct != 32 || s.Code != 1 || s.WindDirDeg != 307 || s.IsDay || s.TemperatureC != 20.3 || s.WindSpeedKMH != 3.6 {
		t.Fatalf("snapshot = %+v", s)
	}
	if s.ObservedAt.Unix() != 1791626400 {
		t.Fatalf("observed at %v", s.ObservedAt)
	}
}

func TestPrecipitationIsScaledToPerHour(t *testing.T) {
	body := []byte(`{"current":{"time":1791626400,"interval":900,"precipitation":0.5,"rain":0.4,"snowfall":0.25,"weather_code":63}}`)
	s, err := parseCurrent(body)
	if err != nil {
		t.Fatal(err)
	}
	if s.PrecipitationMMH != 2 || s.SnowfallCMH != 1 {
		t.Fatalf("per-hour rates: rain %v snow %v", s.PrecipitationMMH, s.SnowfallCMH)
	}
}

func TestParseRejectsBrokenResponses(t *testing.T) {
	for _, body := range []string{`{}`, `not json`, `{"current":{"time":0}}`} {
		if _, err := parseCurrent([]byte(body)); err == nil {
			t.Errorf("accepted %q", body)
		}
	}
}

func TestCurrentSendsTheDocumentedParameters(t *testing.T) {
	var got http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(r.Context())
		_, _ = w.Write(fixture(t))
	}))
	defer srv.Close()

	snap, err := New(srv.URL, nil).Current(context.Background(), 37.5665, 126.978)
	if err != nil || snap.Code != 1 {
		t.Fatalf("Current: %+v %v", snap, err)
	}
	q := got.URL.Query()
	if got.URL.Path != "/v1/forecast" || q.Get("latitude") != "37.57" || q.Get("longitude") != "126.98" ||
		q.Get("timeformat") != "unixtime" || q.Get("current") != currentVars {
		t.Fatalf("request = %s", got.URL)
	}
}

func TestCurrentReportsUpstreamErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "limit", http.StatusTooManyRequests)
	}))
	defer srv.Close()
	if _, err := New(srv.URL, nil).Current(context.Background(), 1, 1); err == nil {
		t.Fatal("a 429 was accepted")
	}
}
