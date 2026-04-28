#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_CONFIG="${DEPLOY_CONFIG:-$ROOT_DIR/.env.deploy.local}"
if [[ -f "$DEPLOY_CONFIG" ]]; then
  # shellcheck disable=SC1090
  source "$DEPLOY_CONFIG"
fi

REMOTE_HOST="${REMOTE_HOST:-}"
REMOTE_DIR="${REMOTE_DIR:-}"
REMOTE_URL="${REMOTE_URL:-$(git -C "$ROOT_DIR" config --get remote.origin.url || true)}"
REMOTE_CLONE_URL="${REMOTE_CLONE_URL:-}"
REMOTE_FLUXGATE_HOST_BIND="${REMOTE_FLUXGATE_HOST_BIND:-}"
REMOTE_FLUXGATE_HTTP_PORT="${REMOTE_FLUXGATE_HTTP_PORT:-}"
REMOTE_PUBLIC_BASE_URL="${REMOTE_PUBLIC_BASE_URL:-}"
REMOTE_ADMIN_BOOTSTRAP_USERNAME="${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}"
REMOTE_ADMIN_BOOTSTRAP_PASSWORD="${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}"
VERIFY_BASE_URL="${VERIFY_BASE_URL:-http://127.0.0.1:${REMOTE_FLUXGATE_HTTP_PORT:-18080}}"
BRANCH_NAME="${BRANCH_NAME:-$(git -C "$ROOT_DIR" rev-parse --abbrev-ref HEAD)}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(date -u +%Y%m%d-%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"

if [[ -z "$REMOTE_CLONE_URL" ]]; then
  case "$REMOTE_URL" in
    git@github.com:*) REMOTE_CLONE_URL="https://github.com/${REMOTE_URL#git@github.com:}" ;;
    *) REMOTE_CLONE_URL="$REMOTE_URL" ;;
  esac
fi

if [[ -z "$REMOTE_HOST" || -z "$REMOTE_DIR" || -z "$REMOTE_URL" || -z "$REMOTE_CLONE_URL" ]]; then
  log "missing deploy config: set REMOTE_HOST, REMOTE_DIR, REMOTE_URL and REMOTE_CLONE_URL via environment or $DEPLOY_CONFIG"
  exit 64
fi

cd "$ROOT_DIR"

log "push and deploy started: branch=$BRANCH_NAME deploy_id=$DEPLOY_ID"
"$ROOT_DIR/scripts/qa/public-scan.sh"
run_logged git push -u origin "$BRANCH_NAME"

ssh "$REMOTE_HOST" \
  "REMOTE_DIR='$REMOTE_DIR' REMOTE_CLONE_URL='$REMOTE_CLONE_URL' BRANCH_NAME='$BRANCH_NAME' DEPLOY_ID='$DEPLOY_ID' REMOTE_FLUXGATE_HOST_BIND='$REMOTE_FLUXGATE_HOST_BIND' REMOTE_FLUXGATE_HTTP_PORT='$REMOTE_FLUXGATE_HTTP_PORT' REMOTE_PUBLIC_BASE_URL='$REMOTE_PUBLIC_BASE_URL' REMOTE_ADMIN_BOOTSTRAP_USERNAME='$REMOTE_ADMIN_BOOTSTRAP_USERNAME' REMOTE_ADMIN_BOOTSTRAP_PASSWORD='$REMOTE_ADMIN_BOOTSTRAP_PASSWORD' VERIFY_BASE_URL='$VERIFY_BASE_URL' bash -s" <<'REMOTE'
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
if [[ -n "${REMOTE_FLUXGATE_HOST_BIND:-}" || -n "${REMOTE_FLUXGATE_HTTP_PORT:-}" || -n "${REMOTE_PUBLIC_BASE_URL:-}" || -n "${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}" || -n "${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}" ]]; then
  FLUXGATE_HOST_BIND="${REMOTE_FLUXGATE_HOST_BIND:-}" \
    FLUXGATE_HTTP_PORT="${REMOTE_FLUXGATE_HTTP_PORT:-}" \
    PUBLIC_BASE_URL="${REMOTE_PUBLIC_BASE_URL:-}" \
    ADMIN_BOOTSTRAP_USERNAME="${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}" \
    ADMIN_BOOTSTRAP_PASSWORD="${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}" \
    scripts/deploy/configure-access.sh
fi
DEPLOY_ID="$DEPLOY_ID" REMOTE_BUILD_ENABLED=true scripts/deploy/deploy-remote.sh
PUBLIC_BASE_URL="$VERIFY_BASE_URL" DEPLOY_ID="$DEPLOY_ID" scripts/deploy/verify-remote.sh
REMOTE

log "push and deploy finished: deploy_id=$DEPLOY_ID"
