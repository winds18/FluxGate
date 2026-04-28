#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_CONFIG="${DEPLOY_CONFIG:-$ROOT_DIR/.env.deploy.local}"
if [[ -f "$DEPLOY_CONFIG" ]]; then
  # shellcheck disable=SC1090
  source "$DEPLOY_CONFIG"
fi

REMOTE_HOST="${REMOTE_HOST:-}"
REMOTE_DIR="${REMOTE_DIR:-}"
SERVICE="${1:-fluxgate}"
LINES="${LINES:-200}"
FOLLOW="${FOLLOW:-false}"

if [[ -z "$REMOTE_HOST" || -z "$REMOTE_DIR" ]]; then
  log "missing deploy config: set REMOTE_HOST and REMOTE_DIR via environment or $DEPLOY_CONFIG"
  exit 64
fi

case "$SERVICE" in
  fluxgate | sing-box | all) ;;
  *)
    log "unknown service: $SERVICE (expected fluxgate, sing-box, or all)"
    exit 64
    ;;
esac

log "reading remote logs: service=$SERVICE lines=$LINES follow=$FOLLOW"
ssh "$REMOTE_HOST" "REMOTE_DIR='$REMOTE_DIR' SERVICE='$SERVICE' LINES='$LINES' FOLLOW='$FOLLOW' bash -s" <<'REMOTE'
set -euo pipefail
cd "$REMOTE_DIR"

args=(logs --tail "$LINES" --timestamps)
if [[ "$FOLLOW" == "true" ]]; then
  args+=(--follow)
fi

if [[ "$SERVICE" != "all" ]]; then
  args+=("$SERVICE")
fi

docker compose "${args[@]}"
REMOTE
