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
			URI:        "vless://00000000-0000-0000-0000-000000000042@example.com:443?security=tls&sni=edge.example.com&flow=xtls-rprx-vision&insecure=1&disable_sni=1&alpn=h2,http/1.1&type=ws&path=%2Fvless&host=ws.example.com#hk",
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
			URI:        "tuic://00000000-0000-0000-0000-000000000048:qa-placeholder@example.io:443?congestion_control=bbr&udp_relay_mode=native&sni=tuic.example.io&alpn=h3&insecure=1#tuic",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         49,
			URI:        "anytls://qa-placeholder@example.chat:443?sni=anytls.example.chat&alpn=h2,http/1.1&idle_session_check_interval=20s&idle_session_timeout=45s&min_idle_session=2&insecure=1#anytls",
			Protocol:   "anytls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         50,
			URI:        "shadowtls://qa-placeholder@example.help:443?version=3&sni=shadow.example.help&alpn=h2&insecure=1#shadowtls",
			Protocol:   "shadowtls",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         51,
			URI:        "naive://qa-user:qa-placeholder@example.news:443?sni=naive.example.news&quic=1&quic_congestion_control=bbr&udp_over_tcp=1&insecure_concurrency=2&alpn=h3&insecure=1&disable_sni=1#naive",
			Protocol:   "naive",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         52,
			URI:        "hysteria://qa-placeholder@example.zone:443?auth_str=qa-auth&up_mbps=20&down_mbps=80&obfs=obfs-placeholder&recv_window_conn=1048576&recv_window=2097152&disable_mtu_discovery=1&protocol=udp&sni=hysteria.example.zone&alpn=h3&insecure=1#hysteria",
			Protocol:   "hysteria",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         53,
			URI:        "https://qa-user:qa-placeholder@example.proxy:8443/connect?sni=http-proxy.example.proxy&skip-cert-verify=1&disable_sni=1&alpn=h2,http%2F1.1#http",
			Protocol:   "https",
			ServerPort: 8443,
			Status:     "active",
		},
		{
			ID:         54,
			URI:        "socks5://qa-user:qa-placeholder@example.socks:1080?network=udp&udp_over_tcp=1#socks",
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
			URI:        "direct://default#Direct",
			Protocol:   "direct",
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
			URI:        "block://default#Block",
			Protocol:   "block",
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
			URI:        "vless://00000000-0000-0000-0000-000000000077@example.grpc:443?security=tls&type=grpc&service_name=fluxgate&idle_timeout=30s&ping_timeout=10s&permit_without_stream=1&multi_mode=1#grpc-keepalive",
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
			URI:        "vless://00000000-0000-0000-0000-000000000074@example.http:443?security=tls&type=http&host=h2.example.test,h2-backup.example.test&path=/h2&method=GET&idle_timeout=20s&ping_timeout=10s#http",
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
			URI:        "trojan://trojan-placeholder@example.upgrade:443?security=tls&type=httpupgrade&host=upgrade.example.test&path=/upgrade#upgrade",
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

func TestBuildConfigPreservesVMessGRPCTransport(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID: 71,
			URI: vmessURI(t, map[string]any{
				"add":        "vmess.grpc.example",
				"port":       "443",
				"id":         "00000000-0000-0000-0000-000000000071",
				"aid":        "0",
				"scy":        "auto",
				"net":        "grpc",
				"path":       "fluxgate-vmess",
				"multi_mode": "1",
				"tls":        "tls",
				"sni":        "vmess.grpc.example",
				"ps":         "vmess-grpc",
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

func TestBuildConfigSupportsVMessUserinfoURI(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         72,
			URI:        "vmess://00000000-0000-0000-0000-000000000072@userinfo.vmess.example:443?encryption=auto&security=tls&type=ws&path=%2Fvmess&host=ws.vmess.example&sni=sni.vmess.example&alpn=h2,http%2F1.1&insecure=1&disable_sni=1&fp=chrome#VMess%20Userinfo",
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

func TestBuildConfigSupportsHysteria2QueryPassword(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         80,
			URI:        "hy2://query-password.hy2.example:443?password=hy2-query-placeholder&obfs=salamander&obfs-password=obfs-placeholder&sni=hy2-query.example&insecure=1#hy2-query",
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
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "hy2-query.example" || tls["insecure"] != true {
		t.Fatalf("unexpected hysteria2 query password tls: %+v", outbound["tls"])
	}
}

func TestBuildConfigSupportsHysteriaTokenAuth(t *testing.T) {
	config := BuildConfig(nil, nil, []store.Node{
		{
			ID:         81,
			URI:        "hysteria://token-auth.hysteria.example:443?token=hysteria-token-placeholder&up_mbps=30&down_mbps=90&obfs=obfs-placeholder&sni=hysteria-token.example&insecure=1#hysteria-token",
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
	tls, ok := outbound["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "hysteria-token.example" || tls["insecure"] != true {
		t.Fatalf("unexpected hysteria token auth tls: %+v", outbound["tls"])
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
