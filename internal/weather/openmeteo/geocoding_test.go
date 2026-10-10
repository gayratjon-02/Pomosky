package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const geoFixture = `{"results":[
 {"id":1,"name":"Tashkent","latitude":41.26465,"longitude":69.21627,"timezone":"Asia/Tashkent","country":"Uzbekistan","admin1":"Tashkent"},
 {"id":2,"name":"NoZone","latitude":1,"longitude":2,"country":"X"}
],"generationtime_ms":1.0}`

func TestParsePlacesSkipsResultsWithoutTimezone(t *testing.T) {
	places, err := parsePlaces([]byte(geoFixture))
	if err != nil || len(places) != 1 {
		t.Fatalf("places = %+v, err = %v", places, err)
	}
	p := places[0]
	if p.Name != "Tashkent" || p.Country != "Uzbekistan" || p.Timezone != "Asia/Tashkent" || p.Lat != 41.26465 {
		t.Fatalf("place = %+v", p)
	}
}

func TestParsePlacesNoResultsKey(t *testing.T) {
	places, err := parsePlaces([]byte(`{"generationtime_ms":0.5}`)) // Open-Meteo omits "results" when nothing matches
	if err != nil || len(places) != 0 {
		t.Fatalf("places = %+v, err = %v", places, err)
	}
}

func TestSearchSendsTheDocumentedParameters(t *testing.T) {
	var got http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(r.Context())
		_, _ = w.Write([]byte(geoFixture))
	}))
	defer srv.Close()
	if _, err := NewGeocoder(srv.URL, nil).Search(context.Background(), "Tashkent", 5); err != nil {
		t.Fatal(err)
	}
	q := got.URL.Query()
	if got.URL.Path != "/v1/search" || q.Get("name") != "Tashkent" || q.Get("count") != "5" || q.Get("language") != "en" {
		t.Fatalf("request = %s", got.URL)
	}
}
