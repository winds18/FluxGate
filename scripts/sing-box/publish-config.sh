#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

CONFIG_PATH="${SING_BOX_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.json}"
PREVIOUS_PATH="${SING_BOX_PREVIOUS_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.previous.json}"
CONTAINER_NAME="${SING_BOX_CONTAINER_NAME:-fluxgate-sing-box}"

ensure_dir "$(dirname "$CONFIG_PATH")"
if [[ -f "$CONFIG_PATH" ]]; then
  run_logged cp "$CONFIG_PATH" "$PREVIOUS_PATH"
fi

run_logged "$ROOT_DIR/scripts/sing-box/check-config.sh"
require_cmd docker
run_logged docker restart "$CONTAINER_NAME"
log "sing-box config published"
