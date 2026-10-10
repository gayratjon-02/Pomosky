package realtime

import (
	"log/slog"
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
)

// Handler serves GET /ws/{id}: it upgrades the request to a WebSocket,
// wraps the connection into a Conn and runs the read and write loops.
type Handler struct {
	hub *Hub
	log *slog.Logger
}

// NewHandler creates the WebSocket handler.
func NewHandler(hub *Hub, log *slog.Logger) *Handler {
	return &Handler{hub: hub, log: log}
}

// ServeHTTP implements http.Handler.
//
// TODO(stage 3): add the WebSocket library, upgrade the connection, enforce an
// origin check, send the initial state, run the read/write loops with
// ping/pong every ~25 seconds and per-connection rate limits.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.NotImplemented(w)
}
