#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

CONTAINER_NAME="${SING_BOX_CONTAINER_NAME:-fluxgate-sing-box}"

run_logged "$ROOT_DIR/scripts/sing-box/check-config.sh"
require_cmd docker
run_logged docker restart "$CONTAINER_NAME"
log "sing-box config reloaded"
