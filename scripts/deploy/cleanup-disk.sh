#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

WARN_PERCENT="${DISK_WARN_PERCENT:-80}"
CRITICAL_PERCENT="${DISK_CRITICAL_PERCENT:-90}"
DRY_RUN="${DRY_RUN:-true}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-14}"
DEPLOY_ID="${DEPLOY_ID:-deploy-$(timestamp)}"
LOG_DIR="$ROOT_DIR/logs/deploy"
ensure_dir "$LOG_DIR"
LOG_FILE="$LOG_DIR/cleanup-${DEPLOY_ID}.log"

exec > >(tee "$LOG_FILE") 2>&1

usage_percent() {
  df -P "$ROOT_DIR" | awk 'NR==2 {gsub("%", "", $5); print $5}'
}

run_or_show() {
  if [[ "$DRY_RUN" == "true" ]]; then
    log "[dry-run] $*"
  else
    run_logged "$@"
  fi
}

current_usage="$(usage_percent)"
log "disk usage: ${current_usage}%"

if (( current_usage < WARN_PERCENT )); then
  log "disk below warning threshold, no cleanup needed"
  exit 0
fi

log "disk reached warning threshold, starting safe cleanup"

if command -v docker >/dev/null 2>&1; then
  run_or_show docker image prune -f
  run_or_show docker container prune -f
  run_or_show docker builder prune -f --filter "until=168h"
fi

if [[ -d "$ROOT_DIR/logs" ]]; then
  if [[ "$DRY_RUN" == "true" ]]; then
    find "$ROOT_DIR/logs" -type f -mtime +"$RETENTION_DAYS" -print
  else
    find "$ROOT_DIR/logs" -type f -mtime +"$RETENTION_DAYS" -delete
  fi
fi

if [[ -d "$ROOT_DIR/data/backups" ]]; then
  if [[ "$DRY_RUN" == "true" ]]; then
    find "$ROOT_DIR/data/backups" -type f -mtime +"$RETENTION_DAYS" -print
  else
    find "$ROOT_DIR/data/backups" -type f -mtime +"$RETENTION_DAYS" -delete
  fi
fi

new_usage="$(usage_percent)"
log "disk usage after cleanup: ${new_usage}%"

if (( new_usage >= CRITICAL_PERCENT )); then
  log "disk remains above critical threshold"
  exit 2
fi

log "cleanup finished: $LOG_FILE"
