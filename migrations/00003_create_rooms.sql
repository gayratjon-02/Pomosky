-- +goose Up
CREATE TABLE rooms (
    id              TEXT PRIMARY KEY CHECK (char_length(id) = 8),
    owner_user_id   TEXT REFERENCES users (id) ON DELETE SET NULL,
    host_token_hash TEXT NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    city            TEXT NOT NULL DEFAULT '',
    lat_rounded     DOUBLE PRECISION,
    lon_rounded     DOUBLE PRECISION,
    tz              TEXT NOT NULL DEFAULT 'UTC',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX rooms_owner_idx ON rooms (owner_user_id);

-- +goose Down
DROP TABLE rooms;
