package upstream

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type singBoxSubscription struct {
	Outbounds []map[string]any `json:"outbounds"`
}

func SingBoxJSONURIList(content string) string {
	var doc singBoxSubscription
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &doc); err != nil {
		return ""
	}
	if len(doc.Outbounds) == 0 {
		return ""
	}

	var uris []string
	for _, outbound := range doc.Outbounds {
		if uri := singBoxOutboundURI(outbound); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func singBoxOutboundURI(outbound map[string]any) string {
	outboundType := strings.ToLower(singBoxString(outbound, "type"))
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
	if sni := singBoxTLSString(outbound, "server_name"); sni != "" {
		proxy["sni"] = sni
	}
	if singBoxTLSBool(outbound, "insecure") {
		proxy["insecure"] = "true"
	}
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
	if tlsMap(outbound) != nil {
		proxy["tls"] = "true"
	}
	if sni := singBoxTLSString(outbound, "server_name"); sni != "" {
		proxy["sni"] = sni
	}
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
	if tlsMap(outbound) != nil {
		doc["tls"] = "tls"
	}
	if transport, ok := outbound["transport"].(map[string]any); ok {
		if transportType := strings.TrimSpace(stringFromAnyValue(transport["type"])); transportType != "" {
			doc["net"] = transportType
		}
		doc["path"] = strings.TrimSpace(stringFromAnyValue(transport["path"]))
		if headers, ok := transport["headers"].(map[string]any); ok {
			doc["host"] = strings.TrimSpace(stringFromAnyValue(headers["Host"]))
			if doc["host"] == "" {
				doc["host"] = strings.TrimSpace(stringFromAnyValue(headers["host"]))
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
		if obfsType := strings.TrimSpace(stringFromAnyValue(obfs["type"])); obfsType != "" {
			values.Set("obfs", obfsType)
		}
		if obfsPassword := strings.TrimSpace(stringFromAnyValue(obfs["password"])); obfsPassword != "" {
			values.Set("obfs-password", obfsPassword)
		}
	}
	appendTLSQueryValues(outbound, values)
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
	if boolFromAnyValue(outbound["udp_over_stream"]) {
		values.Set("udp_over_stream", "1")
	}
	if boolFromAnyValue(outbound["zero_rtt_handshake"]) {
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

func singBoxName(outbound map[string]any) string {
	return firstNonEmptyString(singBoxString(outbound, "tag"), singBoxString(outbound, "name"))
}

func singBoxString(outbound map[string]any, key string) string {
	return strings.TrimSpace(stringFromAnyValue(outbound[key]))
}

func singBoxPort(outbound map[string]any) string {
	value := intFromAnyValue(outbound["server_port"])
	if value <= 0 {
		value = intFromAnyValue(outbound["port"])
	}
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func singBoxTLSString(outbound map[string]any, key string) string {
	tls := tlsMap(outbound)
	if tls == nil {
		return ""
	}
	return strings.TrimSpace(stringFromAnyValue(tls[key]))
}

func singBoxTLSBool(outbound map[string]any, key string) bool {
	tls := tlsMap(outbound)
	if tls == nil {
		return false
	}
	return boolFromAnyValue(tls[key])
}

func appendTLSQueryValues(outbound map[string]any, values url.Values) {
	tls := tlsMap(outbound)
	if tls == nil {
		return
	}
	if serverName := strings.TrimSpace(stringFromAnyValue(tls["server_name"])); serverName != "" {
		values.Set("sni", serverName)
	}
	if boolFromAnyValue(tls["insecure"]) {
		values.Set("insecure", "1")
	}
	if boolFromAnyValue(tls["disable_sni"]) {
		values.Set("disable_sni", "1")
	}
	if alpn := stringListFromAnyValue(tls["alpn"]); len(alpn) > 0 {
		values.Set("alpn", strings.Join(alpn, ","))
	}
}

func tlsMap(outbound map[string]any) map[string]any {
	tls, ok := outbound["tls"].(map[string]any)
	if !ok {
		return nil
	}
	if enabled, ok := tls["enabled"]; ok && !boolFromAnyValue(enabled) {
		return nil
	}
	return tls
}

func singBoxMap(outbound map[string]any, key string) map[string]any {
	value, ok := outbound[key].(map[string]any)
	if !ok {
		return nil
	}
	return value
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
