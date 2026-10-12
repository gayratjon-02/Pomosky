package ports

import (
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// PhaseNotifier is told when a focus or break phase of a room ran to its end.
// PhaseEnded must not block: it is called while the room is locked.
type PhaseNotifier interface {
	PhaseEnded(id room.ID, phase room.Phase, label string)
}

// PushSubscriptions lets browsers ask for a Web Push message when a phase of a
// room ends, also while the page is closed or the screen is locked.
type PushSubscriptions interface {
	// PublicKey is the VAPID application server key the browser subscribes with.
	PublicKey() string
	// Subscribe adds (or refreshes) a subscription of a room. It returns
	// apperr.ErrInvalid for an address that is not a known push service.
	Subscribe(id room.ID, sub models.PushSubscription) error
	// Unsubscribe forgets an endpoint in every room.
	Unsubscribe(endpoint string)
}
