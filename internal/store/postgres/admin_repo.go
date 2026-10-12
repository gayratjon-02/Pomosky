package postgres

import (
	"context"
	"sort"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// AdminRepo implements ports.AdminRepository.
type AdminRepo struct{ db *DB }

var _ ports.AdminRepository = (*AdminRepo)(nil)

// NewAdminRepo creates the repository.
func NewAdminRepo(db *DB) *AdminRepo { return &AdminRepo{db: db} }

func (r *AdminRepo) Totals(ctx context.Context) (models.AdminTotals, error) {
	row, err := r.db.q.AdminTotals(ctx)
	if err != nil {
		return models.AdminTotals{}, mapErr(err)
	}
	return models.AdminTotals{
		Users: int(row.Users), Rooms: int(row.Rooms),
		RegisteredSeconds: row.RegisteredMs / 1000, RegisteredSessions: int(row.RegisteredSessions),
		AnonymousSeconds: row.AnonymousMs / 1000, AnonymousSessions: int(row.AnonymousSessions),
		HiddenUsers: int(row.HiddenUsers),
	}, nil
}

func (r *AdminRepo) NewUsersSince(ctx context.Context, since time.Time) (int, error) {
	n, err := r.db.q.AdminNewUsersSince(ctx, since)
	return int(n), mapErr(err)
}

func (r *AdminRepo) ActiveHostsSince(ctx context.Context, since time.Time) (int, error) {
	n, err := r.db.q.AdminActiveHostsSince(ctx, since)
	return int(n), mapErr(err)
}

func (r *AdminRepo) Daily(ctx context.Context, from, to time.Time, tz string) ([]models.AdminDay, error) {
	newUsers, err := r.db.q.AdminDailyNewUsers(ctx, sqlcgen.AdminDailyNewUsersParams{Tz: tz, FromTs: from, ToTs: to})
	if err != nil {
		return nil, mapErr(err)
	}
	reg, err := r.db.q.AdminDailyRegisteredFocus(ctx, sqlcgen.AdminDailyRegisteredFocusParams{Tz: tz, FromTs: from, ToTs: to})
	if err != nil {
		return nil, mapErr(err)
	}
	anon, err := r.db.q.AdminDailyAnonymousFocus(ctx, sqlcgen.AdminDailyAnonymousFocusParams{Tz: tz, FromTs: from, ToTs: to})
	if err != nil {
		return nil, mapErr(err)
	}

	byDate := map[string]*models.AdminDay{}
	day := func(t time.Time) *models.AdminDay {
		key := t.Format("2006-01-02")
		d, ok := byDate[key]
		if !ok {
			d = &models.AdminDay{Date: key}
			byDate[key] = d
		}
		return d
	}
	for _, x := range newUsers {
		day(x.Day).NewUsers = int(x.Users)
	}
	for _, x := range reg {
		d := day(x.Day)
		d.RegisteredSeconds, d.RegisteredSessions, d.ActiveHosts = x.FocusedMs/1000, int(x.Sessions), int(x.Hosts)
	}
	for _, x := range anon {
		d := day(x.Day)
		d.AnonymousSeconds, d.AnonymousSessions = x.FocusedMs/1000, int(x.Sessions)
	}
	out := make([]models.AdminDay, 0, len(byDate))
	for _, d := range byDate {
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out, nil
}

func (r *AdminRepo) Users(ctx context.Context, mostFocused bool, limit, offset int) ([]models.AdminUser, int, error) {
	total, err := r.db.q.AdminCountUsers(ctx)
	if err != nil {
		return nil, 0, mapErr(err)
	}
	out := []models.AdminUser{}
	if mostFocused {
		rows, err := r.db.q.AdminUsersMostFocused(ctx, sqlcgen.AdminUsersMostFocusedParams{MaxRows: int32(limit), Skip: int32(offset)})
		if err != nil {
			return nil, 0, mapErr(err)
		}
		for _, x := range rows {
			out = append(out, adminUser(x.Email.String, x.Provider, x.Nickname, x.AvatarUrl, x.CreatedAt, x.ShowOnLeaderboard, x.FocusedMs, int(x.Sessions), x.LastAt))
		}
	} else {
		rows, err := r.db.q.AdminUsersNewestFirst(ctx, sqlcgen.AdminUsersNewestFirstParams{MaxRows: int32(limit), Skip: int32(offset)})
		if err != nil {
			return nil, 0, mapErr(err)
		}
		for _, x := range rows {
			out = append(out, adminUser(x.Email.String, x.Provider, x.Nickname, x.AvatarUrl, x.CreatedAt, x.ShowOnLeaderboard, x.FocusedMs, int(x.Sessions), x.LastAt))
		}
	}
	return out, int(total), nil
}

func adminUser(email, provider, nickname, avatar string, joined time.Time, show bool, ms int64, sessions int, last any) models.AdminUser {
	u := models.AdminUser{
		Email: email, Provider: provider, Nickname: nickname, AvatarURL: avatar, JoinedAt: joined,
		ShowOnLeaderboard: show, FocusSeconds: ms / 1000, Sessions: sessions,
	}
	if t, ok := last.(time.Time); ok {
		u.LastFocusAt = &t
	}
	return u
}

func (r *AdminRepo) RecentAnonymous(ctx context.Context, limit int) ([]models.AnonymousSession, error) {
	rows, err := r.db.q.AdminRecentAnonymousFocus(ctx, int32(limit))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]models.AnonymousSession, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnonymousSession{StartedAt: row.StartedAt, Focused: time.Duration(row.FocusedMs) * time.Millisecond})
	}
	return out, nil
}

func (r *AdminRepo) DatabaseSize(ctx context.Context) (int64, error) {
	n, err := r.db.q.AdminDatabaseSize(ctx)
	return n, mapErr(err)
}
