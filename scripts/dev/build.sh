#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

OUT_DIR="${OUT_DIR:-$ROOT_DIR/tmp/build}"
ensure_dir "$OUT_DIR"

log "building FluxGate binary"
cd "$ROOT_DIR"
run_logged go build -trimpath -o "$OUT_DIR/fluxgate" ./cmd/fluxgate
log "binary built: $OUT_DIR/fluxgate"
