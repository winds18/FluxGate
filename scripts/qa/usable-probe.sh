#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_CONFIG="${DEPLOY_CONFIG:-$ROOT_DIR/.env.deploy.local}"
if [[ -f "$DEPLOY_CONFIG" ]]; then
  # shellcheck disable=SC1090
  source "$DEPLOY_CONFIG"
fi

BASE_URL="${1:-${BASE_URL:-${REMOTE_BROWSER_BASE_URL:-${BROWSER_BASE_URL:-http://127.0.0.1:8080}}}}"
ADMIN_USERNAME="${ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}}}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/usable-probe/$(timestamp)}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
COOKIE_JAR="$OUT_DIR/cookies.txt"

require_cmd curl
require_cmd node
require_cmd nc
ensure_dir "$OUT_DIR"

cleanup() {
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

if [[ -z "$ADMIN_USERNAME" || -z "$ADMIN_PASSWORD" ]]; then
  log "ADMIN_USERNAME and ADMIN_PASSWORD are required for usable probe"
  exit 64
fi

json_expr() {
  local file="$1"
  local expr="$2"
  node -e '
const fs = require("fs");
const file = process.argv[1];
const expr = process.argv[2];
const data = JSON.parse(fs.readFileSync(file, "utf8"));
const value = Function("data", `return (${expr});`)(data);
if (value === undefined || value === null) {
  process.exit(2);
}
if (typeof value === "object") {
  console.log(JSON.stringify(value));
} else {
  console.log(value);
}
' "$file" "$expr"
}

assert_min() {
  local value="$1"
  local min="$2"
  local label="$3"
  if (( value < min )); then
    log "usable probe failed: $label expected >= $min, got $value"
    exit 1
  fi
}

probe_subscription_url() {
  local url="$1"
  local target="$2"
  local out="$3"
  local headers="$out.headers"
  local probe_url="$url"
  if [[ "$target" != "default" ]]; then
    if [[ "$probe_url" == *\?* ]]; then
      probe_url="$probe_url&target=$target"
    else
      probe_url="$probe_url?target=$target"
    fi
  fi
  log "+ curl -fsS -D $headers <subscription:$target:redacted> -o $out"
  curl -fsS -D "$headers" "$probe_url" -o "$out"
  if ! tr -d '\r' <"$headers" | grep -Eiq '^subscription-userinfo:'; then
    log "usable probe failed: $target subscription missing subscription-userinfo header"
    exit 1
  fi
}

log "running usable probe against $BASE_URL"
curl -fsS "$BASE_URL/healthz" -o "$OUT_DIR/healthz.json"
curl -fsS "$BASE_URL/readyz" -o "$OUT_DIR/readyz.json"

log "+ curl -fsS -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o $OUT_DIR/login.json"
curl -fsS -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
  -H 'content-type: application/json' \
  --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -o "$OUT_DIR/login.json"

curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/overview" -o "$OUT_DIR/overview.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/virtual-nodes" -o "$OUT_DIR/virtual-nodes.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/policies" -o "$OUT_DIR/policies.json"
curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/check" -o "$OUT_DIR/sing-box-check.json"

teams="$(json_expr "$OUT_DIR/overview.json" "data.teams || 0")"
users="$(json_expr "$OUT_DIR/overview.json" "data.users || 0")"
tokens="$(json_expr "$OUT_DIR/overview.json" "data.tokens || 0")"
sources="$(json_expr "$OUT_DIR/overview.json" "data.sources || 0")"
nodes="$(json_expr "$OUT_DIR/overview.json" "data.nodes || 0")"
virtual_nodes="$(json_expr "$OUT_DIR/overview.json" "data.virtual_nodes || 0")"
policies="$(json_expr "$OUT_DIR/overview.json" "data.policies || 0")"

assert_min "$teams" 1 "teams"
assert_min "$users" 1 "users"
assert_min "$tokens" 1 "tokens"
assert_min "$sources" 1 "sources"
assert_min "$nodes" 1 "nodes"
assert_min "$virtual_nodes" 1 "virtual_nodes"
assert_min "$policies" 1 "policies"

config_valid="$(json_expr "$OUT_DIR/sing-box-check.json" "data.valid === true")"
inbound_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.inbound_count || 0")"
user_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.user_count || 0")"
upstream_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.upstream_outbound_count || 0")"
config_hash="$(json_expr "$OUT_DIR/sing-box-check.json" "data.config_hash || ''")"
if [[ "$config_valid" != "true" ]]; then
  log "usable probe failed: sing-box config is not valid"
  exit 1
fi
assert_min "$inbound_count" 1 "sing-box inbound_count"
assert_min "$user_count" 1 "sing-box user_count"
assert_min "$upstream_count" 1 "sing-box upstream_outbound_count"

virtual_node_name="$(json_expr "$OUT_DIR/virtual-nodes.json" "data.find((item) => item.status === 'active')?.name || ''")"
virtual_node_port="$(json_expr "$OUT_DIR/virtual-nodes.json" "data.find((item) => item.status === 'active')?.listen_port || 0")"
if [[ -z "$virtual_node_name" || "$virtual_node_port" == "0" ]]; then
  log "usable probe failed: no active virtual node with listen port"
  exit 1
fi

gateway_host="${GATEWAY_PROBE_HOST:-$(node -e 'const raw = process.argv[1]; console.log(new URL(raw).hostname);' "$BASE_URL")}"
gateway_port="${GATEWAY_PROBE_PORT:-$virtual_node_port}"
log "+ nc -z -w 3 $gateway_host $gateway_port"
nc -z -w 3 "$gateway_host" "$gateway_port"

if [[ -n "${FLUXGATE_QA_SUBSCRIPTION_URL:-}" ]]; then
  probe_subscription_url "$FLUXGATE_QA_SUBSCRIPTION_URL" "default" "$OUT_DIR/subscription-default.txt"
  probe_subscription_url "$FLUXGATE_QA_SUBSCRIPTION_URL" "clash" "$OUT_DIR/subscription-clash.yaml"
  probe_subscription_url "$FLUXGATE_QA_SUBSCRIPTION_URL" "sing-box" "$OUT_DIR/subscription-sing-box.json"
  if ! grep -F "$virtual_node_name" "$OUT_DIR/subscription-clash.yaml" >/dev/null; then
    log "usable probe failed: clash subscription missing active virtual node $virtual_node_name"
    exit 1
  fi
  if ! node -e '
const fs = require("fs");
const file = process.argv[1];
const name = process.argv[2];
const data = JSON.parse(fs.readFileSync(file, "utf8"));
const outbounds = Array.isArray(data.outbounds) ? data.outbounds : [];
process.exit(outbounds.some((item) => item && item.tag === name) ? 0 : 1);
' "$OUT_DIR/subscription-sing-box.json" "$virtual_node_name"; then
    log "usable probe failed: sing-box subscription missing active virtual node $virtual_node_name"
    exit 1
  fi
  subscription_probe="enabled"
else
  subscription_probe="skipped"
  log "FLUXGATE_QA_SUBSCRIPTION_URL is not set; skipping direct subscription body probe"
fi

log "usable probe passed: teams=$teams users=$users tokens=$tokens sources=$sources nodes=$nodes virtual_nodes=$virtual_nodes policies=$policies inbound=$inbound_count users_in_config=$user_count upstreams=$upstream_count gateway=$gateway_host:$gateway_port virtual_node=$virtual_node_name config_hash=$config_hash subscription_probe=$subscription_probe"
if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "usable probe artifacts: $OUT_DIR"
else
  log "usable probe finished; temporary artifacts will be removed"
fi
