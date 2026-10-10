package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// RoomRepo implements ports.RoomRepository.
type RoomRepo struct{ db *DB }

var _ ports.RoomRepository = (*RoomRepo)(nil)

// NewRoomRepo creates the repository.
func NewRoomRepo(db *DB) *RoomRepo { return &RoomRepo{db: db} }

func (r *RoomRepo) Create(ctx context.Context, rec models.Room) error {
	owner := sql.NullString{}
	if rec.OwnerUserID != nil {
		owner = sql.NullString{String: string(*rec.OwnerUserID), Valid: true}
	}
	tz := rec.Location.TZ
	if tz == "" {
		tz = "UTC"
	}
	return mapErr(r.db.q.CreateRoom(ctx, sqlcgen.CreateRoomParams{
		ID:            string(rec.ID),
		OwnerUserID:   owner,
		HostTokenHash: rec.HostTokenHash,
		Name:          rec.Name,
		City:          rec.Location.City,
		LatRounded:    sql.NullFloat64{Float64: rec.Location.LatRounded, Valid: rec.Location.LatRounded != 0 || rec.Location.LonRounded != 0},
		LonRounded:    sql.NullFloat64{Float64: rec.Location.LonRounded, Valid: rec.Location.LatRounded != 0 || rec.Location.LonRounded != 0},
		Tz:            tz,
	}))
}

func (r *RoomRepo) ByID(ctx context.Context, id room.ID) (models.Room, error) {
	row, err := r.db.q.GetRoom(ctx, string(id))
	if err != nil {
		return models.Room{}, mapErr(err)
	}
	rec := models.Room{
		ID:            room.ID(row.ID),
		HostTokenHash: row.HostTokenHash,
		Name:          row.Name,
		CreatedAt:     row.CreatedAt,
		Location: room.Location{
			City:       row.City,
			LatRounded: row.LatRounded.Float64,
			LonRounded: row.LonRounded.Float64,
			TZ:         row.Tz,
		},
	}
	if row.OwnerUserID.Valid {
		owner := models.UserID(row.OwnerUserID.String)
		rec.OwnerUserID = &owner
	}
	return rec, nil
}

func (r *RoomRepo) SetOwner(ctx context.Context, id room.ID, owner models.UserID) error {
	return mapErr(r.db.q.SetRoomOwner(ctx, sqlcgen.SetRoomOwnerParams{
		ID: string(id), OwnerUserID: sql.NullString{String: string(owner), Valid: true},
	}))
}

func (r *RoomRepo) UpdateLocation(ctx context.Context, id room.ID, loc room.Location) error {
	return mapErr(r.db.q.UpdateRoomLocation(ctx, sqlcgen.UpdateRoomLocationParams{
		ID:         string(id),
		City:       loc.City,
		LatRounded: sql.NullFloat64{Float64: loc.LatRounded, Valid: true},
		LonRounded: sql.NullFloat64{Float64: loc.LonRounded, Valid: true},
		Tz:         loc.TZ,
	}))
}

// DeleteStale removes rooms created before the cutoff that have no focus history.
func (r *RoomRepo) DeleteStale(ctx context.Context, before time.Time) (int64, error) {
	n, err := r.db.q.DeleteStaleEmptyRooms(ctx, before)
	return n, mapErr(err)
}
