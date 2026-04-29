package upstream

import (
	"strings"
)

func SurgeProxyURIList(content string) string {
	if strings.HasPrefix(strings.TrimSpace(content), "{") {
		return ""
	}
	var uris []string
	inProxySection := false
	sawSection := false
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			section := strings.ToLower(strings.TrimSpace(strings.Trim(line, "[]")))
			inProxySection = section == "proxy" || section == "proxies" || section == "proxy list"
			sawSection = true
			continue
		}
		if sawSection && !inProxySection {
			continue
		}
		if uri := surgeProxyURI(line); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func surgeProxyURI(line string) string {
	name, fields, ok := parseSurgeProxyLine(line)
	if !ok || len(fields) == 0 {
		return ""
	}

	protocol := strings.ToLower(fields[0])
	positionals, options := surgePositionalsAndOptions(fields[1:])
	proxy := map[string]string{
		"name":     name,
		"type":     protocol,
		"server":   surgeFirstValue(options, positionals, 0, "server", "host", "address"),
		"port":     surgeFirstValue(options, positionals, 1, "port", "server-port", "server_port"),
		"username": surgeFirstValue(options, positionals, 2, "username", "user"),
		"password": surgeFirstValue(options, positionals, 3, "password", "passwd", "psk"),
	}
	surgeApplyTLSOptions(proxy, options)
	surgeApplyTransportOptions(proxy, options)
	surgeApplyNetworkOptions(proxy, options)

	switch protocol {
	case "ss", "shadowsocks":
		proxy["method"] = surgeFirstValue(options, positionals, 2, "encrypt-method", "method", "cipher")
		proxy["password"] = surgeFirstValue(options, positionals, 3, "password", "passwd")
		if plugin := surgeOption(options, "plugin"); plugin != "" {
			proxy["plugin"] = plugin
		}
		if pluginOpts := surgeOption(options, "plugin-opts", "plugin_opts", "plugin-options", "plugin_options"); pluginOpts != "" {
			proxy["plugin_opts"] = pluginOpts
		}
		return clashShadowsocksURI(proxy)
	case "vless":
		proxy["uuid"] = surgeFirstValue(options, positionals, 2, "uuid", "id", "username", "user", "password")
		if flow := surgeOption(options, "flow"); flow != "" {
			proxy["flow"] = flow
		}
		return clashVLESSURI(proxy)
	case "vmess", "vmess-aead":
		proxy["uuid"] = surgeFirstValue(options, positionals, 2, "uuid", "id", "username", "user", "password")
		proxy["alterid"] = surgeOption(options, "alter-id", "alter_id", "alterid", "aid")
		proxy["cipher"] = surgeOption(options, "encrypt-method", "method", "cipher", "security")
		if headerType := surgeOption(options, "header-type", "header_type"); headerType != "" {
			proxy["header-type"] = headerType
		}
		return clashVMessURI(proxy)
	case "hysteria2", "hy2":
		proxy["type"] = "hysteria2"
		proxy["password"] = surgeFirstValue(options, positionals, 2, "password", "passwd", "auth", "auth-str", "auth_str", "token")
		surgeCopyOptions(proxy, options,
			"obfs", "obfs-type", "obfs_type",
			"obfs-password", "obfs_password",
			"up-mbps", "up_mbps", "upmbps",
			"down-mbps", "down_mbps", "downmbps",
			"pinsha256", "pin-sha256", "pin_sha256",
			"certificate-public-key-sha256", "certificate_public_key_sha256", "fingerprint",
		)
		return clashHysteria2URI(proxy)
	case "tuic":
		proxy["uuid"] = surgeFirstValue(options, positionals, 2, "uuid", "id", "username", "user")
		proxy["password"] = surgeFirstValue(options, positionals, 3, "password", "passwd", "psk", "token")
		surgeCopyOptions(proxy, options,
			"congestion-control", "congestion_control", "congestion-controller", "congestion_controller",
			"udp-relay-mode", "udp_relay_mode",
		)
		return clashTUICURI(proxy)
	case "trojan":
		proxy["password"] = surgeFirstValue(options, positionals, 2, "password", "passwd", "psk")
		if proxy["tls"] == "" {
			proxy["tls"] = "true"
		}
		return clashTrojanURI(proxy)
	case "http", "https":
		if protocol == "https" {
			proxy["type"] = "https"
			proxy["tls"] = "true"
		}
		return clashHTTPURI(proxy)
	case "socks", "socks5", "socks4", "socks4a":
		if protocol == "socks" {
			proxy["type"] = "socks5"
		}
		return clashSOCKSURI(proxy)
	default:
		return ""
	}
}

func parseSurgeProxyLine(line string) (string, []string, bool) {
	name, body, ok := strings.Cut(line, "=")
	if !ok {
		return "", nil, false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, false
	}
	fields := splitSurgeFields(body)
	if len(fields) == 0 {
		return "", nil, false
	}
	return name, fields, true
}

func splitSurgeFields(value string) []string {
	var fields []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, r := range value {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quote != 0 {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ',' {
			if field := strings.TrimSpace(current.String()); field != "" {
				fields = append(fields, field)
			}
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if field := strings.TrimSpace(current.String()); field != "" {
		fields = append(fields, field)
	}
	return fields
}

func surgePositionalsAndOptions(fields []string) ([]string, map[string]string) {
	positionals := make([]string, 0, len(fields))
	options := map[string]string{}
	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			positionals = append(positionals, strings.TrimSpace(field))
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			options[key] = value
		}
	}
	return positionals, options
}

func surgeFirstValue(options map[string]string, positionals []string, index int, keys ...string) string {
	if value := surgeOption(options, keys...); value != "" {
		return value
	}
	if index >= 0 && index < len(positionals) {
		return strings.TrimSpace(positionals[index])
	}
	return ""
}

func surgeOption(options map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(options[strings.ToLower(key)]); value != "" {
			return value
		}
	}
	return ""
}

func surgeCopyOptions(proxy map[string]string, options map[string]string, keys ...string) {
	for _, key := range keys {
		if value := surgeOption(options, key); value != "" {
			proxy[strings.ToLower(key)] = value
		}
	}
}

func surgeApplyTLSOptions(proxy map[string]string, options map[string]string) {
	if sni := surgeOption(options, "sni", "tls-host", "tls_host", "servername", "server_name"); sni != "" {
		proxy["sni"] = sni
	}
	if alpn := surgeOption(options, "alpn"); alpn != "" {
		proxy["alpn"] = alpn
	}
	if surgeBoolOption(options, "tls", "over-tls", "over_tls") {
		proxy["tls"] = "true"
	}
	if surgeBoolOption(options, "skip-cert-verify", "skip_cert_verify", "insecure") || strings.EqualFold(surgeOption(options, "tls-verification", "tls_verification"), "false") {
		proxy["insecure"] = "true"
	}
	if surgeBoolOption(options, "disable-sni", "disable_sni") {
		proxy["disable-sni"] = "true"
	}
}

func surgeApplyTransportOptions(proxy map[string]string, options map[string]string) {
	if surgeBoolOption(options, "ws", "websocket") {
		proxy["network"] = "ws"
	}
	if obfs := strings.ToLower(surgeOption(options, "obfs")); obfs == "ws" || obfs == "websocket" {
		proxy["network"] = "ws"
	}
	if path := surgeOption(options, "ws-path", "ws_path", "obfs-uri", "obfs_uri", "path"); path != "" {
		proxy["ws-path"] = path
	}
	if host := surgeOption(options, "ws-host", "ws_host", "obfs-host", "obfs_host", "host"); host != "" {
		proxy["ws-headers.host"] = host
	}
	if headerHost := surgeWSHeaderHost(surgeOption(options, "ws-headers", "ws_headers")); headerHost != "" {
		proxy["ws-headers.host"] = headerHost
	}
}

func surgeApplyNetworkOptions(proxy map[string]string, options map[string]string) {
	if network := surgeOption(options, "network", "protocol"); network != "" {
		proxy["network"] = network
	}
	if surgeBoolOption(options, "udp-relay", "udp_relay") && proxy["network"] == "" {
		proxy["network"] = "udp"
	}
}

func surgeBoolOption(options map[string]string, keys ...string) bool {
	switch strings.ToLower(surgeOption(options, keys...)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func surgeWSHeaderHost(value string) string {
	for _, item := range strings.FieldsFunc(value, func(r rune) bool {
		return r == '|' || r == ';'
	}) {
		key, host, ok := strings.Cut(item, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "host") {
			return strings.TrimSpace(host)
		}
	}
	return ""
}
