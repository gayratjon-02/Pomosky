package realtime

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

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

func TestApplyRunsTheCoreAndBroadcastsState(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	rh := newRoomHub(room.New("abcd2345", "hash", now), log)

	c := NewClient("c1", nil, 4)
	rh.Join(c)

	err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 25 * time.Minute, Label: "api"}, now)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case payload := <-c.Outbox():
		env, err := Decode(payload)
		if err != nil || env.Type != TypeState {
			t.Fatalf("unexpected frame: %v, %+v", err, env)
		}
		var data StateData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			t.Fatal(err)
		}
		if data.Timer.Phase != "focus" || data.Timer.DurationMS != (25*time.Minute).Milliseconds() || data.ServerNowMS != now.UnixMilli() {
			t.Fatalf("unexpected state: %+v", data)
		}
	default:
		t.Fatal("no state was broadcast")
	}

	// An invalid event must neither change the state nor broadcast anything.
	if err := rh.Apply(room.TimerResume{}, now); err == nil {
		t.Fatal("resume on a running timer should fail")
	}
	select {
	case <-c.Outbox():
		t.Fatal("a failed event must not broadcast")
	default:
	}
	if got := rh.State().Timer.Phase; got != room.PhaseFocus {
		t.Fatalf("phase = %q after failed event", got)
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
