package analytics

import (
	"net/url"
	"testing"
)

const (
	uaIPhoneSafari = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1"
	uaInstagram    = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Instagram 350.0.0.0.0 (iPhone16,2)"
	uaTikTok       = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Mobile Safari/537.36 musical_ly_2024 BytedanceWebview/d8a21c6"
	uaMacChrome    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"
)

func TestParseUserAgent(t *testing.T) {
	for _, tc := range []struct {
		name, ua            string
		device, os, browser string
		bot                 bool
	}{
		{"iphone safari", uaIPhoneSafari, "mobile", "iOS", "Safari", false},
		{"instagram in-app", uaInstagram, "mobile", "iOS", "Instagram", false},
		{"tiktok in-app", uaTikTok, "mobile", "Android", "TikTok", false},
		{"mac chrome", uaMacChrome, "desktop", "macOS", "Chrome", false},
		{"googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", "", "", "", true},
		{"link preview", "facebookexternalhit/1.1", "", "", "", true},
		{"empty", "", "", "", "", true},
	} {
		c := ParseUserAgent(tc.ua)
		if c.Bot != tc.bot || (!tc.bot && (c.Device != tc.device || c.OS != tc.os || c.Browser != tc.browser)) {
			t.Errorf("%s: got %+v", tc.name, c)
		}
	}
}

func TestSource(t *testing.T) {
	web := ParseUserAgent(uaMacChrome)
	app := ParseUserAgent(uaInstagram)
	for _, tc := range []struct {
		name         string
		query        string
		ref          string
		c            Client
		src, med, cp string
	}{
		{"utm wins", "utm_source=Instagram&utm_medium=reel&utm_campaign=launch", "https://t.co/x", web, "instagram", "reel", "launch"},
		{"referrer", "", "https://l.instagram.com/?u=x", web, "instagram", "referral", ""},
		{"www stripped", "", "https://www.example.org/post", web, "example.org", "referral", ""},
		{"own site is not a source", "", "https://pomosky.test/", web, "direct", "", ""},
		{"in-app without referrer", "", "", app, "instagram", "social", ""},
		{"direct", "", "", web, "direct", "", ""},
		{"x.com is not box.com", "", "https://box.com/", web, "box.com", "referral", ""},
		{"google country domain", "", "https://www.google.co.kr/", web, "google", "referral", ""},
	} {
		q, _ := url.ParseQuery(tc.query)
		s, m, c := Source(q, tc.ref, "pomosky.test", tc.c)
		if s != tc.src || m != tc.med || c != tc.cp {
			t.Errorf("%s: got %q %q %q, want %q %q %q", tc.name, s, m, c, tc.src, tc.med, tc.cp)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	for in, want := range map[string]string{
		"/": "/", "/r/abcd2345": "/r/:id", "/r/abcd2345/host": "/r/:id/host", "/r/new": "/r/new",
		"/privacy/": "/privacy", "/account": "/account", "/auth/callback": "/auth/callback",
	} {
		if got := NormalizePath(in); got != want {
			t.Errorf("NormalizePath(%q) = %q, want %q", in, got, want)
		}
	}
}
