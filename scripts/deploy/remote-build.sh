#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

IMAGE="${FLUXGATE_IMAGE:-ghcr.io/winds18/fluxgate:${FLUXGATE_IMAGE_TAG:-latest}}"
SING_BOX_IMAGE="${SING_BOX_IMAGE:-fluxgate-sing-box:${FLUXGATE_IMAGE_TAG:-latest}}"
SING_BOX_CUSTOM_BUILD_ENABLED="${SING_BOX_CUSTOM_BUILD_ENABLED:-true}"
SING_BOX_VERSION="${SING_BOX_VERSION:-latest}"
SING_BOX_BUILD_TAGS="${SING_BOX_BUILD_TAGS:-with_v2ray_api with_quic with_grpc with_wireguard with_utls with_acme with_clash_api}"
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

if [[ "$SING_BOX_CUSTOM_BUILD_ENABLED" == "true" ]]; then
  log "building sing-box image remotely: $SING_BOX_IMAGE"
  run_logged docker build \
    --build-arg "GOPROXY=${GOPROXY:-https://goproxy.cn,https://proxy.golang.org,direct}" \
    --build-arg "SING_BOX_VERSION=$SING_BOX_VERSION" \
    --build-arg "SING_BOX_BUILD_TAGS=$SING_BOX_BUILD_TAGS" \
    -f Dockerfile.sing-box \
    -t "$SING_BOX_IMAGE" .
  log "remote sing-box build complete: $SING_BOX_IMAGE"
fi
