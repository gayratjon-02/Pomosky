package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
)

type trafficDayJSON struct {
	Date         string `json:"date"`
	Pageviews    int    `json:"pageviews"`
	Visitors     int    `json:"visitors"`
	RoomsCreated int    `json:"roomsCreated"`
	Logins       int    `json:"logins"`
}

type trafficCountJSON struct {
	Label    string `json:"label"`
	Visitors int    `json:"visitors"`
}

type trafficJSON struct {
	Days  []trafficDayJSON `json:"days"`
	Today trafficDayJSON   `json:"today"`
	Pages []struct {
		Path      string `json:"path"`
		Pageviews int    `json:"pageviews"`
		Visitors  int    `json:"visitors"`
	} `json:"pages"`
	Sources []struct {
		Source       string `json:"source"`
		Medium       string `json:"medium"`
		Visitors     int    `json:"visitors"`
		RoomsCreated int    `json:"roomsCreated"`
		Logins       int    `json:"logins"`
	} `json:"sources"`
	Devices   []trafficCountJSON `json:"devices"`
	Systems   []trafficCountJSON `json:"systems"`
	Browsers  []trafficCountJSON `json:"browsers"`
	Countries []trafficCountJSON `json:"countries"`
	Funnel    struct {
		Visitors     int `json:"visitors"`
		RoomViewers  int `json:"roomViewers"`
		RoomsCreated int `json:"roomsCreated"`
		LoginStarted int `json:"loginStarted"`
		LoginDone    int `json:"loginDone"`
	} `json:"funnel"`
}

func toTrafficDayJSON(d models.TrafficDay) trafficDayJSON {
	return trafficDayJSON{Date: d.Date, Pageviews: d.Pageviews, Visitors: d.Visitors, RoomsCreated: d.RoomsCreated, Logins: d.Logins}
}

func toTrafficCounts(in []models.TrafficCount) []trafficCountJSON {
	out := make([]trafficCountJSON, 0, len(in))
	for _, c := range in {
		label := c.Label
		if label == "" {
			label = "unknown"
		}
		out = append(out, trafficCountJSON{Label: label, Visitors: c.Visitors})
	}
	return out
}

func toTrafficJSON(o models.TrafficOverview) trafficJSON {
	var j trafficJSON
	j.Days = make([]trafficDayJSON, 0, len(o.Days))
	for _, d := range o.Days {
		j.Days = append(j.Days, toTrafficDayJSON(d))
	}
	j.Today = toTrafficDayJSON(o.Today)
	for _, p := range o.Pages {
		j.Pages = append(j.Pages, struct {
			Path      string `json:"path"`
			Pageviews int    `json:"pageviews"`
			Visitors  int    `json:"visitors"`
		}{p.Path, p.Pageviews, p.Visitors})
	}
	for _, s := range o.Sources {
		j.Sources = append(j.Sources, struct {
			Source       string `json:"source"`
			Medium       string `json:"medium"`
			Visitors     int    `json:"visitors"`
			RoomsCreated int    `json:"roomsCreated"`
			Logins       int    `json:"logins"`
		}{s.Source, s.Medium, s.Visitors, s.RoomsCreated, s.Logins})
	}
	j.Devices, j.Systems, j.Browsers, j.Countries = toTrafficCounts(o.Devices), toTrafficCounts(o.Systems), toTrafficCounts(o.Browsers), toTrafficCounts(o.Countries)
	f := o.Funnel
	j.Funnel.Visitors, j.Funnel.RoomViewers, j.Funnel.RoomsCreated, j.Funnel.LoginStarted, j.Funnel.LoginDone = f.Visitors, f.RoomViewers, f.RoomsCreated, f.LoginStarted, f.LoginDone
	return j
}

// AdminTrafficAPI serves GET /api/admin/traffic: anonymous visitor numbers of the last 30 UTC days.
func (c *Controllers) AdminTrafficAPI(w http.ResponseWriter, r *http.Request) {
	if !c.adminGate(w, r, true) {
		return
	}
	if c.deps.Traffic == nil {
		httpx.WriteError(w, apperr.ErrNotImplemented)
		return
	}
	o, err := c.deps.Traffic.Traffic(r.Context())
	if err != nil {
		c.deps.Log.Error("admin traffic", "err", err)
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toTrafficJSON(o))
}
