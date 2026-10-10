package controllers

import "net/http"

// track counts a usage event (a room created, a login step). It is a no-op
// when analytics is off and never affects the response.
func (c *Controllers) track(r *http.Request, kind string) {
	if c.deps.Visits != nil {
		c.deps.Visits.Track(r, kind)
	}
}
