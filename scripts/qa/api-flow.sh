#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/api-flow/$(timestamp)}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
ADMIN_USERNAME="${ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-}}"
COOKIE_JAR="$OUT_DIR/cookies.txt"
ensure_dir "$OUT_DIR"
require_cmd curl
require_cmd node

cleanup() {
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

json_value() {
  node -e "const fs=require('fs'); const data=JSON.parse(fs.readFileSync(0,'utf8')); console.log($1);"
}

post_json() {
  local path="$1"
  local payload="$2"
  local out="$3"
  run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL$path" -H 'content-type: application/json' --data "$payload" -o "$out"
}

log "running API flow against $BASE_URL"
if [[ -z "$ADMIN_USERNAME" || -z "$ADMIN_PASSWORD" ]]; then
  log "ADMIN_USERNAME and ADMIN_PASSWORD are required for API flow"
  exit 64
fi

log "+ curl -fsS -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o $OUT_DIR/login.json"
curl -fsS -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
  -H 'content-type: application/json' \
  --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -o "$OUT_DIR/login.json"

post_json "/api/teams" '{"name":"QA Team","description":"automated smoke"}' "$OUT_DIR/team.json"
team_id="$(json_value "data.id" <"$OUT_DIR/team.json")"

post_json "/api/users" "{\"team_id\":$team_id,\"name\":\"QA User\",\"email\":\"qa@example.test\"}" "$OUT_DIR/user.json"
user_id="$(json_value "data.id" <"$OUT_DIR/user.json")"

post_json "/api/sources" '{"name":"机场A","type":"manual"}' "$OUT_DIR/source-a.json"
source_a_id="$(json_value "data.id" <"$OUT_DIR/source-a.json")"
post_json "/api/sources" '{"name":"机场A","type":"manual"}' "$OUT_DIR/source-b.json"
source_b_prefix="$(json_value "JSON.stringify(data.display_prefix)" <"$OUT_DIR/source-b.json")"
post_json "/api/sources" '{"name":"订阅源A","type":"subscription","raw_content":"vless://uuid@example.net:443#新加坡%2001"}' "$OUT_DIR/source-subscription.json"
subscription_source_id="$(json_value "data.id" <"$OUT_DIR/source-subscription.json")"

post_json "/api/nodes/import" "{\"source_id\":$source_a_id,\"content\":\"vless://uuid@example.com:443#香港%2001\"}" "$OUT_DIR/import.json"
post_json "/api/sources/$subscription_source_id/refresh" '{}' "$OUT_DIR/source-refresh.json"
source_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-refresh.json")"
post_json "/api/virtual-nodes" '{"name":"FluxGate-HK","listen_protocol":"vless","listen_port":8443}' "$OUT_DIR/virtual-node.json"
post_json "/api/tokens" "{\"user_id\":$user_id,\"name\":\"QA Token\",\"expire_days\":30,\"quota_bytes\":1048576}" "$OUT_DIR/token.json"
plain_token="$(json_value "data.plain_token" <"$OUT_DIR/token.json")"

log "+ curl -fsS $BASE_URL/sub/<redacted>?target=clash -o $OUT_DIR/subscription.yaml"
curl -fsS "$BASE_URL/sub/$plain_token?target=clash" -o "$OUT_DIR/subscription.yaml"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/generate" -o "$OUT_DIR/sing-box.json"

if [[ "$source_b_prefix" != '"[机场A-2] "' ]]; then
  log "unexpected auto prefix for duplicate source: $source_b_prefix"
  exit 1
fi

if [[ "$source_refresh_imported" != "1" ]]; then
  log "unexpected subscription refresh import count: $source_refresh_imported"
  exit 1
fi

if ! grep -q 'FluxGate-HK' "$OUT_DIR/subscription.yaml"; then
  log "subscription missing FluxGate-HK"
  exit 1
fi

if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "API flow passed; artifacts: $OUT_DIR"
else
  log "API flow passed; temporary artifacts will be removed"
fi
