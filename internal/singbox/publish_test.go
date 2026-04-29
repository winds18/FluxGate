package singbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

func TestPublishConfigWritesCurrentAndPreviousConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	previousPath := filepath.Join(dir, "config.previous.json")
	if err := os.WriteFile(configPath, []byte(`{"old":true}`), 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	config := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "qa-user"),
	}, []store.VirtualNode{
		{Name: "hk", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, nil, time.Date(2026, 4, 29, 3, 49, 0, 0, time.UTC))

	result, err := PublishConfig(config, configPath, previousPath)
	if err != nil {
		t.Fatalf("publish config: %v", err)
	}
	if !result.Valid || !result.Published || !result.PreviousSaved || result.ConfigHash == "" {
		t.Fatalf("unexpected publish result: %+v", result)
	}
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), `"inbounds"`) {
		t.Fatalf("published config missing inbounds: %s", string(body))
	}
	previous, err := os.ReadFile(previousPath)
	if err != nil {
		t.Fatalf("read previous config: %v", err)
	}
	if string(previous) != `{"old":true}` {
		t.Fatalf("unexpected previous config: %s", string(previous))
	}
}

func TestPublishConfigRejectsInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	result, err := PublishConfig(Config{}, filepath.Join(dir, "config.json"), filepath.Join(dir, "previous.json"))
	if err == nil {
		t.Fatal("expected invalid config error")
	}
	if result.Valid || result.Published {
		t.Fatalf("invalid config should not publish: %+v", result)
	}
}
