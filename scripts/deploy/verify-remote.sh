#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

BASE_URL="${VERIFY_BASE_URL:-http://127.0.0.1:${FLUXGATE_HTTP_PORT:-18080}}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
VERIFY_RETRIES="${VERIFY_RETRIES:-30}"
VERIFY_RETRY_DELAY_SECONDS="${VERIFY_RETRY_DELAY_SECONDS:-1}"
ADMIN_USERNAME="${VERIFY_ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-}}"
ADMIN_PASSWORD="${VERIFY_ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-}}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/verify-${DEPLOY_ID}.log"
COOKIE_JAR="$LOG_DIR/verify-${DEPLOY_ID}.cookies"

cleanup() {
  rm -f "$COOKIE_JAR"
}
trap cleanup EXIT

exec > >(tee "$LOG_FILE") 2>&1

curl_with_retry() {
  local path="$1"
  local attempt
  for ((attempt = 1; attempt <= VERIFY_RETRIES; attempt++)); do
    if curl -fsS "$BASE_URL$path"; then
      return 0
    fi
    log "verification retry $attempt/$VERIFY_RETRIES for $path"
    sleep "$VERIFY_RETRY_DELAY_SECONDS"
  done
  log "verification failed after $VERIFY_RETRIES attempts: $path"
  return 1
}

log "remote verification started: $BASE_URL"
if command -v docker >/dev/null 2>&1; then
  docker compose ps || true
fi

if command -v curl >/dev/null 2>&1; then
  log "+ curl -fsS $BASE_URL/healthz"
  curl_with_retry "/healthz"
  log "+ curl -fsS $BASE_URL/readyz"
  curl_with_retry "/readyz"
  if [[ -n "$ADMIN_USERNAME" && -n "$ADMIN_PASSWORD" ]]; then
    log "+ curl -fsS -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o /dev/null"
    curl -fsS -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
      -H 'content-type: application/json' \
      --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
      -o /dev/null
    log "+ curl -fsS -b $COOKIE_JAR $BASE_URL/api/overview"
    curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/overview"
  fi
else
  log "curl missing, skip HTTP verification"
fi

log "remote verification finished: $LOG_FILE"
