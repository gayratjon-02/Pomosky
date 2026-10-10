package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

const testRoomID = "abcd2345"

// ---- test doubles

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type seqIDs struct{ n atomic.Int64 }

func (s *seqIDs) NewRoomID() room.ID   { return room.ID(fmt.Sprintf("room%04d", s.n.Add(1))) }
func (s *seqIDs) NewToken() string     { return fmt.Sprintf("token-%d", s.n.Add(1)) }
func (s *seqIDs) NewSessionID() string { return fmt.Sprintf("session-%d", s.n.Add(1)) }

// fakeHosts accepts the token "secret" for the room hash "hash".
type fakeHosts struct{}

func (fakeHosts) Verify(token, hash string) bool { return token == "secret" && hash == "hash" }

// ---- test server and client

type testServer struct {
	url   string
	hub   *Hub
	clock *fakeClock
}

func newTestServer(t *testing.T, tweak func(*HandlerConfig)) *testServer {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	clock := &fakeClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)}
	hub := NewHub(log, clock, nil)
	hub.Register(room.New(testRoomID, "hash", clock.Now()), "")

	cfg := HandlerConfig{Hosts: fakeHosts{}, Clock: clock, IDs: &seqIDs{}}
	if tweak != nil {
		tweak(&cfg)
	}
	h, err := NewHandler(hub, log, cfg)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /ws/{id}", h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &testServer{url: "ws" + strings.TrimPrefix(srv.URL, "http"), hub: hub, clock: clock}
}

type testClient struct {
	t   *testing.T
	c   *websocket.Conn
	ctx context.Context
}

func (s *testServer) dial(t *testing.T) *testClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	c, _, err := websocket.Dial(ctx, s.url+"/ws/"+testRoomID, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.CloseNow() })
	return &testClient{t: t, c: c, ctx: ctx}
}

func (tc *testClient) send(typ string, data any) {
	tc.t.Helper()
	b, err := Encode(typ, data)
	if err != nil {
		tc.t.Fatal(err)
	}
	tc.sendRaw(b)
}

func (tc *testClient) sendRaw(b []byte) {
	tc.t.Helper()
	if err := tc.c.Write(tc.ctx, websocket.MessageText, b); err != nil {
		tc.t.Fatal(err)
	}
}

// next reads one frame, failing the test if none arrives in time.
func (tc *testClient) next() Envelope {
	tc.t.Helper()
	ctx, cancel := context.WithTimeout(tc.ctx, 3*time.Second)
	defer cancel()
	_, b, err := tc.c.Read(ctx)
	if err != nil {
		tc.t.Fatalf("expected a frame: %v", err)
	}
	env, err := Decode(b)
	if err != nil {
		tc.t.Fatal(err)
	}
	return env
}

// waitFor skips frames until one of the given type arrives.
func (tc *testClient) waitFor(typ string) Envelope {
	tc.t.Helper()
	for i := 0; i < 50; i++ {
		if env := tc.next(); env.Type == typ {
			return env
		}
	}
	tc.t.Fatalf("no %q frame received", typ)
	return Envelope{}
}

// waitState skips frames until a state that satisfies ok arrives.
func (tc *testClient) waitState(ok func(StateData) bool) StateData {
	tc.t.Helper()
	for i := 0; i < 50; i++ {
		env := tc.next()
		if env.Type != TypeState {
			continue
		}
		var st StateData
		if err := json.Unmarshal(env.Data, &st); err != nil {
			tc.t.Fatal(err)
		}
		if ok(st) {
			return st
		}
	}
	tc.t.Fatal("no matching state received")
	return StateData{}
}

func (tc *testClient) waitError(code string) {
	tc.t.Helper()
	env := tc.waitFor(TypeError)
	var e ErrorData
	if err := json.Unmarshal(env.Data, &e); err != nil {
		tc.t.Fatal(err)
	}
	if e.Code != code {
		tc.t.Fatalf("error code = %q, want %q", e.Code, code)
	}
}

func (tc *testClient) becomeHost() {
	tc.t.Helper()
	tc.send(TypeHostAuth, HostAuthData{Token: "secret"})
	tc.waitFor(TypeHostOK)
}

func viewersIs(n int) func(StateData) bool { return func(s StateData) bool { return s.Viewers == n } }

// ---- tests

func TestViewerCountFollowsConnections(t *testing.T) {
	s := newTestServer(t, nil)
	a := s.dial(t)
	a.waitState(viewersIs(1))

	b := s.dial(t)
	a.waitState(viewersIs(2))
	b.waitState(viewersIs(2))

	b.c.CloseNow()
	a.waitState(viewersIs(1))
}

func TestOnlyTheHostControlsTheTimer(t *testing.T) {
	s := newTestServer(t, nil)
	viewer := s.dial(t)
	viewer.waitState(viewersIs(1))

	viewer.send(TypeTimerStart, TimerStartData{Phase: "focus", DurationMS: 1_500_000})
	viewer.waitError(CodeForbidden)

	viewer.send(TypeHostAuth, HostAuthData{Token: "wrong"})
	viewer.waitError(CodeForbidden)
	viewer.send(TypeTimerStart, TimerStartData{Phase: "focus", DurationMS: 1_500_000})
	viewer.waitError(CodeForbidden)

	viewer.becomeHost()
	viewer.send(TypeTimerStart, TimerStartData{Phase: "focus", DurationMS: 1_500_000, Label: "api"})
	st := viewer.waitState(func(s StateData) bool { return s.Timer.Phase == "focus" })
	if st.Timer.Label != "api" || st.Timer.DurationMS != 1_500_000 {
		t.Fatalf("unexpected timer: %+v", st.Timer)
	}
}

func TestThreeClientsSeeTheSameTimer(t *testing.T) {
	s := newTestServer(t, nil)
	host, v1, v2 := s.dial(t), s.dial(t), s.dial(t)
	for _, c := range []*testClient{host, v1, v2} {
		c.waitState(viewersIs(3))
	}
	host.becomeHost()

	host.send(TypeTimerStart, TimerStartData{Phase: "focus", DurationMS: 25 * 60 * 1000})
	running := func(s StateData) bool { return s.Timer.Phase == "focus" }
	states := []StateData{host.waitState(running), v1.waitState(running), v2.waitState(running)}

	for i, st := range states {
		remaining := st.Timer.StartedAtMS + st.Timer.DurationMS - st.ServerNowMS
		if remaining != 25*60*1000 {
			t.Fatalf("client %d computes remaining = %d ms, want 1500000", i, remaining)
		}
		if st.Timer != states[0].Timer {
			t.Fatalf("client %d sees a different timer: %+v vs %+v", i, st.Timer, states[0].Timer)
		}
	}

	// Pause and resume reach everyone too.
	s.clock.Advance(10 * time.Minute)
	host.send(TypeTimerPause, nil)
	paused := func(s StateData) bool { return s.Timer.Phase == "paused" }
	for _, c := range []*testClient{host, v1, v2} {
		st := c.waitState(paused)
		if st.Timer.PausedRemainingMS != 15*60*1000 || st.Timer.PausedFrom != "focus" {
			t.Fatalf("unexpected paused timer: %+v", st.Timer)
		}
	}
}

func TestPingAnswersWithServerTime(t *testing.T) {
	s := newTestServer(t, nil)
	c := s.dial(t)
	s.clock.Advance(42 * time.Second)

	c.send(TypePing, nil)
	env := c.waitFor(TypePong)
	var pong PongData
	if err := json.Unmarshal(env.Data, &pong); err != nil {
		t.Fatal(err)
	}
	if want := s.clock.Now().UnixMilli(); pong.ServerNowMS != want {
		t.Fatalf("serverNowMs = %d, want %d", pong.ServerNowMS, want)
	}
}

func TestReactionsReachEveryone(t *testing.T) {
	s := newTestServer(t, nil)
	a, b := s.dial(t), s.dial(t)
	a.waitState(viewersIs(2))
	b.waitState(viewersIs(2))

	a.send(TypeReaction, ReactionData{Kind: "fire"})
	for _, c := range []*testClient{a, b} {
		env := c.waitFor(TypeReactionTick)
		var tick ReactionTickData
		if err := json.Unmarshal(env.Data, &tick); err != nil {
			t.Fatal(err)
		}
		if tick.Kind != "fire" || tick.Count != 1 {
			t.Fatalf("tick = %+v", tick)
		}
	}

	a.send(TypeReaction, ReactionData{Kind: "angry"})
	a.waitError(CodeInvalid)
}

func TestBadInputIsAnsweredAndTheConnectionStaysOpen(t *testing.T) {
	s := newTestServer(t, nil)
	c := s.dial(t)
	c.waitState(viewersIs(1))

	c.sendRaw([]byte("this is not json"))
	c.waitError(CodeBadRequest)

	c.send("no.such.type", nil)
	c.waitError(CodeBadRequest)

	c.send(TypeJoin, nil) // missing data
	c.waitError(CodeBadRequest)

	c.send(TypeJoin, JoinData{Nickname: "x"}) // too short
	c.waitError(CodeInvalid)

	c.becomeHost()
	c.send(TypeTimerStart, TimerStartData{Phase: "party", DurationMS: 60_000})
	c.waitError(CodeInvalid)
	c.send(TypeTimerStart, TimerStartData{Phase: "focus", DurationMS: 1 << 62}) // would overflow time.Duration
	c.waitError(CodeInvalid)
	c.send(TypeTimerResume, nil)
	c.waitError(CodeInvalidTransition)

	c.send(TypePing, nil)
	c.waitFor(TypePong) // still alive
}

func TestValidNicknameIsAccepted(t *testing.T) {
	s := newTestServer(t, nil)
	c := s.dial(t)
	c.waitState(viewersIs(1))
	c.send(TypeJoin, JoinData{Nickname: "Haeyeon"})
	c.send(TypePing, nil)
	// The first non-state frame after a valid join must be the pong, not an error.
	for {
		env := c.next()
		if env.Type == TypeState {
			continue
		}
		if env.Type != TypePong {
			t.Fatalf("got %q after a valid join, want pong", env.Type)
		}
		return
	}
}

func TestMessageRateIsLimited(t *testing.T) {
	s := newTestServer(t, func(c *HandlerConfig) { c.MsgRatePerSec = 1; c.MsgBurst = 3 })
	c := s.dial(t)
	c.waitState(viewersIs(1))

	for i := 0; i < 10; i++ {
		c.send(TypePing, nil)
	}
	c.waitError(CodeRateLimited)
}

func TestOversizedMessageClosesTheConnection(t *testing.T) {
	s := newTestServer(t, func(c *HandlerConfig) { c.MaxMessageBytes = 256 })
	c := s.dial(t)
	c.waitState(viewersIs(1))

	c.sendRaw([]byte(strings.Repeat("x", 2048)))
	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()
	for {
		if _, _, err := c.c.Read(ctx); err != nil {
			if got := websocket.CloseStatus(err); got != websocket.StatusMessageTooBig {
				t.Fatalf("close status = %v (%v), want MessageTooBig", got, err)
			}
			return
		}
	}
}

func TestRepeatedWrongHostTokensCloseTheConnection(t *testing.T) {
	s := newTestServer(t, func(c *HandlerConfig) { c.MaxAuthFailures = 3 })
	c := s.dial(t)
	c.waitState(viewersIs(1))

	for i := 0; i < 3; i++ {
		c.send(TypeHostAuth, HostAuthData{Token: "guess"})
	}
	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()
	for {
		if _, _, err := c.c.Read(ctx); err != nil {
			if ctx.Err() != nil {
				t.Fatal("the server did not close the connection")
			}
			return
		}
	}
}

func TestIdleConnectionsAreClosed(t *testing.T) {
	s := newTestServer(t, func(c *HandlerConfig) { c.IdleTimeout = 150 * time.Millisecond })
	c := s.dial(t)
	c.waitState(viewersIs(1))

	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()
	for {
		if _, _, err := c.c.Read(ctx); err != nil {
			if ctx.Err() != nil {
				t.Fatal("an idle connection was not closed")
			}
			return
		}
	}
}

func TestMissingOrInvalidRoomIs404(t *testing.T) {
	s := newTestServer(t, nil)
	for _, id := range []string{"abcd2346", "ABCD2345", "short"} {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, resp, err := websocket.Dial(ctx, s.url+"/ws/"+id, nil)
		cancel()
		if err == nil {
			t.Fatalf("dialing room %q should fail", id)
		}
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			t.Fatalf("room %q: response = %+v, want 404", id, resp)
		}
	}
}

func TestHostRecoverRestoresAnIdleRoomOnly(t *testing.T) {
	s := newTestServer(t, nil)
	host := s.dial(t)
	host.waitState(viewersIs(1))
	host.becomeHost()

	started := s.clock.Now().Add(-10 * time.Minute).UnixMilli()
	host.send(TypeTimerRecover, TimerRecoverData{Timer: TimerView{
		Phase: "focus", StartedAtMS: started, DurationMS: 25 * 60 * 1000, Label: "restored",
	}})
	st := host.waitState(func(s StateData) bool { return s.Timer.Phase == "focus" })
	if st.Timer.Label != "restored" || st.Timer.StartedAtMS != started {
		t.Fatalf("unexpected recovered timer: %+v", st.Timer)
	}

	// A second recovery must not overwrite the running timer.
	host.send(TypeTimerRecover, TimerRecoverData{Timer: TimerView{
		Phase: "break", StartedAtMS: started, DurationMS: 60_000, Label: "other",
	}})
	if got := s.hub.mustRoom(t, testRoomID).State().Timer.Label; got != "restored" {
		t.Fatalf("the timer was overwritten: label = %q", got)
	}
}

// mustRoom is a test helper to reach the live room.
func (h *Hub) mustRoom(t *testing.T, id room.ID) *RoomHub {
	t.Helper()
	rh, ok := h.Room(id)
	if !ok {
		t.Fatalf("room %q is not live", id)
	}
	return rh
}

func TestOnlyTheHostSetsTheLocationAndItIsRounded(t *testing.T) {
	s := newTestServer(t, nil)
	viewer := s.dial(t)
	viewer.waitState(viewersIs(1))

	viewer.send(TypeLocationSet, LocationSetData{City: "Seoul", Lat: 37.5665, Lon: 126.978, TZ: "Asia/Seoul"})
	viewer.waitError(CodeForbidden)

	viewer.becomeHost()
	viewer.send(TypeLocationSet, LocationSetData{City: "Seoul", Lat: 37.56654321, Lon: 126.97801234, TZ: "Asia/Seoul"})
	st := viewer.waitState(func(s StateData) bool { return s.City == "Seoul" })
	if st.Location == nil || st.Location.Lat != 37.57 || st.Location.Lon != 126.98 || st.Location.TZ != "Asia/Seoul" {
		t.Fatalf("location = %+v", st.Location)
	}
}

func TestBadLocationsAreRejected(t *testing.T) {
	s := newTestServer(t, nil)
	host := s.dial(t)
	host.waitState(viewersIs(1))
	host.becomeHost()

	for name, d := range map[string]LocationSetData{
		"latitude out of range":  {City: "x", Lat: 91, Lon: 0, TZ: "UTC"},
		"longitude out of range": {City: "x", Lat: 0, Lon: 181, TZ: "UTC"},
		"unknown zone":           {City: "x", Lat: 1, Lon: 1, TZ: "Mars/Olympus"},
		"missing zone":           {City: "x", Lat: 1, Lon: 1},
		"long city":              {City: strings.Repeat("a", 200), Lat: 1, Lon: 1, TZ: "UTC"},
	} {
		host.send(TypeLocationSet, d)
		host.waitError(CodeInvalid)
		if got := s.hub.mustRoom(t, testRoomID).State().Location; got.TZ != "" {
			t.Fatalf("%s: location was stored: %+v", name, got)
		}
	}
}

func TestConnectionsPerClientAreLimited(t *testing.T) {
	s := newTestServer(t, func(c *HandlerConfig) {
		c.MaxConnsPerIP = 2
		c.ClientIP = func(*http.Request) string { return "1.2.3.4" }
	})
	a := s.dial(t)
	b := s.dial(t)
	a.waitState(viewersIs(1))
	b.waitState(viewersIs(2))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, resp, err := websocket.Dial(ctx, s.url+"/ws/"+testRoomID, nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("third connection: err=%v resp=%v", err, resp)
	}

	// Closing one frees a slot.
	a.c.Close(websocket.StatusNormalClosure, "bye")
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, _, err := websocket.Dial(ctx, s.url+"/ws/"+testRoomID, nil)
		if err == nil {
			c.CloseNow()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("slot was not released: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
