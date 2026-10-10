package realtime

// Client is one connected browser tab.
type Client struct {
	ID       string
	Nickname string
	IsHost   bool

	conn Conn
	send chan []byte
}

// NewClient creates a client with an outbox of the given size.
func NewClient(id string, conn Conn, outbox int) *Client {
	return &Client{ID: id, conn: conn, send: make(chan []byte, outbox)}
}

// Send queues a message without blocking. It returns false if the outbox is
// full, in which case the message is dropped (a slow client must never stall
// the room).
func (c *Client) Send(payload []byte) bool {
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

// Outbox is the channel the write loop drains.
func (c *Client) Outbox() <-chan []byte { return c.send }

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
