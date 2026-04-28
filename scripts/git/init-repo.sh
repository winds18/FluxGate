#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

REMOTE_URL="${REMOTE_URL:-git@github.com:winds18/FluxGate.git}"
BRANCH_NAME="${BRANCH_NAME:-codex/phase-1-foundation}"

cd "$ROOT_DIR"

if [[ ! -d .git ]]; then
  run_logged git init
fi

if git remote get-url origin >/dev/null 2>&1; then
  current="$(git remote get-url origin)"
  if [[ "$current" != "$REMOTE_URL" ]]; then
    run_logged git remote set-url origin "$REMOTE_URL"
  fi
else
  run_logged git remote add origin "$REMOTE_URL"
fi

if git rev-parse --verify "$BRANCH_NAME" >/dev/null 2>&1; then
  run_logged git switch "$BRANCH_NAME"
else
  run_logged git switch -c "$BRANCH_NAME"
fi

log "git repository ready: branch=$BRANCH_NAME remote=$REMOTE_URL"
