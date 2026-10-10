-- +goose Up
-- First-party traffic analytics. No cookies, no IP address, no room id: a visitor
-- is a hash that changes every UTC day, so nobody can be followed across days.
CREATE TABLE visits (
    id       BIGSERIAL PRIMARY KEY,
    at       TIMESTAMPTZ NOT NULL,
    day      DATE NOT NULL,           -- UTC day of "at"
    kind     TEXT NOT NULL,           -- pageview, room_created, login_started, login_done
    path     TEXT NOT NULL,           -- normalised, e.g. /r/:id
    visitor  TEXT NOT NULL,           -- daily hash of client address and browser
    source   TEXT NOT NULL DEFAULT '', -- utm_source, referrer site or in-app browser
    medium   TEXT NOT NULL DEFAULT '',
    campaign TEXT NOT NULL DEFAULT '',
    device   TEXT NOT NULL DEFAULT '', -- mobile, tablet, desktop
    os       TEXT NOT NULL DEFAULT '',
    browser  TEXT NOT NULL DEFAULT '',
    country  TEXT NOT NULL DEFAULT ''  -- two letters, only when a proxy supplies it
);
CREATE INDEX visits_day_kind_idx ON visits (day, kind);
CREATE INDEX visits_day_visitor_idx ON visits (day, visitor);

-- +goose Down
DROP TABLE visits;
