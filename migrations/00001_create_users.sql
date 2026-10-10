-- +goose Up
CREATE TABLE users (
    id               TEXT PRIMARY KEY,
    email            TEXT UNIQUE,
    provider         TEXT NOT NULL,
    provider_subject TEXT NOT NULL,
    nickname         TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_subject)
);

-- +goose Down
DROP TABLE users;
