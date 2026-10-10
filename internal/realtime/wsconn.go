package realtime

import (
	"context"
	"errors"

	"github.com/coder/websocket"
)

// ErrUnsupportedFrame is returned when a client sends a non-text frame.
var ErrUnsupportedFrame = errors.New("realtime: only text frames are supported")

// wsConn adapts a github.com/coder/websocket connection to Conn.
//
// Note from the library docs: an error or an expired context on any call closes
// the connection, so read deadlines are expressed as context timeouts.
type wsConn struct {
	c *websocket.Conn
}

var _ Conn = wsConn{}

func (w wsConn) ReadMessage(ctx context.Context) ([]byte, error) {
	typ, data, err := w.c.Read(ctx)
	if err != nil {
		return nil, err
	}
	if typ != websocket.MessageText {
		return nil, ErrUnsupportedFrame
	}
	return data, nil
}

func (w wsConn) WriteMessage(ctx context.Context, payload []byte) error {
	return w.c.Write(ctx, websocket.MessageText, payload)
}

// Close closes the connection immediately, without the close handshake.
func (w wsConn) Close() error { return w.c.CloseNow() }
