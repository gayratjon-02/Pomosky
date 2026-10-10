package realtime

import (
	"encoding/json"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Message types. See docs/ws-protocol.md.
const (
	// client -> server
	TypeJoin        = "join"
	TypeHostAuth    = "host.auth"
	TypeTimerStart  = "timer.start"
	TypeTimerPause  = "timer.pause"
	TypeTimerResume = "timer.resume"
	TypeTimerReset  = "timer.reset"
	TypeReaction    = "reaction"
	TypePing        = "ping"

	// server -> client
	TypeState        = "state"
	TypeTimerUpdated = "timer.updated"
	TypePresence     = "presence"
	TypeReactionTick = "reaction.tick"
	TypeWeather      = "weather"
	TypePong         = "pong"
	TypeError        = "error"
)

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

type ReactionData struct {
	Kind string `json:"kind"`
}

// Server -> client payloads.

type TimerView struct {
	Phase             string `json:"phase"`
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
	ServerNowMS int64          `json:"serverNowMs"`
}

type PresenceData struct {
	Viewers int `json:"viewers"`
}

type ReactionTickData struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
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
		StartedAtMS:       started,
		DurationMS:        t.Duration.Milliseconds(),
		PausedRemainingMS: t.PausedRemaining.Milliseconds(),
		Label:             t.Label,
	}
}

// ToStateData converts the core state to the snapshot sent on join.
func ToStateData(s room.State, now time.Time) StateData {
	reactions := make(map[string]int, len(s.Reactions))
	for k, v := range s.Reactions {
		reactions[string(k)] = v
	}
	return StateData{
		RoomID:      s.ID.String(),
		Timer:       ToTimerView(s.Timer),
		Viewers:     s.Viewers,
		Reactions:   reactions,
		City:        s.Location.City,
		ServerNowMS: now.UnixMilli(),
	}
}
