-- +goose Up
CREATE TABLE focus_sessions (
    id           BIGSERIAL PRIMARY KEY,
    room_id      TEXT NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    started_at   TIMESTAMPTZ NOT NULL,
    ended_at     TIMESTAMPTZ NOT NULL,
    peak_viewers INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX focus_sessions_room_started_idx ON focus_sessions (room_id, started_at DESC);

-- +goose Down
DROP TABLE focus_sessions;
