#!/usr/bin/env bash
# Writes a gzip dump of the database. Run on the server (cron), then copy it OFF the server.
#   sudo ./backup.sh [folder]        default folder: /var/backups/pomosky, keeps the last 14
set -euo pipefail
dir="${1:-/var/backups/pomosky}"
mkdir -p "$dir"
file="$dir/pomosky-$(date +%Y%m%d-%H%M%S).sql.gz"
sudo -u postgres pg_dump pomosky | gzip >"$file"
chmod 600 "$file"
ls -1t "$dir"/pomosky-*.sql.gz | tail -n +15 | xargs -r rm -f
echo "$file"
