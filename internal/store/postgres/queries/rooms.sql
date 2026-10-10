-- name: CreateRoom :exec
INSERT INTO rooms (id, owner_user_id, host_token_hash, name, city, lat_rounded, lon_rounded, tz)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetRoom :one
SELECT * FROM rooms WHERE id = $1;

-- name: SetRoomOwner :exec
UPDATE rooms SET owner_user_id = $2 WHERE id = $1;

-- name: UpdateRoomLocation :exec
UPDATE rooms SET city = $2, lat_rounded = $3, lon_rounded = $4, tz = $5 WHERE id = $1;
