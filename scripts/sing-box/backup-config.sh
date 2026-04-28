#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

CONFIG_PATH="${SING_BOX_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.json}"
BACKUP_DIR="${SING_BOX_BACKUP_DIR:-$ROOT_DIR/data/sing-box/versions}"
ensure_dir "$BACKUP_DIR"

if [[ ! -f "$CONFIG_PATH" ]]; then
  log "sing-box config not found, skip backup: $CONFIG_PATH"
  exit 0
fi

target="$BACKUP_DIR/config-$(timestamp).json"
run_logged cp "$CONFIG_PATH" "$target"
log "sing-box config backup created: $target"
