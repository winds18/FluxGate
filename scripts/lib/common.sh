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

setup_go_cache() {
  export GOCACHE="${GOCACHE:-$ROOT_DIR/tmp/go-build-cache}"
  ensure_dir "$GOCACHE"
}

run_logged() {
  log "+ $*"
  "$@"
}

add_no_proxy_host() {
  local host="$1"
  if [[ -z "$host" ]]; then
    return
  fi

  local current="${NO_PROXY:-${no_proxy:-}}"
  if [[ ",$current," != *",$host,"* ]]; then
    current="${current:+$current,}$host"
  fi

  export NO_PROXY="$current"
  export no_proxy="$current"
}

configure_no_proxy_for_url() {
  local url="$1"
  local host="${url#*://}"
  host="${host%%/*}"
  if [[ "$host" == \[*\]* ]]; then
    host="${host#\[}"
    host="${host%%\]*}"
  else
    host="${host%%:*}"
  fi

  add_no_proxy_host "$host"
  add_no_proxy_host "127.0.0.1"
  add_no_proxy_host "localhost"
  add_no_proxy_host "::1"
}
