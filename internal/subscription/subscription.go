package subscription

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

type Request struct {
	Target        string
	GatewayHost   string
	PublicBaseURL string
}

type Response struct {
	ContentType string
	Body        []byte
}

func Build(req Request, token store.TokenWithAccount, virtualNodes []store.VirtualNode) (Response, error) {
	target := strings.ToLower(strings.TrimSpace(req.Target))
	if target == "" {
		target = "clash"
	}

	switch target {
	case "sing-box", "singbox":
		body, err := buildSingBox(req, token, virtualNodes)
		return Response{ContentType: "application/json; charset=utf-8", Body: body}, err
	default:
		body := buildClash(req, token, virtualNodes)
		return Response{ContentType: "text/yaml; charset=utf-8", Body: []byte(body)}, nil
	}
}

func SetUserInfoHeader(header http.Header, token store.TokenWithAccount) {
	total := token.QuotaBytes
	used := token.UsedUploadBytes + token.UsedDownloadBytes
	expire := int64(0)
	if token.ExpireAt != nil {
		expire = token.ExpireAt.Unix()
	}
	header.Set("subscription-userinfo", fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", token.UsedUploadBytes, token.UsedDownloadBytes, total, expire))
	header.Set("profile-update-interval", "24")
	header.Set("x-fluxgate-used-bytes", fmt.Sprintf("%d", used))
}

func TokenUsable(token store.TokenWithAccount, now time.Time) (bool, string) {
	if token.Status != "active" {
		return false, "token is not active"
	}
	if token.GatewayAccount.Status != "active" {
		return false, "gateway account is not active"
	}
	if token.ExpireAt != nil && !token.ExpireAt.After(now) {
		return false, "token expired"
	}
	if token.QuotaBytes > 0 && token.UsedUploadBytes+token.UsedDownloadBytes >= token.QuotaBytes {
		return false, "token over quota"
	}
	return true, ""
}

func buildClash(req Request, token store.TokenWithAccount, virtualNodes []store.VirtualNode) string {
	var b strings.Builder
	b.WriteString("mixed-port: 7890\n")
	b.WriteString("allow-lan: false\n")
	b.WriteString("mode: rule\n")
	b.WriteString("log-level: info\n")
	b.WriteString("proxies:\n")
	for _, node := range virtualNodes {
		if node.Status != "active" || node.ListenProtocol != "vless" {
			continue
		}
		fmt.Fprintf(&b, "  - name: %q\n", node.Name)
		b.WriteString("    type: vless\n")
		fmt.Fprintf(&b, "    server: %s\n", req.GatewayHost)
		fmt.Fprintf(&b, "    port: %d\n", node.ListenPort)
		fmt.Fprintf(&b, "    uuid: %s\n", token.GatewayAccount.UUID)
		b.WriteString("    udp: true\n")
		b.WriteString("    network: tcp\n")
	}
	b.WriteString("proxy-groups:\n")
	b.WriteString("  - name: FluxGate\n")
	b.WriteString("    type: select\n")
	b.WriteString("    proxies:\n")
	hasProxy := false
	for _, node := range virtualNodes {
		if node.Status != "active" || node.ListenProtocol != "vless" {
			continue
		}
		hasProxy = true
		fmt.Fprintf(&b, "      - %q\n", node.Name)
	}
	if !hasProxy {
		b.WriteString("      - DIRECT\n")
	}
	b.WriteString("rules:\n")
	b.WriteString("  - MATCH,FluxGate\n")
	return b.String()
}

func buildSingBox(req Request, token store.TokenWithAccount, virtualNodes []store.VirtualNode) ([]byte, error) {
	outbounds := []map[string]any{
		{"type": "direct", "tag": "direct"},
	}
	for _, node := range virtualNodes {
		if node.Status != "active" || node.ListenProtocol != "vless" {
			continue
		}
		outbounds = append(outbounds, map[string]any{
			"type":        "vless",
			"tag":         node.Name,
			"server":      req.GatewayHost,
			"server_port": node.ListenPort,
			"uuid":        token.GatewayAccount.UUID,
			"network":     "tcp",
		})
	}
	doc := map[string]any{
		"log":       map[string]any{"level": "info"},
		"outbounds": outbounds,
		"route":     map[string]any{"final": "direct"},
	}
	return json.MarshalIndent(doc, "", "  ")
}
