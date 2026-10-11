#!/usr/bin/env bash
# Downloads the current free DB-IP "IP to Country Lite" table (CC BY 4.0, https://db-ip.com)
# to /opt/pomosky/geoip/dbip-country-lite.csv.gz and restarts Pomosky so it is read.
# DB-IP publishes a new file at the start of every month. Run as root, for example from cron:
#   0 4 3 * * /opt/pomosky/update-geoip.sh
# Set GEOIP_FILE=/opt/pomosky/geoip/dbip-country-lite.csv.gz in /etc/pomosky/pomosky.env once.
set -euo pipefail

dir=/opt/pomosky/geoip
mkdir -p "$dir"
tmp="$(mktemp "$dir/new.XXXXXX")"
trap 'rm -f "$tmp"' EXIT

ok=""
for month in "$(date -u +%Y-%m)" "$(date -u -d '1 month ago' +%Y-%m)"; do
  if curl -fsSL --max-time 180 -o "$tmp" "https://download.db-ip.com/free/dbip-country-lite-${month}.csv.gz"; then
    ok="$month"
    break
  fi
done
[ -n "$ok" ] || { echo "could not download the table" >&2; exit 1; }

# A truncated or wrong file must never replace the working one.
gzip -t "$tmp"
[ "$(zcat "$tmp" | wc -l)" -gt 100000 ] || { echo "the table looks too small" >&2; exit 1; }

chown pomosky:pomosky "$tmp"
chmod 644 "$tmp"
mv "$tmp" "$dir/dbip-country-lite.csv.gz"
trap - EXIT
systemctl restart pomosky
echo "installed DB-IP country table ${ok}"
