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
	server := vmessJSONString(item, "add")
	port := vmessJSONPort(item, "port")
	uuid := vmessJSONString(item, "id")
	if server == "" || port == "" || uuid == "" {
		return ""
	}

	proxy := map[string]string{
		"name":        firstNonEmptyString(vmessJSONString(item, "ps"), fallbackName),
		"server":      server,
		"port":        port,
		"uuid":        uuid,
		"alterid":     vmessJSONString(item, "aid"),
		"cipher":      firstNonEmptyString(vmessJSONString(item, "scy"), vmessJSONString(item, "cipher")),
		"network":     firstNonEmptyString(vmessJSONString(item, "net"), vmessJSONString(item, "network")),
		"header-type": vmessJSONString(item, "type"),
		"host":        vmessJSONString(item, "host"),
		"path":        vmessJSONString(item, "path"),
		"sni":         firstNonEmptyString(vmessJSONString(item, "sni"), vmessJSONString(item, "serverName"), vmessJSONString(item, "server_name")),
		"alpn":        strings.Join(stringListFromAnyValue(item["alpn"]), ","),
	}
	if packetEncoding := firstNonEmptyString(vmessJSONString(item, "packetEncoding"), vmessJSONString(item, "packet_encoding"), vmessJSONString(item, "packet-encoding")); packetEncoding != "" {
		proxy["packet-encoding"] = packetEncoding
	}
	if vmessJSONTLSEnabled(item["tls"]) || strings.EqualFold(vmessJSONString(item, "security"), "tls") {
		proxy["tls"] = "true"
	}
	if boolFromAnyValue(item["allowInsecure"]) || boolFromAnyValue(item["allow_insecure"]) || boolFromAnyValue(item["skip-cert-verify"]) || boolFromAnyValue(item["skip_cert_verify"]) {
		proxy["insecure"] = "true"
	}
	return clashVMessURI(proxy)
}

func vmessJSONString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return strings.TrimSpace(stringFromAnyValue(values[key]))
}

func vmessJSONPort(values map[string]any, key string) string {
	port := intFromAnyValue(values[key])
	if port <= 0 {
		return ""
	}
	return strconv.Itoa(port)
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
