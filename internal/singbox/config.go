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

func buildHysteria2Outbound(node store.Node) (map[string]any, bool) {
	if node.Status != "active" || (node.Protocol != "hysteria2" && node.Protocol != "hy2") {
		return nil, false
	}
	parsed, err := url.Parse(strings.TrimSpace(node.URI))
	if err != nil || (parsed.Scheme != "hysteria2" && parsed.Scheme != "hy2") || parsed.Hostname() == "" {
		return nil, false
	}
	password := hysteria2Password(parsed.User)
	if password == "" {
		return nil, false
	}

	query := parsed.Query()
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

	tls := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(query.Get("sni"), parsed.Hostname()); serverName != "" {
		tls["server_name"] = serverName
	}
	if boolQuery(query.Get("insecure")) {
		tls["insecure"] = true
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
	authStr := firstNonEmpty(query.Get("auth_str"), query.Get("auth-str"), query.Get("password"), parsed.User.Username())
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
