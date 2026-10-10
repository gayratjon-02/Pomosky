package room

import "time"

// Reduce applies ev to s and returns the next state.
//
// It is pure: the same inputs always give the same result and s is never
// modified (the returned State may share nothing mutable with s). now is the
// server time of the event. On error the original state is returned unchanged.
//
// TODO(stage 2): viewer, reaction, location and recovery events.
func Reduce(s State, ev Event, now time.Time) (State, error) {
	var (
		next State
		err  error
	)
	switch e := ev.(type) {
	case TimerStart:
		next, err = startTimer(s, e, now)
	case TimerPause:
		next, err = pauseTimer(s, now)
	case TimerResume:
		next, err = resumeTimer(s, now)
	case TimerReset:
		next, err = resetTimer(s)
	case ViewerJoined, ViewerLeft, ReactionSent, LocationSet, HostRecovered:
		return s, ErrNotImplemented
	default:
		return s, ErrUnknownEvent
	}
	if err != nil {
		return s, err
	}
	next.UpdatedAt = now
	return next, nil
}
