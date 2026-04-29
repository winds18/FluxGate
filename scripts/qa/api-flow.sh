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
post_json "/api/policies" "{\"name\":\"QA 默认策略\",\"scope_type\":\"team\",\"scope_id\":$team_id,\"max_nodes\":5}" "$OUT_DIR/policy.json"
policy_id="$(json_value "data.id" <"$OUT_DIR/policy.json")"
policy_scope_id="$(json_value "data.scope_id" <"$OUT_DIR/policy.json")"
policy_max_nodes="$(json_value "data.max_nodes" <"$OUT_DIR/policy.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/policies" -o "$OUT_DIR/policies.json"
policy_list_count="$(json_value "data.length" <"$OUT_DIR/policies.json")"

post_json "/api/users" "{\"team_id\":$team_id,\"name\":\"QA User\",\"email\":\"qa@example.test\"}" "$OUT_DIR/user.json"
user_id="$(json_value "data.id" <"$OUT_DIR/user.json")"

post_json "/api/sources" '{"name":"机场A","type":"manual"}' "$OUT_DIR/source-a.json"
source_a_id="$(json_value "data.id" <"$OUT_DIR/source-a.json")"
post_json "/api/sources" '{"name":"机场A","type":"manual"}' "$OUT_DIR/source-b.json"
source_b_prefix="$(json_value "JSON.stringify(data.display_prefix)" <"$OUT_DIR/source-b.json")"
clash_subscription_raw="$(node -e 'process.stdout.write(JSON.stringify(`proxies:
  - name: "新加坡 01"
    type: ss
    server: example.sub
    port: 8388
    cipher: aes-128-gcm
    password: "qa-placeholder"
`));')"
post_json "/api/sources" "{\"name\":\"订阅源A\",\"type\":\"subscription\",\"raw_content\":$clash_subscription_raw,\"refresh_interval_minutes\":5}" "$OUT_DIR/source-subscription.json"
subscription_source_id="$(json_value "data.id" <"$OUT_DIR/source-subscription.json")"
subscription_refresh_interval="$(json_value "data.refresh_interval_minutes" <"$OUT_DIR/source-subscription.json")"
sip008_subscription_raw="$(node -e 'process.stdout.write(JSON.stringify(JSON.stringify({version:1,servers:[{remarks:"首尔 04",server:"sip008.example.sub",server_port:8388,method:"aes-128-gcm",password:"qa-placeholder"}]})));')"
post_json "/api/sources" "{\"name\":\"SIP008订阅\",\"type\":\"subscription\",\"raw_content\":$sip008_subscription_raw,\"refresh_interval_minutes\":0}" "$OUT_DIR/source-sip008.json"
sip008_source_id="$(json_value "data.id" <"$OUT_DIR/source-sip008.json")"
singbox_subscription_raw="$(node -e 'process.stdout.write(JSON.stringify(JSON.stringify({outbounds:[{type:"shadowsocks",tag:"香港 05",server:"singbox.example.sub",server_port:8388,method:"aes-128-gcm",password:"qa-placeholder"}]})));')"
post_json "/api/sources" "{\"name\":\"sing-box订阅\",\"type\":\"subscription\",\"raw_content\":$singbox_subscription_raw,\"refresh_interval_minutes\":0}" "$OUT_DIR/source-sing-box.json"
singbox_source_id="$(json_value "data.id" <"$OUT_DIR/source-sing-box.json")"
vmess_uri="$(node -e 'const doc={add:"vmess.example.net",port:"443",id:"00000000-0000-0000-0000-000000000046",aid:"0",scy:"auto",net:"ws",host:"ws.example.test",path:"/ws",tls:"tls",sni:"vmess.example.net",ps:"VMess QA"}; process.stdout.write("vmess://"+Buffer.from(JSON.stringify(doc)).toString("base64url"));')"
hysteria2_uri="hysteria2://qa-placeholder@example.dev:443?obfs=salamander&obfs-password=obfs-placeholder&sni=hy2.example.dev&insecure=1#首尔%2002"
tuic_uri="tuic://00000000-0000-0000-0000-000000000048:qa-placeholder@example.io:443?congestion_control=bbr&udp_relay_mode=native&sni=tuic.example.io&alpn=h3&insecure=1#大阪%2001"
anytls_uri="anytls://qa-placeholder@example.chat:443?sni=anytls.example.chat&alpn=h2,http/1.1&idle_session_check_interval=20s&idle_session_timeout=45s&min_idle_session=2&insecure=1#香港%2002"
shadowtls_uri="shadowtls://qa-placeholder@example.help:443?version=3&sni=shadow.example.help&alpn=h2&insecure=1#东京%2002"
naive_uri="naive://qa-user:qa-placeholder@example.news:443?sni=naive.example.news&quic=1&quic_congestion_control=bbr&udp_over_tcp=1&insecure_concurrency=2#新加坡%2002"
hysteria_uri="hysteria://qa-placeholder@example.zone:443?auth_str=qa-auth&up_mbps=20&down_mbps=80&obfs=obfs-placeholder&protocol=udp&sni=hysteria.example.zone&alpn=h3&insecure=1#香港%2003"
http_uri="https://qa-user:qa-placeholder@example.proxy:8443/connect?sni=http-proxy.example.proxy&insecure=1#东京%2003"
socks_uri="socks5://qa-user:qa-placeholder@example.socks:1080?network=udp&udp_over_tcp=1#首尔%2003"
ssh_uri="ssh://qa-user:qa-placeholder@example.ssh:22?private_key_path=keys%2Fqa_id_ed25519&host_key_algorithms=ssh-ed25519,rsa-sha2-512&client_version=SSH-2.0-FluxGateQA#香港%2004"
wireguard_uri="wireguard://example.wg:51820?private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=&peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==&pre_shared_key=cHNrLXBsYWNlaG9sZGVy&local_address=10.66.0.2/32,fd00::2/128&allowed_ips=0.0.0.0/0,::/0&reserved=1,2,3&mtu=1420#台北%2001"
tor_uri="tor://default?executable_path=/usr/bin/tor&extra_args=--quiet,--SocksPort,auto&data_directory=cache%2Ftor&torrc.ClientOnly=1#匿名%2001"

post_json "/api/nodes/import" "{\"source_id\":$source_a_id,\"content\":\"vless://uuid@example.com:443#香港%2001\\ntrojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京%2001\\nss://aes-128-gcm:qa-placeholder@example.net:8388#首尔%2001\\n$vmess_uri\\n$hysteria2_uri\\n$tuic_uri\\n$anytls_uri\\n$shadowtls_uri\\n$naive_uri\\n$hysteria_uri\\n$http_uri\\n$socks_uri\\n$ssh_uri\\n$wireguard_uri\\n$tor_uri\"}" "$OUT_DIR/import.json"
post_json "/api/sources/$subscription_source_id/refresh" '{}' "$OUT_DIR/source-refresh.json"
source_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-refresh.json")"
post_json "/api/sources/$sip008_source_id/refresh" '{}' "$OUT_DIR/source-sip008-refresh.json"
sip008_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-sip008-refresh.json")"
post_json "/api/sources/$singbox_source_id/refresh" '{}' "$OUT_DIR/source-sing-box-refresh.json"
singbox_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-sing-box-refresh.json")"
post_json "/api/virtual-nodes" '{"name":"FluxGate-HK","listen_protocol":"vless","listen_port":8443}' "$OUT_DIR/virtual-node.json"
post_json "/api/tokens" "{\"user_id\":$user_id,\"name\":\"QA Token\",\"expire_days\":30,\"quota_bytes\":1048576}" "$OUT_DIR/token.json"
token_id="$(json_value "data.token.id" <"$OUT_DIR/token.json")"
plain_token="$(json_value "data.plain_token" <"$OUT_DIR/token.json")"
post_json "/api/tokens/$token_id/extend" '{"extend_days":30}' "$OUT_DIR/token-extend.json"
token_extended_status="$(json_value "data.status" <"$OUT_DIR/token-extend.json")"
post_json "/api/tokens/$token_id/quota" '{"quota_bytes":1073741824}' "$OUT_DIR/token-quota.json"
token_quota_after="$(json_value "data.quota_bytes" <"$OUT_DIR/token-quota.json")"
post_json "/api/tokens/$token_id/revoke" '{}' "$OUT_DIR/token-revoke.json"
token_revoked_status="$(json_value "data.status" <"$OUT_DIR/token-revoke.json")"
log "+ curl -sS -o $OUT_DIR/subscription-revoked.json -w <http_code> $BASE_URL/sub/<redacted>?target=clash"
revoked_http_status="$(curl -sS -o "$OUT_DIR/subscription-revoked.json" -w "%{http_code}" "$BASE_URL/sub/$plain_token?target=clash")"
gateway_user="fg_u_${user_id}_t_${token_id}"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/generate" -o "$OUT_DIR/sing-box-revoked.json"
post_json "/api/tokens/$token_id/restore" '{}' "$OUT_DIR/token-restore.json"
token_restored_status="$(json_value "data.status" <"$OUT_DIR/token-restore.json")"

log "+ curl -fsS $BASE_URL/sub/<redacted>?target=clash -o $OUT_DIR/subscription.yaml"
curl -fsS "$BASE_URL/sub/$plain_token?target=clash" -o "$OUT_DIR/subscription.yaml"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/generate" -o "$OUT_DIR/sing-box.json"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/check" -o "$OUT_DIR/sing-box-check.json"
config_check_valid="$(json_value "data.valid" <"$OUT_DIR/sing-box-check.json")"
config_check_hash="$(json_value "data.config_hash" <"$OUT_DIR/sing-box-check.json")"
config_check_upstreams="$(json_value "data.upstream_outbound_count" <"$OUT_DIR/sing-box-check.json")"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/publish" -o "$OUT_DIR/sing-box-publish.json"
config_publish_done="$(json_value "data.published" <"$OUT_DIR/sing-box-publish.json")"
config_publish_hash="$(json_value "data.config_hash" <"$OUT_DIR/sing-box-publish.json")"
config_publish_restart_required="$(json_value "data.restart_required" <"$OUT_DIR/sing-box-publish.json")"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/publish" -o "$OUT_DIR/sing-box-publish-again.json"
config_publish_again_previous="$(json_value "data.previous_saved" <"$OUT_DIR/sing-box-publish-again.json")"
config_publish_again_restart_required="$(json_value "data.restart_required" <"$OUT_DIR/sing-box-publish-again.json")"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/rollback" -o "$OUT_DIR/sing-box-rollback.json"
config_rollback_done="$(json_value "data.rolled_back" <"$OUT_DIR/sing-box-rollback.json")"
config_rollback_hash="$(json_value "data.config_hash" <"$OUT_DIR/sing-box-rollback.json")"
config_rollback_restart_required="$(json_value "data.restart_required" <"$OUT_DIR/sing-box-rollback.json")"

if [[ "$source_b_prefix" != '"[机场A-2] "' ]]; then
  log "unexpected auto prefix for duplicate source: $source_b_prefix"
  exit 1
fi

if [[ "$policy_id" -lt 1 || "$policy_scope_id" != "$team_id" || "$policy_max_nodes" != "5" || "$policy_list_count" -lt 1 ]]; then
  log "policy create/list should work: id=$policy_id scope_id=$policy_scope_id team_id=$team_id max_nodes=$policy_max_nodes list_count=$policy_list_count"
  exit 1
fi

if [[ "$source_refresh_imported" != "1" ]]; then
  log "unexpected subscription refresh import count: $source_refresh_imported"
  exit 1
fi

if [[ "$subscription_refresh_interval" != "5" ]]; then
  log "unexpected subscription refresh interval: $subscription_refresh_interval"
  exit 1
fi

if [[ "$sip008_refresh_imported" != "1" ]]; then
  log "unexpected SIP008 refresh import count: $sip008_refresh_imported"
  exit 1
fi

if [[ "$singbox_refresh_imported" != "1" ]]; then
  log "unexpected sing-box JSON refresh import count: $singbox_refresh_imported"
  exit 1
fi

if [[ "$token_extended_status" != "active" ]]; then
  log "unexpected token status after extend: $token_extended_status"
  exit 1
fi

if [[ "$token_quota_after" != "1074790400" ]]; then
  log "unexpected token quota after add: $token_quota_after"
  exit 1
fi

if [[ "$token_revoked_status" != "revoked" || "$revoked_http_status" != "403" ]]; then
  log "revoked token should reject subscription: token_status=$token_revoked_status http_status=$revoked_http_status"
  exit 1
fi

if grep -q "$gateway_user" "$OUT_DIR/sing-box-revoked.json"; then
  log "revoked token should not appear in sing-box config"
  exit 1
fi

if [[ "$token_restored_status" != "active" ]]; then
  log "unexpected token status after restore: $token_restored_status"
  exit 1
fi

if ! grep -q 'FluxGate-HK' "$OUT_DIR/subscription.yaml"; then
  log "subscription missing FluxGate-HK"
  exit 1
fi

if ! grep -q "$gateway_user" "$OUT_DIR/sing-box.json"; then
  log "restored token should appear in sing-box config"
  exit 1
fi

if [[ "$config_check_valid" != "true" || -z "$config_check_hash" || "$config_check_upstreams" -lt 1 ]]; then
  log "sing-box config check should pass with upstreams: valid=$config_check_valid hash=$config_check_hash upstreams=$config_check_upstreams"
  exit 1
fi

if [[ "$config_publish_done" != "true" || "$config_publish_hash" != "$config_check_hash" ]]; then
  log "sing-box config publish should persist checked config: published=$config_publish_done publish_hash=$config_publish_hash check_hash=$config_check_hash"
  exit 1
fi

if [[ "$config_publish_restart_required" != "true" || "$config_publish_again_restart_required" != "true" ]]; then
  log "sing-box config publish should mark restart required: first=$config_publish_restart_required second=$config_publish_again_restart_required"
  exit 1
fi

if [[ "$config_publish_again_previous" != "true" || "$config_rollback_done" != "true" || "$config_rollback_hash" != "$config_check_hash" ]]; then
  log "sing-box config rollback should restore previous checked config: previous=$config_publish_again_previous rollback=$config_rollback_done rollback_hash=$config_rollback_hash check_hash=$config_check_hash"
  exit 1
fi

if [[ "$config_rollback_restart_required" != "true" ]]; then
  log "sing-box config rollback should mark restart required: rollback=$config_rollback_restart_required"
  exit 1
fi

if ! grep -q '"tag": "up_' "$OUT_DIR/sing-box.json"; then
  log "active upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "trojan"' "$OUT_DIR/sing-box.json"; then
  log "active trojan upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "shadowsocks"' "$OUT_DIR/sing-box.json"; then
  log "active shadowsocks upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "vmess"' "$OUT_DIR/sing-box.json"; then
  log "active vmess upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "hysteria2"' "$OUT_DIR/sing-box.json"; then
  log "active hysteria2 upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "tuic"' "$OUT_DIR/sing-box.json"; then
  log "active tuic upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "anytls"' "$OUT_DIR/sing-box.json"; then
  log "active anytls upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "shadowtls"' "$OUT_DIR/sing-box.json"; then
  log "active shadowtls upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "naive"' "$OUT_DIR/sing-box.json"; then
  log "active naive upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "hysteria"' "$OUT_DIR/sing-box.json"; then
  log "active hysteria upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "http"' "$OUT_DIR/sing-box.json"; then
  log "active http upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "socks"' "$OUT_DIR/sing-box.json"; then
  log "active socks upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "ssh"' "$OUT_DIR/sing-box.json"; then
  log "active ssh upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "wireguard"' "$OUT_DIR/sing-box.json"; then
  log "active wireguard upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"type": "tor"' "$OUT_DIR/sing-box.json"; then
  log "active tor upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"final": "FluxGate-upstreams"' "$OUT_DIR/sing-box.json"; then
  log "sing-box route final should use upstream selector when upstream nodes exist"
  exit 1
fi

if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "API flow passed; artifacts: $OUT_DIR"
else
  log "API flow passed; temporary artifacts will be removed"
fi
