-- Anonymous traffic analytics. Visitors are daily hashes: counts are per UTC day.

-- name: CreateVisit :exec
INSERT INTO visits (at, day, kind, path, visitor, source, medium, campaign, device, os, browser, country)
VALUES ($1, ($1 AT TIME ZONE 'UTC')::date, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: TrafficDaily :many
SELECT day,
       COUNT(*) FILTER (WHERE kind = 'pageview')::int AS pageviews,
       COUNT(DISTINCT visitor) FILTER (WHERE kind = 'pageview')::int AS visitors,
       COUNT(DISTINCT visitor) FILTER (WHERE kind = 'room_created')::int AS rooms_created,
       COUNT(DISTINCT visitor) FILTER (WHERE kind = 'login_done')::int AS logins
FROM visits
WHERE day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY day ORDER BY day;

-- name: TrafficPages :many
SELECT path, COUNT(*)::int AS pageviews, COUNT(DISTINCT (day, visitor))::int AS visitors
FROM visits
WHERE kind = 'pageview' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY path ORDER BY pageviews DESC LIMIT 15;

-- name: TrafficSources :many
WITH first_touch AS (
  SELECT DISTINCT ON (day, visitor) day, visitor, source, medium
  FROM visits
  WHERE kind = 'pageview' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
  ORDER BY day, visitor, at
)
SELECT f.source, f.medium, COUNT(*)::int AS visitors,
       COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM visits v WHERE v.day = f.day AND v.visitor = f.visitor AND v.kind = 'room_created'))::int AS rooms_created,
       COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM visits v WHERE v.day = f.day AND v.visitor = f.visitor AND v.kind = 'login_done'))::int AS logins
FROM first_touch f
GROUP BY f.source, f.medium ORDER BY visitors DESC LIMIT 20;

-- name: TrafficDevices :many
SELECT device AS label, COUNT(DISTINCT (day, visitor))::int AS visitors
FROM visits WHERE kind = 'pageview' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY device ORDER BY visitors DESC;

-- name: TrafficSystems :many
SELECT os AS label, COUNT(DISTINCT (day, visitor))::int AS visitors
FROM visits WHERE kind = 'pageview' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY os ORDER BY visitors DESC;

-- name: TrafficBrowsers :many
SELECT browser AS label, COUNT(DISTINCT (day, visitor))::int AS visitors
FROM visits WHERE kind = 'pageview' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY browser ORDER BY visitors DESC;

-- name: TrafficCountries :many
SELECT country AS label, COUNT(DISTINCT (day, visitor))::int AS visitors
FROM visits WHERE kind = 'pageview' AND country <> '' AND day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date
GROUP BY country ORDER BY visitors DESC LIMIT 15;

-- name: TrafficFunnel :one
SELECT
  COUNT(DISTINCT (day, visitor)) FILTER (WHERE kind = 'pageview')::int AS visitors,
  COUNT(DISTINCT (day, visitor)) FILTER (WHERE kind = 'pageview' AND path IN ('/r/:id', '/r/:id/host'))::int AS room_viewers,
  COUNT(DISTINCT (day, visitor)) FILTER (WHERE kind = 'room_created')::int AS rooms_created,
  COUNT(DISTINCT (day, visitor)) FILTER (WHERE kind = 'login_started')::int AS login_started,
  COUNT(DISTINCT (day, visitor)) FILTER (WHERE kind = 'login_done')::int AS login_done
FROM visits WHERE day >= sqlc.arg(from_day)::date AND day < sqlc.arg(to_day)::date;
