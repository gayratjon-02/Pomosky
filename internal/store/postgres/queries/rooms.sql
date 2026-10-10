-- name: CreateRoom :exec
INSERT INTO rooms (id, owner_user_id, host_token_hash, name, city, lat_rounded, lon_rounded, tz)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetRoom :one
SELECT * FROM rooms WHERE id = $1;

-- name: SetRoomOwner :exec
UPDATE rooms SET owner_user_id = $2 WHERE id = $1;

-- name: UpdateRoomLocation :exec
UPDATE rooms SET city = $2, lat_rounded = $3, lon_rounded = $4, tz = $5 WHERE id = $1;

-- Rooms nobody has touched for a long time and that have no focus history.
-- name: DeleteStaleEmptyRooms :execrows
DELETE FROM rooms r
WHERE r.created_at < $1
  AND NOT EXISTS (SELECT 1 FROM focus_sessions fs WHERE fs.room_id = r.id);
