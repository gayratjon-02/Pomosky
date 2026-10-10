#!/usr/bin/env bash
# Restores a dump created by backup.sh into the running database.
# WARNING: this overwrites existing data. Test restores regularly on a spare machine.
set -euo pipefail

FILE="${1:?usage: restore.sh <backup.sql.gz>}"
read -r -p "This will overwrite the database. Continue? [y/N] " ans
[ "$ans" = "y" ] || exit 1

gunzip -c "$FILE" | docker compose exec -T postgres psql -U pomosky -d pomosky
echo "restored from $FILE"
