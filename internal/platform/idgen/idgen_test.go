package idgen

import "testing"

func TestGeneratedIDsAreValidAndUnique(t *testing.T) {
	g := Generator{}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := g.NewRoomID()
		if !id.Valid() {
			t.Fatalf("generated invalid room ID %q", id)
		}
		if seen[id.String()] {
			t.Fatalf("duplicate room ID %q", id)
		}
		seen[id.String()] = true
	}
	if len(g.NewToken()) < 40 {
		t.Fatal("token too short")
	}
}
