-- name: GetUserStats :one
SELECT * FROM user_stats WHERE user_id = $1;

-- name: UpsertUserStats :exec
INSERT INTO user_stats (user_id, total_focus_ms, streak_days, last_active_date)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
SET total_focus_ms = user_stats.total_focus_ms + EXCLUDED.total_focus_ms,
    streak_days = EXCLUDED.streak_days,
    last_active_date = EXCLUDED.last_active_date;
