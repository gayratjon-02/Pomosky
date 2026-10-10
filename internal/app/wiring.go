package app

import (
	"log/slog"
	"strings"

	"github.com/gayratjon-02/Pomosky/internal/auth"
	"github.com/gayratjon-02/Pomosky/internal/auth/google"
	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/mail"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// newMailer picks how sign-in emails are delivered: SMTP when configured, the
// log in development, and nothing otherwise (which switches magic links off).
func newMailer(cfg config.Config, log *slog.Logger) ports.Mailer {
	switch {
	case cfg.Mail.SMTPHost != "" && cfg.Mail.From != "":
		return &mail.SMTPMailer{
			Host: cfg.Mail.SMTPHost, Port: cfg.Mail.SMTPPort,
			Username: cfg.Mail.Username, Password: cfg.Mail.Password, From: cfg.Mail.From,
		}
	case !cfg.IsProduction():
		return mail.LogMailer{Log: log}
	default:
		return nil
	}
}

// oauthProviders returns the social login providers that have credentials.
func oauthProviders(cfg config.Config) []auth.Provider {
	var providers []auth.Provider
	if cfg.OAuth.GoogleClientID != "" && cfg.OAuth.GoogleClientSecret != "" {
		redirect := strings.TrimRight(cfg.BaseURL, "/") + "/auth/google/callback"
		providers = append(providers, google.New(cfg.OAuth.GoogleClientID, cfg.OAuth.GoogleClientSecret, redirect))
	}
	return providers
}
