package controllers

import (
	"hash/fnv"
	"math/rand"
	"net/http"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

// DemoStats serves GET /api/stats/demo: made-up example statistics in the real
// shape. The statistics page shows them (blurred, labelled as an example) to
// visitors who are not logged in, so they can see what logging in gives them.
func (c *Controllers) DemoStats(w http.ResponseWriter, r *http.Request) {
	month, ok := parseMonth(r, time.Now())
	if !ok {
		http.Error(w, "bad month", http.StatusBadRequest)
		return
	}
	writeStats(w, fixtureMonth(month, time.Now()))
}

func fixtureMonth(month, now time.Time) models.MonthStats {
	cur := fixtureOne(month, now)
	cur.PrevTotalSeconds = fixtureOne(month.AddDate(0, -1, 0), now).TotalSeconds
	return cur
}

func fixtureOne(month, now time.Time) models.MonthStats {
	h := fnv.New64a()
	_, _ = h.Write([]byte(month.Format("2006-01")))
	rng := rand.New(rand.NewSource(int64(h.Sum64())))

	out := models.MonthStats{Month: month.Format("2006-01")}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	streak := 0
	for d := month; d.Month() == month.Month(); d = d.AddDate(0, 0, 1) {
		day := models.DayStat{Date: d.Format("2006-01-02")}
		if !d.After(today) && rng.Float64() < 0.72 {
			day.Sessions = 1 + rng.Intn(5)
			day.Seconds = int64(day.Sessions) * int64(20+rng.Intn(40)) * 60
		}
		if day.Seconds > 0 {
			out.ActiveDays++
			out.Sessions += day.Sessions
			out.TotalSeconds += day.Seconds
			streak++
		} else if !d.After(today) {
			streak = 0
		}
		out.Days = append(out.Days, day)
	}
	out.StreakDays = streak
	return out
}
