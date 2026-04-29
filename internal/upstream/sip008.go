package upstream

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func SIP008URIList(content string) string {
	var doc struct {
		Version int              `json:"version"`
		Servers []map[string]any `json:"servers"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &doc); err != nil {
		return ""
	}
	if doc.Version != 1 || len(doc.Servers) == 0 {
		return ""
	}

	var uris []string
	for _, server := range doc.Servers {
		if uri := sip008ServerURI(server); uri != "" {
			uris = append(uris, uri)
		}
	}
	return strings.Join(uris, "\n")
}

func sip008ServerURI(server map[string]any) string {
	host := jsonFieldString(server, "server", "host")
	port := jsonFieldString(server, "server_port", "serverPort", "port")
	method := jsonFieldString(server, "method", "cipher")
	password := jsonFieldString(server, "password")
	if host == "" || port == "" || method == "" || password == "" {
		return ""
	}
	return (&url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(method, password),
		Host:     net.JoinHostPort(host, port),
		Fragment: firstNonEmptyString(jsonFieldString(server, "remarks", "name", "id"), host),
	}).String()
}

func jsonFieldString(fields map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := fields[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case float64:
			if typed == float64(int64(typed)) {
				return strconv.FormatInt(int64(typed), 10)
			}
			return strconv.FormatFloat(typed, 'f', -1, 64)
		case json.Number:
			return typed.String()
		}
	}
	return ""
}
