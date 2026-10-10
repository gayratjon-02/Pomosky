package realtime

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

type captureRecorder struct{ got []models.FocusSession }

func (c *captureRecorder) RecordFocus(s models.FocusSession) { c.got = append(c.got, s) }

var t0 = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

func ownedRoom(owner models.UserID) (*RoomHub, *captureRecorder) {
	rec := &captureRecorder{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return newRoomHub(room.New("abcd2345", "hash", t0), owner, log, rec), rec
}

func startFocus(t *testing.T, rh *RoomHub, at time.Time, d time.Duration) {
	t.Helper()
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: d}, at); err != nil {
		t.Fatal(err)
	}
}

func TestResetRecordsTheFocusedTime(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)
	if err := rh.Apply(room.TimerReset{}, t0.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if len(rec.got) != 1 || rec.got[0].Focused != 10*time.Minute || rec.got[0].RoomID != "abcd2345" {
		t.Fatalf("recorded %+v", rec.got)
	}
	if !rec.got[0].StartedAt.Equal(t0) || !rec.got[0].EndedAt.Equal(t0.Add(10*time.Minute)) {
		t.Fatalf("times %v - %v", rec.got[0].StartedAt, rec.got[0].EndedAt)
	}
}

func TestNaturalEndIsRecordedOnceBySweep(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)

	rh.sweep(t0.Add(24*time.Minute), time.Hour) // not over yet
	if len(rec.got) != 0 {
		t.Fatalf("recorded too early: %+v", rec.got)
	}
	rh.sweep(t0.Add(26*time.Minute), time.Hour)
	rh.sweep(t0.Add(27*time.Minute), time.Hour) // a second sweep must not record it again
	if len(rec.got) != 1 || rec.got[0].Focused != 25*time.Minute || !rec.got[0].StartedAt.Equal(t0) {
		t.Fatalf("recorded %+v", rec.got)
	}
	// Resetting later does not count the same phase twice.
	if err := rh.Apply(room.TimerReset{}, t0.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if len(rec.got) != 1 {
		t.Fatalf("double count: %+v", rec.got)
	}
}

func TestPausedTimeIsNotCounted(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)
	if err := rh.Apply(room.TimerPause{}, t0.Add(5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := rh.Apply(room.TimerReset{}, t0.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if len(rec.got) != 1 || rec.got[0].Focused != 5*time.Minute {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestStartingAfterTheEndRecordsThePhaseOnce(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)

	// The next phase may start once the focus phase has run to its end; the
	// focus time is recorded once, whether or not the sweep saw it first.
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseBreak, Duration: 5 * time.Minute}, t0.Add(26*time.Minute)); err != nil {
		t.Fatal(err)
	}
	rh.sweep(t0.Add(27*time.Minute), time.Hour)
	if len(rec.got) != 1 || rec.got[0].Focused != 25*time.Minute {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestStartingWhileRunningIsRefusedAndRecordsNothing(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)
	err := rh.Apply(room.TimerStart{Phase: room.PhaseFocus, Duration: 50 * time.Minute}, t0.Add(8*time.Minute))
	if !errors.Is(err, room.ErrInvalidTransition) {
		t.Fatalf("start while running: %v", err)
	}
	if len(rec.got) != 0 || rh.State().Timer.Duration != 25*time.Minute {
		t.Fatalf("the refused start changed something: %+v %+v", rec.got, rh.State().Timer)
	}
}

func TestAnonymousRoomsAreRecordedAsAnonymous(t *testing.T) {
	anon, rec := ownedRoom("")
	startFocus(t, anon, t0, 25*time.Minute)
	_ = anon.Apply(room.TimerReset{}, t0.Add(10*time.Minute))
	if len(rec.got) != 1 || !rec.got[0].Anonymous || rec.got[0].Focused != 10*time.Minute {
		t.Fatalf("recorded %+v", rec.got)
	}
	owned, ownedRec := ownedRoom("u1")
	startFocus(t, owned, t0, 25*time.Minute)
	_ = owned.Apply(room.TimerReset{}, t0.Add(10*time.Minute))
	if len(ownedRec.got) != 1 || ownedRec.got[0].Anonymous {
		t.Fatalf("owned room recorded %+v", ownedRec.got)
	}
}

func TestNothingIsRecordedForBreaksOrBlips(t *testing.T) {

	rh, rec := ownedRoom("u1")
	if err := rh.Apply(room.TimerStart{Phase: room.PhaseBreak, Duration: 5 * time.Minute}, t0); err != nil {
		t.Fatal(err)
	}
	_ = rh.Apply(room.TimerReset{}, t0.Add(4*time.Minute))
	startFocus(t, rh, t0.Add(10*time.Minute), 25*time.Minute)
	_ = rh.Apply(room.TimerReset{}, t0.Add(10*time.Minute+10*time.Second)) // 10 s: an accidental start
	if len(rec.got) != 0 {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestPeakViewersIsRecorded(t *testing.T) {
	rh, rec := ownedRoom("u1")
	a, b := &Client{ID: "a"}, &Client{ID: "b"}
	rh.Join(a)
	rh.Join(b)
	rh.Leave(b)
	startFocus(t, rh, t0, 25*time.Minute)
	_ = rh.Apply(room.TimerReset{}, t0.Add(5*time.Minute))
	if len(rec.got) != 1 || rec.got[0].PeakViewers != 2 {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestEmptyRoomsExpireAndRecordTheRunningPhase(t *testing.T) {
	rh, rec := ownedRoom("u1")
	c := &Client{ID: "a"}
	rh.Join(c)
	startFocus(t, rh, t0, 60*time.Minute)

	if rh.sweep(t0.Add(time.Minute), 10*time.Minute) {
		t.Fatal("a room with a client expired")
	}
	rh.Leave(c)
	if rh.sweep(t0.Add(2*time.Minute), 10*time.Minute) {
		t.Fatal("expired as soon as it became empty")
	}
	if rh.sweep(t0.Add(11*time.Minute), 10*time.Minute) {
		t.Fatal("expired before the idle time passed")
	}
	if !rh.sweep(t0.Add(13*time.Minute), 10*time.Minute) {
		t.Fatal("did not expire after the idle time")
	}
	if len(rec.got) != 1 || rec.got[0].Focused != 13*time.Minute {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestReconnectingClientKeepsTheRoomAlive(t *testing.T) {
	rh, _ := ownedRoom("")
	c := &Client{ID: "a"}
	rh.Join(c)
	rh.Leave(c)
	rh.sweep(t0, 10*time.Minute) // starts the empty timer
	rh.Join(c)                   // the host reloaded the page
	rh.sweep(t0.Add(5*time.Minute), 10*time.Minute)
	rh.Leave(c)
	if rh.sweep(t0.Add(12*time.Minute), 10*time.Minute) {
		t.Fatal("the empty timer was not reset by the reconnect")
	}
}
