package singbox

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

func TestBuildConfigFiltersUnusableGatewayTokens(t *testing.T) {
	now := time.Date(2026, 4, 29, 1, 10, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "active-user"),
		gatewayToken("active", "active", "vless", &future, 100, 10, 20, "future-user"),
		gatewayToken("revoked", "active", "vless", nil, 0, 0, 0, "revoked-user"),
		gatewayToken("active", "revoked", "vless", nil, 0, 0, 0, "revoked-account-user"),
		gatewayToken("active", "active", "vless", &past, 0, 0, 0, "expired-user"),
		gatewayToken("active", "active", "vless", nil, 100, 40, 60, "quota-user"),
		gatewayToken("active", "active", "trojan", nil, 0, 0, 0, "unsupported-user"),
	}, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, nil, nil, now)

	if len(config.Inbounds) != 1 {
		t.Fatalf("expected one inbound, got %+v", config.Inbounds)
	}
	users := config.Inbounds[0].Users
	if len(users) != 2 {
		t.Fatalf("expected two usable users, got %+v", users)
	}
	if users[0].Name != "active-user" || users[1].Name != "future-user" {
		t.Fatalf("unexpected users in generated config: %+v", users)
	}

	stats, ok := config.Experimental["v2ray_api"].(map[string]any)["stats"].(map[string]any)
	if !ok {
		t.Fatalf("missing v2ray stats config: %+v", config.Experimental)
	}
	statUsers, ok := stats["users"].([]string)
	if !ok {
		t.Fatalf("missing stat users: %+v", stats)
	}
	if len(statUsers) != 2 || statUsers[0] != "active-user" || statUsers[1] != "future-user" {
		t.Fatalf("unexpected stat users: %+v", statUsers)
	}
}

func TestBuildConfigAddsSupportedUpstreamOutbounds(t *testing.T) {
	ssCredential := base64.RawURLEncoding.EncodeToString([]byte("aes-128-gcm:qa-placeholder"))
	vmessURI := vmessURI(t, map[string]any{
		"add":           "vmess.example.net",
		"port":          "443",
		"id":            "00000000-0000-0000-0000-000000000046",
		"aid":           "0",
		"scy":           "auto",
		"net":           "ws",
		"host":          "ws.example.net",
		"path":          "/ws",
		"tls":           "tls",
		"sni":           "vmess.example.net",
		"allowInsecure": true,
		"disable_sni":   "1",
		"alpn":          "h2,http/1.1",
		"ps":            "vmess",
	})
	config := buildConfig(nil, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, []store.Node{
		{
			ID:         42,
			URI:        "vless://00000000-0000-0000-0000-000000000042@example.com:443?security=tls&sni=edge.example.com&flow=xtls-rprx-vision&packet_encoding=xudp&insecure=1&disable_sni=1&alpn=h2,http/1.1&type=ws&path=%2Fvless&host=ws.example.com#hk",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         43,
			URI:        "vless://00000000-0000-0000-0000-000000000043@example.net:443#inactive",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "inactive",
		},
		{
			ID:         44,
			URI:        "trojan://qa-placeholder@example.org:443?security=tls&sni=trojan.example.org&skip-cert-verify=1&disable-sni=1&alpn=h2&type=ws&path=%2Ftrojan&host=ws.trojan.example.org#trojan",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         45,
			URI:        "ss://" + ssCredential + "@example.net:8388?plugin=v2ray-plugin&plugin_opts=mode%3Dwebsocket%3Bhost%3Dss.example.net&network=tcp#ss",
			Protocol:   "ss",
			ServerPort: 8388,
			Status:     "active",
		},
		{
			ID:         46,
			URI:        vmessURI,
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         47,
			URI:        "hysteria2://qa-placeholder@example.dev:443?obfs=salamander&obfs-password=obfs-placeholder&up_mbps=40&down_mbps=160&sni=hy2.example.dev&skip-cert-verify=1&disable-sni=1&pinSHA256=hy2-pin-placeholder&fp=chrome#hy2",
			Protocol:   "hysteria2",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         48,
			URI:        "tuic://00000000-0000-0000-0000-000000000048:qa-placeholder@example.io:443?congestion_control=bbr&udp_relay_mode=native&sni=tuic.example.io&alpn=h3&insecure=1&fp=chrome#tuic",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         49,
			URI:        "anytls://qa-placeholder@example.chat:443?sni=anytls.example.chat&alpn=h2,http/1.1&idle_session_check_interval=20s&idle_session_timeout=45s&min_idle_session=2&insecure=1&fp=chrome#anytls",
			Protocol:   "anytls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         50,
			URI:        "shadowtls://qa-placeholder@example.help:443?version=3&sni=shadow.example.help&alpn=h2&insecure=1&fp=chrome#shadowtls",
			Protocol:   "shadowtls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         51,
			URI:        "naive://qa-user:qa-placeholder@example.news:443?sni=naive.example.news&quic=1&quic_congestion_control=bbr&udp_over_tcp=1&insecure_concurrency=2&alpn=h3&insecure=1&disable_sni=1&fp=chrome#naive",
			Protocol:   "naive",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         52,
			URI:        "hysteria://qa-placeholder@example.zone:443?auth_str=qa-auth&up_mbps=20&down_mbps=80&obfs=obfs-placeholder&recv_window_conn=1048576&recv_window=2097152&disable_mtu_discovery=1&protocol=udp&sni=hysteria.example.zone&alpn=h3&insecure=1&fp=chrome#hysteria",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         53,
			URI:        "https://qa-user:qa-placeholder@example.proxy:8443/connect?sni=http-proxy.example.proxy&skip-cert-verify=1&disable_sni=1&alpn=h2,http%2F1.1&fp=chrome#http",
			Protocol:   "https",
			ServerPort: 8443,
			Status:     "active",
		},
		{
			ID:         54,
			URI:        "socks5://qa-user:qa-placeholder@example.socks:1080?udp=1&udp_over_tcp=1#socks",
			Protocol:   "socks5",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         55,
			URI:        "ssh://qa-user:qa-placeholder@example.ssh:22?private_key_path=keys%2Fqa_id_ed25519&host_key_algorithms=ssh-ed25519,rsa-sha2-512&client_version=SSH-2.0-FluxGateQA&cipher=aes128-gcm@openssh.com,chacha20-poly1305@openssh.com&mac=hmac-sha2-256&kex_algorithm=curve25519-sha256#ssh",
			Protocol:   "ssh",
			ServerPort: 22,
			Status:     "active",
		},
		{
			ID:         56,
			URI:        "wireguard://example.wg:51820?private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=&peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==&pre_shared_key=cHNrLXBsYWNlaG9sZGVy&local_address=10.66.0.2/32,fd00::2/128&allowed_ips=0.0.0.0/0,::/0&reserved=1,2,3&workers=2&mtu=1420&network=udp&system_interface=1&interface_name=wg-qa#wireguard",
			Protocol:   "wireguard",
			ServerPort: 51820,
			Status:     "active",
		},
		{
			ID:         57,
			URI:        "tor://default?executable_path=/usr/bin/tor&extra_args=--quiet,--SocksPort,auto&data_directory=cache%2Ftor&torrc.ClientOnly=1&torrc.SocksPort=auto#tor",
			Protocol:   "tor",
			ServerPort: 0,
			Status:     "active",
		},
		{
			ID:         58,
			URI:        "unknown://default#unsupported",
			Protocol:   "unknown",
			ServerPort: 0,
			Status:     "active",
		},
	}, nil, time.Date(2026, 4, 29, 1, 17, 0, 0, time.UTC))

	if config.Route["final"] != upstreamSelectorTag {
		t.Fatalf("expected route final to selector, got %+v", config.Route)
	}
	stats, ok := config.Experimental["v2ray_api"].(map[string]any)["stats"].(map[string]any)
	if !ok {
		t.Fatalf("missing v2ray stats config: %+v", config.Experimental)
	}
	statOutbounds, ok := stats["outbounds"].([]string)
	if !ok {
		t.Fatalf("missing stat outbounds: %+v", stats)
	}
	if len(statOutbounds) != 15 || statOutbounds[0] != "up_42" || statOutbounds[len(statOutbounds)-1] != "up_57" {
		t.Fatalf("unexpected stat outbounds: %+v", statOutbounds)
	}
	vless := findOutbound(config.Outbounds, "up_42")
	if vless == nil {
		t.Fatalf("expected vless outbound up_42, got %+v", config.Outbounds)
	}
	if vless["server"] != "example.com" || vless["server_port"] != 443 {
		t.Fatalf("unexpected vless server fields: %+v", vless)
	}
	if vless["uuid"] != "00000000-0000-0000-0000-000000000042" || vless["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected vless auth fields: %+v", vless)
	}
	if vless["packet_encoding"] != "xudp" {
		t.Fatalf("unexpected vless packet encoding: %+v", vless)
	}
	vlessTLS, ok := vless["tls"].(map[string]any)
	if !ok || vlessTLS["enabled"] != true || vlessTLS["server_name"] != "edge.example.com" || vlessTLS["insecure"] != true || vlessTLS["disable_sni"] != true {
		t.Fatalf("unexpected tls config: %+v", vless["tls"])
	}
	vlessALPN, ok := vlessTLS["alpn"].([]string)
	if !ok || len(vlessALPN) != 2 || vlessALPN[0] != "h2" || vlessALPN[1] != "http/1.1" {
		t.Fatalf("unexpected vless alpn config: %+v", vlessTLS["alpn"])
	}
	vlessTransport, ok := vless["transport"].(map[string]any)
	if !ok || vlessTransport["type"] != "ws" || vlessTransport["path"] != "/vless" {
		t.Fatalf("unexpected vless transport config: %+v", vless["transport"])
	}
	vlessTransportHeaders, ok := vlessTransport["headers"].(map[string]any)
	if !ok || vlessTransportHeaders["Host"] != "ws.example.com" {
		t.Fatalf("unexpected vless transport headers: %+v", vlessTransport["headers"])
	}
	trojan := findOutbound(config.Outbounds, "up_44")
	if trojan == nil {
		t.Fatalf("expected trojan outbound up_44, got %+v", config.Outbounds)
	}
	if trojan["type"] != "trojan" || trojan["server"] != "example.org" || trojan["password"] != "qa-placeholder" {
		t.Fatalf("unexpected trojan outbound fields: %+v", trojan)
	}
	trojanTLS, ok := trojan["tls"].(map[string]any)
	if !ok || trojanTLS["enabled"] != true || trojanTLS["server_name"] != "trojan.example.org" || trojanTLS["insecure"] != true || trojanTLS["disable_sni"] != true {
		t.Fatalf("unexpected trojan tls config: %+v", trojan["tls"])
	}
	trojanALPN, ok := trojanTLS["alpn"].([]string)
	if !ok || len(trojanALPN) != 1 || trojanALPN[0] != "h2" {
		t.Fatalf("unexpected trojan alpn config: %+v", trojanTLS["alpn"])
	}
	trojanTransport, ok := trojan["transport"].(map[string]any)
	if !ok || trojanTransport["type"] != "ws" || trojanTransport["path"] != "/trojan" {
		t.Fatalf("unexpected trojan transport config: %+v", trojan["transport"])
	}
	trojanTransportHeaders, ok := trojanTransport["headers"].(map[string]any)
	if !ok || trojanTransportHeaders["Host"] != "ws.trojan.example.org" {
		t.Fatalf("unexpected trojan transport headers: %+v", trojanTransport["headers"])
	}
	shadowsocks := findOutbound(config.Outbounds, "up_45")
	if shadowsocks == nil {
		t.Fatalf("expected shadowsocks outbound up_45, got %+v", config.Outbounds)
	}
	if shadowsocks["type"] != "shadowsocks" || shadowsocks["server"] != "example.net" || shadowsocks["server_port"] != 8388 {
		t.Fatalf("unexpected shadowsocks server fields: %+v", shadowsocks)
	}
	if shadowsocks["method"] != "aes-128-gcm" || shadowsocks["password"] != "qa-placeholder" {
		t.Fatalf("unexpected shadowsocks auth fields: %+v", shadowsocks)
	}
	if shadowsocks["plugin"] != "v2ray-plugin" || shadowsocks["plugin_opts"] != "mode=websocket;host=ss.example.net" || shadowsocks["network"] != "tcp" {
		t.Fatalf("unexpected shadowsocks plugin fields: %+v", shadowsocks)
	}
	vmess := findOutbound(config.Outbounds, "up_46")
	if vmess == nil {
		t.Fatalf("expected vmess outbound up_46, got %+v", config.Outbounds)
	}
	if vmess["type"] != "vmess" || vmess["server"] != "vmess.example.net" || vmess["server_port"] != 443 {
		t.Fatalf("unexpected vmess server fields: %+v", vmess)
	}
	if vmess["uuid"] != "00000000-0000-0000-0000-000000000046" || vmess["security"] != "auto" {
		t.Fatalf("unexpected vmess auth fields: %+v", vmess)
	}
	vmessTLS, ok := vmess["tls"].(map[string]any)
	if !ok || vmessTLS["enabled"] != true || vmessTLS["server_name"] != "vmess.example.net" || vmessTLS["insecure"] != true || vmessTLS["disable_sni"] != true {
		t.Fatalf("unexpected vmess tls config: %+v", vmess["tls"])
	}
	vmessALPN, ok := vmessTLS["alpn"].([]string)
	if !ok || len(vmessALPN) != 2 || vmessALPN[0] != "h2" || vmessALPN[1] != "http/1.1" {
		t.Fatalf("unexpected vmess alpn config: %+v", vmessTLS["alpn"])
	}
	transport, ok := vmess["transport"].(map[string]any)
	if !ok || transport["type"] != "ws" || transport["path"] != "/ws" {
		t.Fatalf("unexpected vmess transport config: %+v", vmess["transport"])
	}
	hysteria2 := findOutbound(config.Outbounds, "up_47")
	if hysteria2 == nil {
		t.Fatalf("expected hysteria2 outbound up_47, got %+v", config.Outbounds)
	}
	if hysteria2["type"] != "hysteria2" || hysteria2["server"] != "example.dev" || hysteria2["server_port"] != 443 {
		t.Fatalf("unexpected hysteria2 server fields: %+v", hysteria2)
	}
	if hysteria2["password"] != "qa-placeholder" {
		t.Fatalf("unexpected hysteria2 auth fields: %+v", hysteria2)
	}
	if hysteria2["up_mbps"] != 40 || hysteria2["down_mbps"] != 160 {
		t.Fatalf("unexpected hysteria2 bandwidth fields: %+v", hysteria2)
	}
	hysteria2TLS, ok := hysteria2["tls"].(map[string]any)
	if !ok || hysteria2TLS["enabled"] != true || hysteria2TLS["server_name"] != "hy2.example.dev" || hysteria2TLS["insecure"] != true || hysteria2TLS["disable_sni"] != true {
		t.Fatalf("unexpected hysteria2 tls config: %+v", hysteria2["tls"])
	}
	certificatePins, ok := hysteria2TLS["certificate_public_key_sha256"].([]string)
	if !ok || len(certificatePins) != 1 || certificatePins[0] != "hy2-pin-placeholder" {
		t.Fatalf("unexpected hysteria2 certificate pin config: %+v", hysteria2TLS["certificate_public_key_sha256"])
	}
	hysteria2UTLS, ok := hysteria2TLS["utls"].(map[string]any)
	if !ok || hysteria2UTLS["enabled"] != true || hysteria2UTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected hysteria2 utls config: %+v", hysteria2TLS["utls"])
	}
	obfs, ok := hysteria2["obfs"].(map[string]any)
	if !ok || obfs["type"] != "salamander" || obfs["password"] != "obfs-placeholder" {
		t.Fatalf("unexpected hysteria2 obfs config: %+v", hysteria2["obfs"])
	}
	tuic := findOutbound(config.Outbounds, "up_48")
	if tuic == nil {
		t.Fatalf("expected tuic outbound up_48, got %+v", config.Outbounds)
	}
	if tuic["type"] != "tuic" || tuic["server"] != "example.io" || tuic["server_port"] != 443 {
		t.Fatalf("unexpected tuic server fields: %+v", tuic)
	}
	if tuic["uuid"] != "00000000-0000-0000-0000-000000000048" || tuic["password"] != "qa-placeholder" {
		t.Fatalf("unexpected tuic auth fields: %+v", tuic)
	}
	if tuic["congestion_control"] != "bbr" || tuic["udp_relay_mode"] != "native" {
		t.Fatalf("unexpected tuic relay fields: %+v", tuic)
	}
	tuicTLS, ok := tuic["tls"].(map[string]any)
	if !ok || tuicTLS["enabled"] != true || tuicTLS["server_name"] != "tuic.example.io" || tuicTLS["insecure"] != true {
		t.Fatalf("unexpected tuic tls config: %+v", tuic["tls"])
	}
	alpn, ok := tuicTLS["alpn"].([]string)
	if !ok || len(alpn) != 1 || alpn[0] != "h3" {
		t.Fatalf("unexpected tuic alpn config: %+v", tuicTLS["alpn"])
	}
	tuicUTLS, ok := tuicTLS["utls"].(map[string]any)
	if !ok || tuicUTLS["enabled"] != true || tuicUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected tuic utls config: %+v", tuicTLS["utls"])
	}
	anytls := findOutbound(config.Outbounds, "up_49")
	if anytls == nil {
		t.Fatalf("expected anytls outbound up_49, got %+v", config.Outbounds)
	}
	if anytls["type"] != "anytls" || anytls["server"] != "example.chat" || anytls["server_port"] != 443 {
		t.Fatalf("unexpected anytls server fields: %+v", anytls)
	}
	if anytls["password"] != "qa-placeholder" {
		t.Fatalf("unexpected anytls auth fields: %+v", anytls)
	}
	if anytls["idle_session_check_interval"] != "20s" || anytls["idle_session_timeout"] != "45s" || anytls["min_idle_session"] != 2 {
		t.Fatalf("unexpected anytls idle session fields: %+v", anytls)
	}
	anytlsTLS, ok := anytls["tls"].(map[string]any)
	if !ok || anytlsTLS["enabled"] != true || anytlsTLS["server_name"] != "anytls.example.chat" || anytlsTLS["insecure"] != true {
		t.Fatalf("unexpected anytls tls config: %+v", anytls["tls"])
	}
	anytlsALPN, ok := anytlsTLS["alpn"].([]string)
	if !ok || len(anytlsALPN) != 2 || anytlsALPN[0] != "h2" || anytlsALPN[1] != "http/1.1" {
		t.Fatalf("unexpected anytls alpn config: %+v", anytlsTLS["alpn"])
	}
	anytlsUTLS, ok := anytlsTLS["utls"].(map[string]any)
	if !ok || anytlsUTLS["enabled"] != true || anytlsUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected anytls utls config: %+v", anytlsTLS["utls"])
	}
	shadowtls := findOutbound(config.Outbounds, "up_50")
	if shadowtls == nil {
		t.Fatalf("expected shadowtls outbound up_50, got %+v", config.Outbounds)
	}
	if shadowtls["type"] != "shadowtls" || shadowtls["server"] != "example.help" || shadowtls["server_port"] != 443 {
		t.Fatalf("unexpected shadowtls server fields: %+v", shadowtls)
	}
	if shadowtls["version"] != 3 || shadowtls["password"] != "qa-placeholder" {
		t.Fatalf("unexpected shadowtls auth fields: %+v", shadowtls)
	}
	shadowtlsTLS, ok := shadowtls["tls"].(map[string]any)
	if !ok || shadowtlsTLS["enabled"] != true || shadowtlsTLS["server_name"] != "shadow.example.help" || shadowtlsTLS["insecure"] != true {
		t.Fatalf("unexpected shadowtls tls config: %+v", shadowtls["tls"])
	}
	shadowtlsALPN, ok := shadowtlsTLS["alpn"].([]string)
	if !ok || len(shadowtlsALPN) != 1 || shadowtlsALPN[0] != "h2" {
		t.Fatalf("unexpected shadowtls alpn config: %+v", shadowtlsTLS["alpn"])
	}
	shadowtlsUTLS, ok := shadowtlsTLS["utls"].(map[string]any)
	if !ok || shadowtlsUTLS["enabled"] != true || shadowtlsUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected shadowtls utls config: %+v", shadowtlsTLS["utls"])
	}
	naive := findOutbound(config.Outbounds, "up_51")
	if naive == nil {
		t.Fatalf("expected naive outbound up_51, got %+v", config.Outbounds)
	}
	if naive["type"] != "naive" || naive["server"] != "example.news" || naive["server_port"] != 443 {
		t.Fatalf("unexpected naive server fields: %+v", naive)
	}
	if naive["username"] != "qa-user" || naive["password"] != "qa-placeholder" {
		t.Fatalf("unexpected naive auth fields: %+v", naive)
	}
	if naive["quic"] != true || naive["udp_over_tcp"] != true || naive["quic_congestion_control"] != "bbr" || naive["insecure_concurrency"] != 2 {
		t.Fatalf("unexpected naive transport fields: %+v", naive)
	}
	naiveTLS, ok := naive["tls"].(map[string]any)
	if !ok || naiveTLS["enabled"] != true || naiveTLS["server_name"] != "naive.example.news" || naiveTLS["insecure"] != true || naiveTLS["disable_sni"] != true {
		t.Fatalf("unexpected naive tls config: %+v", naive["tls"])
	}
	naiveALPN, ok := naiveTLS["alpn"].([]string)
	if !ok || len(naiveALPN) != 1 || naiveALPN[0] != "h3" {
		t.Fatalf("unexpected naive alpn config: %+v", naiveTLS["alpn"])
	}
	naiveUTLS, ok := naiveTLS["utls"].(map[string]any)
	if !ok || naiveUTLS["enabled"] != true || naiveUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected naive utls config: %+v", naiveTLS["utls"])
	}
	hysteria := findOutbound(config.Outbounds, "up_52")
	if hysteria == nil {
		t.Fatalf("expected hysteria outbound up_52, got %+v", config.Outbounds)
	}
	if hysteria["type"] != "hysteria" || hysteria["server"] != "example.zone" || hysteria["server_port"] != 443 {
		t.Fatalf("unexpected hysteria server fields: %+v", hysteria)
	}
	if hysteria["auth_str"] != "qa-auth" || hysteria["up_mbps"] != 20 || hysteria["down_mbps"] != 80 {
		t.Fatalf("unexpected hysteria auth or bandwidth fields: %+v", hysteria)
	}
	if hysteria["obfs"] != "obfs-placeholder" || hysteria["network"] != "udp" {
		t.Fatalf("unexpected hysteria transport fields: %+v", hysteria)
	}
	if hysteria["recv_window_conn"] != 1048576 || hysteria["recv_window"] != 2097152 || hysteria["disable_mtu_discovery"] != true {
		t.Fatalf("unexpected hysteria window fields: %+v", hysteria)
	}
	hysteriaTLS, ok := hysteria["tls"].(map[string]any)
	if !ok || hysteriaTLS["enabled"] != true || hysteriaTLS["server_name"] != "hysteria.example.zone" || hysteriaTLS["insecure"] != true {
		t.Fatalf("unexpected hysteria tls config: %+v", hysteria["tls"])
	}
	hysteriaALPN, ok := hysteriaTLS["alpn"].([]string)
	if !ok || len(hysteriaALPN) != 1 || hysteriaALPN[0] != "h3" {
		t.Fatalf("unexpected hysteria alpn config: %+v", hysteriaTLS["alpn"])
	}
	hysteriaUTLS, ok := hysteriaTLS["utls"].(map[string]any)
	if !ok || hysteriaUTLS["enabled"] != true || hysteriaUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected hysteria utls config: %+v", hysteriaTLS["utls"])
	}
	httpProxy := findOutbound(config.Outbounds, "up_53")
	if httpProxy == nil {
		t.Fatalf("expected http outbound up_53, got %+v", config.Outbounds)
	}
	if httpProxy["type"] != "http" || httpProxy["server"] != "example.proxy" || httpProxy["server_port"] != 8443 {
		t.Fatalf("unexpected http server fields: %+v", httpProxy)
	}
	if httpProxy["username"] != "qa-user" || httpProxy["password"] != "qa-placeholder" || httpProxy["path"] != "/connect" {
		t.Fatalf("unexpected http auth or path fields: %+v", httpProxy)
	}
	httpTLS, ok := httpProxy["tls"].(map[string]any)
	if !ok || httpTLS["enabled"] != true || httpTLS["server_name"] != "http-proxy.example.proxy" || httpTLS["insecure"] != true || httpTLS["disable_sni"] != true {
		t.Fatalf("unexpected http tls config: %+v", httpProxy["tls"])
	}
	httpALPN, ok := httpTLS["alpn"].([]string)
	if !ok || len(httpALPN) != 2 || httpALPN[0] != "h2" || httpALPN[1] != "http/1.1" {
		t.Fatalf("unexpected http alpn config: %+v", httpTLS["alpn"])
	}
	httpUTLS, ok := httpTLS["utls"].(map[string]any)
	if !ok || httpUTLS["enabled"] != true || httpUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected http utls config: %+v", httpTLS["utls"])
	}
	socks := findOutbound(config.Outbounds, "up_54")
	if socks == nil {
		t.Fatalf("expected socks outbound up_54, got %+v", config.Outbounds)
	}
	if socks["type"] != "socks" || socks["server"] != "example.socks" || socks["server_port"] != 1080 {
		t.Fatalf("unexpected socks server fields: %+v", socks)
	}
	if socks["version"] != "5" || socks["username"] != "qa-user" || socks["password"] != "qa-placeholder" {
		t.Fatalf("unexpected socks auth fields: %+v", socks)
	}
	if socks["network"] != "udp" || socks["udp_over_tcp"] != true {
		t.Fatalf("unexpected socks network fields: %+v", socks)
	}
	ssh := findOutbound(config.Outbounds, "up_55")
	if ssh == nil {
		t.Fatalf("expected ssh outbound up_55, got %+v", config.Outbounds)
	}
	if ssh["type"] != "ssh" || ssh["server"] != "example.ssh" || ssh["server_port"] != 22 {
		t.Fatalf("unexpected ssh server fields: %+v", ssh)
	}
	if ssh["user"] != "qa-user" || ssh["password"] != "qa-placeholder" || ssh["private_key_path"] != "keys/qa_id_ed25519" {
		t.Fatalf("unexpected ssh auth fields: %+v", ssh)
	}
	if ssh["client_version"] != "SSH-2.0-FluxGateQA" {
		t.Fatalf("unexpected ssh client version: %+v", ssh)
	}
	hostKeyAlgorithms, ok := ssh["host_key_algorithms"].([]string)
	if !ok || len(hostKeyAlgorithms) != 2 || hostKeyAlgorithms[0] != "ssh-ed25519" || hostKeyAlgorithms[1] != "rsa-sha2-512" {
		t.Fatalf("unexpected ssh host key algorithms: %+v", ssh["host_key_algorithms"])
	}
	cipher, ok := ssh["cipher"].([]string)
	if !ok || len(cipher) != 2 || cipher[0] != "aes128-gcm@openssh.com" || cipher[1] != "chacha20-poly1305@openssh.com" {
		t.Fatalf("unexpected ssh cipher list: %+v", ssh["cipher"])
	}
	mac, ok := ssh["mac"].([]string)
	if !ok || len(mac) != 1 || mac[0] != "hmac-sha2-256" {
		t.Fatalf("unexpected ssh mac list: %+v", ssh["mac"])
	}
	kexAlgorithm, ok := ssh["kex_algorithm"].([]string)
	if !ok || len(kexAlgorithm) != 1 || kexAlgorithm[0] != "curve25519-sha256" {
		t.Fatalf("unexpected ssh kex algorithms: %+v", ssh["kex_algorithm"])
	}
	wireguard := findOutbound(config.Outbounds, "up_56")
	if wireguard == nil {
		t.Fatalf("expected wireguard outbound up_56, got %+v", config.Outbounds)
	}
	if wireguard["type"] != "wireguard" || wireguard["server"] != "example.wg" || wireguard["server_port"] != 51820 {
		t.Fatalf("unexpected wireguard server fields: %+v", wireguard)
	}
	if wireguard["private_key"] != "cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=" || wireguard["peer_public_key"] != "cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==" || wireguard["pre_shared_key"] != "cHNrLXBsYWNlaG9sZGVy" {
		t.Fatalf("unexpected wireguard key fields: %+v", wireguard)
	}
	localAddress, ok := wireguard["local_address"].([]string)
	if !ok || len(localAddress) != 2 || localAddress[0] != "10.66.0.2/32" || localAddress[1] != "fd00::2/128" {
		t.Fatalf("unexpected wireguard local address: %+v", wireguard["local_address"])
	}
	if wireguard["system_interface"] != true || wireguard["interface_name"] != "wg-qa" || wireguard["workers"] != 2 || wireguard["mtu"] != 1420 || wireguard["network"] != "udp" {
		t.Fatalf("unexpected wireguard interface fields: %+v", wireguard)
	}
	reserved, ok := wireguard["reserved"].([]int)
	if !ok || len(reserved) != 3 || reserved[0] != 1 || reserved[1] != 2 || reserved[2] != 3 {
		t.Fatalf("unexpected wireguard reserved bytes: %+v", wireguard["reserved"])
	}
	peers, ok := wireguard["peers"].([]map[string]any)
	if !ok || len(peers) != 1 {
		t.Fatalf("unexpected wireguard peers: %+v", wireguard["peers"])
	}
	peerAllowedIPs, ok := peers[0]["allowed_ips"].([]string)
	if !ok || len(peerAllowedIPs) != 2 || peerAllowedIPs[0] != "0.0.0.0/0" || peerAllowedIPs[1] != "::/0" {
		t.Fatalf("unexpected wireguard peer allowed ips: %+v", peers[0]["allowed_ips"])
	}
	if peers[0]["server"] != "example.wg" || peers[0]["server_port"] != 51820 || peers[0]["public_key"] != "cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==" || peers[0]["pre_shared_key"] != "cHNrLXBsYWNlaG9sZGVy" {
		t.Fatalf("unexpected wireguard peer fields: %+v", peers[0])
	}
	peerReserved, ok := peers[0]["reserved"].([]int)
	if !ok || len(peerReserved) != 3 || peerReserved[0] != 1 || peerReserved[1] != 2 || peerReserved[2] != 3 {
		t.Fatalf("unexpected wireguard peer reserved bytes: %+v", peers[0]["reserved"])
	}
	tor := findOutbound(config.Outbounds, "up_57")
	if tor == nil {
		t.Fatalf("expected tor outbound up_57, got %+v", config.Outbounds)
	}
	if tor["type"] != "tor" || tor["executable_path"] != "/usr/bin/tor" || tor["data_directory"] != "cache/tor" {
		t.Fatalf("unexpected tor base fields: %+v", tor)
	}
	extraArgs, ok := tor["extra_args"].([]string)
	if !ok || len(extraArgs) != 3 || extraArgs[0] != "--quiet" || extraArgs[1] != "--SocksPort" || extraArgs[2] != "auto" {
		t.Fatalf("unexpected tor extra args: %+v", tor["extra_args"])
	}
	torrc, ok := tor["torrc"].(map[string]any)
	if !ok || torrc["ClientOnly"] != 1 || torrc["SocksPort"] != "auto" {
		t.Fatalf("unexpected torrc fields: %+v", tor["torrc"])
	}
	if findOutbound(config.Outbounds, "up_43") != nil || findOutbound(config.Outbounds, "up_58") != nil {
		t.Fatalf("inactive or unsupported nodes should be skipped: %+v", config.Outbounds)
	}

	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 15 || tags[0] != "up_42" || tags[1] != "up_44" || tags[2] != "up_45" || tags[3] != "up_46" || tags[4] != "up_47" || tags[5] != "up_48" || tags[6] != "up_49" || tags[7] != "up_50" || tags[8] != "up_51" || tags[9] != "up_52" || tags[10] != "up_53" || tags[11] != "up_54" || tags[12] != "up_55" || tags[13] != "up_56" || tags[14] != "up_57" || selector["default"] != "up_42" {
		t.Fatalf("unexpected selector outbounds: %+v", selector)
	}
}

func TestBuildConfigAddsDNSOutboundWithoutSelectingIt(t *testing.T) {
	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "dns-user"),
	}, []store.VirtualNode{
		{Name: "dns-only", ListenProtocol: "vless", ListenPort: 8443, TagSelector: `{"include":["DNS"]}`, Status: "active"},
	}, []store.Node{
		{
			ID:         61,
			URI:        "dns://default#DNS",
			Protocol:   "dns",
			ServerPort: 0,
			Status:     "active",
			Tags:       []string{"DNS"},
		},
		{
			ID:         62,
			URI:        "vless://00000000-0000-0000-0000-000000000062@example.com:443#hk",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	}, nil, time.Date(2026, 4, 29, 16, 38, 0, 0, time.UTC))

	dns := findOutbound(config.Outbounds, "up_61")
	if dns == nil || dns["type"] != "dns" {
		t.Fatalf("expected dns outbound up_61, got %+v", config.Outbounds)
	}
	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 1 || tags[0] != "up_62" || selector["default"] != "up_62" {
		t.Fatalf("dns outbound should not enter traffic selector: %+v", selector)
	}
	stats, ok := config.Experimental["v2ray_api"].(map[string]any)["stats"].(map[string]any)
	if !ok {
		t.Fatalf("missing v2ray stats config: %+v", config.Experimental)
	}
	statOutbounds, ok := stats["outbounds"].([]string)
	if !ok || len(statOutbounds) != 1 || statOutbounds[0] != "up_62" {
		t.Fatalf("dns outbound should not enter stats outbounds: %+v", stats)
	}
	if virtualSelector := findOutbound(config.Outbounds, "vn-dns-only-upstreams"); virtualSelector != nil {
		t.Fatalf("dns outbound should not create a virtual upstream selector: %+v", virtualSelector)
	}
	rules, ok := config.Route["rules"].([]map[string]any)
	if !ok || len(rules) != 1 || rules[0]["outbound"] != "block" {
		t.Fatalf("dns-only virtual selector should route normal traffic to block: %+v", config.Route)
	}
}

func TestBuildConfigAddsDirectOutboundAndSelectsIt(t *testing.T) {
	config := buildConfig(nil, nil, []store.Node{
		{
			ID:         63,
			URI:        "freedom://default#Direct",
			Protocol:   "freedom",
			ServerPort: 0,
			Status:     "active",
		},
	}, nil, time.Date(2026, 4, 29, 16, 50, 0, 0, time.UTC))

	direct := findOutbound(config.Outbounds, "up_63")
	if direct == nil || direct["type"] != "direct" {
		t.Fatalf("expected direct outbound up_63, got %+v", config.Outbounds)
	}
	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 1 || tags[0] != "up_63" || selector["default"] != "up_63" {
		t.Fatalf("direct outbound should enter traffic selector: %+v", selector)
	}
	stats, ok := config.Experimental["v2ray_api"].(map[string]any)["stats"].(map[string]any)
	if !ok {
		t.Fatalf("missing v2ray stats config: %+v", config.Experimental)
	}
	statOutbounds, ok := stats["outbounds"].([]string)
	if !ok || len(statOutbounds) != 1 || statOutbounds[0] != "up_63" {
		t.Fatalf("direct outbound should enter stats outbounds: %+v", stats)
	}
}

func TestBuildConfigAddsBlockOutboundWithoutSelectingIt(t *testing.T) {
	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "block-user"),
	}, []store.VirtualNode{
		{Name: "block-only", ListenProtocol: "vless", ListenPort: 8443, TagSelector: `{"include":["BLOCK"]}`, Status: "active"},
	}, []store.Node{
		{
			ID:         64,
			URI:        "reject-drop://default#Block",
			Protocol:   "reject-drop",
			ServerPort: 0,
			Status:     "active",
			Tags:       []string{"BLOCK"},
		},
		{
			ID:         65,
			URI:        "vless://00000000-0000-0000-0000-000000000065@example.com:443#hk",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	}, nil, time.Date(2026, 4, 29, 16, 58, 0, 0, time.UTC))

	block := findOutbound(config.Outbounds, "up_64")
	if block == nil || block["type"] != "block" {
		t.Fatalf("expected block outbound up_64, got %+v", config.Outbounds)
	}
	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 1 || tags[0] != "up_65" || selector["default"] != "up_65" {
		t.Fatalf("block outbound should not enter traffic selector: %+v", selector)
	}
	stats, ok := config.Experimental["v2ray_api"].(map[string]any)["stats"].(map[string]any)
	if !ok {
		t.Fatalf("missing v2ray stats config: %+v", config.Experimental)
	}
	statOutbounds, ok := stats["outbounds"].([]string)
	if !ok || len(statOutbounds) != 1 || statOutbounds[0] != "up_65" {
		t.Fatalf("block outbound should not enter stats outbounds: %+v", stats)
	}
	if virtualSelector := findOutbound(config.Outbounds, "vn-block-only-upstreams"); virtualSelector != nil {
		t.Fatalf("block outbound should not create a virtual upstream selector: %+v", virtualSelector)
	}
	rules, ok := config.Route["rules"].([]map[string]any)
	if !ok || len(rules) != 1 || rules[0]["outbound"] != "block" {
		t.Fatalf("block-only virtual selector should route normal traffic to built-in block: %+v", config.Route)
	}
}

func TestBuildConfigSupportsCanonicalProtocolAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         96,
			URI:        "shadowsocks://alias-ss.example:8388?method=aes-128-gcm&password=ss-alias-placeholder#ss-alias",
			Protocol:   "shadowsocks",
			ServerPort: 8388,
			Status:     "active",
		},
		{
			ID:         97,
			URI:        "trojan-go://trojan-go-placeholder@alias-trojan.example:443?security=tls&sni=alias-trojan.example#trojan-go-alias",
			Protocol:   "trojan-go",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         98,
			URI:        "vmess-aead://00000000-0000-0000-0000-000000000098@alias-vmess.example:443?tls=1&type=ws&path=%2Fws&host=ws.alias-vmess.example#vmess-aead-alias",
			Protocol:   "vmess-aead",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         99,
			URI:        "any-tls://anytls-alias-placeholder@alias-anytls.example:443?sni=alias-anytls.example#any-tls-alias",
			Protocol:   "any-tls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         100,
			URI:        "shadow-tls://shadowtls-alias-placeholder@alias-shadowtls.example:443?version=3&sni=alias-shadowtls.example#shadow-tls-alias",
			Protocol:   "shadow-tls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         101,
			URI:        "naive-quic://qa-user:naive-alias-placeholder@alias-naive.example:443?serverName=alias-naive.example#naive-quic-alias",
			Protocol:   "naive-quic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         102,
			URI:        "naive-https://qa-user:naive-https-placeholder@alias-naive-https.example:443?serverName=alias-naive-https.example#naive-https-alias",
			Protocol:   "naive-https",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         103,
			URI:        "tuic-v5://00000000-0000-0000-0000-000000000103:tuic-v5-placeholder@alias-tuic-v5.example:443?congestionControl=bbr&udpRelayMode=native&sni=alias-tuic-v5.example#tuic-v5-alias",
			Protocol:   "tuic-v5",
			ServerPort: 443,
			Status:     "active",
		},
	})

	shadowsocks := findOutbound(config.Outbounds, "up_96")
	if shadowsocks == nil || shadowsocks["type"] != "shadowsocks" || shadowsocks["method"] != "aes-128-gcm" || shadowsocks["password"] != "ss-alias-placeholder" {
		t.Fatalf("unexpected shadowsocks alias outbound: %+v", shadowsocks)
	}
	trojan := findOutbound(config.Outbounds, "up_97")
	if trojan == nil || trojan["type"] != "trojan" || trojan["password"] != "trojan-go-placeholder" {
		t.Fatalf("unexpected trojan-go alias outbound: %+v", trojan)
	}
	vmess := findOutbound(config.Outbounds, "up_98")
	if vmess == nil || vmess["type"] != "vmess" || vmess["uuid"] != "00000000-0000-0000-0000-000000000098" {
		t.Fatalf("unexpected vmess-aead alias outbound: %+v", vmess)
	}
	vmessTransport, ok := vmess["transport"].(map[string]any)
	if !ok || vmessTransport["type"] != "ws" || vmessTransport["path"] != "/ws" {
		t.Fatalf("unexpected vmess-aead alias transport: %+v", vmess["transport"])
	}
	anytls := findOutbound(config.Outbounds, "up_99")
	if anytls == nil || anytls["type"] != "anytls" || anytls["password"] != "anytls-alias-placeholder" {
		t.Fatalf("unexpected any-tls alias outbound: %+v", anytls)
	}
	shadowtls := findOutbound(config.Outbounds, "up_100")
	if shadowtls == nil || shadowtls["type"] != "shadowtls" || shadowtls["version"] != 3 || shadowtls["password"] != "shadowtls-alias-placeholder" {
		t.Fatalf("unexpected shadow-tls alias outbound: %+v", shadowtls)
	}
	naive := findOutbound(config.Outbounds, "up_101")
	if naive == nil || naive["type"] != "naive" || naive["username"] != "qa-user" || naive["password"] != "naive-alias-placeholder" || naive["quic"] != true {
		t.Fatalf("unexpected naive-quic alias outbound: %+v", naive)
	}
	naiveHTTPS := findOutbound(config.Outbounds, "up_102")
	if naiveHTTPS == nil || naiveHTTPS["type"] != "naive" || naiveHTTPS["username"] != "qa-user" || naiveHTTPS["password"] != "naive-https-placeholder" || naiveHTTPS["quic"] == true {
		t.Fatalf("unexpected naive-https alias outbound: %+v", naiveHTTPS)
	}
	naiveHTTPSTLS, ok := naiveHTTPS["tls"].(map[string]any)
	if !ok || naiveHTTPSTLS["enabled"] != true || naiveHTTPSTLS["server_name"] != "alias-naive-https.example" {
		t.Fatalf("unexpected naive-https alias tls: %+v", naiveHTTPS["tls"])
	}
	tuic := findOutbound(config.Outbounds, "up_103")
	if tuic == nil || tuic["type"] != "tuic" || tuic["uuid"] != "00000000-0000-0000-0000-000000000103" || tuic["password"] != "tuic-v5-placeholder" || tuic["congestion_control"] != "bbr" || tuic["udp_relay_mode"] != "native" {
		t.Fatalf("unexpected tuic-v5 alias outbound: %+v", tuic)
	}
	tuicTLS, ok := tuic["tls"].(map[string]any)
	if !ok || tuicTLS["enabled"] != true || tuicTLS["server_name"] != "alias-tuic-v5.example" {
		t.Fatalf("unexpected tuic-v5 alias tls: %+v", tuic["tls"])
	}
}

func TestBuildConfigAppliesVirtualNodePolicyToUsers(t *testing.T) {
	now := time.Date(2026, 4, 29, 5, 10, 0, 0, time.UTC)
	teamID := int64(10)
	policyScopeID := teamID
	token := gatewayToken("active", "active", "vless", nil, 0, 0, 0, "team-user")
	token.ID = 30
	token.UserID = 20
	token.UserTeamID = &teamID

	config := buildConfig([]store.TokenWithAccount{token}, []store.VirtualNode{
		{ID: 1, Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
		{ID: 2, Name: "sg", ListenProtocol: "vless", ListenPort: 8444, Status: "active"},
	}, nil, []store.Policy{
		{ScopeType: "team", ScopeID: &policyScopeID, MaxNodes: 1, Status: "active"},
	}, now)

	if len(config.Inbounds) != 1 {
		t.Fatalf("expected one allowed inbound, got %+v", config.Inbounds)
	}
	if len(config.Inbounds[0].Users) != 1 || config.Inbounds[0].Users[0].Name != "team-user" {
		t.Fatalf("first virtual node should contain user: %+v", config.Inbounds[0].Users)
	}
	if config.Inbounds[0].Tag != "vn-hk" {
		t.Fatalf("unexpected allowed inbound: %+v", config.Inbounds[0])
	}
}

func TestBuildConfigRoutesVirtualNodeByTagSelector(t *testing.T) {
	now := time.Date(2026, 4, 29, 5, 35, 0, 0, time.UTC)
	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "active-user"),
	}, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, TagSelector: `{"include":["HK"],"exclude":["Backup"]}`, Status: "active"},
	}, []store.Node{
		{ID: 1, URI: "vless://00000000-0000-0000-0000-000000000001@hk.example:443#hk", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"HK", "Primary"}},
		{ID: 2, URI: "vless://00000000-0000-0000-0000-000000000002@backup.example:443#backup", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"HK", "Backup"}},
		{ID: 3, URI: "vless://00000000-0000-0000-0000-000000000003@sg.example:443#sg", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"SG"}},
	}, nil, now)

	selector := findOutbound(config.Outbounds, "vn-hk-upstreams")
	if selector == nil {
		t.Fatalf("expected virtual node upstream selector, got %+v", config.Outbounds)
	}
	selected, ok := selector["outbounds"].([]string)
	if !ok || len(selected) != 1 || selected[0] != "up_1" || selector["default"] != "up_1" {
		t.Fatalf("unexpected virtual node selected outbounds: %+v", selector)
	}
	rules, ok := config.Route["rules"].([]map[string]any)
	if !ok || len(rules) != 1 {
		t.Fatalf("expected one virtual node route rule, got %+v", config.Route)
	}
	if rules[0]["outbound"] != "vn-hk-upstreams" {
		t.Fatalf("unexpected route rule: %+v", rules[0])
	}
}

func TestBuildConfigRoutesPolicyTagsByAuthUser(t *testing.T) {
	now := time.Date(2026, 4, 29, 5, 46, 0, 0, time.UTC)
	teamID := int64(10)
	scopeID := teamID
	token := gatewayToken("active", "active", "vless", nil, 0, 0, 0, "team-user")
	token.ID = 30
	token.UserID = 20
	token.UserTeamID = &teamID

	config := buildConfig([]store.TokenWithAccount{token}, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, TagSelector: `{"include":["HK"]}`, Status: "active"},
	}, []store.Node{
		{ID: 1, URI: "vless://00000000-0000-0000-0000-000000000001@hk.example:443#hk", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"HK"}},
		{ID: 2, URI: "vless://00000000-0000-0000-0000-000000000002@premium.example:443#premium", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"HK", "Premium"}},
		{ID: 3, URI: "vless://00000000-0000-0000-0000-000000000003@sg.example:443#sg", Protocol: "vless", ServerPort: 443, Status: "active", Tags: []string{"SG", "Premium"}},
	}, []store.Policy{
		{ScopeType: "team", ScopeID: &scopeID, IncludeTags: `["Premium"]`, Status: "active"},
	}, now)

	selector := findOutbound(config.Outbounds, "vn-hk-token-30-upstreams")
	if selector == nil {
		t.Fatalf("expected token policy selector, got %+v", config.Outbounds)
	}
	selected, ok := selector["outbounds"].([]string)
	if !ok || len(selected) != 1 || selected[0] != "up_2" {
		t.Fatalf("policy selector should intersect virtual node tags, got %+v", selector)
	}
	rules, ok := config.Route["rules"].([]map[string]any)
	if !ok || len(rules) < 2 {
		t.Fatalf("expected policy and virtual node route rules, got %+v", config.Route)
	}
	authUsers, ok := rules[0]["auth_user"].([]string)
	if !ok || len(authUsers) != 1 || authUsers[0] != "team-user" || rules[0]["outbound"] != "vn-hk-token-30-upstreams" {
		t.Fatalf("first route rule should target auth_user policy selector: %+v", rules[0])
	}
	if rules[1]["outbound"] != "vn-hk-upstreams" {
		t.Fatalf("virtual node fallback rule should follow policy rule: %+v", rules)
	}
}

func TestBuildConfigPreservesVLESSGRPCTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         70,
			URI:        "vless://00000000-0000-0000-0000-000000000070@example.grpc:443?security=tls&type=grpc&service_name=fluxgate#grpc",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_70")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_70, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "grpc" || transport["service_name"] != "fluxgate" {
		t.Fatalf("unexpected vless grpc transport config: %+v", outbound["transport"])
	}
}

func TestBuildConfigPreservesVLESSGRPCKeepaliveOptions(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         77,
			URI:        "vless://00000000-0000-0000-0000-000000000077@example.grpc:443?security=tls&type=grpc&grpcServiceName=fluxgate&idle_timeout=30s&ping_timeout=10s&permit_without_stream=1&grpcMultiMode=1#grpc-keepalive",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_77")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_77, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "grpc" || transport["service_name"] != "fluxgate" || transport["idle_timeout"] != "30s" || transport["ping_timeout"] != "10s" || transport["permit_without_stream"] != true || transport["multi_mode"] != true {
		t.Fatalf("unexpected vless grpc keepalive config: %+v", outbound["transport"])
	}
}

func TestBuildConfigPreservesVLESSQUICTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         78,
			URI:        "vless://00000000-0000-0000-0000-000000000078@example.quic:443?security=tls&type=quic#quic",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_78")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_78, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "quic" {
		t.Fatalf("unexpected vless quic transport config: %+v", outbound["transport"])
	}
}

func TestBuildConfigPreservesVLESSWebSocketEarlyData(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         76,
			URI:        "vless://00000000-0000-0000-0000-000000000076@example.ws:443?security=tls&type=ws&path=/ws&host=ws.example.test&max_early_data=2048&early_data_header_name=Sec-WebSocket-Protocol#ws-early",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_76")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_76, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "ws" || transport["path"] != "/ws" || transport["max_early_data"] != 2048 || transport["early_data_header_name"] != "Sec-WebSocket-Protocol" {
		t.Fatalf("unexpected vless websocket early data config: %+v", outbound["transport"])
	}
	headers, ok := transport["headers"].(map[string]any)
	if !ok || headers["Host"] != "ws.example.test" {
		t.Fatalf("unexpected vless websocket headers: %+v", transport["headers"])
	}
}

func TestBuildConfigPreservesVLESSHTTPTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         74,
			URI:        "vless://00000000-0000-0000-0000-000000000074@example.http:443?security=tls&type=http&httpHost=h2.example.test,h2-backup.example.test&httpPath=/h2&httpMethod=GET&httpIdleTimeout=20s&httpPingTimeout=10s#http",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_74")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_74, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "http" || transport["path"] != "/h2" || transport["method"] != "GET" || transport["idle_timeout"] != "20s" || transport["ping_timeout"] != "10s" {
		t.Fatalf("unexpected vless http transport config: %+v", outbound["transport"])
	}
	hosts, ok := transport["host"].([]string)
	if !ok || len(hosts) != 2 || hosts[0] != "h2.example.test" || hosts[1] != "h2-backup.example.test" {
		t.Fatalf("unexpected vless http transport hosts: %+v", transport["host"])
	}
}

func TestBuildConfigPreservesTrojanHTTPUpgradeTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         75,
			URI:        "trojan://trojan-placeholder@example.upgrade:443?security=tls&type=httpupgrade&httpUpgradeHost=upgrade.example.test&httpUpgradePath=/upgrade#upgrade",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_75")
	if outbound == nil {
		t.Fatalf("expected trojan outbound up_75, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "httpupgrade" || transport["host"] != "upgrade.example.test" || transport["path"] != "/upgrade" {
		t.Fatalf("unexpected trojan httpupgrade transport config: %+v", outbound["transport"])
	}
}

func TestBuildConfigPreservesURITransportHostAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         175,
			URI:        "vless://00000000-0000-0000-0000-000000000175@ws-authority.example:443?security=tls&type=ws&path=/ws&authority=ws-authority.example.test#ws-authority",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         176,
			URI:        "vless://00000000-0000-0000-0000-000000000176@http-authority.example:443?security=tls&type=http&headers.Host=h2-authority.example.test,h2-backup.example.test&httpPath=/h2#http-authority",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         177,
			URI:        "trojan://trojan-placeholder@upgrade-authority.example:443?security=tls&type=httpupgrade&headers.%3Aauthority=upgrade-authority.example.test&httpUpgradePath=/upgrade#upgrade-authority",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         178,
			URI:        "vmess://00000000-0000-0000-0000-000000000178@vmess-header-host.example:443?encryption=auto&security=tls&type=ws&path=/vmess&headerHost=vmess-header-host.example.test&sni=vmess-header-host.example#vmess-header-host",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	wsOutbound := findOutbound(config.Outbounds, "up_175")
	if wsOutbound == nil {
		t.Fatalf("expected vless outbound up_175, got %+v", config.Outbounds)
	}
	wsTransport, ok := wsOutbound["transport"].(map[string]any)
	if !ok || wsTransport["type"] != "ws" || wsTransport["path"] != "/ws" {
		t.Fatalf("unexpected websocket authority transport: %+v", wsOutbound["transport"])
	}
	wsHeaders, ok := wsTransport["headers"].(map[string]any)
	if !ok || wsHeaders["Host"] != "ws-authority.example.test" {
		t.Fatalf("unexpected websocket authority headers: %+v", wsTransport["headers"])
	}

	httpOutbound := findOutbound(config.Outbounds, "up_176")
	if httpOutbound == nil {
		t.Fatalf("expected vless outbound up_176, got %+v", config.Outbounds)
	}
	httpTransport, ok := httpOutbound["transport"].(map[string]any)
	if !ok || httpTransport["type"] != "http" || httpTransport["path"] != "/h2" {
		t.Fatalf("unexpected http authority transport: %+v", httpOutbound["transport"])
	}
	httpHosts, ok := httpTransport["host"].([]string)
	if !ok || len(httpHosts) != 2 || httpHosts[0] != "h2-authority.example.test" || httpHosts[1] != "h2-backup.example.test" {
		t.Fatalf("unexpected http authority hosts: %+v", httpTransport["host"])
	}

	upgradeOutbound := findOutbound(config.Outbounds, "up_177")
	if upgradeOutbound == nil {
		t.Fatalf("expected trojan outbound up_177, got %+v", config.Outbounds)
	}
	upgradeTransport, ok := upgradeOutbound["transport"].(map[string]any)
	if !ok || upgradeTransport["type"] != "httpupgrade" || upgradeTransport["host"] != "upgrade-authority.example.test" || upgradeTransport["path"] != "/upgrade" {
		t.Fatalf("unexpected httpupgrade authority transport: %+v", upgradeOutbound["transport"])
	}

	vmessOutbound := findOutbound(config.Outbounds, "up_178")
	if vmessOutbound == nil {
		t.Fatalf("expected vmess outbound up_178, got %+v", config.Outbounds)
	}
	vmessTransport, ok := vmessOutbound["transport"].(map[string]any)
	if !ok || vmessTransport["type"] != "ws" || vmessTransport["path"] != "/vmess" {
		t.Fatalf("unexpected vmess header host transport: %+v", vmessOutbound["transport"])
	}
	vmessHeaders, ok := vmessTransport["headers"].(map[string]any)
	if !ok || vmessHeaders["Host"] != "vmess-header-host.example.test" {
		t.Fatalf("unexpected vmess header host headers: %+v", vmessTransport["headers"])
	}
}

func TestBuildConfigPreservesVLESSRealityTLS(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         72,
			URI:        "vless://00000000-0000-0000-0000-000000000072@reality.vless.example:443?security=reality&sni=www.example.test&flow=xtls-rprx-vision&pbk=reality-public-key-placeholder&sid=a1b2c3d4&fp=chrome#reality",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_72")
	if outbound == nil {
		t.Fatalf("expected vless outbound up_72, got %+v", config.Outbounds)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "www.example.test" {
		t.Fatalf("unexpected vless reality tls config: %+v", outbound["tls"])
	}
	reality, ok := tls["reality"].(map[string]any)
	if !ok || reality["enabled"] != true || reality["public_key"] != "reality-public-key-placeholder" || reality["short_id"] != "a1b2c3d4" {
		t.Fatalf("unexpected vless reality config: %+v", tls["reality"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected vless reality utls config: %+v", tls["utls"])
	}
	if outbound["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected vless reality flow: %+v", outbound)
	}
}

func TestBuildConfigPreservesTrojanRealityTLS(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         73,
			URI:        "trojan://trojan-placeholder@reality.trojan.example:443?security=reality&sni=www.example.test&pbk=trojan-reality-public-key&sid=b1c2d3e4&fp=chrome#trojan-reality",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_73")
	if outbound == nil {
		t.Fatalf("expected trojan outbound up_73, got %+v", config.Outbounds)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "www.example.test" {
		t.Fatalf("unexpected trojan reality tls config: %+v", outbound["tls"])
	}
	reality, ok := tls["reality"].(map[string]any)
	if !ok || reality["enabled"] != true || reality["public_key"] != "trojan-reality-public-key" || reality["short_id"] != "b1c2d3e4" {
		t.Fatalf("unexpected trojan reality config: %+v", tls["reality"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected trojan reality utls config: %+v", tls["utls"])
	}
}

func TestBuildConfigSupportsVLESSTrojanTLSQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         212,
			URI:        "vless://00000000-0000-0000-0000-000000000212@vless-tls-alias.example:443?tlsEnabled=1#vless-tls-alias",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         213,
			URI:        "trojan://trojan-placeholder@trojan-tls-alias.example:443?enableTLS=true#trojan-tls-alias",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         214,
			URI:        "vless://00000000-0000-0000-0000-000000000214@vless-over-tls.example:443?overTLS=1#vless-over-tls",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         215,
			URI:        "vless://00000000-0000-0000-0000-000000000215@vless-tls-server-name.example:443?tls=true&tlsServerName=vless-tls-name.example#vless-tls-server-name",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         216,
			URI:        "trojan://trojan-placeholder@trojan-server-name.example:443?tls=tls&server-name=trojan-server-name.example#trojan-server-name",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         217,
			URI:        "vless://00000000-0000-0000-0000-000000000217@vless-skip-cert.example:443?tls=true&skipCertVerify=1#vless-skip-cert",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         218,
			URI:        "trojan://trojan-placeholder@trojan-allow-insecure.example:443?tls=true&allow-insecure=true#trojan-allow-insecure",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         219,
			URI:        "vless://00000000-0000-0000-0000-000000000219@vless-tls-host.example:443?tls=true&tlsHost=vless-host-alias.example#vless-tls-host",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         220,
			URI:        "trojan://trojan-placeholder@trojan-tls-host.example:443?tls=true&tls-host=trojan-host-alias.example#trojan-tls-host",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         221,
			URI:        "vless://00000000-0000-0000-0000-000000000221@vless-disable-sni.example:443?tls=true&disableSni=true#vless-disable-sni",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         222,
			URI:        "trojan://trojan-placeholder@trojan-tls-disable-sni.example:443?tls=true&tls-disable-sni=1#trojan-tls-disable-sni",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
	})

	vlessOutbound := findOutbound(config.Outbounds, "up_212")
	if vlessOutbound == nil {
		t.Fatalf("expected vless outbound up_212, got %+v", config.Outbounds)
	}
	vlessTLS, ok := vlessOutbound["tls"].(map[string]any)
	if !ok || vlessTLS["enabled"] != true || vlessTLS["server_name"] != "vless-tls-alias.example" {
		t.Fatalf("unexpected vless TLS query aliases: %+v", vlessOutbound["tls"])
	}

	trojanOutbound := findOutbound(config.Outbounds, "up_213")
	if trojanOutbound == nil {
		t.Fatalf("expected trojan outbound up_213, got %+v", config.Outbounds)
	}
	trojanTLS, ok := trojanOutbound["tls"].(map[string]any)
	if !ok || trojanTLS["enabled"] != true || trojanTLS["server_name"] != "trojan-tls-alias.example" {
		t.Fatalf("unexpected trojan TLS query aliases: %+v", trojanOutbound["tls"])
	}

	overTLSOutbound := findOutbound(config.Outbounds, "up_214")
	if overTLSOutbound == nil {
		t.Fatalf("expected vless outbound up_214, got %+v", config.Outbounds)
	}
	overTLS, ok := overTLSOutbound["tls"].(map[string]any)
	if !ok || overTLS["enabled"] != true || overTLS["server_name"] != "vless-over-tls.example" {
		t.Fatalf("unexpected vless overTLS query alias: %+v", overTLSOutbound["tls"])
	}

	vlessServerNameOutbound := findOutbound(config.Outbounds, "up_215")
	if vlessServerNameOutbound == nil {
		t.Fatalf("expected vless outbound up_215, got %+v", config.Outbounds)
	}
	vlessServerNameTLS, ok := vlessServerNameOutbound["tls"].(map[string]any)
	if !ok || vlessServerNameTLS["enabled"] != true || vlessServerNameTLS["server_name"] != "vless-tls-name.example" {
		t.Fatalf("unexpected vless tlsServerName query alias: %+v", vlessServerNameOutbound["tls"])
	}

	trojanServerNameOutbound := findOutbound(config.Outbounds, "up_216")
	if trojanServerNameOutbound == nil {
		t.Fatalf("expected trojan outbound up_216, got %+v", config.Outbounds)
	}
	trojanServerNameTLS, ok := trojanServerNameOutbound["tls"].(map[string]any)
	if !ok || trojanServerNameTLS["enabled"] != true || trojanServerNameTLS["server_name"] != "trojan-server-name.example" {
		t.Fatalf("unexpected trojan server-name query alias: %+v", trojanServerNameOutbound["tls"])
	}

	vlessSkipCertOutbound := findOutbound(config.Outbounds, "up_217")
	if vlessSkipCertOutbound == nil {
		t.Fatalf("expected vless outbound up_217, got %+v", config.Outbounds)
	}
	vlessSkipCertTLS, ok := vlessSkipCertOutbound["tls"].(map[string]any)
	if !ok || vlessSkipCertTLS["enabled"] != true || vlessSkipCertTLS["insecure"] != true {
		t.Fatalf("unexpected vless skipCertVerify query alias: %+v", vlessSkipCertOutbound["tls"])
	}

	trojanAllowInsecureOutbound := findOutbound(config.Outbounds, "up_218")
	if trojanAllowInsecureOutbound == nil {
		t.Fatalf("expected trojan outbound up_218, got %+v", config.Outbounds)
	}
	trojanAllowInsecureTLS, ok := trojanAllowInsecureOutbound["tls"].(map[string]any)
	if !ok || trojanAllowInsecureTLS["enabled"] != true || trojanAllowInsecureTLS["insecure"] != true {
		t.Fatalf("unexpected trojan allow-insecure query alias: %+v", trojanAllowInsecureOutbound["tls"])
	}

	vlessTLSHostOutbound := findOutbound(config.Outbounds, "up_219")
	if vlessTLSHostOutbound == nil {
		t.Fatalf("expected vless outbound up_219, got %+v", config.Outbounds)
	}
	vlessTLSHostTLS, ok := vlessTLSHostOutbound["tls"].(map[string]any)
	if !ok || vlessTLSHostTLS["enabled"] != true || vlessTLSHostTLS["server_name"] != "vless-host-alias.example" {
		t.Fatalf("unexpected vless tlsHost query alias: %+v", vlessTLSHostOutbound["tls"])
	}

	trojanTLSHostOutbound := findOutbound(config.Outbounds, "up_220")
	if trojanTLSHostOutbound == nil {
		t.Fatalf("expected trojan outbound up_220, got %+v", config.Outbounds)
	}
	trojanTLSHostTLS, ok := trojanTLSHostOutbound["tls"].(map[string]any)
	if !ok || trojanTLSHostTLS["enabled"] != true || trojanTLSHostTLS["server_name"] != "trojan-host-alias.example" {
		t.Fatalf("unexpected trojan tls-host query alias: %+v", trojanTLSHostOutbound["tls"])
	}

	vlessDisableSNIOutbound := findOutbound(config.Outbounds, "up_221")
	if vlessDisableSNIOutbound == nil {
		t.Fatalf("expected vless outbound up_221, got %+v", config.Outbounds)
	}
	vlessDisableSNITLS, ok := vlessDisableSNIOutbound["tls"].(map[string]any)
	if !ok || vlessDisableSNITLS["enabled"] != true || vlessDisableSNITLS["disable_sni"] != true {
		t.Fatalf("unexpected vless disableSni query alias: %+v", vlessDisableSNIOutbound["tls"])
	}

	trojanTLSDisableSNIOutbound := findOutbound(config.Outbounds, "up_222")
	if trojanTLSDisableSNIOutbound == nil {
		t.Fatalf("expected trojan outbound up_222, got %+v", config.Outbounds)
	}
	trojanTLSDisableSNITLS, ok := trojanTLSDisableSNIOutbound["tls"].(map[string]any)
	if !ok || trojanTLSDisableSNITLS["enabled"] != true || trojanTLSDisableSNITLS["disable_sni"] != true {
		t.Fatalf("unexpected trojan tls-disable-sni query alias: %+v", trojanTLSDisableSNIOutbound["tls"])
	}
}

func TestBuildConfigPreservesVMessGRPCTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID: 71,
			URI: vmessURI(t, map[string]any{
				"add":             "vmess.grpc.example",
				"port":            "443",
				"id":              "00000000-0000-0000-0000-000000000071",
				"aid":             "0",
				"scy":             "auto",
				"net":             "grpc",
				"grpcServiceName": "fluxgate-vmess",
				"grpcMultiMode":   "1",
				"tls":             "tls",
				"sni":             "vmess.grpc.example",
				"ps":              "vmess-grpc",
			}),
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_71")
	if outbound == nil {
		t.Fatalf("expected vmess outbound up_71, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "grpc" || transport["service_name"] != "fluxgate-vmess" || transport["multi_mode"] != true {
		t.Fatalf("unexpected vmess grpc transport config: %+v", outbound["transport"])
	}
}

func TestBuildConfigPreservesVMessHTTPTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID: 82,
			URI: vmessURI(t, map[string]any{
				"add":  "vmess.http.example",
				"port": "443",
				"id":   "00000000-0000-0000-0000-000000000082",
				"aid":  "0",
				"scy":  "auto",
				"net":  "tcp",
				"type": "http",
				"host": "h2.vmess.example,h2-backup.vmess.example",
				"path": "/h2",
				"tls":  "tls",
				"sni":  "vmess.http.example",
				"ps":   "vmess-http",
			}),
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_82")
	if outbound == nil {
		t.Fatalf("expected vmess outbound up_82, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "http" || transport["path"] != "/h2" {
		t.Fatalf("unexpected vmess http transport config: %+v", outbound["transport"])
	}
	hosts, ok := transport["host"].([]string)
	if !ok || len(hosts) != 2 || hosts[0] != "h2.vmess.example" || hosts[1] != "h2-backup.vmess.example" {
		t.Fatalf("unexpected vmess http transport hosts: %+v", transport["host"])
	}
}

func TestBuildConfigPreservesVMessHTTPUpgradeTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID: 182,
			URI: vmessURI(t, map[string]any{
				"add":  "vmess.upgrade.example",
				"port": "443",
				"id":   "00000000-0000-0000-0000-000000000182",
				"aid":  "0",
				"scy":  "auto",
				"net":  "httpupgrade",
				"host": "upgrade.vmess.example",
				"path": "/upgrade",
				"tls":  "tls",
				"sni":  "vmess.upgrade.example",
				"ps":   "vmess-upgrade",
			}),
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         183,
			URI:        "vmess://00000000-0000-0000-0000-000000000183@userinfo-upgrade.vmess.example:443?encryption=auto&security=tls&type=httpupgrade&httpUpgradeHost=userinfo-upgrade.vmess.example&httpUpgradePath=%2Fuserinfo-upgrade&sni=userinfo-upgrade.vmess.example#VMess%20Userinfo%20Upgrade",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_182")
	if outbound == nil {
		t.Fatalf("expected vmess outbound up_182, got %+v", config.Outbounds)
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "httpupgrade" || transport["host"] != "upgrade.vmess.example" || transport["path"] != "/upgrade" {
		t.Fatalf("unexpected vmess httpupgrade transport config: %+v", outbound["transport"])
	}

	userinfoOutbound := findOutbound(config.Outbounds, "up_183")
	if userinfoOutbound == nil {
		t.Fatalf("expected vmess outbound up_183, got %+v", config.Outbounds)
	}
	userinfoTransport, ok := userinfoOutbound["transport"].(map[string]any)
	if !ok || userinfoTransport["type"] != "httpupgrade" || userinfoTransport["host"] != "userinfo-upgrade.vmess.example" || userinfoTransport["path"] != "/userinfo-upgrade" {
		t.Fatalf("unexpected vmess userinfo httpupgrade transport config: %+v", userinfoOutbound["transport"])
	}
}

func TestBuildConfigSupportsVMessUserinfoURI(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         72,
			URI:        "vmess://00000000-0000-0000-0000-000000000072@userinfo.vmess.example:443?encryption=auto&security=tls&type=ws&path=%2Fvmess&host=ws.vmess.example&sni=sni.vmess.example&alpn=h2,http%2F1.1&insecure=1&disable_sni=1&fp=chrome&packet_encoding=packetaddr#VMess%20Userinfo",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_72")
	if outbound == nil {
		t.Fatalf("expected vmess outbound up_72, got %+v", config.Outbounds)
	}
	if outbound["type"] != "vmess" || outbound["server"] != "userinfo.vmess.example" || outbound["server_port"] != 443 {
		t.Fatalf("unexpected vmess userinfo server fields: %+v", outbound)
	}
	if outbound["uuid"] != "00000000-0000-0000-0000-000000000072" || outbound["security"] != "auto" {
		t.Fatalf("unexpected vmess userinfo auth fields: %+v", outbound)
	}
	if outbound["packet_encoding"] != "packetaddr" {
		t.Fatalf("unexpected vmess userinfo packet encoding: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "sni.vmess.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected vmess userinfo tls config: %+v", outbound["tls"])
	}
	alpn, ok := tls["alpn"].([]string)
	if !ok || len(alpn) != 2 || alpn[0] != "h2" || alpn[1] != "http/1.1" {
		t.Fatalf("unexpected vmess userinfo alpn config: %+v", tls["alpn"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected vmess userinfo utls config: %+v", tls["utls"])
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "ws" || transport["path"] != "/vmess" {
		t.Fatalf("unexpected vmess userinfo transport config: %+v", outbound["transport"])
	}
	headers, ok := transport["headers"].(map[string]any)
	if !ok || headers["Host"] != "ws.vmess.example" {
		t.Fatalf("unexpected vmess userinfo transport headers: %+v", transport["headers"])
	}
}

func TestBuildConfigSupportsVMessQueryCredentials(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         88,
			URI:        "vmess://query-vmess.example:443?id=00000000-0000-0000-0000-000000000088&encryption=auto&security=tls&type=grpc&grpcServiceName=query-vmess&grpcMultiMode=1&sni=query-vmess.example&packet_encoding=packetaddr#vmess-query",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_88")
	if outbound == nil {
		t.Fatalf("expected vmess outbound up_88, got %+v", config.Outbounds)
	}
	if outbound["type"] != "vmess" || outbound["server"] != "query-vmess.example" || outbound["server_port"] != 443 {
		t.Fatalf("unexpected vmess query credential server fields: %+v", outbound)
	}
	if outbound["uuid"] != "00000000-0000-0000-0000-000000000088" || outbound["security"] != "auto" || outbound["packet_encoding"] != "packetaddr" {
		t.Fatalf("unexpected vmess query credential auth fields: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "query-vmess.example" {
		t.Fatalf("unexpected vmess query credential tls: %+v", outbound["tls"])
	}
	transport, ok := outbound["transport"].(map[string]any)
	if !ok || transport["type"] != "grpc" || transport["service_name"] != "query-vmess" || transport["multi_mode"] != true {
		t.Fatalf("unexpected vmess query credential transport: %+v", outbound["transport"])
	}
}

func TestBuildConfigSupportsVMessUserinfoTLSQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         210,
			URI:        "vmess://00000000-0000-0000-0000-000000000210@vmess-tls-alias.example:443?encryption=auto&tlsEnabled=1&serverName=alias.vmess.example&disableSNI=1#vmess-tls-alias",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         211,
			URI:        "vmess://00000000-0000-0000-0000-000000000211@vmess-tls-server-name.example:443?encryption=auto&enableTLS=true&tlsServerName=tls-name.vmess.example#vmess-tls-server-name",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         231,
			URI:        "vmess://00000000-0000-0000-0000-000000000231@vmess-tls-host.example:443?encryption=auto&tlsHost=alias-host.vmess.example&tlsAllowInsecure=1&tlsDisableSni=1#vmess-tls-host",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
	})

	serverNameOutbound := findOutbound(config.Outbounds, "up_210")
	if serverNameOutbound == nil {
		t.Fatalf("expected vmess outbound up_210, got %+v", config.Outbounds)
	}
	serverNameTLS, ok := serverNameOutbound["tls"].(map[string]any)
	if !ok || serverNameTLS["enabled"] != true || serverNameTLS["server_name"] != "alias.vmess.example" || serverNameTLS["disable_sni"] != true {
		t.Fatalf("unexpected vmess serverName TLS aliases: %+v", serverNameOutbound["tls"])
	}

	tlsServerNameOutbound := findOutbound(config.Outbounds, "up_211")
	if tlsServerNameOutbound == nil {
		t.Fatalf("expected vmess outbound up_211, got %+v", config.Outbounds)
	}
	tlsServerNameTLS, ok := tlsServerNameOutbound["tls"].(map[string]any)
	if !ok || tlsServerNameTLS["enabled"] != true || tlsServerNameTLS["server_name"] != "tls-name.vmess.example" {
		t.Fatalf("unexpected vmess tlsServerName aliases: %+v", tlsServerNameOutbound["tls"])
	}

	tlsHostOutbound := findOutbound(config.Outbounds, "up_231")
	if tlsHostOutbound == nil {
		t.Fatalf("expected vmess outbound up_231, got %+v", config.Outbounds)
	}
	tlsHostTLS, ok := tlsHostOutbound["tls"].(map[string]any)
	if !ok || tlsHostTLS["enabled"] != true || tlsHostTLS["server_name"] != "alias-host.vmess.example" || tlsHostTLS["insecure"] != true || tlsHostTLS["disable_sni"] != true {
		t.Fatalf("unexpected vmess tlsHost aliases: %+v", tlsHostOutbound["tls"])
	}
}

func TestBuildConfigSupportsTUICQueryCredentialAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         89,
			URI:        "tuic://query-tuic.example:443?id=00000000-0000-0000-0000-000000000089&pass=tuic-query-placeholder&congestionControl=bbr&udpOverStream=1&zeroRttHandshake=1&heartbeatInterval=10s&serverName=query-tuic.example&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#tuic-query",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         189,
			URI:        "tuic://relay-tuic.example:443?id=00000000-0000-0000-0000-000000000189&token=tuic-relay-placeholder&udpRelayMode=native&sni=relay-tuic.example#tuic-query-relay",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         225,
			URI:        "tuic://tuic-tls-alias.example:443?id=00000000-0000-0000-0000-000000000225&password=tuic-tls-placeholder&tlsHost=alias.tuic.example&tlsSkipVerify=1&tlsDisableSni=1#tuic-tls-alias",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_89")
	if outbound == nil {
		t.Fatalf("expected tuic outbound up_89, got %+v", config.Outbounds)
	}
	if outbound["type"] != "tuic" || outbound["server"] != "query-tuic.example" || outbound["server_port"] != 443 {
		t.Fatalf("unexpected tuic query credential server fields: %+v", outbound)
	}
	if outbound["uuid"] != "00000000-0000-0000-0000-000000000089" || outbound["password"] != "tuic-query-placeholder" {
		t.Fatalf("unexpected tuic query credential auth fields: %+v", outbound)
	}
	if outbound["congestion_control"] != "bbr" || outbound["udp_over_stream"] != true {
		t.Fatalf("unexpected tuic query credential relay fields: %+v", outbound)
	}
	if outbound["zero_rtt_handshake"] != true || outbound["heartbeat"] != "10s" {
		t.Fatalf("unexpected tuic query credential timing fields: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "query-tuic.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected tuic query credential tls: %+v", outbound["tls"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected tuic query credential utls: %+v", tls["utls"])
	}

	relayOutbound := findOutbound(config.Outbounds, "up_189")
	if relayOutbound == nil {
		t.Fatalf("expected tuic relay outbound up_189, got %+v", config.Outbounds)
	}
	if relayOutbound["uuid"] != "00000000-0000-0000-0000-000000000189" || relayOutbound["password"] != "tuic-relay-placeholder" || relayOutbound["udp_relay_mode"] != "native" {
		t.Fatalf("unexpected tuic query credential relay alias fields: %+v", relayOutbound)
	}

	tlsAliasOutbound := findOutbound(config.Outbounds, "up_225")
	if tlsAliasOutbound == nil {
		t.Fatalf("expected tuic TLS alias outbound up_225, got %+v", config.Outbounds)
	}
	tlsAlias, ok := tlsAliasOutbound["tls"].(map[string]any)
	if !ok || tlsAlias["enabled"] != true || tlsAlias["server_name"] != "alias.tuic.example" || tlsAlias["insecure"] != true || tlsAlias["disable_sni"] != true {
		t.Fatalf("unexpected tuic TLS alias config: %+v", tlsAliasOutbound["tls"])
	}
}

func TestBuildConfigSupportsJuicityOutbound(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         190,
			URI:        "juicity://00000000-0000-0000-0000-000000000190:juicity-placeholder@juicity.example:443?congestionControl=bbr&serverName=juicity.example&allowInsecure=1&disableSNI=1&alpn=h3&clientFingerprint=chrome#juicity",
			Protocol:   "juicity",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         226,
			URI:        "juicity://juicity-tls-alias.example:443?uuid=00000000-0000-0000-0000-000000000226&password=juicity-tls-placeholder&tlsHost=alias.juicity.example&tlsAllowInsecure=1&tlsDisableSni=1#juicity-tls-alias",
			Protocol:   "juicity",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_190")
	if outbound == nil {
		t.Fatalf("expected juicity outbound up_190, got %+v", config.Outbounds)
	}
	if outbound["type"] != "juicity" || outbound["server"] != "juicity.example" || outbound["server_port"] != 443 {
		t.Fatalf("unexpected juicity server fields: %+v", outbound)
	}
	if outbound["uuid"] != "00000000-0000-0000-0000-000000000190" || outbound["password"] != "juicity-placeholder" || outbound["congestion_control"] != "bbr" {
		t.Fatalf("unexpected juicity auth fields: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "juicity.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected juicity tls: %+v", outbound["tls"])
	}
	alpn, ok := tls["alpn"].([]string)
	if !ok || len(alpn) != 1 || alpn[0] != "h3" {
		t.Fatalf("unexpected juicity alpn: %+v", tls["alpn"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected juicity utls: %+v", tls["utls"])
	}

	tlsAliasOutbound := findOutbound(config.Outbounds, "up_226")
	if tlsAliasOutbound == nil {
		t.Fatalf("expected juicity TLS alias outbound up_226, got %+v", config.Outbounds)
	}
	tlsAlias, ok := tlsAliasOutbound["tls"].(map[string]any)
	if !ok || tlsAlias["enabled"] != true || tlsAlias["server_name"] != "alias.juicity.example" || tlsAlias["insecure"] != true || tlsAlias["disable_sni"] != true {
		t.Fatalf("unexpected juicity TLS alias config: %+v", tlsAliasOutbound["tls"])
	}
}

func TestBuildConfigSupportsAnyTLSShadowTLSQueryCredentialAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         90,
			URI:        "anytls://query-anytls.example:443?pass=anytls-query-placeholder&serverName=query-anytls.example&idleSessionCheckInterval=20s&idleSessionTimeout=30s&minIdleSession=2&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#anytls-query",
			Protocol:   "anytls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         91,
			URI:        "shadowtls://query-shadowtls.example:443?version=3&token=shadowtls-query-placeholder&serverName=query-shadowtls.example&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#shadowtls-query",
			Protocol:   "shadowtls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         227,
			URI:        "anytls://alias-anytls.example:443?pwd=anytls-alias-placeholder&tlsHost=alias.anytls.example&tlsAllowInsecure=1&tlsDisableSni=1#anytls-tls-alias",
			Protocol:   "anytls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         228,
			URI:        "shadowtls://alias-shadowtls.example:443?version=3&secret=shadowtls-alias-placeholder&tlsServerName=alias.shadowtls.example&tlsSkipVerify=1&tlsDisableSNI=1#shadowtls-tls-alias",
			Protocol:   "shadowtls",
			ServerPort: 443,
			Status:     "active",
		},
	})

	anytlsOutbound := findOutbound(config.Outbounds, "up_90")
	if anytlsOutbound == nil {
		t.Fatalf("expected anytls outbound up_90, got %+v", config.Outbounds)
	}
	if anytlsOutbound["type"] != "anytls" || anytlsOutbound["server"] != "query-anytls.example" || anytlsOutbound["password"] != "anytls-query-placeholder" || anytlsOutbound["idle_session_check_interval"] != "20s" || anytlsOutbound["idle_session_timeout"] != "30s" || anytlsOutbound["min_idle_session"] != 2 {
		t.Fatalf("unexpected anytls query credential fields: %+v", anytlsOutbound)
	}
	anytlsTLS, ok := anytlsOutbound["tls"].(map[string]any)
	if !ok || anytlsTLS["enabled"] != true || anytlsTLS["server_name"] != "query-anytls.example" || anytlsTLS["insecure"] != true || anytlsTLS["disable_sni"] != true {
		t.Fatalf("unexpected anytls query credential tls: %+v", anytlsOutbound["tls"])
	}
	anytlsUTLS, ok := anytlsTLS["utls"].(map[string]any)
	if !ok || anytlsUTLS["enabled"] != true || anytlsUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected anytls query credential utls: %+v", anytlsTLS["utls"])
	}

	shadowtlsOutbound := findOutbound(config.Outbounds, "up_91")
	if shadowtlsOutbound == nil {
		t.Fatalf("expected shadowtls outbound up_91, got %+v", config.Outbounds)
	}
	if shadowtlsOutbound["type"] != "shadowtls" || shadowtlsOutbound["server"] != "query-shadowtls.example" || shadowtlsOutbound["version"] != 3 || shadowtlsOutbound["password"] != "shadowtls-query-placeholder" {
		t.Fatalf("unexpected shadowtls query credential fields: %+v", shadowtlsOutbound)
	}
	shadowtlsTLS, ok := shadowtlsOutbound["tls"].(map[string]any)
	if !ok || shadowtlsTLS["enabled"] != true || shadowtlsTLS["server_name"] != "query-shadowtls.example" || shadowtlsTLS["insecure"] != true || shadowtlsTLS["disable_sni"] != true {
		t.Fatalf("unexpected shadowtls query credential tls: %+v", shadowtlsOutbound["tls"])
	}
	shadowtlsUTLS, ok := shadowtlsTLS["utls"].(map[string]any)
	if !ok || shadowtlsUTLS["enabled"] != true || shadowtlsUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected shadowtls query credential utls: %+v", shadowtlsTLS["utls"])
	}

	anytlsAliasOutbound := findOutbound(config.Outbounds, "up_227")
	if anytlsAliasOutbound == nil {
		t.Fatalf("expected anytls TLS alias outbound up_227, got %+v", config.Outbounds)
	}
	anytlsAliasTLS, ok := anytlsAliasOutbound["tls"].(map[string]any)
	if !ok || anytlsAliasTLS["enabled"] != true || anytlsAliasTLS["server_name"] != "alias.anytls.example" || anytlsAliasTLS["insecure"] != true || anytlsAliasTLS["disable_sni"] != true {
		t.Fatalf("unexpected anytls TLS alias config: %+v", anytlsAliasOutbound["tls"])
	}

	shadowtlsAliasOutbound := findOutbound(config.Outbounds, "up_228")
	if shadowtlsAliasOutbound == nil {
		t.Fatalf("expected shadowtls TLS alias outbound up_228, got %+v", config.Outbounds)
	}
	shadowtlsAliasTLS, ok := shadowtlsAliasOutbound["tls"].(map[string]any)
	if !ok || shadowtlsAliasTLS["enabled"] != true || shadowtlsAliasTLS["server_name"] != "alias.shadowtls.example" || shadowtlsAliasTLS["insecure"] != true || shadowtlsAliasTLS["disable_sni"] != true {
		t.Fatalf("unexpected shadowtls TLS alias config: %+v", shadowtlsAliasOutbound["tls"])
	}
}

func TestBuildConfigSupportsHysteria2QueryPassword(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         80,
			URI:        "hy2://query-password.hy2.example:443?password=hy2-query-placeholder&obfs=salamander&obfsPassword=obfs-placeholder&upMbps=35&downMbps=95&serverName=hy2-query.example&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#hy2-query",
			Protocol:   "hy2",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_80")
	if outbound == nil {
		t.Fatalf("expected hysteria2 outbound up_80, got %+v", config.Outbounds)
	}
	if outbound["type"] != "hysteria2" || outbound["server"] != "query-password.hy2.example" || outbound["server_port"] != 443 || outbound["password"] != "hy2-query-placeholder" {
		t.Fatalf("unexpected hysteria2 query password fields: %+v", outbound)
	}
	obfs, ok := outbound["obfs"].(map[string]any)
	if !ok || obfs["type"] != "salamander" || obfs["password"] != "obfs-placeholder" {
		t.Fatalf("unexpected hysteria2 query password obfs: %+v", outbound["obfs"])
	}
	if outbound["up_mbps"] != 35 || outbound["down_mbps"] != 95 {
		t.Fatalf("unexpected hysteria2 query bandwidth aliases: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "hy2-query.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected hysteria2 query password tls: %+v", outbound["tls"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected hysteria2 query password utls: %+v", tls["utls"])
	}
}

func TestBuildConfigSupportsHysteriaTokenAuth(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         81,
			URI:        "hysteria://token-auth.hysteria.example:443?token=hysteria-token-placeholder&upMbps=30&downMbps=90&obfs=obfs-placeholder&recvWindowConn=1048576&recvWindow=2097152&disableMTUDiscovery=1&serverName=hysteria-token.example&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#hysteria-token",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_81")
	if outbound == nil {
		t.Fatalf("expected hysteria outbound up_81, got %+v", config.Outbounds)
	}
	if outbound["type"] != "hysteria" || outbound["server"] != "token-auth.hysteria.example" || outbound["server_port"] != 443 || outbound["auth_str"] != "hysteria-token-placeholder" {
		t.Fatalf("unexpected hysteria token auth fields: %+v", outbound)
	}
	if outbound["up_mbps"] != 30 || outbound["down_mbps"] != 90 || outbound["obfs"] != "obfs-placeholder" {
		t.Fatalf("unexpected hysteria token auth options: %+v", outbound)
	}
	if outbound["recv_window_conn"] != 1048576 || outbound["recv_window"] != 2097152 || outbound["disable_mtu_discovery"] != true {
		t.Fatalf("unexpected hysteria token auth window aliases: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "hysteria-token.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected hysteria token auth tls: %+v", outbound["tls"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected hysteria token auth utls: %+v", tls["utls"])
	}
}

func TestBuildConfigSupportsHysteriaAuthAliasQueries(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         195,
			URI:        "hysteria://auth-base64.hysteria.example:443?authBase64=base64-auth-placeholder&serverName=auth-base64.hysteria.example#hysteria-auth-base64",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         196,
			URI:        "hysteria://auth-str.hysteria.example:443?authStr=hysteria-authstr-placeholder&upMbps=12&downMbps=34#hysteria-auth-str",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         197,
			URI:        "hysteria2://hy2-authstr.example:443?authStr=hy2-authstr-placeholder#hy2-auth-str",
			Protocol:   "hysteria2",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         198,
			URI:        "hy2://hy2-pass.example:443?pass=hy2-pass-placeholder#hy2-pass",
			Protocol:   "hy2",
			ServerPort: 443,
			Status:     "active",
		},
	})

	authBase64Outbound := findOutbound(config.Outbounds, "up_195")
	if authBase64Outbound == nil {
		t.Fatalf("expected hysteria authBase64 outbound up_195, got %+v", config.Outbounds)
	}
	if authBase64Outbound["type"] != "hysteria" || authBase64Outbound["auth"] != "base64-auth-placeholder" {
		t.Fatalf("unexpected hysteria authBase64 fields: %+v", authBase64Outbound)
	}
	authBase64TLS, ok := authBase64Outbound["tls"].(map[string]any)
	if !ok || authBase64TLS["server_name"] != "auth-base64.hysteria.example" {
		t.Fatalf("unexpected hysteria authBase64 tls: %+v", authBase64Outbound["tls"])
	}

	authStrOutbound := findOutbound(config.Outbounds, "up_196")
	if authStrOutbound == nil {
		t.Fatalf("expected hysteria authStr outbound up_196, got %+v", config.Outbounds)
	}
	if authStrOutbound["type"] != "hysteria" || authStrOutbound["auth_str"] != "hysteria-authstr-placeholder" || authStrOutbound["up_mbps"] != 12 || authStrOutbound["down_mbps"] != 34 {
		t.Fatalf("unexpected hysteria authStr fields: %+v", authStrOutbound)
	}

	hy2AuthStrOutbound := findOutbound(config.Outbounds, "up_197")
	if hy2AuthStrOutbound == nil {
		t.Fatalf("expected hysteria2 authStr outbound up_197, got %+v", config.Outbounds)
	}
	if hy2AuthStrOutbound["type"] != "hysteria2" || hy2AuthStrOutbound["password"] != "hy2-authstr-placeholder" {
		t.Fatalf("unexpected hysteria2 authStr fields: %+v", hy2AuthStrOutbound)
	}

	hy2PassOutbound := findOutbound(config.Outbounds, "up_198")
	if hy2PassOutbound == nil {
		t.Fatalf("expected hysteria2 pass outbound up_198, got %+v", config.Outbounds)
	}
	if hy2PassOutbound["type"] != "hysteria2" || hy2PassOutbound["password"] != "hy2-pass-placeholder" {
		t.Fatalf("unexpected hysteria2 pass fields: %+v", hy2PassOutbound)
	}
}

func TestBuildConfigSupportsHysteriaTLSAliasQueries(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         223,
			URI:        "hy2://hy2-tls-alias.example:443?password=hy2-tls-placeholder&tlsServerName=alias.hy2.example&tlsSkipVerify=1&tlsDisableSni=1#hy2-tls-alias",
			Protocol:   "hy2",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         224,
			URI:        "hysteria://hysteria-tls-alias.example:443?token=hysteria-tls-placeholder&tlsHost=alias.hysteria.example&tlsAllowInsecure=1&tlsDisableSNI=1#hysteria-tls-alias",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
	})

	hy2Outbound := findOutbound(config.Outbounds, "up_223")
	if hy2Outbound == nil {
		t.Fatalf("expected hysteria2 outbound up_223, got %+v", config.Outbounds)
	}
	hy2TLS, ok := hy2Outbound["tls"].(map[string]any)
	if !ok || hy2TLS["enabled"] != true || hy2TLS["server_name"] != "alias.hy2.example" || hy2TLS["insecure"] != true || hy2TLS["disable_sni"] != true {
		t.Fatalf("unexpected hysteria2 TLS alias config: %+v", hy2Outbound["tls"])
	}

	hysteriaOutbound := findOutbound(config.Outbounds, "up_224")
	if hysteriaOutbound == nil {
		t.Fatalf("expected hysteria outbound up_224, got %+v", config.Outbounds)
	}
	hysteriaTLS, ok := hysteriaOutbound["tls"].(map[string]any)
	if !ok || hysteriaTLS["enabled"] != true || hysteriaTLS["server_name"] != "alias.hysteria.example" || hysteriaTLS["insecure"] != true || hysteriaTLS["disable_sni"] != true {
		t.Fatalf("unexpected hysteria TLS alias config: %+v", hysteriaOutbound["tls"])
	}
}

func TestBuildConfigSupportsVLESSTrojanQueryCredentials(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         82,
			URI:        "vless://query-vless.example:443?uuid=00000000-0000-0000-0000-000000000094&security=reality&serverName=query-vless.example&publicKey=vless-public-placeholder&shortId=0123abcd&clientFingerprint=chrome&allowInsecure=1&disableSNI=1&type=grpc&serviceName=query-vless&grpcIdleTimeout=20s&grpcPingTimeout=10s&permitWithoutStream=1&multiMode=1#vless-query",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         83,
			URI:        "trojan://query-trojan.example:443?password=trojan-query-placeholder&security=reality&serverName=query-trojan.example&publicKey=trojan-public-placeholder&shortId=4567abcd&clientFingerprint=firefox&skip_cert_verify=1&disableSNI=1&type=ws&wsPath=%2Ftrojan&wsHost=ws.query-trojan.example&maxEarlyData=2048&earlyDataHeaderName=Sec-WebSocket-Protocol#trojan-query",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
	})

	vlessOutbound := findOutbound(config.Outbounds, "up_82")
	if vlessOutbound == nil {
		t.Fatalf("expected vless outbound up_82, got %+v", config.Outbounds)
	}
	if vlessOutbound["type"] != "vless" || vlessOutbound["server"] != "query-vless.example" || vlessOutbound["server_port"] != 443 {
		t.Fatalf("unexpected vless query credential server fields: %+v", vlessOutbound)
	}
	if vlessOutbound["uuid"] != "00000000-0000-0000-0000-000000000094" {
		t.Fatalf("unexpected vless query credential uuid: %+v", vlessOutbound)
	}
	vlessTLS, ok := vlessOutbound["tls"].(map[string]any)
	if !ok || vlessTLS["enabled"] != true || vlessTLS["server_name"] != "query-vless.example" || vlessTLS["insecure"] != true || vlessTLS["disable_sni"] != true {
		t.Fatalf("unexpected vless query credential tls: %+v", vlessOutbound["tls"])
	}
	vlessReality, ok := vlessTLS["reality"].(map[string]any)
	if !ok || vlessReality["enabled"] != true || vlessReality["public_key"] != "vless-public-placeholder" || vlessReality["short_id"] != "0123abcd" {
		t.Fatalf("unexpected vless query credential reality: %+v", vlessTLS["reality"])
	}
	vlessUTLS, ok := vlessTLS["utls"].(map[string]any)
	if !ok || vlessUTLS["enabled"] != true || vlessUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected vless query credential utls: %+v", vlessTLS["utls"])
	}
	vlessTransport, ok := vlessOutbound["transport"].(map[string]any)
	if !ok || vlessTransport["type"] != "grpc" || vlessTransport["service_name"] != "query-vless" || vlessTransport["idle_timeout"] != "20s" || vlessTransport["ping_timeout"] != "10s" || vlessTransport["permit_without_stream"] != true || vlessTransport["multi_mode"] != true {
		t.Fatalf("unexpected vless query credential transport: %+v", vlessOutbound["transport"])
	}

	trojanOutbound := findOutbound(config.Outbounds, "up_83")
	if trojanOutbound == nil {
		t.Fatalf("expected trojan outbound up_83, got %+v", config.Outbounds)
	}
	if trojanOutbound["type"] != "trojan" || trojanOutbound["server"] != "query-trojan.example" || trojanOutbound["server_port"] != 443 {
		t.Fatalf("unexpected trojan query credential server fields: %+v", trojanOutbound)
	}
	if trojanOutbound["password"] != "trojan-query-placeholder" {
		t.Fatalf("unexpected trojan query credential password: %+v", trojanOutbound)
	}
	trojanTLS, ok := trojanOutbound["tls"].(map[string]any)
	if !ok || trojanTLS["enabled"] != true || trojanTLS["server_name"] != "query-trojan.example" || trojanTLS["insecure"] != true || trojanTLS["disable_sni"] != true {
		t.Fatalf("unexpected trojan query credential tls: %+v", trojanOutbound["tls"])
	}
	trojanReality, ok := trojanTLS["reality"].(map[string]any)
	if !ok || trojanReality["enabled"] != true || trojanReality["public_key"] != "trojan-public-placeholder" || trojanReality["short_id"] != "4567abcd" {
		t.Fatalf("unexpected trojan query credential reality: %+v", trojanTLS["reality"])
	}
	trojanUTLS, ok := trojanTLS["utls"].(map[string]any)
	if !ok || trojanUTLS["enabled"] != true || trojanUTLS["fingerprint"] != "firefox" {
		t.Fatalf("unexpected trojan query credential utls: %+v", trojanTLS["utls"])
	}
	trojanTransport, ok := trojanOutbound["transport"].(map[string]any)
	if !ok || trojanTransport["type"] != "ws" || trojanTransport["path"] != "/trojan" || trojanTransport["max_early_data"] != 2048 || trojanTransport["early_data_header_name"] != "Sec-WebSocket-Protocol" {
		t.Fatalf("unexpected trojan query credential transport: %+v", trojanOutbound["transport"])
	}
	headers, ok := trojanTransport["headers"].(map[string]any)
	if !ok || headers["Host"] != "ws.query-trojan.example" {
		t.Fatalf("unexpected trojan query credential headers: %+v", trojanTransport["headers"])
	}
}

func TestBuildConfigSupportsUUIDQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         199,
			URI:        "vless://uuid-alias-vless.example:443?userId=00000000-0000-0000-0000-000000000199&security=tls#vless-uuid-alias",
			Protocol:   "vless",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         200,
			URI:        "vmess://uuid-alias-vmess.example:443?userID=00000000-0000-0000-0000-000000000200&security=tls#vmess-uuid-alias",
			Protocol:   "vmess",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         201,
			URI:        "tuic://uuid-alias-tuic.example:443?userid=00000000-0000-0000-0000-000000000201&token=tuic-uuid-alias-placeholder#tuic-uuid-alias",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         202,
			URI:        "juicity://uuid-alias-juicity.example:443?userId=00000000-0000-0000-0000-000000000202&token=juicity-uuid-alias-placeholder#juicity-uuid-alias",
			Protocol:   "juicity",
			ServerPort: 443,
			Status:     "active",
		},
	})

	vlessOutbound := findOutbound(config.Outbounds, "up_199")
	if vlessOutbound == nil {
		t.Fatalf("expected vless outbound up_199, got %+v", config.Outbounds)
	}
	if vlessOutbound["uuid"] != "00000000-0000-0000-0000-000000000199" {
		t.Fatalf("unexpected vless UUID alias fields: %+v", vlessOutbound)
	}

	vmessOutbound := findOutbound(config.Outbounds, "up_200")
	if vmessOutbound == nil {
		t.Fatalf("expected vmess outbound up_200, got %+v", config.Outbounds)
	}
	if vmessOutbound["uuid"] != "00000000-0000-0000-0000-000000000200" {
		t.Fatalf("unexpected vmess UUID alias fields: %+v", vmessOutbound)
	}

	tuicOutbound := findOutbound(config.Outbounds, "up_201")
	if tuicOutbound == nil {
		t.Fatalf("expected tuic outbound up_201, got %+v", config.Outbounds)
	}
	if tuicOutbound["uuid"] != "00000000-0000-0000-0000-000000000201" || tuicOutbound["password"] != "tuic-uuid-alias-placeholder" {
		t.Fatalf("unexpected tuic UUID alias fields: %+v", tuicOutbound)
	}

	juicityOutbound := findOutbound(config.Outbounds, "up_202")
	if juicityOutbound == nil {
		t.Fatalf("expected juicity outbound up_202, got %+v", config.Outbounds)
	}
	if juicityOutbound["uuid"] != "00000000-0000-0000-0000-000000000202" || juicityOutbound["password"] != "juicity-uuid-alias-placeholder" {
		t.Fatalf("unexpected juicity UUID alias fields: %+v", juicityOutbound)
	}
}

func TestBuildConfigSupportsSecretCredentialQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         203,
			URI:        "trojan://secret-trojan.example:443?credential=trojan-secret-placeholder&security=tls#trojan-secret",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         204,
			URI:        "tuic://secret-tuic.example:443?userID=00000000-0000-0000-0000-000000000204&secret=tuic-secret-placeholder#tuic-secret",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         205,
			URI:        "juicity://secret-juicity.example:443?id=00000000-0000-0000-0000-000000000205&accountPassword=juicity-account-placeholder#juicity-secret",
			Protocol:   "juicity",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         206,
			URI:        "anytls://secret-anytls.example:443?pwd=anytls-pwd-placeholder#anytls-secret",
			Protocol:   "anytls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         207,
			URI:        "shadowtls://secret-shadowtls.example:443?version=3&credentials=shadowtls-credentials-placeholder#shadowtls-secret",
			Protocol:   "shadowtls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         208,
			URI:        "hysteria2://secret-hy2.example:443?account-password=hy2-account-placeholder#hy2-secret",
			Protocol:   "hysteria2",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         209,
			URI:        "hysteria://secret-hysteria.example:443?credential=hysteria-credential-placeholder#hysteria-secret",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
	})

	trojanOutbound := findOutbound(config.Outbounds, "up_203")
	if trojanOutbound == nil || trojanOutbound["password"] != "trojan-secret-placeholder" {
		t.Fatalf("unexpected trojan secret alias fields: %+v", trojanOutbound)
	}

	tuicOutbound := findOutbound(config.Outbounds, "up_204")
	if tuicOutbound == nil || tuicOutbound["uuid"] != "00000000-0000-0000-0000-000000000204" || tuicOutbound["password"] != "tuic-secret-placeholder" {
		t.Fatalf("unexpected tuic secret alias fields: %+v", tuicOutbound)
	}

	juicityOutbound := findOutbound(config.Outbounds, "up_205")
	if juicityOutbound == nil || juicityOutbound["uuid"] != "00000000-0000-0000-0000-000000000205" || juicityOutbound["password"] != "juicity-account-placeholder" {
		t.Fatalf("unexpected juicity secret alias fields: %+v", juicityOutbound)
	}

	anytlsOutbound := findOutbound(config.Outbounds, "up_206")
	if anytlsOutbound == nil || anytlsOutbound["password"] != "anytls-pwd-placeholder" {
		t.Fatalf("unexpected anytls secret alias fields: %+v", anytlsOutbound)
	}

	shadowtlsOutbound := findOutbound(config.Outbounds, "up_207")
	if shadowtlsOutbound == nil || shadowtlsOutbound["password"] != "shadowtls-credentials-placeholder" {
		t.Fatalf("unexpected shadowtls secret alias fields: %+v", shadowtlsOutbound)
	}

	hy2Outbound := findOutbound(config.Outbounds, "up_208")
	if hy2Outbound == nil || hy2Outbound["password"] != "hy2-account-placeholder" {
		t.Fatalf("unexpected hysteria2 secret alias fields: %+v", hy2Outbound)
	}

	hysteriaOutbound := findOutbound(config.Outbounds, "up_209")
	if hysteriaOutbound == nil || hysteriaOutbound["auth_str"] != "hysteria-credential-placeholder" {
		t.Fatalf("unexpected hysteria secret alias fields: %+v", hysteriaOutbound)
	}
}

func TestBuildConfigSupportsShadowsocksQueryCredentials(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         87,
			URI:        "ss://query-ss.example:8388?method=aes-128-gcm&password=ss-query-placeholder&plugin=v2ray-plugin&plugin_opts=mode%3Dwebsocket%3Bhost%3Dss.query.example&network=tcp#ss-query",
			Protocol:   "ss",
			ServerPort: 8388,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_87")
	if outbound == nil {
		t.Fatalf("expected shadowsocks outbound up_87, got %+v", config.Outbounds)
	}
	if outbound["type"] != "shadowsocks" || outbound["server"] != "query-ss.example" || outbound["server_port"] != 8388 {
		t.Fatalf("unexpected shadowsocks query credential server fields: %+v", outbound)
	}
	if outbound["method"] != "aes-128-gcm" || outbound["password"] != "ss-query-placeholder" {
		t.Fatalf("unexpected shadowsocks query credential auth fields: %+v", outbound)
	}
	if outbound["plugin"] != "v2ray-plugin" || outbound["plugin_opts"] != "mode=websocket;host=ss.query.example" || outbound["network"] != "tcp" {
		t.Fatalf("unexpected shadowsocks query credential plugin fields: %+v", outbound)
	}
}

func TestBuildConfigSupportsShadowsocksCredentialQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         99,
			URI:        "ss://ss-alias.example:8388?encryptMethod=chacha20-ietf-poly1305&accountPassword=ss-alias-placeholder&plugin=v2ray-plugin&pluginOptions=mode%3Dwebsocket%3Bhost%3Dss.alias.example&network=udp#ss-alias",
			Protocol:   "ss",
			ServerPort: 8388,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_99")
	if outbound == nil {
		t.Fatalf("expected shadowsocks outbound up_99, got %+v", config.Outbounds)
	}
	if outbound["method"] != "chacha20-ietf-poly1305" || outbound["password"] != "ss-alias-placeholder" {
		t.Fatalf("unexpected shadowsocks alias credential auth fields: %+v", outbound)
	}
	if outbound["plugin"] != "v2ray-plugin" || outbound["plugin_opts"] != "mode=websocket;host=ss.alias.example" || outbound["network"] != "udp" {
		t.Fatalf("unexpected shadowsocks alias plugin fields: %+v", outbound)
	}
}

func TestBuildConfigSupportsNaiveQueryCredentials(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         84,
			URI:        "naive+quic://query-naive.example:443?user=qa-user&pass=naive-query-placeholder&serverName=query-naive.example&quic=1&insecureConcurrency=2&udpOverTcp=1&quicCongestionControl=bbr&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#naive-query",
			Protocol:   "naive+quic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         229,
			URI:        "naive://alias-naive.example:443?user=qa-user&pwd=naive-alias-placeholder&tlsHost=alias.naive.example&tlsAllowInsecure=1&tlsDisableSni=1#naive-tls-alias",
			Protocol:   "naive",
			ServerPort: 443,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_84")
	if outbound == nil {
		t.Fatalf("expected naive outbound up_84, got %+v", config.Outbounds)
	}
	if outbound["type"] != "naive" || outbound["server"] != "query-naive.example" || outbound["server_port"] != 443 {
		t.Fatalf("unexpected naive query credential server fields: %+v", outbound)
	}
	if outbound["username"] != "qa-user" || outbound["password"] != "naive-query-placeholder" || outbound["quic"] != true {
		t.Fatalf("unexpected naive query credential auth fields: %+v", outbound)
	}
	if outbound["insecure_concurrency"] != 2 || outbound["udp_over_tcp"] != true || outbound["quic_congestion_control"] != "bbr" {
		t.Fatalf("unexpected naive query credential transport fields: %+v", outbound)
	}
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "query-naive.example" || tls["insecure"] != true || tls["disable_sni"] != true {
		t.Fatalf("unexpected naive query credential tls: %+v", outbound["tls"])
	}
	utls, ok := tls["utls"].(map[string]any)
	if !ok || utls["enabled"] != true || utls["fingerprint"] != "chrome" {
		t.Fatalf("unexpected naive query credential utls: %+v", tls["utls"])
	}

	aliasOutbound := findOutbound(config.Outbounds, "up_229")
	if aliasOutbound == nil {
		t.Fatalf("expected naive TLS alias outbound up_229, got %+v", config.Outbounds)
	}
	aliasTLS, ok := aliasOutbound["tls"].(map[string]any)
	if !ok || aliasTLS["enabled"] != true || aliasTLS["server_name"] != "alias.naive.example" || aliasTLS["insecure"] != true || aliasTLS["disable_sni"] != true {
		t.Fatalf("unexpected naive TLS alias config: %+v", aliasOutbound["tls"])
	}
}

func TestBuildConfigSupportsHTTPAndSOCKSQueryCredentials(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         85,
			URI:        "https://http-query.example:8443/connect?username=qa-user&password=http-query-placeholder&serverName=http-query.example&allowInsecure=1&disableSNI=1&clientFingerprint=chrome#http-query",
			Protocol:   "https",
			ServerPort: 8443,
			Status:     "active",
		},
		{
			ID:         86,
			URI:        "socks5://socks-query.example:1080?user=qa-user&pass=socks-query-placeholder&udpEnabled=1&udpOverTcp=1#socks-query",
			Protocol:   "socks5",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         87,
			URI:        "socks5h://socks5h-query.example:1080?username=qa-user&password=socks5h-query-placeholder#socks5h-query",
			Protocol:   "socks5h",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         93,
			URI:        "http+tls://http-plus-tls.example:443/connect?username=qa-user&password=http-plus-tls-placeholder&serverName=http-plus-tls.example#http-plus-tls",
			Protocol:   "http+tls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:       94,
			URI:      "http-tls://http-dash-tls.example/connect?user=qa-user&pass=http-dash-tls-placeholder#http-dash-tls",
			Protocol: "http-tls",
			Status:   "active",
		},
		{
			ID:         230,
			URI:        "http://http-alias.example:8080/connect?username=qa-user&password=http-alias-placeholder&tlsServerName=alias.http.example&tlsSkipVerify=1&tlsDisableSNI=1#http-tls-alias",
			Protocol:   "http",
			ServerPort: 8080,
			Status:     "active",
		},
	})

	httpOutbound := findOutbound(config.Outbounds, "up_85")
	if httpOutbound == nil {
		t.Fatalf("expected http outbound up_85, got %+v", config.Outbounds)
	}
	if httpOutbound["type"] != "http" || httpOutbound["server"] != "http-query.example" || httpOutbound["server_port"] != 8443 {
		t.Fatalf("unexpected http query credential server fields: %+v", httpOutbound)
	}
	if httpOutbound["username"] != "qa-user" || httpOutbound["password"] != "http-query-placeholder" || httpOutbound["path"] != "/connect" {
		t.Fatalf("unexpected http query credential auth fields: %+v", httpOutbound)
	}
	httpTLS, ok := httpOutbound["tls"].(map[string]any)
	if !ok || httpTLS["enabled"] != true || httpTLS["server_name"] != "http-query.example" || httpTLS["insecure"] != true || httpTLS["disable_sni"] != true {
		t.Fatalf("unexpected http query credential tls: %+v", httpOutbound["tls"])
	}
	httpUTLS, ok := httpTLS["utls"].(map[string]any)
	if !ok || httpUTLS["enabled"] != true || httpUTLS["fingerprint"] != "chrome" {
		t.Fatalf("unexpected http query credential utls: %+v", httpTLS["utls"])
	}

	socksOutbound := findOutbound(config.Outbounds, "up_86")
	if socksOutbound == nil {
		t.Fatalf("expected socks outbound up_86, got %+v", config.Outbounds)
	}
	if socksOutbound["type"] != "socks" || socksOutbound["server"] != "socks-query.example" || socksOutbound["server_port"] != 1080 || socksOutbound["version"] != "5" {
		t.Fatalf("unexpected socks query credential server fields: %+v", socksOutbound)
	}
	if socksOutbound["username"] != "qa-user" || socksOutbound["password"] != "socks-query-placeholder" || socksOutbound["network"] != "udp" || socksOutbound["udp_over_tcp"] != true {
		t.Fatalf("unexpected socks query credential auth fields: %+v", socksOutbound)
	}

	socks5hOutbound := findOutbound(config.Outbounds, "up_87")
	if socks5hOutbound == nil {
		t.Fatalf("expected socks5h outbound up_87, got %+v", config.Outbounds)
	}
	if socks5hOutbound["type"] != "socks" || socks5hOutbound["server"] != "socks5h-query.example" || socks5hOutbound["server_port"] != 1080 || socks5hOutbound["version"] != "5" {
		t.Fatalf("unexpected socks5h query credential server fields: %+v", socks5hOutbound)
	}
	if socks5hOutbound["username"] != "qa-user" || socks5hOutbound["password"] != "socks5h-query-placeholder" {
		t.Fatalf("unexpected socks5h query credential auth fields: %+v", socks5hOutbound)
	}

	httpPlusTLSOutbound := findOutbound(config.Outbounds, "up_93")
	if httpPlusTLSOutbound == nil {
		t.Fatalf("expected http+tls outbound up_93, got %+v", config.Outbounds)
	}
	if httpPlusTLSOutbound["type"] != "http" || httpPlusTLSOutbound["server"] != "http-plus-tls.example" || httpPlusTLSOutbound["server_port"] != 443 || httpPlusTLSOutbound["path"] != "/connect" {
		t.Fatalf("unexpected http+tls outbound fields: %+v", httpPlusTLSOutbound)
	}
	if httpPlusTLSOutbound["username"] != "qa-user" || httpPlusTLSOutbound["password"] != "http-plus-tls-placeholder" {
		t.Fatalf("unexpected http+tls credential fields: %+v", httpPlusTLSOutbound)
	}
	httpPlusTLSTLS, ok := httpPlusTLSOutbound["tls"].(map[string]any)
	if !ok || httpPlusTLSTLS["enabled"] != true || httpPlusTLSTLS["server_name"] != "http-plus-tls.example" {
		t.Fatalf("unexpected http+tls tls fields: %+v", httpPlusTLSOutbound["tls"])
	}

	httpDashTLSOutbound := findOutbound(config.Outbounds, "up_94")
	if httpDashTLSOutbound == nil {
		t.Fatalf("expected http-tls outbound up_94, got %+v", config.Outbounds)
	}
	if httpDashTLSOutbound["type"] != "http" || httpDashTLSOutbound["server"] != "http-dash-tls.example" || httpDashTLSOutbound["server_port"] != 443 {
		t.Fatalf("unexpected http-tls outbound fields: %+v", httpDashTLSOutbound)
	}
	if httpDashTLSOutbound["username"] != "qa-user" || httpDashTLSOutbound["password"] != "http-dash-tls-placeholder" {
		t.Fatalf("unexpected http-tls credential fields: %+v", httpDashTLSOutbound)
	}
	httpDashTLSTLS, ok := httpDashTLSOutbound["tls"].(map[string]any)
	if !ok || httpDashTLSTLS["enabled"] != true || httpDashTLSTLS["server_name"] != "http-dash-tls.example" {
		t.Fatalf("unexpected http-tls tls fields: %+v", httpDashTLSOutbound["tls"])
	}

	httpAliasOutbound := findOutbound(config.Outbounds, "up_230")
	if httpAliasOutbound == nil {
		t.Fatalf("expected http TLS alias outbound up_230, got %+v", config.Outbounds)
	}
	httpAliasTLS, ok := httpAliasOutbound["tls"].(map[string]any)
	if !ok || httpAliasTLS["enabled"] != true || httpAliasTLS["server_name"] != "alias.http.example" || httpAliasTLS["insecure"] != true || httpAliasTLS["disable_sni"] != true {
		t.Fatalf("unexpected http TLS alias config: %+v", httpAliasOutbound["tls"])
	}
}

func TestBuildConfigSupportsUserPasswordCredentialQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         95,
			URI:        "naive://naive-alias.example:443?accountName=qa-naive-user&pwd=naive-alias-placeholder#naive-alias",
			Protocol:   "naive",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         96,
			URI:        "https://http-alias.example:8443/connect?userName=qa-http-user&accountPassword=http-alias-placeholder#http-alias",
			Protocol:   "https",
			ServerPort: 8443,
			Status:     "active",
		},
		{
			ID:         97,
			URI:        "socks5://socks-alias.example:1080?accountName=qa-socks-user&pwd=socks-alias-placeholder&udp=1#socks-alias",
			Protocol:   "socks5",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         98,
			URI:        "ssh://ssh-alias.example:22?login=qa-ssh-user&accountPassword=ssh-alias-placeholder#ssh-alias",
			Protocol:   "ssh",
			ServerPort: 22,
			Status:     "active",
		},
	})

	naiveOutbound := findOutbound(config.Outbounds, "up_95")
	if naiveOutbound == nil {
		t.Fatalf("expected naive outbound up_95, got %+v", config.Outbounds)
	}
	if naiveOutbound["username"] != "qa-naive-user" || naiveOutbound["password"] != "naive-alias-placeholder" {
		t.Fatalf("unexpected naive alias credential fields: %+v", naiveOutbound)
	}

	httpOutbound := findOutbound(config.Outbounds, "up_96")
	if httpOutbound == nil {
		t.Fatalf("expected http outbound up_96, got %+v", config.Outbounds)
	}
	if httpOutbound["username"] != "qa-http-user" || httpOutbound["password"] != "http-alias-placeholder" || httpOutbound["path"] != "/connect" {
		t.Fatalf("unexpected http alias credential fields: %+v", httpOutbound)
	}

	socksOutbound := findOutbound(config.Outbounds, "up_97")
	if socksOutbound == nil {
		t.Fatalf("expected socks outbound up_97, got %+v", config.Outbounds)
	}
	if socksOutbound["username"] != "qa-socks-user" || socksOutbound["password"] != "socks-alias-placeholder" || socksOutbound["network"] != "udp" {
		t.Fatalf("unexpected socks alias credential fields: %+v", socksOutbound)
	}

	sshOutbound := findOutbound(config.Outbounds, "up_98")
	if sshOutbound == nil {
		t.Fatalf("expected ssh outbound up_98, got %+v", config.Outbounds)
	}
	if sshOutbound["user"] != "qa-ssh-user" || sshOutbound["password"] != "ssh-alias-placeholder" {
		t.Fatalf("unexpected ssh alias credential fields: %+v", sshOutbound)
	}
}

func TestBuildConfigPreservesSOCKSVersionVariants(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         90,
			URI:        "socks4://socks4.example:1080?user=qa-user#socks4",
			Protocol:   "socks4",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         91,
			URI:        "socks4a://socks4a.example:1080?username=qa-user&password=socks4a-placeholder#socks4a",
			Protocol:   "socks4a",
			ServerPort: 1080,
			Status:     "active",
		},
		{
			ID:         92,
			URI:        "socks://socks-version.example:1080?version=4a&user=qa-user&pass=socks-version-placeholder#socks-version",
			Protocol:   "socks",
			ServerPort: 1080,
			Status:     "active",
		},
	})

	socks4 := findOutbound(config.Outbounds, "up_90")
	if socks4 == nil {
		t.Fatalf("expected socks4 outbound up_90, got %+v", config.Outbounds)
	}
	if socks4["type"] != "socks" || socks4["server"] != "socks4.example" || socks4["server_port"] != 1080 || socks4["version"] != "4" {
		t.Fatalf("unexpected socks4 outbound fields: %+v", socks4)
	}
	if socks4["username"] != "qa-user" {
		t.Fatalf("unexpected socks4 credential fields: %+v", socks4)
	}

	socks4a := findOutbound(config.Outbounds, "up_91")
	if socks4a == nil {
		t.Fatalf("expected socks4a outbound up_91, got %+v", config.Outbounds)
	}
	if socks4a["type"] != "socks" || socks4a["server"] != "socks4a.example" || socks4a["server_port"] != 1080 || socks4a["version"] != "4a" {
		t.Fatalf("unexpected socks4a outbound fields: %+v", socks4a)
	}
	if socks4a["username"] != "qa-user" || socks4a["password"] != "socks4a-placeholder" {
		t.Fatalf("unexpected socks4a credential fields: %+v", socks4a)
	}

	socksVersion := findOutbound(config.Outbounds, "up_92")
	if socksVersion == nil {
		t.Fatalf("expected query-version socks outbound up_92, got %+v", config.Outbounds)
	}
	if socksVersion["type"] != "socks" || socksVersion["server"] != "socks-version.example" || socksVersion["server_port"] != 1080 || socksVersion["version"] != "4a" {
		t.Fatalf("unexpected query-version socks outbound fields: %+v", socksVersion)
	}
	if socksVersion["username"] != "qa-user" || socksVersion["password"] != "socks-version-placeholder" {
		t.Fatalf("unexpected query-version socks credential fields: %+v", socksVersion)
	}
}

func TestBuildConfigSupportsSSHQueryCredentialAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         88,
			URI:        "ssh://query-ssh.example:22?username=qa-user&pass=ssh-query-placeholder&privateKey=inline-private-placeholder&privateKeyPath=keys%2Fquery_id_ed25519&privateKeyPassphrase=query-passphrase&hostKey=ssh-ed25519%20AAAAC3NzaC1lZDI1NTE5AAAAIplaceholder&hostKeyAlgorithms=ssh-ed25519,rsa-sha2-512&clientVersion=SSH-2.0-FluxGateQuery&cipher=aes128-gcm@openssh.com,chacha20-poly1305@openssh.com&mac=hmac-sha2-256&kexAlgorithm=curve25519-sha256#ssh-query",
			Protocol:   "ssh",
			ServerPort: 22,
			Status:     "active",
		},
		{
			ID:         188,
			URI:        "ssh://algorithm-alias-ssh.example:22?userName=qa-algo-user&accountPassword=ssh-algo-placeholder&ciphers=aes128-gcm@openssh.com,chacha20-poly1305@openssh.com&macAlgorithms=hmac-sha2-256,hmac-sha2-512&kexAlgorithms=curve25519-sha256,diffie-hellman-group14-sha256#ssh-algo-alias",
			Protocol:   "ssh",
			ServerPort: 22,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_88")
	if outbound == nil {
		t.Fatalf("expected ssh outbound up_88, got %+v", config.Outbounds)
	}
	if outbound["type"] != "ssh" || outbound["server"] != "query-ssh.example" || outbound["server_port"] != 22 {
		t.Fatalf("unexpected ssh query credential server fields: %+v", outbound)
	}
	if outbound["user"] != "qa-user" || outbound["password"] != "ssh-query-placeholder" || outbound["private_key"] != "inline-private-placeholder" || outbound["private_key_path"] != "keys/query_id_ed25519" || outbound["private_key_passphrase"] != "query-passphrase" {
		t.Fatalf("unexpected ssh query credential auth fields: %+v", outbound)
	}
	hostKey, ok := outbound["host_key"].([]string)
	if !ok || len(hostKey) != 1 || hostKey[0] != "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIplaceholder" {
		t.Fatalf("unexpected ssh query credential host key: %+v", outbound["host_key"])
	}
	hostKeyAlgorithms, ok := outbound["host_key_algorithms"].([]string)
	if !ok || len(hostKeyAlgorithms) != 2 || hostKeyAlgorithms[0] != "ssh-ed25519" || hostKeyAlgorithms[1] != "rsa-sha2-512" {
		t.Fatalf("unexpected ssh query credential host key algorithms: %+v", outbound["host_key_algorithms"])
	}
	if outbound["client_version"] != "SSH-2.0-FluxGateQuery" {
		t.Fatalf("unexpected ssh query credential client version: %+v", outbound)
	}
	cipher, ok := outbound["cipher"].([]string)
	if !ok || len(cipher) != 2 || cipher[0] != "aes128-gcm@openssh.com" || cipher[1] != "chacha20-poly1305@openssh.com" {
		t.Fatalf("unexpected ssh query credential cipher: %+v", outbound["cipher"])
	}
	mac, ok := outbound["mac"].([]string)
	if !ok || len(mac) != 1 || mac[0] != "hmac-sha2-256" {
		t.Fatalf("unexpected ssh query credential mac: %+v", outbound["mac"])
	}
	kexAlgorithm, ok := outbound["kex_algorithm"].([]string)
	if !ok || len(kexAlgorithm) != 1 || kexAlgorithm[0] != "curve25519-sha256" {
		t.Fatalf("unexpected ssh query credential kex algorithm: %+v", outbound["kex_algorithm"])
	}

	algorithmAliasOutbound := findOutbound(config.Outbounds, "up_188")
	if algorithmAliasOutbound == nil {
		t.Fatalf("expected ssh outbound up_188, got %+v", config.Outbounds)
	}
	if algorithmAliasOutbound["user"] != "qa-algo-user" || algorithmAliasOutbound["password"] != "ssh-algo-placeholder" {
		t.Fatalf("unexpected ssh algorithm alias credentials: %+v", algorithmAliasOutbound)
	}
	aliasCipher, ok := algorithmAliasOutbound["cipher"].([]string)
	if !ok || len(aliasCipher) != 2 || aliasCipher[0] != "aes128-gcm@openssh.com" || aliasCipher[1] != "chacha20-poly1305@openssh.com" {
		t.Fatalf("unexpected ssh ciphers alias: %+v", algorithmAliasOutbound["cipher"])
	}
	aliasMAC, ok := algorithmAliasOutbound["mac"].([]string)
	if !ok || len(aliasMAC) != 2 || aliasMAC[0] != "hmac-sha2-256" || aliasMAC[1] != "hmac-sha2-512" {
		t.Fatalf("unexpected ssh macAlgorithms alias: %+v", algorithmAliasOutbound["mac"])
	}
	aliasKexAlgorithm, ok := algorithmAliasOutbound["kex_algorithm"].([]string)
	if !ok || len(aliasKexAlgorithm) != 2 || aliasKexAlgorithm[0] != "curve25519-sha256" || aliasKexAlgorithm[1] != "diffie-hellman-group14-sha256" {
		t.Fatalf("unexpected ssh kexAlgorithms alias: %+v", algorithmAliasOutbound["kex_algorithm"])
	}
}

func TestBuildConfigSupportsWireGuardQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         89,
			URI:        "wg://wg-query.example:51820?privateKey=wireguard-private-placeholder&publicKey=wireguard-public-placeholder&localAddresses=10.77.0.2%2F32,fd77%3A%3A2%2F128&preSharedKey=wireguard-psk-placeholder&peerAllowedIps=0.0.0.0%2F0,%3A%3A%2F0&reservedBytes=4,5,6&peerReserved=7,8,9&mtu=1280&systemInterface=1&interfaceName=wg-query#wireguard-query",
			Protocol:   "wg",
			ServerPort: 51820,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_89")
	if outbound == nil {
		t.Fatalf("expected wireguard outbound up_89, got %+v", config.Outbounds)
	}
	if outbound["type"] != "wireguard" || outbound["server"] != "wg-query.example" || outbound["server_port"] != 51820 {
		t.Fatalf("unexpected wireguard query alias server fields: %+v", outbound)
	}
	if outbound["private_key"] != "wireguard-private-placeholder" || outbound["peer_public_key"] != "wireguard-public-placeholder" || outbound["pre_shared_key"] != "wireguard-psk-placeholder" {
		t.Fatalf("unexpected wireguard query alias key fields: %+v", outbound)
	}
	localAddress, ok := outbound["local_address"].([]string)
	if !ok || len(localAddress) != 2 || localAddress[0] != "10.77.0.2/32" || localAddress[1] != "fd77::2/128" {
		t.Fatalf("unexpected wireguard query alias local addresses: %+v", outbound["local_address"])
	}
	if outbound["system_interface"] != true || outbound["interface_name"] != "wg-query" || outbound["mtu"] != 1280 {
		t.Fatalf("unexpected wireguard query alias interface fields: %+v", outbound)
	}
	reserved, ok := outbound["reserved"].([]int)
	if !ok || len(reserved) != 3 || reserved[0] != 4 || reserved[1] != 5 || reserved[2] != 6 {
		t.Fatalf("unexpected wireguard query alias reserved bytes: %+v", outbound["reserved"])
	}
	peers, ok := outbound["peers"].([]map[string]any)
	if !ok || len(peers) != 1 {
		t.Fatalf("unexpected wireguard query alias peers: %+v", outbound["peers"])
	}
	allowedIPs, ok := peers[0]["allowed_ips"].([]string)
	if !ok || len(allowedIPs) != 2 || allowedIPs[0] != "0.0.0.0/0" || allowedIPs[1] != "::/0" {
		t.Fatalf("unexpected wireguard query alias allowed ips: %+v", peers[0]["allowed_ips"])
	}
	peerReserved, ok := peers[0]["reserved"].([]int)
	if !ok || len(peerReserved) != 3 || peerReserved[0] != 7 || peerReserved[1] != 8 || peerReserved[2] != 9 {
		t.Fatalf("unexpected wireguard query alias peer reserved bytes: %+v", peers[0]["reserved"])
	}
}

func TestBuildConfigSupportsTorQueryAliases(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         90,
			URI:        "tor://default?executablePath=%2Fusr%2Fbin%2Ftor&dataDir=cache%2Ftor-query&arg=--quiet&arg=--SocksPort&arg=auto&torrc%5BClientOnly%5D=1&torrc.SocksPort=auto#tor-query",
			Protocol:   "tor",
			ServerPort: 0,
			Status:     "active",
		},
	})

	outbound := findOutbound(config.Outbounds, "up_90")
	if outbound == nil {
		t.Fatalf("expected tor outbound up_90, got %+v", config.Outbounds)
	}
	if outbound["type"] != "tor" || outbound["executable_path"] != "/usr/bin/tor" || outbound["data_directory"] != "cache/tor-query" {
		t.Fatalf("unexpected tor query alias path fields: %+v", outbound)
	}
	extraArgs, ok := outbound["extra_args"].([]string)
	if !ok || len(extraArgs) != 3 || extraArgs[0] != "--quiet" || extraArgs[1] != "--SocksPort" || extraArgs[2] != "auto" {
		t.Fatalf("unexpected tor query alias extra args: %+v", outbound["extra_args"])
	}
	torrc, ok := outbound["torrc"].(map[string]any)
	if !ok || torrc["ClientOnly"] != 1 || torrc["SocksPort"] != "auto" {
		t.Fatalf("unexpected tor query alias torrc: %+v", outbound["torrc"])
	}
}

func gatewayToken(tokenStatus, accountStatus, protocol string, expireAt *time.Time, quotaBytes, usedUploadBytes, usedDownloadBytes int64, authUser string) store.TokenWithAccount {
	return store.TokenWithAccount{
		Token: store.Token{
			Status:            tokenStatus,
			ExpireAt:          expireAt,
			QuotaBytes:        quotaBytes,
			UsedUploadBytes:   usedUploadBytes,
			UsedDownloadBytes: usedDownloadBytes,
		},
		GatewayAccount: store.GatewayAccount{
			Status:   accountStatus,
			Protocol: protocol,
			AuthUser: authUser,
			UUID:     "00000000-0000-0000-0000-000000000000",
		},
	}
}

func findOutbound(outbounds []map[string]any, tag string) map[string]any {
	for _, outbound := range outbounds {
		if outbound["tag"] == tag {
			return outbound
		}
	}
	return nil
}

func vmessURI(t *testing.T, doc map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal vmess doc: %v", err)
	}
	return "vmess://" + base64.RawURLEncoding.EncodeToString(payload)
}
