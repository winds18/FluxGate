package upstream

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxSubscriptionBytes = 5 << 20

var supportedURIPrefixes = []string{
	"anytls://",
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
	if normalized := ClashYAMLURIList(content); normalized != "" {
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
		if normalized := ClashYAMLURIList(string(decoded)); normalized != "" {
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
