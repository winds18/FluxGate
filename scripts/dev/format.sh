#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "formatting Go files"
cd "$ROOT_DIR"
run_logged gofmt -w ./cmd ./internal
