#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

OUT_DIR="${OUT_DIR:-$ROOT_DIR/tmp/build}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
ensure_dir "$OUT_DIR"

cleanup() {
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

log "building FluxGate binary"
cd "$ROOT_DIR"
run_logged go build -trimpath -o "$OUT_DIR/fluxgate" ./cmd/fluxgate
if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "binary built: $OUT_DIR/fluxgate"
else
  log "binary build passed; temporary artifact will be removed"
fi
