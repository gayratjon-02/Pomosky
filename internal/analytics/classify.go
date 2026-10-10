package analytics

import (
	"net/url"
	"strings"
)

// Client describes a browser as far as the traffic charts need it.
type Client struct {
	Device  string // mobile, tablet, desktop
	OS      string
	Browser string // includes in-app browsers such as Instagram and TikTok
	Bot     bool
}

var botMarkers = []string{
	"bot", "crawl", "spider", "slurp", "headless", "curl/", "wget", "python-requests", "go-http-client",
	"monitor", "uptime", "preview", "facebookexternalhit", "whatsapp", "telegrambot", "slackbot",
	"discordbot", "lighthouse", "pingdom", "httpclient", "okhttp", "axios", "node-fetch",
}

// ParseUserAgent classifies a User-Agent string. It is deliberately simple: the
// aim is a readable breakdown, not exact browser versions.
func ParseUserAgent(ua string) Client {
	l := strings.ToLower(ua)
	if ua == "" {
		return Client{Bot: true}
	}
	for _, m := range botMarkers {
		if strings.Contains(l, m) {
			return Client{Bot: true}
		}
	}
	c := Client{Device: "desktop"}
	switch {
	case strings.Contains(l, "ipad") || strings.Contains(l, "tablet"):
		c.Device = "tablet"
	case strings.Contains(l, "iphone") || strings.Contains(l, "ipod") || (strings.Contains(l, "android") && strings.Contains(l, "mobile")):
		c.Device = "mobile"
	case strings.Contains(l, "android"):
		c.Device = "tablet"
	}
	switch {
	case strings.Contains(l, "iphone") || strings.Contains(l, "ipad") || strings.Contains(l, "ipod"):
		c.OS = "iOS"
	case strings.Contains(l, "android"):
		c.OS = "Android"
	case strings.Contains(l, "windows"):
		c.OS = "Windows"
	case strings.Contains(l, "mac os x") || strings.Contains(l, "macintosh"):
		c.OS = "macOS"
	case strings.Contains(l, "linux") || strings.Contains(l, "cros"):
		c.OS = "Linux"
	default:
		c.OS = "Other"
	}
	switch {
	case strings.Contains(l, "instagram"):
		c.Browser = "Instagram"
	case strings.Contains(l, "tiktok") || strings.Contains(l, "musical_ly") || strings.Contains(l, "bytedancewebview") || strings.Contains(l, "trill"):
		c.Browser = "TikTok"
	case strings.Contains(l, "fban") || strings.Contains(l, "fbav"):
		c.Browser = "Facebook"
	case strings.Contains(l, "telegram"):
		c.Browser = "Telegram"
	case strings.Contains(l, "edg/") || strings.Contains(l, "edga/") || strings.Contains(l, "edgios/"):
		c.Browser = "Edge"
	case strings.Contains(l, "firefox") || strings.Contains(l, "fxios"):
		c.Browser = "Firefox"
	case strings.Contains(l, "opr/") || strings.Contains(l, "opera"):
		c.Browser = "Opera"
	case strings.Contains(l, "samsungbrowser"):
		c.Browser = "Samsung"
	case strings.Contains(l, "chrome") || strings.Contains(l, "crios"):
		c.Browser = "Chrome"
	case strings.Contains(l, "safari"):
		c.Browser = "Safari"
	default:
		c.Browser = "Other"
	}
	return c
}

// Source says where a visit came from. utm_source wins, then the referrer's
// site, then an in-app browser (a link opened from Instagram or TikTok often
// has no referrer), otherwise "direct". medium is utm_medium, "referral" for a
// referrer, "social" for an in-app browser and "" for direct visits.
func Source(query url.Values, referrer, ownHost string, c Client) (source, medium, campaign string) {
	campaign = clean(query.Get("utm_campaign"), 60)
	if s := clean(strings.ToLower(query.Get("utm_source")), 40); s != "" {
		m := clean(strings.ToLower(query.Get("utm_medium")), 40)
		return s, m, campaign
	}
	if host := referrerHost(referrer); host != "" && host != strings.ToLower(ownHost) {
		return siteName(host), "referral", campaign
	}
	switch c.Browser {
	case "Instagram", "TikTok", "Facebook", "Telegram":
		return strings.ToLower(c.Browser), "social", campaign
	}
	return "direct", "", campaign
}

// siteName maps well-known hosts to one short name, so m.facebook.com and
// l.facebook.com count together.
func siteName(host string) string {
	for _, p := range []struct{ suffix, name string }{
		{"instagram.com", "instagram"}, {"tiktok.com", "tiktok"}, {"facebook.com", "facebook"}, {"fb.com", "facebook"},
		{"t.co", "twitter"}, {"twitter.com", "twitter"}, {"x.com", "twitter"}, {"youtube.com", "youtube"},
		{"youtu.be", "youtube"}, {"t.me", "telegram"}, {"telegram.org", "telegram"}, {"reddit.com", "reddit"},
		{"linkedin.com", "linkedin"}, {"lnkd.in", "linkedin"}, {"github.com", "github"}, {"news.ycombinator.com", "hackernews"},
		{"producthunt.com", "producthunt"}, {"google.", "google"}, {"bing.com", "bing"}, {"duckduckgo.com", "duckduckgo"},
		{"yandex.", "yandex"}, {"naver.com", "naver"},
	} {
		if matchesSite(host, p.suffix) {
			return p.name
		}
	}
	return host
}

// matchesSite reports whether host is the site or one of its subdomains. A suffix
// ending in a dot ("google.") stands for every country domain of that name.
func matchesSite(host, site string) bool {
	if strings.HasSuffix(site, ".") {
		return strings.HasPrefix(host, site) || strings.Contains(host, "."+site)
	}
	return host == site || strings.HasSuffix(host, "."+site)
}

func referrerHost(ref string) string {
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil || u.Host == "" {
		return ""
	}
	h := strings.ToLower(u.Hostname())
	return strings.TrimPrefix(h, "www.")
}

// NormalizePath keeps the page and drops anything that identifies a room.
func NormalizePath(p string) string {
	switch {
	case p == "/r/new":
		return "/r/new"
	case strings.HasPrefix(p, "/r/"):
		if strings.HasSuffix(p, "/host") {
			return "/r/:id/host"
		}
		return "/r/:id"
	case strings.HasPrefix(p, "/auth/"):
		return p
	}
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	if len(p) > 80 {
		p = p[:80]
	}
	return p
}

// clean trims a client-supplied label: printable, single line, limited length.
func clean(s string, max int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
		if b.Len() >= max {
			break
		}
	}
	return b.String()
}
