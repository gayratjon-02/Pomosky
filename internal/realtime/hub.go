package realtime

import (
	"bytes"
	"context"
	"log/slog"
	"sort"
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
	notifier ports.PhaseNotifier // may be nil
	idleTTL  time.Duration       // how long a room without any client is kept

	mu    sync.RWMutex
	rooms map[room.ID]*RoomHub
}

var (
	_ ports.Broadcaster   = (*Hub)(nil)
	_ ports.LiveRooms     = (*Hub)(nil)
	_ ports.LiveInspector = (*Hub)(nil)
)

// NewHub creates an empty Hub. recorder receives the finished focus phases of
// rooms that have an owner; it may be nil.
func NewHub(log *slog.Logger, clock ports.Clock, recorder ports.FocusRecorder) *Hub {
	return &Hub{log: log, clock: clock, recorder: recorder, idleTTL: DefaultIdleTTL, rooms: map[room.ID]*RoomHub{}}
}

// DefaultIdleTTL is how long a room with no connected client stays alive. It
// is long enough for the host to reload the page or lose the network.
const DefaultIdleTTL = 10 * time.Minute

// SetPhaseNotifier sets who is told when a phase ends (call before serving).
func (h *Hub) SetPhaseNotifier(n ports.PhaseNotifier) { h.notifier = n }

// SetIdleTTL changes how long an empty room is kept (call before serving).
func (h *Hub) SetIdleTTL(d time.Duration) { h.idleTTL = d }

// Open returns the live RoomHub for state.ID, creating it from state if it
// does not exist yet. owner is "" for anonymous rooms.
func (h *Hub) Open(state room.State, owner models.UserID) *RoomHub {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rh, ok := h.rooms[state.ID]; ok {
		return rh
	}
	rh := newRoomHub(state, owner, h.log, h.recorder)
	rh.notifier = h.notifier
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

// Location implements ports.LiveRooms.
func (h *Hub) Location(id room.ID) (room.Location, bool) {
	rh, ok := h.Room(id)
	if !ok {
		return room.Location{}, false
	}
	return rh.State().Location, true
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

// Sweep does the periodic work of the rooms: it records the focus phases that
// ran to their end (nothing else notices that, because the timer is only a
// start time and a duration) and removes rooms that nobody has been connected
// to for the idle TTL. Call it every few seconds (see RunSweeper).
func (h *Hub) Sweep(now time.Time) {
	h.mu.RLock()
	rooms := make(map[room.ID]*RoomHub, len(h.rooms))
	for id, rh := range h.rooms {
		rooms[id] = rh
	}
	h.mu.RUnlock()
	for id, rh := range rooms {
		if rh.sweep(now, h.idleTTL) {
			h.Close(id)
			h.log.Info("room closed after being empty", "room", id)
		}
	}
}

// LiveSummary implements ports.LiveInspector: counts of rooms, connected
// browsers and timer phases right now. Nothing identifying is returned.
func (h *Hub) LiveSummary() models.LiveSummary {
	h.mu.RLock()
	rooms := make([]*RoomHub, 0, len(h.rooms))
	for _, rh := range h.rooms {
		rooms = append(rooms, rh)
	}
	h.mu.RUnlock()

	var sum models.LiveSummary
	sum.Rooms = len(rooms)
	for _, rh := range rooms {
		rh.mu.Lock()
		sum.Clients += len(rh.clients)
		switch rh.state.Timer.Phase {
		case room.PhaseFocus:
			sum.Focus++
		case room.PhaseBreak:
			sum.Break++
		case room.PhasePaused:
			sum.Paused++
		default:
			sum.Idle++
		}
		rh.mu.Unlock()
	}
	return sum
}

// Len returns the number of live rooms.
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
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
	notifier ports.PhaseNotifier // may be nil

	mu       sync.Mutex
	state    room.State
	clients  map[*Client]struct{}
	peak     int       // most clients seen since the last recorded focus phase
	credited time.Time // StartedAt of the focus phase already recorded at its natural end
	notified time.Time // StartedAt of the phase whose end was already announced
	emptyAt  time.Time // since when no client is connected; zero while someone is
	people   []byte    // the last people message sent, to skip identical ones
}

func newRoomHub(state room.State, owner models.UserID, log *slog.Logger, recorder ports.FocusRecorder) *RoomHub {
	r := &RoomHub{log: log, owner: owner, recorder: recorder, state: state, clients: map[*Client]struct{}{}}
	r.people, _ = Encode(TypePeople, r.peopleLocked()) // an empty room has an empty list: no news
	return r
}

// recordLocked writes the focus time of a timer that is ending to the recorder.
func (r *RoomHub) recordLocked(t room.Timer, now time.Time) {
	if r.recorder == nil {
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
		Anonymous:   r.owner == "",
	})
	r.peak = len(r.clients)
}

// sweep records a focus phase that ended on its own and reports whether the
// room has been empty for longer than idleTTL and should be closed.
func (r *RoomHub) sweep(now time.Time, idleTTL time.Duration) (expired bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := r.state.Timer
	if t.Phase == room.PhaseFocus && t.Remaining(now) == 0 && !t.StartedAt.Equal(r.credited) {
		r.recordLocked(t, t.StartedAt.Add(t.Duration)) // at its end, the whole phase has been focused
		r.credited = t.StartedAt
	}

	r.notifyEndLocked(t, now)

	if len(r.clients) > 0 {
		r.emptyAt = time.Time{}
		return false
	}
	if r.emptyAt.IsZero() {
		r.emptyAt = now
		return false
	}
	if now.Sub(r.emptyAt) < idleTTL {
		return false
	}
	r.recordLocked(t, now) // a phase still running when the room closes counts up to now
	return true
}

// lateNotice is how long after its end a phase is still announced. A timer
// that ended long ago (a room restored after a restart) must not ping anyone.
const lateNotice = 2 * time.Minute

// notifyEndLocked tells the notifier, once, that a running phase has run out.
func (r *RoomHub) notifyEndLocked(t room.Timer, now time.Time) {
	if r.notifier == nil || (t.Phase != room.PhaseFocus && t.Phase != room.PhaseBreak) {
		return
	}
	end := t.StartedAt.Add(t.Duration)
	if now.Before(end) || t.StartedAt.Equal(r.notified) {
		return
	}
	r.notified = t.StartedAt
	if now.Sub(end) > lateNotice {
		return
	}
	r.notifier.PhaseEnded(r.state.ID, t.Phase, t.Label)
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

// MaxPeople is how many named people one room lists; the rest are counted.
const MaxPeople = 12

// Announce sends the list of named people to the clients if it changed since
// the last time. Call it after a client joined, left or chose a nickname.
func (r *RoomHub) Announce() {
	r.mu.Lock()
	defer r.mu.Unlock()
	payload, err := Encode(TypePeople, r.peopleLocked())
	if err != nil || bytes.Equal(payload, r.people) {
		return
	}
	r.people = payload
	r.broadcastLocked(payload)
}

// People returns the current people message for a client that just joined.
func (r *RoomHub) People() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	payload, err := Encode(TypePeople, r.peopleLocked())
	if err != nil {
		return nil
	}
	return payload
}

// peopleLocked lists the named people: signed-in users who allow it (once,
// however many tabs they have open) and viewers with a nickname. The order is
// stable so the message only changes when the people do.
func (r *RoomHub) peopleLocked() PeopleData {
	type entry struct {
		key string
		PersonView
	}
	seen := map[string]bool{}
	var all []entry
	for c := range r.clients {
		switch {
		case c.Account != "":
			if !seen["a:"+c.Account] {
				seen["a:"+c.Account] = true
				all = append(all, entry{"a:" + c.Account, PersonView{Name: c.AccountName, AvatarURL: c.AccountAvatar}})
			}
		case c.Nickname != "":
			all = append(all, entry{"n:" + c.ID, PersonView{Name: c.Nickname}})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].key < all[j].key })
	out := PeopleData{People: []PersonView{}}
	for i, e := range all {
		if i >= MaxPeople {
			out.More = len(all) - MaxPeople
			break
		}
		out.People = append(out.People, e.PersonView)
	}
	return out
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
