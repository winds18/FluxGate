#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
TARGET="$ROOT_DIR/logs/diagnostics/$DEPLOY_ID"
ensure_dir "$TARGET"

collect() {
  local name="$1"
  shift
  log "collecting $name"
  { "$@" || true; } >"$TARGET/$name.txt" 2>&1
}

collect git-status git -C "$ROOT_DIR" status --short --branch
collect git-log git -C "$ROOT_DIR" log --oneline -5
collect disk df -h "$ROOT_DIR"

if command -v docker >/dev/null 2>&1; then
  collect docker-ps docker ps
  collect docker-compose-ps docker compose -f "$ROOT_DIR/docker-compose.yml" ps
  collect fluxgate-logs docker logs --tail 200 fluxgate
  collect sing-box-logs docker logs --tail 200 fluxgate-sing-box
fi

if [[ -f "$ROOT_DIR/.env" ]]; then
  sed -n 's/^\([^#=][^=]*\)=.*/\1/p' "$ROOT_DIR/.env" | sort >"$TARGET/env-keys.txt"
fi

log "diagnostics collected: $TARGET"
