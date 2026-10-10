package ports

import "github.com/gayratjon-02/Pomosky/internal/room"

// LiveRooms is the registry of rooms that exist on this server instance right
// now. Controllers use it to create rooms and to decide whether a room page
// can be shown; realtime.Hub implements it.
type LiveRooms interface {
	Register(state room.State)
	Has(id room.ID) bool
}
