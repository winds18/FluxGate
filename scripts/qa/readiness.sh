#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_CONFIG="${DEPLOY_CONFIG:-$ROOT_DIR/.env.deploy.local}"
if [[ -f "$DEPLOY_CONFIG" ]]; then
  # shellcheck disable=SC1090
  source "$DEPLOY_CONFIG"
fi

BASE_URL="${1:-${BASE_URL:-${REMOTE_BROWSER_BASE_URL:-${BROWSER_BASE_URL:-http://127.0.0.1:8080}}}}"
configure_no_proxy_for_url "$BASE_URL"
ADMIN_USERNAME="${ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}}}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
STRICT="${STRICT:-false}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/readiness/$(timestamp)}"
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
  log "ADMIN_USERNAME and ADMIN_PASSWORD are required for readiness check"
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

list_count() {
  json_expr "$1" "data && Array.isArray(data.data) ? data.data.length : (Array.isArray(data) ? data.length : 0)"
}

warn() {
  log "readiness warning: $*"
  warnings=$((warnings + 1))
}

log "running read-only readiness check against $BASE_URL"
curl -fsS "$BASE_URL/healthz" -o "$OUT_DIR/healthz.json"
curl -fsS "$BASE_URL/readyz" -o "$OUT_DIR/readyz.json"

log "+ curl -fsS -c $COOKIE_JAR -X POST $BASE_URL/api/auth/login -H content-type:application/json --data <redacted> -o $OUT_DIR/login.json"
curl -fsS -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
  -H 'content-type: application/json' \
  --data "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -o "$OUT_DIR/login.json"

curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/overview" -o "$OUT_DIR/overview.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/teams" -o "$OUT_DIR/teams.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/users" -o "$OUT_DIR/users.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/tokens" -o "$OUT_DIR/tokens.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/sources" -o "$OUT_DIR/sources.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/nodes" -o "$OUT_DIR/nodes.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/virtual-nodes" -o "$OUT_DIR/virtual-nodes.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/policies" -o "$OUT_DIR/policies.json"
curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/traffic/tokens" -o "$OUT_DIR/traffic-tokens.json"

teams="$(list_count "$OUT_DIR/teams.json")"
users="$(list_count "$OUT_DIR/users.json")"
tokens="$(list_count "$OUT_DIR/tokens.json")"
sources="$(list_count "$OUT_DIR/sources.json")"
nodes="$(list_count "$OUT_DIR/nodes.json")"
virtual_nodes="$(list_count "$OUT_DIR/virtual-nodes.json")"
policies="$(list_count "$OUT_DIR/policies.json")"
traffic_tokens="$(list_count "$OUT_DIR/traffic-tokens.json")"
regions="$(json_expr "$OUT_DIR/nodes.json" "data && Array.isArray(data.data) ? new Set(data.data.map((node) => node.region || '其他')).size : (Array.isArray(data) ? new Set(data.map((node) => node.region || '其他')).size : 0)")"

warnings=0
if [[ "$teams" -lt 1 ]]; then warn "没有团队数据，无法体验团队管理和策略范围"; fi
if [[ "$users" -lt 1 ]]; then warn "没有成员数据，无法体验成员绑定和 Token 归属"; fi
if [[ "$tokens" -lt 1 ]]; then warn "没有 Token 数据，无法体验订阅分发、续期、加额、撤销和恢复"; fi
if [[ "$sources" -lt 1 ]]; then warn "没有上游来源，无法体验来源编辑和同步"; fi
if [[ "$nodes" -lt 1 ]]; then warn "没有节点数据，无法体验地区聚合、节点详情和节点编辑"; fi
if [[ "$virtual_nodes" -lt 1 ]]; then warn "没有虚拟节点，无法体验网关入口配置、虚拟节点订阅输出和行内编辑"; fi
if [[ "$policies" -lt 1 ]]; then warn "没有策略，无法体验团队/成员/Token 策略限制"; fi
if [[ "$traffic_tokens" -lt 1 ]]; then warn "没有 Token 流量摘要，无法体验额度进度条"; fi

log "readiness summary: teams=$teams users=$users tokens=$tokens sources=$sources nodes=$nodes regions=$regions virtual_nodes=$virtual_nodes policies=$policies traffic_tokens=$traffic_tokens warnings=$warnings"

if [[ "$STRICT" == "true" && "$warnings" -gt 0 ]]; then
  log "readiness failed in STRICT mode"
  exit 1
fi

if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "readiness check finished; artifacts: $OUT_DIR"
else
  log "readiness check finished; temporary artifacts will be removed"
fi
