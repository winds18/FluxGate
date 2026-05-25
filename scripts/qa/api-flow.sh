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

header_contains() {
  local file="$1"
  local pattern="$2"
  tr -d '\r' <"$file" | grep -Eiq "$pattern"
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
patch_json "/api/teams/$team_id" '{"name":"QA Team Edited","description":"automated smoke edited","status":"active"}' "$OUT_DIR/team-update.json"
team_updated_name="$(json_value "data.name" <"$OUT_DIR/team-update.json")"
team_updated_description="$(json_value "data.description" <"$OUT_DIR/team-update.json")"
team_updated_status="$(json_value "data.status" <"$OUT_DIR/team-update.json")"
post_json "/api/policies" "{\"name\":\"QA 默认策略\",\"scope_type\":\"team\",\"scope_id\":$team_id,\"include_tags\":\"[\\\"QA-HK\\\"]\",\"allowed_virtual_nodes\":\"[\\\"FluxGate-HK\\\"]\",\"max_nodes\":5}" "$OUT_DIR/policy.json"
policy_id="$(json_value "data.id" <"$OUT_DIR/policy.json")"
policy_scope_id="$(json_value "data.scope_id" <"$OUT_DIR/policy.json")"
policy_max_nodes="$(json_value "data.max_nodes" <"$OUT_DIR/policy.json")"
policy_include_tags="$(json_value "data.include_tags" <"$OUT_DIR/policy.json")"
policy_allowed_virtual_nodes="$(json_value "data.allowed_virtual_nodes" <"$OUT_DIR/policy.json")"
patch_json "/api/policies/$policy_id" "{\"name\":\"QA 默认策略-已调整\",\"scope_type\":\"team\",\"scope_id\":$team_id,\"include_tags\":\"[\\\"QA-HK\\\"]\",\"exclude_tags\":\"[\\\"Backup\\\"]\",\"allowed_virtual_nodes\":\"[\\\"FluxGate-HK\\\"]\",\"max_nodes\":5,\"status\":\"active\"}" "$OUT_DIR/policy-update.json"
policy_updated_name="$(json_value "data.name" <"$OUT_DIR/policy-update.json")"
policy_updated_status="$(json_value "data.status" <"$OUT_DIR/policy-update.json")"
policy_updated_exclude_tags="$(json_value "data.exclude_tags" <"$OUT_DIR/policy-update.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/policies" -o "$OUT_DIR/policies.json"
policy_list_count="$(json_value "data.length" <"$OUT_DIR/policies.json")"

post_json "/api/users" "{\"team_id\":$team_id,\"name\":\"QA User\",\"email\":\"qa@example.test\"}" "$OUT_DIR/user.json"
user_id="$(json_value "data.id" <"$OUT_DIR/user.json")"
patch_json "/api/users/$user_id" "{\"team_id\":$team_id,\"name\":\"QA User Edited\",\"email\":\"qa-edited@example.test\",\"remark\":\"QA remark\",\"status\":\"active\"}" "$OUT_DIR/user-update.json"
user_updated_name="$(json_value "data.name" <"$OUT_DIR/user-update.json")"
user_updated_email="$(json_value "data.email" <"$OUT_DIR/user-update.json")"
user_updated_remark="$(json_value "data.remark" <"$OUT_DIR/user-update.json")"
user_updated_status="$(json_value "data.status" <"$OUT_DIR/user-update.json")"

post_json "/api/sources" '{"name":"机场A","type":"manual","default_tags":"[\"QA-HK\"]"}' "$OUT_DIR/source-a.json"
source_a_id="$(json_value "data.id" <"$OUT_DIR/source-a.json")"
post_json "/api/sources" '{"name":"机场A","type":"manual"}' "$OUT_DIR/source-b.json"
source_b_id="$(json_value "data.id" <"$OUT_DIR/source-b.json")"
source_b_prefix="$(json_value "JSON.stringify(data.display_prefix)" <"$OUT_DIR/source-b.json")"
patch_json "/api/sources/$source_b_id" '{"name":"机场B","type":"subscription","url":"https://example.test/sub","display_prefix":"[手动B]","default_tags":"SG, Backup","refresh_interval_minutes":15}' "$OUT_DIR/source-b-update.json"
source_b_updated_name="$(json_value "data.name" <"$OUT_DIR/source-b-update.json")"
source_b_updated_type="$(json_value "data.type" <"$OUT_DIR/source-b-update.json")"
source_b_updated_url="$(json_value "data.url" <"$OUT_DIR/source-b-update.json")"
source_b_updated_prefix="$(json_value "data.display_prefix" <"$OUT_DIR/source-b-update.json")"
source_b_updated_tags="$(json_value "data.default_tags" <"$OUT_DIR/source-b-update.json")"
source_b_updated_interval="$(json_value "data.refresh_interval_minutes" <"$OUT_DIR/source-b-update.json")"
clash_subscription_raw="$(node -e 'process.stdout.write(JSON.stringify(`proxies:
  - name: "新加坡 01"
    type: ss
    server: example.sub
    port: 8388
    cipher: aes-128-gcm
    password: "qa-placeholder"
  - name: "新加坡 02"
    type: hysteria2
    server: hy2.example.sub
    port: 443
    password: "hy2-placeholder"
    obfs: salamander
    obfs-password: "obfs-placeholder"
    sni: hy2.example.sub
    skip-cert-verify: true
  - { name: "大阪 02", type: tuic, server: tuic.example.sub, port: 443, uuid: "00000000-0000-0000-0000-000000000050", password: "tuic-placeholder", congestion-controller: bbr, udp-relay-mode: native, sni: tuic.example.sub }
  - name: "香港 09"
    type: hysteria
    server: hysteria.example.sub
    port: 443
    auth-str: "hysteria-auth"
    up-mbps: 20
    down-mbps: 80
    obfs: "obfs-placeholder"
    protocol: udp
    sni: hysteria.example.sub
    skip-cert-verify: true
  - name: "东京 HTTP"
    type: http
    server: http.clash.example.sub
    port: 8080
    username: qa-user
    password: "http-placeholder"
    tls: true
    sni: http.clash.example.sub
    skip-cert-verify: true
    path: connect
  - { name: "首尔 SOCKS", type: socks5, server: socks.clash.example.sub, port: 1080, username: "qa-user", password: "socks-placeholder", udp-over-tcp: true, network: udp }
  - name: "香港 AnyTLS"
    type: anytls
    server: anytls.clash.example.sub
    port: 443
    password: "anytls-placeholder"
    idle-session-check-interval: 20s
    idle-session-timeout: 45s
    min-idle-session: 2
    sni: anytls.clash.example.sub
  - { name: "东京 ShadowTLS", type: shadowtls, server: shadowtls.clash.example.sub, port: 443, version: 3, password: "shadow-placeholder", sni: shadowtls.clash.example.sub, skip-cert-verify: true }
  - { name: "新加坡 Naive", type: naive+quic, server: naive.clash.example.sub, port: 443, username: "qa-user", password: "naive-placeholder", sni: naive.clash.example.sub, quic: true, quic-congestion-control: bbr, udp-over-tcp: true, insecure-concurrency: 2 }
  - name: "香港 SSH"
    type: ssh
    server: ssh.clash.example.sub
    port: 22
    username: qa-user
    password: "ssh-placeholder"
    private-key-path: keys/qa_id_ed25519
    host-key-algorithms: ssh-ed25519,rsa-sha2-512
    client-version: SSH-2.0-FluxGateQA
  - name: "台北 WireGuard"
    type: wireguard
    server: wg.clash.example.sub
    port: 51820
    private-key: cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=
    public-key: cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==
    ip: 10.66.0.2/32
    ipv6: fd00::2/128
    pre-shared-key: cHNrLXBsYWNlaG9sZGVy
    allowed-ips: 0.0.0.0/0,::/0
    reserved: 1,2,3
    mtu: 1420
    udp: true
  - name: "东京 VMess"
    type: vmess
    server: vmess.clash.example.sub
    port: 443
    uuid: 00000000-0000-0000-0000-000000000056
    alter-id: 0
    cipher: auto
    network: ws
    ws-path: /ws
    ws-headers.host: ws.clash.example.sub
    tls: true
    sni: vmess.clash.example.sub
    skip-cert-verify: true
    alpn: h2,http/1.1
`));')"
post_json "/api/sources" "{\"name\":\"订阅源A\",\"type\":\"subscription\",\"raw_content\":$clash_subscription_raw,\"refresh_interval_minutes\":5}" "$OUT_DIR/source-subscription.json"
subscription_source_id="$(json_value "data.id" <"$OUT_DIR/source-subscription.json")"
subscription_refresh_interval="$(json_value "data.refresh_interval_minutes" <"$OUT_DIR/source-subscription.json")"
sip008_subscription_raw="$(node -e 'process.stdout.write(JSON.stringify(JSON.stringify({version:1,servers:[{remarks:"首尔 04",server:"sip008.example.sub",server_port:8388,method:"aes-128-gcm",password:"qa-placeholder"}]})));')"
post_json "/api/sources" "{\"name\":\"SIP008订阅\",\"type\":\"subscription\",\"raw_content\":$sip008_subscription_raw,\"refresh_interval_minutes\":0}" "$OUT_DIR/source-sip008.json"
sip008_source_id="$(json_value "data.id" <"$OUT_DIR/source-sip008.json")"
singbox_subscription_raw="$(
node <<'NODE'
const payload = {
  outbounds: [
    { type: "shadowsocks", tag: "香港 05", server: "singbox.example.sub", server_port: 8388, method: "aes-128-gcm", password: "qa-placeholder" },
    { type: "hysteria2", tag: "香港 06", server: "hy2.singbox.example.sub", server_port: 443, password: "hy2-placeholder", obfs: { type: "salamander", password: "obfs-placeholder" }, tls: { enabled: true, server_name: "hy2.singbox.example.sub", insecure: true } },
    { type: "tuic", tag: "东京 06", server: "tuic.singbox.example.sub", server_port: 443, uuid: "00000000-0000-0000-0000-000000000049", password: "tuic-placeholder", congestion_control: "bbr", udp_relay_mode: "native", tls: { enabled: true, server_name: "tuic.singbox.example.sub" } },
    { type: "anytls", tag: "首尔 06", server: "anytls.singbox.example.sub", server_port: 443, password: "anytls-placeholder", idle_session_check_interval: "20s", idle_session_timeout: "45s", min_idle_session: 2, tls: { enabled: true, server_name: "anytls.singbox.example.sub" } },
    { type: "shadowtls", tag: "东京 07", server: "shadow.singbox.example.sub", server_port: 443, version: 3, password: "shadow-placeholder", tls: { enabled: true, server_name: "shadow.singbox.example.sub", insecure: true } },
    { type: "hysteria", tag: "香港 07", server: "hysteria.singbox.example.sub", server_port: 443, auth_str: "hysteria-auth", up_mbps: 20, down_mbps: 80, obfs: "obfs-placeholder", network: "udp", tls: { enabled: true, server_name: "hysteria.singbox.example.sub", insecure: true, alpn: ["h3"] } },
    { type: "http", tag: "首尔 07", server: "http.singbox.example.sub", server_port: 8443, username: "qa-user", password: "http-placeholder", path: "/connect", tls: { enabled: true, server_name: "http.singbox.example.sub", insecure: true } },
    { type: "socks", tag: "大阪 07", server: "socks.singbox.example.sub", server_port: 1080, username: "qa-user", password: "socks-placeholder", version: "5", network: "udp", udp_over_tcp: true },
    { type: "ssh", tag: "香港 08", server: "ssh.singbox.example.sub", server_port: 22, user: "qa-user", password: "ssh-placeholder", private_key_path: "keys/qa_id_ed25519", host_key_algorithms: ["ssh-ed25519", "rsa-sha2-512"], client_version: "SSH-2.0-FluxGateQA" },
    { type: "wireguard", tag: "台北 02", server: "wg.singbox.example.sub", server_port: 51820, private_key: "cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=", local_address: ["10.66.0.2/32", "fd00::2/128"], peers: [{ public_key: "cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==", allowed_ips: ["0.0.0.0/0", "::/0"], pre_shared_key: "cHNrLXBsYWNlaG9sZGVy", reserved: [1, 2, 3] }], mtu: 1420 },
    { type: "tor", tag: "匿名 02", executable_path: "/usr/bin/tor", extra_args: ["--quiet", "--SocksPort", "auto"], data_directory: "cache/tor", torrc: { ClientOnly: "1" } },
  ],
};
process.stdout.write(JSON.stringify(JSON.stringify(payload)));
NODE
)"
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
route_uri="vless://00000000-0000-0000-0000-000000000051@example.route:443#香港-%E7%BE%8E%E5%9B%BD"

post_json "/api/nodes/import" "{\"source_id\":$source_a_id,\"content\":\"vless://uuid@example.com:443#香港%2001\\ntrojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京%2001\\nss://aes-128-gcm:qa-placeholder@example.net:8388#首尔%2001\\n$vmess_uri\\n$hysteria2_uri\\n$tuic_uri\\n$anytls_uri\\n$shadowtls_uri\\n$naive_uri\\n$hysteria_uri\\n$http_uri\\n$socks_uri\\n$ssh_uri\\n$wireguard_uri\\n$tor_uri\\n$route_uri\"}" "$OUT_DIR/import.json"
post_json "/api/sources/$subscription_source_id/refresh" '{}' "$OUT_DIR/source-refresh.json"
source_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-refresh.json")"
post_json "/api/sources/$sip008_source_id/refresh" '{}' "$OUT_DIR/source-sip008-refresh.json"
sip008_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-sip008-refresh.json")"
post_json "/api/sources/$singbox_source_id/refresh" '{}' "$OUT_DIR/source-sing-box-refresh.json"
singbox_refresh_imported="$(json_value "data.result.imported" <"$OUT_DIR/source-sing-box-refresh.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/nodes" -o "$OUT_DIR/nodes.json"
qa_hk_tagged_count="$(json_value "data.filter((node) => (node.tags || []).includes('QA-HK')).length" <"$OUT_DIR/nodes.json")"
china_hk_region_count="$(json_value "data.filter((node) => node.region === '🇨🇳中国|香港').length" <"$OUT_DIR/nodes.json")"
china_tw_region_count="$(json_value "data.filter((node) => node.region === '🇨🇳中国|台湾').length" <"$OUT_DIR/nodes.json")"
us_route_region_count="$(json_value "data.filter((node) => node.raw_name === '香港-美国' && node.region === '🇺🇸美国').length" <"$OUT_DIR/nodes.json")"
singapore_region_count="$(json_value "data.filter((node) => node.region === '🇸🇬新加坡').length" <"$OUT_DIR/nodes.json")"
clash_http_node_count="$(json_value "data.filter((node) => node.raw_name === '东京 HTTP' && node.protocol === 'https').length" <"$OUT_DIR/nodes.json")"
clash_socks_node_count="$(json_value "data.filter((node) => node.raw_name === '首尔 SOCKS' && node.protocol === 'socks5').length" <"$OUT_DIR/nodes.json")"
clash_anytls_node_count="$(json_value "data.filter((node) => node.raw_name === '香港 AnyTLS' && node.protocol === 'anytls').length" <"$OUT_DIR/nodes.json")"
clash_shadowtls_node_count="$(json_value "data.filter((node) => node.raw_name === '东京 ShadowTLS' && node.protocol === 'shadowtls').length" <"$OUT_DIR/nodes.json")"
clash_naive_node_count="$(json_value "data.filter((node) => node.raw_name === '新加坡 Naive' && node.protocol === 'naive+quic').length" <"$OUT_DIR/nodes.json")"
clash_ssh_node_count="$(json_value "data.filter((node) => node.raw_name === '香港 SSH' && node.protocol === 'ssh').length" <"$OUT_DIR/nodes.json")"
clash_wireguard_node_count="$(json_value "data.filter((node) => node.raw_name === '台北 WireGuard' && node.protocol === 'wireguard').length" <"$OUT_DIR/nodes.json")"
clash_vmess_node_count="$(json_value "data.filter((node) => node.raw_name === '东京 VMess' && node.protocol === 'vmess').length" <"$OUT_DIR/nodes.json")"
first_node_id="$(json_value "data[0]?.id ?? 0" <"$OUT_DIR/nodes.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/nodes/$first_node_id" -o "$OUT_DIR/node-detail.json"
node_detail_id="$(json_value "data.id" <"$OUT_DIR/node-detail.json")"
node_detail_uri="$(json_value "data.uri || ''" <"$OUT_DIR/node-detail.json")"
node_detail_hash="$(json_value "data.uri_hash || ''" <"$OUT_DIR/node-detail.json")"
node_detail_source_name="$(json_value "data.source_name || ''" <"$OUT_DIR/node-detail.json")"
patch_json "/api/nodes/$first_node_id" '{"display_name":"QA 手动节点"}' "$OUT_DIR/node-update.json"
node_manual_name="$(json_value "data.display_name" <"$OUT_DIR/node-update.json")"
node_manual_mode="$(json_value "data.name_mode" <"$OUT_DIR/node-update.json")"
post_json "/api/nodes/$first_node_id/reset-display-name" '{}' "$OUT_DIR/node-reset-name.json"
node_reset_name="$(json_value "data.display_name" <"$OUT_DIR/node-reset-name.json")"
node_reset_mode="$(json_value "data.name_mode" <"$OUT_DIR/node-reset-name.json")"
post_json "/api/virtual-nodes" '{"name":"FluxGate-HK","listen_protocol":"vless","listen_port":8443,"tag_selector":"{\"include\":[\"QA-HK\"]}"}' "$OUT_DIR/virtual-node.json"
post_json "/api/virtual-nodes" '{"name":"FluxGate-SG","listen_protocol":"vless","listen_port":8444}' "$OUT_DIR/virtual-node-sg.json"
virtual_sg_id="$(json_value "data.id" <"$OUT_DIR/virtual-node-sg.json")"
patch_json "/api/virtual-nodes/$virtual_sg_id" '{"name":"FluxGate-SG","listen_protocol":"vless","listen_port":9444,"tag_selector":"{\"include\":[\"SG\"],\"exclude\":[\"Backup\"]}","strategy":"selector","status":"inactive"}' "$OUT_DIR/virtual-node-sg-update.json"
virtual_sg_updated_port="$(json_value "data.listen_port" <"$OUT_DIR/virtual-node-sg-update.json")"
virtual_sg_updated_selector="$(json_value "data.tag_selector" <"$OUT_DIR/virtual-node-sg-update.json")"
virtual_sg_updated_status="$(json_value "data.status" <"$OUT_DIR/virtual-node-sg-update.json")"
post_json "/api/tokens" "{\"user_id\":$user_id,\"name\":\"QA Token\",\"expire_days\":30,\"quota_bytes\":1048576}" "$OUT_DIR/token.json"
token_id="$(json_value "data.token.id" <"$OUT_DIR/token.json")"
plain_token="$(json_value "data.plain_token" <"$OUT_DIR/token.json")"
token_subscription_legacy="$(json_value "data.subscription || ''" <"$OUT_DIR/token.json")"
token_subscription_default="$(json_value "data.subscriptions?.default || ''" <"$OUT_DIR/token.json")"
token_subscription_clash="$(json_value "data.subscriptions?.clash || ''" <"$OUT_DIR/token.json")"
token_subscription_sing_box="$(json_value "data.subscriptions?.sing_box || ''" <"$OUT_DIR/token.json")"
expected_gateway_host="$(node -e 'const u = new URL(process.argv[1]); process.stdout.write(u.hostname)' "$BASE_URL")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/traffic/tokens" -o "$OUT_DIR/traffic-tokens.json"
traffic_token_count="$(json_value "data.length" <"$OUT_DIR/traffic-tokens.json")"
traffic_token_used="$(json_value "data.find((row) => row.token_id === $token_id)?.used_total_bytes ?? -1" <"$OUT_DIR/traffic-tokens.json")"
traffic_token_today="$(json_value "data.find((row) => row.token_id === $token_id)?.today_total_bytes ?? -1" <"$OUT_DIR/traffic-tokens.json")"
traffic_token_month="$(json_value "data.find((row) => row.token_id === $token_id)?.month_total_bytes ?? -1" <"$OUT_DIR/traffic-tokens.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/traffic/daily?days=14" -o "$OUT_DIR/traffic-daily.json"
traffic_daily_count="$(json_value "data.length" <"$OUT_DIR/traffic-daily.json")"
traffic_daily_total="$(json_value "data.reduce((sum, row) => sum + (row.total_bytes || 0), 0)" <"$OUT_DIR/traffic-daily.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/traffic/hourly?hours=24" -o "$OUT_DIR/traffic-hourly.json"
traffic_hourly_count="$(json_value "data.length" <"$OUT_DIR/traffic-hourly.json")"
traffic_hourly_total="$(json_value "data.reduce((sum, row) => sum + (row.total_bytes || 0), 0)" <"$OUT_DIR/traffic-hourly.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/traffic/outbounds?days=14" -o "$OUT_DIR/traffic-outbounds.json"
traffic_outbound_count="$(json_value "data.length" <"$OUT_DIR/traffic-outbounds.json")"
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
curl -fsS -D "$OUT_DIR/subscription.headers" "$BASE_URL/sub/$plain_token?target=clash" -o "$OUT_DIR/subscription.yaml"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/generate" -o "$OUT_DIR/sing-box.json"
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/config/check" -o "$OUT_DIR/sing-box-check.json"
config_check_valid="$(json_value "data.valid" <"$OUT_DIR/sing-box-check.json")"
config_check_hash="$(json_value "data.config_hash" <"$OUT_DIR/sing-box-check.json")"
config_check_upstreams="$(json_value "data.upstream_outbound_count" <"$OUT_DIR/sing-box-check.json")"
run_logged curl -fsS -b "$COOKIE_JAR" "$BASE_URL/api/delivery/readiness" -o "$OUT_DIR/delivery-readiness.json"
delivery_ready="$(json_value "data.ready" <"$OUT_DIR/delivery-readiness.json")"
delivery_ready_count="$(json_value "data.ready_count" <"$OUT_DIR/delivery-readiness.json")"
delivery_total_checks="$(json_value "data.total_checks" <"$OUT_DIR/delivery-readiness.json")"
delivery_config_hash="$(json_value "data.config.config_hash" <"$OUT_DIR/delivery-readiness.json")"
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
run_logged curl -fsS -b "$COOKIE_JAR" -X POST "$BASE_URL/api/sing-box/restart" -o "$OUT_DIR/sing-box-restart.json"
config_restart_enabled="$(json_value "data.enabled" <"$OUT_DIR/sing-box-restart.json")"
config_restart_executed="$(json_value "data.executed" <"$OUT_DIR/sing-box-restart.json")"
config_restart_skipped="$(json_value "data.skipped" <"$OUT_DIR/sing-box-restart.json")"

if [[ "$source_b_prefix" != '"[机场A-2] "' ]]; then
  log "unexpected auto prefix for duplicate source: $source_b_prefix"
  exit 1
fi

if [[ "$team_updated_name" != "QA Team Edited" || "$team_updated_description" != "automated smoke edited" || "$team_updated_status" != "active" ]]; then
  log "team update should persist editable fields: name=$team_updated_name description=$team_updated_description status=$team_updated_status"
  exit 1
fi

if [[ "$user_updated_name" != "QA User Edited" || "$user_updated_email" != "qa-edited@example.test" || "$user_updated_remark" != "QA remark" || "$user_updated_status" != "active" ]]; then
  log "user update should persist editable fields: name=$user_updated_name email=$user_updated_email remark=$user_updated_remark status=$user_updated_status"
  exit 1
fi

if [[ "$source_b_updated_name" != "机场B" || "$source_b_updated_type" != "subscription" || "$source_b_updated_url" != "https://example.test/sub" || "$source_b_updated_prefix" != "[手动B] " || "$source_b_updated_tags" != "SG, Backup" || "$source_b_updated_interval" != "15" ]]; then
  log "source update should persist editable fields: name=$source_b_updated_name type=$source_b_updated_type url=$source_b_updated_url prefix=$source_b_updated_prefix tags=$source_b_updated_tags interval=$source_b_updated_interval"
  exit 1
fi

if [[ "$policy_id" -lt 1 || "$policy_scope_id" != "$team_id" || "$policy_max_nodes" != "5" || "$policy_include_tags" != "[\"QA-HK\"]" || "$policy_allowed_virtual_nodes" != "[\"FluxGate-HK\"]" || "$policy_list_count" -lt 1 ]]; then
  log "policy create/list should work: id=$policy_id scope_id=$policy_scope_id team_id=$team_id include_tags=$policy_include_tags allowed_virtual_nodes=$policy_allowed_virtual_nodes max_nodes=$policy_max_nodes list_count=$policy_list_count"
  exit 1
fi

if [[ "$policy_updated_name" != "QA 默认策略-已调整" || "$policy_updated_status" != "active" || "$policy_updated_exclude_tags" != "[\"Backup\"]" ]]; then
  log "policy update should persist editable fields: name=$policy_updated_name status=$policy_updated_status exclude_tags=$policy_updated_exclude_tags"
  exit 1
fi

if [[ "$virtual_sg_updated_port" != "9444" || "$virtual_sg_updated_selector" != "{\"include\":[\"SG\"],\"exclude\":[\"Backup\"]}" || "$virtual_sg_updated_status" != "inactive" ]]; then
  log "virtual node update should persist editable fields: port=$virtual_sg_updated_port selector=$virtual_sg_updated_selector status=$virtual_sg_updated_status"
  exit 1
fi

if [[ "$source_refresh_imported" != "12" ]]; then
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

if [[ "$singbox_refresh_imported" != "11" ]]; then
  log "unexpected sing-box JSON refresh import count: $singbox_refresh_imported"
  exit 1
fi

if [[ "$qa_hk_tagged_count" -lt 1 ]]; then
  log "source default_tags should be attached to imported nodes"
  exit 1
fi

if [[ "$china_hk_region_count" -lt 1 || "$china_tw_region_count" -lt 1 ]]; then
  log "node region normalization should group Hong Kong and Taiwan: hk=$china_hk_region_count tw=$china_tw_region_count"
  exit 1
fi

if [[ "$us_route_region_count" -lt 1 || "$singapore_region_count" -lt 1 ]]; then
  log "node region normalization should group known countries and route destinations: us_route=$us_route_region_count singapore=$singapore_region_count"
  exit 1
fi

if [[ "$clash_http_node_count" -lt 1 || "$clash_socks_node_count" -lt 1 ]]; then
  log "Clash YAML HTTP/SOCKS proxies should import as active nodes: http=$clash_http_node_count socks=$clash_socks_node_count"
  exit 1
fi

if [[ "$clash_anytls_node_count" -lt 1 || "$clash_shadowtls_node_count" -lt 1 || "$clash_naive_node_count" -lt 1 ]]; then
  log "Clash YAML AnyTLS/ShadowTLS/Naive proxies should import as active nodes: anytls=$clash_anytls_node_count shadowtls=$clash_shadowtls_node_count naive=$clash_naive_node_count"
  exit 1
fi

if [[ "$clash_ssh_node_count" -lt 1 ]]; then
  log "Clash YAML SSH proxy should import as active node: ssh=$clash_ssh_node_count"
  exit 1
fi

if [[ "$clash_wireguard_node_count" -lt 1 ]]; then
  log "Clash YAML WireGuard proxy should import as active node: wireguard=$clash_wireguard_node_count"
  exit 1
fi

if [[ "$clash_vmess_node_count" -lt 1 ]]; then
  log "Clash YAML VMess proxy should import as active node: vmess=$clash_vmess_node_count"
  exit 1
fi

if [[ "$first_node_id" -lt 1 || "$node_manual_name" != "QA 手动节点" || "$node_manual_mode" != "manual" || "$node_reset_mode" != "auto" || "$node_reset_name" == "QA 手动节点" ]]; then
  log "node edit/reset should work: id=$first_node_id manual_name=$node_manual_name manual_mode=$node_manual_mode reset_name=$node_reset_name reset_mode=$node_reset_mode"
  exit 1
fi

if [[ "$node_detail_id" != "$first_node_id" || -z "$node_detail_uri" || -z "$node_detail_hash" || -z "$node_detail_source_name" ]]; then
  log "node detail endpoint should return full node information: id=$node_detail_id expected=$first_node_id uri=$node_detail_uri hash=$node_detail_hash source=$node_detail_source_name"
  exit 1
fi

if [[ "$token_extended_status" != "active" ]]; then
  log "unexpected token status after extend: $token_extended_status"
  exit 1
fi

if [[ "$traffic_token_count" -lt 1 || "$traffic_token_used" != "0" || "$traffic_token_today" != "0" || "$traffic_token_month" != "0" ]]; then
  log "traffic token summary should include new token with zero usage: count=$traffic_token_count used=$traffic_token_used today=$traffic_token_today month=$traffic_token_month"
  exit 1
fi

if [[ "$traffic_daily_count" != "14" || "$traffic_daily_total" != "0" ]]; then
  log "traffic daily chart data should return 14 empty days: count=$traffic_daily_count total=$traffic_daily_total"
  exit 1
fi

if [[ "$traffic_hourly_count" != "24" || "$traffic_hourly_total" != "0" ]]; then
  log "traffic hourly chart data should return 24 empty hours: count=$traffic_hourly_count total=$traffic_hourly_total"
  exit 1
fi

if [[ "$traffic_outbound_count" != "0" ]]; then
  log "traffic outbound summary should be empty before stats samples: count=$traffic_outbound_count"
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

if [[ "$token_subscription_legacy" != "$token_subscription_default" || "$token_subscription_default" != "$BASE_URL/sub/$plain_token" ]]; then
  log "token creation should return a default subscription URL based on the current base URL"
  exit 1
fi

if [[ "$token_subscription_clash" != "$token_subscription_default?target=clash" ]]; then
  log "token creation should return the Clash/Mihomo subscription URL"
  exit 1
fi

if [[ "$token_subscription_sing_box" != "$token_subscription_default?target=sing-box" ]]; then
  log "token creation should return the sing-box subscription URL"
  exit 1
fi

if ! grep -q 'FluxGate-HK' "$OUT_DIR/subscription.yaml"; then
  log "subscription missing FluxGate-HK"
  exit 1
fi

if ! grep -Fq "server: $expected_gateway_host" "$OUT_DIR/subscription.yaml"; then
  log "subscription should expose the current request host as Mihomo gateway server: expected $expected_gateway_host"
  exit 1
fi

if grep -Fq 'gateway.example.com' "$OUT_DIR/subscription.yaml"; then
  log "subscription should not expose placeholder gateway.example.com"
  exit 1
fi

if grep -q 'FluxGate-SG' "$OUT_DIR/subscription.yaml"; then
  log "team policy allowed_virtual_nodes should hide FluxGate-SG from subscription"
  exit 1
fi

if ! header_contains "$OUT_DIR/subscription.headers" '^subscription-userinfo: upload=0; download=0; total=1074790400; expire=[0-9]+$'; then
  log "subscription response should expose standard traffic userinfo header"
  exit 1
fi

if ! header_contains "$OUT_DIR/subscription.headers" '^x-fluxgate-used-bytes: 0$'; then
  log "subscription response should expose FluxGate used bytes header"
  exit 1
fi

if ! header_contains "$OUT_DIR/subscription.headers" '^x-fluxgate-quota-bytes: 1074790400$'; then
  log "subscription response should expose FluxGate quota bytes header"
  exit 1
fi

if ! header_contains "$OUT_DIR/subscription.headers" '^x-fluxgate-remaining-bytes: 1074790400$'; then
  log "subscription response should expose FluxGate remaining bytes header"
  exit 1
fi

if grep -q 'FluxGate-SG' "$OUT_DIR/sing-box.json"; then
  log "team policy allowed_virtual_nodes should hide FluxGate-SG from sing-box config"
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

if [[ "$delivery_ready" != "true" || "$delivery_ready_count" != "$delivery_total_checks" || "$delivery_config_hash" != "$config_check_hash" ]]; then
  log "delivery readiness should match usable config: ready=$delivery_ready ready_count=$delivery_ready_count total=$delivery_total_checks delivery_hash=$delivery_config_hash check_hash=$config_check_hash"
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

if [[ "$config_restart_enabled" != "false" || "$config_restart_executed" != "false" || "$config_restart_skipped" != "true" ]]; then
  log "sing-box restart should be disabled by default: enabled=$config_restart_enabled executed=$config_restart_executed skipped=$config_restart_skipped"
  exit 1
fi

if ! grep -q '"tag": "up_' "$OUT_DIR/sing-box.json"; then
  log "active upstream node should appear as sing-box outbound"
  exit 1
fi

if ! grep -q '"tag": "vn-FluxGate-HK-upstreams"' "$OUT_DIR/sing-box.json"; then
  log "virtual node tag_selector should create dedicated upstream selector"
  exit 1
fi

policy_route_count="$(json_value "(data.route.rules || []).filter((rule) => rule.outbound === 'vn-FluxGate-HK-token-$token_id-upstreams' && (rule.auth_user || []).includes('$gateway_user')).length" <"$OUT_DIR/sing-box.json")"
if [[ "$policy_route_count" != "1" ]]; then
  log "policy include_tags should create auth_user route rule to token selector"
  exit 1
fi

if ! grep -q '"outbound": "vn-FluxGate-HK-upstreams"' "$OUT_DIR/sing-box.json"; then
  log "virtual node tag_selector should create route rule to dedicated selector"
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
