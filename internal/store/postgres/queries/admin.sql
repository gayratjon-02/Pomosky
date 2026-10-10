-- Queries for the operator's admin page. Counts and totals only, plus a user list.

-- name: AdminTotals :one
SELECT
  (SELECT COUNT(*) FROM users)::int AS users,
  (SELECT COUNT(*) FROM rooms)::int AS rooms,
  (SELECT COALESCE(SUM(focused_ms), 0) FROM focus_sessions)::bigint AS registered_ms,
  (SELECT COUNT(*) FROM focus_sessions)::int AS registered_sessions,
  (SELECT COALESCE(SUM(focused_ms), 0) FROM anonymous_focus)::bigint AS anonymous_ms,
  (SELECT COUNT(*) FROM anonymous_focus)::int AS anonymous_sessions,
  (SELECT COUNT(*) FROM users WHERE NOT show_on_leaderboard)::int AS hidden_users;

-- name: AdminNewUsersSince :one
SELECT COUNT(*)::int FROM users WHERE created_at >= $1;

-- name: AdminActiveHostsSince :one
SELECT COUNT(DISTINCT r.owner_user_id)::int
FROM focus_sessions fs JOIN rooms r ON r.id = fs.room_id
WHERE fs.started_at >= $1;

-- name: AdminDailyNewUsers :many
SELECT (created_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day, COUNT(*)::int AS users
FROM users
WHERE created_at >= sqlc.arg(from_ts)::timestamptz AND created_at < sqlc.arg(to_ts)::timestamptz
GROUP BY day ORDER BY day;

-- name: AdminDailyRegisteredFocus :many
SELECT (fs.started_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day,
       SUM(fs.focused_ms)::bigint AS focused_ms,
       COUNT(*)::int AS sessions,
       COUNT(DISTINCT r.owner_user_id)::int AS hosts
FROM focus_sessions fs JOIN rooms r ON r.id = fs.room_id
WHERE fs.started_at >= sqlc.arg(from_ts)::timestamptz AND fs.started_at < sqlc.arg(to_ts)::timestamptz
GROUP BY day ORDER BY day;

-- name: AdminDailyAnonymousFocus :many
SELECT (started_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day,
       SUM(focused_ms)::bigint AS focused_ms,
       COUNT(*)::int AS sessions
FROM anonymous_focus
WHERE started_at >= sqlc.arg(from_ts)::timestamptz AND started_at < sqlc.arg(to_ts)::timestamptz
GROUP BY day ORDER BY day;

-- name: AdminCountUsers :one
SELECT COUNT(*)::int FROM users;

-- name: AdminUsersNewestFirst :many
SELECT u.email, u.provider, u.nickname, u.avatar_url, u.created_at, u.show_on_leaderboard,
       COALESCE(a.ms, 0)::bigint AS focused_ms, COALESCE(a.n, 0)::int AS sessions, a.last_at
FROM users u
LEFT JOIN (
  SELECT r.owner_user_id AS uid, SUM(fs.focused_ms) AS ms, COUNT(*) AS n, MAX(fs.started_at) AS last_at
  FROM focus_sessions fs JOIN rooms r ON r.id = fs.room_id GROUP BY r.owner_user_id
) a ON a.uid = u.id
ORDER BY u.created_at DESC
LIMIT sqlc.arg(max_rows)::int OFFSET sqlc.arg(skip)::int;

-- name: AdminUsersMostFocused :many
SELECT u.email, u.provider, u.nickname, u.avatar_url, u.created_at, u.show_on_leaderboard,
       COALESCE(a.ms, 0)::bigint AS focused_ms, COALESCE(a.n, 0)::int AS sessions, a.last_at
FROM users u
LEFT JOIN (
  SELECT r.owner_user_id AS uid, SUM(fs.focused_ms) AS ms, COUNT(*) AS n, MAX(fs.started_at) AS last_at
  FROM focus_sessions fs JOIN rooms r ON r.id = fs.room_id GROUP BY r.owner_user_id
) a ON a.uid = u.id
ORDER BY COALESCE(a.ms, 0) DESC, u.created_at ASC
LIMIT sqlc.arg(max_rows)::int OFFSET sqlc.arg(skip)::int;

-- name: AdminDatabaseSize :one
SELECT pg_database_size(current_database())::bigint;
