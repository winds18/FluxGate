#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/deploy-${DEPLOY_ID}.log"

exec > >(tee "$LOG_FILE") 2>&1

require_cmd docker
cd "$ROOT_DIR"

log "deployment started: $DEPLOY_ID"
run_logged "$ROOT_DIR/scripts/deploy/probe-env.sh"
export DRY_RUN="${DRY_RUN:-false}"
run_logged "$ROOT_DIR/scripts/deploy/cleanup-disk.sh"
run_logged "$ROOT_DIR/scripts/db/backup.sh"
run_logged "$ROOT_DIR/scripts/sing-box/backup-config.sh"

REMOTE_BUILD_ENABLED="${REMOTE_BUILD_ENABLED:-true}"
if [[ "$REMOTE_BUILD_ENABLED" == "true" ]]; then
  run_logged "$ROOT_DIR/scripts/deploy/remote-build.sh"
  if ! docker compose pull sing-box; then
    if docker image inspect ghcr.io/sagernet/sing-box:latest >/dev/null 2>&1; then
      log "sing-box pull failed; using cached ghcr.io/sagernet/sing-box:latest image"
    else
      log "sing-box pull failed and no cached image is available"
      exit 1
    fi
  fi
else
  run_logged docker compose pull
fi

run_logged docker compose up -d
log "deployment finished: $DEPLOY_ID"
