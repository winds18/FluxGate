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
      "$ROOT_DIR/data/qa-fluxgate.db-wal"

rm -rf "$ROOT_DIR/tmp/build"
rm -f "$ROOT_DIR/tmp/qa-server.log"
find "$ROOT_DIR/tmp" -type d -empty -delete 2>/dev/null || true
rm -rf "$ROOT_DIR/logs/fluxgate" "$ROOT_DIR/logs/fluxgate-qa"

if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
  rm -rf "$ROOT_DIR/logs/qa"
  rm -rf "$ROOT_DIR/logs/deploy"
  rm -rf "$ROOT_DIR/logs/diagnostics"
fi

find "$ROOT_DIR" -name .DS_Store -type f -delete

log "QA temporary artifacts cleaned"
