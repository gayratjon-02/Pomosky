-- +goose Up
-- Users appear on the public leaderboard (nickname and picture) unless they switch it off.
ALTER TABLE users ADD COLUMN show_on_leaderboard BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE users DROP COLUMN show_on_leaderboard;
