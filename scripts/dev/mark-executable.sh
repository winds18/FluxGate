#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

log "marking project scripts executable"
find "$ROOT_DIR/scripts" -type f -name '*.sh' -print0 | xargs -0 chmod +x
log "script permissions updated"
