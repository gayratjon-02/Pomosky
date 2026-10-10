package ports

import "github.com/gayratjon-02/Pomosky/internal/room"

// Broadcaster sends an encoded message to everyone connected to a room.
type Broadcaster interface {
	Publish(id room.ID, payload []byte)
}
