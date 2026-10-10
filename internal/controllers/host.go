package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/views"
)

// HostPanel serves GET /r/{id}/host: the host control panel.
//
// The page itself is not secret; control is granted over the WebSocket by the
// host token the browser keeps (see realtime host.auth). Anyone without the
// token sees the panel but every timer command is rejected.
func (c *Controllers) HostPanel(w http.ResponseWriter, r *http.Request) {
	id, ok := c.liveRoom(w, r)
	if !ok {
		return
	}
	c.render(w, r, http.StatusOK, "host", views.PageData{
		Title:   "Host panel",
		Scripts: []string{"/static/host.js"},
		Data:    views.RoomPage{ID: id.String(), ShareURL: c.shareURL(id)},
	})
}
