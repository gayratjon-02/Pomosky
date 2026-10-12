package controllers

import (
	"context"
	"log/slog"

	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/views"
	"github.com/gayratjon-02/Pomosky/internal/weather"
)

// WeatherSource is what controllers need from the weather layer.
// *weather.Service implements it.
type WeatherSource interface {
	Current(ctx context.Context, lat, lon float64) (weather.Snapshot, error)
}

// HostHasher turns a raw host token into the value stored for a room.
// auth.HostTokens implements it.
type HostHasher interface {
	Hash(token string) string
}

// Deps are the collaborators of the controllers, wired in internal/app.
// Fields may be nil until the corresponding stage is implemented.
type Deps struct {
	Config     config.Config
	Log        *slog.Logger
	Views      *views.Renderer
	Rooms      ports.RoomRepository
	Accounts   ports.Accounts    // nil: login is off, everything else still works
	Stats      ports.StatsReader // nil without a database
	Admin      ports.AdminReader // nil without a database
	Live       ports.LiveRooms
	HostTokens HostHasher
	Clock      ports.Clock
	IDs        ports.IDGenerator
	Weather    WeatherSource
	Places     PlaceSearcher           // nil: no place search
	Visits     ports.VisitTracker      // nil: usage is not counted
	Traffic    ports.TrafficReader     // nil without a database
	Push       ports.PushSubscriptions // nil: Web Push is off
}

// Controllers groups all handlers; methods are split over files per area.
type Controllers struct {
	deps Deps
}

// New creates the controllers.
func New(d Deps) *Controllers { return &Controllers{deps: d} }
