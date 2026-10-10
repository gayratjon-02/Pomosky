#!/usr/bin/env bash
# Pulls the latest code and restarts the stack. Run on the server, from the repo root.
set -euo pipefail

cd "$(dirname "$0")/.."
git pull --ff-only
docker compose up -d --build
docker compose ps
curl -fsS "http://localhost:8080/healthz" >/dev/null 2>&1 || echo "note: app is only reachable inside the compose network; check https://\$DOMAIN/healthz"
