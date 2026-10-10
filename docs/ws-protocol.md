# WebSocket protocol

Endpoint: `GET /ws/{roomId}` (upgrade). Every frame is a JSON text frame:

```json
{ "type": "timer.start", "data": { "phase": "focus", "durationMs": 1500000, "label": "Go backend" } }
```

Keep `internal/realtime/protocol.go` and `web/src/core/protocol.ts` in sync.

## Client -> server

| type | data | who |
|---|---|---|
| `join` | `{ nickname }` | viewer (optional, validated) |
| `host.auth` | `{ token }` | host; answered with `host.ok` or an `error` |
| `timer.start` | `{ phase: "focus" or "break", durationMs, label }` | host |
| `timer.pause` / `timer.resume` / `timer.reset` | none | host |
| `timer.recover` | `{ timer }` (the host's last known timer) | host, used after a server restart; ignored if the room already has a timer |
| `location.set` | `{ city, lat, lon, tz }` (tz is an IANA name; the server rounds lat/lon to 2 decimals, about 1 km) | host |
| `reaction` | `{ kind }` (clap, coffee, fire, heart) | anyone |
| `ping` | none | anyone, every ~25 s and a few times right after connecting (clock sync) |

## Server -> client

| type | data |
|---|---|
| `state` | `{ roomId, timer, viewers, reactions, city, location?, serverNowMs }`: sent on join and after every change; carries the viewer count. `location` is `{ lat, lon, tz }` (rounded) and is absent until the host sets it |
| `people` | `{ people: [{ name, avatarUrl? }], more }`: the named people in the room (signed-in users who allow it, once however many tabs, and viewers who chose a nickname), at most 12, `more` counts the rest. Sent on join and whenever the list changes; no ids or emails |
| `host.ok` | none |
| `reaction.tick` | `{ kind, count }` |
| `weather` | not used: clients fetch `GET /api/weather?room={id}` instead (see api.md) |
| `pong` | `{ serverNowMs }` |
| `error` | `{ code, message }` |

`timer` is `{ phase, pausedFrom?, startedAtMs, durationMs, pausedRemainingMs, label }`.

Error codes: `bad_request`, `forbidden`, `invalid`, `invalid_transition`, `rate_limited`.

## Timer maths

`remaining = startedAtMs + durationMs - (clientNow + offset)` while running;
`pausedRemainingMs` while paused. `offset = serverNowMs - (sentAt + rtt/2)` from a
`ping`/`pong` round trip, smoothed with the median of recent samples.

## Rules

- Only the host (after `host.auth`) can send `timer.*` messages; anyone else gets `forbidden`.
- `timer.start` is accepted only when the timer is idle or the running phase has reached its end. While a phase
  runs or is paused it is answered with `invalid_transition`, so pressing start twice never restarts the clock.
  Pause, resume or reset first. `timer.pause` needs a running phase, `timer.resume` a paused one.
- Messages are small: the server limits the frame size and the message rate per connection
  (`rate_limited`). Binary frames are not supported.
- The heartbeat is required: a connection that sends nothing for a long time is closed, and
  idle TCP connections are dropped by some cloud networks.
- The browser must send `Origin` equal to the page's host (same-origin); other origins are rejected.
