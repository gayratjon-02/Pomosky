package room

import "time"

// Phase is the phase of the room timer.
type Phase string

const (
	PhaseIdle   Phase = "idle"
	PhaseFocus  Phase = "focus"
	PhaseBreak  Phase = "break"
	PhasePaused Phase = "paused"
)

// ReactionKind is one of the fixed set of reactions viewers can send.
type ReactionKind string

const (
	ReactionClap   ReactionKind = "clap"
	ReactionCoffee ReactionKind = "coffee"
	ReactionFire   ReactionKind = "fire"
	ReactionHeart  ReactionKind = "heart"
)

// Location is the rounded host location used for the sky.
// Coordinates are rounded (~1 km); exact coordinates are never stored.
type Location struct {
	City       string
	LatRounded float64
	LonRounded float64
	TZ         string // IANA time zone, e.g. "Asia/Seoul"
}

// State is the complete in-memory state of one room.
type State struct {
	ID            ID
	HostTokenHash string
	Timer         Timer
	Viewers       int
	Reactions     map[ReactionKind]int
	Location      Location
	UpdatedAt     time.Time
}

// New creates the initial state of a freshly created room.
func New(id ID, hostTokenHash string, now time.Time) State {
	return State{
		ID:            id,
		HostTokenHash: hostTokenHash,
		Timer:         Timer{Phase: PhaseIdle},
		Reactions:     map[ReactionKind]int{},
		UpdatedAt:     now,
	}
}
