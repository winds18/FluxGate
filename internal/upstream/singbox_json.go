package upstream

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func SingBoxJSONURIList(content string) string {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}
	outbounds := singBoxObjectList(singBoxValue(doc, "outbounds"))
	endpoints := singBoxObjectList(singBoxValue(doc, "endpoints"))
	if len(outbounds)+len(endpoints) == 0 {
		return ""
	}

	var uris []string
	for _, outbound := range outbounds {
		if uri := singBoxOutboundURI(outbound); uri != "" {
			uris = append(uris, uri)
		}
	}
	for _, endpoint := range endpoints {
		if uri := singBoxEndpointURI(endpoint); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func singBoxObjectList(value any) []map[string]any {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				items = append(items, mapped)
			}
		}
		return items
	case []map[string]any:
		return typed
	case map[string]any:
		if len(typed) == 0 {
			return nil
		}
		if singBoxString(typed, "type") == "" {
			return singBoxObjectMap(typed)
		}
		return []map[string]any{typed}
	default:
		return nil
	}
}

func singBoxObjectMap(items map[string]any) []map[string]any {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		appendSingBoxMappedObjects(&result, key, items[key])
	}
	return result
}

func appendSingBoxMappedObjects(result *[]map[string]any, key string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		appendSingBoxMappedObject(result, typed, key)
	case []any:
		for index, item := range typed {
			mapped, ok := item.(map[string]any)
			if !ok {
				continue
			}
			fallbackName := key
			if len(typed) > 1 {
				fallbackName = key + "-" + strconv.Itoa(index+1)
			}
			appendSingBoxMappedObject(result, mapped, fallbackName)
		}
	}
}

func appendSingBoxMappedObject(result *[]map[string]any, item map[string]any, fallbackName string) {
	if len(item) == 0 {
		return
	}
	if singBoxName(item) == "" {
		item["tag"] = fallbackName
	}
	*result = append(*result, item)
}

func normalizedSingBoxOutboundType(outboundType string) string {
	switch strings.ToLower(strings.TrimSpace(outboundType)) {
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
	case "http+tls", "http-tls":
		return "https"
	case "naive+https", "naive-https":
		return "naive"
	case "naive-quic":
		return "naive+quic"
	case "socks5h":
		return "socks5"
	default:
		return strings.ToLower(strings.TrimSpace(outboundType))
	}
}

func singBoxOutboundURI(outbound map[string]any) string {
	outboundType := normalizedSingBoxOutboundType(singBoxString(outbound, "type"))
	switch outboundType {
	case "shadowsocks":
		return singBoxShadowsocksURI(outbound)
	case "trojan":
		return singBoxTrojanURI(outbound)
	case "vless":
		return singBoxVLESSURI(outbound)
	case "vmess":
		return singBoxVMessURI(outbound)
	case "hysteria2":
		return singBoxHysteria2URI(outbound)
	case "tuic":
		return singBoxTUICURI(outbound)
	case "anytls":
		return singBoxAnyTLSURI(outbound)
	case "shadowtls":
		return singBoxShadowTLSURI(outbound)
	case "naive", "naive+quic":
		return singBoxNaiveURI(outbound)
	case "hysteria":
		return singBoxHysteriaURI(outbound)
	case "http", "https":
		return singBoxHTTPURI(outbound)
	case "socks", "socks4", "socks4a", "socks5":
		return singBoxSOCKSURI(outbound)
	case "ssh":
		return singBoxSSHURI(outbound)
	case "wireguard", "wg":
		return singBoxWireGuardURI(outbound)
	case "tor":
		return singBoxTorURI(outbound)
	case "dns":
		return singBoxDNSURI(outbound)
	case "direct":
		return singBoxInternalURI("direct", outbound, "Direct")
	case "block":
		return singBoxInternalURI("block", outbound, "Block")
	default:
		return ""
	}
}

func singBoxEndpointURI(endpoint map[string]any) string {
	endpointType := strings.ToLower(singBoxString(endpoint, "type"))
	switch endpointType {
	case "wireguard", "wg":
		return singBoxWireGuardEndpointURI(endpoint)
	default:
		return ""
	}
}

func singBoxShadowsocksURI(outbound map[string]any) string {
	proxy := map[string]string{
		"name":     singBoxName(outbound),
		"server":   singBoxString(outbound, "server"),
		"port":     singBoxPort(outbound),
		"method":   singBoxString(outbound, "method"),
		"password": singBoxString(outbound, "password"),
	}
	for _, key := range []string{"plugin", "plugin_opts", "network"} {
		if value := singBoxString(outbound, key); value != "" {
			proxy[key] = value
		}
	}
	return clashShadowsocksURI(proxy)
}

func singBoxTrojanURI(outbound map[string]any) string {
	proxy := map[string]string{
		"name":     singBoxName(outbound),
		"server":   singBoxString(outbound, "server"),
		"port":     singBoxPort(outbound),
		"password": singBoxString(outbound, "password"),
	}
	if tlsMap(outbound) != nil {
		proxy["tls"] = "true"
	}
	appendSingBoxTLSProxyValues(outbound, proxy)
	appendSingBoxTransportProxyValues(outbound, proxy)
	return clashTrojanURI(proxy)
}

func singBoxVLESSURI(outbound map[string]any) string {
	proxy := map[string]string{
		"name":   singBoxName(outbound),
		"server": singBoxString(outbound, "server"),
		"port":   singBoxPort(outbound),
		"uuid":   singBoxString(outbound, "uuid"),
		"flow":   singBoxString(outbound, "flow"),
	}
	if packetEncoding := firstNonEmptyString(singBoxString(outbound, "packet_encoding"), singBoxString(outbound, "packet-encoding"), singBoxString(outbound, "packetEncoding")); packetEncoding != "" {
		proxy["packet-encoding"] = packetEncoding
	}
	if tlsMap(outbound) != nil {
		proxy["tls"] = "true"
	}
	appendSingBoxTLSProxyValues(outbound, proxy)
	appendSingBoxTransportProxyValues(outbound, proxy)
	return clashVLESSURI(proxy)
}

func singBoxVMessURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	uuid := singBoxString(outbound, "uuid")
	if server == "" || port == "" || uuid == "" {
		return ""
	}

	doc := map[string]string{
		"v":    "2",
		"ps":   singBoxName(outbound),
		"add":  server,
		"port": port,
		"id":   uuid,
		"aid":  singBoxString(outbound, "alter_id"),
		"scy":  firstNonEmptyString(singBoxString(outbound, "security"), "auto"),
		"net":  "tcp",
		"type": "",
		"host": "",
		"path": "",
		"tls":  "",
		"sni":  singBoxTLSString(outbound, "server_name"),
	}
	if doc["aid"] == "" {
		doc["aid"] = "0"
	}
	if packetEncoding := firstNonEmptyString(singBoxString(outbound, "packet_encoding"), singBoxString(outbound, "packet-encoding"), singBoxString(outbound, "packetEncoding")); packetEncoding != "" {
		doc["packet_encoding"] = packetEncoding
	}
	if tls := tlsMap(outbound); tls != nil {
		doc["tls"] = "tls"
		if singBoxBool(tls, "disable_sni") {
			doc["disable_sni"] = "1"
		}
		if utls := singBoxMap(tls, "utls"); utls != nil && singBoxBool(utls, "enabled") {
			if fingerprint := singBoxString(utls, "fingerprint"); fingerprint != "" {
				doc["fp"] = fingerprint
			}
		}
	}
	if transport := singBoxMap(outbound, "transport"); transport != nil {
		if transportType := singBoxString(transport, "type"); transportType != "" {
			doc["net"] = transportType
		}
		doc["path"] = singBoxString(transport, "path")
		if strings.EqualFold(doc["net"], "grpc") {
			if serviceName := singBoxString(transport, "service_name"); serviceName != "" {
				doc["path"] = serviceName
			}
		}
		if strings.EqualFold(doc["net"], "http") || strings.EqualFold(doc["net"], "h2") {
			if hosts := singBoxStringList(transport, "host"); len(hosts) > 0 {
				doc["host"] = strings.Join(hosts, ",")
			}
		}
		if headers := singBoxMap(transport, "headers"); headers != nil {
			host := singBoxString(headers, "Host")
			if host != "" {
				doc["host"] = host
			}
		}
	}

	encoded, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return "vmess://" + base64.RawURLEncoding.EncodeToString(encoded)
}

func singBoxHysteria2URI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	password := singBoxString(outbound, "password")
	if server == "" || port == "" || password == "" {
		return ""
	}

	values := url.Values{}
	if obfs := singBoxMap(outbound, "obfs"); obfs != nil {
		if obfsType := singBoxString(obfs, "type"); obfsType != "" {
			values.Set("obfs", obfsType)
		}
		if obfsPassword := singBoxString(obfs, "password"); obfsPassword != "" {
			values.Set("obfs-password", obfsPassword)
		}
	}
	for _, item := range []struct {
		key      string
		outbound string
	}{
		{key: "up_mbps", outbound: "up_mbps"},
		{key: "down_mbps", outbound: "down_mbps"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.key, value)
		}
	}
	appendTLSQueryValues(outbound, values)
	appendCertificatePinQueryValue(outbound, values)
	return proxyURL("hysteria2", password, server, port, values, singBoxName(outbound))
}

func singBoxTUICURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	uuid := singBoxString(outbound, "uuid")
	password := singBoxString(outbound, "password")
	if server == "" || port == "" || uuid == "" || password == "" {
		return ""
	}

	values := url.Values{}
	for _, item := range []struct {
		key      string
		outbound string
	}{
		{key: "congestion_control", outbound: "congestion_control"},
		{key: "udp_relay_mode", outbound: "udp_relay_mode"},
		{key: "heartbeat", outbound: "heartbeat"},
		{key: "network", outbound: "network"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.key, value)
		}
	}
	if singBoxBool(outbound, "udp_over_stream") {
		values.Set("udp_over_stream", "1")
	}
	if singBoxBool(outbound, "zero_rtt_handshake") {
		values.Set("zero_rtt_handshake", "1")
	}
	appendTLSQueryValues(outbound, values)
	return (&url.URL{
		Scheme:   "tuic",
		User:     url.UserPassword(uuid, password),
		Host:     net.JoinHostPort(server, port),
		RawQuery: values.Encode(),
		Fragment: singBoxName(outbound),
	}).String()
}

func singBoxAnyTLSURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	password := singBoxString(outbound, "password")
	if server == "" || port == "" || password == "" {
		return ""
	}

	values := url.Values{}
	for _, item := range []struct {
		key      string
		outbound string
	}{
		{key: "idle_session_check_interval", outbound: "idle_session_check_interval"},
		{key: "idle_session_timeout", outbound: "idle_session_timeout"},
		{key: "min_idle_session", outbound: "min_idle_session"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.key, value)
		}
	}
	appendTLSQueryValues(outbound, values)
	return proxyURL("anytls", password, server, port, values, singBoxName(outbound))
}

func singBoxShadowTLSURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	if server == "" || port == "" {
		return ""
	}

	version := singBoxInt(outbound, "version")
	if version <= 0 {
		version = 1
	}
	password := singBoxString(outbound, "password")
	if version >= 2 && password == "" {
		return ""
	}

	values := url.Values{}
	values.Set("version", strconv.Itoa(version))
	appendTLSQueryValues(outbound, values)
	var user *url.Userinfo
	if password != "" {
		user = url.User(password)
	}
	return singBoxProxyURL("shadowtls", server, port, "", user, values, singBoxName(outbound))
}

func singBoxHysteriaURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	if server == "" || port == "" {
		return ""
	}

	auth := singBoxString(outbound, "auth")
	authBase64 := singBoxString(outbound, "auth_base64")
	authStr := firstNonEmptyString(singBoxString(outbound, "auth_str"), singBoxString(outbound, "password"))
	if auth == "" && authBase64 == "" && authStr == "" {
		return ""
	}

	values := url.Values{}
	if authBase64 != "" {
		values.Set("auth_base64", authBase64)
	} else if auth != "" {
		values.Set("auth", auth)
	}
	for _, item := range []struct {
		key      string
		outbound string
	}{
		{key: "up", outbound: "up"},
		{key: "up_mbps", outbound: "up_mbps"},
		{key: "down", outbound: "down"},
		{key: "down_mbps", outbound: "down_mbps"},
		{key: "obfs", outbound: "obfs"},
		{key: "recv_window_conn", outbound: "recv_window_conn"},
		{key: "recv_window", outbound: "recv_window"},
		{key: "network", outbound: "network"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.key, value)
		}
	}
	if singBoxBool(outbound, "disable_mtu_discovery") {
		values.Set("disable_mtu_discovery", "1")
	}
	appendTLSQueryValues(outbound, values)
	var user *url.Userinfo
	if authStr != "" {
		user = url.User(authStr)
	}
	return singBoxProxyURL("hysteria", server, port, "", user, values, singBoxName(outbound))
}

func singBoxNaiveURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	username := firstNonEmptyString(singBoxString(outbound, "username"), singBoxString(outbound, "user"))
	password := singBoxString(outbound, "password")
	if server == "" || port == "" || username == "" || password == "" {
		return ""
	}

	scheme := "naive"
	values := url.Values{}
	if normalizedSingBoxOutboundType(singBoxString(outbound, "type")) == "naive+quic" || singBoxBool(outbound, "quic") {
		scheme = "naive+quic"
		values.Set("quic", "1")
	}
	for _, item := range []struct {
		key      string
		outbound string
	}{
		{key: "insecure_concurrency", outbound: "insecure_concurrency"},
		{key: "quic_congestion_control", outbound: "quic_congestion_control"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.key, value)
		}
	}
	if singBoxBool(outbound, "udp_over_tcp") {
		values.Set("udp_over_tcp", "1")
	}
	appendTLSQueryValues(outbound, values)
	return proxyURLWithUser(scheme, url.UserPassword(username, password), server, port, "", values, singBoxName(outbound))
}

func singBoxHTTPURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	if server == "" || port == "" {
		return ""
	}

	scheme := "http"
	if normalizedSingBoxOutboundType(singBoxString(outbound, "type")) == "https" || tlsMap(outbound) != nil {
		scheme = "https"
	}
	values := url.Values{}
	appendTLSQueryValues(outbound, values)
	var user *url.Userinfo
	username := singBoxString(outbound, "username")
	password := singBoxString(outbound, "password")
	if username != "" || password != "" {
		if password != "" {
			user = url.UserPassword(username, password)
		} else {
			user = url.User(username)
		}
	}
	path := singBoxString(outbound, "path")
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return singBoxProxyURL(scheme, server, port, path, user, values, singBoxName(outbound))
}

func singBoxSOCKSURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	if server == "" || port == "" {
		return ""
	}

	scheme := "socks5"
	outboundType := normalizedSingBoxOutboundType(singBoxString(outbound, "type"))
	switch outboundType {
	case "socks4", "socks4a", "socks5":
		scheme = outboundType
	default:
		switch strings.ToLower(singBoxString(outbound, "version")) {
		case "4":
			scheme = "socks4"
		case "4a":
			scheme = "socks4a"
		}
	}

	values := url.Values{}
	if network := singBoxString(outbound, "network"); network != "" {
		values.Set("network", network)
	} else if singBoxBool(outbound, "udp") || singBoxBool(outbound, "udp_relay") || singBoxBool(outbound, "udp-relay") {
		values.Set("udp", "1")
	}
	if singBoxBool(outbound, "udp_over_tcp") {
		values.Set("udp_over_tcp", "1")
	}
	var user *url.Userinfo
	username := singBoxString(outbound, "username")
	password := singBoxString(outbound, "password")
	if username != "" || password != "" {
		if password != "" {
			user = url.UserPassword(username, password)
		} else {
			user = url.User(username)
		}
	}
	return singBoxProxyURL(scheme, server, port, "", user, values, singBoxName(outbound))
}

func singBoxSSHURI(outbound map[string]any) string {
	server := singBoxString(outbound, "server")
	port := singBoxPort(outbound)
	if server == "" {
		return ""
	}
	if port == "" {
		port = "22"
	}

	values := url.Values{}
	for _, item := range []struct {
		query    string
		outbound string
	}{
		{query: "private_key", outbound: "private_key"},
		{query: "private_key_path", outbound: "private_key_path"},
		{query: "private_key_passphrase", outbound: "private_key_passphrase"},
		{query: "client_version", outbound: "client_version"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.query, value)
		}
	}
	for _, item := range []struct {
		query    string
		outbound string
	}{
		{query: "host_key", outbound: "host_key"},
		{query: "host_key_algorithms", outbound: "host_key_algorithms"},
		{query: "cipher", outbound: "cipher"},
		{query: "mac", outbound: "mac"},
		{query: "kex_algorithm", outbound: "kex_algorithm"},
	} {
		if list := singBoxStringList(outbound, item.outbound); len(list) > 0 {
			values.Set(item.query, strings.Join(list, ","))
		}
	}

	var user *url.Userinfo
	username := firstNonEmptyString(singBoxString(outbound, "user"), singBoxString(outbound, "username"))
	password := singBoxString(outbound, "password")
	if username != "" || password != "" {
		if password != "" {
			user = url.UserPassword(username, password)
		} else {
			user = url.User(username)
		}
	}
	return singBoxProxyURL("ssh", server, port, "", user, values, singBoxName(outbound))
}

func singBoxWireGuardURI(outbound map[string]any) string {
	peer := singBoxFirstPeer(outbound)
	server := firstNonEmptyString(singBoxString(outbound, "server"), singBoxMapString(peer, "server"))
	port := firstNonEmptyString(singBoxPort(outbound), singBoxMapPort(peer, "server_port", "port"))
	if port == "" {
		port = "51820"
	}

	localAddress := singBoxStringList(outbound, "local_address")
	if len(localAddress) == 0 {
		localAddress = singBoxStringList(outbound, "address")
	}
	privateKey := singBoxString(outbound, "private_key")
	peerPublicKey := firstNonEmptyString(
		singBoxString(outbound, "peer_public_key"),
		singBoxString(outbound, "public_key"),
		singBoxMapString(peer, "public_key"),
		singBoxMapString(peer, "peer_public_key"),
	)
	if server == "" || privateKey == "" || peerPublicKey == "" || len(localAddress) == 0 {
		return ""
	}

	values := url.Values{}
	values.Set("private_key", privateKey)
	values.Set("peer_public_key", peerPublicKey)
	values.Set("local_address", strings.Join(localAddress, ","))
	for _, item := range []struct {
		query    string
		outbound string
	}{
		{query: "interface_name", outbound: "interface_name"},
		{query: "network", outbound: "network"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.query, value)
		}
	}
	if singBoxBool(outbound, "system_interface") {
		values.Set("system_interface", "1")
	}
	for _, item := range []struct {
		query string
		key   string
	}{
		{query: "workers", key: "workers"},
		{query: "mtu", key: "mtu"},
	} {
		if value := singBoxInt(outbound, item.key); value > 0 {
			values.Set(item.query, strconv.Itoa(value))
		}
	}
	if preSharedKey := firstNonEmptyString(singBoxString(outbound, "pre_shared_key"), singBoxMapString(peer, "pre_shared_key")); preSharedKey != "" {
		values.Set("pre_shared_key", preSharedKey)
	}
	if allowedIPs := firstNonEmptyStringList(
		singBoxStringList(outbound, "allowed_ips"),
		stringListFromAnyValue(peerValue(peer, "allowed_ips")),
	); len(allowedIPs) > 0 {
		values.Set("allowed_ips", strings.Join(allowedIPs, ","))
	}
	if reserved := firstNonEmptyStringList(
		singBoxStringList(outbound, "reserved"),
		stringListFromAnyValue(peerValue(peer, "reserved")),
	); len(reserved) > 0 {
		values.Set("reserved", strings.Join(reserved, ","))
	}
	return singBoxProxyURL("wireguard", server, port, "", nil, values, singBoxName(outbound))
}

func singBoxWireGuardEndpointURI(endpoint map[string]any) string {
	peer := singBoxFirstPeer(endpoint)
	server := firstNonEmptyString(singBoxMapString(peer, "address"), singBoxMapString(peer, "server"))
	port := firstNonEmptyString(singBoxMapPort(peer, "port", "server_port"), "51820")
	localAddress := singBoxStringList(endpoint, "address")
	if len(localAddress) == 0 {
		localAddress = singBoxStringList(endpoint, "local_address")
	}
	privateKey := singBoxString(endpoint, "private_key")
	peerPublicKey := firstNonEmptyString(
		singBoxString(endpoint, "peer_public_key"),
		singBoxMapString(peer, "public_key"),
		singBoxMapString(peer, "peer_public_key"),
	)
	if server == "" || privateKey == "" || peerPublicKey == "" || len(localAddress) == 0 {
		return ""
	}

	values := url.Values{}
	values.Set("private_key", privateKey)
	values.Set("peer_public_key", peerPublicKey)
	values.Set("local_address", strings.Join(localAddress, ","))
	if singBoxBool(endpoint, "system") || singBoxBool(endpoint, "system_interface") {
		values.Set("system_interface", "1")
	}
	if interfaceName := firstNonEmptyString(singBoxString(endpoint, "name"), singBoxString(endpoint, "interface_name")); interfaceName != "" {
		values.Set("interface_name", interfaceName)
	}
	if network := singBoxString(endpoint, "network"); network != "" {
		values.Set("network", network)
	}
	for _, item := range []struct {
		query string
		key   string
	}{
		{query: "workers", key: "workers"},
		{query: "mtu", key: "mtu"},
	} {
		if value := singBoxInt(endpoint, item.key); value > 0 {
			values.Set(item.query, strconv.Itoa(value))
		}
	}
	if preSharedKey := singBoxMapString(peer, "pre_shared_key"); preSharedKey != "" {
		values.Set("pre_shared_key", preSharedKey)
	}
	if allowedIPs := stringListFromAnyValue(peerValue(peer, "allowed_ips")); len(allowedIPs) > 0 {
		values.Set("allowed_ips", strings.Join(allowedIPs, ","))
	}
	if reserved := stringListFromAnyValue(peerValue(peer, "reserved")); len(reserved) > 0 {
		values.Set("reserved", strings.Join(reserved, ","))
	}
	return singBoxProxyURL("wireguard", server, port, "", nil, values, singBoxName(endpoint))
}

func singBoxTorURI(outbound map[string]any) string {
	values := url.Values{}
	for _, item := range []struct {
		query    string
		outbound string
	}{
		{query: "executable_path", outbound: "executable_path"},
		{query: "data_directory", outbound: "data_directory"},
	} {
		if value := singBoxString(outbound, item.outbound); value != "" {
			values.Set(item.query, value)
		}
	}
	if extraArgs := singBoxStringList(outbound, "extra_args"); len(extraArgs) > 0 {
		values.Set("extra_args", strings.Join(extraArgs, ","))
	}
	if torrc := singBoxMap(outbound, "torrc"); torrc != nil {
		for key, value := range torrc {
			option := strings.TrimSpace(key)
			if option == "" {
				continue
			}
			if scalar := strings.TrimSpace(stringFromAnyValue(value)); scalar != "" {
				values.Set("torrc."+option, scalar)
			}
		}
	}

	result := &url.URL{
		Scheme:   "tor",
		Host:     "default",
		Fragment: singBoxName(outbound),
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func singBoxDNSURI(outbound map[string]any) string {
	return singBoxInternalURI("dns", outbound, "DNS")
}

func singBoxInternalURI(scheme string, outbound map[string]any, fallbackName string) string {
	name := firstNonEmptyString(singBoxName(outbound), fallbackName)
	return (&url.URL{
		Scheme:   scheme,
		Host:     "default",
		Fragment: name,
	}).String()
}

func singBoxProxyURL(scheme, server, port, path string, user *url.Userinfo, values url.Values, fragment string) string {
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

func singBoxName(outbound map[string]any) string {
	return firstNonEmptyString(singBoxString(outbound, "tag"), singBoxString(outbound, "name"))
}

func singBoxValue(values map[string]any, keys ...string) any {
	return jsonFieldValue(values, keys...)
}

func singBoxString(outbound map[string]any, key string) string {
	return jsonFieldString(outbound, key)
}

func singBoxPort(outbound map[string]any) string {
	value := singBoxInt(outbound, "server_port")
	if value <= 0 {
		value = singBoxInt(outbound, "port")
	}
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func singBoxInt(values map[string]any, key string) int {
	return intFromAnyValue(singBoxValue(values, key))
}

func singBoxBool(values map[string]any, key string) bool {
	return boolFromAnyValue(singBoxValue(values, key))
}

func singBoxStringList(values map[string]any, key string) []string {
	return stringListFromAnyValue(singBoxValue(values, key))
}

func singBoxTLSString(outbound map[string]any, key string) string {
	tls := tlsMap(outbound)
	if tls == nil {
		return ""
	}
	return singBoxString(tls, key)
}

func appendSingBoxTLSProxyValues(outbound map[string]any, proxy map[string]string) {
	tls := tlsMap(outbound)
	if tls == nil {
		return
	}
	if serverName := singBoxString(tls, "server_name"); serverName != "" {
		proxy["sni"] = serverName
	}
	if singBoxBool(tls, "insecure") {
		proxy["insecure"] = "true"
	}
	if singBoxBool(tls, "disable_sni") {
		proxy["disable-sni"] = "true"
	}
	if alpn := singBoxStringList(tls, "alpn"); len(alpn) > 0 {
		proxy["alpn"] = strings.Join(alpn, ",")
	}
	if reality := singBoxMap(tls, "reality"); reality != nil && singBoxBool(reality, "enabled") {
		proxy["security"] = "reality"
		if publicKey := singBoxString(reality, "public_key"); publicKey != "" {
			proxy["pbk"] = publicKey
		}
		if shortID := singBoxString(reality, "short_id"); shortID != "" {
			proxy["sid"] = shortID
		}
	}
	if utls := singBoxMap(tls, "utls"); utls != nil && singBoxBool(utls, "enabled") {
		if fingerprint := singBoxString(utls, "fingerprint"); fingerprint != "" {
			proxy["fp"] = fingerprint
		}
	}
}

func appendSingBoxTransportProxyValues(outbound map[string]any, proxy map[string]string) {
	transport := singBoxMap(outbound, "transport")
	if transport == nil {
		return
	}
	if transportType := singBoxString(transport, "type"); transportType != "" {
		proxy["network"] = transportType
	}
	if path := singBoxString(transport, "path"); path != "" {
		proxy["path"] = path
	}
	if strings.EqualFold(singBoxString(transport, "type"), "ws") {
		if maxEarlyData := singBoxInt(transport, "max_early_data"); maxEarlyData > 0 {
			proxy["max_early_data"] = strconv.Itoa(maxEarlyData)
		}
		if earlyDataHeaderName := singBoxString(transport, "early_data_header_name"); earlyDataHeaderName != "" {
			proxy["early_data_header_name"] = earlyDataHeaderName
		}
	}
	if strings.EqualFold(singBoxString(transport, "type"), "http") {
		if hosts := singBoxStringList(transport, "host"); len(hosts) > 0 {
			proxy["host"] = strings.Join(hosts, ",")
		}
		if method := singBoxString(transport, "method"); method != "" {
			proxy["method"] = method
		}
		if idleTimeout := singBoxString(transport, "idle_timeout"); idleTimeout != "" {
			proxy["idle_timeout"] = idleTimeout
		}
		if pingTimeout := singBoxString(transport, "ping_timeout"); pingTimeout != "" {
			proxy["ping_timeout"] = pingTimeout
		}
	}
	if strings.EqualFold(singBoxString(transport, "type"), "httpupgrade") {
		if hosts := singBoxStringList(transport, "host"); len(hosts) > 0 {
			proxy["host"] = strings.Join(hosts, ",")
		}
	}
	if strings.EqualFold(singBoxString(transport, "type"), "grpc") {
		if idleTimeout := singBoxString(transport, "idle_timeout"); idleTimeout != "" {
			proxy["idle_timeout"] = idleTimeout
		}
		if pingTimeout := singBoxString(transport, "ping_timeout"); pingTimeout != "" {
			proxy["ping_timeout"] = pingTimeout
		}
		if singBoxBool(transport, "permit_without_stream") {
			proxy["permit_without_stream"] = "1"
		}
	}
	if serviceName := singBoxString(transport, "service_name"); serviceName != "" {
		proxy["service_name"] = serviceName
	}
	if headers := singBoxMap(transport, "headers"); headers != nil {
		if hosts := singBoxStringList(headers, "Host"); len(hosts) > 0 {
			proxy["host"] = strings.Join(hosts, ",")
		}
	}
}

func appendTLSQueryValues(outbound map[string]any, values url.Values) {
	tls := tlsMap(outbound)
	if tls == nil {
		return
	}
	if serverName := singBoxString(tls, "server_name"); serverName != "" {
		values.Set("sni", serverName)
	}
	if singBoxBool(tls, "insecure") {
		values.Set("insecure", "1")
	}
	if singBoxBool(tls, "disable_sni") {
		values.Set("disable_sni", "1")
	}
	if alpn := singBoxStringList(tls, "alpn"); len(alpn) > 0 {
		values.Set("alpn", strings.Join(alpn, ","))
	}
	if utls := singBoxMap(tls, "utls"); utls != nil && singBoxBool(utls, "enabled") {
		if fingerprint := singBoxString(utls, "fingerprint"); fingerprint != "" {
			values.Set("fp", fingerprint)
		}
	}
}

func appendCertificatePinQueryValue(outbound map[string]any, values url.Values) {
	tls := tlsMap(outbound)
	if tls == nil {
		return
	}
	pins := firstNonEmptyStringList(
		singBoxStringList(tls, "certificate_public_key_sha256"),
		singBoxStringList(tls, "certificate-public-key-sha256"),
	)
	if len(pins) > 0 {
		values.Set("pinSHA256", strings.Join(pins, ","))
	}
}

func tlsMap(outbound map[string]any) map[string]any {
	tls := singBoxMap(outbound, "tls")
	if tls == nil {
		return nil
	}
	if enabled := singBoxValue(tls, "enabled"); enabled != nil && !boolFromAnyValue(enabled) {
		return nil
	}
	return tls
}

func singBoxMap(outbound map[string]any, key string) map[string]any {
	value, ok := singBoxValue(outbound, key).(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func singBoxFirstPeer(outbound map[string]any) map[string]any {
	peers, ok := singBoxValue(outbound, "peers").([]any)
	if !ok || len(peers) == 0 {
		return nil
	}
	peer, _ := peers[0].(map[string]any)
	return peer
}

func singBoxMapString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return singBoxString(values, key)
}

func singBoxMapPort(values map[string]any, keys ...string) string {
	if values == nil {
		return ""
	}
	for _, key := range keys {
		if value := singBoxInt(values, key); value > 0 {
			return strconv.Itoa(value)
		}
	}
	return ""
}

func peerValue(values map[string]any, key string) any {
	if values == nil {
		return nil
	}
	return singBoxValue(values, key)
}

func firstNonEmptyStringList(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func stringFromAnyValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return ""
	}
}

func stringListFromAnyValue(value any) []string {
	switch typed := value.(type) {
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if value := strings.TrimSpace(stringFromAnyValue(item)); value != "" {
				result = append(result, value)
			}
		}
		return result
	case []string:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if value := strings.TrimSpace(item); value != "" {
				result = append(result, value)
			}
		}
		return result
	case string:
		return splitCommaList(typed)
	default:
		return nil
	}
}

func splitCommaList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func intFromAnyValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return 0
}

func boolFromAnyValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "on":
			return true
		default:
			return false
		}
	case float64:
		return typed != 0
	case json.Number:
		parsed, err := typed.Int64()
		return err == nil && parsed != 0
	default:
		return false
	}
}
