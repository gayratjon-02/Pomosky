-- name: CreateUser :exec
INSERT INTO users (id, email, provider, provider_subject, nickname)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
