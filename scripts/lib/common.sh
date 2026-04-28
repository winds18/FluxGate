#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_ROOT="${LOG_ROOT:-$ROOT_DIR/logs}"

timestamp() {
  date -u +"%Y%m%d-%H%M%S"
}

log() {
  printf '[%s] %s\n' "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$*"
}

ensure_dir() {
  mkdir -p "$1"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log "missing command: $1"
    exit 127
  fi
}

run_logged() {
  log "+ $*"
  "$@"
}
