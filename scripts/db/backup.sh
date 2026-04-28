#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DB_PATH="${DB_PATH:-$ROOT_DIR/data/fluxgate.db}"
BACKUP_DIR="${BACKUP_DIR:-$ROOT_DIR/data/backups}"
ensure_dir "$BACKUP_DIR"

if [[ ! -f "$DB_PATH" ]]; then
  log "database not found, skip backup: $DB_PATH"
  exit 0
fi

target="$BACKUP_DIR/fluxgate-$(timestamp).db"
run_logged cp "$DB_PATH" "$target"
log "database backup created: $target"
