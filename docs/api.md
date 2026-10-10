# HTTP routes

| Method | Path | Purpose | Status |
|---|---|---|---|
| GET | `/` | Landing | implemented |
| GET | `/privacy` | Privacy page | implemented |
| GET | `/healthz` | Health check | implemented |
| GET | `/r/{id}` | Room (viewer); `?mode=record` for record mode | implemented |
| GET | `/r/{id}/host` | Host panel | implemented |
| POST | `/r/new` | Create a room; the host token goes in the URL fragment of the redirect. Logged-in hosts own the room | implemented |
| GET | `/ws/{id}` | WebSocket (see ws-protocol.md); 404 if the room does not exist | implemented |
| GET | `/api/weather?room={id}` | Cached weather at the host's place of a live room (JSON, see below). 404 until the host sets a place | implemented |
| GET, POST | `/login` | Login page; POST starts Google or emails a magic link (rate limited) | implemented |
| GET, POST | `/auth/callback` | Magic link: GET shows a confirmation, POST spends the token and signs in | implemented |
| GET | `/auth/google/callback` | Google redirect target (`code`, `state`) | implemented |
| POST | `/logout` | End the session | implemented |
| GET | `/account` | Monthly statistics. Logged-in: your own. Anonymous: example data behind a login prompt | implemented |
| POST | `/account/delete` | Delete the account and its data (`confirm=delete`) | implemented |
| GET | `/api/places?q=` | Place search for the host (JSON `{places:[{name,region,country,lat,lon,tz}]}`), rate limited, via Open-Meteo geocoding | implemented |
| GET | `/api/stats/demo?month=YYYY-MM` | Made-up example statistics in the same shape, for the logged-out preview | implemented |
| GET | `/api/stats/leaderboard?period=today|week|month|all&tz=` | Most focused users who allow it, best first (JSON `{period, entries:[{rank,name,avatarUrl,seconds,sessions,you}]}`), no ids or emails | implemented |
| POST | `/account/nickname`, `/account/leaderboard` | Change the public name; show or hide yourself on the leaderboard | implemented |
| GET | `/api/stats/community?tz=` | Focus time of everyone: today, last 7 days, this month, all time and the last 30 days (JSON), public and cached for a minute | implemented |
| GET | `/api/stats?month=YYYY-MM&tz=Area/City` | Monthly focus statistics of the logged-in user (JSON, see below) | implemented |
| GET | `/api/admin/traffic` | Anonymous traffic of the last 30 UTC days for operators: visitors and page views per day, rooms created, logins, sources, pages, devices, browsers, countries and the funnel (JSON, 404 for everyone else) | implemented |
| GET | `/static/*` | Built frontend assets | implemented |
| GET | `/__dev/events` | Reload event stream (SSE); only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/reload.js` | Browser auto-reload script; only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/mobile?path=/account&w=390&h=844` | Shows a page in a phone-sized frame (the dev server lets the site frame itself); only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/sky` | Sky lab playground (sliders for time, place and weather); only with `DEV_RELOAD=true` | implemented |

## GET /api/stats

Query: `month=YYYY-MM` (default: the current month). Needs a login; anonymous requests get 401.
Days are grouped in the time zone given by `tz` (IANA name, the browser sends its own); an unknown name falls back to UTC. `streakDays` counts consecutive days with focus time ending today (or yesterday while today is still empty); for a past month it ends on the month's last day.

```json
{
  "month": "2026-10",
  "totalSeconds": 86400,
  "sessions": 31,
  "activeDays": 18,
  "streakDays": 4,
  "prevTotalSeconds": 72000,
  "days": [{ "date": "2026-10-01", "seconds": 5400, "sessions": 3 }]
}
```

`days` lists every day of the month in order, including days with no focus time.

Every POST needs the CSRF token (hidden field `csrf` or header `X-CSRF-Token`, same value as the `pomosky_csrf` cookie).

## GET /api/weather

Query: `room={id}`. The place is the host's rounded location; viewers never send theirs.
Data comes from Open-Meteo (non-commercial free API, CC-BY 4.0, so pages that show it credit
Open-Meteo.com), cached on the server for `WEATHER_CACHE_TTL`, with one shared upstream call per
place and at most 60 upstream calls a minute. If the upstream fails, the last snapshot is served.

```json
{
  "cloudCoverPct": 32, "precipitationMmh": 0, "snowfallCmh": 0,
  "windSpeedKmh": 3.6, "windDirDeg": 307, "fogDensity": 0, "thunder": false,
  "isDay": false, "temperatureC": 20.3, "code": 1,
  "observedAtMs": 1791626400000, "fetchedAtMs": 1791626500000
}
```

`precipitationMmh` and `snowfallCmh` are hourly rates (Open-Meteo reports sums over the last
15 minutes and they are scaled). `fogDensity` and `thunder` come from the WMO weather code, which
also fills precipitation when the numbers are 0 but the code says rain or snow.
Errors: 400 bad room id, 404 unknown room or no place set, 502 weather unavailable.

## GET /api/stats/community

Query: `tz=Area/City` (days are taken in this zone; unknown names fall back to UTC). Public, cached for a minute.
It sums the focus time of logged-in hosts and of anonymous hosts (stored only as start time and length).

```json
{
  "today": { "seconds": 3600, "sessions": 2 },
  "week":  { "seconds": 5400, "sessions": 3 },
  "month": { "seconds": 6000, "sessions": 4 },
  "all":   { "seconds": 13200, "sessions": 7 },
  "days":  [{ "date": "2026-09-11", "seconds": 0, "sessions": 0 }]
}
```

`week` is the last 7 days including today, `month` the current calendar month, `days` the last 30 days (every day present).

## Traffic analytics

Every successful HTML page counts as a `pageview`; creating a room, starting a login and finishing a login are counted as `room_created`, `login_started` and `login_done`. Nothing is stored that identifies a person: no cookie, no IP address, no room id (paths are normalised to `/r/:id`). The visitor is `HMAC(key, UTC day, client address, User-Agent)`, truncated, so it changes every day. Bots, `DNT: 1` and `Sec-GPC: 1` are skipped. The source is `utm_source` (with `utm_medium`, `utm_campaign`), else the referrer's site, else an in-app browser (Instagram, TikTok, Facebook, Telegram), else `direct`. The country comes from `CF-IPCountry` when the site runs behind Cloudflare. Use `?utm_source=instagram&utm_medium=reel&utm_campaign=launch` in links to tell channels apart.
