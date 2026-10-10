-- +goose Up
-- Focused time is stored explicitly: a session can be paused, so ended_at - started_at is not it.
ALTER TABLE focus_sessions ADD COLUMN focused_ms BIGINT NOT NULL DEFAULT 0;

-- Deleting a user must delete the rooms they own, and with them the focus history.
ALTER TABLE rooms DROP CONSTRAINT rooms_owner_user_id_fkey;
ALTER TABLE rooms ADD CONSTRAINT rooms_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES users (id) ON DELETE CASCADE;

-- Statistics are computed from focus_sessions, so the pre-aggregated table is not needed.
DROP TABLE user_stats;

-- One-time email sign-in links. Only a hash of the token is stored.
CREATE TABLE login_tokens (
    token_hash TEXT PRIMARY KEY,
    email      TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX login_tokens_expires_at_idx ON login_tokens (expires_at);

-- +goose Down
DROP TABLE login_tokens;

CREATE TABLE user_stats (
    user_id          TEXT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    total_focus_ms   BIGINT NOT NULL DEFAULT 0,
    streak_days      INTEGER NOT NULL DEFAULT 0,
    last_active_date DATE
);

ALTER TABLE rooms DROP CONSTRAINT rooms_owner_user_id_fkey;
ALTER TABLE rooms ADD CONSTRAINT rooms_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE focus_sessions DROP COLUMN focused_ms;
