package singbox

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

type Config struct {
	Log          map[string]any   `json:"log"`
	Inbounds     []Inbound        `json:"inbounds"`
	Outbounds    []map[string]any `json:"outbounds"`
	Route        map[string]any   `json:"route"`
	Experimental map[string]any   `json:"experimental,omitempty"`
}

type Inbound struct {
	Type       string      `json:"type"`
	Tag        string      `json:"tag"`
	Listen     string      `json:"listen"`
	ListenPort int         `json:"listen_port"`
	Users      []VLESSUser `json:"users"`
}

type VLESSUser struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Flow string `json:"flow"`
}

const upstreamSelectorTag = "FluxGate-upstreams"

func BuildConfig(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode, upstreamNodes []store.Node) Config {
	return buildConfig(tokens, virtualNodes, upstreamNodes, time.Now().UTC())
}

func buildConfig(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode, upstreamNodes []store.Node, now time.Time) Config {
	users := make([]VLESSUser, 0, len(tokens))
	for _, token := range tokens {
		if !gatewayTokenUsable(token, now) {
			continue
		}
		users = append(users, VLESSUser{
			Name: token.GatewayAccount.AuthUser,
			UUID: token.GatewayAccount.UUID,
			Flow: "",
		})
	}

	inbounds := make([]Inbound, 0, len(virtualNodes))
	for _, node := range virtualNodes {
		if node.Status != "active" || node.ListenProtocol != "vless" {
			continue
		}
		inbounds = append(inbounds, Inbound{
			Type:       "vless",
			Tag:        "vn-" + node.Name,
			Listen:     "::",
			ListenPort: node.ListenPort,
			Users:      users,
		})
	}

	upstreamOutbounds, upstreamTags := buildUpstreamOutbounds(upstreamNodes)
	outbounds := []map[string]any{
		{"type": "direct", "tag": "direct"},
		{"type": "block", "tag": "block"},
	}
	outbounds = append(outbounds, upstreamOutbounds...)
	finalOutbound := "direct"
	if len(upstreamTags) > 0 {
		finalOutbound = upstreamSelectorTag
		outbounds = append(outbounds, map[string]any{
			"type":      "selector",
			"tag":       upstreamSelectorTag,
			"outbounds": upstreamTags,
			"default":   upstreamTags[0],
		})
	}

	return Config{
		Log: map[string]any{
			"level": "info",
		},
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Route: map[string]any{
			"final": finalOutbound,
		},
		Experimental: map[string]any{
			"v2ray_api": map[string]any{
				"listen": "0.0.0.0:9090",
				"stats": map[string]any{
					"enabled":  true,
					"inbounds": inboundTags(inbounds),
					"users":    userNames(users),
				},
			},
		},
	}
}

func gatewayTokenUsable(token store.TokenWithAccount, now time.Time) bool {
	if token.Status != "active" {
		return false
	}
	if token.GatewayAccount.Status != "active" {
		return false
	}
	if token.GatewayAccount.Protocol != "vless" {
		return false
	}
	if token.ExpireAt != nil && !token.ExpireAt.After(now) {
		return false
	}
	if token.QuotaBytes > 0 && token.UsedUploadBytes+token.UsedDownloadBytes >= token.QuotaBytes {
		return false
	}
	return true
}

func buildUpstreamOutbounds(nodes []store.Node) ([]map[string]any, []string) {
	outbounds := make([]map[string]any, 0, len(nodes))
	tags := make([]string, 0, len(nodes))
	for _, node := range nodes {
		outbound, ok := buildVLESSOutbound(node)
		if !ok {
			continue
		}
		tags = append(tags, outbound["tag"].(string))
		outbounds = append(outbounds, outbound)
	}
	return outbounds, tags
}

func buildVLESSOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "vless" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "vless" || parsed.Hostname() == "" {
		return nil, false
	}
	uuid := parsed.User.Username()
	if uuid == "" {
		return nil, false
	}
	port := node.ServerPort
	if port <= 0 {
		port = 443
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type":        "vless",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": port,
		"uuid":        uuid,
	}
	if flow := strings.TrimSpace(query.Get("flow")); flow != "" {
		outbound["flow"] = flow
	}
	if strings.EqualFold(query.Get("security"), "tls") {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
			tls["server_name"] = serverName
		}
		outbound["tls"] = tls
	}
	return outbound, true
}

func upstreamTag(node store.Node) string {
	return fmt.Sprintf("up_%d", node.ID)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func Marshal(config Config) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}

func inboundTags(inbounds []Inbound) []string {
	tags := make([]string, 0, len(inbounds))
	for _, inbound := range inbounds {
		tags = append(tags, inbound.Tag)
	}
	return tags
}

func userNames(users []VLESSUser) []string {
	names := make([]string, 0, len(users))
	for _, user := range users {
		names = append(names, user.Name)
	}
	return names
}
