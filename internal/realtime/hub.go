package realtime

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Hub owns all live rooms of this server instance.
type Hub struct {
	log   *slog.Logger
	clock ports.Clock

	mu    sync.RWMutex
	rooms map[room.ID]*RoomHub
}

var (
	_ ports.Broadcaster = (*Hub)(nil)
	_ ports.LiveRooms   = (*Hub)(nil)
)

// NewHub creates an empty Hub.
func NewHub(log *slog.Logger, clock ports.Clock) *Hub {
	return &Hub{log: log, clock: clock, rooms: map[room.ID]*RoomHub{}}
}

// Open returns the live RoomHub for state.ID, creating it from state if it
// does not exist yet.
func (h *Hub) Open(state room.State) *RoomHub {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rh, ok := h.rooms[state.ID]; ok {
		return rh
	}
	rh := newRoomHub(state, h.log)
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
func (h *Hub) Register(state room.State) { h.Open(state) }

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

// RoomHub is the live side of one room: its state and its connected clients.
type RoomHub struct {
	log *slog.Logger

	mu      sync.Mutex
	state   room.State
	clients map[*Client]struct{}
}

func newRoomHub(state room.State, log *slog.Logger) *RoomHub {
	return &RoomHub{log: log, state: state, clients: map[*Client]struct{}{}}
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

	next, err := room.Reduce(r.state, ev, now)
	if err != nil {
		return err
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
