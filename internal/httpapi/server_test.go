package httpapi

import (
	"net/http/httptest"
	"testing"

	"github.com/winds18/FluxGate/internal/config"
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
