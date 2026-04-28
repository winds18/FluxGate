#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

URL="${URL:-http://127.0.0.1:8080}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/screenshots/$(timestamp)}"
ensure_dir "$OUT_DIR"

if command -v npx >/dev/null 2>&1; then
  log "capturing screenshot with Playwright: $URL"
  if ! npx --yes playwright screenshot --full-page "$URL" "$OUT_DIR/fluxgate.png"; then
    log "Playwright browser missing or unavailable; installing Chromium and retrying"
    run_logged npx --yes playwright install chromium
    run_logged npx --yes playwright screenshot --full-page "$URL" "$OUT_DIR/fluxgate.png"
  fi
  log "screenshot saved: $OUT_DIR/fluxgate.png"
else
  log "npx missing, screenshot skipped"
  exit 1
fi
