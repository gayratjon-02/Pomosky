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
| GET | `/static/*` | Built frontend assets | implemented |
| GET | `/__dev/events` | Reload event stream (SSE); only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/reload.js` | Browser auto-reload script; only with `DEV_RELOAD=true` | implemented |
| GET | `/__dev/sky` | Sky lab playground (sliders for time, place and weather); only with `DEV_RELOAD=true` | implemented |

State-changing forms use CSRF tokens (TODO stage 4).
