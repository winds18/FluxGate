#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

REMOTE_HOST="${REMOTE_HOST:-66.10}"
REMOTE_DIR="${REMOTE_DIR:-/home/wings/docker/FluxGate}"
REMOTE_URL="${REMOTE_URL:-git@github.com:winds18/FluxGate.git}"
REMOTE_CLONE_URL="${REMOTE_CLONE_URL:-https://github.com/winds18/FluxGate.git}"
BRANCH_NAME="${BRANCH_NAME:-$(git -C "$ROOT_DIR" rev-parse --abbrev-ref HEAD)}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(date -u +%Y%m%d-%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"

cd "$ROOT_DIR"

log "push and deploy started: branch=$BRANCH_NAME deploy_id=$DEPLOY_ID"
run_logged git push -u origin "$BRANCH_NAME"

ssh "$REMOTE_HOST" \
  "REMOTE_DIR='$REMOTE_DIR' REMOTE_CLONE_URL='$REMOTE_CLONE_URL' BRANCH_NAME='$BRANCH_NAME' DEPLOY_ID='$DEPLOY_ID' bash -s" <<'REMOTE'
set -euo pipefail

if [[ ! -d "$REMOTE_DIR/.git" ]]; then
  mkdir -p "$(dirname "$REMOTE_DIR")"
  if [[ -e "$REMOTE_DIR" ]]; then
    if [[ -n "$(find "$REMOTE_DIR" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]]; then
      mv "$REMOTE_DIR" "${REMOTE_DIR}.bak.$(date -u +%Y%m%d-%H%M%S)"
    else
      rmdir "$REMOTE_DIR"
    fi
  fi
  git clone --branch "$BRANCH_NAME" "$REMOTE_CLONE_URL" "$REMOTE_DIR"
fi

cd "$REMOTE_DIR"
git fetch origin "$BRANCH_NAME"
git switch "$BRANCH_NAME"
git reset --hard "origin/$BRANCH_NAME"

scripts/deploy/bootstrap-remote.sh
DEPLOY_ID="$DEPLOY_ID" REMOTE_BUILD_ENABLED=true scripts/deploy/deploy-remote.sh
PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-http://127.0.0.1:8080}" DEPLOY_ID="$DEPLOY_ID" scripts/deploy/verify-remote.sh
REMOTE

log "push and deploy finished: deploy_id=$DEPLOY_ID"
