package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// testDB opens the database named by TEST_DATABASE_URL (skipped when unset),
// applies the migrations and empties the tables.
func testDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().Exec("TRUNCATE users, login_tokens CASCADE"); err != nil {
		t.Fatal(err)
	}
	return db
}

func newUser(t *testing.T, db *DB, id, email string) models.User {
	t.Helper()
	u := models.User{ID: models.UserID(id), Email: &email, Provider: "magiclink", ProviderSubject: email, Nickname: id}
	if err := NewUserRepo(db).Create(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	return u
}

func TestUsers(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewUserRepo(db)
	u := newUser(t, db, "u1", "a@example.com")

	got, err := repo.ByID(ctx, u.ID)
	if err != nil || got.Email == nil || *got.Email != "a@example.com" {
		t.Fatalf("ByID: %+v %v", got, err)
	}
	if _, err := repo.ByEmail(ctx, "a@example.com"); err != nil {
		t.Fatalf("ByEmail: %v", err)
	}
	if _, err := repo.ByIdentity(ctx, "magiclink", "a@example.com"); err != nil {
		t.Fatalf("ByIdentity: %v", err)
	}
	if _, err := repo.ByID(ctx, "nobody"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("missing user: %v", err)
	}
	dup := models.User{ID: "u2", Email: u.Email, Provider: "google", ProviderSubject: "g1"}
	if err := repo.Create(ctx, dup); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("duplicate email: %v", err)
	}
}

func TestSessions(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	u := newUser(t, db, "u1", "a@example.com")
	repo := NewSessionRepo(db)
	now := time.Now().UTC()

	if err := repo.Create(ctx, models.Session{ID: "s1", UserID: u.ID, ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, models.Session{ID: "s2", UserID: u.ID, ExpiresAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if s, err := repo.ByID(ctx, "s1"); err != nil || s.UserID != u.ID {
		t.Fatalf("ByID: %+v %v", s, err)
	}
	if n, err := repo.DeleteExpired(ctx, now); err != nil || n != 1 {
		t.Fatalf("DeleteExpired: %d %v", n, err)
	}
	if err := repo.Delete(ctx, "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ByID(ctx, "s1"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("deleted session: %v", err)
	}
}

func TestLoginTokensWorkOnce(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewLoginTokenRepo(db)
	now := time.Now().UTC()

	if err := repo.Create(ctx, "h1", "a@example.com", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if email, err := repo.Consume(ctx, "h1", now); err != nil || email != "a@example.com" {
		t.Fatalf("first use: %q %v", email, err)
	}
	if _, err := repo.Consume(ctx, "h1", now); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("second use: %v", err)
	}
	if err := repo.Create(ctx, "h2", "a@example.com", now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Consume(ctx, "h2", now); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expired token: %v", err)
	}
}

func TestRoomsAndFocusTotals(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	u := newUser(t, db, "u1", "a@example.com")
	other := newUser(t, db, "u2", "b@example.com")
	rooms := NewRoomRepo(db)
	focus := NewFocusSessionRepo(db)

	owner := u.ID
	if err := rooms.Create(ctx, models.Room{ID: "abcdefgh", OwnerUserID: &owner, HostTokenHash: "x", Location: room.Location{TZ: "Asia/Seoul"}}); err != nil {
		t.Fatal(err)
	}
	otherOwner := other.ID
	if err := rooms.Create(ctx, models.Room{ID: "zzzzzzzz", OwnerUserID: &otherOwner, HostTokenHash: "y"}); err != nil {
		t.Fatal(err)
	}
	got, err := rooms.ByID(ctx, "abcdefgh")
	if err != nil || got.OwnerUserID == nil || *got.OwnerUserID != u.ID || got.Location.TZ != "Asia/Seoul" {
		t.Fatalf("room ByID: %+v %v", got, err)
	}

	// 2026-10-01 23:30 UTC is already 2 October in Seoul (UTC+9).
	utc := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, time.UTC) }
	add := func(id room.ID, start time.Time, minutes int) {
		t.Helper()
		err := focus.Create(ctx, models.FocusSession{RoomID: id, StartedAt: start, EndedAt: start.Add(time.Duration(minutes) * time.Minute), Focused: time.Duration(minutes) * time.Minute})
		if err != nil {
			t.Fatal(err)
		}
	}
	add("abcdefgh", utc(1, 23, 30), 25)
	add("abcdefgh", utc(2, 1, 0), 50)
	add("abcdefgh", utc(3, 5, 0), 30)
	add("zzzzzzzz", utc(2, 1, 0), 99) // someone else's session must not leak in

	totals, err := focus.DailyTotals(ctx, u.ID, utc(1, 0, 0), utc(31, 0, 0), "Asia/Seoul")
	if err != nil {
		t.Fatal(err)
	}
	if len(totals) != 2 || totals[0].Date != "2026-10-02" || totals[0].Seconds != 75*60 || totals[0].Sessions != 2 || totals[1].Date != "2026-10-03" {
		t.Fatalf("Seoul totals: %+v", totals)
	}
	utcTotals, _ := focus.DailyTotals(ctx, u.ID, utc(1, 0, 0), utc(31, 0, 0), "UTC")
	if len(utcTotals) != 3 {
		t.Fatalf("UTC totals: %+v", utcTotals)
	}
	days, err := focus.ActiveDays(ctx, u.ID, utc(1, 0, 0), "Asia/Seoul")
	if err != nil || len(days) != 2 || days[0] != "2026-10-03" {
		t.Fatalf("ActiveDays: %v %v", days, err)
	}

	// Deleting the user removes their rooms and focus history.
	if err := NewUserRepo(db).Delete(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := rooms.ByID(ctx, "abcdefgh"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("room should be gone: %v", err)
	}
	var n int
	if err := db.SQL().QueryRow("SELECT count(*) FROM focus_sessions").Scan(&n); err != nil || n != 1 {
		t.Fatalf("focus sessions left: %d %v", n, err)
	}
}

func TestCleanupQueries(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	u := newUser(t, db, "u1", "a@example.com")
	now := time.Now().UTC()

	tokens := NewLoginTokenRepo(db)
	_ = tokens.Create(ctx, "old", "a@example.com", now.Add(-time.Hour))
	_ = tokens.Create(ctx, "fresh", "a@example.com", now.Add(time.Hour))
	if n, err := tokens.DeleteExpired(ctx, now); err != nil || n != 1 {
		t.Fatalf("DeleteExpired tokens: %d %v", n, err)
	}

	rooms := NewRoomRepo(db)
	owner := u.ID
	for _, id := range []room.ID{"aaaaaaaa", "bbbbbbbb"} {
		if err := rooms.Create(ctx, models.Room{ID: id, OwnerUserID: &owner, HostTokenHash: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	// "bbbbbbbb" has focus history and must survive; "aaaaaaaa" is stale and empty.
	err := NewFocusSessionRepo(db).Create(ctx, models.FocusSession{RoomID: "bbbbbbbb", StartedAt: now, EndedAt: now, Focused: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if n, err := rooms.DeleteStale(ctx, now.Add(time.Hour)); err != nil || n != 1 {
		t.Fatalf("DeleteStale: %d %v", n, err)
	}
	if _, err := rooms.ByID(ctx, "bbbbbbbb"); err != nil {
		t.Fatalf("room with history was deleted: %v", err)
	}
}

func TestCommunityTotalsIncludeAnonymousFocus(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if _, err := db.SQL().Exec("TRUNCATE anonymous_focus"); err != nil {
		t.Fatal(err)
	}
	u := newUser(t, db, "u1", "a@example.com")
	owner := u.ID
	if err := NewRoomRepo(db).Create(ctx, models.Room{ID: "aaaaaaaa", OwnerUserID: &owner, HostTokenHash: "x"}); err != nil {
		t.Fatal(err)
	}
	focus := NewFocusSessionRepo(db)
	day := time.Date(2026, 10, 2, 1, 0, 0, 0, time.UTC)
	if err := focus.Create(ctx, models.FocusSession{RoomID: "aaaaaaaa", StartedAt: day, EndedAt: day, Focused: 30 * time.Minute}); err != nil {
		t.Fatal(err)
	}
	if err := focus.CreateAnonymous(ctx, day.Add(time.Hour), 20*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := focus.CreateAnonymous(ctx, day.AddDate(0, -2, 0), 10*time.Minute); err != nil { // an older one
		t.Fatal(err)
	}

	month, err := focus.CommunityBetween(ctx, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || month.Seconds != 3000 || month.Sessions != 2 {
		t.Fatalf("month: %+v %v", month, err)
	}
	all, _ := focus.CommunityBetween(ctx, time.Unix(0, 0), time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	if all.Seconds != 3600 || all.Sessions != 3 {
		t.Fatalf("all: %+v", all)
	}
	days, err := focus.CommunityDaily(ctx, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), "UTC")
	if err != nil || len(days) != 1 || days[0].Date != "2026-10-02" || days[0].Seconds != 3000 {
		t.Fatalf("daily: %+v %v", days, err)
	}
}

func TestLeaderboardOrdersAndHonoursTheSwitch(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	users := NewUserRepo(db)
	rooms := NewRoomRepo(db)
	focus := NewFocusSessionRepo(db)
	day := time.Date(2026, 10, 2, 1, 0, 0, 0, time.UTC)

	for i, c := range []struct {
		id      string
		minutes int
	}{{"ua", 30}, {"ub", 90}, {"uc", 60}} {
		newUser(t, db, c.id, c.id+"@example.com")
		owner := models.UserID(c.id)
		room := room.ID(fmt.Sprintf("room%04d", i))
		if err := rooms.Create(ctx, models.Room{ID: room, OwnerUserID: &owner, HostTokenHash: "x"}); err != nil {
			t.Fatal(err)
		}
		if err := focus.Create(ctx, models.FocusSession{RoomID: room, StartedAt: day, EndedAt: day, Focused: time.Duration(c.minutes) * time.Minute}); err != nil {
			t.Fatal(err)
		}
	}
	from, to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	got, err := focus.Leaderboard(ctx, from, to, 10)
	if err != nil || len(got) != 3 || got[0].UserID != "ub" || got[1].UserID != "uc" || got[2].UserID != "ua" || got[0].Seconds != 5400 {
		t.Fatalf("order: %+v %v", got, err)
	}
	if limited, _ := focus.Leaderboard(ctx, from, to, 2); len(limited) != 2 {
		t.Fatalf("limit: %d", len(limited))
	}

	// A user who switches it off disappears; switching back brings them back.
	if err := users.SetLeaderboardVisible(ctx, "ub", false); err != nil {
		t.Fatal(err)
	}
	got, _ = focus.Leaderboard(ctx, from, to, 10)
	if len(got) != 2 || got[0].UserID != "uc" {
		t.Fatalf("hidden user still listed: %+v", got)
	}
	_ = users.SetLeaderboardVisible(ctx, "ub", true)
	if got, _ = focus.Leaderboard(ctx, from, to, 10); len(got) != 3 {
		t.Fatalf("user not back: %+v", got)
	}

	// Outside the period nothing is listed.
	if got, _ = focus.Leaderboard(ctx, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), 10); len(got) != 0 {
		t.Fatalf("outside the period: %+v", got)
	}
	if err := users.UpdateNickname(ctx, "ua", "Ayla K"); err != nil {
		t.Fatal(err)
	}
	if u, _ := users.ByID(ctx, "ua"); u.Nickname != "Ayla K" || !u.ShowOnLeaderboard {
		t.Fatalf("user: %+v", u)
	}
}

func TestAdminQueries(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if _, err := db.SQL().Exec("TRUNCATE anonymous_focus"); err != nil {
		t.Fatal(err)
	}
	admin := NewAdminRepo(db)
	focus := NewFocusSessionRepo(db)
	rooms := NewRoomRepo(db)
	now := time.Now().UTC()

	a := newUser(t, db, "ua", "a@example.com")
	b := newUser(t, db, "ub", "b@example.com")
	if err := NewUserRepo(db).SetLeaderboardVisible(ctx, "ub", false); err != nil {
		t.Fatal(err)
	}
	ownerA, ownerB := a.ID, b.ID
	_ = rooms.Create(ctx, models.Room{ID: "adminrma", OwnerUserID: &ownerA, HostTokenHash: "x"})
	_ = rooms.Create(ctx, models.Room{ID: "adminrmb", OwnerUserID: &ownerB, HostTokenHash: "x"})
	_ = focus.Create(ctx, models.FocusSession{RoomID: "adminrma", StartedAt: now.Add(-time.Hour), EndedAt: now, Focused: 40 * time.Minute})
	_ = focus.Create(ctx, models.FocusSession{RoomID: "adminrma", StartedAt: now.Add(-2 * time.Hour), EndedAt: now, Focused: 20 * time.Minute})
	_ = focus.Create(ctx, models.FocusSession{RoomID: "adminrmb", StartedAt: now.Add(-3 * time.Hour), EndedAt: now, Focused: 10 * time.Minute})
	_ = focus.CreateAnonymous(ctx, now.Add(-time.Hour), 30*time.Minute)

	_ = focus.CreateAnonymous(ctx, now.Add(-5*time.Minute), 7*time.Minute)
	recent, err := admin.RecentAnonymous(ctx, 1)
	if err != nil || len(recent) != 1 || recent[0].Focused != 7*time.Minute {
		t.Fatalf("recent anonymous (newest first, limited): %+v %v", recent, err)
	}
	if _, err := db.SQL().Exec("DELETE FROM anonymous_focus WHERE focused_ms = $1", (7 * time.Minute).Milliseconds()); err != nil {
		t.Fatal(err)
	}

	tot, err := admin.Totals(ctx)
	if err != nil || tot.Users != 2 || tot.Rooms != 2 || tot.RegisteredSeconds != 4200 || tot.RegisteredSessions != 3 || tot.AnonymousSeconds != 1800 || tot.AnonymousSessions != 1 || tot.HiddenUsers != 1 {
		t.Fatalf("totals: %+v %v", tot, err)
	}
	if n, err := admin.NewUsersSince(ctx, now.Add(-time.Hour)); err != nil || n != 2 {
		t.Fatalf("new users: %d %v", n, err)
	}
	if n, err := admin.ActiveHostsSince(ctx, now.Add(-90*time.Minute)); err != nil || n != 1 {
		t.Fatalf("active hosts since 90 min: %d %v", n, err) // only user a focused in the last 90 minutes
	}
	if n, _ := admin.ActiveHostsSince(ctx, now.Add(-4*time.Hour)); n != 2 {
		t.Fatalf("active hosts since 4 h: %d", n)
	}

	days, err := admin.Daily(ctx, now.Add(-48*time.Hour), now.Add(24*time.Hour), "UTC")
	if err != nil || len(days) == 0 {
		t.Fatalf("daily: %+v %v", days, err)
	}
	var regSec, anonSec int64
	var newUsers int
	for _, d := range days {
		regSec += d.RegisteredSeconds
		anonSec += d.AnonymousSeconds
		newUsers += d.NewUsers
	}
	if regSec != 4200 || anonSec != 1800 || newUsers != 2 {
		t.Fatalf("daily sums: reg %d anon %d users %d", regSec, anonSec, newUsers)
	}

	best, total, err := admin.Users(ctx, true, 10, 0)
	if err != nil || total != 2 || len(best) != 2 || best[0].Email != "a@example.com" || best[0].FocusSeconds != 3600 || best[0].Sessions != 2 || best[0].LastFocusAt == nil {
		t.Fatalf("most focused: %+v total %d %v", best, total, err)
	}
	if best[1].ShowOnLeaderboard || best[1].FocusSeconds != 600 {
		t.Fatalf("second user: %+v", best[1])
	}
	newest, _, _ := admin.Users(ctx, false, 1, 1)
	if len(newest) != 1 {
		t.Fatalf("paging: %+v", newest)
	}
	if size, err := admin.DatabaseSize(ctx); err != nil || size <= 0 {
		t.Fatalf("database size: %d %v", size, err)
	}
}
