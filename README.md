# Pomosky

Pomodoro + sky. The host runs a shared pomodoro timer, viewers join by link and work together. The sky behind the room shows the real weather and sun position at the host's location.

- Backend and pages: Go (server-side rendering, MVC)
- Interactive parts: plain TypeScript modules (no UI framework)
- Database: PostgreSQL
- Hosting: Azure VM (Docker Compose, Caddy)

Status: early development, see [docs/plan.md](docs/plan.md) for the stages and [docs/architecture.md](docs/architecture.md) for the structure.

## Development

Requirements: Go (see `go.mod`) and Node 22.

```bash
make dev
```

Opens the server on http://localhost:8080 (change it with `ADDR=:8081 make dev`). It rebuilds and
restarts the Go server when Go files change, rebuilds the frontend (`vite build --watch`), reads
templates from disk, and reloads the browser tab whenever a template, a built asset or the server changes.
Stop it with Ctrl+C. The auto-reload only exists with `DEV_RELOAD=true`; the server refuses to start with it in production.

Other commands: `make test`, `make vet`, `make fmt`, `make run` (build the frontend once, then run the server).
