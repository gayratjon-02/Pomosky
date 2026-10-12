package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/i18n"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// maxPushBody bounds the JSON a browser sends when it subscribes.
const maxPushBody = 4 << 10

// PushKey serves GET /api/push/key: the public key a browser subscribes with.
// It answers 404 when Web Push is not set up.
func (c *Controllers) PushKey(w http.ResponseWriter, r *http.Request) {
	if c.deps.Push == nil {
		httpx.WriteError(w, apperr.ErrNotFound)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"publicKey": c.deps.Push.PublicKey()})
}

type pushSubscribeBody struct {
	Room         string `json:"room"`
	Lang         string `json:"lang"`
	Subscription struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	} `json:"subscription"`
}

// PushSubscribe serves POST /api/push/subscribe: ask for an alert when a phase
// of a live room ends. Browsers call it on every visit, which also refreshes it.
func (c *Controllers) PushSubscribe(w http.ResponseWriter, r *http.Request) {
	if c.deps.Push == nil || c.deps.Live == nil {
		httpx.WriteError(w, apperr.ErrNotFound)
		return
	}
	var body pushSubscribeBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPushBody)).Decode(&body); err != nil {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	id := room.ID(body.Room)
	if !id.Valid() {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	if !c.deps.Live.Has(id) {
		httpx.WriteError(w, apperr.ErrNotFound)
		return
	}
	err := c.deps.Push.Subscribe(id, models.PushSubscription{
		Endpoint: body.Subscription.Endpoint,
		P256dh:   body.Subscription.Keys.P256dh,
		Auth:     body.Subscription.Keys.Auth,
		Lang:     string(i18n.Parse(body.Lang)),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PushUnsubscribe serves POST /api/push/unsubscribe with {"endpoint": "..."}.
func (c *Controllers) PushUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if c.deps.Push == nil {
		httpx.WriteError(w, apperr.ErrNotFound)
		return
	}
	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPushBody)).Decode(&body); err != nil || body.Endpoint == "" {
		httpx.WriteError(w, apperr.ErrInvalid)
		return
	}
	c.deps.Push.Unsubscribe(body.Endpoint)
	w.WriteHeader(http.StatusNoContent)
}
