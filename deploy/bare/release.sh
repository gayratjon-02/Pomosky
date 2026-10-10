#!/usr/bin/env bash
# Builds Pomosky on your computer and installs it on the server (no Docker).
# Run from anywhere:   deploy/bare/release.sh <user>@<server> [arm64|amd64]
# The server must have been prepared with install.sh. The architecture defaults
# to arm64 (Oracle Ampere); use amd64 for ordinary x86 servers.
set -euo pipefail

target="${1:?usage: release.sh <user>@<server> [arm64|amd64]}"
arch="${2:-arm64}"
root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

echo "==> frontend"
(cd web && npm ci --silent && npm run build --silent)

echo "==> backend (linux/${arch})"
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -ldflags="-s -w" -o bin/pomosky-linux ./cmd/server

echo "==> upload"
ssh "$target" 'mkdir -p /tmp/pomosky-release'
scp -q bin/pomosky-linux "$target:/tmp/pomosky-release/pomosky"
tar -C web/dist -cf - . | ssh "$target" 'rm -rf /tmp/pomosky-release/static && mkdir -p /tmp/pomosky-release/static && tar -C /tmp/pomosky-release/static -xf -'

echo "==> install and restart"
ssh "$target" 'sudo bash -s' <<'REMOTE'
set -euo pipefail
install -o pomosky -g pomosky -m 755 /tmp/pomosky-release/pomosky /opt/pomosky/pomosky.new
mv /opt/pomosky/pomosky.new /opt/pomosky/pomosky
rm -rf /opt/pomosky/static.new
cp -r /tmp/pomosky-release/static /opt/pomosky/static.new
chown -R pomosky:pomosky /opt/pomosky/static.new
rm -rf /opt/pomosky/static.old
[ -d /opt/pomosky/static ] && mv /opt/pomosky/static /opt/pomosky/static.old
mv /opt/pomosky/static.new /opt/pomosky/static
systemctl restart pomosky
sleep 2
systemctl is-active pomosky
rm -rf /tmp/pomosky-release /opt/pomosky/static.old
REMOTE

echo "==> done. Logs: ssh $target 'journalctl -u pomosky -f'"
