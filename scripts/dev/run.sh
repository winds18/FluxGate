#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "starting FluxGate locally"
cd "$ROOT_DIR"
export APP_ENV="${APP_ENV:-development}"
export HTTP_ADDR="${HTTP_ADDR:-127.0.0.1:8080}"
export DB_PATH="${DB_PATH:-$ROOT_DIR/data/fluxgate.db}"
export LOG_DIR="${LOG_DIR:-$ROOT_DIR/logs/fluxgate}"
export ADMIN_BOOTSTRAP_USERNAME="${ADMIN_BOOTSTRAP_USERNAME:-admin}"
export ADMIN_BOOTSTRAP_PASSWORD="${ADMIN_BOOTSTRAP_PASSWORD:-dev-admin-change-me}"
export SESSION_SECRET="${SESSION_SECRET:-dev-session-secret-change-me}"
setup_go_cache
run_logged go run ./cmd/fluxgate
