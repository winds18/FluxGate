package singbox

import (
	"encoding/json"
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

func BuildConfig(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode) Config {
	return buildConfig(tokens, virtualNodes, time.Now().UTC())
}

func buildConfig(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode, now time.Time) Config {
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

	return Config{
		Log: map[string]any{
			"level": "info",
		},
		Inbounds: inbounds,
		Outbounds: []map[string]any{
			{"type": "direct", "tag": "direct"},
			{"type": "block", "tag": "block"},
		},
		Route: map[string]any{
			"final": "direct",
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
