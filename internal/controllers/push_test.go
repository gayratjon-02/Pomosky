package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

type fakePush struct {
	subs    map[room.ID][]models.PushSubscription
	removed string
	err     error
}

func (f *fakePush) PublicKey() string { return "PUBKEY" }
func (f *fakePush) Subscribe(id room.ID, s models.PushSubscription) error {
	if f.err != nil {
		return f.err
	}
	if f.subs == nil {
		f.subs = map[room.ID][]models.PushSubscription{}
	}
	f.subs[id] = append(f.subs[id], s)
	return nil
}
func (f *fakePush) Unsubscribe(endpoint string) { f.removed = endpoint }

func pushSetup(t *testing.T, p *fakePush) *Controllers {
	t.Helper()
	live := &fakeLive{}
	live.Register(room.New("abcd2345", "h", time.Now()), "")
	c := newTestControllers(t)
	c.deps.Live = live
	if p != nil {
		c.deps.Push = p
	}
	return c
}

func postPush(h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("POST", "/api/push/x", strings.NewReader(body)))
	return rec
}

func subscribeBody(room string) string {
	return fmt.Sprintf(`{"room":%q,"subscription":{"endpoint":"https://fcm.googleapis.com/fcm/send/abc","keys":{"p256dh":"P","auth":"A"}}}`, room)
}

func TestPushIsOffWithoutKeys(t *testing.T) {
	c := pushSetup(t, nil)
	rec := httptest.NewRecorder()
	c.PushKey(rec, httptest.NewRequest("GET", "/api/push/key", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("key status %d", rec.Code)
	}
	if rec := postPush(c.PushSubscribe, subscribeBody("abcd2345")); rec.Code != http.StatusNotFound {
		t.Fatalf("subscribe status %d", rec.Code)
	}
}

func TestPushKeyIsServed(t *testing.T) {
	c := pushSetup(t, &fakePush{})
	rec := httptest.NewRecorder()
	c.PushKey(rec, httptest.NewRequest("GET", "/api/push/key", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"publicKey":"PUBKEY"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestPushSubscribeStoresTheSubscriptionOfALiveRoom(t *testing.T) {
	p := &fakePush{}
	c := pushSetup(t, p)
	if rec := postPush(c.PushSubscribe, subscribeBody("abcd2345")); rec.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	got := p.subs["abcd2345"]
	if len(got) != 1 || got[0].Endpoint != "https://fcm.googleapis.com/fcm/send/abc" || got[0].P256dh != "P" || got[0].Auth != "A" {
		t.Fatalf("stored %+v", got)
	}
}

func TestPushSubscribeRejectsBadInput(t *testing.T) {
	p := &fakePush{}
	c := pushSetup(t, p)
	for name, tc := range map[string]struct {
		body string
		want int
	}{
		"not json":     {"nope", http.StatusBadRequest},
		"bad room id":  {subscribeBody("x"), http.StatusBadRequest},
		"unknown room": {subscribeBody("bcde2345"), http.StatusNotFound},
	} {
		if rec := postPush(c.PushSubscribe, tc.body); rec.Code != tc.want {
			t.Errorf("%s: status %d, want %d", name, rec.Code, tc.want)
		}
	}
	p.err = apperr.ErrInvalid
	if rec := postPush(c.PushSubscribe, subscribeBody("abcd2345")); rec.Code != http.StatusBadRequest {
		t.Errorf("service refusal: status %d", rec.Code)
	}
	big := `{"room":"abcd2345","pad":"` + strings.Repeat("a", 5000) + `"}`
	if rec := postPush(c.PushSubscribe, big); rec.Code != http.StatusBadRequest {
		t.Errorf("oversized body: status %d", rec.Code)
	}
}

func TestPushUnsubscribe(t *testing.T) {
	p := &fakePush{}
	c := pushSetup(t, p)
	if rec := postPush(c.PushUnsubscribe, `{"endpoint":"https://fcm.googleapis.com/fcm/send/abc"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
	if p.removed != "https://fcm.googleapis.com/fcm/send/abc" {
		t.Fatalf("removed %q", p.removed)
	}
	if rec := postPush(c.PushUnsubscribe, `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty endpoint: status %d", rec.Code)
	}
}
