#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
FLUXGATE_HOST_BIND="${FLUXGATE_HOST_BIND:-}"
FLUXGATE_HTTP_PORT="${FLUXGATE_HTTP_PORT:-}"
PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-}"

if [[ -z "$FLUXGATE_HOST_BIND" && -z "$FLUXGATE_HTTP_PORT" && -z "$PUBLIC_BASE_URL" ]]; then
  log "no access settings supplied; set FLUXGATE_HOST_BIND, FLUXGATE_HTTP_PORT or PUBLIC_BASE_URL"
  exit 64
fi

if [[ -n "$FLUXGATE_HOST_BIND" ]]; then
  case "$FLUXGATE_HOST_BIND" in
    127.0.0.1|0.0.0.0|localhost) ;;
    *)
      if [[ ! "$FLUXGATE_HOST_BIND" =~ ^[0-9]{1,3}(\.[0-9]{1,3}){3}$ ]]; then
        log "invalid FLUXGATE_HOST_BIND: $FLUXGATE_HOST_BIND"
        exit 64
      fi
      ;;
  esac
fi

if [[ -n "$FLUXGATE_HTTP_PORT" ]]; then
  if [[ ! "$FLUXGATE_HTTP_PORT" =~ ^[0-9]+$ || "$FLUXGATE_HTTP_PORT" -lt 1 || "$FLUXGATE_HTTP_PORT" -gt 65535 ]]; then
    log "invalid FLUXGATE_HTTP_PORT: $FLUXGATE_HTTP_PORT"
    exit 64
  fi
fi

ensure_dir "$(dirname "$ENV_FILE")"
if [[ ! -f "$ENV_FILE" ]]; then
  if [[ -f "$ROOT_DIR/.env.example" ]]; then
    run_logged cp "$ROOT_DIR/.env.example" "$ENV_FILE"
  else
    : >"$ENV_FILE"
  fi
fi

upsert_env() {
  local key="$1"
  local value="$2"
  local tmp_file="$ENV_FILE.tmp"

  [[ -z "$value" ]] && return 0

  awk -v key="$key" -v value="$value" '
    BEGIN { found = 0 }
    $0 ~ "^" key "=" {
      print key "=" value
      found = 1
      next
    }
    { print }
    END {
      if (found == 0) {
        print key "=" value
      }
    }
  ' "$ENV_FILE" >"$tmp_file"
  mv "$tmp_file" "$ENV_FILE"
}

upsert_env FLUXGATE_HOST_BIND "$FLUXGATE_HOST_BIND"
upsert_env FLUXGATE_HTTP_PORT "$FLUXGATE_HTTP_PORT"
upsert_env PUBLIC_BASE_URL "$PUBLIC_BASE_URL"

log "access config updated: env=$ENV_FILE"
if [[ -n "$FLUXGATE_HOST_BIND" ]]; then
  log "FLUXGATE_HOST_BIND=$FLUXGATE_HOST_BIND"
fi
if [[ -n "$FLUXGATE_HTTP_PORT" ]]; then
  log "FLUXGATE_HTTP_PORT=$FLUXGATE_HTTP_PORT"
fi
if [[ -n "$PUBLIC_BASE_URL" ]]; then
  log "PUBLIC_BASE_URL configured"
fi
