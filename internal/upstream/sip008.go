package upstream

import (
	"encoding/json"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func SIP008URIList(content string) string {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}
	servers := sip008ServerList(jsonFieldValue(doc, "servers"))
	if jsonFieldString(doc, "version") != "1" || len(servers) == 0 {
		return ""
	}

	var uris []string
	for _, server := range servers {
		if uri := sip008ServerURI(server); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func sip008ServerList(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		servers := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if server, ok := item.(map[string]any); ok {
				servers = append(servers, server)
			}
		}
		return servers
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		servers := make([]map[string]any, 0, len(keys))
		for _, key := range keys {
			server, ok := typed[key].(map[string]any)
			if !ok || len(server) == 0 {
				continue
			}
			if shadowsocksJSONName(server) == "" {
				server["remarks"] = key
			}
			servers = append(servers, server)
		}
		return servers
	default:
		return nil
	}
}

func sip008ServerURI(server map[string]any) string {
	host := shadowsocksJSONHost(server)
	port := shadowsocksJSONPort(server)
	method := shadowsocksJSONMethod(server)
	password := shadowsocksJSONPassword(server)
	if host == "" || port == "" || method == "" || password == "" {
		return ""
	}
	values := url.Values{}
	if plugin := jsonFieldString(server, "plugin"); plugin != "" {
		values.Set("plugin", plugin)
	}
	if pluginOpts := jsonFieldString(server, "plugin_opts", "pluginOpts", "plugin-options", "plugin_options"); pluginOpts != "" {
		values.Set("plugin_opts", pluginOpts)
	}
	if network := jsonFieldString(server, "network"); network != "" {
		values.Set("network", network)
	}
	result := &url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(method, password),
		Host:     net.JoinHostPort(host, port),
		Fragment: firstNonEmptyString(shadowsocksJSONName(server), host),
	}
	if len(values) > 0 {
		result.RawQuery = values.Encode()
	}
	return result.String()
}

func shadowsocksJSONName(fields map[string]any) string {
	return jsonFieldString(fields, "remarks", "remark", "name", "displayName", "display_name", "display-name", "nodeName", "node_name", "node-name", "label", "title", "tag", "id")
}

func shadowsocksJSONHost(fields map[string]any) string {
	return jsonFieldString(fields, "server", "host", "hostname", "address", "addr", "serverAddress", "server_address", "server-address", "serverHost", "server_host", "server-host", "remoteHost", "remote_host", "remote-host", "nodeHost", "node_host", "node-host", "endpoint")
}

func shadowsocksJSONPort(fields map[string]any) string {
	return jsonFieldString(fields, "server_port", "serverPort", "server-port", "port", "remotePort", "remote_port", "remote-port", "nodePort", "node_port", "node-port", "portNumber", "port_number", "port-number")
}

func shadowsocksJSONMethod(fields map[string]any) string {
	return jsonFieldString(fields, "method", "cipher", "encryption", "security", "encryptMethod", "encrypt_method", "encrypt-method")
}

func shadowsocksJSONPassword(fields map[string]any) string {
	return jsonFieldString(fields, "password", "pass", "passwd", "pwd", "psk", "token")
}

func jsonFieldString(fields map[string]any, keys ...string) string {
	if fields == nil {
		return ""
	}
	for _, key := range keys {
		value, ok := fields[key]
		if !ok {
			continue
		}
		if formatted := formatJSONFieldString(value); formatted != "" {
			return formatted
		}
	}
	actualKeys := make([]string, 0, len(fields))
	for key := range fields {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)
	for _, key := range keys {
		normalizedKey := normalizedJSONURIKey(key)
		for _, actualKey := range actualKeys {
			if normalizedJSONURIKey(actualKey) != normalizedKey {
				continue
			}
			if formatted := formatJSONFieldString(fields[actualKey]); formatted != "" {
				return formatted
			}
		}
	}
	return ""
}

func formatJSONFieldString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func jsonFieldValue(fields map[string]any, keys ...string) any {
	if fields == nil {
		return nil
	}
	for _, key := range keys {
		if value, ok := fields[key]; ok {
			return value
		}
	}
	actualKeys := make([]string, 0, len(fields))
	for key := range fields {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)
	for _, key := range keys {
		normalizedKey := normalizedJSONURIKey(key)
		for _, actualKey := range actualKeys {
			if normalizedJSONURIKey(actualKey) == normalizedKey {
				return fields[actualKey]
			}
		}
	}
	return nil
}
