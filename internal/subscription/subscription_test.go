package subscription

import (
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestSetUserInfoHeaderIncludesTrafficQuota(t *testing.T) {
	expireAt := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	header := http.Header{}

	SetUserInfoHeader(header, store.TokenWithAccount{
		Token: store.Token{
			Status:            "active",
			ExpireAt:          &expireAt,
			QuotaBytes:        1000,
			UsedUploadBytes:   200,
			UsedDownloadBytes: 300,
		},
	})

	wantUserInfo := "upload=200; download=300; total=1000; expire=1893456000"
	if got := header.Get("subscription-userinfo"); got != wantUserInfo {
		t.Fatalf("unexpected subscription-userinfo: %q", got)
	}
	for name, want := range map[string]string{
		"profile-update-interval":    "24",
		"x-fluxgate-upload-bytes":    "200",
		"x-fluxgate-download-bytes":  "300",
		"x-fluxgate-used-bytes":      "500",
		"x-fluxgate-quota-bytes":     "1000",
		"x-fluxgate-remaining-bytes": "500",
	} {
		if got := header.Get(name); got != want {
			t.Fatalf("unexpected %s: got %q want %q", name, got, want)
		}
	}
}

func TestSetUserInfoHeaderClampsRemainingQuota(t *testing.T) {
	header := http.Header{}

	SetUserInfoHeader(header, store.TokenWithAccount{
		Token: store.Token{
			QuotaBytes:        400,
			UsedUploadBytes:   250,
			UsedDownloadBytes: 300,
		},
	})

	if got := header.Get("x-fluxgate-remaining-bytes"); got != "0" {
		t.Fatalf("remaining quota should not go negative: %q", got)
	}
}
