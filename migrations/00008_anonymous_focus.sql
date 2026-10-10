-- +goose Up
-- Focus time of rooms hosted without an account, for the community totals.
-- It holds no room id, no user and no location: only when and how long.
CREATE TABLE anonymous_focus (
    id         BIGSERIAL PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL,
    focused_ms BIGINT NOT NULL
);
CREATE INDEX anonymous_focus_started_at_idx ON anonymous_focus (started_at);

-- +goose Down
DROP TABLE anonymous_focus;
