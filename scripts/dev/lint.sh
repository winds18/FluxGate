#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "running Go vet"
cd "$ROOT_DIR"
run_logged go vet ./...
