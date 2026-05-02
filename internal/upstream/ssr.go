package upstream

import (
	"encoding/base64"
	"net"
	"net/url"
	"strings"
)

func canonicalSSRURI(rawURI string) string {
	rawURI = strings.TrimSpace(rawURI)
	index := strings.Index(rawURI, "://")
	if index < 0 {
		return ""
	}
	decoded := decodeSSRBase64(rawURI[index+3:])
	if decoded == "" {
		return ""
	}

	body, rawQuery, _ := strings.Cut(decoded, "/?")
	if body == decoded {
		body, rawQuery, _ = strings.Cut(decoded, "?")
	}
	parts := strings.SplitN(body, ":", 6)
	if len(parts) != 6 {
		return ""
	}

	server := strings.TrimSpace(parts[0])
	port := strings.TrimSpace(parts[1])
	method := strings.TrimSpace(parts[3])
	password := decodeSSRBase64(parts[5])
	if server == "" || port == "" || method == "" || password == "" {
		return ""
	}

	result := &url.URL{
		Scheme: "ss",
		User:   url.UserPassword(method, password),
		Host:   net.JoinHostPort(server, port),
	}
	if remarks := ssrQueryBase64Value(rawQuery, "remarks"); remarks != "" {
		result.Fragment = remarks
	}
	return result.String()
}

func ssrQueryBase64Value(rawQuery, key string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ""
	}
	return decodeSSRBase64(values.Get(key))
}

func decodeSSRBase64(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	candidates := []string{value}
	if strings.Contains(value, " ") {
		candidates = append(candidates, strings.ReplaceAll(value, " ", "+"))
	}
	for _, candidate := range candidates {
		for _, encoding := range []*base64.Encoding{
			base64.RawURLEncoding,
			base64.URLEncoding,
			base64.RawStdEncoding,
			base64.StdEncoding,
		} {
			if decoded, err := encoding.DecodeString(candidate); err == nil {
				return string(decoded)
			}
		}
		if remainder := len(candidate) % 4; remainder != 0 {
			padded := candidate + strings.Repeat("=", 4-remainder)
			for _, encoding := range []*base64.Encoding{
				base64.URLEncoding,
				base64.StdEncoding,
			} {
				if decoded, err := encoding.DecodeString(padded); err == nil {
					return string(decoded)
				}
			}
		}
	}
	return ""
}
