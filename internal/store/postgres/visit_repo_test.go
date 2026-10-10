package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

func TestVisitsTraffic(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if _, err := db.SQL().Exec("TRUNCATE visits"); err != nil {
		t.Fatal(err)
	}
	repo := NewVisitRepo(db)
	day := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	add := func(at time.Time, kind, path, visitor, source string) {
		t.Helper()
		v := models.Visit{At: at, Kind: kind, Path: path, Visitor: visitor, Source: source, Medium: "social", Device: "mobile", OS: "iOS", Browser: "Instagram", Country: "UZ"}
		if err := repo.Create(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	// Visitor a: Instagram, opens the landing, creates a room, logs in.
	add(day, models.VisitPageview, "/", "a", "instagram")
	add(day.Add(time.Minute), models.VisitRoomCreated, "/r/new", "a", "")
	add(day.Add(2*time.Minute), models.VisitPageview, "/r/:id/host", "a", "")
	add(day.Add(3*time.Minute), models.VisitLoginDone, "/auth/google/callback", "a", "")
	// Visitor b: TikTok, only looks. Visitor c: Instagram, only looks.
	add(day.Add(time.Hour), models.VisitPageview, "/", "b", "tiktok")
	add(day.Add(2*time.Hour), models.VisitPageview, "/", "c", "instagram")

	got, err := repo.Traffic(ctx, day.Truncate(24*time.Hour), day.Truncate(24*time.Hour).AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Days) != 1 || got.Days[0].Pageviews != 4 || got.Days[0].Visitors != 3 || got.Days[0].RoomsCreated != 1 || got.Days[0].Logins != 1 {
		t.Fatalf("days %+v", got.Days)
	}
	if got.Funnel.Visitors != 3 || got.Funnel.RoomViewers != 1 || got.Funnel.RoomsCreated != 1 || got.Funnel.LoginDone != 1 {
		t.Fatalf("funnel %+v", got.Funnel)
	}
	var insta *models.TrafficSource
	for i := range got.Sources {
		if got.Sources[i].Source == "instagram" {
			insta = &got.Sources[i]
		}
	}
	if insta == nil || insta.Visitors != 2 || insta.RoomsCreated != 1 || insta.Logins != 1 {
		t.Fatalf("sources %+v", got.Sources)
	}
	if len(got.Pages) == 0 || got.Pages[0].Path != "/" || got.Pages[0].Pageviews != 3 {
		t.Fatalf("pages %+v", got.Pages)
	}
	if len(got.Browsers) != 1 || got.Browsers[0].Label != "Instagram" || got.Browsers[0].Visitors != 3 || got.Countries[0].Label != "UZ" {
		t.Fatalf("browsers %+v countries %+v", got.Browsers, got.Countries)
	}
}
