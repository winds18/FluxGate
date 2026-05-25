#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

IMAGE="${FLUXGATE_IMAGE:-ghcr.io/winds18/fluxgate:${FLUXGATE_IMAGE_TAG:-latest}}"
REMOTE_HOST="${REMOTE_HOST:-}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
TARGET_GOOS="${TARGET_GOOS:-linux}"
TARGET_GOARCH="${TARGET_GOARCH:-amd64}"
TARGET_PLATFORM="${TARGET_PLATFORM:-$TARGET_GOOS/$TARGET_GOARCH}"
CLEANUP_IMAGE="${CLEANUP_IMAGE:-true}"
LOAD_REMOTE="${LOAD_REMOTE:-true}"
WORK_DIR="$ROOT_DIR/tmp/prebuilt-image/$DEPLOY_ID"

if [[ "$LOAD_REMOTE" == "true" && -z "$REMOTE_HOST" ]]; then
  log "REMOTE_HOST is required to load the prebuilt image"
  exit 64
fi

cleanup() {
  rm -rf "$WORK_DIR"
  if [[ "$CLEANUP_IMAGE" == "true" ]]; then
    docker image rm "$IMAGE" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

find_ca_bundle() {
  local candidates=(
    "/etc/ssl/certs/ca-certificates.crt"
    "/etc/ssl/cert.pem"
    "/usr/local/etc/openssl@3/cert.pem"
    "/opt/homebrew/etc/openssl@3/cert.pem"
  )
  for candidate in "${candidates[@]}"; do
    if [[ -s "$candidate" ]]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done
  return 1
}

require_cmd go
require_cmd docker
require_cmd ssh
ensure_dir "$WORK_DIR"

ca_bundle="$(find_ca_bundle || true)"
if [[ -z "$ca_bundle" ]]; then
  log "no local CA bundle found for prebuilt image"
  exit 1
fi

log "building Linux binary for image: os=$TARGET_GOOS arch=$TARGET_GOARCH"
cd "$ROOT_DIR"
GOOS="$TARGET_GOOS" GOARCH="$TARGET_GOARCH" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$WORK_DIR/fluxgate" ./cmd/fluxgate
cp "$ca_bundle" "$WORK_DIR/ca-certificates.crt"

log "building prebuilt Docker image locally: $IMAGE platform=$TARGET_PLATFORM"
run_logged docker build --platform "$TARGET_PLATFORM" -f "$ROOT_DIR/Dockerfile.prebuilt" -t "$IMAGE" "$WORK_DIR"

if [[ "$LOAD_REMOTE" == "true" ]]; then
  log "loading prebuilt Docker image on remote host: $IMAGE"
  docker save "$IMAGE" | ssh "$REMOTE_HOST" docker load
  log "prebuilt image loaded on remote host: $IMAGE"
else
  log "remote image load skipped: $IMAGE"
fi
