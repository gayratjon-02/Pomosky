package realtime

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Message types. See docs/ws-protocol.md and keep web/src/core/protocol.ts in sync.
const (
	// client -> server
	TypeJoin         = "join"
	TypeHostAuth     = "host.auth"
	TypeTimerStart   = "timer.start"
	TypeTimerPause   = "timer.pause"
	TypeTimerResume  = "timer.resume"
	TypeTimerReset   = "timer.reset"
	TypeTimerRecover = "timer.recover"
	TypeReaction     = "reaction"
	TypeLocationSet  = "location.set"
	TypePing         = "ping"

	// server -> client
	TypeState        = "state"
	TypeHostOK       = "host.ok"
	TypeReactionTick = "reaction.tick"
	TypeWeather      = "weather" // reserved for stage 6
	TypePong         = "pong"
	TypeError        = "error"
)

// Error codes sent in ErrorData.Code.
const (
	CodeBadRequest        = "bad_request"
	CodeForbidden         = "forbidden"
	CodeInvalid           = "invalid"
	CodeInvalidTransition = "invalid_transition"
	CodeRateLimited       = "rate_limited"
)

// maxPhaseMS bounds client-supplied durations before they are converted to
// time.Duration, so a huge value cannot overflow. The core applies the real
// limits (room.MaxPhaseDuration).
const maxPhaseMS = int64(100 * time.Hour / time.Millisecond)

// ErrBadTimer is returned for a timer that cannot be converted.
var ErrBadTimer = errors.New("realtime: invalid timer")

// Envelope is the frame of every message.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Client -> server payloads.

type JoinData struct {
	Nickname string `json:"nickname"`
}

type HostAuthData struct {
	Token string `json:"token"`
}

type TimerStartData struct {
	Phase      string `json:"phase"`
	DurationMS int64  `json:"durationMs"`
	Label      string `json:"label"`
}

type TimerRecoverData struct {
	Timer TimerView `json:"timer"`
}

type LocationSetData struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	TZ   string  `json:"tz"`
}

type ReactionData struct {
	Kind string `json:"kind"`
}

// Server -> client payloads.

type TimerView struct {
	Phase             string `json:"phase"`
	PausedFrom        string `json:"pausedFrom,omitempty"`
	StartedAtMS       int64  `json:"startedAtMs"`
	DurationMS        int64  `json:"durationMs"`
	PausedRemainingMS int64  `json:"pausedRemainingMs"`
	Label             string `json:"label"`
}

type StateData struct {
	RoomID      string         `json:"roomId"`
	Timer       TimerView      `json:"timer"`
	Viewers     int            `json:"viewers"`
	Reactions   map[string]int `json:"reactions"`
	City        string         `json:"city"`
	Location    *LocationView  `json:"location,omitempty"` // the host's rounded place; absent until the host sets it
	ServerNowMS int64          `json:"serverNowMs"`
}

// LocationView is the host's place, rounded to about 1 km by the core.
type LocationView struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	TZ  string  `json:"tz"`
}

type ReactionTickData struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

type PongData struct {
	ServerNowMS int64 `json:"serverNowMs"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Encode builds a framed message.
func Encode(typ string, data any) ([]byte, error) {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	return json.Marshal(Envelope{Type: typ, Data: raw})
}

// Decode parses a framed message.
func Decode(b []byte) (Envelope, error) {
	var e Envelope
	err := json.Unmarshal(b, &e)
	return e, err
}

// ToTimerView converts the core timer to its wire form.
func ToTimerView(t room.Timer) TimerView {
	var started int64
	if !t.StartedAt.IsZero() {
		started = t.StartedAt.UnixMilli()
	}
	return TimerView{
		Phase:             string(t.Phase),
		PausedFrom:        string(t.PausedFrom),
		StartedAtMS:       started,
		DurationMS:        t.Duration.Milliseconds(),
		PausedRemainingMS: t.PausedRemaining.Milliseconds(),
		Label:             t.Label,
	}
}

// FromTimerView converts a client-supplied timer to the core form. It only
// rejects values that cannot be represented; the core validates the semantics.
func FromTimerView(v TimerView) (room.Timer, error) {
	if v.DurationMS < 0 || v.DurationMS > maxPhaseMS ||
		v.PausedRemainingMS < 0 || v.PausedRemainingMS > maxPhaseMS {
		return room.Timer{}, ErrBadTimer
	}
	t := room.Timer{
		Phase:           room.Phase(v.Phase),
		PausedFrom:      room.Phase(v.PausedFrom),
		Duration:        time.Duration(v.DurationMS) * time.Millisecond,
		PausedRemaining: time.Duration(v.PausedRemainingMS) * time.Millisecond,
		Label:           v.Label,
	}
	if v.StartedAtMS != 0 {
		t.StartedAt = time.UnixMilli(v.StartedAtMS).UTC()
	}
	return t, nil
}

// ToTimerStart converts a client request to the core event.
func ToTimerStart(d TimerStartData) (room.TimerStart, error) {
	if d.DurationMS < 0 || d.DurationMS > maxPhaseMS {
		return room.TimerStart{}, ErrBadTimer
	}
	return room.TimerStart{
		Phase:    room.Phase(d.Phase),
		Duration: time.Duration(d.DurationMS) * time.Millisecond,
		Label:    d.Label,
	}, nil
}

// ToStateData converts the core state to the snapshot sent to clients.
func ToStateData(s room.State, now time.Time) StateData {
	reactions := make(map[string]int, len(s.Reactions))
	for k, v := range s.Reactions {
		reactions[string(k)] = v
	}
	var loc *LocationView
	if l := s.Location; l.City != "" || l.TZ != "" || l.LatRounded != 0 || l.LonRounded != 0 {
		loc = &LocationView{Lat: l.LatRounded, Lon: l.LonRounded, TZ: l.TZ}
	}
	return StateData{
		Location:    loc,
		RoomID:      s.ID.String(),
		Timer:       ToTimerView(s.Timer),
		Viewers:     s.Viewers,
		Reactions:   reactions,
		City:        s.Location.City,
		ServerNowMS: now.UnixMilli(),
	}
}
