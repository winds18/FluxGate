package subscription

import (
	"strings"
	"testing"

	"github.com/winds18/FluxGate/internal/store"
)

func TestBuildClashSubscription(t *testing.T) {
	resp, err := Build(Request{
		Target:      "clash",
		GatewayHost: "gateway.example.com",
	}, store.TokenWithAccount{
		GatewayAccount: store.GatewayAccount{UUID: "00000000-0000-4000-8000-000000000000"},
	}, []store.VirtualNode{
		{Name: "FluxGate-HK", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	})
	if err != nil {
		t.Fatalf("build subscription: %v", err)
	}
	body := string(resp.Body)
	for _, want := range []string{"FluxGate-HK", "gateway.example.com", "8443"} {
		if !strings.Contains(body, want) {
			t.Fatalf("subscription missing %q:\n%s", want, body)
		}
	}
}
