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
	var scopes []yamlFieldScope
	inProxies := false
	proxiesIndent := -1
	proxyItemIndent := -1

	for _, rawLine := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " \t"))
		if inProxies && indent <= proxiesIndent && !strings.HasPrefix(trimmed, "- ") {
			if len(current) > 0 {
				proxies = append(proxies, current)
			}
			current = nil
			scopes = nil
			inProxies = false
			proxiesIndent = -1
			proxyItemIndent = -1
		}
		if !inProxies {
			if inlineProxies := clashYAMLInlineProxies(trimmed); len(inlineProxies) > 0 {
				proxies = append(proxies, inlineProxies...)
				continue
			}
			if isClashYAMLProxiesField(trimmed) {
				inProxies = true
				proxiesIndent = indent
			}
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			if current != nil && proxyItemIndent >= 0 && indent > proxyItemIndent {
				scopes = trimYAMLFieldScopes(scopes, indent)
				storeYAMLProxyListItem(current, scopes, strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
				continue
			}
			if len(current) > 0 {
				proxies = append(proxies, current)
			}
			current = map[string]string{}
			scopes = nil
			proxyItemIndent = indent
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			rest = stripYAMLNodeAnchor(rest)
			if strings.HasPrefix(rest, "{") && strings.HasSuffix(rest, "}") {
				mergeProxyFields(current, parseInlineMap(rest))
				continue
			}
			key, value, ok := parseYAMLField(rest)
			if ok {
				storeYAMLProxyField(current, nil, key, value)
			}
			continue
		}
		if current == nil {
			continue
		}
		scopes = trimYAMLFieldScopes(scopes, indent)
		key, value, ok := parseYAMLField(trimmed)
		if ok {
			storeYAMLProxyField(current, scopes, key, value)
			if value == "" {
				scopes = append(scopes, yamlFieldScope{
					Indent: indent,
					Key:    strings.ToLower(key),
				})
			}
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

type yamlFieldScope struct {
	Indent int
	Key    string
}

func isClashYAMLProxiesField(line string) bool {
	key, value, ok := parseYAMLField(line)
	return ok && strings.EqualFold(strings.TrimSpace(key), "proxies") && isYAMLBlockListFieldValue(value)
}

func isYAMLBlockListFieldValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	return strings.HasPrefix(value, "&") && !strings.ContainsAny(value, "[]{}")
}

func clashYAMLInlineProxies(line string) []map[string]string {
	key, value, ok := parseYAMLRawField(line)
	if !ok || !strings.EqualFold(strings.TrimSpace(key), "proxies") {
		return nil
	}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil
	}
	return parseYAMLInlineProxyList(value)
}

func normalizedClashProxyType(proxyType string) string {
	switch strings.ToLower(strings.TrimSpace(proxyType)) {
	case "trojan-go":
		return "trojan"
	case "vmess-aead":
		return "vmess"
	case "hy2":
		return "hysteria2"
	case "any-tls":
		return "anytls"
	case "shadow-tls":
		return "shadowtls"
	case "naive-quic":
		return "naive+quic"
	case "socks5h":
		return "socks5"
	default:
		return strings.ToLower(strings.TrimSpace(proxyType))
	}
}

func clashProxyURI(proxy map[string]string) string {
	proxyType := normalizedClashProxyType(firstMapValue(proxy, "type"))
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
	case "direct":
		return clashInternalURI("direct", proxy, "Direct")
	case "block", "reject", "reject-drop", "reject-no-drop", "reject-tinygif":
		return clashInternalURI("block", proxy, "Block")
	case "dns":
		return clashInternalURI("dns", proxy, "DNS")
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
	values := url.Values{}
	if plugin := firstMapValue(proxy, "plugin"); plugin != "" {
		values.Set("plugin", plugin)
	}
	if pluginOpts := firstMapValue(proxy, "plugin-opts", "plugin_opts", "plugin-options", "plugin_options"); pluginOpts != "" {
		values.Set("plugin_opts", pluginOpts)
	}
	if network := firstMapValue(proxy, "network", "protocol"); network != "" {
		values.Set("network", network)
	}
	result := &url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(method, password),
		Host:     net.JoinHostPort(server, port),
		Fragment: firstMapValue(proxy, "name"),
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func clashTrojanURI(proxy map[string]string) string {
	server := firstMapValue(proxy, "server")
	port := firstMapValue(proxy, "port")
	password := firstMapValue(proxy, "password")
	if server == "" || port == "" || password == "" {
		return ""
	}
	values := url.Values{}
	if clashRealityPublicKey(proxy) != "" || strings.EqualFold(firstMapValue(proxy, "security"), "reality") {
		values.Set("security", "reality")
	} else if tlsEnabled(proxy) || strings.EqualFold(firstMapValue(proxy, "security"), "tls") {
		values.Set("security", "tls")
	}
	appendClashTransportQueryValues(proxy, values)
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
	if clashRealityPublicKey(proxy) != "" || strings.EqualFold(firstMapValue(proxy, "security"), "reality") {
		values.Set("security", "reality")
	} else if tlsEnabled(proxy) {
		values.Set("security", "tls")
	}
	if flow := firstMapValue(proxy, "flow"); flow != "" {
		values.Set("flow", flow)
	}
	if packetEncoding := firstMapValue(proxy, "packet-encoding", "packet_encoding", "packetEncoding", "packetencoding"); packetEncoding != "" {
		values.Set("packet_encoding", packetEncoding)
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
	network := firstMapValue(proxy, "network")
	path := firstMapValue(proxy, "ws-path", "ws_path", "path")
	serviceName := firstMapValue(proxy, "grpc-service-name", "grpc_service_name", "grpc-opts.grpc-service-name", "grpc_opts.grpc_service_name", "service-name", "service_name")
	if network == "" && serviceName != "" {
		network = "grpc"
	}
	if network == "" {
		network = "tcp"
	}
	if strings.EqualFold(network, "grpc") && serviceName != "" {
		path = serviceName
	}
	doc := map[string]string{
		"v":    "2",
		"ps":   firstMapValue(proxy, "name"),
		"add":  server,
		"port": port,
		"id":   uuid,
		"aid":  firstNonEmptyString(firstMapValue(proxy, "alterid", "alter-id", "alter_id"), "0"),
		"scy":  firstNonEmptyString(firstMapValue(proxy, "cipher"), "auto"),
		"net":  network,
		"type": firstMapValue(proxy, "header-type", "header_type"),
		"host": firstMapValue(proxy, "ws-headers.host", "host"),
		"path": path,
		"tls":  "",
		"sni":  firstMapValue(proxy, "sni", "servername", "server_name"),
		"alpn": firstMapValue(proxy, "alpn"),
	}
	if grpcMultiMode := firstMapValue(proxy, "grpc-opts.multi-mode", "grpc_opts.multi_mode", "grpc-opts.multi_mode", "grpc_opts.multi-mode", "grpc-multi-mode", "grpc_multi_mode", "multi-mode", "multi_mode"); grpcMultiMode != "" {
		doc["multi_mode"] = grpcMultiMode
	}
	if packetEncoding := firstMapValue(proxy, "packet-encoding", "packet_encoding", "packetEncoding", "packetencoding"); packetEncoding != "" {
		doc["packet_encoding"] = packetEncoding
	}
	if boolMapValue(proxy, "disable-sni", "disable_sni") {
		doc["disable_sni"] = "1"
	}
	if fingerprint := firstMapValue(proxy, "fp", "fingerprint", "client-fingerprint", "client_fingerprint", "clientFingerprint", "clientfingerprint"); fingerprint != "" {
		doc["fp"] = fingerprint
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
	if upMbps := firstMapValue(proxy, "up-mbps", "up_mbps", "upmbps"); upMbps != "" {
		values.Set("up_mbps", upMbps)
	}
	if downMbps := firstMapValue(proxy, "down-mbps", "down_mbps", "downmbps"); downMbps != "" {
		values.Set("down_mbps", downMbps)
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
	if boolMapValue(proxy, "disable-sni", "disable_sni") {
		values.Set("disable_sni", "1")
	}
	if certificatePin := firstMapValue(proxy, "pinSHA256", "pin-sha256", "pin_sha256", "certificate-public-key-sha256", "certificate_public_key_sha256", "fingerprint"); certificatePin != "" {
		values.Set("pinSHA256", certificatePin)
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
	if recvWindowConn := firstMapValue(proxy, "recv-window-conn", "recv_window_conn"); recvWindowConn != "" {
		values.Set("recv_window_conn", recvWindowConn)
	}
	if recvWindow := firstMapValue(proxy, "recv-window", "recv_window"); recvWindow != "" {
		values.Set("recv_window", recvWindow)
	}
	if boolMapValue(proxy, "disable-mtu-discovery", "disable_mtu_discovery") {
		values.Set("disable_mtu_discovery", "1")
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
	if boolMapValue(proxy, "disable-sni", "disable_sni") {
		values.Set("disable_sni", "1")
	}
	if alpn := firstMapValue(proxy, "alpn"); alpn != "" {
		values.Set("alpn", alpn)
	}
	if fingerprint := firstMapValue(proxy, "client-fingerprint", "client_fingerprint", "fingerprint", "fp"); fingerprint != "" {
		values.Set("fp", fingerprint)
	}

	scheme := "http"
	if proxyType == "https" || tlsEnabled(proxy) || sni != "" || values.Get("insecure") != "" || values.Get("disable_sni") != "" || values.Get("alpn") != "" || values.Get("fp") != "" {
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
	switch proxyType := normalizedClashProxyType(firstMapValue(proxy, "type")); proxyType {
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
	} else if boolMapValue(proxy, "udp", "udp-relay", "udp_relay") {
		values.Set("udp", "1")
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

	proxyType := normalizedClashProxyType(firstMapValue(proxy, "type"))
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

func clashInternalURI(scheme string, proxy map[string]string, fallbackName string) string {
	return (&url.URL{
		Scheme:   scheme,
		Host:     "default",
		Fragment: firstNonEmptyString(firstMapValue(proxy, "name"), fallbackName),
	}).String()
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
	if publicKey := clashRealityPublicKey(proxy); publicKey != "" {
		values.Set("pbk", publicKey)
	}
	if shortID := firstMapValue(proxy, "reality-opts.short-id", "reality_opts.short_id", "short-id", "short_id", "sid"); shortID != "" {
		values.Set("sid", shortID)
	}
	if fingerprint := firstMapValue(proxy, "client-fingerprint", "client_fingerprint", "fingerprint", "fp"); fingerprint != "" {
		values.Set("fp", fingerprint)
	}
	if spiderX := firstMapValue(proxy, "reality-opts.spider-x", "reality_opts.spider_x", "spider-x", "spider_x", "spx"); spiderX != "" {
		values.Set("spx", spiderX)
	}
}

func clashRealityPublicKey(proxy map[string]string) string {
	return firstMapValue(proxy, "reality-opts.public-key", "reality_opts.public_key", "public-key", "public_key", "pbk")
}

func appendClashTransportQueryValues(proxy map[string]string, values url.Values) {
	transportType := firstMapValue(proxy, "network", "net", "transport")
	if strings.EqualFold(transportType, "h2") {
		transportType = "http"
	}
	wsPath := firstMapValue(proxy, "ws-path", "ws_path")
	wsHost := firstMapValue(proxy, "ws-headers.host", "ws_headers.host", "ws-opts.headers.host", "ws_opts.headers.host", "ws-host", "ws_host")
	wsMaxEarlyData := firstMapValue(proxy, "ws-opts.max-early-data", "ws_opts.max_early_data", "ws-opts.max_early_data", "ws_opts.max-early-data", "max-early-data", "max_early_data")
	wsEarlyDataHeaderName := firstMapValue(proxy, "ws-opts.early-data-header-name", "ws_opts.early_data_header_name", "ws-opts.early_data_header_name", "ws_opts.early-data-header-name", "early-data-header-name", "early_data_header_name")
	httpPath := firstMapValue(proxy, "http-opts.path", "http_opts.path")
	httpHost := firstMapValue(proxy, "http-opts.host", "http_opts.host", "http-opts.headers.host", "http_opts.headers.host")
	httpUpgradePath := firstMapValue(proxy, "httpupgrade-opts.path", "httpupgrade_opts.path", "http-upgrade-opts.path", "http_upgrade_opts.path")
	httpUpgradeHost := firstMapValue(proxy, "httpupgrade-opts.host", "httpupgrade_opts.host", "httpupgrade-opts.headers.host", "httpupgrade_opts.headers.host", "http-upgrade-opts.host", "http_upgrade_opts.host", "http-upgrade-opts.headers.host", "http_upgrade_opts.headers.host")
	httpMethod := firstMapValue(proxy, "http-opts.method", "http_opts.method", "method")
	httpIdleTimeout := firstMapValue(proxy, "http-opts.idle-timeout", "http_opts.idle_timeout", "idle-timeout", "idle_timeout")
	httpPingTimeout := firstMapValue(proxy, "http-opts.ping-timeout", "http_opts.ping_timeout", "ping-timeout", "ping_timeout")
	grpcIdleTimeout := firstMapValue(proxy, "grpc-opts.idle-timeout", "grpc_opts.idle_timeout", "grpc-opts.idle_timeout", "grpc_opts.idle-timeout", "grpc-idle-timeout", "grpc_idle_timeout", "idle-timeout", "idle_timeout")
	grpcPingTimeout := firstMapValue(proxy, "grpc-opts.ping-timeout", "grpc_opts.ping_timeout", "grpc-opts.ping_timeout", "grpc_opts.ping-timeout", "grpc-ping-timeout", "grpc_ping_timeout", "ping-timeout", "ping_timeout")
	grpcPermitWithoutStream := firstMapValue(proxy, "grpc-opts.permit-without-stream", "grpc_opts.permit_without_stream", "grpc-opts.permit_without_stream", "grpc_opts.permit-without-stream", "permit-without-stream", "permit_without_stream")
	grpcMultiMode := firstMapValue(proxy, "grpc-opts.multi-mode", "grpc_opts.multi_mode", "grpc-opts.multi_mode", "grpc_opts.multi-mode", "grpc-multi-mode", "grpc_multi_mode", "multi-mode", "multi_mode")
	path := firstNonEmptyString(wsPath, httpPath, httpUpgradePath, firstMapValue(proxy, "path"))
	host := firstNonEmptyString(wsHost, httpHost, httpUpgradeHost, firstMapValue(proxy, "host"))
	serviceName := firstMapValue(proxy, "grpc-service-name", "grpc_service_name", "grpc-opts.grpc-service-name", "grpc_opts.grpc_service_name", "service-name", "service_name")
	if strings.EqualFold(transportType, "http-upgrade") || strings.EqualFold(transportType, "http_upgrade") {
		transportType = "httpupgrade"
	}
	if strings.EqualFold(transportType, "tls") {
		transportType = ""
	}
	if transportType == "" && (httpPath != "" || httpHost != "" || httpMethod != "" || httpIdleTimeout != "" || httpPingTimeout != "") {
		transportType = "http"
	}
	if transportType == "" && (wsPath != "" || wsHost != "" || wsMaxEarlyData != "" || wsEarlyDataHeaderName != "") {
		transportType = "ws"
	}
	if transportType == "" && (httpUpgradePath != "" || httpUpgradeHost != "") {
		transportType = "httpupgrade"
	}
	if transportType == "" && (serviceName != "" || grpcIdleTimeout != "" || grpcPingTimeout != "" || grpcPermitWithoutStream != "" || grpcMultiMode != "") {
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
	if strings.EqualFold(transportType, "ws") || strings.EqualFold(transportType, "websocket") {
		if wsMaxEarlyData != "" {
			values.Set("max_early_data", wsMaxEarlyData)
		}
		if wsEarlyDataHeaderName != "" {
			values.Set("early_data_header_name", wsEarlyDataHeaderName)
		}
	}
	if strings.EqualFold(transportType, "http") {
		if httpMethod != "" {
			values.Set("method", httpMethod)
		}
		if httpIdleTimeout != "" {
			values.Set("idle_timeout", httpIdleTimeout)
		}
		if httpPingTimeout != "" {
			values.Set("ping_timeout", httpPingTimeout)
		}
	}
	if strings.EqualFold(transportType, "grpc") {
		if grpcIdleTimeout != "" {
			values.Set("idle_timeout", grpcIdleTimeout)
		}
		if grpcPingTimeout != "" {
			values.Set("ping_timeout", grpcPingTimeout)
		}
		if grpcPermitWithoutStream != "" {
			values.Set("permit_without_stream", grpcPermitWithoutStream)
		}
		if grpcMultiMode != "" {
			values.Set("multi_mode", grpcMultiMode)
		}
	}
}

func parseInlineMap(value string) map[string]string {
	result := map[string]string{}
	parseInlineMapFields(result, nil, value)
	return result
}

func parseInlineMapFields(result map[string]string, scopes []string, value string) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.TrimSuffix(value, "}"), "{")
	for _, part := range splitOutsideQuotes(value, ',') {
		key, fieldValue, ok := parseYAMLField(strings.TrimSpace(part))
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		if len(scopes) > 0 {
			parts := make([]string, 0, len(scopes)+1)
			parts = append(parts, scopes...)
			parts = append(parts, key)
			result[strings.Join(parts, ".")] = fieldValue
		}
		result[key] = fieldValue
		if isInlineMapLiteral(fieldValue) {
			parseInlineMapFields(result, append(scopes, key), fieldValue)
		}
	}
}

func isInlineMapLiteral(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

func trimYAMLFieldScopes(scopes []yamlFieldScope, indent int) []yamlFieldScope {
	for len(scopes) > 0 && scopes[len(scopes)-1].Indent >= indent {
		scopes = scopes[:len(scopes)-1]
	}
	return scopes
}

func storeYAMLProxyField(target map[string]string, scopes []yamlFieldScope, key, value string) {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return
	}
	if len(scopes) > 0 {
		parts := make([]string, 0, len(scopes)+1)
		for _, scope := range scopes {
			parts = append(parts, scope.Key)
		}
		parts = append(parts, key)
		target[strings.Join(parts, ".")] = value
	}
	target[key] = value
}

func storeYAMLProxyListItem(target map[string]string, scopes []yamlFieldScope, value string) {
	value = parseYAMLScalar(value)
	if len(scopes) == 0 || value == "" {
		return
	}
	parts := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		parts = append(parts, scope.Key)
	}
	scopedKey := strings.Join(parts, ".")
	appendMapCSV(target, scopedKey, value)
	if lastKey := scopes[len(scopes)-1].Key; lastKey != scopedKey {
		appendMapCSV(target, lastKey, value)
	}
}

func appendMapCSV(target map[string]string, key, value string) {
	if key == "" || value == "" {
		return
	}
	if existing := strings.TrimSpace(target[key]); existing != "" {
		target[key] = existing + "," + value
		return
	}
	target[key] = value
}

func parseYAMLField(line string) (string, string, bool) {
	key, value, ok := parseYAMLRawField(line)
	if !ok {
		return "", "", false
	}
	return key, parseYAMLScalar(value), true
}

func parseYAMLRawField(line string) (string, string, bool) {
	index := indexOutsideQuotes(line, ':')
	if index < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:index])
	value := strings.TrimSpace(stripInlineComment(line[index+1:]))
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func parseYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return parseYAMLInlineList(value)
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

func parseYAMLInlineList(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(value, "]"), "["))
	if value == "" {
		return ""
	}
	items := make([]string, 0)
	for _, part := range splitOutsideQuotes(value, ',') {
		item := parseYAMLScalar(strings.TrimSpace(part))
		if item != "" {
			items = append(items, item)
		}
	}
	return strings.Join(items, ",")
}

func parseYAMLInlineProxyList(value string) []map[string]string {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(value, "]"), "["))
	if value == "" {
		return nil
	}
	var proxies []map[string]string
	for _, part := range splitOutsideQuotes(value, ',') {
		item := stripYAMLNodeAnchor(part)
		if !isInlineMapLiteral(item) {
			continue
		}
		proxy := parseInlineMap(item)
		if len(proxy) > 0 {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

func stripYAMLNodeAnchor(value string) string {
	value = strings.TrimSpace(value)
	for strings.HasPrefix(value, "&") {
		fields := strings.Fields(value)
		if len(fields) == 0 || len(fields[0]) <= 1 {
			return value
		}
		value = strings.TrimSpace(strings.TrimPrefix(value, fields[0]))
	}
	return value
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
	depth := 0
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
		switch ch {
		case '{', '[':
			depth++
			continue
		case '}', ']':
			if depth > 0 {
				depth--
			}
			continue
		}
		if ch == separator && depth == 0 {
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
