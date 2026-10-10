package postgres

import (
	"context"
	"errors"
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
