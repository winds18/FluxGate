#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

CONFIG_PATH="${SING_BOX_CONFIG_PATH:-$ROOT_DIR/data/sing-box/config.json}"
if [[ ! -f "$CONFIG_PATH" ]]; then
  log "sing-box config not found: $CONFIG_PATH"
  exit 1
fi

if command -v sing-box >/dev/null 2>&1; then
  run_logged sing-box check -c "$CONFIG_PATH"
else
  require_cmd docker
  run_logged docker run --rm -v "$(dirname "$CONFIG_PATH"):/etc/sing-box" ghcr.io/sagernet/sing-box check -c /etc/sing-box/"$(basename "$CONFIG_PATH")"
fi
