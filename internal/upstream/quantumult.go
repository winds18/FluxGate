package upstream

import (
	"net"
	"strings"
)

func QuantumultXURIList(content string) string {
	if strings.HasPrefix(strings.TrimSpace(content), "{") {
		return ""
	}
	var uris []string
	inServerSection := false
	sawSection := false
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			section := strings.ToLower(strings.TrimSpace(strings.Trim(line, "[]")))
			inServerSection = section == "server_local" || section == "server_remote" || section == "server"
			sawSection = true
			continue
		}
		if sawSection && !inServerSection {
			continue
		}
		if uri := quantumultXProxyURI(line); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func quantumultXProxyURI(line string) string {
	protocol, fields, ok := parseQuantumultXProxyLine(line)
	if !ok || len(fields) == 0 {
		return ""
	}
	positionals, options := surgePositionalsAndOptions(fields)
	server, port := quantumultXServerPort(positionals, options)
	if server == "" || port == "" {
		return ""
	}
	name := firstNonEmptyString(surgeOption(options, "tag", "name", "remarks", "ps"), server)
	proxy := map[string]string{
		"name":   name,
		"type":   protocol,
		"server": server,
		"port":   port,
	}
	surgeApplyTLSOptions(proxy, options)
	quantumultXApplyTransportOptions(proxy, options)
	surgeApplyNetworkOptions(proxy, options)

	switch protocol {
	case "ss", "shadowsocks":
		proxy["type"] = "ss"
		proxy["method"] = surgeFirstValue(options, positionals, 1, "method", "cipher", "encrypt-method")
		proxy["password"] = surgeFirstValue(options, positionals, 2, "password", "passwd", "pass")
		return clashShadowsocksURI(proxy)
	case "trojan":
		proxy["password"] = surgeFirstValue(options, positionals, 1, "password", "passwd", "pass")
		if proxy["tls"] == "" {
			proxy["tls"] = "true"
		}
		return clashTrojanURI(proxy)
	case "vless":
		proxy["uuid"] = surgeFirstValue(options, positionals, 1, "uuid", "id", "password", "passwd")
		if flow := surgeOption(options, "flow"); flow != "" {
			proxy["flow"] = flow
		}
		return clashVLESSURI(proxy)
	case "vmess", "vmess-aead":
		proxy["type"] = "vmess"
		proxy["uuid"] = surgeFirstValue(options, positionals, 1, "uuid", "id", "password", "passwd")
		proxy["cipher"] = surgeOption(options, "method", "cipher", "security")
		proxy["alterid"] = surgeOption(options, "alter-id", "alter_id", "alterid", "aid")
		return clashVMessURI(proxy)
	case "http", "https":
		proxy["username"] = surgeOption(options, "username", "user")
		proxy["password"] = surgeOption(options, "password", "passwd", "pass")
		if protocol == "https" {
			proxy["type"] = "https"
			proxy["tls"] = "true"
		}
		return clashHTTPURI(proxy)
	case "socks", "socks5":
		proxy["type"] = "socks5"
		proxy["username"] = surgeOption(options, "username", "user")
		proxy["password"] = surgeOption(options, "password", "passwd", "pass")
		return clashSOCKSURI(proxy)
	default:
		return ""
	}
}

func parseQuantumultXProxyLine(line string) (string, []string, bool) {
	protocol, body, ok := strings.Cut(line, "=")
	if !ok {
		return "", nil, false
	}
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	switch protocol {
	case "ss", "shadowsocks", "trojan", "vless", "vmess", "vmess-aead", "http", "https", "socks", "socks5":
	default:
		return "", nil, false
	}
	fields := splitSurgeFields(body)
	if len(fields) == 0 {
		return "", nil, false
	}
	return protocol, fields, true
}

func quantumultXServerPort(positionals []string, options map[string]string) (string, string) {
	server := surgeOption(options, "server", "host", "address")
	port := surgeOption(options, "port", "server-port", "server_port")
	if server != "" && port != "" {
		return server, port
	}
	if len(positionals) == 0 {
		return server, port
	}
	if parsedServer, parsedPort := splitQuantumultXHostPort(positionals[0]); parsedServer != "" && parsedPort != "" {
		return parsedServer, parsedPort
	}
	if server == "" {
		server = strings.TrimSpace(positionals[0])
	}
	if port == "" && len(positionals) > 1 {
		port = strings.TrimSpace(positionals[1])
	}
	return server, port
}

func splitQuantumultXHostPort(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if host, port, err := net.SplitHostPort(value); err == nil {
		return strings.Trim(host, "[]"), port
	}
	index := strings.LastIndex(value, ":")
	if index <= 0 || index >= len(value)-1 {
		return "", ""
	}
	host := strings.TrimSpace(value[:index])
	port := strings.TrimSpace(value[index+1:])
	if host == "" || port == "" || strings.Contains(host, ":") {
		return "", ""
	}
	return host, port
}

func quantumultXApplyTransportOptions(proxy map[string]string, options map[string]string) {
	obfs := strings.ToLower(surgeOption(options, "obfs", "obfs-type", "obfs_type"))
	if obfs == "ws" || obfs == "websocket" || obfs == "wss" {
		proxy["network"] = "ws"
	}
	if obfs == "wss" && proxy["tls"] == "" {
		proxy["tls"] = "true"
	}
	if path := surgeOption(options, "obfs-uri", "obfs_uri", "ws-path", "ws_path", "path"); path != "" {
		proxy["ws-path"] = path
	}
	if host := surgeOption(options, "obfs-host", "obfs_host", "ws-host", "ws_host", "host"); host != "" {
		proxy["ws-headers.host"] = host
	}
}
