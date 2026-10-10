-- name: CreateFocusSession :exec
INSERT INTO focus_sessions (room_id, started_at, ended_at, focused_ms, peak_viewers)
VALUES ($1, $2, $3, $4, $5);

-- Focus time per calendar day of the owner, days taken in the given time zone.
-- name: DailyFocusTotals :many
SELECT (fs.started_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day,
       SUM(fs.focused_ms)::bigint AS focused_ms,
       COUNT(*)::int AS sessions
FROM focus_sessions fs
JOIN rooms r ON r.id = fs.room_id
WHERE r.owner_user_id = sqlc.arg(owner)::text
  AND fs.started_at >= sqlc.arg(from_ts)::timestamptz
  AND fs.started_at < sqlc.arg(to_ts)::timestamptz
GROUP BY day
ORDER BY day;

-- Days with any focus since a moment, newest first (used for the streak).
-- name: ActiveFocusDays :many
SELECT DISTINCT (fs.started_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day
FROM focus_sessions fs
JOIN rooms r ON r.id = fs.room_id
WHERE r.owner_user_id = sqlc.arg(owner)::text
  AND fs.started_at >= sqlc.arg(since)::timestamptz
ORDER BY day DESC;
