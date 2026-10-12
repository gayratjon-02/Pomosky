// Package config loads runtime configuration from environment variables.
//
// It is the only package that reads the environment. Everything else
// receives a Config value from the composition root (internal/app).
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the complete runtime configuration of the server.
type Config struct {
	Env           string // "development" or "production"
	Addr          string // listen address, e.g. ":8080"
	BaseURL       string // public base URL, used for links and OAuth callbacks
	StaticDir     string // directory with built frontend assets
	DatabaseURL   string
	HostTokenSalt string
	RoomIdleTTL   time.Duration // how long a room with nobody connected stays alive
	AdminEmails   []string      // logged-in users with these (verified) emails may open /admin
	// AdminHost, when set, is the only host that serves /admin and /api/admin/*
	// (everywhere else they answer 404). CookieDomain, e.g. ".example.com", lets the
	// login session cookie reach that host as well.
	AdminHost    string
	CookieDomain string
	// GeoIPFile is the DB-IP "IP to Country Lite" CSV (plain or .gz). When set, the
	// country of a visit is looked up from the address if no proxy header gives it.
	GeoIPFile string
	// VAPID keys switch on Web Push (alerts when a phase ends while the page is
	// closed). Generate them with `go run ./cmd/vapid`. VAPIDSubject is a mailto: or
	// https: contact for the push services; it defaults to BaseURL.
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

	// TrustProxy makes the server believe X-Forwarded-For. Enable it only when
	// every request comes through a proxy you control (Caddy), otherwise clients
	// could spoof their IP and dodge rate limits.
	TrustProxy bool

	// Development helpers. Dev enables template reloading from DevViewsDir and
	// the browser auto-reload endpoints. It is rejected in production.
	Dev         bool
	DevViewsDir string

	Weather WeatherConfig
	OAuth   OAuthConfig
	Mail    MailConfig
}

// WeatherConfig configures the weather adapter and its cache.
type WeatherConfig struct {
	BaseURL      string
	GeocodingURL string // place search
	CacheTTL     time.Duration
}

// OAuthConfig holds the credentials of the optional social login.
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
}

// MailConfig holds the settings of the optional magic-link mailer.
type MailConfig struct {
	From     string
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

// Load reads the configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		Env:             getenv("APP_ENV", "development"),
		Addr:            getenv("ADDR", ":"+getenv("PORT", "8080")),
		BaseURL:         getenv("BASE_URL", "http://localhost:8080"),
		StaticDir:       getenv("STATIC_DIR", "web/dist"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		HostTokenSalt:   os.Getenv("HOST_TOKEN_SALT"),
		RoomIdleTTL:     getenvDuration("ROOM_IDLE_TTL", 10*time.Minute),
		AdminEmails:     splitEmails(os.Getenv("ADMIN_EMAILS")),
		AdminHost:       strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_HOST"))),
		CookieDomain:    strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_DOMAIN"))),
		GeoIPFile:       strings.TrimSpace(os.Getenv("GEOIP_FILE")),
		VAPIDPublicKey:  strings.TrimSpace(os.Getenv("VAPID_PUBLIC_KEY")),
		VAPIDPrivateKey: strings.TrimSpace(os.Getenv("VAPID_PRIVATE_KEY")),
		VAPIDSubject:    strings.TrimSpace(os.Getenv("VAPID_SUBJECT")),
		TrustProxy:      getenvBool("TRUST_PROXY", false),
		Dev:             getenvBool("DEV_RELOAD", false),
		DevViewsDir:     getenv("DEV_VIEWS_DIR", "internal/views"),
		Weather: WeatherConfig{
			// Default is the public Open-Meteo host; verify its terms and limits.
			BaseURL:      getenv("OPEN_METEO_BASE", "https://api.open-meteo.com"),
			GeocodingURL: getenv("OPEN_METEO_GEOCODING_BASE", "https://geocoding-api.open-meteo.com"),
			CacheTTL:     getenvDuration("WEATHER_CACHE_TTL", 10*time.Minute),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		},
		Mail: MailConfig{
			From:     os.Getenv("MAIL_FROM"),
			SMTPHost: os.Getenv("SMTP_HOST"),
			SMTPPort: getenvInt("SMTP_PORT", 587),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
		},
	}
}

// IsProduction reports whether the server runs in production mode.
func (c Config) IsProduction() bool { return c.Env == "production" }

// Validate rejects configurations that are unsafe to run.
func (c Config) Validate() error {
	if !c.IsProduction() {
		return nil
	}
	switch {
	case c.Dev:
		return errors.New("config: DEV_RELOAD must not be enabled in production")
	case c.DatabaseURL == "":
		return errors.New("config: DATABASE_URL is required in production")
	case c.HostTokenSalt == "":
		return errors.New("config: HOST_TOKEN_SALT is required in production")
	}
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// splitEmails parses a comma or space separated list into lower-case addresses.
func splitEmails(raw string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t' }) {
		if e := strings.ToLower(strings.TrimSpace(part)); e != "" {
			out = append(out, e)
		}
	}
	return out
}
