package models

import "time"

// Visit kinds.
const (
	VisitPageview     = "pageview"
	VisitRoomCreated  = "room_created"
	VisitLoginStarted = "login_started"
	VisitLoginDone    = "login_done"
)

// Visit is one anonymous usage event. Visitor is a hash that changes every UTC
// day; there is no IP address, cookie or room id in it.
type Visit struct {
	At       time.Time
	Kind     string
	Path     string
	Visitor  string
	Source   string
	Medium   string
	Campaign string
	Device   string
	OS       string
	Browser  string
	Country  string
}

// TrafficDay is one UTC day of the traffic charts. Visitors are distinct daily hashes.
type TrafficDay struct {
	Date         string
	Pageviews    int
	Visitors     int
	RoomsCreated int
	Logins       int
}

// TrafficPage is a row of the top pages table.
type TrafficPage struct {
	Path      string
	Pageviews int
	Visitors  int
}

// TrafficSource is where visitors came from, by the first page of their day.
type TrafficSource struct {
	Source       string
	Medium       string
	Visitors     int
	RoomsCreated int // of those visitors, how many created a room the same day
	Logins       int
}

// TrafficCount is a labelled count (device, operating system, browser, country).
type TrafficCount struct {
	Label    string
	Visitors int
}

// TrafficFunnel counts distinct visitor-days over the window.
type TrafficFunnel struct {
	Visitors     int
	RoomViewers  int // opened a room page
	RoomsCreated int
	LoginStarted int
	LoginDone    int
}

// TrafficOverview is what the admin traffic section shows (last 30 UTC days).
type TrafficOverview struct {
	Days      []TrafficDay
	Pages     []TrafficPage
	Sources   []TrafficSource
	Devices   []TrafficCount
	Systems   []TrafficCount
	Browsers  []TrafficCount
	Countries []TrafficCount
	Funnel    TrafficFunnel
	Today     TrafficDay
}
