#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "bootstrapping remote deployment directory"
ensure_dir "$ROOT_DIR/data/backups"
ensure_dir "$ROOT_DIR/data/sing-box/versions"
ensure_dir "$ROOT_DIR/logs/deploy"
ensure_dir "$ROOT_DIR/logs/fluxgate"
ensure_dir "$ROOT_DIR/logs/sing-box"
ensure_dir "$ROOT_DIR/logs/diagnostics"

if [[ ! -f "$ROOT_DIR/data/sing-box/config.json" ]]; then
  cat >"$ROOT_DIR/data/sing-box/config.json" <<'JSON'
{
  "log": {
    "level": "info"
  },
  "inbounds": [],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct"
    },
    {
      "type": "block",
      "tag": "block"
    }
  ],
  "route": {
    "final": "direct"
  }
}
JSON
  log "minimal sing-box config created"
fi

if [[ ! -f "$ROOT_DIR/.env" && -f "$ROOT_DIR/.env.example" ]]; then
  run_logged cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
  log ".env created from .env.example; edit secrets before production use"
fi

log "remote bootstrap complete"
