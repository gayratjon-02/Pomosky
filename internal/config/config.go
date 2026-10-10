// Package config loads runtime configuration from environment variables.
//
// It is the only package that reads the environment. Everything else
// receives a Config value from the composition root (internal/app).
package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config is the complete runtime configuration of the server.
type Config struct {
	Env           string // "development" or "production"
	Addr          string // listen address, e.g. ":8080"
	BaseURL       string // public base URL, used for links and OAuth callbacks
	StaticDir     string // directory with built frontend assets
	DatabaseURL   string
	SessionSecret string
	HostTokenSalt string

	Weather WeatherConfig
	OAuth   OAuthConfig
	Mail    MailConfig
}

// WeatherConfig configures the weather adapter and its cache.
type WeatherConfig struct {
	BaseURL  string
	CacheTTL time.Duration
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
		Env:           getenv("APP_ENV", "development"),
		Addr:          getenv("ADDR", ":"+getenv("PORT", "8080")),
		BaseURL:       getenv("BASE_URL", "http://localhost:8080"),
		StaticDir:     getenv("STATIC_DIR", "web/dist"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		HostTokenSalt: os.Getenv("HOST_TOKEN_SALT"),
		Weather: WeatherConfig{
			// Default is the public Open-Meteo host; verify its terms and limits.
			BaseURL:  getenv("OPEN_METEO_BASE", "https://api.open-meteo.com"),
			CacheTTL: getenvDuration("WEATHER_CACHE_TTL", 10*time.Minute),
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
	case c.DatabaseURL == "":
		return errors.New("config: DATABASE_URL is required in production")
	case c.SessionSecret == "":
		return errors.New("config: SESSION_SECRET is required in production")
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
