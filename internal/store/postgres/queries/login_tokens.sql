-- name: CreateLoginToken :exec
INSERT INTO login_tokens (token_hash, email, expires_at) VALUES ($1, $2, $3);

-- A token works once: consuming it deletes it, and expired tokens are not returned.
-- name: ConsumeLoginToken :one
DELETE FROM login_tokens WHERE token_hash = $1 AND expires_at > $2 RETURNING email;

-- name: DeleteExpiredLoginTokens :execrows
DELETE FROM login_tokens WHERE expires_at <= $1;
