package ports

import (
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// LiveRooms is the registry of rooms that exist on this server instance right
// now. Controllers use it to create rooms and to decide whether a room page
// can be shown; realtime.Hub implements it.
type LiveRooms interface {
	// Register opens a room. owner is the logged-in host whose statistics the
	// room's focus time counts for, or "" for an anonymous host.
	Register(state room.State, owner models.UserID)
	Has(id room.ID) bool
}

// FocusRecorder receives every finished focus phase of a room with an owner.
// RecordFocus must not block: it is called while a room is locked.
type FocusRecorder interface {
	RecordFocus(s models.FocusSession)
}
