package analytics

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// TrafficWindowDays is how many UTC days the traffic section covers.
const TrafficWindowDays = 30

// Reader implements ports.TrafficReader.
type Reader struct {
	repo  ports.VisitRepository
	clock ports.Clock
}

var _ ports.TrafficReader = (*Reader)(nil)

// NewReader creates the reader.
func NewReader(repo ports.VisitRepository, clock ports.Clock) *Reader {
	return &Reader{repo: repo, clock: clock}
}

// Traffic implements ports.TrafficReader: the last 30 UTC days, every day present.
func (r *Reader) Traffic(ctx context.Context) (models.TrafficOverview, error) {
	now := r.clock.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, 0, -(TrafficWindowDays - 1))
	to := today.AddDate(0, 0, 1)

	out, err := r.repo.Traffic(ctx, from, to)
	if err != nil {
		return out, err
	}
	byDate := make(map[string]models.TrafficDay, len(out.Days))
	for _, d := range out.Days {
		byDate[d.Date] = d
	}
	out.Days = out.Days[:0]
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		day, ok := byDate[date]
		if !ok {
			day = models.TrafficDay{Date: date}
		}
		out.Days = append(out.Days, day)
	}
	out.Today = out.Days[len(out.Days)-1]
	return out, nil
}
