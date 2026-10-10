package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/gayratjon-02/Pomosky/internal/moderation"
	"github.com/gayratjon-02/Pomosky/internal/platform/ratelimit"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// HostVerifier checks a host token against the hash stored for a room.
// auth.HostTokens implements it.
type HostVerifier interface {
	Verify(token, hash string) bool
}

type denyHosts struct{}

func (denyHosts) Verify(string, string) bool { return false }

// HandlerConfig configures the WebSocket handler. Clock and IDs are required;
// every other zero value falls back to a safe default.
type HandlerConfig struct {
	Hosts HostVerifier // nil means nobody can become host
	Clock ports.Clock
	IDs   ports.IDGenerator

	IdleTimeout     time.Duration // close a connection that sends nothing for this long (default 75s)
	WriteTimeout    time.Duration // default 5s
	MaxMessageBytes int64         // default 4096
	OutboxSize      int           // default 32
	MsgRatePerSec   float64       // all messages, per connection (default 20)
	MsgBurst        int           // default 40
	ReactRatePerSec float64       // reactions only, per connection (default 5)
	ReactBurst      int           // default 10
	MaxAuthFailures int           // wrong host tokens before the connection is closed (default 5)
}

func (c HandlerConfig) withDefaults() HandlerConfig {
	if c.Hosts == nil {
		c.Hosts = denyHosts{}
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 75 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 5 * time.Second
	}
	if c.MaxMessageBytes == 0 {
		c.MaxMessageBytes = 4096
	}
	if c.OutboxSize == 0 {
		c.OutboxSize = 32
	}
	if c.MsgRatePerSec == 0 {
		c.MsgRatePerSec = 20
	}
	if c.MsgBurst == 0 {
		c.MsgBurst = 40
	}
	if c.ReactRatePerSec == 0 {
		c.ReactRatePerSec = 5
	}
	if c.ReactBurst == 0 {
		c.ReactBurst = 10
	}
	if c.MaxAuthFailures == 0 {
		c.MaxAuthFailures = 5
	}
	return c
}

// Handler serves GET /ws/{id}: it upgrades the request to a WebSocket, joins
// the room and runs the read and write loops until the client goes away.
type Handler struct {
	hub *Hub
	log *slog.Logger
	cfg HandlerConfig
}

// NewHandler creates the WebSocket handler.
func NewHandler(hub *Hub, log *slog.Logger, cfg HandlerConfig) (*Handler, error) {
	if cfg.Clock == nil || cfg.IDs == nil {
		return nil, errors.New("realtime: HandlerConfig needs Clock and IDs")
	}
	return &Handler{hub: hub, log: log, cfg: cfg.withDefaults()}, nil
}

// ServeHTTP implements http.Handler. The browser's Origin must match the
// request host (the library's default same-origin check).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := room.ID(r.PathValue("id"))
	rh, ok := h.hub.Room(id)
	if !id.Valid() || !ok {
		http.NotFound(w, r)
		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.log.Debug("websocket accept failed", "err", err) // Accept already answered the request
		return
	}
	c.SetReadLimit(h.cfg.MaxMessageBytes)
	h.serve(r.Context(), rh, wsConn{c})
}

// session is the per-connection state of the read loop.
type session struct {
	client       *Client
	rh           *RoomHub
	msgLimit     *ratelimit.Limiter
	reactLimit   *ratelimit.Limiter
	authFailures int
}

func (h *Handler) serve(parent context.Context, rh *RoomHub, conn Conn) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	client := NewClient(h.cfg.IDs.NewSessionID(), conn, h.cfg.OutboxSize)
	defer client.Close()

	// Register before announcing, so the new client receives the state too.
	rh.Join(client)
	_ = rh.Apply(room.ViewerJoined{}, h.cfg.Clock.Now())
	defer func() {
		rh.Leave(client)
		_ = rh.Apply(room.ViewerLeft{}, h.cfg.Clock.Now())
	}()

	go h.writeLoop(ctx, cancel, client, conn)

	now := h.cfg.Clock.Now
	s := &session{
		client:     client,
		rh:         rh,
		msgLimit:   ratelimit.New(h.cfg.MsgRatePerSec, h.cfg.MsgBurst, now),
		reactLimit: ratelimit.New(h.cfg.ReactRatePerSec, h.cfg.ReactBurst, now),
	}
	for {
		readCtx, stop := context.WithTimeout(ctx, h.cfg.IdleTimeout)
		data, err := conn.ReadMessage(readCtx)
		stop()
		if err != nil {
			return
		}
		if !s.msgLimit.Allow("msg") {
			h.sendError(client, CodeRateLimited, "too many messages")
			continue
		}
		if !h.handle(s, data) {
			return
		}
	}
}

// writeLoop sends queued messages. Any write error ends the connection.
func (h *Handler) writeLoop(ctx context.Context, cancel context.CancelFunc, c *Client, conn Conn) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.Outbox():
			wctx, stop := context.WithTimeout(ctx, h.cfg.WriteTimeout)
			err := conn.WriteMessage(wctx, msg)
			stop()
			if err != nil {
				return
			}
		}
	}
}

// handle processes one message. It returns false when the connection must be closed.
func (h *Handler) handle(s *session, data []byte) bool {
	env, err := Decode(data)
	if err != nil {
		h.sendError(s.client, CodeBadRequest, "malformed message")
		return true
	}
	now := h.cfg.Clock.Now()

	switch env.Type {
	case TypePing:
		h.send(s.client, TypePong, PongData{ServerNowMS: now.UnixMilli()})

	case TypeJoin:
		var d JoinData
		if !h.decodeData(s.client, env, &d) {
			return true
		}
		if err := moderation.ValidateNickname(d.Nickname); err != nil {
			h.sendError(s.client, CodeInvalid, "invalid nickname")
			return true
		}
		s.client.Nickname = strings.TrimSpace(d.Nickname)

	case TypeHostAuth:
		var d HostAuthData
		if !h.decodeData(s.client, env, &d) {
			return true
		}
		if d.Token != "" && h.cfg.Hosts.Verify(d.Token, s.rh.State().HostTokenHash) {
			s.client.IsHost = true
			s.authFailures = 0
			h.send(s.client, TypeHostOK, nil)
			return true
		}
		s.authFailures++
		h.sendError(s.client, CodeForbidden, "wrong host token")
		return s.authFailures < h.cfg.MaxAuthFailures

	case TypeReaction:
		var d ReactionData
		if !h.decodeData(s.client, env, &d) {
			return true
		}
		if !s.reactLimit.Allow("react") {
			h.sendError(s.client, CodeRateLimited, "too many reactions")
			return true
		}
		if err := s.rh.React(room.ReactionKind(d.Kind), now); err != nil {
			h.sendError(s.client, CodeInvalid, "unknown reaction")
		}

	case TypeTimerStart, TypeTimerPause, TypeTimerResume, TypeTimerReset, TypeTimerRecover:
		if !s.client.IsHost {
			h.sendError(s.client, CodeForbidden, "only the host can control the timer")
			return true
		}
		h.timerCommand(s, env, now)

	default:
		h.sendError(s.client, CodeBadRequest, "unknown message type")
	}
	return true
}

// timerCommand converts a host message to a core event and applies it.
func (h *Handler) timerCommand(s *session, env Envelope, now time.Time) {
	var ev room.Event
	switch env.Type {
	case TypeTimerStart:
		var d TimerStartData
		if !h.decodeData(s.client, env, &d) {
			return
		}
		start, err := ToTimerStart(d)
		if err != nil {
			h.sendError(s.client, CodeInvalid, "invalid timer")
			return
		}
		ev = start
	case TypeTimerPause:
		ev = room.TimerPause{}
	case TypeTimerResume:
		ev = room.TimerResume{}
	case TypeTimerReset:
		ev = room.TimerReset{}
	case TypeTimerRecover:
		var d TimerRecoverData
		if !h.decodeData(s.client, env, &d) {
			return
		}
		t, err := FromTimerView(d.Timer)
		if err != nil {
			h.sendError(s.client, CodeInvalid, "invalid timer")
			return
		}
		ev = room.HostRecovered{Timer: t}
	}

	if err := s.rh.Apply(ev, now); err != nil {
		switch {
		case errors.Is(err, room.ErrInvalidTransition):
			h.sendError(s.client, CodeInvalidTransition, "not possible right now")
		case errors.Is(err, room.ErrInvalidEvent):
			h.sendError(s.client, CodeInvalid, "invalid value")
		default:
			h.log.Error("applying a room event failed", "type", env.Type, "err", err)
			h.sendError(s.client, CodeBadRequest, "request failed")
		}
	}
}

// decodeData parses env.Data into v and reports a bad_request on failure.
func (h *Handler) decodeData(c *Client, env Envelope, v any) bool {
	if len(env.Data) == 0 || json.Unmarshal(env.Data, v) != nil {
		h.sendError(c, CodeBadRequest, "malformed data")
		return false
	}
	return true
}

func (h *Handler) send(c *Client, typ string, data any) {
	b, err := Encode(typ, data)
	if err != nil {
		h.log.Error("encoding a message failed", "type", typ, "err", err)
		return
	}
	c.Send(b)
}

func (h *Handler) sendError(c *Client, code, msg string) {
	h.send(c, TypeError, ErrorData{Code: code, Message: msg})
}
