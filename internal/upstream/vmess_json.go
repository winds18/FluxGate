package upstream

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

func VMessJSONURIList(content string) string {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.UseNumber()
	var doc any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}
	var uris []string
	collectVMessJSONURIs("", doc, &uris)
	return strings.Join(uris, "\n")
}

func collectVMessJSONURIs(name string, value any, uris *[]string) {
	switch typed := value.(type) {
	case string:
		if normalized := VMessJSONURIList(typed); normalized != "" {
			*uris = append(*uris, normalized)
		}
	case []any:
		for index, item := range typed {
			fallbackName := name
			if fallbackName != "" && len(typed) > 1 {
				fallbackName = fallbackName + "-" + strconv.Itoa(index+1)
			}
			collectVMessJSONURIs(fallbackName, item, uris)
		}
	case map[string]any:
		if uri := vmessJSONURI(typed, name); uri != "" {
			*uris = append(*uris, uri)
			return
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectVMessJSONURIs(key, typed[key], uris)
		}
	}
}

func vmessJSONURI(item map[string]any, fallbackName string) string {
	if vmessJSONHasNonVMessProtocol(item) {
		return ""
	}
	server := vmessJSONString(item, "add")
	port := vmessJSONPort(item, "port")
	uuid := vmessJSONString(item, "id")
	if (server == "" || port == "" || uuid == "") && vmessJSONAllowsAliasCore(item) {
		server = firstNonEmptyString(server, vmessJSONString(item,
			"address",
			"addr",
			"server",
			"serverAddress",
			"server_address",
			"server-address",
			"serverHost",
			"server_host",
			"server-host",
			"remoteHost",
			"remote_host",
			"remote-host",
			"nodeHost",
			"node_host",
			"node-host",
			"hostname",
			"endpoint",
		))
		port = firstNonEmptyString(port, vmessJSONPort(item, "serverPort", "server_port", "server-port", "remotePort", "remote_port", "remote-port", "nodePort", "node_port", "node-port", "portNumber", "port_number", "port-number"))
		uuid = firstNonEmptyString(uuid, vmessJSONString(item, "uuid", "user_id", "userId", "userid", "user-id"))
	}
	if server == "" || port == "" || uuid == "" {
		return ""
	}

	proxy := map[string]string{
		"name":        firstNonEmptyString(vmessJSONString(item, "ps", "name", "displayName", "display_name", "display-name", "nodeName", "node_name", "node-name", "label", "title", "remarks", "remark", "tag"), fallbackName),
		"server":      server,
		"port":        port,
		"uuid":        uuid,
		"alterid":     vmessJSONString(item, "aid", "alterId", "alter_id", "alter-id"),
		"cipher":      firstNonEmptyString(vmessJSONString(item, "scy"), vmessJSONString(item, "cipher")),
		"network":     firstNonEmptyString(vmessJSONString(item, "net"), vmessJSONString(item, "network"), vmessJSONString(item, "transport")),
		"header-type": vmessJSONString(item, "type"),
		"host":        vmessJSONString(item, "host"),
		"path":        vmessJSONString(item, "path"),
		"sni":         firstNonEmptyString(vmessJSONString(item, "sni"), vmessJSONString(item, "serverName"), vmessJSONString(item, "server_name"), vmessJSONString(item, "server-name")),
		"alpn":        strings.Join(stringListFromAnyValue(vmessJSONValue(item, "alpn")), ","),
	}
	if packetEncoding := firstNonEmptyString(vmessJSONString(item, "packetEncoding"), vmessJSONString(item, "packet_encoding"), vmessJSONString(item, "packet-encoding")); packetEncoding != "" {
		proxy["packet-encoding"] = packetEncoding
	}
	if vmessJSONBool(item, "disable_sni", "disable-sni", "disableSNI", "disableSni") {
		proxy["disable_sni"] = "true"
	}
	if fingerprint := firstNonEmptyString(
		vmessJSONString(item, "fp"),
		vmessJSONString(item, "fingerprint"),
		vmessJSONString(item, "clientFingerprint"),
		vmessJSONString(item, "client_fingerprint"),
		vmessJSONString(item, "client-fingerprint"),
	); fingerprint != "" {
		proxy["fp"] = fingerprint
	}
	if vmessJSONTLSEnabled(vmessJSONValue(item, "tls")) || strings.EqualFold(vmessJSONString(item, "security"), "tls") {
		proxy["tls"] = "true"
	}
	if vmessJSONBool(item, "allowInsecure", "allow_insecure", "skip-cert-verify", "skip_cert_verify") {
		proxy["insecure"] = "true"
	}
	return clashVMessURI(proxy)
}

func vmessJSONValue(values map[string]any, keys ...string) any {
	if values == nil {
		return nil
	}
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	actualKeys := make([]string, 0, len(values))
	for key := range values {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)
	for _, key := range keys {
		normalizedKey := normalizedJSONURIKey(key)
		for _, actualKey := range actualKeys {
			if normalizedJSONURIKey(actualKey) == normalizedKey {
				return values[actualKey]
			}
		}
	}
	return nil
}

func vmessJSONString(values map[string]any, keys ...string) string {
	if values == nil {
		return ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(stringFromAnyValue(vmessJSONValue(values, key))); value != "" {
			return value
		}
	}
	return ""
}

func vmessJSONPort(values map[string]any, keys ...string) string {
	if values == nil {
		return ""
	}
	for _, key := range keys {
		port := intFromAnyValue(vmessJSONValue(values, key))
		if port > 0 {
			return strconv.Itoa(port)
		}
	}
	return ""
}

func vmessJSONBool(values map[string]any, keys ...string) bool {
	if values == nil {
		return false
	}
	for _, key := range keys {
		if boolFromAnyValue(vmessJSONValue(values, key)) {
			return true
		}
	}
	return false
}

func vmessJSONAllowsAliasCore(values map[string]any) bool {
	if strings.TrimSpace(stringFromAnyValue(vmessJSONValue(values, "v"))) != "" {
		return true
	}
	return vmessJSONAliasProtocol(values) == "vmess"
}

func vmessJSONHasNonVMessProtocol(values map[string]any) bool {
	for _, key := range []string{
		"protocol",
		"proto",
		"scheme",
		"nodeType",
		"node_type",
		"node-type",
		"serverType",
		"server_type",
		"server-type",
		"protocolType",
		"protocol_type",
		"protocol-type",
		"proxyType",
		"proxy_type",
		"proxy-type",
		"proxyProtocol",
		"proxy_protocol",
		"proxy-protocol",
	} {
		if protocol := normalizedVMessJSONProtocol(vmessJSONString(values, key)); protocol != "" && protocol != "vmess" {
			return true
		}
	}
	switch normalizedVMessJSONProtocol(vmessJSONString(values, "type")) {
	case "", "vmess", "none", "http":
		return false
	default:
		return true
	}
}

func vmessJSONAliasProtocol(values map[string]any) string {
	for _, key := range []string{
		"protocol",
		"proto",
		"scheme",
		"nodeType",
		"node_type",
		"node-type",
		"serverType",
		"server_type",
		"server-type",
		"protocolType",
		"protocol_type",
		"protocol-type",
		"proxyType",
		"proxy_type",
		"proxy-type",
		"proxyProtocol",
		"proxy_protocol",
		"proxy-protocol",
	} {
		if protocol := normalizedVMessJSONProtocol(vmessJSONString(values, key)); protocol != "" {
			return protocol
		}
	}
	return ""
}

func normalizedVMessJSONProtocol(value string) string {
	normalized := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(strings.TrimSpace(value)))
	if normalized == "vmessaead" {
		return "vmess"
	}
	return normalized
}

func vmessJSONTLSEnabled(value any) bool {
	if boolFromAnyValue(value) {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(stringFromAnyValue(value))) {
	case "tls", "xtls":
		return true
	default:
		return false
	}
}
