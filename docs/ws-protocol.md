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
| `reaction` | `{ kind }` (clap, coffee, fire, heart) | anyone |
| `ping` | none | anyone, every ~25 s and a few times right after connecting (clock sync) |

## Server -> client

| type | data |
|---|---|
| `state` | `{ roomId, timer, viewers, reactions, city, serverNowMs }`: sent on join and after every change; carries the viewer count |
| `host.ok` | none |
| `reaction.tick` | `{ kind, count }` |
| `weather` | reserved for stage 6 |
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
- Messages are small: the server limits the frame size and the message rate per connection
  (`rate_limited`). Binary frames are not supported.
- The heartbeat is required: a connection that sends nothing for a long time is closed, and
  idle TCP connections are dropped by some cloud networks.
- The browser must send `Origin` equal to the page's host (same-origin); other origins are rejected.
