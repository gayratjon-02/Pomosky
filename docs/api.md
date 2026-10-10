# HTTP routes

| Method | Path | Purpose | Status |
|---|---|---|---|
| GET | `/` | Landing | implemented |
| GET | `/privacy` | Privacy page | implemented |
| GET | `/healthz` | Health check | implemented |
| GET | `/r/{id}` | Room (viewer); `?mode=record` for record mode | stub |
| GET | `/r/{id}/host` | Host panel | stub |
| POST | `/r/new` | Create a room, show host token once | stub |
| GET | `/ws/{id}` | WebSocket (see ws-protocol.md); 404 if the room does not exist | implemented |
| GET | `/api/weather?room={id}` | Cached weather snapshot | stub |
| GET/POST | `/login`, `/auth/callback`, `/logout` | Optional login | stub |
| GET | `/account` | Profile and statistics | stub |
| GET | `/api/stats?month=YYYY-MM` | Monthly focus statistics of the logged-in user (JSON, see below) | stub |
| GET | `/static/*` | Built frontend assets | implemented |
| GET | `/__dev/events` | Reload event stream (SSE); only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/reload.js` | Browser auto-reload script; only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/account`, `/__dev/api/stats` | Account page with a fake user and fixture statistics; only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/sky` | Sky lab playground (sliders for time, place and weather); only with `DEV_RELOAD=true` | implemented |

## GET /api/stats

Query: `month=YYYY-MM` (default: the current month). Needs a login; anonymous requests get 401.
Days are grouped in the user's time zone (TODO(stage 4): decide how the zone is sent; likely a `tz` parameter or a stored profile setting).

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

State-changing forms use CSRF tokens (TODO stage 4).
