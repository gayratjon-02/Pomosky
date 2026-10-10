#!/usr/bin/env bash
# Creates deploy/.env from deploy/.env.example with fresh random secrets.
# It never overwrites an existing .env. Usage: init-env.sh [domain]
set -euo pipefail

cd "$(dirname "$0")/.."
if [ -e .env ]; then
  echo ".env already exists, leaving it alone" >&2
  exit 1
fi

random() { openssl rand -base64 48 | tr -d '/+=\n' | cut -c1-48; }

domain="${1:-pomosky.example.com}"
umask 077
sed \
  -e "s|^DOMAIN=.*|DOMAIN=${domain}|" \
  -e "s|^BASE_URL=.*|BASE_URL=https://${domain}|" \
  -e "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$(random)|" \
  -e "s|^HOST_TOKEN_SALT=.*|HOST_TOKEN_SALT=$(random)|" \
  .env.example > .env
chmod 600 .env
echo "created deploy/.env (mode 600) for ${domain}. Add login/SMTP settings if you want them."
