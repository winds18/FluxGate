#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "running Sub-Store extraction QA"
setup_go_cache
run_logged go test ./internal/substore ./internal/upstreamsync ./internal/httpapi -run 'TestExtractorExtractsNormalizedNodes|TestRefreshSourceUsesSubStoreExtraction|TestRefreshSourceAPIUsesSubStoreExtraction' -count=1
log "Sub-Store extraction QA passed"
