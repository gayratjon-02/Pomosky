# Pomosky web (frontend)

Plain TypeScript modules bundled with Vite. There is no UI framework: pages are
rendered on the server (Go `html/template`), and these modules only add the
interactive parts (timer, WebSocket, sky canvas, host controls).

```
src/
  entries/   one entry per page bundle (landing, room, host, record)
  core/      protocol types, server clock sync, timer maths, reconnecting WebSocket
  sky/       live sky renderer: sun, moon, stars, clouds, rain, snow, fog, lightning
  room/      page controllers wiring core + sky + DOM
  ui/        small UI pieces (reactions, presence, toast, QR)
  styles/    CSS per page area
public/      files copied as-is (robots.txt, ...)
dist/        build output (git-ignored), served by the Go server under /static/
```

Commands: `npm install`, `npm run dev`, `npm run build`, `npm run typecheck`.

Build output uses stable file names (`/static/room.js`, `/static/base.css`) so the
Go templates can reference them directly. TODO: switch to hashed file names plus
the Vite manifest for long-lived caching.
