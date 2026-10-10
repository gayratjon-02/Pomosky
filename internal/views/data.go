package views

import "github.com/gayratjon-02/Pomosky/internal/models"

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
	Title        string
	Description  string
	CanonicalURL string
	OGImage      string
	Flash        string
	CSRFToken    string
	User         *models.User // nil for anonymous visitors
	Scripts      []string     // page-specific frontend bundles
	DevReload    bool         // development only: include the browser auto-reload script
	Data         any          // page-specific data
}
