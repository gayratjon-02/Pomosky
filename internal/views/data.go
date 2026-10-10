package views

import "github.com/gayratjon-02/Pomosky/internal/models"

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
	Data         any          // page-specific data
}
