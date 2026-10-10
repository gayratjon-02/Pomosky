package realtime

import (
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

	rh.sweep(t0.Add(24 * time.Minute)) // not over yet
	if len(rec.got) != 0 {
		t.Fatalf("recorded too early: %+v", rec.got)
	}
	rh.sweep(t0.Add(26 * time.Minute))
	rh.sweep(t0.Add(27 * time.Minute)) // a second sweep must not record it again
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

func TestStartingANewPhaseEndsTheOldOne(t *testing.T) {
	rh, rec := ownedRoom("u1")
	startFocus(t, rh, t0, 25*time.Minute)
	startFocus(t, rh, t0.Add(8*time.Minute), 50*time.Minute)
	if len(rec.got) != 1 || rec.got[0].Focused != 8*time.Minute {
		t.Fatalf("recorded %+v", rec.got)
	}
}

func TestNothingIsRecordedForAnonymousBreaksOrBlips(t *testing.T) {
	anon, anonRec := ownedRoom("")
	startFocus(t, anon, t0, 25*time.Minute)
	_ = anon.Apply(room.TimerReset{}, t0.Add(10*time.Minute))
	anon.sweep(t0.Add(time.Hour))
	if len(anonRec.got) != 0 {
		t.Fatalf("anonymous room recorded %+v", anonRec.got)
	}

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
