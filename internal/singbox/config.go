package singbox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/policy"
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
	return buildConfig(tokens, virtualNodes, upstreamNodes, nil, time.Now().UTC())
}

func BuildConfigWithPolicies(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode, upstreamNodes []store.Node, policies []store.Policy) Config {
	return buildConfig(tokens, virtualNodes, upstreamNodes, policies, time.Now().UTC())
}

func buildConfig(tokens []store.TokenWithAccount, virtualNodes []store.VirtualNode, upstreamNodes []store.Node, policies []store.Policy, now time.Time) Config {
	inbounds := make([]Inbound, 0, len(virtualNodes))
	for _, node := range virtualNodes {
		if node.Status != "active" || node.ListenProtocol != "vless" {
			continue
		}
		users := vlessUsersForVirtualNode(tokens, node, virtualNodes, policies, now)
		if len(users) == 0 {
			continue
		}
		inbounds = append(inbounds, Inbound{
			Type:       "vless",
			Tag:        virtualNodeInboundTag(node),
			Listen:     "::",
			ListenPort: node.ListenPort,
			Users:      users,
		})
	}

	upstreamOutbounds, upstreamTags, upstreamTagByNodeID := buildUpstreamOutbounds(upstreamNodes)
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
	routeRules, virtualNodeOutbounds := buildVirtualNodeUpstreamRouting(tokens, inbounds, virtualNodes, upstreamNodes, upstreamTagByNodeID, policies, now)
	outbounds = append(outbounds, virtualNodeOutbounds...)
	route := map[string]any{
		"final": finalOutbound,
	}
	if len(routeRules) > 0 {
		route["rules"] = routeRules
	}

	return Config{
		Log: map[string]any{
			"level": "info",
		},
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Route:     route,
		Experimental: map[string]any{
			"v2ray_api": map[string]any{
				"listen": "0.0.0.0:9090",
				"stats": map[string]any{
					"enabled":   true,
					"inbounds":  inboundTags(inbounds),
					"outbounds": upstreamTags,
					"users":     userNamesFromInbounds(inbounds),
				},
			},
		},
	}
}

func vlessUsersForVirtualNode(tokens []store.TokenWithAccount, node store.VirtualNode, virtualNodes []store.VirtualNode, policies []store.Policy, now time.Time) []VLESSUser {
	users := make([]VLESSUser, 0, len(tokens))
	for _, token := range tokens {
		if !gatewayTokenUsable(token, now) {
			continue
		}
		if !policy.VirtualNodeAllowed(token, node, virtualNodes, policies) {
			continue
		}
		users = append(users, VLESSUser{
			Name: token.GatewayAccount.AuthUser,
			UUID: token.GatewayAccount.UUID,
			Flow: "",
		})
	}
	return users
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

func buildUpstreamOutbounds(nodes []store.Node) ([]map[string]any, []string, map[int64]string) {
	outbounds := make([]map[string]any, 0, len(nodes))
	tags := make([]string, 0, len(nodes))
	tagByNodeID := map[int64]string{}
	for _, node := range nodes {
		outbound, ok := buildNodeOutbound(node)
		if !ok {
			continue
		}
		tag := outbound["tag"].(string)
		if routeableUpstreamOutbound(outbound) {
			tags = append(tags, tag)
			tagByNodeID[node.ID] = tag
		}
		outbounds = append(outbounds, outbound)
	}
	return outbounds, tags, tagByNodeID
}

func routeableUpstreamOutbound(outbound map[string]any) bool {
	switch strings.ToLower(strings.TrimSpace(stringFromAny(outbound["type"]))) {
	case "dns", "block":
		return false
	default:
		return true
	}
}

type routeTagSelector struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

func buildVirtualNodeUpstreamRouting(tokens []store.TokenWithAccount, inbounds []Inbound, virtualNodes []store.VirtualNode, upstreamNodes []store.Node, upstreamTagByNodeID map[int64]string, policies []store.Policy, now time.Time) ([]map[string]any, []map[string]any) {
	inboundExists := map[string]bool{}
	for _, inbound := range inbounds {
		inboundExists[inbound.Tag] = true
	}

	var rules []map[string]any
	var outbounds []map[string]any
	for _, node := range virtualNodes {
		inboundTag := virtualNodeInboundTag(node)
		if !inboundExists[inboundTag] {
			continue
		}
		selector, configured := parseVirtualNodeTagSelector(node.TagSelector)
		for _, token := range tokens {
			if !gatewayTokenUsable(token, now) || !policy.VirtualNodeAllowed(token, node, virtualNodes, policies) {
				continue
			}
			policySelector, matched := policy.EffectiveTagSelector(token, policies)
			if !matched || len(policySelector.Include)+len(policySelector.Exclude) == 0 {
				continue
			}
			selectedTags := selectUpstreamTags(upstreamNodes, upstreamTagByNodeID, selector, configured, routeTagSelector{
				Include: policySelector.Include,
				Exclude: policySelector.Exclude,
			})
			targetOutbound := "block"
			if len(selectedTags) > 0 {
				targetOutbound = tokenUpstreamSelectorTag(node, token)
				outbounds = append(outbounds, map[string]any{
					"type":      "selector",
					"tag":       targetOutbound,
					"outbounds": selectedTags,
					"default":   selectedTags[0],
				})
			}
			rules = append(rules, map[string]any{
				"inbound":   []string{inboundTag},
				"auth_user": []string{token.GatewayAccount.AuthUser},
				"action":    "route",
				"outbound":  targetOutbound,
			})
		}
		if !configured {
			continue
		}
		selectedTags := selectUpstreamTags(upstreamNodes, upstreamTagByNodeID, selector, true)
		targetOutbound := "block"
		if len(selectedTags) > 0 {
			targetOutbound = virtualNodeUpstreamSelectorTag(node)
			outbounds = append(outbounds, map[string]any{
				"type":      "selector",
				"tag":       targetOutbound,
				"outbounds": selectedTags,
				"default":   selectedTags[0],
			})
		}
		rules = append(rules, map[string]any{
			"inbound":  []string{inboundTag},
			"action":   "route",
			"outbound": targetOutbound,
		})
	}
	return rules, outbounds
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
	case "hysteria2", "hy2":
		return buildHysteria2Outbound(node)
	case "tuic":
		return buildTUICOutbound(node)
	case "anytls":
		return buildAnyTLSOutbound(node)
	case "shadowtls":
		return buildShadowTLSOutbound(node)
	case "naive", "naive+https", "naive+quic":
		return buildNaiveOutbound(node)
	case "hysteria":
		return buildHysteriaOutbound(node)
	case "http", "https":
		return buildHTTPOutbound(node)
	case "socks", "socks4", "socks4a", "socks5":
		return buildSOCKSOutbound(node)
	case "ssh":
		return buildSSHOutbound(node)
	case "wireguard", "wg":
		return buildWireGuardOutbound(node)
	case "tor":
		return buildTorOutbound(node)
	case "dns":
		return buildDNSOutbound(node)
	case "direct":
		return buildDirectOutbound(node)
	case "block":
		return buildBlockOutbound(node)
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
	if strings.EqualFold(query.Get("security"), "tls") ||
		strings.EqualFold(query.Get("security"), "reality") ||
		firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name")) != "" ||
		boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) ||
		boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) ||
		strings.TrimSpace(query.Get("alpn")) != "" ||
		firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")) != "" ||
		firstNonEmpty(query.Get("fp"), query.Get("fingerprint"), query.Get("client-fingerprint")) != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
			tls["server_name"] = serverName
		}
		if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
			tls["insecure"] = true
		}
		if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
			tls["disable_sni"] = true
		}
		if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
			tls["alpn"] = alpn
		}
		if strings.EqualFold(query.Get("security"), "reality") || firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")) != "" {
			reality := map[string]any{"enabled": true}
			if publicKey := firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")); publicKey != "" {
				reality["public_key"] = publicKey
			}
			if shortID := firstNonEmpty(query.Get("sid"), query.Get("short_id"), query.Get("short-id")); shortID != "" {
				reality["short_id"] = shortID
			}
			tls["reality"] = reality
		}
		if fingerprint := firstNonEmpty(query.Get("fp"), query.Get("fingerprint"), query.Get("client-fingerprint")); fingerprint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": fingerprint,
			}
		}
		outbound["tls"] = tls
	}
	if transport := transportFromQuery(query); transport != nil {
		outbound["transport"] = transport
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
	if strings.EqualFold(query.Get("security"), "tls") ||
		strings.EqualFold(query.Get("security"), "reality") ||
		firstNonEmpty(query.Get("sni"), query.Get("peer"), query.Get("servername"), query.Get("server_name")) != "" ||
		boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) ||
		boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) ||
		strings.TrimSpace(query.Get("alpn")) != "" ||
		firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")) != "" ||
		firstNonEmpty(query.Get("fp"), query.Get("fingerprint"), query.Get("client-fingerprint")) != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(query.Get("sni"), query.Get("peer"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
			tls["server_name"] = serverName
		}
		if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
			tls["insecure"] = true
		}
		if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
			tls["disable_sni"] = true
		}
		if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
			tls["alpn"] = alpn
		}
		if strings.EqualFold(query.Get("security"), "reality") || firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")) != "" {
			reality := map[string]any{"enabled": true}
			if publicKey := firstNonEmpty(query.Get("pbk"), query.Get("public_key"), query.Get("public-key")); publicKey != "" {
				reality["public_key"] = publicKey
			}
			if shortID := firstNonEmpty(query.Get("sid"), query.Get("short_id"), query.Get("short-id")); shortID != "" {
				reality["short_id"] = shortID
			}
			tls["reality"] = reality
		}
		if fingerprint := firstNonEmpty(query.Get("fp"), query.Get("fingerprint"), query.Get("client-fingerprint")); fingerprint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": fingerprint,
			}
		}
		outbound["tls"] = tls
	}
	if transport := transportFromQuery(query); transport != nil {
		outbound["transport"] = transport
	}
	return outbound, true
}

func transportFromQuery(query url.Values) map[string]any {
	transportType := strings.ToLower(firstNonEmpty(query.Get("type"), query.Get("network"), query.Get("net")))
	switch transportType {
	case "ws", "websocket":
		transport := map[string]any{"type": "ws"}
		if path := firstNonEmpty(query.Get("path"), query.Get("ws_path"), query.Get("ws-path")); path != "" {
			transport["path"] = path
		}
		if host := firstNonEmpty(query.Get("host"), query.Get("ws_host"), query.Get("ws-host")); host != "" {
			transport["headers"] = map[string]any{"Host": host}
		}
		if maxEarlyData := intQuery(firstNonEmpty(query.Get("max_early_data"), query.Get("max-early-data"))); maxEarlyData > 0 {
			transport["max_early_data"] = maxEarlyData
		}
		if earlyDataHeaderName := firstNonEmpty(query.Get("early_data_header_name"), query.Get("early-data-header-name")); earlyDataHeaderName != "" {
			transport["early_data_header_name"] = earlyDataHeaderName
		}
		return transport
	case "grpc":
		transport := map[string]any{"type": "grpc"}
		if serviceName := firstNonEmpty(query.Get("service_name"), query.Get("serviceName"), query.Get("grpc_service_name"), query.Get("grpc-service-name")); serviceName != "" {
			transport["service_name"] = serviceName
		}
		if idleTimeout := firstNonEmpty(query.Get("idle_timeout"), query.Get("idle-timeout"), query.Get("grpc_idle_timeout"), query.Get("grpc-idle-timeout")); idleTimeout != "" {
			transport["idle_timeout"] = idleTimeout
		}
		if pingTimeout := firstNonEmpty(query.Get("ping_timeout"), query.Get("ping-timeout"), query.Get("grpc_ping_timeout"), query.Get("grpc-ping-timeout")); pingTimeout != "" {
			transport["ping_timeout"] = pingTimeout
		}
		if boolQuery(firstNonEmpty(query.Get("permit_without_stream"), query.Get("permit-without-stream"))) {
			transport["permit_without_stream"] = true
		}
		if boolQuery(firstNonEmpty(query.Get("multi_mode"), query.Get("multi-mode"), query.Get("grpc_multi_mode"), query.Get("grpc-multi-mode"))) {
			transport["multi_mode"] = true
		}
		return transport
	case "quic":
		return map[string]any{"type": "quic"}
	case "http", "h2":
		transport := map[string]any{"type": "http"}
		if hosts := splitCSV(firstNonEmpty(query.Get("host"), query.Get("http_host"), query.Get("http-host"))); len(hosts) > 0 {
			transport["host"] = hosts
		}
		if path := firstNonEmpty(query.Get("path"), query.Get("http_path"), query.Get("http-path")); path != "" {
			transport["path"] = path
		}
		if method := strings.TrimSpace(query.Get("method")); method != "" {
			transport["method"] = method
		}
		if idleTimeout := firstNonEmpty(query.Get("idle_timeout"), query.Get("idle-timeout")); idleTimeout != "" {
			transport["idle_timeout"] = idleTimeout
		}
		if pingTimeout := firstNonEmpty(query.Get("ping_timeout"), query.Get("ping-timeout")); pingTimeout != "" {
			transport["ping_timeout"] = pingTimeout
		}
		return transport
	case "httpupgrade", "http-upgrade", "http_upgrade":
		transport := map[string]any{"type": "httpupgrade"}
		if host := firstNonEmpty(query.Get("host"), query.Get("httpupgrade_host"), query.Get("httpupgrade-host"), query.Get("http_upgrade_host"), query.Get("http-upgrade-host")); host != "" {
			transport["host"] = host
		}
		if path := firstNonEmpty(query.Get("path"), query.Get("httpupgrade_path"), query.Get("httpupgrade-path"), query.Get("http_upgrade_path"), query.Get("http-upgrade-path")); path != "" {
			transport["path"] = path
		}
		return transport
	default:
		return nil
	}
}

func buildShadowsocksOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "ss" {
		return nil, false
	}
	method, password, server, port, ok := parseShadowsocksURI(node.URI, node.ServerPort)
	if !ok {
		return nil, false
	}
	outbound := map[string]any{
		"type":        "shadowsocks",
		"tag":         upstreamTag(node),
		"server":      server,
		"server_port": port,
		"method":      method,
		"password":    password,
	}
	if parsed, err := url.Parse(strings.TrimSpace(node.URI)); err == nil {
		query := parsed.Query()
		if plugin := strings.TrimSpace(query.Get("plugin")); plugin != "" {
			outbound["plugin"] = plugin
		}
		if pluginOpts := firstNonEmpty(query.Get("plugin_opts"), query.Get("plugin-opts"), query.Get("plugin_options"), query.Get("plugin-options")); pluginOpts != "" {
			outbound["plugin_opts"] = pluginOpts
		}
		if network := strings.TrimSpace(query.Get("network")); network != "" {
			outbound["network"] = network
		}
	}
	return outbound, true
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
	vmessInsecure := boolFromAny(doc["allowInsecure"]) ||
		boolFromAny(doc["allowinsecure"]) ||
		boolFromAny(doc["insecure"]) ||
		boolFromAny(doc["skip-cert-verify"]) ||
		boolFromAny(doc["skip_cert_verify"])
	vmessDisableSNI := boolFromAny(doc["disable_sni"]) || boolFromAny(doc["disable-sni"])
	vmessALPN := splitCSV(stringFromAny(doc["alpn"]))
	vmessFingerprint := firstNonEmpty(
		stringFromAny(doc["fp"]),
		stringFromAny(doc["fingerprint"]),
		stringFromAny(doc["client-fingerprint"]),
		stringFromAny(doc["client_fingerprint"]),
		stringFromAny(doc["clientFingerprint"]),
	)
	if strings.EqualFold(stringFromAny(doc["tls"]), "tls") ||
		strings.TrimSpace(stringFromAny(doc["sni"])) != "" ||
		vmessInsecure ||
		vmessDisableSNI ||
		len(vmessALPN) > 0 ||
		vmessFingerprint != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(stringFromAny(doc["sni"]), stringFromAny(doc["host"]), server); serverName != "" {
			tls["server_name"] = serverName
		}
		if vmessInsecure {
			tls["insecure"] = true
		}
		if vmessDisableSNI {
			tls["disable_sni"] = true
		}
		if len(vmessALPN) > 0 {
			tls["alpn"] = vmessALPN
		}
		if vmessFingerprint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": vmessFingerprint,
			}
		}
		outbound["tls"] = tls
	}
	if transport := vmessTransportFromDoc(doc); transport != nil {
		outbound["transport"] = transport
	}
	return outbound, true
}

func vmessTransportFromDoc(doc map[string]any) map[string]any {
	transportType := strings.ToLower(strings.TrimSpace(stringFromAny(doc["net"])))
	switch transportType {
	case "ws", "websocket":
		transport := map[string]any{"type": "ws"}
		if path := strings.TrimSpace(stringFromAny(doc["path"])); path != "" {
			transport["path"] = path
		}
		if host := strings.TrimSpace(stringFromAny(doc["host"])); host != "" {
			transport["headers"] = map[string]any{"Host": host}
		}
		return transport
	case "grpc":
		transport := map[string]any{"type": "grpc"}
		if serviceName := firstNonEmpty(
			stringFromAny(doc["path"]),
			stringFromAny(doc["service_name"]),
			stringFromAny(doc["serviceName"]),
			stringFromAny(doc["grpc_service_name"]),
			stringFromAny(doc["grpc-service-name"]),
		); serviceName != "" {
			transport["service_name"] = serviceName
		}
		if boolFromAny(doc["multi_mode"]) || boolFromAny(doc["multi-mode"]) || boolFromAny(doc["grpc_multi_mode"]) || boolFromAny(doc["grpc-multi-mode"]) {
			transport["multi_mode"] = true
		}
		return transport
	case "http", "h2":
		return vmessHTTPTransportFromDoc(doc)
	case "tcp":
		headerType := strings.ToLower(strings.TrimSpace(stringFromAny(doc["type"])))
		if headerType == "http" {
			return vmessHTTPTransportFromDoc(doc)
		}
		return nil
	default:
		return nil
	}
}

func vmessHTTPTransportFromDoc(doc map[string]any) map[string]any {
	transport := map[string]any{"type": "http"}
	if hosts := splitCSV(stringFromAny(doc["host"])); len(hosts) > 0 {
		transport["host"] = hosts
	}
	if path := strings.TrimSpace(stringFromAny(doc["path"])); path != "" {
		transport["path"] = path
	}
	return transport
}

func buildHysteria2Outbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || (node.Protocol != "hysteria2" && node.Protocol != "hy2") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || (parsed.Scheme != "hysteria2" && parsed.Scheme != "hy2") || parsed.Hostname() == "" {
		return nil, false
	}
	query := parsed.Query()
	password := firstNonEmpty(
		hysteria2Password(parsed.User),
		query.Get("password"),
		query.Get("auth"),
		query.Get("auth_str"),
		query.Get("auth-str"),
		query.Get("token"),
	)
	if password == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "hysteria2",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
		"password":    password,
	}

	if obfsType := strings.TrimSpace(query.Get("obfs")); obfsType != "" {
		obfs := map[string]any{"type": obfsType}
		if obfsPassword := strings.TrimSpace(query.Get("obfs-password")); obfsPassword != "" {
			obfs["password"] = obfsPassword
		}
		outbound["obfs"] = obfs
	}
	if upMbps := intQuery(firstNonEmpty(query.Get("up_mbps"), query.Get("up-mbps"), query.Get("upmbps"))); upMbps > 0 {
		outbound["up_mbps"] = upMbps
	}
	if downMbps := intQuery(firstNonEmpty(query.Get("down_mbps"), query.Get("down-mbps"), query.Get("downmbps"))); downMbps > 0 {
		outbound["down_mbps"] = downMbps
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
		tls["insecure"] = true
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		tls["disable_sni"] = true
	}
	if fingerprint := firstNonEmpty(query.Get("pinSHA256"), query.Get("pin-sha256"), query.Get("fingerprint")); fingerprint != "" {
		tls["certificate_public_key_sha256"] = []string{fingerprint}
	}
	if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
		tls["alpn"] = alpn
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildTUICOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "tuic" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "tuic" || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	uuid, password := tuicCredentials(parsed.User, query.Get("uuid"), query.Get("password"))
	if uuid == "" || password == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "tuic",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
		"uuid":        uuid,
		"password":    password,
	}

	if congestionControl := firstNonEmpty(query.Get("congestion_control"), query.Get("congestion-controller")); congestionControl != "" {
		outbound["congestion_control"] = congestionControl
	}
	if boolQuery(firstNonEmpty(query.Get("udp_over_stream"), query.Get("udp-over-stream"))) {
		outbound["udp_over_stream"] = true
	} else if udpRelayMode := firstNonEmpty(query.Get("udp_relay_mode"), query.Get("udp-relay-mode")); udpRelayMode != "" {
		outbound["udp_relay_mode"] = udpRelayMode
	}
	if boolQuery(firstNonEmpty(query.Get("zero_rtt_handshake"), query.Get("zero-rtt-handshake"), query.Get("reduce-rtt"))) {
		outbound["zero_rtt_handshake"] = true
	}
	if heartbeat := firstNonEmpty(query.Get("heartbeat"), query.Get("heartbeat-interval")); heartbeat != "" {
		outbound["heartbeat"] = heartbeat
	}
	if network := strings.TrimSpace(query.Get("network")); network != "" {
		outbound["network"] = network
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
		tls["insecure"] = true
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		tls["disable_sni"] = true
	}
	if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
		tls["alpn"] = alpn
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildAnyTLSOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "anytls" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "anytls" || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	password := anyTLSPassword(parsed.User, query.Get("password"))
	if password == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "anytls",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
		"password":    password,
	}
	if checkInterval := firstNonEmpty(query.Get("idle_session_check_interval"), query.Get("idle-session-check-interval")); checkInterval != "" {
		outbound["idle_session_check_interval"] = checkInterval
	}
	if timeout := firstNonEmpty(query.Get("idle_session_timeout"), query.Get("idle-session-timeout")); timeout != "" {
		outbound["idle_session_timeout"] = timeout
	}
	if minIdleSession := intQuery(firstNonEmpty(query.Get("min_idle_session"), query.Get("min-idle-session"))); minIdleSession > 0 {
		outbound["min_idle_session"] = minIdleSession
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
		tls["insecure"] = true
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		tls["disable_sni"] = true
	}
	if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
		tls["alpn"] = alpn
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildShadowTLSOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "shadowtls" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "shadowtls" || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	version := intQuery(query.Get("version"))
	if version <= 0 {
		version = 1
	}
	password := anyTLSPassword(parsed.User, query.Get("password"))
	if version >= 2 && password == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "shadowtls",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
		"version":     version,
	}
	if password != "" {
		outbound["password"] = password
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("peer"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
		tls["insecure"] = true
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		tls["disable_sni"] = true
	}
	if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
		tls["alpn"] = alpn
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildNaiveOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || !strings.HasPrefix(node.Protocol, "naive") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || !strings.HasPrefix(parsed.Scheme, "naive") || parsed.Hostname() == "" {
		return nil, false
	}
	username, password := userPassword(parsed.User)
	if username == "" || password == "" {
		return nil, false
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type":        "naive",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
		"username":    username,
		"password":    password,
	}
	if concurrency := intQuery(firstNonEmpty(query.Get("insecure_concurrency"), query.Get("insecure-concurrency"))); concurrency > 0 {
		outbound["insecure_concurrency"] = concurrency
	}
	if boolQuery(firstNonEmpty(query.Get("udp_over_tcp"), query.Get("udp-over-tcp"))) {
		outbound["udp_over_tcp"] = true
	}
	if parsed.Scheme == "naive+quic" || boolQuery(query.Get("quic")) {
		outbound["quic"] = true
	}
	if congestionControl := firstNonEmpty(query.Get("quic_congestion_control"), query.Get("quic-congestion-control")); congestionControl != "" {
		outbound["quic_congestion_control"] = congestionControl
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if certificate := firstNonEmpty(query.Get("certificate"), query.Get("cert")); certificate != "" {
		tls["certificate"] = certificate
	}
	if certificatePath := firstNonEmpty(query.Get("certificate_path"), query.Get("certificate-path"), query.Get("cert_path"), query.Get("cert-path")); certificatePath != "" {
		tls["certificate_path"] = certificatePath
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildHysteriaOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "hysteria" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "hysteria" || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	auth := firstNonEmpty(query.Get("auth"), query.Get("auth_base64"), query.Get("auth-base64"))
	authStr := firstNonEmpty(query.Get("auth_str"), query.Get("auth-str"), query.Get("password"), query.Get("token"), parsed.User.Username())
	if auth == "" && authStr == "" {
		return nil, false
	}

	outbound := map[string]any{
		"type":        "hysteria",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 443),
	}
	if auth != "" {
		outbound["auth"] = auth
	}
	if authStr != "" {
		outbound["auth_str"] = authStr
	}
	if up := firstNonEmpty(query.Get("up"), query.Get("up_speed"), query.Get("up-speed")); up != "" {
		outbound["up"] = up
	}
	if upMbps := intQuery(firstNonEmpty(query.Get("up_mbps"), query.Get("up-mbps"), query.Get("upmbps"))); upMbps > 0 {
		outbound["up_mbps"] = upMbps
	}
	if down := firstNonEmpty(query.Get("down"), query.Get("down_speed"), query.Get("down-speed")); down != "" {
		outbound["down"] = down
	}
	if downMbps := intQuery(firstNonEmpty(query.Get("down_mbps"), query.Get("down-mbps"), query.Get("downmbps"))); downMbps > 0 {
		outbound["down_mbps"] = downMbps
	}
	if obfs := strings.TrimSpace(query.Get("obfs")); obfs != "" {
		outbound["obfs"] = obfs
	}
	if recvWindowConn := intQuery(firstNonEmpty(query.Get("recv_window_conn"), query.Get("recv-window-conn"))); recvWindowConn > 0 {
		outbound["recv_window_conn"] = recvWindowConn
	}
	if recvWindow := intQuery(firstNonEmpty(query.Get("recv_window"), query.Get("recv-window"))); recvWindow > 0 {
		outbound["recv_window"] = recvWindow
	}
	if boolQuery(firstNonEmpty(query.Get("disable_mtu_discovery"), query.Get("disable-mtu-discovery"))) {
		outbound["disable_mtu_discovery"] = true
	}
	if network := firstNonEmpty(query.Get("network"), query.Get("protocol")); network != "" {
		outbound["network"] = network
	}

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("peer"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
		tls["insecure"] = true
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		tls["disable_sni"] = true
	}
	if alpn := splitCSV(query.Get("alpn")); len(alpn) > 0 {
		tls["alpn"] = alpn
	}
	outbound["tls"] = tls

	return outbound, true
}

func buildHTTPOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || (node.Protocol != "http" && node.Protocol != "https") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return nil, false
	}

	defaultPort := 80
	if parsed.Scheme == "https" {
		defaultPort = 443
	}
	outbound := map[string]any{
		"type":        "http",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, defaultPort),
	}

	username, password := userPassword(parsed.User)
	if username != "" {
		outbound["username"] = username
	}
	if password != "" {
		outbound["password"] = password
	}

	query := parsed.Query()
	if path := firstNonEmpty(query.Get("path"), parsed.EscapedPath()); path != "" && path != "/" {
		outbound["path"] = path
	}

	if parsed.Scheme == "https" || strings.EqualFold(query.Get("security"), "tls") || boolQuery(query.Get("tls")) || firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name")) != "" {
		tls := map[string]any{"enabled": true}
		if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name"), parsed.Hostname()); serverName != "" {
			tls["server_name"] = serverName
		}
		if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"))) {
			tls["insecure"] = true
		}
		outbound["tls"] = tls
	}

	return outbound, true
}

func buildSOCKSOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || !strings.HasPrefix(node.Protocol, "socks") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || !strings.HasPrefix(parsed.Scheme, "socks") || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type":        "socks",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 1080),
	}
	if version := socksVersion(parsed.Scheme, query.Get("version")); version != "" {
		outbound["version"] = version
	}
	username, password := userPassword(parsed.User)
	if username != "" {
		outbound["username"] = username
	}
	if password != "" {
		outbound["password"] = password
	}
	if network := firstNonEmpty(query.Get("network"), query.Get("protocol")); network != "" {
		outbound["network"] = network
	}
	if boolQuery(firstNonEmpty(query.Get("udp_over_tcp"), query.Get("udp-over-tcp"), query.Get("uot"))) {
		outbound["udp_over_tcp"] = true
	}

	return outbound, true
}

func buildSSHOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "ssh" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "ssh" || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type":        "ssh",
		"tag":         upstreamTag(node),
		"server":      parsed.Hostname(),
		"server_port": portWithFallback(parsed.Port(), node.ServerPort, 22),
	}
	if user := firstNonEmpty(parsed.User.Username(), query.Get("user"), query.Get("username")); user != "" {
		outbound["user"] = user
	}
	_, userPasswordValue := userPassword(parsed.User)
	if password := firstNonEmpty(userPasswordValue, query.Get("password")); password != "" {
		outbound["password"] = password
	}
	if privateKey := firstNonEmpty(query.Get("private_key"), query.Get("private-key")); privateKey != "" {
		outbound["private_key"] = privateKey
	}
	if privateKeyPath := firstNonEmpty(query.Get("private_key_path"), query.Get("private-key-path")); privateKeyPath != "" {
		outbound["private_key_path"] = privateKeyPath
	}
	if privateKeyPassphrase := firstNonEmpty(query.Get("private_key_passphrase"), query.Get("private-key-passphrase")); privateKeyPassphrase != "" {
		outbound["private_key_passphrase"] = privateKeyPassphrase
	}
	if hostKey := splitCSV(firstNonEmpty(query.Get("host_key"), query.Get("host-key"))); len(hostKey) > 0 {
		outbound["host_key"] = hostKey
	}
	if hostKeyAlgorithms := splitCSV(firstNonEmpty(query.Get("host_key_algorithms"), query.Get("host-key-algorithms"))); len(hostKeyAlgorithms) > 0 {
		outbound["host_key_algorithms"] = hostKeyAlgorithms
	}
	if clientVersion := firstNonEmpty(query.Get("client_version"), query.Get("client-version")); clientVersion != "" {
		outbound["client_version"] = clientVersion
	}
	if cipher := splitCSV(query.Get("cipher")); len(cipher) > 0 {
		outbound["cipher"] = cipher
	}
	if mac := splitCSV(query.Get("mac")); len(mac) > 0 {
		outbound["mac"] = mac
	}
	if kexAlgorithm := splitCSV(firstNonEmpty(query.Get("kex_algorithm"), query.Get("kex-algorithm"))); len(kexAlgorithm) > 0 {
		outbound["kex_algorithm"] = kexAlgorithm
	}

	return outbound, true
}

func buildWireGuardOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || (node.Protocol != "wireguard" && node.Protocol != "wg") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || (parsed.Scheme != "wireguard" && parsed.Scheme != "wg") || parsed.Hostname() == "" {
		return nil, false
	}

	query := parsed.Query()
	privateKey := firstNonEmpty(query.Get("private_key"), query.Get("private-key"), parsed.User.Username())
	peerPublicKey := firstNonEmpty(query.Get("peer_public_key"), query.Get("peer-public-key"), query.Get("public_key"), query.Get("public-key"))
	localAddress := splitCSV(firstNonEmpty(query.Get("local_address"), query.Get("local-address"), query.Get("address")))
	if privateKey == "" || peerPublicKey == "" || len(localAddress) == 0 {
		return nil, false
	}

	serverPort := portWithFallback(parsed.Port(), node.ServerPort, 51820)
	outbound := map[string]any{
		"type":            "wireguard",
		"tag":             upstreamTag(node),
		"server":          parsed.Hostname(),
		"server_port":     serverPort,
		"local_address":   localAddress,
		"private_key":     privateKey,
		"peer_public_key": peerPublicKey,
	}
	if systemInterface := boolQuery(firstNonEmpty(query.Get("system_interface"), query.Get("system-interface"), query.Get("system"))); systemInterface {
		outbound["system_interface"] = true
	}
	if interfaceName := firstNonEmpty(query.Get("interface_name"), query.Get("interface-name"), query.Get("name")); interfaceName != "" {
		outbound["interface_name"] = interfaceName
	}
	preSharedKey := firstNonEmpty(query.Get("pre_shared_key"), query.Get("pre-shared-key"), query.Get("preshared_key"), query.Get("psk"))
	if preSharedKey != "" {
		outbound["pre_shared_key"] = preSharedKey
	}
	if reserved, ok := byteListQuery(query.Get("reserved")); ok {
		outbound["reserved"] = reserved
	}
	if workers := intQuery(query.Get("workers")); workers > 0 {
		outbound["workers"] = workers
	}
	if mtu := intQuery(query.Get("mtu")); mtu > 0 {
		outbound["mtu"] = mtu
	}
	if network := firstNonEmpty(query.Get("network"), query.Get("protocol")); network != "" {
		outbound["network"] = network
	}

	if allowedIPs := splitCSV(firstNonEmpty(query.Get("allowed_ips"), query.Get("allowed-ips"), query.Get("peer_allowed_ips"), query.Get("peer-allowed-ips"))); len(allowedIPs) > 0 {
		peer := map[string]any{
			"server":      parsed.Hostname(),
			"server_port": serverPort,
			"public_key":  peerPublicKey,
			"allowed_ips": allowedIPs,
		}
		if preSharedKey != "" {
			peer["pre_shared_key"] = preSharedKey
		}
		if reserved, ok := byteListQuery(firstNonEmpty(query.Get("peer_reserved"), query.Get("peer-reserved"), query.Get("reserved"))); ok {
			peer["reserved"] = reserved
		}
		outbound["peers"] = []map[string]any{peer}
	}

	return outbound, true
}

func buildTorOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "tor" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "tor" {
		return nil, false
	}

	query := parsed.Query()
	outbound := map[string]any{
		"type": "tor",
		"tag":  upstreamTag(node),
	}
	if executablePath := firstNonEmpty(query.Get("executable_path"), query.Get("executable-path")); executablePath != "" {
		outbound["executable_path"] = executablePath
	}
	if extraArgs := splitCSV(firstNonEmpty(query.Get("extra_args"), query.Get("extra-args"), query.Get("args"))); len(extraArgs) > 0 {
		outbound["extra_args"] = extraArgs
	}
	if dataDirectory := firstNonEmpty(query.Get("data_directory"), query.Get("data-directory")); dataDirectory != "" {
		outbound["data_directory"] = dataDirectory
	}
	if torrc := torrcQuery(query); len(torrc) > 0 {
		outbound["torrc"] = torrc
	}

	return outbound, true
}

func buildDNSOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "dns" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "dns" {
		return nil, false
	}
	return map[string]any{
		"type": "dns",
		"tag":  upstreamTag(node),
	}, true
}

func buildDirectOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "direct" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "direct" {
		return nil, false
	}
	return map[string]any{
		"type": "direct",
		"tag":  upstreamTag(node),
	}, true
}

func buildBlockOutbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || node.Protocol != "block" {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || parsed.Scheme != "block" {
		return nil, false
	}
	return map[string]any{
		"type": "block",
		"tag":  upstreamTag(node),
	}, true
}

func parseVMessURI(rawURI string) (map[string]any, bool) {
	rawURI = strings.TrimSpace(rawURI)
	if !strings.HasPrefix(rawURI, "vmess://") {
		return nil, false
	}
	if doc, ok := parseVMessUserinfoURI(rawURI); ok {
		return doc, true
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

func parseVMessUserinfoURI(rawURI string) (map[string]any, bool) {
	parsed, err := url.Parse(rawURI)
	if err != nil || parsed.Scheme != "vmess" || parsed.Hostname() == "" || parsed.User == nil {
		return nil, false
	}
	uuid := strings.TrimSpace(parsed.User.Username())
	if uuid == "" {
		return nil, false
	}
	query := parsed.Query()
	doc := map[string]any{
		"v":    "2",
		"ps":   strings.TrimSpace(parsed.Fragment),
		"add":  parsed.Hostname(),
		"port": firstNonEmpty(parsed.Port(), "443"),
		"id":   uuid,
		"aid":  firstNonEmpty(query.Get("alterId"), query.Get("alterid"), query.Get("alter-id"), query.Get("aid"), "0"),
		"scy":  firstNonEmpty(query.Get("encryption"), query.Get("scy"), query.Get("cipher"), "auto"),
		"net":  firstNonEmpty(query.Get("type"), query.Get("net"), query.Get("network"), "tcp"),
		"type": firstNonEmpty(query.Get("headerType"), query.Get("header-type"), query.Get("header_type")),
		"host": firstNonEmpty(query.Get("host"), query.Get("authority")),
		"path": query.Get("path"),
		"sni":  firstNonEmpty(query.Get("sni"), query.Get("servername"), query.Get("server_name")),
		"alpn": query.Get("alpn"),
	}
	if multiMode := firstNonEmpty(query.Get("multi_mode"), query.Get("multi-mode"), query.Get("grpc_multi_mode"), query.Get("grpc-multi-mode")); multiMode != "" {
		doc["multi_mode"] = multiMode
	}
	if strings.EqualFold(query.Get("security"), "tls") ||
		strings.EqualFold(query.Get("tls"), "tls") ||
		boolQuery(query.Get("tls")) {
		doc["tls"] = "tls"
	}
	if boolQuery(firstNonEmpty(query.Get("insecure"), query.Get("skip-cert-verify"), query.Get("skip_cert_verify"), query.Get("allowInsecure"), query.Get("allow_insecure"))) {
		doc["allowInsecure"] = "1"
	}
	if boolQuery(firstNonEmpty(query.Get("disable_sni"), query.Get("disable-sni"))) {
		doc["disable_sni"] = "1"
	}
	if fingerprint := firstNonEmpty(query.Get("fp"), query.Get("fingerprint"), query.Get("client-fingerprint"), query.Get("client_fingerprint"), query.Get("clientFingerprint")); fingerprint != "" {
		doc["fp"] = fingerprint
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

func boolFromAny(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return boolQuery(stringFromAny(value))
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

func hysteria2Password(user *url.Userinfo) string {
	if user == nil {
		return ""
	}
	password := user.Username()
	if value, ok := user.Password(); ok {
		password += ":" + value
	}
	return strings.TrimSpace(password)
}

func tuicCredentials(user *url.Userinfo, queryUUID string, queryPassword string) (string, string) {
	uuid := strings.TrimSpace(queryUUID)
	password := strings.TrimSpace(queryPassword)
	if user == nil {
		return uuid, password
	}
	if username := strings.TrimSpace(user.Username()); username != "" {
		uuid = username
	}
	if value, ok := user.Password(); ok && strings.TrimSpace(value) != "" {
		password = strings.TrimSpace(value)
	}
	return uuid, password
}

func anyTLSPassword(user *url.Userinfo, queryPassword string) string {
	if user == nil {
		return strings.TrimSpace(queryPassword)
	}
	password := user.Username()
	if value, ok := user.Password(); ok {
		password += ":" + value
	}
	if strings.TrimSpace(password) == "" {
		return strings.TrimSpace(queryPassword)
	}
	return strings.TrimSpace(password)
}

func userPassword(user *url.Userinfo) (string, string) {
	if user == nil {
		return "", ""
	}
	username := strings.TrimSpace(user.Username())
	password, ok := user.Password()
	password = strings.TrimSpace(password)
	if !ok {
		return username, ""
	}
	return username, password
}

func socksVersion(scheme string, queryVersion string) string {
	version := strings.TrimSpace(queryVersion)
	if version == "" {
		version = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(scheme)), "socks")
	}
	if version == "" {
		version = "5"
	}
	switch strings.ToLower(version) {
	case "4", "4a", "5":
		return strings.ToLower(version)
	default:
		return ""
	}
}

func byteListQuery(value string) ([]int, bool) {
	parts := splitCSV(value)
	if len(parts) == 0 {
		return nil, false
	}
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 || value > 255 {
			return nil, false
		}
		result = append(result, value)
	}
	return result, true
}

func torrcQuery(query url.Values) map[string]any {
	torrc := map[string]any{}
	for key, values := range query {
		option := ""
		switch {
		case strings.HasPrefix(key, "torrc."):
			option = strings.TrimPrefix(key, "torrc.")
		case strings.HasPrefix(key, "torrc_"):
			option = strings.TrimPrefix(key, "torrc_")
		}
		option = strings.TrimSpace(option)
		if option == "" || len(values) == 0 {
			continue
		}
		value := strings.TrimSpace(values[len(values)-1])
		if value == "" {
			continue
		}
		torrc[option] = scalarQueryValue(value)
	}
	return torrc
}

func scalarQueryValue(value string) any {
	if integer, err := strconv.Atoi(value); err == nil {
		return integer
	}
	switch strings.ToLower(value) {
	case "true", "yes", "on":
		return true
	case "false", "no", "off":
		return false
	default:
		return value
	}
}

func boolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func splitCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func intQuery(value string) int {
	result, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return result
}

func upstreamTag(node store.Node) string {
	return fmt.Sprintf("up_%d", node.ID)
}

func virtualNodeInboundTag(node store.VirtualNode) string {
	return "vn-" + node.Name
}

func virtualNodeUpstreamSelectorTag(node store.VirtualNode) string {
	return "vn-" + node.Name + "-upstreams"
}

func tokenUpstreamSelectorTag(node store.VirtualNode, token store.TokenWithAccount) string {
	if token.ID > 0 {
		return fmt.Sprintf("vn-%s-token-%d-upstreams", node.Name, token.ID)
	}
	return "vn-" + node.Name + "-user-" + token.GatewayAccount.AuthUser + "-upstreams"
}

func parseVirtualNodeTagSelector(raw string) (routeTagSelector, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return routeTagSelector{}, false
	}
	var document struct {
		Include     []string `json:"include"`
		Exclude     []string `json:"exclude"`
		IncludeTags []string `json:"include_tags"`
		ExcludeTags []string `json:"exclude_tags"`
	}
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		return routeTagSelector{}, false
	}
	selector := routeTagSelector{
		Include: cleanTagSelectorValues(append(document.Include, document.IncludeTags...)),
		Exclude: cleanTagSelectorValues(append(document.Exclude, document.ExcludeTags...)),
	}
	return selector, len(selector.Include) > 0 || len(selector.Exclude) > 0
}

func selectUpstreamTags(nodes []store.Node, upstreamTagByNodeID map[int64]string, selector routeTagSelector, selectorConfigured bool, extraSelectors ...routeTagSelector) []string {
	var selected []string
	for _, node := range nodes {
		upstreamTag, ok := upstreamTagByNodeID[node.ID]
		if !ok || !nodeMatchesTagSelectors(node, selector, selectorConfigured, extraSelectors...) {
			continue
		}
		selected = append(selected, upstreamTag)
	}
	return selected
}

func nodeMatchesTagSelectors(node store.Node, selector routeTagSelector, selectorConfigured bool, extraSelectors ...routeTagSelector) bool {
	if selectorConfigured && !nodeMatchesTagSelector(node, selector) {
		return false
	}
	for _, extra := range extraSelectors {
		if !nodeMatchesTagSelector(node, extra) {
			return false
		}
	}
	return true
}

func nodeMatchesTagSelector(node store.Node, selector routeTagSelector) bool {
	nodeTags := tagSet(node.Tags)
	return (len(selector.Include) == 0 || containsAnyTag(nodeTags, selector.Include)) && !containsAnyTag(nodeTags, selector.Exclude)
}

func cleanTagSelectorValues(values []string) []string {
	seen := map[string]bool{}
	var cleaned []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func tagSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result[strings.ToLower(value)] = true
		}
	}
	return result
}

func containsAnyTag(set map[string]bool, values []string) bool {
	for _, value := range values {
		if set[strings.ToLower(strings.TrimSpace(value))] {
			return true
		}
	}
	return false
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

func userNamesFromInbounds(inbounds []Inbound) []string {
	var names []string
	seen := map[string]bool{}
	for _, inbound := range inbounds {
		for _, name := range userNames(inbound.Users) {
			if seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}
