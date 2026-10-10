-- +goose Up
-- Profile picture URL from the login provider (Google). Empty when there is none.
ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_url;
