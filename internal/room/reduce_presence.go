package room

func viewerJoined(s State) (State, error) {
	s.Viewers++
	return s, nil
}

// viewerLeft never lets the count go below zero, so a duplicate "left" event
// from a flaky connection cannot corrupt the number.
func viewerLeft(s State) (State, error) {
	if s.Viewers > 0 {
		s.Viewers--
	}
	return s, nil
}

func validReaction(k ReactionKind) bool {
	switch k {
	case ReactionClap, ReactionCoffee, ReactionFire, ReactionHeart:
		return true
	default:
		return false
	}
}

// reactionSent counts one reaction. The map is copied first: State values are
// shared by callers, and Reduce must never modify the map of its input.
func reactionSent(s State, e ReactionSent) (State, error) {
	if !validReaction(e.Kind) {
		return s, ErrInvalidEvent
	}
	next := make(map[ReactionKind]int, len(s.Reactions)+1)
	for k, v := range s.Reactions {
		next[k] = v
	}
	next[e.Kind]++
	s.Reactions = next
	return s, nil
}
