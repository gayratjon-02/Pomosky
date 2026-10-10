# Runbook

## First deploy (Azure VM)

1. Prepare the VM: Docker + Compose, firewall (NSG) allowing only 80/443 (and SSH from your IP), SSH key login only.
2. Point DNS (A record) at the VM's static public IP.
3. `cd deploy && ./scripts/init-env.sh <your-domain>` (random secrets, mode 600) and add login/SMTP settings if wanted. Variables: `docs/env.md`.
4. `docker compose up -d --build`
5. Check `https://<domain>/healthz`.
6. The server applies the database migrations itself on start. Login works only when Google credentials or SMTP settings are set (see `.env.example`); everything else works without them.

## First deploy without Docker (small or free servers, e.g. Oracle Always Free)

1. On the server: `sudo ./install.sh <domain>` (PostgreSQL, Caddy, systemd service, secrets in `/etc/pomosky/pomosky.env`).
2. On your computer: `deploy/bare/release.sh <user>@<server-ip> [arm64|amd64]` builds and installs; run it again for every update.
3. Check `https://<domain>/healthz`. Logs: `journalctl -u pomosky -f`.
4. Backups: `sudo deploy/bare/backup.sh` from cron, and copy the file off the server.

## Frontend file names

The built files carry a content hash (`landing.CWdZvdpn.js`) and `dist/.vite/manifest.json` maps logical names
to them. The server reads the manifest to write the right names into the pages, serves hashed files with
`Cache-Control: immutable` for a year and never serves dot files. A deploy therefore cannot mix old and new
files in a browser cache. Keep `.vite/` when copying `web/dist` (`release.sh` and the Dockerfile do).

## Update

`deploy/scripts/deploy.sh` (pull, rebuild, restart).

## Backup and restore

- `deploy/scripts/backup.sh` writes a gzip dump. Copy it **off the VM** (e.g. Azure Blob Storage) on a schedule (cron).
- `deploy/scripts/restore.sh <file>` restores. **Test the restore** on a spare machine regularly.

## Rollback

Redeploy the previous git commit (`git checkout <sha>` then `docker compose up -d --build`).
Database migrations are forward-only in production; write `Down` migrations but test them before relying on them.

## Incidents

| Symptom | First checks |
|---|---|
| Site down | `docker compose ps`, `docker compose logs app caddy`, VM disk full? |
| WebSocket drops every few minutes | heartbeat running? proxy/network idle timeout (see ws-protocol.md) |
| Weather missing | upstream status, cache TTL, logs for provider errors (stale data is served when upstream fails) |
| Database errors | `docker compose logs postgres`, disk space, connection count |
| Cost spike | outbound traffic after a viral video: check Azure cost analysis, consider a CDN for static assets |
