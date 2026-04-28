#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
require_cmd curl

log "running smoke tests against $BASE_URL"
run_logged curl -fsS "$BASE_URL/healthz"
run_logged curl -fsS "$BASE_URL/readyz"
run_logged curl -fsS "$BASE_URL/api/overview"
log "smoke tests passed"
