package analytics

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// Collector implements ports.VisitTracker. Track only queues the event; a
// background worker writes it, so a slow database never slows a page.
//
// Privacy: no cookie, no IP address and no room id is stored (the country is
// looked up from the address and only its two-letter code is kept). The visitor is an
// HMAC of the UTC day, the client address and the browser, so it cannot be
// linked from one day to the next. Visitors who send Do Not Track or Global
// Privacy Control, and bots, are not counted at all.
type Collector struct {
	repo       ports.VisitRepository
	log        *slog.Logger
	clock      ports.Clock
	key        []byte
	trustProxy bool
	geo        CountryLookup // optional: the country of an address when no proxy header says

	ch   chan models.Visit
	done chan struct{}
	once sync.Once
}

var _ ports.VisitTracker = (*Collector)(nil)

// CountryLookup finds the two-letter country of an IP address ("" when unknown).
type CountryLookup interface {
	Country(ip string) string
}

// SetCountryLookup makes the collector fill in the country itself when the
// request carries no proxy header such as CF-IPCountry. Call it before traffic starts.
func (c *Collector) SetCountryLookup(l CountryLookup) { c.geo = l }

// NewCollector starts the worker. key is a server secret (it keeps the daily
// hashes from being guessed). Call Close to flush and stop.
func NewCollector(repo ports.VisitRepository, log *slog.Logger, clock ports.Clock, key string, trustProxy bool) *Collector {
	c := &Collector{
		repo: repo, log: log, clock: clock, key: []byte("pomosky-visits:" + key), trustProxy: trustProxy,
		ch: make(chan models.Visit, 512), done: make(chan struct{}),
	}
	go c.run()
	return c
}

// Track implements ports.VisitTracker.
func (c *Collector) Track(r *http.Request, kind string) {
	if r.Header.Get("DNT") == "1" || r.Header.Get("Sec-GPC") == "1" {
		return
	}
	client := ParseUserAgent(r.UserAgent())
	if client.Bot {
		return
	}
	now := c.clock.Now().UTC()
	ip := httpx.ClientIP(r, c.trustProxy)
	cc := country(r)
	if cc == "" && c.geo != nil {
		cc = c.geo.Country(ip)
	}
	source, medium, campaign := Source(r.URL.Query(), r.Referer(), hostOnly(r.Host), client)
	v := models.Visit{
		At: now, Kind: kind, Path: NormalizePath(r.URL.Path),
		Visitor: c.visitor(now, ip, r.UserAgent()),
		Source:  source, Medium: medium, Campaign: campaign,
		Device: client.Device, OS: client.OS, Browser: client.Browser,
		Country: cc,
	}
	select {
	case c.ch <- v:
	default:
		c.log.Warn("visit dropped, queue is full")
	}
}

// visitor is the daily anonymous id.
func (c *Collector) visitor(now time.Time, ip, ua string) string {
	m := hmac.New(sha256.New, c.key)
	m.Write([]byte(now.Format("2006-01-02")))
	m.Write([]byte{0})
	m.Write([]byte(ip))
	m.Write([]byte{0})
	m.Write([]byte(ua))
	return hex.EncodeToString(m.Sum(nil))[:16]
}

// country reads the two-letter country a proxy such as Cloudflare adds. Without one it is empty.
func country(r *http.Request) string {
	cc := strings.ToUpper(strings.TrimSpace(r.Header.Get("CF-IPCountry")))
	if len(cc) != 2 || cc == "XX" || cc == "T1" {
		return ""
	}
	for _, ch := range cc {
		if ch < 'A' || ch > 'Z' {
			return ""
		}
	}
	return cc
}

func hostOnly(h string) string {
	if host, _, err := net.SplitHostPort(h); err == nil {
		return host
	}
	return h
}

// Close writes what is still queued and stops the worker.
func (c *Collector) Close() {
	c.once.Do(func() { close(c.ch) })
	<-c.done
}

func (c *Collector) run() {
	defer close(c.done)
	for v := range c.ch {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := c.repo.Create(ctx, v); err != nil {
			c.log.Error("store visit", "kind", v.Kind, "err", err)
		}
		cancel()
	}
}
