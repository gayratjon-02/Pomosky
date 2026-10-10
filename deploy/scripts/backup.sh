#!/usr/bin/env bash
# Dumps the database to a timestamped gzip file. Run from deploy/ on the server,
# then copy the file OFF the server (e.g. Azure Blob Storage). A backup that
# lives only on the same VM is not a backup.
set -euo pipefail

OUT_DIR="${1:-./backups}"
mkdir -p "$OUT_DIR"
FILE="$OUT_DIR/pomosky-$(date -u +%Y%m%dT%H%M%SZ).sql.gz"

docker compose exec -T postgres pg_dump -U pomosky -d pomosky | gzip > "$FILE"
echo "wrote $FILE"
