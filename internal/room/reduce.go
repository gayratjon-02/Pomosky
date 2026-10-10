package room

import "time"

// Reduce applies ev to s and returns the next state.
//
// It is pure: the same inputs always give the same result. now is the server
// time of the event.
//
// TODO(stage 2): implement every case and cover it with table tests.
func Reduce(s State, ev Event, now time.Time) (State, error) {
	switch ev.(type) {
	case TimerStart, TimerPause, TimerResume, TimerReset,
		ViewerJoined, ViewerLeft, ReactionSent, LocationSet, HostRecovered:
		return s, ErrNotImplemented
	default:
		return s, ErrUnknownEvent
	}
}
