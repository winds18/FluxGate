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

	outbounds := v2rayObjectList(doc["outbounds"])
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
	switch strings.ToLower(v2rayString(outbound, "protocol")) {
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
	case "socks":
		return v2raySOCKSServerURIs(outbound)
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

func v2rayVNextURIs(outbound map[string]any, build func(map[string]string) string) []string {
	settings := v2rayMap(outbound["settings"])
	vnexts := v2rayObjectList(settings["vnext"])
	total := v2rayVNextUserCount(vnexts)
	sequence := 0
	var uris []string
	for _, vnext := range vnexts {
		users := v2rayObjectList(vnext["users"])
		if len(users) == 0 {
			users = []map[string]any{{}}
		}
		for _, user := range users {
			sequence++
			proxy := map[string]string{
				"name":    v2rayName(outbound, vnext, user, v2rayString(vnext, "address"), sequence, total),
				"server":  v2rayString(vnext, "address"),
				"port":    v2rayPort(vnext, "port"),
				"uuid":    firstNonEmptyString(v2rayString(user, "id"), v2rayString(user, "uuid")),
				"alterid": firstNonEmptyString(v2rayString(user, "alterId"), v2rayString(user, "alter_id")),
				"cipher":  firstNonEmptyString(v2rayString(user, "security"), v2rayString(user, "encryption")),
				"flow":    v2rayString(user, "flow"),
			}
			appendV2RayStreamProxyValues(outbound, proxy)
			if uri := build(proxy); uri != "" {
				uris = append(uris, uri)
			}
		}
	}
	return uris
}

func v2rayServerURIs(outbound map[string]any, protocol string, build func(map[string]string) string) []string {
	settings := v2rayMap(outbound["settings"])
	servers := v2rayObjectList(settings["servers"])
	total := len(servers)
	var uris []string
	for index, server := range servers {
		proxy := map[string]string{
			"name":     v2rayName(outbound, server, nil, firstNonEmptyString(v2rayString(server, "address"), v2rayString(server, "server")), index+1, total),
			"server":   firstNonEmptyString(v2rayString(server, "address"), v2rayString(server, "server")),
			"port":     v2rayPort(server, "port", "server_port"),
			"password": v2rayString(server, "password"),
		}
		if protocol == "shadowsocks" {
			proxy["method"] = firstNonEmptyString(v2rayString(server, "method"), v2rayString(server, "cipher"))
			if plugin := v2rayString(server, "plugin"); plugin != "" {
				proxy["plugin"] = plugin
			}
			if pluginOpts := firstNonEmptyString(v2rayString(server, "plugin_opts"), v2rayString(server, "plugin-opts")); pluginOpts != "" {
				proxy["plugin_opts"] = pluginOpts
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
	settings := v2rayMap(outbound["settings"])
	servers := v2rayObjectList(settings["servers"])
	total := v2rayProxyServerUserCount(servers)
	sequence := 0
	var uris []string
	for _, server := range servers {
		users := v2rayObjectList(server["users"])
		if len(users) == 0 {
			users = []map[string]any{{}}
		}
		for _, user := range users {
			sequence++
			proxy := map[string]string{
				"name":     v2rayName(outbound, server, user, firstNonEmptyString(v2rayString(server, "address"), v2rayString(server, "server")), sequence, total),
				"type":     proxyType,
				"server":   firstNonEmptyString(v2rayString(server, "address"), v2rayString(server, "server")),
				"port":     v2rayPort(server, "port", "server_port"),
				"username": firstNonEmptyString(v2rayString(user, "user"), v2rayString(user, "username")),
				"password": firstNonEmptyString(v2rayString(user, "pass"), v2rayString(user, "password")),
			}
			appendV2RayStreamProxyValues(outbound, proxy)
			if uri := build(proxy); uri != "" {
				uris = append(uris, uri)
			}
		}
	}
	return uris
}

func appendV2RayStreamProxyValues(outbound map[string]any, proxy map[string]string) {
	stream := v2rayMap(outbound["streamSettings"])
	if stream == nil {
		stream = v2rayMap(outbound["stream_settings"])
	}
	if stream == nil {
		return
	}

	if network := v2rayString(stream, "network"); network != "" {
		proxy["network"] = network
	}
	security := strings.ToLower(v2rayString(stream, "security"))
	tlsSettings := firstNonNilMap(v2rayMap(stream["tlsSettings"]), v2rayMap(stream["tls_settings"]), v2rayMap(stream["xtlsSettings"]))
	realitySettings := firstNonNilMap(v2rayMap(stream["realitySettings"]), v2rayMap(stream["reality_settings"]))
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
	if serverName := firstNonEmptyString(v2rayString(tls, "serverName"), v2rayString(tls, "server_name")); serverName != "" {
		proxy["sni"] = serverName
	}
	if boolFromAnyValue(tls["allowInsecure"]) || boolFromAnyValue(tls["allow_insecure"]) {
		proxy["insecure"] = "true"
	}
	if alpn := stringListFromAnyValue(tls["alpn"]); len(alpn) > 0 {
		proxy["alpn"] = strings.Join(alpn, ",")
	}
}

func appendV2RayRealityValues(reality map[string]any, proxy map[string]string) {
	if reality == nil {
		return
	}
	if serverName := firstNonEmptyString(v2rayString(reality, "serverName"), v2rayString(reality, "server_name")); serverName != "" {
		proxy["sni"] = serverName
	}
	if publicKey := firstNonEmptyString(v2rayString(reality, "publicKey"), v2rayString(reality, "public_key")); publicKey != "" {
		proxy["pbk"] = publicKey
	}
	if shortID := firstNonEmptyString(v2rayString(reality, "shortId"), v2rayString(reality, "short_id")); shortID != "" {
		proxy["sid"] = shortID
	}
	if fingerprint := firstNonEmptyString(v2rayString(reality, "fingerprint"), v2rayString(reality, "fp")); fingerprint != "" {
		proxy["fp"] = fingerprint
	}
	if spiderX := firstNonEmptyString(v2rayString(reality, "spiderX"), v2rayString(reality, "spider_x")); spiderX != "" {
		proxy["spx"] = spiderX
	}
}

func appendV2RayTransportValues(stream map[string]any, proxy map[string]string) {
	if ws := firstNonNilMap(v2rayMap(stream["wsSettings"]), v2rayMap(stream["ws_settings"])); ws != nil {
		if path := v2rayString(ws, "path"); path != "" {
			proxy["ws-path"] = path
		}
		if headers := v2rayMap(ws["headers"]); headers != nil {
			if hosts := firstNonEmptyStringList(stringListFromAnyValue(headers["Host"]), stringListFromAnyValue(headers["host"])); len(hosts) > 0 {
				proxy["ws-headers.host"] = strings.Join(hosts, ",")
			}
		}
		if maxEarlyData := intFromAnyValue(ws["maxEarlyData"]); maxEarlyData > 0 {
			proxy["max-early-data"] = strconv.Itoa(maxEarlyData)
		}
		if earlyHeader := firstNonEmptyString(v2rayString(ws, "earlyDataHeaderName"), v2rayString(ws, "early_data_header_name")); earlyHeader != "" {
			proxy["early-data-header-name"] = earlyHeader
		}
	}
	if tcp := firstNonNilMap(v2rayMap(stream["tcpSettings"]), v2rayMap(stream["tcp_settings"])); tcp != nil {
		appendV2RayTCPHeaderValues(tcp, proxy)
	}
	if grpc := firstNonNilMap(v2rayMap(stream["grpcSettings"]), v2rayMap(stream["grpc_settings"])); grpc != nil {
		if serviceName := firstNonEmptyString(v2rayString(grpc, "serviceName"), v2rayString(grpc, "service_name")); serviceName != "" {
			proxy["grpc-service-name"] = serviceName
		}
		if idleTimeout := firstNonEmptyString(v2rayString(grpc, "idleTimeout"), v2rayString(grpc, "idle_timeout")); idleTimeout != "" {
			proxy["grpc-idle-timeout"] = idleTimeout
		}
		if pingTimeout := firstNonEmptyString(
			v2rayString(grpc, "pingTimeout"),
			v2rayString(grpc, "ping_timeout"),
			v2rayString(grpc, "healthCheckTimeout"),
			v2rayString(grpc, "health_check_timeout"),
		); pingTimeout != "" {
			proxy["grpc-ping-timeout"] = pingTimeout
		}
		if boolFromAnyValue(grpc["permitWithoutStream"]) || boolFromAnyValue(grpc["permit_without_stream"]) {
			proxy["permit-without-stream"] = "1"
		}
		if boolFromAnyValue(grpc["multiMode"]) || boolFromAnyValue(grpc["multi_mode"]) {
			proxy["grpc-multi-mode"] = "1"
		}
	}
	if http := firstNonNilMap(v2rayMap(stream["httpSettings"]), v2rayMap(stream["http_settings"]), v2rayMap(stream["h2Settings"])); http != nil {
		if hosts := stringListFromAnyValue(http["host"]); len(hosts) > 0 {
			proxy["http-opts.host"] = strings.Join(hosts, ",")
		}
		if path := v2rayString(http, "path"); path != "" {
			proxy["http-opts.path"] = path
		}
		if method := v2rayString(http, "method"); method != "" {
			proxy["http-opts.method"] = method
		}
	}
	if upgrade := firstNonNilMap(v2rayMap(stream["httpupgradeSettings"]), v2rayMap(stream["httpupgrade_settings"]), v2rayMap(stream["httpUpgradeSettings"])); upgrade != nil {
		if host := v2rayString(upgrade, "host"); host != "" {
			proxy["httpupgrade-opts.host"] = host
		}
		if path := v2rayString(upgrade, "path"); path != "" {
			proxy["httpupgrade-opts.path"] = path
		}
	}
}

func appendV2RayTCPHeaderValues(tcp map[string]any, proxy map[string]string) {
	header := firstNonNilMap(v2rayMap(tcp["header"]), v2rayMap(tcp["headers"]))
	if header == nil || !strings.EqualFold(v2rayString(header, "type"), "http") {
		return
	}
	proxy["network"] = "http"
	request := v2rayMap(header["request"])
	if request == nil {
		return
	}
	if method := v2rayString(request, "method"); method != "" {
		proxy["http-opts.method"] = method
	}
	if paths := stringListFromAnyValue(request["path"]); len(paths) > 0 {
		proxy["http-opts.path"] = strings.Join(paths, ",")
	}
	if headers := v2rayMap(request["headers"]); headers != nil {
		if hosts := firstNonEmptyStringList(stringListFromAnyValue(headers["Host"]), stringListFromAnyValue(headers["host"])); len(hosts) > 0 {
			proxy["http-opts.host"] = strings.Join(hosts, ",")
		}
	}
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
		"id", "uuid", "password", "method", "cipher", "users", "vnext", "servers",
	} {
		if _, ok := value[key]; ok {
			return true
		}
	}
	return false
}

func v2rayVNextUserCount(vnexts []map[string]any) int {
	total := 0
	for _, vnext := range vnexts {
		users := v2rayObjectList(vnext["users"])
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
		users := v2rayObjectList(server["users"])
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

func v2rayString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return strings.TrimSpace(stringFromAnyValue(values[key]))
}

func v2rayPort(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if port := intFromAnyValue(values[key]); port > 0 {
			return strconv.Itoa(port)
		}
	}
	return ""
}

func firstNonNilMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
