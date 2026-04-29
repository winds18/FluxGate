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
	}, nil, now)

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
		"add":  "vmess.example.net",
		"port": "443",
		"id":   "00000000-0000-0000-0000-000000000046",
		"aid":  "0",
		"scy":  "auto",
		"net":  "ws",
		"host": "ws.example.net",
		"path": "/ws",
		"tls":  "tls",
		"sni":  "vmess.example.net",
		"ps":   "vmess",
	})
	config := buildConfig(nil, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, []store.Node{
		{
			ID:         42,
			URI:        "vless://00000000-0000-0000-0000-000000000042@example.com:443?security=tls&sni=edge.example.com&flow=xtls-rprx-vision#hk",
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
			URI:        "trojan://qa-placeholder@example.org:443?security=tls&sni=trojan.example.org#trojan",
			Protocol:   "trojan",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         45,
			URI:        "ss://" + ssCredential + "@example.net:8388#ss",
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
			URI:        "hysteria2://qa-placeholder@example.dev:443?obfs=salamander&obfs-password=obfs-placeholder&sni=hy2.example.dev&insecure=1#hy2",
			Protocol:   "hysteria2",
			ServerPort: 443,
			Status:     "active",
		},
		{
			ID:         48,
			URI:        "tuic://placeholder@example.org:443#unsupported",
			Protocol:   "tuic",
			ServerPort: 443,
			Status:     "active",
		},
	}, time.Date(2026, 4, 29, 1, 17, 0, 0, time.UTC))

	if config.Route["final"] != upstreamSelectorTag {
		t.Fatalf("expected route final to selector, got %+v", config.Route)
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
	tls, ok := vless["tls"].(map[string]any)
	if !ok || tls["enabled"] != true || tls["server_name"] != "edge.example.com" {
		t.Fatalf("unexpected tls config: %+v", vless["tls"])
	}
	trojan := findOutbound(config.Outbounds, "up_44")
	if trojan == nil {
		t.Fatalf("expected trojan outbound up_44, got %+v", config.Outbounds)
	}
	if trojan["type"] != "trojan" || trojan["server"] != "example.org" || trojan["password"] != "qa-placeholder" {
		t.Fatalf("unexpected trojan outbound fields: %+v", trojan)
	}
	trojanTLS, ok := trojan["tls"].(map[string]any)
	if !ok || trojanTLS["enabled"] != true || trojanTLS["server_name"] != "trojan.example.org" {
		t.Fatalf("unexpected trojan tls config: %+v", trojan["tls"])
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
	if !ok || vmessTLS["enabled"] != true || vmessTLS["server_name"] != "vmess.example.net" {
		t.Fatalf("unexpected vmess tls config: %+v", vmess["tls"])
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
	hysteria2TLS, ok := hysteria2["tls"].(map[string]any)
	if !ok || hysteria2TLS["enabled"] != true || hysteria2TLS["server_name"] != "hy2.example.dev" || hysteria2TLS["insecure"] != true {
		t.Fatalf("unexpected hysteria2 tls config: %+v", hysteria2["tls"])
	}
	obfs, ok := hysteria2["obfs"].(map[string]any)
	if !ok || obfs["type"] != "salamander" || obfs["password"] != "obfs-placeholder" {
		t.Fatalf("unexpected hysteria2 obfs config: %+v", hysteria2["obfs"])
	}
	if findOutbound(config.Outbounds, "up_43") != nil || findOutbound(config.Outbounds, "up_48") != nil {
		t.Fatalf("inactive or unsupported nodes should be skipped: %+v", config.Outbounds)
	}

	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 5 || tags[0] != "up_42" || tags[1] != "up_44" || tags[2] != "up_45" || tags[3] != "up_46" || tags[4] != "up_47" || selector["default"] != "up_42" {
		t.Fatalf("unexpected selector outbounds: %+v", selector)
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
