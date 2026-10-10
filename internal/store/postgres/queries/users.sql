-- name: CreateUser :exec
INSERT INTO users (id, email, provider, provider_subject, nickname, avatar_url)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateUserAvatar :exec
UPDATE users SET avatar_url = $2 WHERE id = $1;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByIdentity :one
SELECT * FROM users WHERE provider = $1 AND provider_subject = $2;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: SetUserLeaderboardVisible :exec
UPDATE users SET show_on_leaderboard = $2 WHERE id = $1;

-- name: UpdateUserNickname :exec
UPDATE users SET nickname = $2 WHERE id = $1;
