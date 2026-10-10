# Architecture

Pomosky is a single Go application that renders HTML on the server (SSR, MVC)
and serves a WebSocket for the live timer. A small TypeScript bundle adds the
interactive parts in the browser.

## Layers

```
cmd/server                         entry point
internal/app                       composition root: the ONLY place that wires everything

  HTTP in                           realtime in
  controllers  ── views             realtime (Hub, RoomHub, Client)
  middleware
        \                              /
         \                            /
          room (pure core)  <── models ──> ports (interfaces)
                                              ^
              adapters:  store/postgres, auth/*, weather/openmeteo, platform/*
```

- **room**: pure core. `Reduce(state, event, now) -> state`. Standard library only.
- **models**: plain data types.
- **ports**: interfaces for repositories, clock, ID generator, mail, broadcaster.
- **adapters**: implement ports (Postgres, Open-Meteo, OAuth, ...).
- **controllers / views / middleware**: the HTTP side (MVC "C" and "V").
- **realtime**: WebSocket hub; talks to the core through `room.Reduce`.
- **app**: builds adapters, injects them, registers routes.

## Dependency rules

Enforced by `internal/archtest` (run with `go test ./...`). Each package may import
only the internal packages listed in `allowed` in `archtest_test.go`. In short:

| Package | May import |
|---|---|
| room, apperr, weather, config | nothing internal (room/apperr/weather/models: no third-party either) |
| devreload | nothing internal (development-only helper, standard library only) |
| models | room |
| ports | room, models |
| store/postgres | models, ports, room, apperr |
| auth | models, ports, apperr |
| realtime | room, ports, moderation, platform/ratelimit |
| middleware | models, auth, platform/httpx, platform/ratelimit |
| controllers | models, ports, room, views, config, weather, platform/httpx, apperr |
| app, cmd | anything (composition root) |

When a rule blocks you, the design is usually telling you something: move the
shared type down (to models or room) or add a port instead of importing an adapter.

## Request flow (page)

```
browser -> Caddy -> middleware (recover, request id, logging, security headers, optional auth)
        -> controller -> models/ports (data) -> views (HTML) -> browser
```

## Realtime flow

```
browser -- WebSocket --> realtime.Handler -> Client -> RoomHub.Apply(event)
                                                         |
                                       room.Reduce(state, event, now)   (pure)
                                                         |
                                  new state -> Encode -> Broadcast to all Clients
```

- The server owns the timer. Clients compute the remaining time from the
  server's `startedAt + duration` and a clock offset; they never tick over the network.
- Slow clients are dropped from a broadcast instead of blocking the room.
- Room state lives in memory; after a restart the host reconnects and sends
  `HostRecovered`. Only finished focus sessions and accounts are stored in Postgres.

## Optional login

`OptionalAuth` middleware attaches a user when a valid session cookie exists and
**never blocks**. Only host actions require a host token or room ownership.
