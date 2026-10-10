package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// RoomShow serves GET /r/{id}: the viewer page of a room.
//
// TODO(stage 5): load the room, render "room" (or "record" for ?mode=record)
// with Open Graph tags and the room/sky bundles.
func (c *Controllers) RoomShow(w http.ResponseWriter, r *http.Request) {
	if !room.ID(r.PathValue("id")).Valid() {
		c.NotFound(w, r)
		return
	}
	httpx.NotImplemented(w)
}

// RoomCreate serves POST /r/new: create a room and show the host token once.
//
// TODO(stage 3/4): generate ID and host token, store the hash, redirect to the
// host panel.
func (c *Controllers) RoomCreate(w http.ResponseWriter, r *http.Request) {
	httpx.NotImplemented(w)
}
