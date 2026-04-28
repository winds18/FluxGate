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
run_logged docker compose pull
run_logged docker compose up -d
log "deployment finished: $DEPLOY_ID"
