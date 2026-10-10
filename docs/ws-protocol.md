# WebSocket protocol

Endpoint: `GET /ws/{roomId}` (upgrade). Every frame is JSON:

```json
{ "type": "timer.start", "data": { "phase": "focus", "durationMs": 1500000, "label": "Go backend" } }
```

Keep `internal/realtime/protocol.go` and `web/src/core/protocol.ts` in sync.

## Client -> server

| type | data | who |
|---|---|---|
| `join` | `{ nickname }` | viewer |
| `host.auth` | `{ token }` | host |
| `timer.start` | `{ phase, durationMs, label }` | host |
| `timer.pause` / `timer.resume` / `timer.reset` | none | host |
| `reaction` | `{ kind }` (clap, coffee, fire, heart) | viewer |
| `ping` | none | any, every ~25 s |

## Server -> client

| type | data |
|---|---|
| `state` | `{ roomId, timer, viewers, reactions, city, serverNowMs }` |
| `timer.updated` | `{ timer }` |
| `presence` | `{ viewers }` |
| `reaction.tick` | `{ kind, count }` |
| `weather` | snapshot for the sky |
| `pong` | none |
| `error` | `{ code, message }` |

## Timer maths

`remaining = startedAtMs + durationMs - (clientNow + offset)` while running;
`pausedRemainingMs` while paused. `offset = serverNowMs - (sentAt + rtt/2)`,
smoothed with the median of recent samples.

## Rules

- Messages are small; the server limits frame size and rate per connection.
- The heartbeat is required: idle TCP connections are dropped by some cloud networks.
