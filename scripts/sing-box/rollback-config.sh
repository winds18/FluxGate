#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

CONFIG_PATH="${SING_BOX_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.json}"
PREVIOUS_PATH="${SING_BOX_PREVIOUS_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.previous.json}"
CONTAINER_NAME="${SING_BOX_CONTAINER_NAME:-fluxgate-sing-box}"

if [[ ! -f "$PREVIOUS_PATH" ]]; then
  log "previous config not found: $PREVIOUS_PATH"
  exit 1
fi

run_logged cp "$PREVIOUS_PATH" "$CONFIG_PATH"
require_cmd docker
run_logged docker restart "$CONTAINER_NAME"
log "sing-box config rolled back"
