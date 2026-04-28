#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

PORT="${PORT:-8081}"
BASE_URL="http://127.0.0.1:$PORT"
QA_DB="$ROOT_DIR/data/qa-fluxgate.db"
QA_LOG_DIR="$ROOT_DIR/logs/fluxgate-qa"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"

cleanup() {
  PORT="$PORT" "$ROOT_DIR/scripts/dev/stop.sh" >/dev/null 2>&1 || true
  KEEP_ARTIFACTS="$KEEP_ARTIFACTS" "$ROOT_DIR/scripts/qa/cleanup.sh" >/dev/null 2>&1 || true
}
trap cleanup EXIT

log "running local QA suite"
PORT="$PORT" "$ROOT_DIR/scripts/dev/stop.sh" >/dev/null 2>&1 || true
KEEP_ARTIFACTS="$KEEP_ARTIFACTS" "$ROOT_DIR/scripts/qa/cleanup.sh"
ensure_dir "$ROOT_DIR/tmp"

DB_PATH="$QA_DB" LOG_DIR="$QA_LOG_DIR" HTTP_ADDR="127.0.0.1:$PORT" "$ROOT_DIR/scripts/dev/run.sh" >"$ROOT_DIR/tmp/qa-server.log" 2>&1 &

ready=false
for _ in {1..30}; do
  if curl -fsS "$BASE_URL/healthz" >/dev/null 2>&1; then
    ready=true
    break
  fi
  sleep 1
done
if [[ "$ready" != "true" ]]; then
  log "local server did not become ready"
  exit 1
fi

BASE_URL="$BASE_URL" "$ROOT_DIR/scripts/qa/smoke.sh"
BASE_URL="$BASE_URL" "$ROOT_DIR/scripts/qa/api-flow.sh"
URL="$BASE_URL" "$ROOT_DIR/scripts/qa/screenshot.sh"

log "local QA suite passed"
