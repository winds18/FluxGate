package upstream

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func ClashYAMLURIList(content string) string {
	var proxies []map[string]string
	var current map[string]string
	inProxies := false

	for _, rawLine := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " \t"))
		if !inProxies {
			if trimmed == "proxies:" {
				inProxies = true
			}
			continue
		}
		if indent == 0 && !strings.HasPrefix(trimmed, "- ") {
			break
		}
		if strings.HasPrefix(trimmed, "- ") {
			if len(current) > 0 {
				proxies = append(proxies, current)
			}
			current = map[string]string{}
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if strings.HasPrefix(rest, "{") && strings.HasSuffix(rest, "}") {
				mergeProxyFields(current, parseInlineMap(rest))
				continue
			}
			key, value, ok := parseYAMLField(rest)
			if ok {
				current[strings.ToLower(key)] = value
			}
			continue
		}
		if current == nil {
			continue
		}
		key, value, ok := parseYAMLField(trimmed)
		if ok {
			current[strings.ToLower(key)] = value
		}
	}
	if len(current) > 0 {
		proxies = append(proxies, current)
	}

	var uris []string
	for _, proxy := range proxies {
		if uri := clashProxyURI(proxy); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func clashProxyURI(proxy map[string]string) string {
	proxyType := strings.ToLower(firstMapValue(proxy, "type"))
	switch proxyType {
	case "ss", "shadowsocks":
		return clashShadowsocksURI(proxy)
	case "trojan":
		return clashTrojanURI(proxy)
	case "vless":
		return clashVLESSURI(proxy)
	case "vmess":
		return clashVMessURI(proxy)
	case "hysteria2", "hy2":
		return clashHysteria2URI(proxy)
	case "tuic":
		return clashTUICURI(proxy)
	case "hysteria":
		return clashHysteriaURI(proxy)
	case "http", "https":
		return clashHTTPURI(proxy)
	case "socks", "socks4", "socks4a", "socks5":
		return clashSOCKSURI(proxy)
	case "anytls":
		return clashAnyTLSURI(proxy)
	case "shadowtls":
		return clashShadowTLSURI(proxy)
	case "naive", "naive+quic":
		return clashNaiveURI(proxy)
	case "ssh":
		return clashSSHURI(proxy)
	case "wireguard", "wg":
		return clashWireGuardURI(proxy)
	default:
		return ""
	}
}

func clashShadowsocksURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	method := firstMapValue(proxy, "cipher", "method")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || method == "" || password == "" {
		return ""
	}
	return (&url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(method, password),
		Host:     net.JoinHostPort(server, port),
		Fragment: firstMapValue(proxy, "name"),
	}).String()
}

func clashTrojanURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || password == "" {
		return ""
	}
	values := url.Values{}
	if tlsEnabled(proxy) {
		values.Set("security", "tls")
	}
	appendClashTLSQueryValues(proxy, values)
	return proxyURL("trojan", password, server, port, values, firstMapValue(proxy, "name"))
}

func clashVLESSURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	uuid := firstMapValue(proxy, "uuid")
	if server == "" || port == "" || uuid == "" {
		return ""
	}
	values := url.Values{}
	if tlsEnabled(proxy) {
		values.Set("security", "tls")
	}
	if flow := firstMapValue(proxy, "flow"); flow != "" {
		values.Set("flow", flow)
	}
	appendClashTransportQueryValues(proxy, values)
	appendClashTLSQueryValues(proxy, values)
	return proxyURL("vless", uuid, server, port, values, firstMapValue(proxy, "name"))
}

func clashVMessURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	uuid := firstMapValue(proxy, "uuid")
	if server == "" || port == "" || uuid == "" {
		return ""
	}
	doc := map[string]string{
		"v":    "2",
		"ps":   firstMapValue(proxy, "name"),
		"add":  server,
		"port": port,
		"id":   uuid,
		"aid":  firstNonEmptyString(firstMapValue(proxy, "alterid", "alter-id", "alter_id"), "0"),
		"scy":  firstNonEmptyString(firstMapValue(proxy, "cipher"), "auto"),
		"net":  firstNonEmptyString(firstMapValue(proxy, "network"), "tcp"),
		"type": firstMapValue(proxy, "header-type", "header_type"),
		"host": firstMapValue(proxy, "ws-headers.host", "host"),
		"path": firstMapValue(proxy, "ws-path", "ws_path", "path"),
		"tls":  "",
		"sni":  firstMapValue(proxy, "sni", "servername", "server_name"),
		"alpn": firstMapValue(proxy, "alpn"),
	}
	if tlsEnabled(proxy) {
		doc["tls"] = "tls"
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		doc["allowInsecure"] = "1"
	}
	encoded, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return "vmess://" + base64.RawURLEncoding.EncodeToString(encoded)
}

func clashHysteria2URI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	password := firstMapValue(proxy, "password", "auth", "auth-str", "auth_str")
	if server == "" || port == "" || password == "" {
		return ""
	}
	values := url.Values{}
	if obfs := firstMapValue(proxy, "obfs", "obfs-type", "obfs_type"); obfs != "" {
		values.Set("obfs", obfs)
	}
	if obfsPassword := firstMapValue(proxy, "obfs-password", "obfs_password"); obfsPassword != "" {
		values.Set("obfs-password", obfsPassword)
	}
	if sni := firstMapValue(proxy, "sni", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
	if alpn := firstMapValue(proxy, "alpn"); alpn != "" {
		values.Set("alpn", alpn)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}
	return proxyURL("hysteria2", password, server, port, values, firstMapValue(proxy, "name"))
}

func clashTUICURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	uuid := firstMapValue(proxy, "uuid")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || uuid == "" || password == "" {
		return ""
	}
	values := url.Values{}
	if congestionControl := firstMapValue(proxy, "congestion-control", "congestion_control", "congestion-controller", "congestion_controller"); congestionControl != "" {
		values.Set("congestion_control", congestionControl)
	}
	if udpRelayMode := firstMapValue(proxy, "udp-relay-mode", "udp_relay_mode"); udpRelayMode != "" {
		values.Set("udp_relay_mode", udpRelayMode)
	}
	if sni := firstMapValue(proxy, "sni", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
	if alpn := firstMapValue(proxy, "alpn"); alpn != "" {
		values.Set("alpn", alpn)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}
	result := &url.URL{
		Scheme:   "tuic",
		User:     url.UserPassword(uuid, password),
		Host:     net.JoinHostPort(server, port),
		Fragment: firstMapValue(proxy, "name"),
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func clashHysteriaURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	authStr := firstMapValue(proxy, "auth-str", "auth_str", "password")
	auth := firstMapValue(proxy, "auth", "auth-base64", "auth_base64")
	if server == "" || port == "" || (authStr == "" && auth == "") {
		return ""
	}
	values := url.Values{}
	if auth != "" {
		values.Set("auth", auth)
	}
	if up := firstMapValue(proxy, "up", "up-speed", "up_speed"); up != "" {
		values.Set("up", up)
	}
	if down := firstMapValue(proxy, "down", "down-speed", "down_speed"); down != "" {
		values.Set("down", down)
	}
	if upMbps := firstMapValue(proxy, "up-mbps", "up_mbps", "upmbps"); upMbps != "" {
		values.Set("up_mbps", upMbps)
	}
	if downMbps := firstMapValue(proxy, "down-mbps", "down_mbps", "downmbps"); downMbps != "" {
		values.Set("down_mbps", downMbps)
	}
	if obfs := firstMapValue(proxy, "obfs"); obfs != "" {
		values.Set("obfs", obfs)
	}
	if network := firstMapValue(proxy, "protocol", "network"); network != "" {
		values.Set("network", network)
	}
	if sni := firstMapValue(proxy, "sni", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
	if alpn := firstMapValue(proxy, "alpn"); alpn != "" {
		values.Set("alpn", alpn)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}
	result := &url.URL{
		Scheme:   "hysteria",
		Host:     net.JoinHostPort(server, port),
		Fragment: firstMapValue(proxy, "name"),
	}
	if authStr != "" {
		result.User = url.User(authStr)
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func clashHTTPURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	if server == "" || port == "" {
		return ""
	}

	proxyType := strings.ToLower(firstMapValue(proxy, "type"))
	values := url.Values{}
	sni := firstMapValue(proxy, "sni", "servername", "server_name")
	if sni != "" {
		values.Set("sni", sni)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}

	scheme := "http"
	if proxyType == "https" || tlsEnabled(proxy) || sni != "" || values.Get("insecure") != "" {
		scheme = "https"
	}

	path := firstMapValue(proxy, "path")
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return proxyURLWithUser(scheme, clashUserInfo(proxy), server, port, path, values, firstMapValue(proxy, "name"))
}

func clashSOCKSURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	if server == "" || port == "" {
		return ""
	}

	scheme := "socks5"
	switch proxyType := strings.ToLower(firstMapValue(proxy, "type")); proxyType {
	case "socks4", "socks4a", "socks5":
		scheme = proxyType
	default:
		switch strings.ToLower(firstMapValue(proxy, "version")) {
		case "4":
			scheme = "socks4"
		case "4a":
			scheme = "socks4a"
		}
	}

	values := url.Values{}
	if network := firstMapValue(proxy, "network", "protocol"); network != "" {
		values.Set("network", network)
	}
	if boolMapValue(proxy, "udp-over-tcp", "udp_over_tcp", "uot") {
		values.Set("udp_over_tcp", "1")
	}
	return proxyURLWithUser(scheme, clashUserInfo(proxy), server, port, "", values, firstMapValue(proxy, "name"))
}

func clashAnyTLSURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || password == "" {
		return ""
	}

	values := url.Values{}
	if checkInterval := firstMapValue(proxy, "idle-session-check-interval", "idle_session_check_interval"); checkInterval != "" {
		values.Set("idle_session_check_interval", checkInterval)
	}
	if timeout := firstMapValue(proxy, "idle-session-timeout", "idle_session_timeout"); timeout != "" {
		values.Set("idle_session_timeout", timeout)
	}
	if minIdleSession := firstMapValue(proxy, "min-idle-session", "min_idle_session"); minIdleSession != "" {
		values.Set("min_idle_session", minIdleSession)
	}
	appendClashTLSQueryValues(proxy, values)
	return proxyURL("anytls", password, server, port, values, firstMapValue(proxy, "name"))
}

func clashShadowTLSURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	if server == "" || port == "" {
		return ""
	}

	version := firstNonEmptyString(firstMapValue(proxy, "version"), "1")
	password := firstMapValue(proxy, "password")
	if version != "1" && password == "" {
		return ""
	}

	values := url.Values{}
	values.Set("version", version)
	appendClashTLSQueryValues(proxy, values)
	var user *url.Userinfo
	if password != "" {
		user = url.User(password)
	}
	return proxyURLWithUser("shadowtls", user, server, port, "", values, firstMapValue(proxy, "name"))
}

func clashNaiveURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	username := firstMapValue(proxy, "username", "user")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || username == "" || password == "" {
		return ""
	}

	proxyType := strings.ToLower(firstMapValue(proxy, "type"))
	scheme := "naive"
	values := url.Values{}
	if proxyType == "naive+quic" || boolMapValue(proxy, "quic") {
		scheme = "naive+quic"
		values.Set("quic", "1")
	}
	if concurrency := firstMapValue(proxy, "insecure-concurrency", "insecure_concurrency"); concurrency != "" {
		values.Set("insecure_concurrency", concurrency)
	}
	if boolMapValue(proxy, "udp-over-tcp", "udp_over_tcp", "uot") {
		values.Set("udp_over_tcp", "1")
	}
	if congestionControl := firstMapValue(proxy, "quic-congestion-control", "quic_congestion_control"); congestionControl != "" {
		values.Set("quic_congestion_control", congestionControl)
	}
	appendClashTLSQueryValues(proxy, values)
	return proxyURLWithUser(scheme, url.UserPassword(username, password), server, port, "", values, firstMapValue(proxy, "name"))
}

func clashSSHURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstNonEmptyString(firstMapValue(proxy, "port"), "22")
	username := firstMapValue(proxy, "username", "user")
	if server == "" || username == "" {
		return ""
	}

	values := url.Values{}
	for _, item := range []struct {
		query string
		keys  []string
	}{
		{query: "private_key", keys: []string{"private-key", "private_key"}},
		{query: "private_key_path", keys: []string{"private-key-path", "private_key_path"}},
		{query: "private_key_passphrase", keys: []string{"private-key-passphrase", "private_key_passphrase"}},
		{query: "client_version", keys: []string{"client-version", "client_version"}},
		{query: "host_key", keys: []string{"host-key", "host_key"}},
		{query: "host_key_algorithms", keys: []string{"host-key-algorithms", "host_key_algorithms"}},
		{query: "cipher", keys: []string{"cipher"}},
		{query: "mac", keys: []string{"mac"}},
		{query: "kex_algorithm", keys: []string{"kex-algorithm", "kex_algorithm"}},
	} {
		if value := firstMapValue(proxy, item.keys...); value != "" {
			values.Set(item.query, value)
		}
	}

	password := firstMapValue(proxy, "password")
	user := url.User(username)
	if password != "" {
		user = url.UserPassword(username, password)
	}
	return proxyURLWithUser("ssh", user, server, port, "", values, firstMapValue(proxy, "name"))
}

func clashWireGuardURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstNonEmptyString(firstMapValue(proxy, "port"), "51820")
	privateKey := firstMapValue(proxy, "private-key", "private_key")
	peerPublicKey := firstMapValue(proxy, "public-key", "public_key", "peer-public-key", "peer_public_key")
	localAddress := clashWireGuardLocalAddress(proxy)
	if server == "" || privateKey == "" || peerPublicKey == "" || localAddress == "" {
		return ""
	}

	values := url.Values{}
	values.Set("private_key", privateKey)
	values.Set("peer_public_key", peerPublicKey)
	values.Set("local_address", localAddress)
	for _, item := range []struct {
		query string
		keys  []string
	}{
		{query: "pre_shared_key", keys: []string{"pre-shared-key", "pre_shared_key", "preshared-key", "preshared_key", "psk"}},
		{query: "allowed_ips", keys: []string{"allowed-ips", "allowed_ips", "peer-allowed-ips", "peer_allowed_ips"}},
		{query: "reserved", keys: []string{"reserved", "peer-reserved", "peer_reserved"}},
		{query: "workers", keys: []string{"workers"}},
		{query: "mtu", keys: []string{"mtu"}},
		{query: "network", keys: []string{"network", "protocol"}},
		{query: "interface_name", keys: []string{"interface-name", "interface_name"}},
	} {
		if value := firstMapValue(proxy, item.keys...); value != "" {
			values.Set(item.query, value)
		}
	}
	if boolMapValue(proxy, "udp") && values.Get("network") == "" {
		values.Set("network", "udp")
	}
	if boolMapValue(proxy, "system-interface", "system_interface", "system") {
		values.Set("system_interface", "1")
	}
	return proxyURLWithUser("wireguard", nil, server, port, "", values, firstMapValue(proxy, "name"))
}

func clashWireGuardLocalAddress(proxy map[string]string) string {
	if address := firstMapValue(proxy, "local-address", "local_address", "address"); address != "" {
		return address
	}
	var addresses []string
	if ipv4 := firstMapValue(proxy, "ip", "ipv4"); ipv4 != "" {
		addresses = append(addresses, ipv4)
	}
	if ipv6 := firstMapValue(proxy, "ipv6"); ipv6 != "" {
		addresses = append(addresses, ipv6)
	}
	return strings.Join(addresses, ",")
}

func proxyURL(scheme, user, server, port string, values url.Values, fragment string) string {
	result := &url.URL{
		Scheme:   scheme,
		User:     url.User(user),
		Host:     net.JoinHostPort(server, port),
		Fragment: fragment,
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func proxyURLWithUser(scheme string, user *url.Userinfo, server, port, path string, values url.Values, fragment string) string {
	result := &url.URL{
		Scheme:   scheme,
		User:     user,
		Host:     net.JoinHostPort(server, port),
		Path:     path,
		Fragment: fragment,
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func clashUserInfo(proxy map[string]string) *url.Userinfo {
	username := firstMapValue(proxy, "username", "user")
	password := firstMapValue(proxy, "password")
	if username == "" && password == "" {
		return nil
	}
	if password != "" {
		return url.UserPassword(username, password)
	}
	return url.User(username)
}

func appendClashTLSQueryValues(proxy map[string]string, values url.Values) {
	if sni := firstMapValue(proxy, "sni", "peer", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}
	if boolMapValue(proxy, "disable-sni", "disable_sni") {
		values.Set("disable_sni", "1")
	}
	if alpn := firstMapValue(proxy, "alpn"); alpn != "" {
		values.Set("alpn", alpn)
	}
}

func appendClashTransportQueryValues(proxy map[string]string, values url.Values) {
	transportType := firstMapValue(proxy, "network", "net", "transport")
	path := firstMapValue(proxy, "ws-path", "ws_path", "path")
	host := firstMapValue(proxy, "ws-headers.host", "ws_headers.host", "ws-host", "ws_host", "host")
	serviceName := firstMapValue(proxy, "grpc-service-name", "grpc_service_name", "grpc-opts.grpc-service-name", "grpc_opts.grpc_service_name", "service-name", "service_name")
	if strings.EqualFold(transportType, "tls") {
		transportType = ""
	}
	if transportType == "" && (path != "" || host != "") {
		transportType = "ws"
	}
	if transportType == "" && serviceName != "" {
		transportType = "grpc"
	}
	if transportType != "" {
		values.Set("type", transportType)
	}
	if path != "" {
		values.Set("path", path)
	}
	if host != "" {
		values.Set("host", host)
	}
	if serviceName != "" {
		values.Set("service_name", serviceName)
	}
}

func parseInlineMap(value string) map[string]string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.TrimSuffix(value, "}"), "{")
	result := map[string]string{}
	for _, part := range splitOutsideQuotes(value, ',') {
		key, fieldValue, ok := parseYAMLField(strings.TrimSpace(part))
		if ok {
			result[strings.ToLower(key)] = fieldValue
		}
	}
	return result
}

func parseYAMLField(line string) (string, string, bool) {
	index := indexOutsideQuotes(line, ':')
	if index < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:index])
	value := strings.TrimSpace(stripInlineComment(line[index+1:]))
	if key == "" {
		return "", "", false
	}
	return key, parseYAMLScalar(value), true
}

func parseYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		unquoted, err := strconv.Unquote(value)
		if err == nil {
			return strings.TrimSpace(unquoted)
		}
	}
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return strings.TrimSpace(strings.ReplaceAll(value[1:len(value)-1], "''", "'"))
	}
	return strings.TrimSpace(value)
}

func stripInlineComment(value string) string {
	quote := rune(0)
	escaped := false
	for index, ch := range value {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && quote == '"' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		if ch == '#' && (index == 0 || value[index-1] == ' ' || value[index-1] == '\t') {
			return value[:index]
		}
	}
	return value
}

func splitOutsideQuotes(value string, separator rune) []string {
	var parts []string
	start := 0
	quote := rune(0)
	escaped := false
	for index, ch := range value {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && quote == '"' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		if ch == separator {
			parts = append(parts, value[start:index])
			start = index + 1
		}
	}
	parts = append(parts, value[start:])
	return parts
}

func indexOutsideQuotes(value string, target rune) int {
	quote := rune(0)
	escaped := false
	for index, ch := range value {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' && quote == '"' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		if ch == target {
			return index
		}
	}
	return -1
}

func mergeProxyFields(target map[string]string, fields map[string]string) {
	for key, value := range fields {
		target[key] = value
	}
}

func firstMapValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[strings.ToLower(key)]); value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func tlsEnabled(values map[string]string) bool {
	return boolMapValue(values, "tls") || strings.EqualFold(firstMapValue(values, "network"), "tls")
}

func boolMapValue(values map[string]string, keys ...string) bool {
	value := strings.ToLower(firstMapValue(values, keys...))
	return value == "true" || value == "1" || value == "yes" || value == "on"
}
