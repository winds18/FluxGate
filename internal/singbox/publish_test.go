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
	if !result.Valid || !result.Published || !result.PreviousSaved || !result.RestartRequired || result.ConfigHash == "" {
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
	if result.Valid || result.Published || result.RestartRequired {
		t.Fatalf("invalid config should not publish: %+v", result)
	}
}

func TestRollbackConfigRestoresPreviousConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	previousPath := filepath.Join(dir, "config.previous.json")

	currentConfig := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "current-user"),
	}, []store.VirtualNode{
		{Name: "current", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}, nil, time.Date(2026, 4, 29, 3, 59, 0, 0, time.UTC))
	previousConfig := buildConfig([]store.TokenWithAccount{
		gatewayToken("active", "active", "vless", nil, 0, 0, 0, "previous-user"),
	}, []store.VirtualNode{
		{Name: "previous", ListenProtocol: "vless", ListenPort: 9443, Status: "active"},
	}, nil, time.Date(2026, 4, 29, 3, 59, 0, 0, time.UTC))

	currentBody, err := Marshal(currentConfig)
	if err != nil {
		t.Fatalf("marshal current: %v", err)
	}
	previousBody, err := Marshal(previousConfig)
	if err != nil {
		t.Fatalf("marshal previous: %v", err)
	}
	if err := os.WriteFile(configPath, currentBody, 0o644); err != nil {
		t.Fatalf("seed current: %v", err)
	}
	if err := os.WriteFile(previousPath, previousBody, 0o644); err != nil {
		t.Fatalf("seed previous: %v", err)
	}

	result, err := RollbackConfig(configPath, previousPath)
	if err != nil {
		t.Fatalf("rollback config: %v", err)
	}
	if !result.Valid || !result.RolledBack || !result.RestartRequired || result.ConfigHash == "" {
		t.Fatalf("unexpected rollback result: %+v", result)
	}
	restored, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read restored config: %v", err)
	}
	if string(restored) != string(previousBody) {
		t.Fatalf("expected previous config restored")
	}
}
