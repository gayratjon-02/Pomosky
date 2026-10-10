package realtime

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Hub owns all live rooms of this server instance.
type Hub struct {
	log      *slog.Logger
	clock    ports.Clock
	recorder ports.FocusRecorder // may be nil

	mu    sync.RWMutex
	rooms map[room.ID]*RoomHub
}

var (
	_ ports.Broadcaster = (*Hub)(nil)
	_ ports.LiveRooms   = (*Hub)(nil)
)

// NewHub creates an empty Hub. recorder receives the finished focus phases of
// rooms that have an owner; it may be nil.
func NewHub(log *slog.Logger, clock ports.Clock, recorder ports.FocusRecorder) *Hub {
	return &Hub{log: log, clock: clock, recorder: recorder, rooms: map[room.ID]*RoomHub{}}
}

// Open returns the live RoomHub for state.ID, creating it from state if it
// does not exist yet. owner is "" for anonymous rooms.
func (h *Hub) Open(state room.State, owner models.UserID) *RoomHub {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rh, ok := h.rooms[state.ID]; ok {
		return rh
	}
	rh := newRoomHub(state, owner, h.log, h.recorder)
	h.rooms[state.ID] = rh
	return rh
}

// Room returns the live RoomHub for id.
func (h *Hub) Room(id room.ID) (*RoomHub, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rh, ok := h.rooms[id]
	return rh, ok
}

// Register implements ports.LiveRooms. An already registered room is kept as is.
func (h *Hub) Register(state room.State, owner models.UserID) { h.Open(state, owner) }

// Has implements ports.LiveRooms.
func (h *Hub) Has(id room.ID) bool {
	_, ok := h.Room(id)
	return ok
}

// Close removes a room from the hub.
func (h *Hub) Close(id room.ID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, id)
}

// Publish implements ports.Broadcaster.
func (h *Hub) Publish(id room.ID, payload []byte) {
	if rh, ok := h.Room(id); ok {
		rh.Broadcast(payload)
	}
}

// Sweep records the focus phases that ran to their end in rooms with an owner.
// Nothing else notices the end of a phase, because the timer is only a start
// time and a duration. Call it every few seconds (see RunSweeper).
func (h *Hub) Sweep(now time.Time) {
	h.mu.RLock()
	rooms := make([]*RoomHub, 0, len(h.rooms))
	for _, rh := range h.rooms {
		rooms = append(rooms, rh)
	}
	h.mu.RUnlock()
	for _, rh := range rooms {
		rh.sweep(now)
	}
}

// RunSweeper calls Sweep every interval until ctx is done.
func (h *Hub) RunSweeper(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			h.Sweep(now)
		}
	}
}

// minRecordedFocus is the shortest focus time that counts toward statistics;
// it keeps accidental starts out.
const minRecordedFocus = 30 * time.Second

// RoomHub is the live side of one room: its state and its connected clients.
type RoomHub struct {
	log      *slog.Logger
	owner    models.UserID // "" for anonymous rooms
	recorder ports.FocusRecorder

	mu       sync.Mutex
	state    room.State
	clients  map[*Client]struct{}
	peak     int       // most clients seen since the last recorded focus phase
	credited time.Time // StartedAt of the focus phase already recorded at its natural end
}

func newRoomHub(state room.State, owner models.UserID, log *slog.Logger, recorder ports.FocusRecorder) *RoomHub {
	return &RoomHub{log: log, owner: owner, recorder: recorder, state: state, clients: map[*Client]struct{}{}}
}

// recordLocked writes the focus time of a timer that is ending to the recorder.
func (r *RoomHub) recordLocked(t room.Timer, now time.Time) {
	if r.owner == "" || r.recorder == nil {
		return
	}
	focused, ok := room.FocusedTime(t, now)
	if !ok || focused < minRecordedFocus {
		return
	}
	if t.Phase == room.PhaseFocus && t.StartedAt.Equal(r.credited) {
		return // already recorded when it ended on its own
	}
	r.recorder.RecordFocus(models.FocusSession{
		RoomID:      r.state.ID,
		StartedAt:   now.Add(-focused),
		EndedAt:     now,
		Focused:     focused,
		PeakViewers: r.peak,
	})
	r.peak = len(r.clients)
}

func (r *RoomHub) sweep(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := r.state.Timer
	if r.owner == "" || t.Phase != room.PhaseFocus || t.Remaining(now) > 0 || t.StartedAt.Equal(r.credited) {
		return
	}
	end := t.StartedAt.Add(t.Duration)
	r.recordLocked(t, end) // at its end, the whole phase has been focused
	r.credited = t.StartedAt
}

// State returns a copy of the current room state.
func (r *RoomHub) State() room.State {
	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneState(r.state)
}

// Join registers a client and returns the number of connected clients.
func (r *RoomHub) Join(c *Client) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c] = struct{}{}
	if len(r.clients) > r.peak {
		r.peak = len(r.clients)
	}
	return len(r.clients)
}

// Leave unregisters a client and returns the number of connected clients.
func (r *RoomHub) Leave(c *Client) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
	return len(r.clients)
}

// Broadcast sends payload to every client without blocking on slow ones.
func (r *RoomHub) Broadcast(payload []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.broadcastLocked(payload)
}

// Apply runs an event through the pure core and broadcasts the new state.
func (r *RoomHub) Apply(ev room.Event, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	prev := r.state.Timer
	next, err := room.Reduce(r.state, ev, now)
	if err != nil {
		return err
	}
	switch ev.(type) {
	case room.TimerStart, room.TimerReset:
		r.recordLocked(prev, now) // the previous focus phase ends here
	}
	r.state = next

	payload, err := Encode(TypeState, ToStateData(next, now))
	if err != nil {
		return err
	}
	r.broadcastLocked(payload)
	return nil
}

// React counts a reaction and broadcasts only a small tick, not the whole
// state, so a burst of reactions stays cheap.
func (r *RoomHub) React(kind room.ReactionKind, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	next, err := room.Reduce(r.state, room.ReactionSent{Kind: kind}, now)
	if err != nil {
		return err
	}
	r.state = next

	payload, err := Encode(TypeReactionTick, ReactionTickData{Kind: string(kind), Count: next.Reactions[kind]})
	if err != nil {
		return err
	}
	r.broadcastLocked(payload)
	return nil
}

func (r *RoomHub) broadcastLocked(payload []byte) {
	for c := range r.clients {
		if !c.Send(payload) {
			r.log.Debug("dropping message for slow client", "client", c.ID)
		}
	}
}

func cloneState(s room.State) room.State {
	reactions := make(map[room.ReactionKind]int, len(s.Reactions))
	for k, v := range s.Reactions {
		reactions[k] = v
	}
	s.Reactions = reactions
	return s
}
