#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

PORT="${PORT:-8080}"

if ! command -v lsof >/dev/null 2>&1; then
  log "lsof missing, cannot stop by port"
  exit 0
fi

pids="$(lsof -ti TCP:"$PORT" -sTCP:LISTEN || true)"
if [[ -z "$pids" ]]; then
  log "no local FluxGate process listening on port $PORT"
  exit 0
fi

for pid in $pids; do
  log "stopping process on port $PORT: pid=$pid"
  kill "$pid"
done
