#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

BASE_URL="${PUBLIC_BASE_URL:-http://127.0.0.1:8080}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/verify-${DEPLOY_ID}.log"

exec > >(tee "$LOG_FILE") 2>&1

log "remote verification started: $BASE_URL"
if command -v docker >/dev/null 2>&1; then
  docker compose ps || true
fi

if command -v curl >/dev/null 2>&1; then
  run_logged curl -fsS "$BASE_URL/healthz"
  run_logged curl -fsS "$BASE_URL/readyz"
else
  log "curl missing, skip HTTP verification"
fi

log "remote verification finished: $LOG_FILE"
