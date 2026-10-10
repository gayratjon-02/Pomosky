package ports

import "github.com/gayratjon-02/Pomosky/internal/room"

// IDGenerator creates unguessable identifiers.
type IDGenerator interface {
	NewRoomID() room.ID
	NewToken() string // host tokens, magic-link tokens
	NewSessionID() string
}
