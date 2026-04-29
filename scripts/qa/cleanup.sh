#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"

log "cleaning QA temporary artifacts"

rm -f "$ROOT_DIR/data/fluxgate.db" \
      "$ROOT_DIR/data/fluxgate.db-shm" \
      "$ROOT_DIR/data/fluxgate.db-wal" \
      "$ROOT_DIR/data/qa-fluxgate.db" \
      "$ROOT_DIR/data/qa-fluxgate.db-shm" \
      "$ROOT_DIR/data/qa-fluxgate.db-wal" \
      "$ROOT_DIR/data/sing-box/config.json" \
      "$ROOT_DIR/data/sing-box/config.previous.json"

rm -rf "$ROOT_DIR/logs/fluxgate" "$ROOT_DIR/logs/fluxgate-qa"

if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
  rm -rf "$ROOT_DIR/tmp"
  rm -rf "$ROOT_DIR/logs/qa"
  rm -rf "$ROOT_DIR/logs/deploy"
  rm -rf "$ROOT_DIR/logs/diagnostics"
  rm -rf "$ROOT_DIR/test-results"
  rm -rf "$ROOT_DIR/playwright-report"
else
  rm -rf "$ROOT_DIR/tmp/build"
  rm -rf "$ROOT_DIR/tmp/go-build-cache"
  rm -rf "$ROOT_DIR/tmp/playwright-browser-login-output"
  rm -f "$ROOT_DIR/tmp/qa-server.log"
  find "$ROOT_DIR/tmp" -type d -empty -delete 2>/dev/null || true
fi

ensure_dir "$ROOT_DIR/data"
ensure_dir "$ROOT_DIR/data/sing-box"
find "$ROOT_DIR" -name .DS_Store -type f -delete

log "QA temporary artifacts cleaned"
