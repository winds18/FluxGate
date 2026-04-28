#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "running Go test suite"
cd "$ROOT_DIR"
run_logged go test ./...
