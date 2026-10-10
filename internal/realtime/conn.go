package realtime

import "context"

// Conn is the minimal WebSocket connection the hub needs. A transport adapter
// (for example around github.com/coder/websocket) implements it.
type Conn interface {
	ReadMessage(ctx context.Context) ([]byte, error)
	WriteMessage(ctx context.Context, payload []byte) error
	Close() error
}
