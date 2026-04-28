#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-}}"
COOKIE_JAR="${COOKIE_JAR:-$ROOT_DIR/tmp/qa-smoke-cookies.txt}"
require_cmd curl

log "running smoke tests against $BASE_URL"
run_logged curl -fsS "$BASE_URL/healthz"
run_logged curl -fsS "$BASE_URL/readyz"
status_code="$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/overview")"
if [[ "$status_code" != "401" ]]; then
  log "expected unauthenticated overview to return 401, got $status_code"
  exit 1
fi
if [[ -n "$ADMIN_USERNAME" && -n "$ADMIN_PASSWORD" ]]; then
  ensure_dir "$(dirname "$COOKIE_JAR")"
  login_headers="$COOKIE_JAR.headers"
  log "+ curl -fsS -D $login_headers -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o /dev/null"
  curl -fsS -D "$login_headers" -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
    -H 'content-type: application/json' \
    --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
    -o /dev/null
  if [[ "$BASE_URL" == http://* ]] && grep -i '^Set-Cookie:.*Secure' "$login_headers" >/dev/null; then
    log "HTTP login response must not set Secure cookies"
    exit 1
  fi
  run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/overview"
  rm -f "$COOKIE_JAR" "$login_headers"
fi
log "smoke tests passed"
