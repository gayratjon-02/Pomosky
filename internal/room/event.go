package room

import "time"

// Event is something that happens to a room. All inputs (WebSocket messages,
// recovery after restart) are converted to Events before reaching Reduce.
type Event interface {
	EventName() string
}

// TimerStart starts a new focus or break phase.
type TimerStart struct {
	Phase    Phase
	Duration time.Duration
	Label    string
}

// TimerPause pauses the running phase.
type TimerPause struct{}

// TimerResume resumes a paused phase.
type TimerResume struct{}

// TimerReset returns the timer to idle.
type TimerReset struct{}

// ViewerJoined increments the viewer count.
type ViewerJoined struct{}

// ViewerLeft decrements the viewer count.
type ViewerLeft struct{}

// ReactionSent counts a viewer reaction.
type ReactionSent struct{ Kind ReactionKind }

// LocationSet sets the (rounded) host location.
type LocationSet struct{ Location Location }

// HostRecovered restores the timer after a server restart, sent by the
// reconnecting host.
type HostRecovered struct{ Timer Timer }

func (TimerStart) EventName() string    { return "timer.start" }
func (TimerPause) EventName() string    { return "timer.pause" }
func (TimerResume) EventName() string   { return "timer.resume" }
func (TimerReset) EventName() string    { return "timer.reset" }
func (ViewerJoined) EventName() string  { return "viewer.joined" }
func (ViewerLeft) EventName() string    { return "viewer.left" }
func (ReactionSent) EventName() string  { return "reaction.sent" }
func (LocationSet) EventName() string   { return "location.set" }
func (HostRecovered) EventName() string { return "host.recovered" }
