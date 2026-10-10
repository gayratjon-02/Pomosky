package realtime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// echoServer accepts one WebSocket connection, wraps it in wsConn and reports
// every message (or the read error) on the returned channel.
func echoServer(t *testing.T) (url string, got chan error, msgs chan string) {
	t.Helper()
	got = make(chan error, 1)
	msgs = make(chan string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		conn := wsConn{c}
		defer conn.Close()
		for {
			b, err := conn.ReadMessage(r.Context())
			if err != nil {
				got <- err
				return
			}
			msgs <- string(b)
			_ = conn.WriteMessage(r.Context(), append([]byte("echo:"), b...))
		}
	}))
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http"), got, msgs
}

func TestWsConnTextRoundTrip(t *testing.T) {
	url, _, msgs := echoServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	if err := c.Write(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if m := <-msgs; m != "hello" {
		t.Fatalf("server got %q", m)
	}
	_, reply, err := c.Read(ctx)
	if err != nil || string(reply) != "echo:hello" {
		t.Fatalf("reply = %q, err = %v", reply, err)
	}
}

func TestWsConnRejectsBinaryFrames(t *testing.T) {
	url, got, _ := echoServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	if err := c.Write(ctx, websocket.MessageBinary, []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-got:
		if !errors.Is(err, ErrUnsupportedFrame) {
			t.Fatalf("error = %v, want ErrUnsupportedFrame", err)
		}
	case <-ctx.Done():
		t.Fatal("server did not report the binary frame")
	}
}
