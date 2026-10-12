package views

import (
	"html/template"

	"github.com/gayratjon-02/Pomosky/internal/i18n"
	"github.com/gayratjon-02/Pomosky/internal/models"
)

// RoomPage is the Data of the room, host and record pages.
type RoomPage struct {
	ID       string
	ShareURL string
}

// MobileLab is the Data of the development mobile preview page.
type MobileLab struct {
	Path          string
	Width, Height int
}

// LoginPage is the Data of the login page.
type LoginPage struct {
	Google       bool   // show "Continue with Google"
	MagicLink    bool   // show the email form
	ConfirmToken string // set on the magic-link confirmation step
}

// AccountPage is the Data of the account page.
type AccountPage struct {
	StatsURL string // where the page fetches the monthly statistics from
	Preview  bool   // visitor is not logged in: example data behind a login prompt
}

// PageData is what every page template receives.
type PageData struct {
	Lang         string // language of the page: en, uz, ru or ko
	Title        string
	Description  string
	CanonicalURL string
	OGImage      string
	Flash        string
	CSRFToken    string
	Mood         string       // the visitor's chosen mood (a valid id from moods.go), "" for the live sky
	User         *models.User // nil for anonymous visitors
	Scripts      []string     // page-specific frontend bundles
	DevReload    bool         // development only: include the browser auto-reload script
	Data         any          // page-specific data
}

// LanguageOption is one entry of the language switcher.
type LanguageOption struct {
	Code, Name string
	Current    bool
}

// Languages lists the languages the visitor can switch to, marking the current one.
func (p PageData) Languages() []LanguageOption {
	out := make([]LanguageOption, 0, len(i18n.Supported))
	for _, l := range i18n.Supported {
		out = append(out, LanguageOption{Code: string(l), Name: i18n.Names[l], Current: string(l) == p.Lang})
	}
	return out
}

// T translates an English text into the language of the page (English if there is no translation).
func (p PageData) T(key string) string { return i18n.T(i18n.Parse(p.Lang), key) }

// H is T for texts with simple markup (links, emphasis). Only our own locale files can
// reach it, never visitor input, so the result is trusted HTML.
func (p PageData) H(key string) template.HTML { return template.HTML(p.T(key)) } //nolint:gosec
