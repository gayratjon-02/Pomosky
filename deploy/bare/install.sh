#!/usr/bin/env bash
# One-time setup of a fresh Ubuntu/Debian server WITHOUT Docker: PostgreSQL,
# Caddy (automatic HTTPS) and a systemd service for Pomosky.
#
# Run on the server:   sudo ./install.sh <domain>
# <domain> must already point to this server (A record). Use "localhost" to skip HTTPS.
# Safe to run again: it keeps the existing database password and env file.
set -euo pipefail

domain="${1:?usage: sudo ./install.sh <domain>}"
[ "$(id -u)" -eq 0 ] || { echo "run as root (sudo)" >&2; exit 1; }

APP_USER=pomosky
APP_DIR=/opt/pomosky
ENV_FILE=/etc/pomosky/pomosky.env
random() { openssl rand -base64 48 | tr -d '/+=\n' | cut -c1-48; }

# Small servers (under 2 GB RAM) get a 2 GB swap file so a memory spike does not kill the database.
if [ "$(awk '/MemTotal/ {print int($2/1024)}' /proc/meminfo)" -lt 2000 ] && [ -z "$(swapon --show --noheadings)" ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile >/dev/null
  swapon /swapfile
  grep -q '^/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo "created a 2 GB swap file"
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y postgresql curl gnupg openssl ca-certificates

# Caddy from the distribution's own packages. (Caddy's third-party apt repository
# answered "402 Payment Required" in October 2026, so it is not used.)
rm -f /etc/apt/sources.list.d/caddy-stable.list /usr/share/keyrings/caddy-stable-archive-keyring.gpg
if ! command -v caddy >/dev/null; then
  apt-get update -y
  apt-get install -y caddy
fi

# Service user and folders.
id "$APP_USER" >/dev/null 2>&1 || useradd --system --home "$APP_DIR" --shell /usr/sbin/nologin "$APP_USER"
mkdir -p "$APP_DIR/static" /etc/pomosky
chown -R "$APP_USER":"$APP_USER" "$APP_DIR"

# Database and settings, created once.
if [ ! -f "$ENV_FILE" ]; then
  db_password="$(random)"
  sudo -u postgres psql -v ON_ERROR_STOP=1 <<SQL
DO \$\$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'pomosky') THEN
    CREATE ROLE pomosky LOGIN PASSWORD '${db_password}';
  ELSE
    ALTER ROLE pomosky PASSWORD '${db_password}';
  END IF;
END \$\$;
SQL
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='pomosky'" | grep -q 1 \
    || sudo -u postgres createdb -O pomosky pomosky

  scheme="https"; [ "$domain" = "localhost" ] && scheme="http"
  umask 077
  cat >"$ENV_FILE" <<ENV
APP_ENV=production
ADDR=127.0.0.1:8080
BASE_URL=${scheme}://${domain}
STATIC_DIR=${APP_DIR}/static
TRUST_PROXY=true
DATABASE_URL=postgres://pomosky:${db_password}@127.0.0.1:5432/pomosky?sslmode=disable
HOST_TOKEN_SALT=$(random)
ROOM_IDLE_TTL=10m
WEATHER_CACHE_TTL=10m
# Login (optional). See docs/env.md:
# OAUTH_GOOGLE_CLIENT_ID=
# OAUTH_GOOGLE_CLIENT_SECRET=
# MAIL_FROM=
# SMTP_HOST=
# SMTP_PORT=587
# SMTP_USERNAME=
# SMTP_PASSWORD=
ENV
  chown root:"$APP_USER" "$ENV_FILE"
  chmod 640 "$ENV_FILE"
  echo "created $ENV_FILE"
else
  echo "keeping existing $ENV_FILE"
fi

# systemd service.
cat >/etc/systemd/system/pomosky.service <<UNIT
[Unit]
Description=Pomosky
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
User=${APP_USER}
Group=${APP_USER}
WorkingDirectory=${APP_DIR}
EnvironmentFile=${ENV_FILE}
ExecStart=${APP_DIR}/pomosky
Restart=always
RestartSec=3
# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
PrivateDevices=true
ProtectKernelTunables=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
MemoryMax=400M

[Install]
WantedBy=multi-user.target
UNIT

# Caddy: HTTPS, compression, WebSocket upgrades (reverse_proxy handles them).
if [ "$domain" = "localhost" ]; then site=":80"; else site="$domain"; fi
cat >/etc/caddy/Caddyfile <<CADDY
${site} {
	encode zstd gzip
	# Cache headers of /static/* come from the app (hashed files are immutable).
	reverse_proxy 127.0.0.1:8080
	header {
		Strict-Transport-Security "max-age=31536000"
		-Server
	}
}
CADDY

systemctl daemon-reload
systemctl enable pomosky >/dev/null
systemctl reload caddy 2>/dev/null || systemctl restart caddy

cat <<DONE

Server is ready. Next, from your computer:   deploy/bare/release.sh <user>@<server-ip>
Then check:  https://${domain}/healthz
Remember to open ports 80 and 443 in the cloud firewall (and keep 22 for your IP only).
DONE
