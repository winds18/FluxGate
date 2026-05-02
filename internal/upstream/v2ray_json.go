package upstream

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

func V2RayJSONURIList(content string) string {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}

	outbounds := v2rayObjectList(v2rayValue(doc, "outbounds"))
	if len(outbounds) == 0 {
		return ""
	}

	var uris []string
	for _, outbound := range outbounds {
		uris = append(uris, v2rayOutboundURIs(outbound)...)
	}
	return strings.Join(uris, "\n")
}

func v2rayOutboundURIs(outbound map[string]any) []string {
	protocol := normalizedV2RayProtocol(v2rayString(outbound, "protocol"))
	switch protocol {
	case "vmess":
		return v2rayVNextURIs(outbound, clashVMessURI)
	case "vless":
		return v2rayVNextURIs(outbound, clashVLESSURI)
	case "trojan":
		return v2rayServerURIs(outbound, "trojan", clashTrojanURI)
	case "shadowsocks", "ss":
		return v2rayServerURIs(outbound, "shadowsocks", clashShadowsocksURI)
	case "http":
		return v2rayHTTPServerURIs(outbound)
	case "https":
		return v2rayProxyServerURIs(outbound, "https", clashHTTPURI)
	case "socks":
		return v2raySOCKSServerURIs(outbound)
	case "socks4", "socks4a", "socks5":
		return v2rayProxyServerURIs(outbound, protocol, clashSOCKSURI)
	case "freedom", "direct":
		uri := clashInternalURI("direct", map[string]string{"name": v2rayName(outbound, nil, nil, "Direct", 1, 1)}, "Direct")
		return []string{uri}
	case "blackhole", "block":
		uri := clashInternalURI("block", map[string]string{"name": v2rayName(outbound, nil, nil, "Block", 1, 1)}, "Block")
		return []string{uri}
	case "dns":
		uri := clashInternalURI("dns", map[string]string{"name": v2rayName(outbound, nil, nil, "DNS", 1, 1)}, "DNS")
		return []string{uri}
	default:
		return nil
	}
}

func normalizedV2RayProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "trojan-go":
		return "trojan"
	case "vmess-aead":
		return "vmess"
	case "http+tls", "http-tls", "https":
		return "https"
	case "socks5h":
		return "socks5"
	case "socks4", "socks4a", "socks5":
		return strings.ToLower(strings.TrimSpace(protocol))
	case "reject", "reject-drop", "reject-no-drop", "reject-tinygif":
		return "block"
	default:
		return strings.ToLower(strings.TrimSpace(protocol))
	}
}

func v2rayVNextURIs(outbound map[string]any, build func(map[string]string) string) []string {
	settings := v2rayFieldMap(outbound, "settings")
	vnexts := v2rayObjectList(v2rayValue(settings, "vnext"))
	total := v2rayVNextUserCount(vnexts)
	sequence := 0
	var uris []string
	for _, vnext := range vnexts {
		users := v2rayObjectList(v2rayValue(vnext, "users"))
		if len(users) == 0 {
			users = []map[string]any{{}}
		}
		server := v2rayVNextServer(vnext)
		for _, user := range users {
			sequence++
			proxy := map[string]string{
				"name":    v2rayName(outbound, vnext, user, server, sequence, total),
				"server":  server,
				"port":    v2rayPort(vnext),
				"uuid":    v2rayVNextUserID(user),
				"alterid": firstNonEmptyString(v2rayString(user, "alterId"), v2rayString(user, "alter_id"), v2rayString(user, "alter-id"), v2rayString(user, "aid")),
				"cipher":  firstNonEmptyString(v2rayString(user, "security"), v2rayString(user, "encryption"), v2rayString(user, "cipher")),
				"flow":    v2rayVNextUserFlow(user),
			}
			if packetEncoding := firstNonEmptyString(
				v2rayString(user, "packetEncoding"),
				v2rayString(user, "packet_encoding"),
				v2rayString(user, "packet-encoding"),
			); packetEncoding != "" {
				proxy["packet-encoding"] = packetEncoding
			}
			appendV2RayStreamProxyValues(outbound, proxy)
			if uri := build(proxy); uri != "" {
				uris = append(uris, uri)
			}
		}
	}
	return uris
}

func v2rayVNextUserID(user map[string]any) string {
	return firstNonEmptyString(
		v2rayString(user, "id"),
		v2rayString(user, "uuid"),
		v2rayString(user, "userId"),
		v2rayString(user, "user_id"),
		v2rayString(user, "user-id"),
	)
}

func v2rayVNextUserFlow(user map[string]any) string {
	return firstNonEmptyString(
		v2rayString(user, "flow"),
		v2rayString(user, "flowName"),
		v2rayString(user, "flow_name"),
		v2rayString(user, "flow-name"),
		v2rayString(user, "xtlsFlow"),
		v2rayString(user, "xtls_flow"),
		v2rayString(user, "xtls-flow"),
	)
}

func v2rayVNextServer(vnext map[string]any) string {
	return v2rayEndpointAddress(vnext)
}

func v2rayEndpointAddress(values map[string]any) string {
	return firstNonEmptyString(
		v2rayString(values, "address"),
		v2rayString(values, "server"),
		v2rayString(values, "host"),
		v2rayString(values, "add"),
		v2rayString(values, "serverAddress"),
		v2rayString(values, "serverHost"),
		v2rayString(values, "remoteHost"),
		v2rayString(values, "nodeHost"),
		v2rayString(values, "endpoint"),
	)
}

func v2rayServerURIs(outbound map[string]any, protocol string, build func(map[string]string) string) []string {
	settings := v2rayFieldMap(outbound, "settings")
	servers := v2rayObjectList(v2rayValue(settings, "servers"))
	total := len(servers)
	var uris []string
	for index, server := range servers {
		address := v2rayEndpointAddress(server)
		proxy := map[string]string{
			"name":     v2rayName(outbound, server, nil, address, index+1, total),
			"server":   address,
			"port":     v2rayPort(server),
			"password": firstNonEmptyString(v2rayString(server, "password"), v2rayString(server, "pass")),
		}
		if protocol == "shadowsocks" {
			proxy["method"] = firstNonEmptyString(v2rayString(server, "method"), v2rayString(server, "cipher"), v2rayString(server, "security"), v2rayString(server, "encryption"))
			if plugin := firstNonEmptyString(v2rayString(server, "plugin"), v2rayString(settings, "plugin")); plugin != "" {
				proxy["plugin"] = plugin
			}
			if pluginOpts := firstNonEmptyString(
				v2rayString(server, "plugin_opts"),
				v2rayString(server, "plugin-opts"),
				v2rayString(server, "plugin_options"),
				v2rayString(server, "plugin-options"),
				v2rayString(settings, "plugin_opts"),
				v2rayString(settings, "plugin-opts"),
				v2rayString(settings, "plugin_options"),
				v2rayString(settings, "plugin-options"),
			); pluginOpts != "" {
				proxy["plugin_opts"] = pluginOpts
			}
			if network := firstNonEmptyString(
				v2rayString(server, "network"),
				v2rayString(server, "protocol"),
				v2rayString(settings, "network"),
				v2rayString(settings, "protocol"),
			); network != "" {
				proxy["network"] = network
			}
		} else {
			appendV2RayStreamProxyValues(outbound, proxy)
		}
		if uri := build(proxy); uri != "" {
			uris = append(uris, uri)
		}
	}
	return uris
}

func v2rayHTTPServerURIs(outbound map[string]any) []string {
	return v2rayProxyServerURIs(outbound, "http", clashHTTPURI)
}

func v2raySOCKSServerURIs(outbound map[string]any) []string {
	return v2rayProxyServerURIs(outbound, "socks5", clashSOCKSURI)
}

func v2rayProxyServerURIs(outbound map[string]any, proxyType string, build func(map[string]string) string) []string {
	settings := v2rayFieldMap(outbound, "settings")
	servers := v2rayObjectList(v2rayValue(settings, "servers"))
	total := v2rayProxyServerUserCount(servers)
	sequence := 0
	var uris []string
	for _, server := range servers {
		users := v2rayProxyServerUsers(server)
		if len(users) == 0 {
			users = []map[string]any{{}}
		}
		address := v2rayEndpointAddress(server)
		for _, user := range users {
			sequence++
			proxy := map[string]string{
				"name":     v2rayName(outbound, server, user, address, sequence, total),
				"type":     proxyType,
				"server":   address,
				"port":     v2rayPort(server),
				"username": firstNonEmptyString(v2rayString(user, "user"), v2rayString(user, "username"), v2rayString(user, "account")),
				"password": firstNonEmptyString(v2rayString(user, "pass"), v2rayString(user, "password")),
			}
			appendV2RayStreamProxyValues(outbound, proxy)
			if proxyType == "socks5" && v2raySOCKSUDPEnabled(settings, server) {
				proxy["udp"] = "true"
			}
			if proxyType == "socks5" && v2raySOCKSUDPOverTCPEnabled(settings, server) {
				proxy["udp_over_tcp"] = "true"
			}
			if uri := build(proxy); uri != "" {
				uris = append(uris, uri)
			}
		}
	}
	return uris
}

func v2rayProxyServerUsers(server map[string]any) []map[string]any {
	for _, key := range []string{"users", "accounts"} {
		if users := v2rayObjectList(v2rayValue(server, key)); len(users) > 0 {
			return users
		}
		if key == "accounts" {
			if users := v2rayScalarAccountUsers(v2rayValue(server, key)); len(users) > 0 {
				return users
			}
		}
	}
	if user := firstNonEmptyString(v2rayString(server, "user"), v2rayString(server, "username"), v2rayString(server, "account")); user != "" {
		return []map[string]any{{
			"user": user,
			"pass": firstNonEmptyString(v2rayString(server, "pass"), v2rayString(server, "password")),
		}}
	}
	if password := firstNonEmptyString(v2rayString(server, "pass"), v2rayString(server, "password")); password != "" {
		return []map[string]any{{"pass": password}}
	}
	return nil
}

func v2rayScalarAccountUsers(value any) []map[string]any {
	accounts, ok := value.(map[string]any)
	if !ok || v2rayLooksLikeObject(accounts) {
		return nil
	}
	keys := make([]string, 0, len(accounts))
	for key := range accounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	users := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		password := strings.TrimSpace(stringFromAnyValue(accounts[key]))
		if key == "" || password == "" {
			continue
		}
		users = append(users, map[string]any{
			"name": key,
			"user": key,
			"pass": password,
		})
	}
	return users
}

func v2raySOCKSUDPEnabled(settings, server map[string]any) bool {
	for _, values := range []map[string]any{settings, server} {
		if boolFromAnyValue(v2rayValue(values, "udp")) ||
			boolFromAnyValue(v2rayValue(values, "udpEnabled")) ||
			boolFromAnyValue(v2rayValue(values, "udp_relay")) {
			return true
		}
	}
	return false
}

func v2raySOCKSUDPOverTCPEnabled(settings, server map[string]any) bool {
	for _, values := range []map[string]any{settings, server} {
		if boolFromAnyValue(v2rayValue(values, "udp_over_tcp")) ||
			boolFromAnyValue(v2rayValue(values, "udpOverTcp")) ||
			boolFromAnyValue(v2rayValue(values, "uot")) {
			return true
		}
	}
	return false
}

func appendV2RayStreamProxyValues(outbound map[string]any, proxy map[string]string) {
	stream := v2rayFieldMap(outbound, "streamSettings")
	if stream == nil {
		return
	}

	if network := v2rayString(stream, "network"); network != "" {
		proxy["network"] = network
	}
	security := strings.ToLower(v2rayString(stream, "security"))
	tlsSettings := firstNonNilMap(v2rayFieldMap(stream, "tlsSettings"), v2rayFieldMap(stream, "xtlsSettings"))
	realitySettings := v2rayFieldMap(stream, "realitySettings")
	if security == "tls" || security == "xtls" || tlsSettings != nil {
		proxy["tls"] = "true"
	}
	if security == "reality" || realitySettings != nil {
		proxy["security"] = "reality"
	}
	appendV2RayTLSValues(tlsSettings, proxy)
	appendV2RayRealityValues(realitySettings, proxy)
	appendV2RayTransportValues(stream, proxy)
}

func appendV2RayTLSValues(tls map[string]any, proxy map[string]string) {
	if tls == nil {
		return
	}
	if serverNames := firstNonEmptyStringList(
		stringListFromAnyValue(v2rayValue(tls, "serverName")),
		stringListFromAnyValue(v2rayValue(tls, "serverNames")),
	); len(serverNames) > 0 {
		proxy["sni"] = serverNames[0]
	}
	if boolFromAnyValue(v2rayValue(tls, "allowInsecure")) ||
		boolFromAnyValue(v2rayValue(tls, "skip_cert_verify")) ||
		boolFromAnyValue(v2rayValue(tls, "insecure")) {
		proxy["insecure"] = "true"
	}
	if boolFromAnyValue(v2rayValue(tls, "disable_sni")) {
		proxy["disable_sni"] = "true"
	}
	if alpn := stringListFromAnyValue(v2rayValue(tls, "alpn")); len(alpn) > 0 {
		proxy["alpn"] = strings.Join(alpn, ",")
	}
	if fingerprint := firstNonEmptyString(
		v2rayString(tls, "fingerprint"),
		v2rayString(tls, "fp"),
		v2rayString(tls, "clientFingerprint"),
		v2rayString(tls, "client_fingerprint"),
		v2rayString(tls, "client-fingerprint"),
		v2rayUTLSFingerprint(tls),
	); fingerprint != "" {
		proxy["fp"] = fingerprint
	}
}

func appendV2RayRealityValues(reality map[string]any, proxy map[string]string) {
	if reality == nil {
		return
	}
	if serverNames := firstNonEmptyStringList(
		stringListFromAnyValue(v2rayValue(reality, "serverName")),
		stringListFromAnyValue(v2rayValue(reality, "serverNames")),
	); len(serverNames) > 0 {
		proxy["sni"] = serverNames[0]
	}
	if publicKey := firstNonEmptyString(v2rayString(reality, "publicKey"), v2rayString(reality, "public_key"), v2rayString(reality, "public-key")); publicKey != "" {
		proxy["pbk"] = publicKey
	}
	if shortIDs := firstNonEmptyStringList(
		stringListFromAnyValue(v2rayValue(reality, "shortId")),
		stringListFromAnyValue(v2rayValue(reality, "shortIds")),
	); len(shortIDs) > 0 {
		proxy["sid"] = shortIDs[0]
	}
	if fingerprint := firstNonEmptyString(v2rayString(reality, "fingerprint"), v2rayString(reality, "fp")); fingerprint != "" {
		proxy["fp"] = fingerprint
	}
	if spiderX := firstNonEmptyString(v2rayString(reality, "spiderX"), v2rayString(reality, "spider_x"), v2rayString(reality, "spider-x")); spiderX != "" {
		proxy["spx"] = spiderX
	}
}

func v2rayUTLSFingerprint(tls map[string]any) string {
	utls := v2rayFieldMap(tls, "utls")
	if utls == nil {
		return ""
	}
	return firstNonEmptyString(v2rayString(utls, "fingerprint"), v2rayString(utls, "fp"))
}

func appendV2RayTransportValues(stream map[string]any, proxy map[string]string) {
	if ws := v2rayFieldMap(stream, "wsSettings"); ws != nil {
		if path := v2rayString(ws, "path"); path != "" {
			proxy["ws-path"] = path
		}
		if hosts := v2rayHeaderHosts(ws); len(hosts) > 0 {
			proxy["ws-headers.host"] = strings.Join(hosts, ",")
		}
		maxEarlyData := firstPositiveIntFromAnyValue(v2rayValue(ws, "maxEarlyData"))
		if maxEarlyData > 0 {
			proxy["max-early-data"] = strconv.Itoa(maxEarlyData)
		}
		if earlyHeader := firstNonEmptyString(
			v2rayString(ws, "earlyDataHeaderName"),
			v2rayString(ws, "early_data_header_name"),
			v2rayString(ws, "early-data-header-name"),
		); earlyHeader != "" {
			proxy["early-data-header-name"] = earlyHeader
		}
	}
	if tcp := v2rayFieldMap(stream, "tcpSettings"); tcp != nil {
		appendV2RayTCPHeaderValues(tcp, proxy)
	}
	if grpc := v2rayFieldMap(stream, "grpcSettings"); grpc != nil {
		if serviceName := firstNonEmptyString(v2rayString(grpc, "serviceName"), v2rayString(grpc, "service_name"), v2rayString(grpc, "service-name")); serviceName != "" {
			proxy["grpc-service-name"] = serviceName
		}
		if idleTimeout := firstNonEmptyString(v2rayString(grpc, "idleTimeout"), v2rayString(grpc, "idle_timeout"), v2rayString(grpc, "idle-timeout")); idleTimeout != "" {
			proxy["grpc-idle-timeout"] = idleTimeout
		}
		if pingTimeout := firstNonEmptyString(
			v2rayString(grpc, "pingTimeout"),
			v2rayString(grpc, "ping_timeout"),
			v2rayString(grpc, "ping-timeout"),
			v2rayString(grpc, "healthCheckTimeout"),
			v2rayString(grpc, "health_check_timeout"),
			v2rayString(grpc, "health-check-timeout"),
		); pingTimeout != "" {
			proxy["grpc-ping-timeout"] = pingTimeout
		}
		if boolFromAnyValue(v2rayValue(grpc, "permitWithoutStream")) {
			proxy["permit-without-stream"] = "1"
		}
		if boolFromAnyValue(v2rayValue(grpc, "multiMode")) {
			proxy["grpc-multi-mode"] = "1"
		}
	}
	if quic := v2rayFieldMap(stream, "quicSettings"); quic != nil && proxy["network"] == "" {
		proxy["network"] = "quic"
	}
	if http := firstNonNilMap(
		v2rayFieldMap(stream, "httpSettings"),
		v2rayFieldMap(stream, "h2Settings"),
	); http != nil {
		if hosts := v2rayHeaderHosts(http); len(hosts) > 0 {
			proxy["http-opts.host"] = strings.Join(hosts, ",")
		}
		if paths := stringListFromAnyValue(v2rayValue(http, "path")); len(paths) > 0 {
			proxy["http-opts.path"] = strings.Join(paths, ",")
		}
		if method := v2rayString(http, "method"); method != "" {
			proxy["http-opts.method"] = method
		}
	}
	if upgrade := firstNonNilMap(
		v2rayFieldMap(stream, "httpupgradeSettings"),
		v2rayFieldMap(stream, "httpUpgradeSettings"),
	); upgrade != nil {
		if hosts := firstNonEmptyStringList(stringListFromAnyValue(v2rayValue(upgrade, "host")), v2rayHeaderHosts(upgrade)); len(hosts) > 0 {
			proxy["httpupgrade-opts.host"] = strings.Join(hosts, ",")
		}
		if path := v2rayString(upgrade, "path"); path != "" {
			proxy["httpupgrade-opts.path"] = path
		}
	}
}

func appendV2RayTCPHeaderValues(tcp map[string]any, proxy map[string]string) {
	header := firstNonNilMap(v2rayFieldMap(tcp, "header"), v2rayFieldMap(tcp, "headers"))
	if header == nil || !strings.EqualFold(v2rayString(header, "type"), "http") {
		return
	}
	proxy["network"] = "http"
	request := v2rayFieldMap(header, "request")
	if request == nil {
		return
	}
	if method := v2rayString(request, "method"); method != "" {
		proxy["http-opts.method"] = method
	}
	if paths := stringListFromAnyValue(v2rayValue(request, "path")); len(paths) > 0 {
		proxy["http-opts.path"] = strings.Join(paths, ",")
	}
	if hosts := v2rayHeaderHosts(request); len(hosts) > 0 {
		proxy["http-opts.host"] = strings.Join(hosts, ",")
	}
}

func v2rayHeaderHosts(values map[string]any) []string {
	directHosts := firstNonEmptyStringList(
		stringListFromAnyValue(v2rayValue(values, "host")),
		stringListFromAnyValue(v2rayValue(values, "authority")),
		stringListFromAnyValue(v2rayValue(values, ":authority")),
	)
	if len(directHosts) > 0 {
		return directHosts
	}
	headers := v2rayFieldMap(values, "headers")
	if headers == nil {
		return nil
	}
	return firstNonEmptyStringList(
		stringListFromAnyValue(v2rayValue(headers, "Host")),
		stringListFromAnyValue(v2rayValue(headers, "authority")),
		stringListFromAnyValue(v2rayValue(headers, ":authority")),
	)
}

func v2rayObjectList(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped := v2rayMap(item); mapped != nil {
				result = append(result, mapped)
			}
		}
		return result
	case map[string]any:
		if v2rayLooksLikeObject(typed) {
			return []map[string]any{typed}
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		result := make([]map[string]any, 0, len(keys))
		for _, key := range keys {
			appendV2RayMappedObjects(&result, key, typed[key])
		}
		return result
	default:
		return nil
	}
}

func appendV2RayMappedObjects(result *[]map[string]any, key string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		if v2rayString(typed, "name") == "" && v2rayString(typed, "tag") == "" && v2rayString(typed, "email") == "" {
			typed["name"] = key
		}
		*result = append(*result, typed)
	case []any:
		for index, item := range typed {
			mapped := v2rayMap(item)
			if mapped == nil {
				continue
			}
			if v2rayString(mapped, "name") == "" && v2rayString(mapped, "tag") == "" && v2rayString(mapped, "email") == "" {
				name := key
				if len(typed) > 1 {
					name = key + "-" + strconv.Itoa(index+1)
				}
				mapped["name"] = name
			}
			*result = append(*result, mapped)
		}
	}
}

func v2rayLooksLikeObject(value map[string]any) bool {
	for _, key := range []string{
		"protocol", "settings", "streamSettings", "stream_settings", "address", "server", "port",
		"id", "uuid", "user", "username", "pass", "password", "method", "cipher", "users", "accounts", "vnext", "servers",
	} {
		if v2rayValue(value, key) != nil {
			return true
		}
	}
	return false
}

func v2rayVNextUserCount(vnexts []map[string]any) int {
	total := 0
	for _, vnext := range vnexts {
		users := v2rayObjectList(v2rayValue(vnext, "users"))
		if len(users) == 0 {
			total++
			continue
		}
		total += len(users)
	}
	return total
}

func v2rayProxyServerUserCount(servers []map[string]any) int {
	total := 0
	for _, server := range servers {
		users := v2rayProxyServerUsers(server)
		if len(users) == 0 {
			total++
			continue
		}
		total += len(users)
	}
	return total
}

func v2rayName(outbound, endpoint, user map[string]any, fallback string, sequence, total int) string {
	base := firstNonEmptyString(
		v2rayString(user, "email"),
		v2rayString(user, "name"),
		v2rayString(endpoint, "email"),
		v2rayString(endpoint, "name"),
		v2rayString(outbound, "tag"),
		v2rayString(outbound, "name"),
		fallback,
	)
	if base == "" {
		base = "v2ray"
	}
	if total > 1 && base == firstNonEmptyString(v2rayString(outbound, "tag"), v2rayString(outbound, "name"), fallback) {
		return base + "-" + strconv.Itoa(sequence)
	}
	return base
}

func v2rayMap(value any) map[string]any {
	mapped, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return mapped
}

func v2rayFieldMap(values map[string]any, key string) map[string]any {
	return v2rayMap(v2rayValue(values, key))
}

func v2rayValue(values map[string]any, keys ...string) any {
	return jsonFieldValue(values, keys...)
}

func v2rayString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return jsonFieldString(values, key)
}

func v2rayPort(values map[string]any, keys ...string) string {
	if len(keys) == 0 {
		keys = []string{
			"port",
			"server_port",
			"serverPort",
			"server-port",
			"remotePort",
			"remote_port",
			"remote-port",
			"nodePort",
			"node_port",
			"node-port",
			"portNumber",
			"port_number",
			"port-number",
		}
	}
	for _, key := range keys {
		if port := intFromAnyValue(v2rayValue(values, key)); port > 0 {
			return strconv.Itoa(port)
		}
	}
	return ""
}

func firstPositiveIntFromAnyValue(values ...any) int {
	for _, value := range values {
		if converted := intFromAnyValue(value); converted > 0 {
			return converted
		}
	}
	return 0
}

func firstNonNilMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
