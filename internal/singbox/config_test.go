package singbox

import (
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
			URI:        "ss://placeholder@example.org:443#unsupported",
			Protocol:   "ss",
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
	if findOutbound(config.Outbounds, "up_43") != nil || findOutbound(config.Outbounds, "up_45") != nil {
		t.Fatalf("inactive or unsupported nodes should be skipped: %+v", config.Outbounds)
	}

	selector := findOutbound(config.Outbounds, upstreamSelectorTag)
	if selector == nil {
		t.Fatalf("expected upstream selector, got %+v", config.Outbounds)
	}
	tags, ok := selector["outbounds"].([]string)
	if !ok || len(tags) != 2 || tags[0] != "up_42" || tags[1] != "up_44" || selector["default"] != "up_42" {
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
