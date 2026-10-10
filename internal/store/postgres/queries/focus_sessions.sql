-- name: CreateFocusSession :exec
INSERT INTO focus_sessions (room_id, started_at, ended_at, peak_viewers)
VALUES ($1, $2, $3, $4);

-- name: ListFocusSessionsByRoom :many
SELECT * FROM focus_sessions WHERE room_id = $1 ORDER BY started_at DESC LIMIT $2;
