package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// VisitRepo implements ports.VisitRepository.
type VisitRepo struct{ db *DB }

var _ ports.VisitRepository = (*VisitRepo)(nil)

// NewVisitRepo creates the repository.
func NewVisitRepo(db *DB) *VisitRepo { return &VisitRepo{db: db} }

func (r *VisitRepo) Create(ctx context.Context, v models.Visit) error {
	return mapErr(r.db.q.CreateVisit(ctx, sqlcgen.CreateVisitParams{
		At: v.At, Kind: v.Kind, Path: v.Path, Visitor: v.Visitor, Source: v.Source, Medium: v.Medium,
		Campaign: v.Campaign, Device: v.Device, Os: v.OS, Browser: v.Browser, Country: v.Country,
	}))
}

func (r *VisitRepo) Traffic(ctx context.Context, from, to time.Time) (models.TrafficOverview, error) {
	var out models.TrafficOverview
	days, err := r.db.q.TrafficDaily(ctx, sqlcgen.TrafficDailyParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, d := range days {
		out.Days = append(out.Days, models.TrafficDay{
			Date: d.Day.Format("2006-01-02"), Pageviews: int(d.Pageviews), Visitors: int(d.Visitors),
			RoomsCreated: int(d.RoomsCreated), Logins: int(d.Logins),
		})
	}
	pages, err := r.db.q.TrafficPages(ctx, sqlcgen.TrafficPagesParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, p := range pages {
		out.Pages = append(out.Pages, models.TrafficPage{Path: p.Path, Pageviews: int(p.Pageviews), Visitors: int(p.Visitors)})
	}
	sources, err := r.db.q.TrafficSources(ctx, sqlcgen.TrafficSourcesParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, s := range sources {
		out.Sources = append(out.Sources, models.TrafficSource{
			Source: s.Source, Medium: s.Medium, Visitors: int(s.Visitors), RoomsCreated: int(s.RoomsCreated), Logins: int(s.Logins),
		})
	}
	devices, err := r.db.q.TrafficDevices(ctx, sqlcgen.TrafficDevicesParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, c := range devices {
		out.Devices = append(out.Devices, models.TrafficCount{Label: c.Label, Visitors: int(c.Visitors)})
	}
	systems, err := r.db.q.TrafficSystems(ctx, sqlcgen.TrafficSystemsParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, c := range systems {
		out.Systems = append(out.Systems, models.TrafficCount{Label: c.Label, Visitors: int(c.Visitors)})
	}
	browsers, err := r.db.q.TrafficBrowsers(ctx, sqlcgen.TrafficBrowsersParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, c := range browsers {
		out.Browsers = append(out.Browsers, models.TrafficCount{Label: c.Label, Visitors: int(c.Visitors)})
	}
	countries, err := r.db.q.TrafficCountries(ctx, sqlcgen.TrafficCountriesParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	for _, c := range countries {
		out.Countries = append(out.Countries, models.TrafficCount{Label: c.Label, Visitors: int(c.Visitors)})
	}
	f, err := r.db.q.TrafficFunnel(ctx, sqlcgen.TrafficFunnelParams{FromDay: from, ToDay: to})
	if err != nil {
		return out, mapErr(err)
	}
	out.Funnel = models.TrafficFunnel{
		Visitors: int(f.Visitors), RoomViewers: int(f.RoomViewers), RoomsCreated: int(f.RoomsCreated),
		LoginStarted: int(f.LoginStarted), LoginDone: int(f.LoginDone),
	}
	return out, nil
}
