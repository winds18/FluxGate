#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

IMAGE="${IMAGE:-fluxgate:local-qa}"
CLEANUP_IMAGE="${CLEANUP_IMAGE:-true}"

cleanup() {
  if [[ "$CLEANUP_IMAGE" == "true" ]]; then
    docker image rm "$IMAGE" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

require_cmd docker
log "building FluxGate Docker image: $IMAGE"
cd "$ROOT_DIR"
run_logged docker build --build-arg "GOPROXY=${GOPROXY:-https://goproxy.cn,https://proxy.golang.org,direct}" -t "$IMAGE" .
log "Docker image build passed: $IMAGE"
