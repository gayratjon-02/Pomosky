package postgres

import (
	"context"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// FocusSessionRepo implements ports.FocusSessionRepository.
type FocusSessionRepo struct{ db *DB }

var _ ports.FocusSessionRepository = (*FocusSessionRepo)(nil)

// NewFocusSessionRepo creates the repository.
func NewFocusSessionRepo(db *DB) *FocusSessionRepo { return &FocusSessionRepo{db: db} }

func (r *FocusSessionRepo) Create(ctx context.Context, s models.FocusSession) error {
	return mapErr(r.db.q.CreateFocusSession(ctx, sqlcgen.CreateFocusSessionParams{
		RoomID:      string(s.RoomID),
		StartedAt:   s.StartedAt,
		EndedAt:     s.EndedAt,
		FocusedMs:   s.Focused.Milliseconds(),
		PeakViewers: int32(s.PeakViewers),
	}))
}

func (r *FocusSessionRepo) DailyTotals(ctx context.Context, owner models.UserID, from, to time.Time, tz string) ([]models.DayStat, error) {
	rows, err := r.db.q.DailyFocusTotals(ctx, sqlcgen.DailyFocusTotalsParams{Tz: tz, Owner: string(owner), FromTs: from, ToTs: to})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]models.DayStat, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.DayStat{
			Date:     row.Day.Format("2006-01-02"),
			Seconds:  row.FocusedMs / 1000,
			Sessions: int(row.Sessions),
		})
	}
	return out, nil
}

func (r *FocusSessionRepo) ActiveDays(ctx context.Context, owner models.UserID, since time.Time, tz string) ([]string, error) {
	days, err := r.db.q.ActiveFocusDays(ctx, sqlcgen.ActiveFocusDaysParams{Tz: tz, Owner: string(owner), Since: since})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]string, 0, len(days))
	for _, d := range days {
		out = append(out, d.Format("2006-01-02"))
	}
	return out, nil
}

func (r *FocusSessionRepo) CreateAnonymous(ctx context.Context, startedAt time.Time, focused time.Duration) error {
	return mapErr(r.db.q.CreateAnonymousFocus(ctx, sqlcgen.CreateAnonymousFocusParams{StartedAt: startedAt, FocusedMs: focused.Milliseconds()}))
}

func (r *FocusSessionRepo) CommunityBetween(ctx context.Context, from, to time.Time) (models.Period, error) {
	row, err := r.db.q.CommunityFocusBetween(ctx, sqlcgen.CommunityFocusBetweenParams{FromTs: from, ToTs: to})
	if err != nil {
		return models.Period{}, mapErr(err)
	}
	return models.Period{Seconds: row.FocusedMs / 1000, Sessions: int(row.Sessions)}, nil
}

func (r *FocusSessionRepo) CommunityDaily(ctx context.Context, from, to time.Time, tz string) ([]models.DayStat, error) {
	rows, err := r.db.q.CommunityDailyFocus(ctx, sqlcgen.CommunityDailyFocusParams{Tz: tz, FromTs: from, ToTs: to})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]models.DayStat, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.DayStat{Date: row.Day.Format("2006-01-02"), Seconds: row.FocusedMs / 1000, Sessions: int(row.Sessions)})
	}
	return out, nil
}

func (r *FocusSessionRepo) Leaderboard(ctx context.Context, from, to time.Time, limit int) ([]models.LeaderEntry, error) {
	rows, err := r.db.q.LeaderboardBetween(ctx, sqlcgen.LeaderboardBetweenParams{FromTs: from, ToTs: to, MaxRows: int32(limit)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]models.LeaderEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.LeaderEntry{
			UserID: models.UserID(row.UserID), Nickname: row.Nickname, AvatarURL: row.AvatarUrl,
			Seconds: row.FocusedMs / 1000, Sessions: int(row.Sessions),
		})
	}
	return out, nil
}
