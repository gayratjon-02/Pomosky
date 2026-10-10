package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// HostPanel serves GET /r/{id}/host: the host control panel.
//
// TODO(stage 5): verify the host token (auth.VerifyHostToken) or room
// ownership, then render "host".
func (c *Controllers) HostPanel(w http.ResponseWriter, r *http.Request) {
	if !room.ID(r.PathValue("id")).Valid() {
		c.NotFound(w, r)
		return
	}
	httpx.NotImplemented(w)
}
