#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/probe-${DEPLOY_ID}.log"

exec > >(tee "$LOG_FILE") 2>&1

log "probe started: $DEPLOY_ID"
log "root: $ROOT_DIR"
log "os: $(uname -a)"
log "arch: $(uname -m)"

if command -v docker >/dev/null 2>&1; then
  log "docker: $(docker --version)"
  docker ps --format 'container={{.Names}} image={{.Image}} status={{.Status}}' || true
else
  log "docker: missing"
fi

if docker compose version >/dev/null 2>&1; then
  log "docker compose: $(docker compose version)"
else
  log "docker compose: missing"
fi

log "memory:"
if command -v free >/dev/null 2>&1; then
  free -h
else
  vm_stat || true
fi

log "disk:"
df -h "$ROOT_DIR" || df -h .
df -i "$ROOT_DIR" || true

for port in 8080 18080 8443 8444 8445; do
  if command -v lsof >/dev/null 2>&1; then
    if lsof -i TCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      log "port $port: occupied"
    else
      log "port $port: free"
    fi
  fi
done

if [[ -f "$ROOT_DIR/.env" ]]; then
  log ".env: present"
  sed -n 's/^\([^#=][^=]*\)=.*/env_key=\1/p' "$ROOT_DIR/.env" | sort
else
  log ".env: missing"
fi

if [[ -f "$ROOT_DIR/data/sing-box/config.json" ]]; then
  log "sing-box config: present"
else
  log "sing-box config: missing"
fi

log "probe finished: $LOG_FILE"
