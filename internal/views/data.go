package views

import "github.com/gayratjon-02/Pomosky/internal/models"

// RoomPage is the Data of the room, host and record pages.
type RoomPage struct {
	ID       string
	ShareURL string
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
