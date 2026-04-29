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
	if sni := firstMapValue(proxy, "sni", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
	if boolMapValue(proxy, "skip-cert-verify", "skip_cert_verify", "insecure") {
		values.Set("insecure", "1")
	}
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
	if sni := firstMapValue(proxy, "sni", "servername", "server_name"); sni != "" {
		values.Set("sni", sni)
	}
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
	}
	if tlsEnabled(proxy) {
		doc["tls"] = "tls"
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
