package httpapi

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/winds18/FluxGate/internal/config"
	"github.com/winds18/FluxGate/internal/store"
)

func TestPublicBaseURLUsesCurrentRequestHost(t *testing.T) {
	server := &Server{cfg: config.Config{PublicBaseURL: "https://configured.example"}}
	request := httptest.NewRequest("GET", "http://192.168.66.10:18080/api/tokens", nil)

	if got, want := server.publicBaseURL(request), "http://192.168.66.10:18080"; got != want {
		t.Fatalf("public base URL mismatch: got %q want %q", got, want)
	}
}

func TestPublicBaseURLUsesForwardedHeaders(t *testing.T) {
	server := &Server{cfg: config.Config{PublicBaseURL: "https://configured.example"}}
	request := httptest.NewRequest("GET", "http://127.0.0.1:8080/api/tokens", nil)
	request.Header.Set("x-forwarded-host", "flux.example.com")
	request.Header.Set("x-forwarded-proto", "https")

	if got, want := server.publicBaseURL(request), "https://flux.example.com"; got != want {
		t.Fatalf("public base URL mismatch: got %q want %q", got, want)
	}
}

func TestPublicBaseURLFallsBackForUnsafeHost(t *testing.T) {
	server := &Server{cfg: config.Config{PublicBaseURL: "https://configured.example/"}}
	request := httptest.NewRequest("GET", "http://127.0.0.1:8080/api/tokens", nil)
	request.Host = "bad.example/@token"

	if got, want := server.publicBaseURL(request), "https://configured.example"; got != want {
		t.Fatalf("public base URL mismatch: got %q want %q", got, want)
	}
}

func TestPublicGatewayHostUsesCurrentRequestHostWithoutManagementPort(t *testing.T) {
	server := &Server{cfg: config.Config{GatewayHost: "gateway.example.com"}}
	request := httptest.NewRequest("GET", "http://192.168.66.10:18080/sub/token?target=clash", nil)

	if got, want := server.publicGatewayHost(request), "192.168.66.10"; got != want {
		t.Fatalf("public gateway host mismatch: got %q want %q", got, want)
	}
}

func TestPublicGatewayHostUsesForwardedHost(t *testing.T) {
	server := &Server{cfg: config.Config{GatewayHost: "gateway.example.com"}}
	request := httptest.NewRequest("GET", "http://127.0.0.1:8080/sub/token?target=clash", nil)
	request.Header.Set("x-forwarded-host", "fluxgate.shimo.proxy.nbyitian.top")

	if got, want := server.publicGatewayHost(request), "fluxgate.shimo.proxy.nbyitian.top"; got != want {
		t.Fatalf("public gateway host mismatch: got %q want %q", got, want)
	}
}

func TestPublicGatewayHostFallsBackToConfiguredGatewayHost(t *testing.T) {
	server := &Server{cfg: config.Config{
		GatewayHost:   "edge.example.com",
		PublicBaseURL: "https://configured.example/",
	}}
	request := httptest.NewRequest("GET", "http://127.0.0.1:8080/sub/token?target=clash", nil)
	request.Host = "bad.example/@token"

	if got, want := server.publicGatewayHost(request), "edge.example.com"; got != want {
		t.Fatalf("public gateway host mismatch: got %q want %q", got, want)
	}
}

func TestDeliveryReadinessReflectsUsableInstance(t *testing.T) {
	ctx := context.Background()
	db := openHTTPTestStore(t)
	server := &Server{
		cfg:   config.Config{Version: "test", TokenSecret: "test-secret"},
		store: db,
	}

	empty, err := server.deliveryReadiness(ctx)
	if err != nil {
		t.Fatalf("empty delivery readiness: %v", err)
	}
	if empty.Ready {
		t.Fatalf("empty instance should not be ready: %+v", empty)
	}
	if empty.ReadyCount != 0 || empty.TotalChecks != 6 || len(empty.Checks) != 6 {
		t.Fatalf("unexpected empty readiness shape: %+v", empty)
	}

	team, err := db.CreateTeam(ctx, store.CreateTeamInput{Name: "QA Team"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	user, err := db.CreateUser(ctx, store.CreateUserInput{TeamID: &team.ID, Name: "QA User"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	source, err := db.CreateSource(ctx, store.CreateSourceInput{Name: "QA Source", Type: "manual", DefaultTags: `["QA-HK"]`})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if _, err := db.ImportNodes(ctx, store.ImportNodesInput{
		SourceID: source.ID,
		Content:  "vless://00000000-0000-0000-0000-000000000001@example.com:443#香港%2001",
	}); err != nil {
		t.Fatalf("import nodes: %v", err)
	}
	if _, err := db.CreateToken(ctx, "test-secret", "https://flux.example", store.CreateTokenInput{
		UserID:     user.ID,
		Name:       "QA Token",
		ExpireDays: 30,
		QuotaBytes: 1 << 30,
	}); err != nil {
		t.Fatalf("create token: %v", err)
	}
	if _, err := db.CreateVirtualNode(ctx, store.CreateVirtualNodeInput{
		Name:           "FluxGate-HK",
		ListenProtocol: "vless",
		ListenPort:     8443,
		TagSelector:    `{"include":["QA-HK"]}`,
		Strategy:       "selector",
	}); err != nil {
		t.Fatalf("create virtual node: %v", err)
	}
	if _, err := db.CreatePolicy(ctx, store.CreatePolicyInput{
		Name:                "QA Policy",
		ScopeType:           "team",
		ScopeID:             &team.ID,
		AllowedVirtualNodes: `["FluxGate-HK"]`,
		MaxNodes:            1,
	}); err != nil {
		t.Fatalf("create policy: %v", err)
	}

	ready, err := server.deliveryReadiness(ctx)
	if err != nil {
		t.Fatalf("delivery readiness: %v", err)
	}
	if !ready.Ready || ready.ReadyCount != ready.TotalChecks {
		t.Fatalf("expected usable instance to be ready: %+v", ready)
	}
	if !ready.Config.Valid || ready.Config.InboundCount < 1 || ready.Config.UserCount < 1 || ready.Config.UpstreamOutboundCount < 1 {
		t.Fatalf("expected usable sing-box config summary: %+v", ready.Config)
	}
	if len(ready.NextActions) != 2 || ready.NextActions[0] != "复制 Token 订阅地址导入客户端" {
		t.Fatalf("unexpected next actions for ready instance: %+v", ready.NextActions)
	}
}

func openHTTPTestStore(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "fluxgate.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
