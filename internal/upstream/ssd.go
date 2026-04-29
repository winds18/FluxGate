package upstream

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func SSDURIList(content string) string {
	var uris []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "ssd://") {
			continue
		}
		if normalized := SSDJSONURIList(strings.TrimSpace(line[len("ssd://"):])); normalized != "" {
			uris = append(uris, normalized)
		}
	}
	if len(uris) > 0 {
		return strings.Join(uris, "\n")
	}
	return SSDJSONURIList(content)
}

func SSDJSONURIList(content string) string {
	decoded := decodeSSDContent(strings.TrimSpace(content))
	if decoded == "" {
		return ""
	}
	decoder := json.NewDecoder(strings.NewReader(decoded))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}
	servers := ssdServerList(doc["servers"])
	if len(servers) == 0 {
		return ""
	}

	var uris []string
	for _, server := range servers {
		if uri := ssdServerURI(doc, server); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func decodeSSDContent(content string) string {
	if content == "" {
		return ""
	}
	if strings.HasPrefix(content, "{") {
		return content
	}
	if strings.HasPrefix(strings.ToLower(content), "ssd://") {
		content = content[len("ssd://"):]
	}
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '\r', '\n', '\t', ' ':
			return -1
		default:
			return r
		}
	}, content)
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(compact)
		if err == nil {
			return string(decoded)
		}
	}
	return ""
}

func ssdServerList(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if server, ok := item.(map[string]any); ok {
				result = append(result, server)
			}
		}
		return result
	default:
		return nil
	}
}

func ssdServerURI(doc, server map[string]any) string {
	host := jsonFieldString(server, "server", "host")
	port := firstNonEmptyString(jsonFieldString(server, "port", "server_port", "serverPort"), jsonFieldString(doc, "port", "server_port", "serverPort"))
	method := firstNonEmptyString(jsonFieldString(server, "encryption", "method", "cipher"), jsonFieldString(doc, "encryption", "method", "cipher"))
	password := firstNonEmptyString(jsonFieldString(server, "password"), jsonFieldString(doc, "password"))
	if host == "" || port == "" || method == "" || password == "" {
		return ""
	}
	proxy := map[string]string{
		"name":     firstNonEmptyString(jsonFieldString(server, "remarks", "remark", "name", "id"), host),
		"server":   host,
		"port":     port,
		"method":   method,
		"password": password,
	}
	if plugin := jsonFieldString(server, "plugin"); plugin != "" {
		proxy["plugin"] = plugin
	}
	if pluginOpts := jsonFieldString(server, "plugin_options", "pluginOptions", "plugin_opts", "plugin-opts", "plugin-options"); pluginOpts != "" {
		proxy["plugin_opts"] = pluginOpts
	}
	if network := jsonFieldString(server, "network"); network != "" {
		proxy["network"] = network
	}
	return clashShadowsocksURI(proxy)
}
