package singbox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
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
		outbound, ok := buildNodeOutbound(node)
		if !ok {
			continue
		}
		tags = append(tags, outbound["tag"].(string))
		outbounds = append(outbounds, outbound)
	}
	return outbounds, tags
}

func buildNodeOutbound(node store.Node) (map[string]any, bool) {
	switch node.Protocol {
	case "vless":
		return buildVLESSOutbound(node)
	case "trojan":
		return buildTrojanOutbound(node)
	case "ss":
		return buildShadowsocksOutbound(node)
	case "vmess":
		return buildVMessOutbound(node)
	default:
		return nil, false
	}
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

func buildTrojanOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "trojan" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "trojan" || parsed.Hostname() == "" {
		return nil, false
	}
	password := parsed.User.Username()
	if password == "" {
		return nil, false
	}
	port := node.ServerPort
	if port <= 0 {
		port = 443
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type":        "trojan",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": port,
		"password":    password,
	}
	if strings.EqualFold(query.Get("security"), "tls") || firstNonEmpty(query.Get("sni"), query.Get("peer")) != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(query.Get("sni"), query.Get("peer"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
			tls["server_name"] = serverName
		}
		outbound["tls"] = tls
	}
	return outbound, true
}

func buildShadowsocksOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "ss" {
		return nil, false
	}
	method, password, server, port, ok := parseShadowsocksURI(node.URI, node.ServerPort)
	if !ok {
		return nil, false
	}
	return map[string]any{
		"type":        "shadowsocks",
		"tag":         upstreamTag(node),
		"server":      server,
		"server_port": port,
		"method":      method,
		"password":    password,
	}, true
}

func buildVMessOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "vmess" {
		return nil, false
	}
	doc, ok := parseVMessURI(node.URI)
	if !ok {
		return nil, false
	}
	port := intFromAny(doc["port"])
	if port <= 0 {
		port = node.ServerPort
	}
	if port <= 0 {
		port = 443
	}
	server := strings.TrimSpace(stringFromAny(doc["add"]))
	uuid := strings.TrimSpace(stringFromAny(doc["id"]))
	if server == "" || uuid == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "vmess",
		"tag":         upstreamTag(node),
		"server":      server,
		"server_port": port,
		"uuid":        uuid,
	}
	if security := strings.TrimSpace(stringFromAny(doc["scy"])); security != "" {
		outbound["security"] = security
	}
	if alterID := intFromAny(doc["aid"]); alterID > 0 {
		outbound["alter_id"] = alterID
	}
	if strings.EqualFold(stringFromAny(doc["tls"]), "tls") || strings.TrimSpace(stringFromAny(doc["sni"])) != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(stringFromAny(doc["sni"]), stringFromAny(doc["host"]), server); serverName != "" {
			tls["server_name"] = serverName
		}
		outbound["tls"] = tls
	}
	if strings.EqualFold(stringFromAny(doc["net"]), "ws") {
		transport := map[string]any{"type": "ws"}
		if path := strings.TrimSpace(stringFromAny(doc["path"])); path != "" {
			transport["path"] = path
		}
		if host := strings.TrimSpace(stringFromAny(doc["host"])); host != "" {
			transport["headers"] = map[string]any{"Host": host}
		}
		outbound["transport"] = transport
	}
	return outbound, true
}

func parseVMessURI(rawURI string) (map[string]any, bool) {
	rawURI = strings.TrimSpace(rawURI)
	if !strings.HasPrefix(rawURI, "vmess://") {
		return nil, false
	}
	payload := strings.TrimPrefix(rawURI, "vmess://")
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, false
	}
	decoded, ok := decodeBase64URL(payload)
	if !ok {
		return nil, false
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(decoded), &doc); err != nil {
		return nil, false
	}
	return doc, true
}

func parseShadowsocksURI(rawURI string, fallbackPort int) (string, string, string, int, bool) {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil || parsed.Scheme != "ss" {
		return "", "", "", 0, false
	}
	if parsed.User != nil && parsed.Hostname() != "" {
		method, password, ok := shadowsocksUserInfo(parsed.User)
		if !ok {
			return "", "", "", 0, false
		}
		return method, password, parsed.Hostname(), portWithFallback(parsed.Port(), fallbackPort, 8388), true
	}

	encoded := strings.TrimPrefix(strings.TrimSpace(rawURI), "ss://")
	encoded = strings.SplitN(encoded, "#", 2)[0]
	encoded = strings.SplitN(encoded, "?", 2)[0]
	decoded, ok := decodeBase64URL(encoded)
	if !ok {
		return "", "", "", 0, false
	}
	legacy, err := url.Parse("ss://" + decoded)
	if err != nil || legacy.User == nil || legacy.Hostname() == "" {
		return "", "", "", 0, false
	}
	method, password, ok := shadowsocksUserInfo(legacy.User)
	if !ok {
		return "", "", "", 0, false
	}
	return method, password, legacy.Hostname(), portWithFallback(legacy.Port(), fallbackPort, 8388), true
}

func shadowsocksUserInfo(user *url.Userinfo) (string, string, bool) {
	if user == nil {
		return "", "", false
	}
	method := strings.TrimSpace(user.Username())
	password, hasPassword := user.Password()
	password = strings.TrimSpace(password)
	if hasPassword && method != "" && password != "" {
		return method, password, true
	}
	decoded, ok := decodeBase64URL(method)
	if !ok {
		return "", "", false
	}
	method, password, ok = strings.Cut(decoded, ":")
	method = strings.TrimSpace(method)
	password = strings.TrimSpace(password)
	return method, password, ok && method != "" && password != ""
}

func decodeBase64URL(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(value)
		if err == nil {
			return string(decoded), true
		}
	}
	return "", false
}

func portWithFallback(raw string, fallbackPort int, defaultPort int) int {
	if raw != "" {
		port, err := strconv.Atoi(raw)
		if err == nil && port > 0 {
			return port
		}
	}
	if fallbackPort > 0 {
		return fallbackPort
	}
	return defaultPort
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case json.Number:
		result, err := typed.Int64()
		if err == nil {
			return int(result)
		}
	case string:
		result, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return result
		}
	}
	return 0
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
