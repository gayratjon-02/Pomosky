-- +goose Up
CREATE TABLE user_stats (
    user_id          TEXT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    total_focus_ms   BIGINT NOT NULL DEFAULT 0,
    streak_days      INTEGER NOT NULL DEFAULT 0,
    last_active_date DATE
);

-- +goose Down
DROP TABLE user_stats;
