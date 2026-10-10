package models

import (
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Room is the persisted record of a room (the live state is room.State).
type Room struct {
	ID            room.ID
	OwnerUserID   *UserID
	HostTokenHash string
	Name          string
	Location      room.Location
	CreatedAt     time.Time
}
