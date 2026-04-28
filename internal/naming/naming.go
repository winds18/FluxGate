package naming

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

func AutoPrefix(sourceName, sourceURL, fallback string, sequence int) string {
	label := strings.TrimSpace(sourceName)
	if label == "" {
		label = hostnameLabel(sourceURL)
	}
	if label == "" {
		label = strings.TrimSpace(fallback)
	}
	if label == "" {
		label = "Source"
	}
	if sequence > 1 {
		label = label + "-" + intString(sequence)
	}
	return "[" + label + "] "
}

func NormalizeManualPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return ""
	}
	if strings.HasSuffix(prefix, " ") {
		return prefix
	}
	return prefix + " "
}

func DisplayName(prefix, rawName string) string {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		rawName = "Unnamed"
	}
	return prefix + rawName
}

func RawNameFromURI(rawURI string) string {
	rawURI = strings.TrimSpace(rawURI)
	if rawURI == "" {
		return "Unnamed"
	}

	if strings.HasPrefix(rawURI, "vmess://") {
		if name := vmessName(rawURI); name != "" {
			return name
		}
	}

	parsed, err := url.Parse(rawURI)
	if err == nil {
		if parsed.Fragment != "" {
			if fragment, err := url.QueryUnescape(parsed.Fragment); err == nil && strings.TrimSpace(fragment) != "" {
				return strings.TrimSpace(fragment)
			}
			return strings.TrimSpace(parsed.Fragment)
		}
		if parsed.Hostname() != "" {
			return parsed.Hostname()
		}
		if parsed.Scheme != "" {
			return parsed.Scheme + "-node"
		}
	}

	if len(rawURI) > 48 {
		return rawURI[:48]
	}
	return rawURI
}

func ProtocolFromURI(rawURI string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil || parsed.Scheme == "" {
		return "unknown"
	}
	return strings.ToLower(parsed.Scheme)
}

func ServerFromURI(rawURI string) (string, int) {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return "", 0
	}
	host := parsed.Hostname()
	port := 0
	if parsed.Port() != "" {
		for _, ch := range parsed.Port() {
			if ch < '0' || ch > '9' {
				return host, 0
			}
			port = port*10 + int(ch-'0')
		}
	}
	return host, port
}

func hostnameLabel(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return host
}

func vmessName(rawURI string) string {
	payload := strings.TrimPrefix(rawURI, "vmess://")
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(payload)
	}
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(payload)
	}
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(payload)
	}
	if err != nil {
		return ""
	}

	var doc struct {
		PS string `json:"ps"`
	}
	if err := json.Unmarshal(decoded, &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.PS)
}

func intString(value int) string {
	if value == 0 {
		return "0"
	}
	var chars [20]byte
	i := len(chars)
	for value > 0 {
		i--
		chars[i] = byte('0' + value%10)
		value /= 10
	}
	return string(chars[i:])
}
