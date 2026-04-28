#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "running database migrations through FluxGate startup check"
cd "$ROOT_DIR"
export APP_ENV="${APP_ENV:-migration}"
export DB_PATH="${DB_PATH:-$ROOT_DIR/data/fluxgate.db}"
run_logged go run ./cmd/fluxgate migrate
log "migration command finished"
