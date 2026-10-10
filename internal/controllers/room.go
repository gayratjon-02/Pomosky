package controllers

import (
	"net/http"
	"strings"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
	"github.com/gayratjon-02/Pomosky/internal/room"
	"github.com/gayratjon-02/Pomosky/internal/views"
)

// shareURL is the public link of a room.
func (c *Controllers) shareURL(id room.ID) string {
	return strings.TrimRight(c.deps.Config.BaseURL, "/") + "/r/" + id.String()
}

// liveRoom returns the ID from the path if that room is live, and answers 404 otherwise.
func (c *Controllers) liveRoom(w http.ResponseWriter, r *http.Request) (room.ID, bool) {
	id := room.ID(r.PathValue("id"))
	if !id.Valid() || c.deps.Live == nil || !c.deps.Live.Has(id) {
		c.render(w, http.StatusNotFound, "error", views.PageData{
			Title: "This room is not live",
			Data:  "The host may have ended the session. Ask for a new link, or start your own room.",
		})
		return "", false
	}
	return id, true
}

// RoomShow serves GET /r/{id}: the viewer page of a room, or the camera-friendly
// record mode with ?mode=record.
func (c *Controllers) RoomShow(w http.ResponseWriter, r *http.Request) {
	id, ok := c.liveRoom(w, r)
	if !ok {
		return
	}
	page, script := "room", "/static/room.js"
	if r.URL.Query().Get("mode") == "record" {
		page, script = "record", "/static/record.js"
	}
	link := c.shareURL(id)
	c.render(w, http.StatusOK, page, views.PageData{
		Title:        "Work together",
		Description:  "Join this pomodoro room and work together under a live sky.",
		CanonicalURL: link,
		Scripts:      []string{script},
		Data:         views.RoomPage{ID: id.String(), ShareURL: link},
	})
}

// RoomCreate serves POST /r/new: it creates a live room and sends the host to
// the host panel. The host token travels in the URL fragment, which browsers
// never send to the server, so it does not end up in any log.
func (c *Controllers) RoomCreate(w http.ResponseWriter, r *http.Request) {
	if c.deps.Live == nil || c.deps.HostTokens == nil || c.deps.IDs == nil || c.deps.Clock == nil {
		httpx.NotImplemented(w)
		return
	}
	var id room.ID
	for i := 0; i < 5; i++ {
		if candidate := c.deps.IDs.NewRoomID(); !c.deps.Live.Has(candidate) {
			id = candidate
			break
		}
	}
	if id == "" {
		http.Error(w, "could not allocate a room, please try again", http.StatusInternalServerError)
		return
	}
	token := c.deps.IDs.NewToken()
	c.deps.Live.Register(room.New(id, c.deps.HostTokens.Hash(token), c.deps.Clock.Now()))
	http.Redirect(w, r, "/r/"+id.String()+"/host#token="+token, http.StatusSeeOther)
}
