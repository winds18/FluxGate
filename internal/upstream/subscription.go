package upstream

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const maxSubscriptionBytes = 5 << 20

var supportedURIPrefixes = []string{
	"anytls://",
	"any-tls://",
	"blackhole://",
	"block://",
	"dns://",
	"direct://",
	"freedom://",
	"http://",
	"http+tls://",
	"http-tls://",
	"https://",
	"hy2://",
	"hysteria://",
	"hysteria2://",
	"naive://",
	"naive+https://",
	"naive+quic://",
	"naive-quic://",
	"reject://",
	"reject-drop://",
	"reject-no-drop://",
	"reject-tinygif://",
	"shadowtls://",
	"shadow-tls://",
	"shadowsocks://",
	"socks://",
	"socks4://",
	"socks4a://",
	"socks5://",
	"socks5h://",
	"ss://",
	"ssh://",
	"tor://",
	"trojan://",
	"trojan-go://",
	"tuic://",
	"vless://",
	"vmess://",
	"vmess-aead://",
	"wg://",
	"wireguard://",
}

type Fetcher struct {
	Client *http.Client
}

func (f Fetcher) Fetch(ctx context.Context, sourceURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid subscription url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported subscription url scheme")
	}

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("user-agent", "FluxGate")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("subscription returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSubscriptionBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) > maxSubscriptionBytes {
		return "", fmt.Errorf("subscription response exceeds %d bytes", maxSubscriptionBytes)
	}
	return string(body), nil
}

func NormalizeContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", errors.New("empty subscription content")
	}
	if normalized := URIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := SSDURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := SurgeProxyURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := QuantumultXURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := JSONURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := VMessJSONURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := ClashYAMLURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := SIP008URIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := SingBoxJSONURIList(content); normalized != "" {
		return normalized, nil
	}
	if normalized := V2RayJSONURIList(content); normalized != "" {
		return normalized, nil
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
		if err != nil {
			continue
		}
		if normalized := URIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := SSDURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := SurgeProxyURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := QuantumultXURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := JSONURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := VMessJSONURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := ClashYAMLURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := SIP008URIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := SingBoxJSONURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
		if normalized := V2RayJSONURIList(string(decoded)); normalized != "" {
			return normalized, nil
		}
	}
	return "", errors.New("subscription content does not contain supported node URIs")
}

func URIList(content string) string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		for _, prefix := range supportedURIPrefixes {
			if strings.HasPrefix(lower, prefix) {
				lines = append(lines, canonicalSpecialURI(line))
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

func canonicalSpecialURI(rawURI string) string {
	lower := strings.ToLower(strings.TrimSpace(rawURI))
	switch {
	case strings.HasPrefix(lower, "freedom://"):
		return rewriteURIScheme(rawURI, "direct")
	case strings.HasPrefix(lower, "blackhole://"),
		strings.HasPrefix(lower, "reject://"),
		strings.HasPrefix(lower, "reject-drop://"),
		strings.HasPrefix(lower, "reject-no-drop://"),
		strings.HasPrefix(lower, "reject-tinygif://"):
		return rewriteURIScheme(rawURI, "block")
	case strings.HasPrefix(lower, "shadowsocks://"):
		return rewriteURIScheme(rawURI, "ss")
	case strings.HasPrefix(lower, "http+tls://"),
		strings.HasPrefix(lower, "http-tls://"):
		return rewriteURIScheme(rawURI, "https")
	case strings.HasPrefix(lower, "hy2://"):
		return rewriteURIScheme(rawURI, "hysteria2")
	case strings.HasPrefix(lower, "any-tls://"):
		return rewriteURIScheme(rawURI, "anytls")
	case strings.HasPrefix(lower, "shadow-tls://"):
		return rewriteURIScheme(rawURI, "shadowtls")
	case strings.HasPrefix(lower, "naive-quic://"):
		return rewriteURIScheme(rawURI, "naive+quic")
	case strings.HasPrefix(lower, "socks5h://"):
		return rewriteURIScheme(rawURI, "socks5")
	case strings.HasPrefix(lower, "trojan-go://"):
		return rewriteURIScheme(rawURI, "trojan")
	case strings.HasPrefix(lower, "vmess-aead://"):
		return rewriteURIScheme(rawURI, "vmess")
	case strings.HasPrefix(lower, "wg://"):
		return rewriteURIScheme(rawURI, "wireguard")
	default:
		return rawURI
	}
}

func rewriteURIScheme(rawURI, scheme string) string {
	if index := strings.Index(rawURI, "://"); index >= 0 {
		return scheme + rawURI[index:]
	}
	return rawURI
}

func JSONURIList(content string) string {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.UseNumber()
	var doc any
	if err := decoder.Decode(&doc); err != nil {
		return ""
	}
	var uris []string
	collectJSONURIs("", doc, &uris)
	return URIList(strings.Join(uris, "\n"))
}

func collectJSONURIs(name string, value any, uris *[]string) {
	switch typed := value.(type) {
	case string:
		if normalized := normalizedStringURIList(typed); normalized != "" {
			*uris = append(*uris, applyJSONURIName(normalized, name))
		}
	case []any:
		for _, item := range typed {
			collectJSONURIs(name, item, uris)
		}
	case map[string]any:
		nodeName := firstJSONString(typed, "name", "displayName", "display_name", "display-name", "nodeName", "node_name", "node-name", "label", "title", "remarks", "remark", "tag", "ps", "id")
		if nodeName == "" {
			nodeName = name
		}
		if normalized := structuredJSONDocumentURI(typed); normalized != "" {
			*uris = append(*uris, normalized)
			return
		}
		vmessName := firstNonEmptyString(firstJSONString(typed, "ps", "name", "displayName", "display_name", "display-name", "nodeName", "node_name", "node-name", "label", "title", "remarks", "remark", "tag"), name)
		if uri := vmessJSONURI(typed, vmessName); uri != "" {
			*uris = append(*uris, uri)
			return
		}
		if uri := clashJSONProxyURI(typed, nodeName); uri != "" {
			*uris = append(*uris, uri)
			return
		}
		handled := map[string]bool{}
		for _, key := range []string{"uri", "url", "link", "share", "share_link", "shareLink", "share_url", "share-url", "shareUrl", "shareURL", "subscription_url", "subscription-url", "subscriptionUrl", "subscriptionURL", "node_url", "node-url", "nodeUrl", "nodeURL"} {
			handled[key] = true
			if item, ok := typed[key]; ok {
				collectJSONURIs(nodeName, item, uris)
			}
		}
		for _, key := range []string{"content", "raw", "raw_content", "raw-content", "rawContent", "subscription", "sub", "payload", "body", "text", "result", "response"} {
			handled[key] = true
			if item, ok := typed[key]; ok {
				collectJSONURIs("", item, uris)
			}
		}
		for _, key := range []string{"uris", "nodes", "proxies", "items", "servers", "subscriptions", "urls", "links", "data", "results", "payloads", "list", "nodeList", "node_list", "node-list", "proxyList", "proxy_list", "proxy-list", "serverList", "server_list", "server-list", "subscriptionList", "subscription_list", "subscription-list", "urlList", "url_list", "url-list", "linkList", "link_list", "link-list", "records", "rows", "entries"} {
			handled[key] = true
			if item, ok := typed[key]; ok {
				collectJSONURIs("", item, uris)
			}
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if handled[key] || isJSONURIMetadataKey(key) {
				continue
			}
			item := typed[key]
			collectJSONURIs(key, item, uris)
		}
	}
}

func structuredJSONDocumentURI(values map[string]any) string {
	if _, ok := values["outbounds"]; !ok {
		if _, ok := values["endpoints"]; !ok {
			return ""
		}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	content := string(raw)
	for _, normalize := range []func(string) string{
		SingBoxJSONURIList,
		V2RayJSONURIList,
	} {
		if normalized := normalize(content); normalized != "" {
			return normalized
		}
	}
	return ""
}

func clashJSONProxyURI(values map[string]any, fallbackName string) string {
	proxy := map[string]string{}
	collectJSONProxyFields(proxy, nil, values)
	applyClashJSONProxyAliases(proxy)
	if firstMapValue(proxy, "name") == "" {
		if name := firstNonEmptyString(firstMapValue(proxy, "tag", "remarks", "ps", "id"), fallbackName); name != "" {
			proxy["name"] = name
		}
	}
	return clashProxyURI(proxy)
}

func applyClashJSONProxyAliases(proxy map[string]string) {
	for _, item := range []struct {
		canonical string
		aliases   []string
	}{
		{canonical: "name", aliases: []string{"displayname", "display_name", "display-name", "nodename", "node_name", "node-name", "label", "title", "remarks", "remark", "tag", "ps"}},
		{canonical: "type", aliases: []string{"protocol", "proto", "scheme", "nodetype", "node_type", "node-type", "servertype", "server_type", "server-type", "protocoltype", "protocol_type", "protocol-type", "proxytype", "proxy_type", "proxy-type", "proxyprotocol", "proxy_protocol", "proxy-protocol"}},
		{canonical: "server", aliases: []string{"host", "hostname", "address", "addr", "add", "serveraddress", "server_address", "server-address", "serverhost", "server_host", "server-host", "remotehost", "remote_host", "remote-host", "nodehost", "node_host", "node-host", "endpoint", "endpointaddress", "endpoint_address", "endpoint-address"}},
		{canonical: "port", aliases: []string{"server_port", "serverport", "server-port", "remoteport", "remote_port", "remote-port", "nodeport", "node_port", "node-port", "portnumber", "port_number", "port-number"}},
		{canonical: "uuid", aliases: []string{"id", "user_id", "userid", "user-id"}},
		{canonical: "cipher", aliases: []string{"method", "encryption", "encryptmethod", "encrypt-method", "encrypt_method"}},
		{canonical: "password", aliases: []string{"pass", "passwd", "pwd", "psk", "token"}},
	} {
		if firstMapValue(proxy, item.canonical) != "" {
			continue
		}
		for _, alias := range item.aliases {
			if value := firstMapValue(proxy, alias); value != "" {
				proxy[item.canonical] = value
				if item.canonical == "type" || (item.canonical == "server" && strings.EqualFold(alias, "host")) {
					delete(proxy, strings.ToLower(alias))
				}
				break
			}
		}
	}
}

func collectJSONProxyFields(target map[string]string, scopes []string, values map[string]any) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		normalizedKey := normalizedStructuredProxyKey(strings.ToLower(strings.TrimSpace(key)))
		if normalizedKey == "" {
			continue
		}
		value := values[key]
		switch typed := value.(type) {
		case map[string]any:
			collectJSONProxyFields(target, append(scopes, normalizedKey), typed)
		default:
			if scalar := strings.Join(stringListFromAnyValue(value), ","); scalar != "" {
				storeJSONProxyField(target, scopes, normalizedKey, scalar)
				continue
			}
			if scalar := strings.TrimSpace(stringFromAnyValue(value)); scalar != "" {
				storeJSONProxyField(target, scopes, normalizedKey, scalar)
			}
		}
	}
}

func normalizedStructuredProxyKey(key string) string {
	switch key {
	case "allowinsecure":
		return "insecure"
	case "clientfingerprint":
		return "client-fingerprint"
	case "disablesni":
		return "disable-sni"
	case "earlydataheadername":
		return "early-data-header-name"
	case "grpcopts":
		return "grpc-opts"
	case "grpcservicename":
		return "grpc-service-name"
	case "httpopts":
		return "http-opts"
	case "httpupgradeopts":
		return "httpupgrade-opts"
	case "maxearlydata":
		return "max-early-data"
	case "idletimeout":
		return "idle-timeout"
	case "pingtimeout":
		return "ping-timeout"
	case "permitwithoutstream":
		return "permit-without-stream"
	case "multimode":
		return "multi-mode"
	case "pluginoptions":
		return "plugin-options"
	case "pluginopts":
		return "plugin-opts"
	case "privatekey":
		return "private-key"
	case "privatekeypassphrase":
		return "private-key-passphrase"
	case "privatekeypath":
		return "private-key-path"
	case "publickey":
		return "public-key"
	case "realityopts":
		return "reality-opts"
	case "skipcertverify":
		return "skip-cert-verify"
	case "servicename":
		return "service-name"
	case "wshost":
		return "ws-host"
	case "wsheaders":
		return "ws-headers"
	case "wsopts":
		return "ws-opts"
	case "wspath":
		return "ws-path"
	default:
		return key
	}
}

func storeJSONProxyField(target map[string]string, scopes []string, key, value string) {
	if len(scopes) > 0 {
		parts := make([]string, 0, len(scopes)+1)
		parts = append(parts, scopes...)
		parts = append(parts, key)
		target[strings.Join(parts, ".")] = value
	}
	if !shouldPromoteStructuredProxyField(scopes) {
		return
	}
	if len(scopes) > 0 && strings.TrimSpace(target[key]) != "" {
		return
	}
	target[key] = value
}

func shouldPromoteStructuredProxyField(scopes []string) bool {
	return len(scopes) == 0 || scopes[0] != "transport"
}

func firstJSONString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isJSONURIMetadataKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "name", "displayname", "display_name", "display-name", "nodename", "node_name", "node-name", "label", "title", "remarks", "remark", "tag", "ps", "id", "type", "protocol", "scheme":
		return true
	default:
		return false
	}
}

func applyJSONURIName(normalized, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return normalized
	}
	lines := strings.Split(normalized, "\n")
	for index, line := range lines {
		parsed, err := url.Parse(strings.TrimSpace(line))
		if err != nil || parsed.Scheme == "" || parsed.Fragment != "" {
			continue
		}
		parsed.Fragment = name
		lines[index] = parsed.String()
	}
	return strings.Join(lines, "\n")
}

func normalizedStringURIList(value string) string {
	for _, normalize := range []func(string) string{
		URIList,
		SSDURIList,
		SurgeProxyURIList,
		QuantumultXURIList,
		VMessJSONURIList,
		ClashYAMLURIList,
		SIP008URIList,
		SingBoxJSONURIList,
		V2RayJSONURIList,
	} {
		if normalized := normalize(value); normalized != "" {
			return normalized
		}
	}
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '\r', '\n', '\t', ' ':
			return -1
		default:
			return r
		}
	}, value)
	if compact == "" {
		return ""
	}
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(compact)
		if err != nil {
			continue
		}
		decodedContent := string(decoded)
		for _, normalize := range []func(string) string{
			URIList,
			SSDURIList,
			SurgeProxyURIList,
			QuantumultXURIList,
			VMessJSONURIList,
			ClashYAMLURIList,
			SIP008URIList,
			SingBoxJSONURIList,
			V2RayJSONURIList,
		} {
			if normalized := normalize(decodedContent); normalized != "" {
				return normalized
			}
		}
	}
	return ""
}
