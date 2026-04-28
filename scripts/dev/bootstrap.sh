#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "bootstrapping FluxGate development environment"
require_cmd go

cd "$ROOT_DIR"
run_logged go mod tidy
ensure_dir "$ROOT_DIR/data/sing-box"
ensure_dir "$ROOT_DIR/logs/fluxgate"
ensure_dir "$ROOT_DIR/logs/qa/screenshots"
log "bootstrap complete"
