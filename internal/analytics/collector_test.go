package analytics

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time { return c.t }

type memRepo struct {
	mu   sync.Mutex
	got  []models.Visit
	over models.TrafficOverview
}

func (m *memRepo) Create(_ context.Context, v models.Visit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.got = append(m.got, v)
	return nil
}

func (m *memRepo) Traffic(context.Context, time.Time, time.Time) (models.TrafficOverview, error) {
	return m.over, nil
}

func newTestCollector(repo *memRepo, clock *fakeClock) *Collector {
	return NewCollector(repo, slog.New(slog.NewTextHandler(io.Discard, nil)), clock, "secret", false)
}

func TestCollectorRecordsAnAnonymousVisit(t *testing.T) {
	repo := &memRepo{}
	clock := &fakeClock{t: time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)}
	c := newTestCollector(repo, clock)

	r := httptest.NewRequest("GET", "http://pomosky.test/r/abcd2345?utm_source=instagram&utm_medium=reel", nil)
	r.RemoteAddr = "203.0.113.7:5555"
	r.Header.Set("User-Agent", uaIPhoneSafari)
	r.Header.Set("CF-IPCountry", "uz")
	c.Track(r, models.VisitPageview)
	c.Close()

	if len(repo.got) != 1 {
		t.Fatalf("recorded %d visits", len(repo.got))
	}
	v := repo.got[0]
	if v.Path != "/r/:id" || v.Source != "instagram" || v.Medium != "reel" || v.Device != "mobile" || v.Country != "UZ" || v.Kind != models.VisitPageview {
		t.Fatalf("visit %+v", v)
	}
	if v.Visitor == "" || len(v.Visitor) != 16 {
		t.Fatalf("visitor %q", v.Visitor)
	}
}

func TestVisitorChangesEveryDayButNotWithinOne(t *testing.T) {
	c := newTestCollector(&memRepo{}, &fakeClock{})
	defer c.Close()
	d1 := time.Date(2026, 10, 10, 1, 0, 0, 0, time.UTC)
	same := c.visitor(d1, "203.0.113.7", "ua")
	if c.visitor(d1.Add(10*time.Hour), "203.0.113.7", "ua") != same {
		t.Fatal("the visitor changed within a day")
	}
	if c.visitor(d1.AddDate(0, 0, 1), "203.0.113.7", "ua") == same {
		t.Fatal("the visitor did not change on the next day")
	}
	if c.visitor(d1, "203.0.113.8", "ua") == same {
		t.Fatal("different addresses share a visitor")
	}
}

func TestCollectorSkipsBotsAndPrivacySignals(t *testing.T) {
	repo := &memRepo{}
	c := newTestCollector(repo, &fakeClock{t: time.Now()})
	for _, set := range map[string]func(h map[string]string){
		"bot":  func(h map[string]string) { h["User-Agent"] = "Googlebot/2.1" },
		"dnt":  func(h map[string]string) { h["DNT"] = "1" },
		"gpc":  func(h map[string]string) { h["Sec-GPC"] = "1" },
		"none": func(h map[string]string) { delete(h, "User-Agent") },
	} {
		h := map[string]string{"User-Agent": uaMacChrome}
		set(h)
		r := httptest.NewRequest("GET", "http://pomosky.test/", nil)
		for k, v := range h {
			r.Header.Set(k, v)
		}
		c.Track(r, models.VisitPageview)
	}
	c.Close()
	if len(repo.got) != 0 {
		t.Fatalf("recorded %+v", repo.got)
	}
}

func TestReaderFillsEveryDay(t *testing.T) {
	repo := &memRepo{over: models.TrafficOverview{Days: []models.TrafficDay{{Date: "2026-10-10", Pageviews: 5, Visitors: 3}}}}
	r := NewReader(repo, &fakeClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)})
	got, err := r.Traffic(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Days) != TrafficWindowDays || got.Days[0].Date != "2026-09-11" || got.Days[len(got.Days)-1].Date != "2026-10-10" {
		t.Fatalf("days %d first %s", len(got.Days), got.Days[0].Date)
	}
	if got.Today.Pageviews != 5 || got.Today.Visitors != 3 {
		t.Fatalf("today %+v", got.Today)
	}
}

type fakeGeo map[string]string

func (g fakeGeo) Country(ip string) string { return g[ip] }

func TestCollectorLooksUpTheCountryWhenNoProxyHeaderSaysSo(t *testing.T) {
	repo := &memRepo{}
	c := newTestCollector(repo, &fakeClock{t: time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)})
	c.SetCountryLookup(fakeGeo{"203.0.113.7": "KR"})
	for _, tc := range []struct{ ip, header string }{
		{"203.0.113.7:1", ""},   // looked up
		{"203.0.113.7:1", "uz"}, // the proxy header wins
		{"198.51.100.1:1", ""},  // unknown address
	} {
		r := httptest.NewRequest("GET", "http://pomosky.test/", nil)
		r.RemoteAddr = tc.ip
		r.Header.Set("User-Agent", uaIPhoneSafari)
		if tc.header != "" {
			r.Header.Set("CF-IPCountry", tc.header)
		}
		c.Track(r, models.VisitPageview)
	}
	c.Close()
	got := []string{}
	for _, v := range repo.got {
		got = append(got, v.Country)
	}
	if len(got) != 3 || got[0] != "KR" || got[1] != "UZ" || got[2] != "" {
		t.Fatalf("countries %q", got)
	}
}
