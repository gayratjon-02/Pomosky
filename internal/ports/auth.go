package ports

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

// Accounts is what controllers need from the optional login system.
// auth.Service implements it. Every method is safe to call when login is
// disabled: the Enabled methods report what is available.
type Accounts interface {
	MagicLinkEnabled() bool
	// SendMagicLink emails a one-time sign-in link. It returns apperr.ErrInvalid
	// for a malformed address.
	SendMagicLink(ctx context.Context, email string) error
	CompleteMagicLink(ctx context.Context, token string) (models.Login, error)

	ProviderEnabled(name string) bool
	// ProviderStart returns the URL to send the browser to and the state value
	// that must come back on the callback.
	ProviderStart(name string) (authURL, state string, err error)
	CompleteProvider(ctx context.Context, name, code string) (models.Login, error)

	SignOut(ctx context.Context, token string) error
	// DeleteAccount permanently deletes the user, their rooms and focus history.
	DeleteAccount(ctx context.Context, id models.UserID) error
}

// StatsReader serves the monthly statistics of a user.
type StatsReader interface {
	// Month returns the statistics of month ("YYYY-MM") with days grouped in
	// the time zone tz (IANA name; invalid names fall back to UTC).
	Month(ctx context.Context, user models.UserID, month, tz string) (models.MonthStats, error)
}
