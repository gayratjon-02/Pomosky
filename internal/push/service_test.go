package push

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

func newService(t *testing.T) *Service {
	t.Helper()
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)), pub, priv, "mailto:ops@example.com")
}

// browserKeys makes the keys a browser would send with a subscription.
func browserKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	k, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 16)
	_, _ = rand.Read(secret)
	return base64.RawURLEncoding.EncodeToString(k.PublicKey().Bytes()), base64.RawURLEncoding.EncodeToString(secret)
}

func TestOnlyKnownPushServicesAreAccepted(t *testing.T) {
	s := newService(t)
	p, a := browserKeys(t)
	for _, endpoint := range []string{
		"http://fcm.googleapis.com/fcm/send/abc",       // not https
		"https://127.0.0.1/steal",                      // internal address
		"https://evil.example.com/push",                // unknown host
		"https://user@fcm.googleapis.com/fcm/send/abc", // credentials in the URL
		"https://googleapis.com.evil.example/push",     // look-alike
		"",
	} {
		err := s.Subscribe("abcd2345", models.PushSubscription{Endpoint: endpoint, P256dh: p, Auth: a})
		if !errors.Is(err, apperr.ErrInvalid) {
			t.Errorf("%q: got %v, want ErrInvalid", endpoint, err)
		}
	}
	for _, endpoint := range []string{
		"https://fcm.googleapis.com/fcm/send/abc",
		"https://updates.push.services.mozilla.com/wpush/v2/abc",
		"https://web.push.apple.com/abc",
	} {
		if err := s.Subscribe("abcd2345", models.PushSubscription{Endpoint: endpoint, P256dh: p, Auth: a}); err != nil {
			t.Errorf("%q: %v", endpoint, err)
		}
	}
}

func TestSubscribeRefreshesWithoutDuplicates(t *testing.T) {
	s := newService(t)
	p, a := browserKeys(t)
	sub := models.PushSubscription{Endpoint: "https://fcm.googleapis.com/fcm/send/abc", P256dh: p, Auth: a}
	for i := 0; i < 3; i++ {
		if err := s.Subscribe("abcd2345", sub); err != nil {
			t.Fatal(err)
		}
	}
	if s.Count() != 1 {
		t.Fatalf("count %d, want 1", s.Count())
	}
	s.Unsubscribe(sub.Endpoint)
	if s.Count() != 0 {
		t.Fatalf("count %d after unsubscribe", s.Count())
	}
}

func TestPruneDropsClosedRooms(t *testing.T) {
	s := newService(t)
	p, a := browserKeys(t)
	_ = s.Subscribe("aaaa2222", models.PushSubscription{Endpoint: "https://fcm.googleapis.com/fcm/send/1", P256dh: p, Auth: a})
	_ = s.Subscribe("bbbb3333", models.PushSubscription{Endpoint: "https://fcm.googleapis.com/fcm/send/2", P256dh: p, Auth: a})
	s.Prune(func(id room.ID) bool { return id == "aaaa2222" })
	if s.Count() != 1 {
		t.Fatalf("count %d, want 1", s.Count())
	}
}

func TestPhaseEndSendsAnEncryptedMessageAndForgetsGoneBrowsers(t *testing.T) {
	var posts atomic.Int32
	status := atomic.Int32{}
	status.Store(http.StatusCreated)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Authorization"), "vapid ") ||
			r.Header.Get("Content-Encoding") != "aes128gcm" {
			t.Errorf("unexpected push request: %s %v", r.Method, r.Header)
		}
		w.WriteHeader(int(status.Load()))
	}))
	defer srv.Close()

	s := newService(t)
	s.allowHost = func(string) bool { return true }
	s.client = srv.Client()
	p, a := browserKeys(t)
	// The test server speaks plain http, so store the subscription directly.
	s.rooms["abcd2345"] = map[string]models.PushSubscription{srv.URL: {Endpoint: srv.URL, P256dh: p, Auth: a}}
	s.total = 1

	s.PhaseEnded("abcd2345", room.PhaseFocus, "essay")
	waitFor(t, func() bool { return posts.Load() == 1 })
	if s.Count() != 1 {
		t.Fatal("a delivered subscription must stay")
	}

	status.Store(http.StatusGone)
	s.PhaseEnded("abcd2345", room.PhaseBreak, "")
	waitFor(t, func() bool { return s.Count() == 0 })
}

func TestEndMessage(t *testing.T) {
	if title, body := EndMessage(room.PhaseFocus, " essay "); title != "Focus session finished" || body != "essay. Time for a break." {
		t.Errorf("focus: %q / %q", title, body)
	}
	if title, _ := EndMessage(room.PhaseBreak, "x"); title != "Break is over" {
		t.Errorf("break: %q", title)
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out")
}
