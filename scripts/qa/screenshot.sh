#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

URL="${URL:-http://127.0.0.1:8080}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/screenshots/$(timestamp)}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
SCREENSHOT_VIEWPORTS="${SCREENSHOT_VIEWPORTS:-1440x900 1280x720 390x844}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --keep)
      KEEP_ARTIFACTS="true"
      shift
      ;;
    --out-dir)
      OUT_DIR="${2:-}"
      if [[ -z "$OUT_DIR" ]]; then
        log "--out-dir requires a value"
        exit 64
      fi
      shift 2
      ;;
    *)
      URL="$1"
      shift
      ;;
  esac
done

ensure_dir "$OUT_DIR"

read -r -a VIEWPORTS <<< "$SCREENSHOT_VIEWPORTS"
if [[ "${#VIEWPORTS[@]}" -eq 0 ]]; then
  log "SCREENSHOT_VIEWPORTS must include at least one WIDTHxHEIGHT value"
  exit 64
fi

cleanup() {
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

capture_viewport() {
  local viewport="$1"
  if [[ ! "$viewport" =~ ^([0-9]+)x([0-9]+)$ ]]; then
    log "invalid screenshot viewport: $viewport"
    return 64
  fi

  local playwright_viewport="${BASH_REMATCH[1]},${BASH_REMATCH[2]}"
  local output_path="$OUT_DIR/fluxgate-${viewport}.png"
  log "capturing screenshot with Playwright: $URL viewport=$viewport"
  npx --yes playwright screenshot --full-page --viewport-size "$playwright_viewport" "$URL" "$output_path"
}

capture_matrix() {
  local viewport
  for viewport in "${VIEWPORTS[@]}"; do
    capture_viewport "$viewport" || return $?
  done
}

if command -v npx >/dev/null 2>&1; then
  if ! capture_matrix; then
    log "Playwright browser missing or unavailable; installing Chromium and retrying"
    run_logged npx --yes playwright install chromium
    capture_matrix
  fi
  if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
    for viewport in "${VIEWPORTS[@]}"; do
      log "screenshot saved: $OUT_DIR/fluxgate-${viewport}.png"
    done
  else
    log "screenshot matrix captured; temporary artifacts will be removed"
  fi
else
  log "npx missing, screenshot skipped"
  exit 1
fi
