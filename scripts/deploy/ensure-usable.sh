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
VIRTUAL_NODE_NAME="${VIRTUAL_NODE_NAME:-FluxGate-Gateway}"
VIRTUAL_NODE_PORT="${VIRTUAL_NODE_PORT:-8443}"
VIRTUAL_NODE_SELECTOR="${VIRTUAL_NODE_SELECTOR:-{}}"
POLICY_NAME="${POLICY_NAME:-FluxGate 默认可用策略}"
TOKEN_EXTEND_DAYS="${TOKEN_EXTEND_DAYS:-30}"
TOKEN_QUOTA_BYTES="${TOKEN_QUOTA_BYTES:-1073741824}"
PUBLISH_CONFIG="${PUBLISH_CONFIG:-true}"
RELOAD_SING_BOX="${RELOAD_SING_BOX:-true}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/deploy/ensure-usable/$(timestamp)}"
COOKIE_JAR="$OUT_DIR/cookies.txt"

require_cmd curl
require_cmd node
ensure_dir "$OUT_DIR"

cleanup() {
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

if [[ -z "$ADMIN_USERNAME" || -z "$ADMIN_PASSWORD" ]]; then
  log "ADMIN_USERNAME and ADMIN_PASSWORD are required for ensure-usable"
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
const list = (value) => Array.isArray(value) ? value : (Array.isArray(value?.data) ? value.data : []);
const value = Function("data", "list", `return (${expr});`)(data, list);
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

post_json() {
  local path="$1"
  local payload="$2"
  local out="$3"
  run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL$path" -H 'content-type: application/json' --data "$payload" -o "$out"
}

patch_json() {
  local path="$1"
  local payload="$2"
  local out="$3"
  run_logged curl -fsS -b "$COOKIE_JAR" -X PATCH "$BASE_URL$path" -H 'content-type: application/json' --data "$payload" -o "$out"
}

fetch_state() {
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/teams" -o "$OUT_DIR/teams.json"
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/users" -o "$OUT_DIR/users.json"
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/tokens" -o "$OUT_DIR/tokens.json"
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/nodes" -o "$OUT_DIR/nodes.json"
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/virtual-nodes" -o "$OUT_DIR/virtual-nodes.json"
  curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/policies" -o "$OUT_DIR/policies.json"
}

log "ensuring usable FluxGate delivery against $BASE_URL"
curl -fsS "$BASE_URL/healthz" -o "$OUT_DIR/healthz.json"
curl -fsS "$BASE_URL/readyz" -o "$OUT_DIR/readyz.json"

log "+ curl -fsS -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o $OUT_DIR/login.json"
curl -fsS -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
  -H 'content-type: application/json' \
  --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -o "$OUT_DIR/login.json"

fetch_state

team_id="$(json_expr "$OUT_DIR/teams.json" "list(data).find((row) => row.status === 'active')?.id || list(data)[0]?.id || 0")"
user_id="$(json_expr "$OUT_DIR/users.json" "list(data).find((row) => row.status === 'active')?.id || list(data)[0]?.id || 0")"
token_id="$(json_expr "$OUT_DIR/tokens.json" "(() => {
  const usable = list(data).find((row) =>
    row.status === 'active' &&
    row.gateway_account?.status === 'active' &&
    row.gateway_account?.protocol === 'vless' &&
    (!row.expire_at || Date.parse(row.expire_at) > Date.now()) &&
    (!(Number(row.quota_bytes || 0) > 0) || (Number(row.used_upload_bytes || 0) + Number(row.used_download_bytes || 0) < Number(row.quota_bytes || 0)))
  );
  return usable?.id || list(data).find((row) => row.status === 'active' && row.gateway_account?.status === 'active' && row.gateway_account?.protocol === 'vless')?.id || 0;
})()")"
active_nodes="$(json_expr "$OUT_DIR/nodes.json" "list(data).filter((row) => row.status === 'active').length")"

if [[ "$team_id" -lt 1 || "$user_id" -lt 1 || "$token_id" -lt 1 || "$active_nodes" -lt 1 ]]; then
  log "instance is not ready for automatic usability closeout: team_id=$team_id user_id=$user_id token_id=$token_id active_nodes=$active_nodes"
  log "create at least one team, user, active token, and upstream node before running this script"
  exit 1
fi

token_expired="$(TOKEN_ID="$token_id" json_expr "$OUT_DIR/tokens.json" "(() => {
  const row = list(data).find((item) => item.id === Number(process.env.TOKEN_ID));
  return Boolean(row?.expire_at && Date.parse(row.expire_at) <= Date.now());
})()")"
token_over_quota="$(TOKEN_ID="$token_id" json_expr "$OUT_DIR/tokens.json" "(() => {
  const row = list(data).find((item) => item.id === Number(process.env.TOKEN_ID));
  return Boolean(Number(row?.quota_bytes || 0) > 0 && Number(row?.used_upload_bytes || 0) + Number(row?.used_download_bytes || 0) >= Number(row?.quota_bytes || 0));
})()")"
if [[ "$token_expired" == "true" ]]; then
  post_json "/api/tokens/$token_id/extend" "{\"extend_days\":$TOKEN_EXTEND_DAYS}" "$OUT_DIR/token-extend.json"
  log "extended expired usable token: id=$token_id days=$TOKEN_EXTEND_DAYS"
fi
if [[ "$token_over_quota" == "true" ]]; then
  post_json "/api/tokens/$token_id/quota" "{\"quota_bytes\":$TOKEN_QUOTA_BYTES}" "$OUT_DIR/token-quota.json"
  log "added quota to over-quota usable token: id=$token_id bytes=$TOKEN_QUOTA_BYTES"
fi
if [[ "$token_expired" == "true" || "$token_over_quota" == "true" ]]; then
  fetch_state
fi
token_team_id="$(TOKEN_ID="$token_id" json_expr "$OUT_DIR/tokens.json" "list(data).find((row) => row.id === Number(process.env.TOKEN_ID))?.user_team_id || 0")"

virtual_node_id="$(json_expr "$OUT_DIR/virtual-nodes.json" "list(data).find((row) => row.status === 'active')?.id || 0")"
if [[ "$virtual_node_id" -lt 1 ]]; then
  payload="$(
    NAME="$VIRTUAL_NODE_NAME" PORT="$VIRTUAL_NODE_PORT" SELECTOR="$VIRTUAL_NODE_SELECTOR" node <<'NODE'
const port = Number(process.env.PORT || "8443");
console.log(JSON.stringify({
  name: process.env.NAME,
  listen_protocol: "vless",
  listen_port: port,
  tag_selector: process.env.SELECTOR || "{}",
  strategy: "selector",
}));
NODE
  )"
  post_json "/api/virtual-nodes" "$payload" "$OUT_DIR/virtual-node-create.json"
  virtual_node_id="$(json_expr "$OUT_DIR/virtual-node-create.json" "data.id || 0")"
  virtual_node_name="$(json_expr "$OUT_DIR/virtual-node-create.json" "data.name || ''")"
  log "created usable virtual node: id=$virtual_node_id name=$virtual_node_name port=$VIRTUAL_NODE_PORT"
else
  virtual_node_name="$(json_expr "$OUT_DIR/virtual-nodes.json" "list(data).find((row) => row.id === Number('$virtual_node_id'))?.name || ''")"
  log "found active virtual node: id=$virtual_node_id name=$virtual_node_name"
fi

if [[ -z "$virtual_node_name" ]]; then
  log "usable virtual node name is empty"
  exit 1
fi

policy_scope_type="team"
policy_scope_id="$team_id"
if [[ "$token_team_id" -gt 0 ]]; then
  policy_scope_id="$token_team_id"
elif [[ "$policy_scope_id" -lt 1 ]]; then
  policy_scope_type="user"
  policy_scope_id="$user_id"
fi

policy_id="$(POLICY_NAME="$POLICY_NAME" json_expr "$OUT_DIR/policies.json" "list(data).find((row) => row.name === process.env.POLICY_NAME)?.id || 0")"
allowed_virtual_nodes="$(
  VNODE="$virtual_node_name" node <<'NODE'
console.log(JSON.stringify([process.env.VNODE]));
NODE
)"
policy_payload="$(
  NAME="$POLICY_NAME" SCOPE_TYPE="$policy_scope_type" SCOPE_ID="$policy_scope_id" ALLOWED="$allowed_virtual_nodes" node <<'NODE'
console.log(JSON.stringify({
  name: process.env.NAME,
  scope_type: process.env.SCOPE_TYPE,
  scope_id: Number(process.env.SCOPE_ID),
  include_tags: "[]",
  exclude_tags: "[]",
  allowed_virtual_nodes: process.env.ALLOWED,
  max_nodes: 1,
  status: "active",
}));
NODE
)"
if [[ "$policy_id" -lt 1 ]]; then
  post_json "/api/policies" "$policy_payload" "$OUT_DIR/policy-create.json"
  policy_id="$(json_expr "$OUT_DIR/policy-create.json" "data.id || 0")"
  log "created usable default policy: id=$policy_id scope=$policy_scope_type:$policy_scope_id allowed=$allowed_virtual_nodes"
else
  patch_json "/api/policies/$policy_id" "$policy_payload" "$OUT_DIR/policy-update.json"
  log "updated usable default policy: id=$policy_id scope=$policy_scope_type:$policy_scope_id allowed=$allowed_virtual_nodes"
fi

run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/check" -o "$OUT_DIR/sing-box-check.json"
config_valid="$(json_expr "$OUT_DIR/sing-box-check.json" "data.valid === true")"
inbound_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.inbound_count || 0")"
user_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.user_count || 0")"
upstream_count="$(json_expr "$OUT_DIR/sing-box-check.json" "data.upstream_outbound_count || 0")"
config_hash="$(json_expr "$OUT_DIR/sing-box-check.json" "data.config_hash || ''")"

if [[ "$config_valid" != "true" || "$inbound_count" -lt 1 || "$user_count" -lt 1 || "$upstream_count" -lt 1 ]]; then
  log "generated sing-box config is not usable: valid=$config_valid inbound=$inbound_count users=$user_count upstreams=$upstream_count hash=$config_hash"
  exit 1
fi

if [[ "$PUBLISH_CONFIG" == "true" ]]; then
  run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/publish" -o "$OUT_DIR/sing-box-publish.json"
  published="$(json_expr "$OUT_DIR/sing-box-publish.json" "data.published === true")"
  publish_hash="$(json_expr "$OUT_DIR/sing-box-publish.json" "data.config_hash || ''")"
  if [[ "$published" != "true" || "$publish_hash" != "$config_hash" ]]; then
    log "sing-box config publish failed or hash mismatch: published=$published publish_hash=$publish_hash check_hash=$config_hash"
    exit 1
  fi
  log "published usable sing-box config: hash=$publish_hash"
fi

if [[ "$RELOAD_SING_BOX" == "true" ]]; then
  if [[ -n "${REMOTE_HOST:-}" && -n "${REMOTE_DIR:-}" ]]; then
    log "restarting remote sing-box container"
    ssh "$REMOTE_HOST" "cd '$REMOTE_DIR' && docker compose restart sing-box"
  else
    log "REMOTE_HOST or REMOTE_DIR not configured; skipping data-plane reload"
  fi
fi

log "usable closeout complete: virtual_node=$virtual_node_name policy_id=$policy_id config_hash=$config_hash"
if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "ensure-usable artifacts: $OUT_DIR"
else
  log "ensure-usable temporary artifacts will be removed"
fi
