package upstream

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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
	"juicity://",
	"naive://",
	"naive+https://",
	"naive-https://",
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
	"ssr://",
	"ssh://",
	"tor://",
	"trojan://",
	"trojan-go://",
	"tuic://",
	"tuic-v5://",
	"tuic5://",
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
				if canonical := canonicalSpecialURI(line); canonical != "" {
					lines = append(lines, canonical)
				}
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
	case strings.HasPrefix(lower, "ssr://"):
		return canonicalSSRURI(rawURI)
	case strings.HasPrefix(lower, "http+tls://"),
		strings.HasPrefix(lower, "http-tls://"):
		return rewriteURIScheme(rawURI, "https")
	case strings.HasPrefix(lower, "hy2://"):
		return rewriteURIScheme(rawURI, "hysteria2")
	case strings.HasPrefix(lower, "any-tls://"):
		return rewriteURIScheme(rawURI, "anytls")
	case strings.HasPrefix(lower, "shadow-tls://"):
		return rewriteURIScheme(rawURI, "shadowtls")
	case strings.HasPrefix(lower, "naive-https://"):
		return rewriteURIScheme(rawURI, "naive+https")
	case strings.HasPrefix(lower, "naive-quic://"):
		return rewriteURIScheme(rawURI, "naive+quic")
	case strings.HasPrefix(lower, "socks5h://"):
		return rewriteURIScheme(rawURI, "socks5")
	case strings.HasPrefix(lower, "trojan-go://"):
		return rewriteURIScheme(rawURI, "trojan")
	case strings.HasPrefix(lower, "vmess-aead://"):
		return rewriteURIScheme(rawURI, "vmess")
	case strings.HasPrefix(lower, "tuic-v5://"),
		strings.HasPrefix(lower, "tuic5://"):
		return rewriteURIScheme(rawURI, "tuic")
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
		collectJSONInlineProxyProviderAliases(typed, handled, uris)
		collectJSONURIKeyAliases(typed, handled, []string{"uri", "url", "link", "share", "share_link", "shareLink", "share_url", "share-url", "shareUrl", "shareURL", "subscription_url", "subscription-url", "subscriptionUrl", "subscriptionURL", "node_url", "node-url", "nodeUrl", "nodeURL"}, nodeName, uris)
		if nodeName != "" {
			collectJSONURIKeyAliases(typed, handled, []string{"subscribe_url", "subscribe-url", "subscribeUrl", "subscribeURL", "sub_url", "sub-url", "subUrl", "subURL", "download_url", "download-url", "downloadUrl", "downloadURL"}, nodeName, uris)
		} else {
			markJSONURIKeyAliases(typed, handled, []string{"subscribe_url", "subscribe-url", "subscribeUrl", "subscribeURL", "sub_url", "sub-url", "subUrl", "subURL", "download_url", "download-url", "downloadUrl", "downloadURL"})
		}
		collectJSONURIKeyAliases(typed, handled, []string{"content", "raw", "raw_content", "raw-content", "rawContent", "subscription", "sub", "payload", "body", "text", "result", "response"}, "", uris)
		collectJSONURIKeyAliases(typed, handled, []string{"uris", "nodes", "proxies", "items", "servers", "subscriptions", "urls", "links", "data", "results", "payloads", "list", "nodeList", "node_list", "node-list", "proxyList", "proxy_list", "proxy-list", "serverList", "server_list", "server-list", "subscriptionList", "subscription_list", "subscription-list", "urlList", "url_list", "url-list", "linkList", "link_list", "link-list", "records", "rows", "entries"}, "", uris)
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

func collectJSONInlineProxyProviderAliases(values map[string]any, handled map[string]bool, uris *[]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		switch normalizedJSONURIKey(key) {
		case "proxyprovider", "proxyproviders":
			handled[key] = true
			collectJSONInlineProxyProviders(values[key], uris)
		}
	}
}

func collectJSONInlineProxyProviders(value any, uris *[]string) {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			collectJSONInlineProxyProviders(item, uris)
		}
	case map[string]any:
		if collectJSONInlineProviderProxyLists(typed, uris) {
			return
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectJSONInlineProxyProviders(typed[key], uris)
		}
	}
}

func collectJSONInlineProviderProxyLists(values map[string]any, uris *[]string) bool {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	collected := false
	for _, alias := range []string{"proxies", "proxy_list", "proxies_list", "nodes", "node_list", "server_list", "items", "list", "entries"} {
		normalizedAlias := normalizedJSONURIKey(alias)
		for _, key := range keys {
			if normalizedJSONURIKey(key) != normalizedAlias {
				continue
			}
			collectJSONURIs("", values[key], uris)
			collected = true
		}
	}
	return collected
}

func collectJSONURIKeyAliases(values map[string]any, handled map[string]bool, aliases []string, name string, uris *[]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, alias := range aliases {
		normalizedAlias := normalizedJSONURIKey(alias)
		for _, key := range keys {
			if handled[key] || normalizedJSONURIKey(key) != normalizedAlias {
				continue
			}
			handled[key] = true
			collectJSONURIs(name, values[key], uris)
		}
	}
}

func markJSONURIKeyAliases(values map[string]any, handled map[string]bool, aliases []string) {
	aliasSet := make(map[string]bool, len(aliases))
	for _, alias := range aliases {
		aliasSet[normalizedJSONURIKey(alias)] = true
	}
	for key := range values {
		if aliasSet[normalizedJSONURIKey(key)] {
			handled[key] = true
		}
	}
}

func normalizedJSONURIKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

func structuredJSONDocumentURI(values map[string]any) string {
	if jsonFieldValue(values, "outbounds") == nil &&
		jsonFieldValue(values, "outbound") == nil &&
		jsonFieldValue(values, "endpoints") == nil &&
		jsonFieldValue(values, "endpoint") == nil {
		return ""
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
	normalizeStructuredProxyServerPort(proxy)
	normalizeStructuredProxyTLSFields(proxy)
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
		{canonical: "type", aliases: []string{"protocol", "proto", "scheme", "protocolname", "protocol_name", "protocol-name", "nodetype", "node_type", "node-type", "servertype", "server_type", "server-type", "protocoltype", "protocol_type", "protocol-type", "proxytype", "proxy_type", "proxy-type", "proxyprotocol", "proxy_protocol", "proxy-protocol"}},
		{canonical: "server", aliases: []string{"host", "hostname", "address", "addr", "add", "serveraddress", "server_address", "server-address", "serverhost", "server_host", "server-host", "remotehost", "remote_host", "remote-host", "nodehost", "node_host", "node-host", "endpoint", "endpointaddress", "endpoint_address", "endpoint-address"}},
		{canonical: "port", aliases: []string{"server_port", "serverport", "server-port", "remoteport", "remote_port", "remote-port", "nodeport", "node_port", "node-port", "portnumber", "port_number", "port-number"}},
		{canonical: "uuid", aliases: []string{"id", "user_id", "userid", "user-id"}},
		{canonical: "cipher", aliases: []string{"method", "encryption", "encryptmethod", "encrypt-method", "encrypt_method"}},
		{canonical: "username", aliases: []string{"user", "user_name", "user-name", "account", "accountname", "account_name", "account-name", "login", "loginname", "login_name", "login-name"}},
		{canonical: "password", aliases: []string{"pass", "passwd", "pwd", "psk", "token", "secret", "credential", "credentials", "accountpassword", "account_password", "account-password"}},
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

func normalizeStructuredProxyServerPort(proxy map[string]string) {
	for _, key := range []string{"server", "host", "address", "endpoint"} {
		raw := firstMapValue(proxy, key)
		if raw == "" {
			continue
		}
		host, port := splitStructuredProxyHostPort(raw)
		if host == "" || port == "" {
			continue
		}
		proxy["server"] = host
		if firstMapValue(proxy, "port") == "" {
			proxy["port"] = port
		}
		return
	}
}

func splitStructuredProxyHostPort(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if host, port, err := net.SplitHostPort(value); err == nil && strings.TrimSpace(host) != "" && strings.TrimSpace(port) != "" {
		return strings.Trim(host, "[]"), strings.TrimSpace(port)
	}
	if strings.ContainsAny(value, " \t\r\n/\\") {
		return "", ""
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
	return strings.Trim(host, "[]"), port
}

func normalizeStructuredProxyTLSFields(proxy map[string]string) {
	if firstMapValue(proxy, "tls") == "" && boolMapValue(proxy, "tls.enabled", "tls.enable") {
		proxy["tls"] = "true"
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
			if scalar := normalizedStructuredProxyValue(normalizedKey, value); scalar != "" {
				storeJSONProxyField(target, scopes, normalizedKey, scalar)
			}
		}
	}
}

func normalizedStructuredProxyValue(key string, value any) string {
	values := stringListFromAnyValue(value)
	if len(values) > 0 {
		if structuredProxyKeyPrefersFirstValue(key) {
			return strings.TrimSpace(values[0])
		}
		return strings.Join(values, ",")
	}
	return strings.TrimSpace(stringFromAnyValue(value))
}

func structuredProxyKeyPrefersFirstValue(key string) bool {
	switch key {
	case "servername", "short-id":
		return true
	default:
		return false
	}
}

func normalizedStructuredProxyKey(key string) string {
	switch key {
	case "allow_insecure":
		return "insecure"
	case "allow-insecure":
		return "insecure"
	case "allowinsecure":
		return "insecure"
	case "clientfingerprint":
		return "client-fingerprint"
	case "disablesni":
		return "disable-sni"
	case "earlydataheadername":
		return "early-data-header-name"
	case "enable_tls":
		return "tls"
	case "enable-tls":
		return "tls"
	case "enabletls":
		return "tls"
	case "flow_name":
		return "flow"
	case "flow-name":
		return "flow"
	case "flowname":
		return "flow"
	case "grpcopts":
		return "grpc-opts"
	case "grpc_settings":
		return "grpc-opts"
	case "grpc-settings":
		return "grpc-opts"
	case "grpcsettings":
		return "grpc-opts"
	case "grpcservicename":
		return "grpc-service-name"
	case "h2_settings":
		return "http-opts"
	case "h2-settings":
		return "http-opts"
	case "h2settings":
		return "http-opts"
	case "httpopts":
		return "http-opts"
	case "http_settings":
		return "http-opts"
	case "http-settings":
		return "http-opts"
	case "httpsettings":
		return "http-opts"
	case "http_upgrade_settings":
		return "httpupgrade-opts"
	case "http-upgrade-settings":
		return "httpupgrade-opts"
	case "httpupgrade_settings":
		return "httpupgrade-opts"
	case "httpupgrade-settings":
		return "httpupgrade-opts"
	case "httpupgradesettings":
		return "httpupgrade-opts"
	case "httpupgradeopts":
		return "httpupgrade-opts"
	case "maxearlydata":
		return "max-early-data"
	case "nettype":
		return "network"
	case "network_type":
		return "network"
	case "network-type":
		return "network"
	case "networktype":
		return "network"
	case "over_tls":
		return "tls"
	case "over-tls":
		return "tls"
	case "overtls":
		return "tls"
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
	case "quic_settings":
		return "quic-opts"
	case "quic-settings":
		return "quic-opts"
	case "quicsettings":
		return "quic-opts"
	case "realityopts":
		return "reality-opts"
	case "skipcertverify":
		return "skip-cert-verify"
	case "skip_certificate_verify":
		return "skip-cert-verify"
	case "skip-certificate-verify":
		return "skip-cert-verify"
	case "skipcertificateverify":
		return "skip-cert-verify"
	case "skip_verify":
		return "skip-cert-verify"
	case "skip-verify":
		return "skip-cert-verify"
	case "skipverify":
		return "skip-cert-verify"
	case "servicename":
		return "service-name"
	case "server-name":
		return "servername"
	case "server_names":
		return "servername"
	case "server-names":
		return "servername"
	case "servernames":
		return "servername"
	case "shortid":
		return "short-id"
	case "shortids":
		return "short-id"
	case "spiderx":
		return "spider-x"
	case "transport_type":
		return "transport"
	case "transport-type":
		return "transport"
	case "transporttype":
		return "transport"
	case "tcp_settings":
		return "tcp-opts"
	case "tcp-settings":
		return "tcp-opts"
	case "tcpsettings":
		return "tcp-opts"
	case "tls_host":
		return "servername"
	case "tls-host":
		return "servername"
	case "tlshost":
		return "servername"
	case "tls_server_name":
		return "servername"
	case "tls-server-name":
		return "servername"
	case "tlsservername":
		return "servername"
	case "tls_enable":
		return "tls"
	case "tls-enable":
		return "tls"
	case "tlsenable":
		return "tls"
	case "tls_enabled":
		return "tls"
	case "tls-enabled":
		return "tls"
	case "tlsenabled":
		return "tls"
	case "tls_settings":
		return "tls"
	case "tls-settings":
		return "tls"
	case "tlssettings":
		return "tls"
	case "tls_options":
		return "tls"
	case "tls-options":
		return "tls"
	case "tlsoptions":
		return "tls"
	case "tls_allow_insecure":
		return "insecure"
	case "tls-allow-insecure":
		return "insecure"
	case "tlsallowinsecure":
		return "insecure"
	case "tls_insecure":
		return "insecure"
	case "tls-insecure":
		return "insecure"
	case "tlsinsecure":
		return "insecure"
	case "tls_skip_verify":
		return "skip-cert-verify"
	case "tls-skip-verify":
		return "skip-cert-verify"
	case "tlsskipverify":
		return "skip-cert-verify"
	case "xtls_flow":
		return "flow"
	case "xtls-flow":
		return "flow"
	case "xtlsflow":
		return "flow"
	case "wshost":
		return "ws-host"
	case "wsheaders":
		return "ws-headers"
	case "wsopts":
		return "ws-opts"
	case "ws_settings":
		return "ws-opts"
	case "ws-settings":
		return "ws-opts"
	case "wssettings":
		return "ws-opts"
	case "web_socket_settings":
		return "ws-opts"
	case "web-socket-settings":
		return "ws-opts"
	case "websocketsettings":
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
	return len(scopes) == 0 || (scopes[0] != "transport" && scopes[0] != "tcp-opts" && scopes[0] != "quic-opts")
}

func firstJSONString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
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
			if normalizedJSONURIKey(actualKey) != normalizedKey {
				continue
			}
			if value, ok := values[actualKey].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func isJSONURIMetadataKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "name", "displayname", "display_name", "display-name", "nodename", "node_name", "node-name", "label", "title", "remarks", "remark", "tag", "ps", "id", "type", "protocol", "scheme",
		"success", "ok", "status", "code", "message", "msg", "error", "errors", "reason",
		"meta", "metadata", "pagination", "page", "page_size", "page-size", "pagesize", "pages",
		"total", "count", "limit", "offset", "next", "previous", "prev",
		"traceid", "trace_id", "trace-id", "requestid", "request_id", "request-id",
		"proxyprovider", "proxy_provider", "proxy-provider", "proxyproviders", "proxy_providers", "proxy-providers",
		"ruleprovider", "rule_provider", "rule-provider", "ruleproviders", "rule_providers", "rule-providers",
		"proxygroup", "proxy_group", "proxy-group", "proxygroups", "proxy_groups", "proxy-groups",
		"healthcheck", "health_check", "health-check", "rule", "rules":
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
