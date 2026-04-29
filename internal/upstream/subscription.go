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
	"block://",
	"dns://",
	"direct://",
	"http://",
	"https://",
	"hy2://",
	"hysteria://",
	"hysteria2://",
	"naive://",
	"naive+https://",
	"naive+quic://",
	"shadowtls://",
	"socks://",
	"socks4://",
	"socks4a://",
	"socks5://",
	"ss://",
	"ssh://",
	"tor://",
	"trojan://",
	"tuic://",
	"vless://",
	"vmess://",
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
				lines = append(lines, line)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
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
		nodeName := firstJSONString(typed, "name", "remarks", "tag", "ps", "id")
		if nodeName == "" {
			nodeName = name
		}
		if normalized := structuredJSONDocumentURI(typed); normalized != "" {
			*uris = append(*uris, normalized)
			return
		}
		vmessName := firstNonEmptyString(firstJSONString(typed, "ps", "name", "remarks", "tag"), name)
		if uri := vmessJSONURI(typed, vmessName); uri != "" {
			*uris = append(*uris, uri)
			return
		}
		if uri := clashJSONProxyURI(typed, nodeName); uri != "" {
			*uris = append(*uris, uri)
			return
		}
		handled := map[string]bool{}
		for _, key := range []string{"uri", "url", "link", "share", "share_link"} {
			handled[key] = true
			if item, ok := typed[key]; ok {
				collectJSONURIs(nodeName, item, uris)
			}
		}
		for _, key := range []string{"content", "raw", "raw_content", "rawContent", "subscription", "sub", "payload", "body", "text", "result", "response"} {
			handled[key] = true
			if item, ok := typed[key]; ok {
				collectJSONURIs("", item, uris)
			}
		}
		for _, key := range []string{"uris", "nodes", "proxies", "items", "servers", "subscriptions", "urls", "links", "data", "results", "payloads"} {
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
	if firstMapValue(proxy, "name") == "" {
		if name := firstNonEmptyString(firstMapValue(proxy, "tag", "remarks", "ps", "id"), fallbackName); name != "" {
			proxy["name"] = name
		}
	}
	return clashProxyURI(proxy)
}

func collectJSONProxyFields(target map[string]string, scopes []string, values map[string]any) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
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

func storeJSONProxyField(target map[string]string, scopes []string, key, value string) {
	if len(scopes) > 0 {
		parts := make([]string, 0, len(scopes)+1)
		parts = append(parts, scopes...)
		parts = append(parts, key)
		target[strings.Join(parts, ".")] = value
	}
	target[key] = value
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
	case "name", "remarks", "tag", "ps", "id", "type", "protocol":
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
