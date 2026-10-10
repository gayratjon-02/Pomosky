package room

import "testing"

func TestViewerCount(t *testing.T) {
	steps := []struct {
		name string
		ev   Event
		want int
	}{
		{"first join", ViewerJoined{}, 1},
		{"second join", ViewerJoined{}, 2},
		{"leave", ViewerLeft{}, 1},
		{"leave again", ViewerLeft{}, 0},
		{"duplicate leave stays at zero", ViewerLeft{}, 0},
		{"join after zero", ViewerJoined{}, 1},
	}
	s := newTestState()
	for _, st := range steps {
		s = mustReduce(t, s, t0, st.ev)
		if s.Viewers != st.want {
			t.Fatalf("%s: Viewers = %d, want %d", st.name, s.Viewers, st.want)
		}
	}
}

func TestReactionsAreCounted(t *testing.T) {
	s := newTestState()
	for _, k := range []ReactionKind{ReactionClap, ReactionClap, ReactionFire, ReactionCoffee, ReactionHeart, ReactionClap} {
		s = mustReduce(t, s, t0, ReactionSent{Kind: k})
	}
	want := map[ReactionKind]int{ReactionClap: 3, ReactionFire: 1, ReactionCoffee: 1, ReactionHeart: 1}
	for k, n := range want {
		if s.Reactions[k] != n {
			t.Errorf("Reactions[%s] = %d, want %d", k, s.Reactions[k], n)
		}
	}
}

func TestReactionRejectsUnknownKind(t *testing.T) {
	for _, kind := range []ReactionKind{"", "angry", "CLAP", "clap "} {
		wantErr(t, newTestState(), t0, ReactionSent{Kind: kind}, ErrInvalidEvent)
	}
}

func TestReactionDoesNotMutateInputMap(t *testing.T) {
	before := mustReduce(t, newTestState(), t0, ReactionSent{Kind: ReactionClap})

	after := mustReduce(t, before, t0, ReactionSent{Kind: ReactionClap})

	if before.Reactions[ReactionClap] != 1 {
		t.Fatalf("input map was modified: %v", before.Reactions)
	}
	if after.Reactions[ReactionClap] != 2 {
		t.Fatalf("result map = %v, want clap=2", after.Reactions)
	}
}
