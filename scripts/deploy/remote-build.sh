#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

IMAGE="${FLUXGATE_IMAGE:-ghcr.io/winds18/fluxgate:${FLUXGATE_IMAGE_TAG:-latest}}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/remote-build-${DEPLOY_ID}.log"

exec > >(tee "$LOG_FILE") 2>&1

require_cmd docker
log "building image remotely: $IMAGE"
cd "$ROOT_DIR"
run_logged docker build --build-arg "GOPROXY=${GOPROXY:-https://goproxy.cn,https://proxy.golang.org,direct}" -t "$IMAGE" .
log "remote build complete: $IMAGE"
