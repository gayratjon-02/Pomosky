package realtime

import (
	"io"
	"log/slog"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	b, err := Encode(TypePresence, PresenceData{Viewers: 7})
	if err != nil {
		t.Fatal(err)
	}
	env, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != TypePresence {
		t.Fatalf("type = %q", env.Type)
	}
}

func TestRoomHubBroadcast(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rh := newRoomHub(room.State{ID: "abcd2345"}, log)

	c := NewClient("c1", nil, 4)
	if n := rh.Join(c); n != 1 {
		t.Fatalf("Join() = %d, want 1", n)
	}
	rh.Broadcast([]byte("hello"))
	select {
	case got := <-c.Outbox():
		if string(got) != "hello" {
			t.Fatalf("got %q", got)
		}
	default:
		t.Fatal("client did not receive the broadcast")
	}
	if n := rh.Leave(c); n != 0 {
		t.Fatalf("Leave() = %d, want 0", n)
	}
}

func TestSlowClientDoesNotBlock(t *testing.T) {
	c := NewClient("slow", nil, 1)
	if !c.Send([]byte("1")) {
		t.Fatal("first send should fit")
	}
	if c.Send([]byte("2")) {
		t.Fatal("second send should be dropped, not block")
	}
}
