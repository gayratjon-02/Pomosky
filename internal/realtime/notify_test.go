package realtime

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

type endNotice struct {
	id    room.ID
	phase room.Phase
	label string
}

type captureNotifier struct{ got []endNotice }

func (c *captureNotifier) PhaseEnded(id room.ID, phase room.Phase, label string) {
	c.got = append(c.got, endNotice{id, phase, label})
}

func notifiedRoom() (*RoomHub, *captureNotifier) {
	n := &captureNotifier{}
	rh := newRoomHub(room.New("abcd2345", "hash", t0), "", slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	rh.notifier = n
	return rh, n
}

func TestPhaseEndIsAnnouncedOnceWhenItRunsOut(t *testing.T) {
	rh, n := notifiedRoom()
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 25 * time.Minute, Label: "essay"}, t0); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(24*time.Minute), time.Hour)
	if len(n.got) != 0 {
		t.Fatalf("announced too early: %+v", n.got)
	}
	rh.sweep(t0.Add(25*time.Minute+time.Second), time.Hour)
	rh.sweep(t0.Add(25*time.Minute+2*time.Second), time.Hour)
	if len(n.got) != 1 || n.got[0] != (endNotice{"abcd2345", room.PhaseFocus, "essay"}) {
		t.Fatalf("announced %+v", n.got)
	}
}

func TestBreakEndIsAnnouncedToo(t *testing.T) {
	rh, n := notifiedRoom()
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseBreak, Duration: 5 * time.Minute}, t0); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(5*time.Minute+time.Second), time.Hour)
	if len(n.got) != 1 || n.got[0].phase != room.PhaseBreak {
		t.Fatalf("announced %+v", n.got)
	}
}

func TestResetAndPauseAnnounceNothing(t *testing.T) {
	rh, n := notifiedRoom()
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 25 * time.Minute}, t0); err != nil {
		t.Fatal(err)
	}
	if err := rh.Apply(room.TimerPause{}, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(40*time.Minute), time.Hour) // paused: it never runs out
	if err := rh.Apply(room.TimerReset{}, t0.Add(41*time.Minute)); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(70*time.Minute), time.Hour)
	if len(n.got) != 0 {
		t.Fatalf("announced %+v", n.got)
	}
}

func TestAPhaseThatEndedLongAgoIsNotAnnounced(t *testing.T) {
	rh, n := notifiedRoom()
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 25 * time.Minute}, t0); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(3*time.Hour), time.Hour) // e.g. the server was busy or the room was restored late
	if len(n.got) != 0 {
		t.Fatalf("announced %+v", n.got)
	}
}
