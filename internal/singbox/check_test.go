package singbox

import (
	"testing"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

func TestCheckConfigSummarizesGeneratedConfig(t *testing.T) {
	future := time.Date(2026, 4, 29, 4, 0, 0, 0, time.UTC)
	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", &future, 0, 0, 0, "qa-user"),
	}, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, []store.Node{
		{ID: 42, URI: "vless://00000000-0000-0000-0000-000000000042@example.com:443#hk", Protocol: "vless", ServerPort: 443, Status: "active"},
	}, time.Date(2026, 4, 29, 3, 41, 0, 0, time.UTC))

	result, err := CheckConfig(config)
	if err != nil {
		t.Fatalf("check config: %v", err)
	}
	if !result.Valid || result.ConfigHash == "" {
		t.Fatalf("expected valid config with hash, got %+v", result)
	}
	if result.InboundCount != 1 || result.OutboundCount != 4 || result.UpstreamOutboundCount != 1 || result.UserCount != 1 {
		t.Fatalf("unexpected config summary: %+v", result)
	}
}

func TestCheckConfigRejectsInvalidShape(t *testing.T) {
	result, err := CheckConfig(Config{
		Log:      map[string]any{"level": "info"},
		Inbounds: []Inbound{{Type: "vless", Tag: "", ListenPort: 0}},
		Route:    map[string]any{"final": "direct"},
	})
	if err != nil {
		t.Fatalf("check config: %v", err)
	}
	if result.Valid || len(result.Messages) == 0 {
		t.Fatalf("expected invalid config messages, got %+v", result)
	}
}
