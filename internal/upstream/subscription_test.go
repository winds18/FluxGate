package upstream

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNormalizeContentPlainURIList(t *testing.T) {
	got, err := NormalizeContent(" # comment\nvless://uuid@example.com:443#HK\n\n")
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "vless://uuid@example.com:443#HK" {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentSupportsCurrentOutboundURIList(t *testing.T) {
	raw := strings.Join([]string{
		"tuic://00000000-0000-0000-0000-000000000048:qa-placeholder@example.io:443#TUIC",
		"anytls://qa-placeholder@example.chat:443#AnyTLS",
		"block://default#Block",
		"dns://default#DNS",
		"direct://default#Direct",
		"shadowtls://qa-placeholder@example.help:443#ShadowTLS",
		"naive://qa-user:qa-placeholder@example.news:443#Naive",
		"hysteria://qa-placeholder@example.zone:443?auth_str=qa-auth#Hysteria",
		"https://qa-user:qa-placeholder@example.proxy:8443/connect#HTTP",
		"socks5://qa-user:qa-placeholder@example.socks:1080#SOCKS",
		"ssh://qa-user:qa-placeholder@example.ssh:22#SSH",
		"tor://default?executable_path=/usr/bin/tor&data_directory=cache%2Ftor#Tor",
		"wireguard://example.wg:51820?private_key=qa-private&peer_public_key=qa-peer&local_address=10.66.0.2/32#WireGuard",
	}, "\n")
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != raw {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentCanonicalizesSpecialOutboundURIList(t *testing.T) {
	raw := strings.Join([]string{
		"freedom://default#Freedom",
		"blackhole://default#Blackhole",
		"reject://default#Reject",
		"reject-drop://default#RejectDrop",
		"reject-no-drop://default#RejectNoDrop",
		"reject-tinygif://default#RejectTinygif",
	}, "\n")
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"direct://default#Freedom",
		"block://default#Blackhole",
		"block://default#Reject",
		"block://default#RejectDrop",
		"block://default#RejectNoDrop",
		"block://default#RejectTinygif",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentCanonicalizesShadowsocksSchemeAlias(t *testing.T) {
	raw := "shadowsocks://aes-128-gcm:qa-placeholder@ss-alias.example.test:8388#SS%20Alias"
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "ss://aes-128-gcm:qa-placeholder@ss-alias.example.test:8388#SS%20Alias"
	if got != want {
		t.Fatalf("unexpected canonical Shadowsocks URI: %q", got)
	}
}

func TestNormalizeContentCanonicalizesCommonDashedSchemeAliases(t *testing.T) {
	raw := strings.Join([]string{
		"any-tls://anytls-placeholder@anytls-alias.example.test:443#AnyTLS%20Alias",
		"shadow-tls://shadow-placeholder@shadowtls-alias.example.test:443?version=3#ShadowTLS%20Alias",
		"naive-https://qa-user:naive-placeholder@naive-https-alias.example.test:443#Naive%20HTTPS%20Alias",
		"naive-quic://qa-user:naive-placeholder@naive-alias.example.test:443#Naive%20QUIC%20Alias",
	}, "\n")
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"anytls://anytls-placeholder@anytls-alias.example.test:443#AnyTLS%20Alias",
		"shadowtls://shadow-placeholder@shadowtls-alias.example.test:443?version=3#ShadowTLS%20Alias",
		"naive+https://qa-user:naive-placeholder@naive-https-alias.example.test:443#Naive%20HTTPS%20Alias",
		"naive+quic://qa-user:naive-placeholder@naive-alias.example.test:443#Naive%20QUIC%20Alias",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected canonical dashed scheme aliases: %q", got)
	}
}

func TestNormalizeContentCanonicalizesHTTPTLSSchemeAliases(t *testing.T) {
	raw := strings.Join([]string{
		"http+tls://qa-user:http-placeholder@http-tls-alias.example.test:443/connect?sni=http-tls-alias.example.test#HTTP%20TLS%20Plus",
		"http-tls://qa-user:http-placeholder@http-tls-dash-alias.example.test:443/connect?sni=http-tls-dash-alias.example.test#HTTP%20TLS%20Dash",
	}, "\n")
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"https://qa-user:http-placeholder@http-tls-alias.example.test:443/connect?sni=http-tls-alias.example.test#HTTP%20TLS%20Plus",
		"https://qa-user:http-placeholder@http-tls-dash-alias.example.test:443/connect?sni=http-tls-dash-alias.example.test#HTTP%20TLS%20Dash",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected canonical HTTP TLS scheme aliases: %q", got)
	}
}

func TestNormalizeContentCanonicalizesCommonShortSchemeAliases(t *testing.T) {
	raw := strings.Join([]string{
		"hy2://hy2-placeholder@hy2-alias.example.test:443#Hy2%20Alias",
		"wg://wg-alias.example.test:51820?private_key=qa-private&peer_public_key=qa-peer&local_address=10.66.0.2/32#WG%20Alias",
	}, "\n")
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"hysteria2://hy2-placeholder@hy2-alias.example.test:443#Hy2%20Alias",
		"wireguard://wg-alias.example.test:51820?private_key=qa-private&peer_public_key=qa-peer&local_address=10.66.0.2/32#WG%20Alias",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected canonical short scheme aliases: %q", got)
	}
}

func TestNormalizeContentCanonicalizesSOCKS5HSchemeAlias(t *testing.T) {
	raw := "socks5h://qa-user:socks-placeholder@socks5h-alias.example.test:1080?udp=1#SOCKS5H%20Alias"
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "socks5://qa-user:socks-placeholder@socks5h-alias.example.test:1080?udp=1#SOCKS5H%20Alias"
	if got != want {
		t.Fatalf("unexpected canonical SOCKS5H URI: %q", got)
	}
}

func TestNormalizeContentCanonicalizesTrojanGoSchemeAlias(t *testing.T) {
	raw := "trojan-go://trojan-placeholder@trojan-go-alias.example.test:443?security=tls&type=ws&path=%2Fws&host=ws.trojan-go-alias.example.test#TrojanGo%20Alias"
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "trojan://trojan-placeholder@trojan-go-alias.example.test:443?security=tls&type=ws&path=%2Fws&host=ws.trojan-go-alias.example.test#TrojanGo%20Alias"
	if got != want {
		t.Fatalf("unexpected canonical Trojan-Go URI: %q", got)
	}
}

func TestNormalizeContentCanonicalizesVMessAEADSchemeAlias(t *testing.T) {
	raw := "vmess-aead://00000000-0000-0000-0000-000000000088@vmess-aead-alias.example.test:443?encryption=auto&security=tls&type=ws&path=%2Fvmess&host=ws.vmess-aead-alias.example.test#VMess%20AEAD%20Alias"
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "vmess://00000000-0000-0000-0000-000000000088@vmess-aead-alias.example.test:443?encryption=auto&security=tls&type=ws&path=%2Fvmess&host=ws.vmess-aead-alias.example.test#VMess%20AEAD%20Alias"
	if got != want {
		t.Fatalf("unexpected canonical VMess AEAD URI: %q", got)
	}
}

func TestNormalizeContentBase64URIList(t *testing.T) {
	raw := "vless://uuid@example.com:443#HK\nss://example#SG"
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	got, err := NormalizeContent(encoded)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != raw {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentJSONURIList(t *testing.T) {
	raw := `[
  "vless://00000000-0000-0000-0000-000000000080@example.com:443#香港 01",
  {
    "name": "东京 01",
    "uri": "trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京 01"
  }
]`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000080@example.com:443#香港 01",
		"trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京 01",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON URI list: %q", got)
	}
}

func TestNormalizeContentJSONURICollectionObject(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "link": "ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 01"
    }
  ],
  "links": [
    "hysteria2://qa-placeholder@example.dev:443?sni=hy2.example.dev#首尔 02"
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 01",
		"hysteria2://qa-placeholder@example.dev:443?sni=hy2.example.dev#首尔 02",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON URI collection: %q", got)
	}
}

func TestNormalizeContentJSONStructuredProxyObjects(t *testing.T) {
	raw := `{
  "proxies": [
    {
      "name": "JSON 香港 SS",
      "type": "ss",
      "server": "json-ss.example.test",
      "port": 8388,
      "cipher": "aes-128-gcm",
      "password": "qa-placeholder"
    },
    {
      "tag": "JSON 东京 Trojan",
      "type": "trojan",
      "server": "json-trojan.example.test",
      "port": 443,
      "password": "trojan-placeholder",
      "tls": true,
      "sni": "json-trojan.example.test",
      "network": "ws",
      "ws-opts": {
        "path": "/trojan",
        "headers": {
          "Host": "ws.json-trojan.example.test"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON proxy URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "ss://aes-128-gcm:qa-placeholder@json-ss.example.test:8388#JSON%20%E9%A6%99%E6%B8%AF%20SS" {
		t.Fatalf("unexpected structured JSON SS URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@json-trojan.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-trojan.example.test",
		"type=ws",
		"path=%2Ftrojan",
		"host=ws.json-trojan.example.test",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected structured JSON Trojan URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#JSON%20%E4%B8%9C%E4%BA%AC%20Trojan") {
		t.Fatalf("unexpected structured JSON Trojan fragment: %q", lines[1])
	}
}

func TestNormalizeContentJSONStructuredProxyAliases(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "remarks": "JSON 别名 SS",
      "protocol": "shadowsocks",
      "host": "json-alias-ss.example.test",
      "server_port": 8388,
      "method": "aes-128-gcm",
      "pass": "qa-placeholder"
    },
    {
      "name": "JSON 别名 VLESS",
      "protocol": "vless",
      "address": "json-alias-vless.example.test",
      "serverPort": "443",
      "id": "00000000-0000-0000-0000-000000000084",
      "tls": true,
      "servername": "json-alias-vless.example.test"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON alias proxy URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "ss://aes-128-gcm:qa-placeholder@json-alias-ss.example.test:8388#JSON%20%E5%88%AB%E5%90%8D%20SS" {
		t.Fatalf("unexpected structured JSON alias SS URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "vless://00000000-0000-0000-0000-000000000084@json-alias-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-alias-vless.example.test",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected structured JSON alias VLESS URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#JSON%20%E5%88%AB%E5%90%8D%20VLESS") {
		t.Fatalf("unexpected structured JSON alias VLESS fragment: %q", lines[1])
	}
}

func TestNormalizeContentJSONStructuredProviderAliases(t *testing.T) {
	raw := `{
  "servers": [
    {
      "displayName": "JSON Provider VLESS",
      "nodeType": "vless",
      "serverHost": "json-provider-vless.example.test",
      "remotePort": 443,
      "userId": "00000000-0000-0000-0000-000000000096",
      "tls": true,
      "serverName": "json-provider-vless.example.test"
    },
    {
      "label": "JSON Provider SS",
      "proxyProtocol": "shadowsocks",
      "remoteHost": "json-provider-ss.example.test",
      "remote_port": 8388,
      "encryptMethod": "aes-128-gcm",
      "passwd": "qa-placeholder"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON provider alias proxy URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000096@json-provider-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-provider-vless.example.test",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected structured JSON provider VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#JSON%20Provider%20VLESS") {
		t.Fatalf("unexpected structured JSON provider VLESS fragment: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:qa-placeholder@json-provider-ss.example.test:8388#JSON%20Provider%20SS" {
		t.Fatalf("unexpected structured JSON provider SS URI: %q", lines[1])
	}
}

func TestNormalizeContentJSONStructuredNameAliases(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "nodeName": "JSON NodeName VLESS",
      "protocol": "vless",
      "host": "json-nodename-vless.example.test",
      "server_port": 443,
      "user_id": "00000000-0000-0000-0000-000000000097",
      "tls": true,
      "servername": "json-nodename-vless.example.test"
    },
    {
      "remark": "JSON Remark SS",
      "scheme": "shadowsocks",
      "address": "json-remark-ss.example.test",
      "server-port": 8388,
      "method": "aes-128-gcm",
      "pwd": "qa-placeholder"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON name alias proxy URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000097@json-nodename-vless.example.test:443?")
	if !strings.Contains(lines[0], "sni=json-nodename-vless.example.test") ||
		!strings.HasSuffix(lines[0], "#JSON%20NodeName%20VLESS") {
		t.Fatalf("unexpected structured JSON nodeName alias VLESS URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:qa-placeholder@json-remark-ss.example.test:8388#JSON%20Remark%20SS" {
		t.Fatalf("unexpected structured JSON remark alias SS URI: %q", lines[1])
	}
}

func TestNormalizeContentJSONStructuredEndpointAliases(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "name": "JSON ProtocolType Trojan",
      "protocolType": "trojan",
      "nodeHost": "json-protocoltype-trojan.example.test",
      "nodePort": 443,
      "password": "trojan-placeholder",
      "tls": true,
      "sni": "json-protocoltype-trojan.example.test"
    },
    {
      "name": "JSON ServerType SS",
      "serverType": "shadowsocks",
      "endpoint": "json-servertype-ss.example.test",
      "portNumber": 8388,
      "encryption": "aes-128-gcm",
      "pass": "qa-placeholder"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON endpoint alias proxy URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@json-protocoltype-trojan.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=json-protocoltype-trojan.example.test") ||
		!strings.HasSuffix(lines[0], "#JSON%20ProtocolType%20Trojan") {
		t.Fatalf("unexpected structured JSON protocolType Trojan URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:qa-placeholder@json-servertype-ss.example.test:8388#JSON%20ServerType%20SS" {
		t.Fatalf("unexpected structured JSON serverType SS URI: %q", lines[1])
	}
}

func TestNormalizeContentJSONStructuredGenericTransport(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "name": "JSON Transport VLESS",
      "protocol": "vless",
      "address": "json-transport-vless.example.test",
      "serverPort": 443,
      "id": "00000000-0000-0000-0000-000000000095",
      "tls": true,
      "serverName": "json-transport-vless.example.test",
      "transport": {
        "type": "ws",
        "path": "/nested",
        "headers": {
          "Host": "ws.json-transport-vless.example.test"
        },
        "maxEarlyData": 512,
        "earlyDataHeaderName": "Sec-WebSocket-Protocol"
      }
    },
    {
      "name": "JSON Transport Trojan",
      "protocol": "trojan",
      "server": "json-transport-trojan.example.test",
      "port": 443,
      "password": "trojan-placeholder",
      "security": "tls",
      "sni": "json-transport-trojan.example.test",
      "transport": {
        "type": "grpc",
        "serviceName": "fluxgate-json",
        "idleTimeout": "30s",
        "pingTimeout": "10s",
        "permitWithoutStream": true,
        "multiMode": true
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON generic transport URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000095@json-transport-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-transport-vless.example.test",
		"type=ws",
		"path=%2Fnested",
		"host=ws.json-transport-vless.example.test",
		"max_early_data=512",
		"early_data_header_name=Sec-WebSocket-Protocol",
		"#JSON%20Transport%20VLESS",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected structured JSON generic WS URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@json-transport-trojan.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-transport-trojan.example.test",
		"type=grpc",
		"service_name=fluxgate-json",
		"idle_timeout=30s",
		"ping_timeout=10s",
		"permit_without_stream=true",
		"multi_mode=true",
		"#JSON%20Transport%20Trojan",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected structured JSON generic gRPC URI to contain %q: %q", want, lines[1])
		}
	}
}

func TestNormalizeContentJSONStructuredProxyHyphenAliases(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "remarks": "JSON hyphen SS",
      "protocol": "shadowsocks",
      "host": "json-hyphen-ss.example.test",
      "server-port": 8388,
      "encrypt-method": "aes-128-gcm",
      "passwd": "qa-placeholder"
    },
    {
      "name": "JSON short VLESS",
      "proto": "vless",
      "add": "json-short-vless.example.test",
      "server-port": "443",
      "user-id": "00000000-0000-0000-0000-000000000093",
      "tls": true,
      "servername": "json-short-vless.example.test"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 structured JSON hyphen alias proxy URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "ss://aes-128-gcm:qa-placeholder@json-hyphen-ss.example.test:8388#JSON%20hyphen%20SS" {
		t.Fatalf("unexpected structured JSON hyphen alias SS URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "vless://00000000-0000-0000-0000-000000000093@json-short-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=json-short-vless.example.test",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected structured JSON hyphen alias VLESS URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#JSON%20short%20VLESS") {
		t.Fatalf("unexpected structured JSON hyphen alias VLESS fragment: %q", lines[1])
	}
}

func TestNormalizeContentQuantumultXServerLocal(t *testing.T) {
	raw := `[server_local]
shadowsocks=qx-ss.example.test:8388, method=aes-128-gcm, password=qa-placeholder, obfs=wss, obfs-uri=/ss, obfs-host=ws.qx-ss.example.test, tag=香港 QuantumultX SS
hysteria2=qx-hy2.example.test:443, password=hy2-placeholder, obfs=salamander, obfs-password=obfs-placeholder, up-mbps=30, down-mbps=120, over-tls=true, tls-host=qx-hy2.example.test, tls-verification=false, disable-sni=true, alpn=h3, pinSHA256=hy2-pin-placeholder, tag=新加坡 QuantumultX Hysteria2
tuic=qx-tuic.example.test:443, uuid=00000000-0000-0000-0000-000000000087, password=tuic-placeholder, congestion-controller=bbr, udp-relay-mode=native, over-tls=true, tls-host=qx-tuic.example.test, tls-verification=false, alpn=h3, tag=大阪 QuantumultX TUIC
hysteria=qx-hysteria.example.test:443, auth-str=hysteria-auth, up-mbps=20, down-mbps=80, obfs=obfs-placeholder, recv-window-conn=1048576, recv-window=2097152, disable-mtu-discovery=true, protocol=udp, over-tls=true, tls-host=qx-hysteria.example.test, tls-verification=false, alpn=h3, tag=香港 QuantumultX Hysteria
anytls=qx-anytls.example.test:443, password=anytls-placeholder, idle-session-check-interval=20s, idle-session-timeout=45s, min-idle-session=2, over-tls=true, tls-host=qx-anytls.example.test, tls-verification=false, alpn="h2,http/1.1", tag=香港 QuantumultX AnyTLS
trojan=qx-trojan.example.test:443, password=trojan-placeholder, over-tls=true, tls-host=qx-trojan.example.test, tag=东京 QuantumultX Trojan
vless=qx-vless.example.test:443, password=00000000-0000-0000-0000-000000000085, over-tls=true, tls-host=qx-vless.example.test, obfs=ws, obfs-uri=/vless, obfs-host=ws.qx-vless.example.test, tag=首尔 QuantumultX VLESS
vmess=qx-vmess.example.test:443, password=00000000-0000-0000-0000-000000000086, method=auto, over-tls=true, tls-host=qx-vmess.example.test, obfs=wss, obfs-uri=/vmess, obfs-host=ws.qx-vmess.example.test, tag=大阪 QuantumultX VMess
shadowtls=qx-shadowtls.example.test:443, password=shadowtls-placeholder, version=3, tls-host=shadowtls.qx.example.test, tls-verification=false, alpn=h2, tag=香港 QuantumultX ShadowTLS
naive+quic=qx-naive.example.test:443, username=qa-user, password=naive-placeholder, tls-host=naive.qx.example.test, alpn=h3, quic-congestion-control=bbr, udp-over-tcp=true, insecure-concurrency=2, tag=新加坡 QuantumultX Naive
ssh=qx-ssh.example.test:22, username=qa-user, password=ssh-placeholder, private-key-path=keys/qa_id_ed25519, host-key-algorithms="ssh-ed25519,rsa-sha2-512", client-version=SSH-2.0-FluxGateQA, cipher=aes128-gcm@openssh.com, mac=hmac-sha2-256, kex-algorithm=curve25519-sha256, tag=香港 QuantumultX SSH
wireguard=qx-wg.example.test:51820, private-key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=, public-key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==, self-ip=10.66.0.4/32, self-ip-v6=fd00::4/128, pre-shared-key=cHNrLXBsYWNlaG9sZGVy, allowed-ips="0.0.0.0/0,::/0", reserved="7,8,9", mtu=1420, udp=true, interface-name=wg-qx, system-interface=true, tag=台北 QuantumultX WireGuard
http=qx-http.example.test:8080, qa-user, http-placeholder, tag=首尔 QuantumultX HTTP
socks5=qx-socks.example.test:1080, qa-user, socks-placeholder, protocol=udp, tag=大阪 QuantumultX SOCKS
[rewrite_local]
^https://example.test reject`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 14 {
		t.Fatalf("expected 14 Quantumult X proxy URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@qx-ss.example.test:8388?")
	for _, want := range []string{
		"network=ws",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20SS",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Quantumult X SS URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "hysteria2://hy2-placeholder@qx-hy2.example.test:443?")
	for _, want := range []string{
		"obfs=salamander",
		"obfs-password=obfs-placeholder",
		"up_mbps=30",
		"down_mbps=120",
		"sni=qx-hy2.example.test",
		"insecure=1",
		"disable_sni=1",
		"alpn=h3",
		"pinSHA256=hy2-pin-placeholder",
		"#%E6%96%B0%E5%8A%A0%E5%9D%A1%20QuantumultX%20Hysteria2",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Quantumult X Hysteria2 URI to contain %q: %q", want, lines[1])
		}
	}
	assertHasPrefix(t, lines[2], "tuic://00000000-0000-0000-0000-000000000087:tuic-placeholder@qx-tuic.example.test:443?")
	for _, want := range []string{
		"congestion_control=bbr",
		"udp_relay_mode=native",
		"sni=qx-tuic.example.test",
		"insecure=1",
		"alpn=h3",
		"#%E5%A4%A7%E9%98%AA%20QuantumultX%20TUIC",
	} {
		if !strings.Contains(lines[2], want) {
			t.Fatalf("expected Quantumult X TUIC URI to contain %q: %q", want, lines[2])
		}
	}
	assertHasPrefix(t, lines[3], "hysteria://hysteria-auth@qx-hysteria.example.test:443?")
	for _, want := range []string{
		"up_mbps=20",
		"down_mbps=80",
		"obfs=obfs-placeholder",
		"recv_window_conn=1048576",
		"recv_window=2097152",
		"disable_mtu_discovery=1",
		"network=udp",
		"sni=qx-hysteria.example.test",
		"insecure=1",
		"alpn=h3",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20Hysteria",
	} {
		if !strings.Contains(lines[3], want) {
			t.Fatalf("expected Quantumult X Hysteria URI to contain %q: %q", want, lines[3])
		}
	}
	assertHasPrefix(t, lines[4], "anytls://anytls-placeholder@qx-anytls.example.test:443?")
	for _, want := range []string{
		"idle_session_check_interval=20s",
		"idle_session_timeout=45s",
		"min_idle_session=2",
		"sni=qx-anytls.example.test",
		"insecure=1",
		"alpn=h2%2Chttp%2F1.1",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20AnyTLS",
	} {
		if !strings.Contains(lines[4], want) {
			t.Fatalf("expected Quantumult X AnyTLS URI to contain %q: %q", want, lines[4])
		}
	}
	assertHasPrefix(t, lines[5], "trojan://trojan-placeholder@qx-trojan.example.test:443?")
	if !strings.Contains(lines[5], "security=tls") || !strings.Contains(lines[5], "sni=qx-trojan.example.test") {
		t.Fatalf("unexpected Quantumult X Trojan URI: %q", lines[5])
	}
	assertHasPrefix(t, lines[6], "vless://00000000-0000-0000-0000-000000000085@qx-vless.example.test:443?")
	for _, want := range []string{"security=tls", "type=ws", "path=%2Fvless", "host=ws.qx-vless.example.test"} {
		if !strings.Contains(lines[6], want) {
			t.Fatalf("expected Quantumult X VLESS URI to contain %q: %q", want, lines[6])
		}
	}
	assertHasPrefix(t, lines[7], "vmess://")
	decodedVMessText := decodeVMessURIForTest(t, lines[7])
	for _, want := range []string{
		`"ps":"大阪 QuantumultX VMess"`,
		`"add":"qx-vmess.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000086"`,
		`"net":"ws"`,
		`"host":"ws.qx-vmess.example.test"`,
		`"path":"/vmess"`,
		`"tls":"tls"`,
		`"sni":"qx-vmess.example.test"`,
	} {
		if !strings.Contains(decodedVMessText, want) {
			t.Fatalf("expected Quantumult X VMess document to contain %q: %q", want, decodedVMessText)
		}
	}
	assertHasPrefix(t, lines[8], "shadowtls://shadowtls-placeholder@qx-shadowtls.example.test:443?")
	for _, want := range []string{
		"version=3",
		"sni=shadowtls.qx.example.test",
		"insecure=1",
		"alpn=h2",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20ShadowTLS",
	} {
		if !strings.Contains(lines[8], want) {
			t.Fatalf("expected Quantumult X ShadowTLS URI to contain %q: %q", want, lines[8])
		}
	}
	assertHasPrefix(t, lines[9], "naive+quic://qa-user:naive-placeholder@qx-naive.example.test:443?")
	for _, want := range []string{
		"quic=1",
		"quic_congestion_control=bbr",
		"udp_over_tcp=1",
		"insecure_concurrency=2",
		"sni=naive.qx.example.test",
		"alpn=h3",
		"#%E6%96%B0%E5%8A%A0%E5%9D%A1%20QuantumultX%20Naive",
	} {
		if !strings.Contains(lines[9], want) {
			t.Fatalf("expected Quantumult X Naive URI to contain %q: %q", want, lines[9])
		}
	}
	assertHasPrefix(t, lines[10], "ssh://qa-user:ssh-placeholder@qx-ssh.example.test:22?")
	for _, want := range []string{
		"private_key_path=keys%2Fqa_id_ed25519",
		"host_key_algorithms=ssh-ed25519%2Crsa-sha2-512",
		"client_version=SSH-2.0-FluxGateQA",
		"cipher=aes128-gcm%40openssh.com",
		"mac=hmac-sha2-256",
		"kex_algorithm=curve25519-sha256",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20SSH",
	} {
		if !strings.Contains(lines[10], want) {
			t.Fatalf("expected Quantumult X SSH URI to contain %q: %q", want, lines[10])
		}
	}
	assertHasPrefix(t, lines[11], "wireguard://qx-wg.example.test:51820?")
	for _, want := range []string{
		"private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI%3D",
		"peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg%3D%3D",
		"local_address=10.66.0.4%2F32%2Cfd00%3A%3A4%2F128",
		"allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0",
		"reserved=7%2C8%2C9",
		"network=udp",
		"interface_name=wg-qx",
		"system_interface=1",
		"#%E5%8F%B0%E5%8C%97%20QuantumultX%20WireGuard",
	} {
		if !strings.Contains(lines[11], want) {
			t.Fatalf("expected Quantumult X WireGuard URI to contain %q: %q", want, lines[11])
		}
	}
	assertHasPrefix(t, lines[12], "http://qa-user:http-placeholder@qx-http.example.test:8080")
	if !strings.HasSuffix(lines[12], "#%E9%A6%96%E5%B0%94%20QuantumultX%20HTTP") {
		t.Fatalf("unexpected Quantumult X HTTP URI: %q", lines[12])
	}
	assertHasPrefix(t, lines[13], "socks5://qa-user:socks-placeholder@qx-socks.example.test:1080?")
	if !strings.Contains(lines[13], "network=udp") ||
		!strings.HasSuffix(lines[13], "#%E5%A4%A7%E9%98%AA%20QuantumultX%20SOCKS") {
		t.Fatalf("unexpected Quantumult X SOCKS URI: %q", lines[13])
	}
}

func TestNormalizeContentQuantumultXGRPCTransport(t *testing.T) {
	raw := `[server_local]
vless=qx-grpc-vless.example.test:443, password=00000000-0000-0000-0000-000000000092, over-tls=true, tls-host=qx-grpc-vless.example.test, obfs=grpc, grpc-service-name=fluxgate-qx, tag=香港 QuantumultX gRPC VLESS
trojan=qx-grpc-trojan.example.test:443, password=trojan-placeholder, over-tls=true, tls-host=qx-grpc-trojan.example.test, obfs=grpc, service-name=trojan-qx, tag=东京 QuantumultX gRPC Trojan`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Quantumult X gRPC URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000092@qx-grpc-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=qx-grpc-vless.example.test",
		"type=grpc",
		"service_name=fluxgate-qx",
		"#%E9%A6%99%E6%B8%AF%20QuantumultX%20gRPC%20VLESS",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Quantumult X gRPC VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@qx-grpc-trojan.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=qx-grpc-trojan.example.test",
		"type=grpc",
		"service_name=trojan-qx",
		"#%E4%B8%9C%E4%BA%AC%20QuantumultX%20gRPC%20Trojan",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Quantumult X gRPC Trojan URI to contain %q: %q", want, lines[1])
		}
	}
}

func TestNormalizeContentQuantumultXSOCKSUDPFlag(t *testing.T) {
	raw := `[server_local]
socks5=qx-socks-udp.example.test:1080, qa-user, socks-placeholder, udp=true, tag=首尔 QuantumultX SOCKS UDP`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@qx-socks-udp.example.test:1080?")
	if !strings.Contains(got, "udp=1") || strings.Contains(got, "network=udp") {
		t.Fatalf("unexpected Quantumult X SOCKS UDP URI: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20QuantumultX%20SOCKS%20UDP") {
		t.Fatalf("unexpected Quantumult X SOCKS UDP fragment: %q", got)
	}
}

func TestNormalizeContentQuantumultXProtocolAliases(t *testing.T) {
	raw := `[server_local]
trojan-go=qx-trojan-go-alias.example.test:443, password=trojan-placeholder, over-tls=true, tls-host=qx-trojan-go-alias.example.test, tag=东京 QuantumultX Trojan-Go Alias
http+tls=qx-http-plus-tls-alias.example.test:443, qa-user, http-placeholder, tls-host=qx-http-plus-tls-alias.example.test, tag=香港 QuantumultX HTTP Plus TLS Alias
http-tls=qx-http-dash-tls-alias.example.test:443, username=qa-user, password=http-dash-placeholder, tls-host=qx-http-dash-tls-alias.example.test, tag=新加坡 QuantumultX HTTP Dash TLS Alias
naive-https=qx-naive-https-alias.example.test:443, username=qa-user, password=naive-https-placeholder, tls-host=qx-naive-https-alias.example.test, tag=台北 QuantumultX Naive HTTPS Alias
socks5h=qx-socks5h-alias.example.test:1080, qa-user, socks-placeholder, udp=true, tag=首尔 QuantumultX SOCKS5H Alias`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 Quantumult X alias URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@qx-trojan-go-alias.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=qx-trojan-go-alias.example.test") ||
		!strings.HasSuffix(lines[0], "#%E4%B8%9C%E4%BA%AC%20QuantumultX%20Trojan-Go%20Alias") {
		t.Fatalf("unexpected Quantumult X Trojan-Go alias URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "https://qa-user:http-placeholder@qx-http-plus-tls-alias.example.test:443?")
	if !strings.Contains(lines[1], "sni=qx-http-plus-tls-alias.example.test") ||
		!strings.HasSuffix(lines[1], "#%E9%A6%99%E6%B8%AF%20QuantumultX%20HTTP%20Plus%20TLS%20Alias") {
		t.Fatalf("unexpected Quantumult X HTTP Plus TLS alias URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "https://qa-user:http-dash-placeholder@qx-http-dash-tls-alias.example.test:443?")
	if !strings.Contains(lines[2], "sni=qx-http-dash-tls-alias.example.test") ||
		!strings.HasSuffix(lines[2], "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20QuantumultX%20HTTP%20Dash%20TLS%20Alias") {
		t.Fatalf("unexpected Quantumult X HTTP Dash TLS alias URI: %q", lines[2])
	}
	assertHasPrefix(t, lines[3], "naive://qa-user:naive-https-placeholder@qx-naive-https-alias.example.test:443?")
	if !strings.Contains(lines[3], "sni=qx-naive-https-alias.example.test") ||
		!strings.HasSuffix(lines[3], "#%E5%8F%B0%E5%8C%97%20QuantumultX%20Naive%20HTTPS%20Alias") {
		t.Fatalf("unexpected Quantumult X Naive HTTPS alias URI: %q", lines[3])
	}
	assertHasPrefix(t, lines[4], "socks5://qa-user:socks-placeholder@qx-socks5h-alias.example.test:1080?")
	if !strings.Contains(lines[4], "udp=1") ||
		!strings.HasSuffix(lines[4], "#%E9%A6%96%E5%B0%94%20QuantumultX%20SOCKS5H%20Alias") {
		t.Fatalf("unexpected Quantumult X SOCKS5H alias URI: %q", lines[4])
	}
}

func TestNormalizeContentJSONWrappedBase64URIList(t *testing.T) {
	payload := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000081@example.com:443#香港 02",
		"trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京 02",
	}, "\n")
	raw := `{"subscription":"` + base64.StdEncoding.EncodeToString([]byte(payload)) + `"}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != payload {
		t.Fatalf("unexpected JSON wrapped base64 URI list: %q", got)
	}
}

func TestNormalizeContentJSONWrappedRawContentURIList(t *testing.T) {
	raw := `{
  "data": {
    "raw_content": "ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 03\nhysteria2://qa-placeholder@example.dev:443?sni=hy2.example.dev#首尔 04"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 03",
		"hysteria2://qa-placeholder@example.dev:443?sni=hy2.example.dev#首尔 04",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON wrapped raw content URI list: %q", got)
	}
}

func TestNormalizeContentJSONCommonWrapperFields(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000083@example.com:443",
		"ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 05",
	}, "\n")))
	raw := `{
  "code": 0,
  "message": "ok",
  "result": {
    "payload": "` + payload + `",
    "body": {
      "links": [
        "trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京 05"
      ]
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000083@example.com:443",
		"ss://aes-128-gcm:qa-placeholder@example.net:8388#首尔 05",
		"trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#东京 05",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON common wrapper fields: %q", got)
	}
}

func TestNormalizeContentJSONIgnoresMetadataURLs(t *testing.T) {
	raw := `{
  "success": true,
  "message": "https://status.example.test/not-a-proxy",
  "meta": {
    "docs": "https://docs.example.test/subscription-api",
    "trace_id": "trace-placeholder"
  },
  "pagination": {
    "next": "https://api.example.test/subscriptions?page=2",
    "total": 1
  },
  "errors": [
    {
      "message": "https://errors.example.test/help"
    }
  ],
  "data": {
    "links": [
      "vless://00000000-0000-0000-0000-000000000098@metadata-filter.example.test:443"
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "vless://00000000-0000-0000-0000-000000000098@metadata-filter.example.test:443"
	if got != want {
		t.Fatalf("unexpected JSON metadata filtered URIs: %q", got)
	}
}

func TestNormalizeContentJSONIgnoresClashProviderMetadataURLs(t *testing.T) {
	raw := `{
  "proxies": [
    {
      "name": "JSON Provider Trojan",
      "type": "trojan",
      "server": "provider-node.example.test",
      "port": 443,
      "password": "trojan-placeholder",
      "tls": true,
      "sni": "provider-node.example.test"
    }
  ],
  "proxy-providers": {
    "airport": {
      "type": "http",
      "url": "https://provider-download.example.test/clash.yaml",
      "path": "./providers/airport.yaml",
      "health-check": {
        "enable": true,
        "url": "http://www.gstatic.com/generate_204"
      }
    }
  },
  "proxy-groups": [
    {
      "name": "auto",
      "type": "url-test",
      "proxies": ["JSON Provider Trojan"],
      "url": "http://www.gstatic.com/generate_204"
    }
  ],
  "rule-providers": {
    "reject-list": {
      "type": "http",
      "url": "https://rules.example.test/reject.yaml"
    }
  },
  "rules": [
    "DOMAIN-SUFFIX,example.test,DIRECT"
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@provider-node.example.test:443?")
	for _, unwanted := range []string{
		"provider-download.example.test",
		"gstatic.com",
		"rules.example.test",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("unexpected provider metadata URL imported as node: %q", got)
		}
	}
	if !strings.HasSuffix(got, "#JSON%20Provider%20Trojan") {
		t.Fatalf("unexpected provider Trojan fragment: %q", got)
	}
}

func TestNormalizeContentJSONParsesInlineProxyProviderNodes(t *testing.T) {
	raw := `{
  "proxy-providers": {
    "inline-airport": {
      "type": "inline",
      "url": "https://provider-download.example.test/should-not-import.yaml",
      "health-check": {
        "enable": true,
        "url": "http://www.gstatic.com/generate_204"
      },
      "proxies": [
        {
          "name": "JSON Inline Provider Trojan",
          "type": "trojan",
          "server": "inline-provider-trojan.example.test",
          "port": 443,
          "password": "trojan-placeholder",
          "tls": true,
          "sni": "inline-provider-trojan.example.test"
        },
        {
          "name": "JSON Inline Provider VLESS",
          "type": "vless",
          "server": "inline-provider-vless.example.test",
          "port": 8443,
          "uuid": "00000000-0000-0000-0000-000000000109",
          "flow": "xtls-rprx-vision"
        }
      ]
    }
  },
  "rule-providers": {
    "reject-list": {
      "type": "http",
      "url": "https://rules.example.test/reject.yaml"
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 inline provider URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@inline-provider-trojan.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=inline-provider-trojan.example.test") ||
		!strings.HasSuffix(lines[0], "#JSON%20Inline%20Provider%20Trojan") {
		t.Fatalf("unexpected inline provider Trojan URI: %q", lines[0])
	}
	if lines[1] != "vless://00000000-0000-0000-0000-000000000109@inline-provider-vless.example.test:8443?flow=xtls-rprx-vision#JSON%20Inline%20Provider%20VLESS" {
		t.Fatalf("unexpected inline provider VLESS URI: %q", lines[1])
	}
	for _, unwanted := range []string{
		"provider-download.example.test",
		"gstatic.com",
		"rules.example.test",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("unexpected provider metadata URL imported as node: %q", got)
		}
	}
}

func TestNormalizeContentJSONParsesInlineProxyProviderListAliases(t *testing.T) {
	raw := `{
  "proxyProviders": {
    "node-list-provider": {
      "type": "inline",
      "url": "https://provider-download.example.test/should-not-import.yaml",
      "nodeList": [
        {
          "name": "JSON Provider NodeList SS",
          "type": "ss",
          "server": "provider-nodelist-ss.example.test",
          "port": 8388,
          "cipher": "aes-128-gcm",
          "password": "qa-placeholder"
        }
      ]
    },
    "server-list-provider": {
      "type": "inline",
      "serverList": [
        {
          "name": "JSON Provider ServerList Trojan",
          "type": "trojan",
          "server": "provider-serverlist-trojan.example.test",
          "port": 443,
          "password": "trojan-placeholder"
        }
      ]
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 inline provider alias URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "ss://aes-128-gcm:qa-placeholder@provider-nodelist-ss.example.test:8388#JSON%20Provider%20NodeList%20SS" {
		t.Fatalf("unexpected provider nodeList Shadowsocks URI: %q", lines[0])
	}
	if lines[1] != "trojan://trojan-placeholder@provider-serverlist-trojan.example.test:443#JSON%20Provider%20ServerList%20Trojan" {
		t.Fatalf("unexpected provider serverList Trojan URI: %q", lines[1])
	}
	if strings.Contains(got, "provider-download.example.test") {
		t.Fatalf("unexpected provider download URL imported as node: %q", got)
	}
}

func TestNormalizeContentJSONCommonCollectionFields(t *testing.T) {
	raw := `{
  "data": {
    "list": [
      "vless://00000000-0000-0000-0000-000000000089@list.example.test:443"
    ],
    "records": [
      "trojan://trojan-placeholder@records.example.test:443?security=tls"
    ],
    "rows": [
      "ss://aes-128-gcm:qa-placeholder@rows.example.test:8388"
    ],
    "entries": [
      {
        "name": "香港 entries",
        "uri": "hysteria2://hy2-placeholder@entries.example.test:443"
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000089@list.example.test:443",
		"trojan://trojan-placeholder@records.example.test:443?security=tls",
		"ss://aes-128-gcm:qa-placeholder@rows.example.test:8388",
		"hysteria2://hy2-placeholder@entries.example.test:443#%E9%A6%99%E6%B8%AF%20entries",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON common collection fields: %q", got)
	}
}

func TestNormalizeContentJSONCommonCollectionAliasFields(t *testing.T) {
	raw := `{
  "data": {
    "nodeList": [
      "vless://00000000-0000-0000-0000-000000000092@node-list.example.test:443"
    ],
    "proxy_list": [
      "trojan://trojan-placeholder@proxy-list.example.test:443?security=tls"
    ],
    "serverList": [
      "ss://aes-128-gcm:qa-placeholder@server-list.example.test:8388"
    ],
    "subscription_list": [
      {
        "name": "香港 subscription_list",
        "uri": "hysteria2://hy2-placeholder@subscription-list.example.test:443"
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000092@node-list.example.test:443",
		"trojan://trojan-placeholder@proxy-list.example.test:443?security=tls",
		"ss://aes-128-gcm:qa-placeholder@server-list.example.test:8388",
		"hysteria2://hy2-placeholder@subscription-list.example.test:443#%E9%A6%99%E6%B8%AF%20subscription_list",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON common collection alias fields: %q", got)
	}
}

func TestNormalizeContentJSONWrapperFieldCaseAliases(t *testing.T) {
	raw := `{
  "Data": {
    "NodeList": [
      "vless://00000000-0000-0000-0000-000000000099@case-node-list.example.test:443"
    ],
    "Proxy_List": [
      {
        "Name": "Case Alias SS",
        "Type": "ss",
        "Server": "case-proxy-list.example.test",
        "Port": 8388,
        "Cipher": "aes-128-gcm",
        "Password": "qa-placeholder"
      }
    ],
    "Payload": {
      "LinkList": [
        "trojan://trojan-placeholder@case-link-list.example.test:443?security=tls"
      ]
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"trojan://trojan-placeholder@case-link-list.example.test:443?security=tls",
		"vless://00000000-0000-0000-0000-000000000099@case-node-list.example.test:443",
		"ss://aes-128-gcm:qa-placeholder@case-proxy-list.example.test:8388#Case%20Alias%20SS",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON wrapper field case aliases: %q", got)
	}
}

func TestNormalizeContentJSONNameFieldCaseAliases(t *testing.T) {
	raw := `{
  "Items": [
    {
      "Name": "Case Name VLESS",
      "URL": "vless://00000000-0000-0000-0000-000000000100@case-name.example.test:443"
    },
    {
      "Display_Name": "Case Display SS",
      "ShareURL": "ss://aes-128-gcm:qa-placeholder@case-display.example.test:8388"
    },
    {
      "NodeName": "Case Node Trojan",
      "Subscription_URL": "trojan://trojan-placeholder@case-node.example.test:443?security=tls"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000100@case-name.example.test:443#Case%20Name%20VLESS",
		"ss://aes-128-gcm:qa-placeholder@case-display.example.test:8388#Case%20Display%20SS",
		"trojan://trojan-placeholder@case-node.example.test:443?security=tls#Case%20Node%20Trojan",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON name field case aliases: %q", got)
	}
}

func TestNormalizeContentJSONCommonShareLinkFields(t *testing.T) {
	raw := `{
  "items": [
    {
      "name": "香港 shareUrl",
      "shareUrl": "vless://00000000-0000-0000-0000-000000000091@share-url.example.test:443"
    },
    {
      "remarks": "东京 subscriptionUrl",
      "subscriptionUrl": "trojan://trojan-placeholder@subscription-url.example.test:443?security=tls"
    },
    {
      "tag": "首尔 nodeUrl",
      "node_url": "ss://aes-128-gcm:qa-placeholder@node-url.example.test:8388"
    }
  ],
  "data": {
    "shareLink": "hysteria2://hy2-placeholder@share-link.example.test:443"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000091@share-url.example.test:443#%E9%A6%99%E6%B8%AF%20shareUrl",
		"trojan://trojan-placeholder@subscription-url.example.test:443?security=tls#%E4%B8%9C%E4%BA%AC%20subscriptionUrl",
		"ss://aes-128-gcm:qa-placeholder@node-url.example.test:8388#%E9%A6%96%E5%B0%94%20nodeUrl",
		"hysteria2://hy2-placeholder@share-link.example.test:443",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON common share link fields: %q", got)
	}
}

func TestNormalizeContentJSONCommonHyphenAliasFields(t *testing.T) {
	raw := `{
  "data": {
    "raw-content": "ss://aes-128-gcm:qa-placeholder@raw-content.example.test:8388",
    "items": [
      {
        "name": "香港 share-url",
        "share-url": "vless://00000000-0000-0000-0000-000000000095@share-hyphen.example.test:443"
      },
      {
        "remarks": "东京 subscription-url",
        "subscription-url": "trojan://trojan-placeholder@subscription-hyphen.example.test:443?security=tls"
      },
      {
        "tag": "首尔 node-url",
        "node-url": "hysteria2://hy2-placeholder@node-hyphen.example.test:443"
      }
    ],
    "node-list": [
      "vless://00000000-0000-0000-0000-000000000096@node-hyphen-list.example.test:443"
    ],
    "proxy-list": [
      "trojan://trojan-placeholder@proxy-hyphen-list.example.test:443?security=tls"
    ],
    "server-list": [
      "ss://aes-128-gcm:qa-placeholder@server-hyphen-list.example.test:8388"
    ],
    "subscription-list": [
      "hysteria2://hy2-placeholder@subscription-hyphen-list.example.test:443"
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@raw-content.example.test:8388",
		"vless://00000000-0000-0000-0000-000000000095@share-hyphen.example.test:443#%E9%A6%99%E6%B8%AF%20share-url",
		"trojan://trojan-placeholder@subscription-hyphen.example.test:443?security=tls#%E4%B8%9C%E4%BA%AC%20subscription-url",
		"hysteria2://hy2-placeholder@node-hyphen.example.test:443#%E9%A6%96%E5%B0%94%20node-url",
		"vless://00000000-0000-0000-0000-000000000096@node-hyphen-list.example.test:443",
		"trojan://trojan-placeholder@proxy-hyphen-list.example.test:443?security=tls",
		"ss://aes-128-gcm:qa-placeholder@server-hyphen-list.example.test:8388",
		"hysteria2://hy2-placeholder@subscription-hyphen-list.example.test:443",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON common hyphen alias fields: %q", got)
	}
}

func TestNormalizeContentJSONNamedDownloadLinkFields(t *testing.T) {
	raw := `{
  "data": {
    "download-url": "https://airport.example.test/subscription",
    "items": [
      {
        "name": "香港 subscribeUrl",
        "subscribeUrl": "vless://00000000-0000-0000-0000-000000000097@subscribe-url.example.test:443"
      },
      {
        "remarks": "东京 sub-url",
        "sub-url": "trojan://trojan-placeholder@sub-url.example.test:443?security=tls"
      },
      {
        "tag": "首尔 download_url",
        "download_url": "ss://aes-128-gcm:qa-placeholder@download-url.example.test:8388"
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"vless://00000000-0000-0000-0000-000000000097@subscribe-url.example.test:443#%E9%A6%99%E6%B8%AF%20subscribeUrl",
		"trojan://trojan-placeholder@sub-url.example.test:443?security=tls#%E4%B8%9C%E4%BA%AC%20sub-url",
		"ss://aes-128-gcm:qa-placeholder@download-url.example.test:8388#%E9%A6%96%E5%B0%94%20download_url",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON named download link fields: %q", got)
	}
}

func TestNormalizeContentJSONStructuredAliasFields(t *testing.T) {
	raw := `{
  "data": {
    "nodes": [
      {
        "displayName": "东京 JSON Scheme SS",
        "scheme": "ss",
        "serverAddress": "json-alias-ss.example.test",
        "serverPort": 8388,
        "encryptMethod": "aes-128-gcm",
        "pwd": "qa-placeholder"
      },
      {
        "label": "首尔 JSON ProxyType SOCKS",
        "proxyType": "socks",
        "hostname": "json-alias-socks.example.test",
        "server_port": 1080,
        "userName": "qa-user",
        "pass": "socks-placeholder",
        "version": "4a"
      },
      {
        "title": "香港 JSON Camel Trojan",
        "protocol": "trojan",
        "server": "json-camel-trojan.example.test",
        "port": 443,
        "password": "trojan-placeholder",
        "tls": true,
        "serverName": "json-camel-trojan.example.test",
        "skipCertVerify": true,
        "disableSNI": true,
        "clientFingerprint": "chrome",
        "network": "ws",
        "wsOpts": {
          "path": "/trojan",
          "headers": {
            "Host": "ws.json-camel-trojan.example.test"
          },
          "maxEarlyData": 2048,
          "earlyDataHeaderName": "Sec-WebSocket-Protocol"
        }
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@json-alias-ss.example.test:8388#%E4%B8%9C%E4%BA%AC%20JSON%20Scheme%20SS",
		"socks4a://qa-user:socks-placeholder@json-alias-socks.example.test:1080#%E9%A6%96%E5%B0%94%20JSON%20ProxyType%20SOCKS",
		"trojan://trojan-placeholder@json-camel-trojan.example.test:443?disable_sni=1&early_data_header_name=Sec-WebSocket-Protocol&fp=chrome&host=ws.json-camel-trojan.example.test&insecure=1&max_early_data=2048&path=%2Ftrojan&security=tls&sni=json-camel-trojan.example.test&type=ws#%E9%A6%99%E6%B8%AF%20JSON%20Camel%20Trojan",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON structured alias fields: %q", got)
	}
}

func TestNormalizeContentJSONStructuredVMessTypeUsesProxyParser(t *testing.T) {
	raw := `{
  "nodes": [
    {
      "name": "东京 JSON VMess",
      "type": "vmess",
      "server": "json-vmess.example.test",
      "serverPort": 443,
      "uuid": "00000000-0000-0000-0000-000000000090",
      "network": "ws",
      "wsPath": "/json-vmess",
      "tls": true,
      "sni": "json-vmess.example.test"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	decoded := decodeVMessURIForTest(t, got)
	for _, want := range []string{
		`"ps":"东京 JSON VMess"`,
		`"add":"json-vmess.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000090"`,
		`"net":"ws"`,
		`"path":"/json-vmess"`,
		`"tls":"tls"`,
		`"sni":"json-vmess.example.test"`,
		`"type":""`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected structured JSON VMess document to contain %s: %q", want, decoded)
		}
	}
}

func TestNormalizeContentJSONWrappedStructuredSubscription(t *testing.T) {
	raw := `{
  "data": {
    "raw_content": "proxies:\n  - name: \"香港 JSON 包装\"\n    type: ss\n    server: wrapped-clash.example.net\n    port: 8388\n    cipher: aes-128-gcm\n    password: \"qa-placeholder\""
  },
  "embedded_sing_box": "{\"outbounds\":[{\"type\":\"direct\",\"tag\":\"直连 JSON 包装\"}]}"
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@wrapped-clash.example.net:8388#%E9%A6%99%E6%B8%AF%20JSON%20%E5%8C%85%E8%A3%85",
		"direct://default#%E7%9B%B4%E8%BF%9E%20JSON%20%E5%8C%85%E8%A3%85",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON wrapped structured subscription: %q", got)
	}
}

func TestNormalizeContentJSONNamedObjectMap(t *testing.T) {
	raw := `{
  "香港 03": "vless://00000000-0000-0000-0000-000000000082@example.com:443",
  "新加坡 多线": [
    "ss://YWVzLTEyOC1nY206cGFzc0AxOTIuMC4yLjEwOjgzODg"
  ],
  "东京条目": {
    "name": "东京 03",
    "uri": "trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"trojan://qa-placeholder@example.org:443?security=tls&sni=edge.example.test#%E4%B8%9C%E4%BA%AC%2003",
		"ss://YWVzLTEyOC1nY206cGFzc0AxOTIuMC4yLjEwOjgzODg#%E6%96%B0%E5%8A%A0%E5%9D%A1%20%E5%A4%9A%E7%BA%BF",
		"vless://00000000-0000-0000-0000-000000000082@example.com:443#%E9%A6%99%E6%B8%AF%2003",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected JSON named object map: %q", got)
	}
}

func TestNormalizeContentClashYAML(t *testing.T) {
	raw := `
mixed-port: 7890
proxies:
  - name: "香港 01"
    type: ss
    server: ss.example.test
    port: 8388
    cipher: aes-128-gcm
    password: "qa-placeholder"
    plugin: v2ray-plugin
    plugin-opts: mode=websocket;host=ss-plugin.example.test
    network: tcp
  - { name: "东京 01", type: trojan, server: trojan.example.test, port: 443, password: "trojan-placeholder", sni: edge.example.test, skip-cert-verify: true, disable-sni: true, alpn: "h2,http/1.1", network: ws, ws-opts: { path: /trojan, headers: { Host: ws.trojan.example.test } } }
  - name: "首尔 01"
    type: vless
    server: vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000051
    tls: true
    flow: xtls-rprx-vision
    network: ws
    ws-opts:
      path: /vless
      headers:
        Host: ws.vless.example.test
    sni: vless.example.test
    skip-cert-verify: true
    disable-sni: true
    alpn: h3
  - name: "新加坡 03"
    type: hysteria2
    server: hy2.clash.example.test
    port: 443
    password: "hy2-placeholder"
    obfs: salamander
    obfs-password: "obfs-placeholder"
    up-mbps: 30
    down-mbps: 120
    sni: hy2.clash.example.test
    skip-cert-verify: true
    disable-sni: true
    pinSHA256: clash-hy2-pin
  - { name: "大阪 03", type: tuic, server: tuic.clash.example.test, port: 443, uuid: "00000000-0000-0000-0000-000000000055", password: "tuic-placeholder", congestion-controller: bbr, udp-relay-mode: native, sni: tuic.clash.example.test }
  - name: "香港 04"
    type: hysteria
    server: hysteria.clash.example.test
    port: 443
    auth-str: "hysteria-auth"
    up-mbps: 20
    down-mbps: 80
    obfs: "obfs-placeholder"
    recv-window-conn: 1048576
    recv-window: 2097152
    disable-mtu-discovery: true
    protocol: udp
    sni: hysteria.clash.example.test
    skip-cert-verify: true
  - name: "东京 HTTP"
    type: http
    server: http.clash.example.test
    port: 8080
    username: qa-user
    password: "http-placeholder"
    tls: true
    sni: http.clash.example.test
    skip-cert-verify: true
    disable-sni: true
    alpn: h2,http/1.1
    client-fingerprint: chrome
    path: connect
  - { name: "首尔 SOCKS", type: socks5, server: socks.clash.example.test, port: 1080, username: qa-user, password: "socks-placeholder", udp-over-tcp: true, network: udp }
  - name: "香港 AnyTLS"
    type: anytls
    server: anytls.clash.example.test
    port: 443
    password: "anytls-placeholder"
    idle-session-check-interval: 20s
    idle-session-timeout: 45s
    min-idle-session: 2
    sni: anytls.clash.example.test
  - { name: "东京 ShadowTLS", type: shadowtls, server: shadowtls.clash.example.test, port: 443, version: 3, password: "shadow-placeholder", sni: shadowtls.clash.example.test, skip-cert-verify: true }
  - { name: "新加坡 Naive", type: naive+quic, server: naive.clash.example.test, port: 443, username: qa-user, password: "naive-placeholder", sni: naive.clash.example.test, quic: true, quic-congestion-control: bbr, udp-over-tcp: true, insecure-concurrency: 2 }
  - name: "香港 SSH"
    type: ssh
    server: ssh.clash.example.test
    port: 22
    username: qa-user
    password: "ssh-placeholder"
    private-key-path: keys/qa_id_ed25519
    host-key-algorithms: ssh-ed25519,rsa-sha2-512
    client-version: SSH-2.0-FluxGateQA
  - name: "台北 WireGuard"
    type: wireguard
    server: wg.clash.example.test
    port: 51820
    private-key: cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=
    public-key: cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==
    ip: 10.66.0.2/32
    ipv6: fd00::2/128
    pre-shared-key: cHNrLXBsYWNlaG9sZGVy
    allowed-ips: 0.0.0.0/0,::/0
    reserved: 1,2,3
    mtu: 1420
    udp: true
  - name: "东京 VMess"
    type: vmess
    server: vmess.clash.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000056
    alter-id: 0
    cipher: auto
    network: ws
    ws-path: /ws
    ws-headers.host: ws.clash.example.test
    tls: true
    sni: vmess.clash.example.test
    skip-cert-verify: true
    alpn: h2,http/1.1
proxy-groups:
  - name: Auto
    type: select
    proxies:
      - 香港 01
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 14 {
		t.Fatalf("expected 14 normalized nodes, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ss.example.test:8388?")
	if !strings.Contains(lines[0], "plugin=v2ray-plugin") ||
		!strings.Contains(lines[0], "plugin_opts=mode%3Dwebsocket%3Bhost%3Dss-plugin.example.test") ||
		!strings.Contains(lines[0], "network=tcp") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%2001") {
		t.Fatalf("unexpected clash shadowsocks URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@trojan.example.test:443?")
	if !strings.Contains(lines[1], "sni=edge.example.test") ||
		!strings.Contains(lines[1], "insecure=1") ||
		!strings.Contains(lines[1], "disable_sni=1") ||
		!strings.Contains(lines[1], "alpn=h2%2Chttp%2F1.1") ||
		!strings.Contains(lines[1], "type=ws") ||
		!strings.Contains(lines[1], "path=%2Ftrojan") ||
		!strings.Contains(lines[1], "host=ws.trojan.example.test") ||
		!strings.HasSuffix(lines[1], "#%E4%B8%9C%E4%BA%AC%2001") {
		t.Fatalf("unexpected trojan URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "vless://00000000-0000-0000-0000-000000000051@vless.example.test:443?")
	if !strings.Contains(lines[2], "flow=xtls-rprx-vision") ||
		!strings.Contains(lines[2], "type=ws") ||
		!strings.Contains(lines[2], "path=%2Fvless") ||
		!strings.Contains(lines[2], "host=ws.vless.example.test") ||
		!strings.Contains(lines[2], "security=tls") ||
		!strings.Contains(lines[2], "sni=vless.example.test") ||
		!strings.Contains(lines[2], "insecure=1") ||
		!strings.Contains(lines[2], "disable_sni=1") ||
		!strings.Contains(lines[2], "alpn=h3") {
		t.Fatalf("unexpected vless URI: %q", lines[2])
	}
	assertHasPrefix(t, lines[3], "hysteria2://hy2-placeholder@hy2.clash.example.test:443?")
	if !strings.Contains(lines[3], "obfs=salamander") ||
		!strings.Contains(lines[3], "obfs-password=obfs-placeholder") ||
		!strings.Contains(lines[3], "up_mbps=30") ||
		!strings.Contains(lines[3], "down_mbps=120") ||
		!strings.Contains(lines[3], "insecure=1") ||
		!strings.Contains(lines[3], "disable_sni=1") ||
		!strings.Contains(lines[3], "pinSHA256=clash-hy2-pin") {
		t.Fatalf("unexpected hysteria2 URI: %q", lines[3])
	}
	assertHasPrefix(t, lines[4], "tuic://00000000-0000-0000-0000-000000000055:tuic-placeholder@tuic.clash.example.test:443?")
	if !strings.Contains(lines[4], "congestion_control=bbr") || !strings.Contains(lines[4], "udp_relay_mode=native") {
		t.Fatalf("unexpected tuic URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "hysteria://hysteria-auth@hysteria.clash.example.test:443?")
	if !strings.Contains(lines[5], "up_mbps=20") ||
		!strings.Contains(lines[5], "down_mbps=80") ||
		!strings.Contains(lines[5], "recv_window_conn=1048576") ||
		!strings.Contains(lines[5], "recv_window=2097152") ||
		!strings.Contains(lines[5], "disable_mtu_discovery=1") ||
		!strings.Contains(lines[5], "network=udp") ||
		!strings.Contains(lines[5], "insecure=1") {
		t.Fatalf("unexpected hysteria URI: %q", lines[5])
	}
	assertHasPrefix(t, lines[6], "https://qa-user:http-placeholder@http.clash.example.test:8080/connect?")
	if !strings.Contains(lines[6], "sni=http.clash.example.test") ||
		!strings.Contains(lines[6], "insecure=1") ||
		!strings.Contains(lines[6], "disable_sni=1") ||
		!strings.Contains(lines[6], "alpn=h2%2Chttp%2F1.1") ||
		!strings.Contains(lines[6], "fp=chrome") ||
		!strings.HasSuffix(lines[6], "#%E4%B8%9C%E4%BA%AC%20HTTP") {
		t.Fatalf("unexpected clash http URI: %q", lines[6])
	}
	assertHasPrefix(t, lines[7], "socks5://qa-user:socks-placeholder@socks.clash.example.test:1080?")
	if !strings.Contains(lines[7], "network=udp") || !strings.Contains(lines[7], "udp_over_tcp=1") || !strings.HasSuffix(lines[7], "#%E9%A6%96%E5%B0%94%20SOCKS") {
		t.Fatalf("unexpected clash socks URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "anytls://anytls-placeholder@anytls.clash.example.test:443?")
	if !strings.Contains(lines[8], "idle_session_check_interval=20s") || !strings.Contains(lines[8], "idle_session_timeout=45s") || !strings.Contains(lines[8], "min_idle_session=2") || !strings.Contains(lines[8], "sni=anytls.clash.example.test") {
		t.Fatalf("unexpected clash anytls URI: %q", lines[8])
	}
	assertHasPrefix(t, lines[9], "shadowtls://shadow-placeholder@shadowtls.clash.example.test:443?")
	if !strings.Contains(lines[9], "version=3") || !strings.Contains(lines[9], "sni=shadowtls.clash.example.test") || !strings.Contains(lines[9], "insecure=1") {
		t.Fatalf("unexpected clash shadowtls URI: %q", lines[9])
	}
	assertHasPrefix(t, lines[10], "naive+quic://qa-user:naive-placeholder@naive.clash.example.test:443?")
	if !strings.Contains(lines[10], "quic=1") || !strings.Contains(lines[10], "quic_congestion_control=bbr") || !strings.Contains(lines[10], "udp_over_tcp=1") || !strings.Contains(lines[10], "insecure_concurrency=2") || !strings.Contains(lines[10], "sni=naive.clash.example.test") {
		t.Fatalf("unexpected clash naive URI: %q", lines[10])
	}
	assertHasPrefix(t, lines[11], "ssh://qa-user:ssh-placeholder@ssh.clash.example.test:22?")
	if !strings.Contains(lines[11], "private_key_path=keys%2Fqa_id_ed25519") || !strings.Contains(lines[11], "host_key_algorithms=ssh-ed25519%2Crsa-sha2-512") || !strings.Contains(lines[11], "client_version=SSH-2.0-FluxGateQA") {
		t.Fatalf("unexpected clash ssh URI: %q", lines[11])
	}
	assertHasPrefix(t, lines[12], "wireguard://wg.clash.example.test:51820?")
	if !strings.Contains(lines[12], "private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI") ||
		!strings.Contains(lines[12], "peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg") ||
		!strings.Contains(lines[12], "local_address=10.66.0.2%2F32%2Cfd00%3A%3A2%2F128") ||
		!strings.Contains(lines[12], "allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0") ||
		!strings.Contains(lines[12], "reserved=1%2C2%2C3") ||
		!strings.Contains(lines[12], "network=udp") ||
		!strings.Contains(lines[12], "mtu=1420") {
		t.Fatalf("unexpected clash wireguard URI: %q", lines[12])
	}
	assertHasPrefix(t, lines[13], "vmess://")
	decodedVMess, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[13], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode clash vmess URI: %v", err)
	}
	decodedVMessText := string(decodedVMess)
	if !strings.Contains(decodedVMessText, `"allowInsecure":"1"`) ||
		!strings.Contains(decodedVMessText, `"alpn":"h2,http/1.1"`) ||
		!strings.Contains(decodedVMessText, `"sni":"vmess.clash.example.test"`) {
		t.Fatalf("unexpected clash vmess document: %q", decodedVMessText)
	}
}

func TestNormalizeContentClashYAMLProtocolAliases(t *testing.T) {
	raw := `
proxies:
  - { name: "Clash Trojan-Go Alias", type: trojan-go, server: trojan-go.clash-alias.example.test, port: 443, password: "trojan-placeholder", sni: trojan-go.clash-alias.example.test }
  - { name: "Clash VMess AEAD Alias", type: vmess-aead, server: vmess-aead.clash-alias.example.test, port: 443, uuid: "00000000-0000-0000-0000-000000000057", tls: true, sni: vmess-aead.clash-alias.example.test }
  - { name: "Clash HY2 Alias", type: hy2, server: hy2.clash-alias.example.test, port: 443, password: "hy2-placeholder" }
  - { name: "Clash AnyTLS Alias", type: any-tls, server: anytls.clash-alias.example.test, port: 443, password: "anytls-placeholder" }
  - { name: "Clash ShadowTLS Alias", type: shadow-tls, server: shadowtls.clash-alias.example.test, port: 443, version: 3, password: "shadowtls-placeholder" }
  - { name: "Clash HTTP Plus TLS Alias", type: http+tls, server: http-plus-tls.clash-alias.example.test, port: 443, username: qa-user, password: "http-placeholder", path: /connect }
  - { name: "Clash HTTP Dash TLS Alias", type: http-tls, server: http-dash-tls.clash-alias.example.test, port: 443, username: qa-user, password: "http-dash-placeholder" }
  - { name: "Clash Naive QUIC Alias", type: naive-quic, server: naive.clash-alias.example.test, port: 443, username: qa-user, password: "naive-placeholder" }
  - { name: "Clash Naive HTTPS Alias", type: naive+https, server: naive-https.clash-alias.example.test, port: 443, username: qa-user, password: "naive-https-placeholder", sni: naive-https.clash-alias.example.test }
  - { name: "Clash Naive Dash HTTPS Alias", type: naive-https, server: naive-dash-https.clash-alias.example.test, port: 443, username: qa-user, password: "naive-dash-https-placeholder", sni: naive-dash-https.clash-alias.example.test }
  - { name: "Clash SOCKS5H Alias", type: socks5h, server: socks5h.clash-alias.example.test, port: 1080, username: qa-user, password: "socks-placeholder", network: udp }
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 11 {
		t.Fatalf("expected 11 normalized Clash alias nodes, got %d: %q", len(lines), got)
	}

	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@trojan-go.clash-alias.example.test:443?")
	if !strings.Contains(lines[0], "sni=trojan-go.clash-alias.example.test") ||
		!strings.HasSuffix(lines[0], "#Clash%20Trojan-Go%20Alias") {
		t.Fatalf("unexpected Clash Trojan-Go alias URI: %q", lines[0])
	}

	assertHasPrefix(t, lines[1], "vmess://")
	decodedVMess := decodeVMessURIForTest(t, lines[1])
	for _, want := range []string{
		`"ps":"Clash VMess AEAD Alias"`,
		`"add":"vmess-aead.clash-alias.example.test"`,
		`"id":"00000000-0000-0000-0000-000000000057"`,
		`"tls":"tls"`,
		`"sni":"vmess-aead.clash-alias.example.test"`,
	} {
		if !strings.Contains(decodedVMess, want) {
			t.Fatalf("expected decoded Clash VMess alias URI to contain %q: %q", want, decodedVMess)
		}
	}

	assertHasPrefix(t, lines[2], "hysteria2://hy2-placeholder@hy2.clash-alias.example.test:443#Clash%20HY2%20Alias")
	assertHasPrefix(t, lines[3], "anytls://anytls-placeholder@anytls.clash-alias.example.test:443#Clash%20AnyTLS%20Alias")
	assertHasPrefix(t, lines[4], "shadowtls://shadowtls-placeholder@shadowtls.clash-alias.example.test:443?")
	if !strings.Contains(lines[4], "version=3") ||
		!strings.HasSuffix(lines[4], "#Clash%20ShadowTLS%20Alias") {
		t.Fatalf("unexpected Clash ShadowTLS alias URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "https://qa-user:http-placeholder@http-plus-tls.clash-alias.example.test:443/connect#Clash%20HTTP%20Plus%20TLS%20Alias")
	assertHasPrefix(t, lines[6], "https://qa-user:http-dash-placeholder@http-dash-tls.clash-alias.example.test:443#Clash%20HTTP%20Dash%20TLS%20Alias")
	assertHasPrefix(t, lines[7], "naive+quic://qa-user:naive-placeholder@naive.clash-alias.example.test:443?")
	if !strings.Contains(lines[7], "quic=1") ||
		!strings.HasSuffix(lines[7], "#Clash%20Naive%20QUIC%20Alias") {
		t.Fatalf("unexpected Clash Naive QUIC alias URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "naive://qa-user:naive-https-placeholder@naive-https.clash-alias.example.test:443?")
	if !strings.Contains(lines[8], "sni=naive-https.clash-alias.example.test") ||
		!strings.HasSuffix(lines[8], "#Clash%20Naive%20HTTPS%20Alias") {
		t.Fatalf("unexpected Clash Naive HTTPS alias URI: %q", lines[8])
	}
	assertHasPrefix(t, lines[9], "naive://qa-user:naive-dash-https-placeholder@naive-dash-https.clash-alias.example.test:443?")
	if !strings.Contains(lines[9], "sni=naive-dash-https.clash-alias.example.test") ||
		!strings.HasSuffix(lines[9], "#Clash%20Naive%20Dash%20HTTPS%20Alias") {
		t.Fatalf("unexpected Clash Naive Dash HTTPS alias URI: %q", lines[9])
	}
	assertHasPrefix(t, lines[10], "socks5://qa-user:socks-placeholder@socks5h.clash-alias.example.test:1080?")
	if !strings.Contains(lines[10], "network=udp") ||
		!strings.HasSuffix(lines[10], "#Clash%20SOCKS5H%20Alias") {
		t.Fatalf("unexpected Clash SOCKS5H alias URI: %q", lines[10])
	}
}

func TestNormalizeContentClashYAMLNestedProviderProxies(t *testing.T) {
	raw := `
proxy-providers:
  airport-a:
    type: file
    path: ./airport-a.yaml
    proxies:
      - name: "Provider 香港 01"
        type: ss
        server: provider-ss.example.test
        port: 8388
        cipher: aes-128-gcm
        password: "qa-placeholder"
    health-check:
      enable: true
proxy-groups:
  - name: auto
    type: select
    proxies:
      - Provider 香港 01
proxies:
  - name: "Top 东京 01"
    type: trojan
    server: top-trojan.example.test
    port: 443
    password: "trojan-placeholder"
    tls: true
    sni: top-trojan.example.test
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@provider-ss.example.test:8388#Provider%20%E9%A6%99%E6%B8%AF%2001",
		"trojan://trojan-placeholder@top-trojan.example.test:443?security=tls&sni=top-trojan.example.test#Top%20%E4%B8%9C%E4%BA%AC%2001",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected nested provider proxies: %q", got)
	}
}

func TestNormalizeContentClashYAMLSOCKSUDPFlag(t *testing.T) {
	raw := `
proxies:
  - name: "SOCKS UDP"
    type: socks5
    server: socks-udp.example.test
    port: 1080
    username: qa-user
    password: "socks-placeholder"
    udp: true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@socks-udp.example.test:1080?")
	if !strings.Contains(got, "udp=1") || strings.Contains(got, "network=udp") {
		t.Fatalf("unexpected Clash SOCKS UDP URI: %q", got)
	}
	if !strings.HasSuffix(got, "#SOCKS%20UDP") {
		t.Fatalf("unexpected Clash SOCKS UDP fragment: %q", got)
	}
}

func TestNormalizeContentClashYAMLInlineProxyLists(t *testing.T) {
	raw := `
proxy-providers:
  inline-a:
    type: inline
    proxies: [{ name: "Inline 香港 01", type: ss, server: inline-ss.example.test, port: 8388, cipher: aes-128-gcm, password: "qa-placeholder" }]
proxy-groups:
  - { name: auto, type: select, proxies: ["Inline 香港 01", "Inline 东京 01"] }
proxies: [{ name: "Inline 东京 01", type: trojan, server: inline-trojan.example.test, port: 443, password: "trojan-placeholder", tls: true, sni: inline-trojan.example.test }]
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@inline-ss.example.test:8388#Inline%20%E9%A6%99%E6%B8%AF%2001",
		"trojan://trojan-placeholder@inline-trojan.example.test:443?security=tls&sni=inline-trojan.example.test#Inline%20%E4%B8%9C%E4%BA%AC%2001",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected inline proxy list URIs: %q", got)
	}
}

func TestNormalizeContentClashYAMLProxyListAliases(t *testing.T) {
	raw := `
proxy-providers:
  alias-a:
    type: inline
    proxies-list:
      - name: "Alias Provider 香港 01"
        type: ss
        server: alias-provider-ss.example.test
        port: 8388
        cipher: aes-128-gcm
        password: "qa-placeholder"
proxy-list: [{ name: "Alias Inline 东京 01", type: trojan, server: alias-inline-trojan.example.test, port: 443, password: "trojan-placeholder", tls: true, sni: alias-inline-trojan.example.test }]
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@alias-provider-ss.example.test:8388#Alias%20Provider%20%E9%A6%99%E6%B8%AF%2001",
		"trojan://trojan-placeholder@alias-inline-trojan.example.test:443?security=tls&sni=alias-inline-trojan.example.test#Alias%20Inline%20%E4%B8%9C%E4%BA%AC%2001",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected proxy list alias URIs: %q", got)
	}
}

func TestNormalizeContentClashYAMLAnchoredProxyList(t *testing.T) {
	raw := `
proxies: &airport_nodes
  - name: "Anchor 香港 01"
    type: ss
    server: anchor-ss.example.test
    port: 8388
    cipher: aes-128-gcm
    password: "qa-placeholder"
proxy-groups:
  - name: auto
    type: select
    proxies: *airport_nodes
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "ss://aes-128-gcm:qa-placeholder@anchor-ss.example.test:8388#Anchor%20%E9%A6%99%E6%B8%AF%2001"
	if got != want {
		t.Fatalf("unexpected anchored proxy list URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLAnchoredInlineProxyItems(t *testing.T) {
	raw := `
proxy-providers:
  inline-anchor:
    type: inline
    proxies: [&inline_hk { name: "Inline Anchor 香港 01", type: ss, server: inline-anchor-ss.example.test, port: 8388, cipher: aes-128-gcm, password: "qa-placeholder" }]
proxies:
  - &top_tokyo { name: "Top Anchor 东京 01", type: trojan, server: top-anchor-trojan.example.test, port: 443, password: "trojan-placeholder", tls: true, sni: top-anchor-trojan.example.test }
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:qa-placeholder@inline-anchor-ss.example.test:8388#Inline%20Anchor%20%E9%A6%99%E6%B8%AF%2001",
		"trojan://trojan-placeholder@top-anchor-trojan.example.test:443?security=tls&sni=top-anchor-trojan.example.test#Top%20Anchor%20%E4%B8%9C%E4%BA%AC%2001",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected anchored inline proxy item URIs: %q", got)
	}
}

func TestNormalizeContentClashYAMLInternalOutbounds(t *testing.T) {
	raw := `
proxies:
  - name: "本地直连"
    type: direct
  - name: "拒绝访问"
    type: reject
  - { name: "静默拦截", type: reject-drop }
  - { name: "无丢包拦截", type: reject-no-drop }
  - { name: "小图拦截", type: reject-tinygif }
  - name: "内部 DNS"
    type: dns
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"direct://default#%E6%9C%AC%E5%9C%B0%E7%9B%B4%E8%BF%9E",
		"block://default#%E6%8B%92%E7%BB%9D%E8%AE%BF%E9%97%AE",
		"block://default#%E9%9D%99%E9%BB%98%E6%8B%A6%E6%88%AA",
		"block://default#%E6%97%A0%E4%B8%A2%E5%8C%85%E6%8B%A6%E6%88%AA",
		"block://default#%E5%B0%8F%E5%9B%BE%E6%8B%A6%E6%88%AA",
		"dns://default#%E5%86%85%E9%83%A8%20DNS",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected clash internal outbound URIs: %q", got)
	}
}

func TestNormalizeContentClashYAMLVLESSGRPC(t *testing.T) {
	raw := `
proxies:
  - name: "香港 gRPC"
    type: vless
    server: grpc.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000071
    tls: true
    network: grpc
    grpc-opts:
      grpc-service-name: fluxgate
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000071@grpc.vless.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "type=grpc") ||
		!strings.Contains(got, "service_name=fluxgate") ||
		!strings.HasSuffix(got, "#%E9%A6%99%E6%B8%AF%20gRPC") {
		t.Fatalf("unexpected clash vless grpc URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLCamelCaseFields(t *testing.T) {
	raw := `
proxies:
  - name: "香港 YAML Camel VLESS"
    type: vless
    server: yaml-camel-vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000094
    tls: true
    serverName: yaml-camel-vless.example.test
    skipCertVerify: true
    disableSNI: true
    clientFingerprint: chrome
    wsPath: /camel
    wsHost: ws.yaml-camel-vless.example.test
    maxEarlyData: 2048
    earlyDataHeaderName: Sec-WebSocket-Protocol
  - { name: "东京 YAML Camel Trojan", type: trojan, server: yaml-camel-trojan.example.test, port: 443, password: "trojan-placeholder", tls: true, serverName: yaml-camel-trojan.example.test, allowInsecure: true, wsOpts: { path: /trojan, headers: { Host: ws.yaml-camel-trojan.example.test }, maxEarlyData: 1024, earlyDataHeaderName: X-FluxGate-ED } }
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Clash YAML camelCase URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000094@yaml-camel-vless.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=yaml-camel-vless.example.test",
		"insecure=1",
		"disable_sni=1",
		"fp=chrome",
		"type=ws",
		"path=%2Fcamel",
		"host=ws.yaml-camel-vless.example.test",
		"max_early_data=2048",
		"early_data_header_name=Sec-WebSocket-Protocol",
		"#%E9%A6%99%E6%B8%AF%20YAML%20Camel%20VLESS",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Clash YAML camelCase VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@yaml-camel-trojan.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=yaml-camel-trojan.example.test",
		"insecure=1",
		"type=ws",
		"path=%2Ftrojan",
		"host=ws.yaml-camel-trojan.example.test",
		"max_early_data=1024",
		"early_data_header_name=X-FluxGate-ED",
		"#%E4%B8%9C%E4%BA%AC%20YAML%20Camel%20Trojan",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Clash YAML inline camelCase Trojan URI to contain %q: %q", want, lines[1])
		}
	}
}

func TestNormalizeContentClashYAMLVLESSGRPCKeepaliveOptions(t *testing.T) {
	raw := `
proxies:
  - name: "东京 gRPC Keepalive"
    type: vless
    server: grpc.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000077
    tls: true
    network: grpc
    grpc-opts:
      grpc-service-name: fluxgate
      idle-timeout: 30s
      ping-timeout: 10s
      permit-without-stream: true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000077@grpc.vless.example.test:443?")
	if !strings.Contains(got, "type=grpc") ||
		!strings.Contains(got, "service_name=fluxgate") ||
		!strings.Contains(got, "idle_timeout=30s") ||
		!strings.Contains(got, "ping_timeout=10s") ||
		!strings.Contains(got, "permit_without_stream=true") {
		t.Fatalf("unexpected clash vless grpc keepalive URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLTrojanQUICTransport(t *testing.T) {
	raw := `
proxies:
  - name: "东京 QUIC"
    type: trojan
    server: quic.trojan.example.test
    port: 443
    password: trojan-placeholder
    tls: true
    network: quic
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@quic.trojan.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "type=quic") ||
		!strings.HasSuffix(got, "#%E4%B8%9C%E4%BA%AC%20QUIC") {
		t.Fatalf("unexpected clash trojan quic URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLVLESSHTTPTransport(t *testing.T) {
	raw := `
proxies:
  - name: "香港 HTTP"
    type: vless
    server: http.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000073
    tls: true
    http-opts:
      host: [h2.example.test, h2-backup.example.test]
      path: /h2
      method: GET
      idle-timeout: 20s
      ping-timeout: 10s
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000073@http.vless.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "type=http") ||
		!strings.Contains(got, "host=h2.example.test%2Ch2-backup.example.test") ||
		!strings.Contains(got, "path=%2Fh2") ||
		!strings.Contains(got, "method=GET") ||
		!strings.Contains(got, "idle_timeout=20s") ||
		!strings.Contains(got, "ping_timeout=10s") ||
		!strings.HasSuffix(got, "#%E9%A6%99%E6%B8%AF%20HTTP") {
		t.Fatalf("unexpected clash vless http transport URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLTrojanWebSocketEarlyData(t *testing.T) {
	raw := `
proxies:
  - name: "东京 WS Early"
    type: trojan
    server: ws.trojan.example.test
    port: 443
    password: trojan-placeholder
    tls: true
    network: ws
    ws-opts:
      path: /ws
      headers:
        Host: ws.example.test
      max-early-data: 2048
      early-data-header-name: Sec-WebSocket-Protocol
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@ws.trojan.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "type=ws") ||
		!strings.Contains(got, "path=%2Fws") ||
		!strings.Contains(got, "host=ws.example.test") ||
		!strings.Contains(got, "max_early_data=2048") ||
		!strings.Contains(got, "early_data_header_name=Sec-WebSocket-Protocol") ||
		!strings.HasSuffix(got, "#%E4%B8%9C%E4%BA%AC%20WS%20Early") {
		t.Fatalf("unexpected clash trojan websocket early data URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLTrojanHTTPUpgradeTransport(t *testing.T) {
	raw := `
proxies:
  - name: "东京 HTTPUpgrade"
    type: trojan
    server: upgrade.trojan.example.test
    port: 443
    password: trojan-placeholder
    tls: true
    httpupgrade-opts:
      host: upgrade.example.test
      path: /upgrade
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@upgrade.trojan.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "type=httpupgrade") ||
		!strings.Contains(got, "host=upgrade.example.test") ||
		!strings.Contains(got, "path=%2Fupgrade") ||
		!strings.HasSuffix(got, "#%E4%B8%9C%E4%BA%AC%20HTTPUpgrade") {
		t.Fatalf("unexpected clash trojan httpupgrade URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLInlineLists(t *testing.T) {
	raw := `
proxies:
  - name: "香港 ALPN"
    type: vless
    server: alpn.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000074
    tls: true
    alpn: [h2, http/1.1]
  - name: "台北 WireGuard"
    type: wireguard
    server: wg.inline.example.test
    port: 51820
    private-key: private-key-placeholder
    public-key: public-key-placeholder
    local-address: [10.66.0.2/32, fd00::2/128]
    allowed-ips: [0.0.0.0/0, ::/0]
    reserved: [1, 2, 3]
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 normalized nodes, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000074@alpn.vless.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") || !strings.Contains(lines[0], "alpn=h2%2Chttp%2F1.1") {
		t.Fatalf("unexpected clash vless inline list URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "wireguard://wg.inline.example.test:51820?")
	if !strings.Contains(lines[1], "local_address=10.66.0.2%2F32%2Cfd00%3A%3A2%2F128") ||
		!strings.Contains(lines[1], "allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0") ||
		!strings.Contains(lines[1], "reserved=1%2C2%2C3") {
		t.Fatalf("unexpected clash wireguard inline list URI: %q", lines[1])
	}
}

func TestNormalizeContentClashYAMLBlockLists(t *testing.T) {
	raw := `
proxies:
  - name: "香港 ALPN 块"
    type: vless
    server: block-alpn.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000076
    tls: true
    alpn:
      - h2
      - http/1.1
  - name: "台北 WireGuard 块"
    type: wireguard
    server: wg.block.example.test
    port: 51820
    private-key: private-key-placeholder
    public-key: public-key-placeholder
    local-address:
      - 10.66.0.2/32
      - fd00::2/128
    allowed-ips:
      - 0.0.0.0/0
      - ::/0
    reserved:
      - 1
      - 2
      - 3
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 normalized nodes, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000076@block-alpn.vless.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") || !strings.Contains(lines[0], "alpn=h2%2Chttp%2F1.1") {
		t.Fatalf("unexpected clash vless block list URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "wireguard://wg.block.example.test:51820?")
	if !strings.Contains(lines[1], "local_address=10.66.0.2%2F32%2Cfd00%3A%3A2%2F128") ||
		!strings.Contains(lines[1], "allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0") ||
		!strings.Contains(lines[1], "reserved=1%2C2%2C3") {
		t.Fatalf("unexpected clash wireguard block list URI: %q", lines[1])
	}
}

func TestNormalizeContentClashYAMLVLESSReality(t *testing.T) {
	raw := `
proxies:
  - name: "香港 Reality"
    type: vless
    server: reality.vless.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000075
    flow: xtls-rprx-vision
    network: tcp
    sni: www.example.test
    client-fingerprint: chrome
    reality-opts:
      public-key: reality-public-key-placeholder
      short-id: a1b2c3d4
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000075@reality.vless.example.test:443?")
	if !strings.Contains(got, "security=reality") ||
		!strings.Contains(got, "pbk=reality-public-key-placeholder") ||
		!strings.Contains(got, "sid=a1b2c3d4") ||
		!strings.Contains(got, "fp=chrome") ||
		!strings.Contains(got, "sni=www.example.test") ||
		!strings.Contains(got, "flow=xtls-rprx-vision") {
		t.Fatalf("unexpected clash vless reality URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLTrojanReality(t *testing.T) {
	raw := `
proxies:
  - name: "香港 Trojan Reality"
    type: trojan
    server: reality.trojan.example.test
    port: 443
    password: trojan-placeholder
    sni: www.example.test
    client-fingerprint: chrome
    reality-opts:
      public-key: trojan-reality-public-key
      short-id: b1c2d3e4
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@reality.trojan.example.test:443?")
	if !strings.Contains(got, "security=reality") ||
		!strings.Contains(got, "pbk=trojan-reality-public-key") ||
		!strings.Contains(got, "sid=b1c2d3e4") ||
		!strings.Contains(got, "fp=chrome") ||
		!strings.Contains(got, "sni=www.example.test") {
		t.Fatalf("unexpected clash trojan reality URI: %q", got)
	}
}

func TestNormalizeContentClashYAMLVMessGRPC(t *testing.T) {
	raw := `
proxies:
  - name: "大阪 VMess gRPC"
    type: vmess
    server: grpc.vmess.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000073
    alter-id: 0
    cipher: auto
    tls: true
    network: grpc
    grpc-opts:
      grpc-service-name: fluxgate-vmess
      multi-mode: true
    sni: grpc.vmess.example.test
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vmess://")
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode clash vmess grpc URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"net":"grpc"`) ||
		!strings.Contains(decodedText, `"path":"fluxgate-vmess"`) ||
		!strings.Contains(decodedText, `"multi_mode":"true"`) ||
		!strings.Contains(decodedText, `"tls":"tls"`) ||
		!strings.Contains(decodedText, `"sni":"grpc.vmess.example.test"`) {
		t.Fatalf("unexpected clash vmess grpc document: %q", decodedText)
	}
}

func TestNormalizeContentClashYAMLVMessPacketEncoding(t *testing.T) {
	raw := `
proxies:
  - name: "大阪 VMess Packet Encoding"
    type: vmess
    server: packet.vmess.example.test
    port: 443
    uuid: 00000000-0000-0000-0000-000000000096
    cipher: auto
    packetEncoding: packetaddr
    disable-sni: true
    client-fingerprint: chrome
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vmess://")
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode clash vmess URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"packet_encoding":"packetaddr"`) ||
		!strings.Contains(decodedText, `"disable_sni":"1"`) ||
		!strings.Contains(decodedText, `"fp":"chrome"`) ||
		!strings.Contains(decodedText, `"add":"packet.vmess.example.test"`) {
		t.Fatalf("unexpected clash vmess packet encoding document: %q", decodedText)
	}
}

func TestNormalizeContentSIP008(t *testing.T) {
	raw := `{
  "version": 1,
  "servers": [
    {
      "id": "node-1",
      "remarks": "香港 02",
      "server": "sip008.example.test",
      "server_port": 8388,
      "method": "aes-256-gcm",
      "password": "qa-placeholder",
      "plugin": "obfs-local",
      "plugin_opts": "obfs=http;obfs-host=sip008.example.test",
      "network": "tcp"
    },
    {
      "remarks": "skip me",
      "server": "missing-password.example.test",
      "server_port": 8388,
      "method": "aes-256-gcm"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 normalized node, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-256-gcm:qa-placeholder@sip008.example.test:8388?")
	if !strings.Contains(lines[0], "plugin=obfs-local") ||
		!strings.Contains(lines[0], "plugin_opts=obfs%3Dhttp%3Bobfs-host%3Dsip008.example.test") ||
		!strings.Contains(lines[0], "network=tcp") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%2002") {
		t.Fatalf("unexpected SIP008 URI fragment: %q", lines[0])
	}
}

func TestNormalizeContentSIP008FieldAliases(t *testing.T) {
	raw := `{
  "version": 1,
  "servers": [
    {
      "displayName": "香港 SIP008 别名",
      "address": "sip008-alias.example.test",
      "portNumber": 8388,
      "encryptMethod": "chacha20-ietf-poly1305",
      "pass": "alias-placeholder"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "ss://chacha20-ietf-poly1305:alias-placeholder@sip008-alias.example.test:8388#%E9%A6%99%E6%B8%AF%20SIP008%20%E5%88%AB%E5%90%8D"
	if got != want {
		t.Fatalf("unexpected SIP008 alias URI: %q", got)
	}
}

func TestNormalizeContentSIP008FieldCaseAliases(t *testing.T) {
	raw := `{
  "Version": 1,
  "Servers": [
    {
      "Display_Name": "香港 SIP008 大小写别名",
      "Address": "sip008-case.example.test",
      "Server-Port": 8388,
      "Encrypt_Method": "aes-128-gcm",
      "Password": "case-placeholder",
      "Plugin": "v2ray-plugin",
      "PluginOpts": "mode=websocket;host=sip008-case.example.test",
      "Network": "udp"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "ss://aes-128-gcm:case-placeholder@sip008-case.example.test:8388?network=udp&plugin=v2ray-plugin&plugin_opts=mode%3Dwebsocket%3Bhost%3Dsip008-case.example.test#%E9%A6%99%E6%B8%AF%20SIP008%20%E5%A4%A7%E5%B0%8F%E5%86%99%E5%88%AB%E5%90%8D"
	if got != want {
		t.Fatalf("unexpected SIP008 case alias URI: %q", got)
	}
}

func TestNormalizeContentSIP008ServerObjectMap(t *testing.T) {
	raw := `{
  "version": 1,
  "servers": {
    "香港 SIP008 映射": {
      "server": "sip008-map.example.test",
      "server_port": 8388,
      "method": "aes-128-gcm",
      "password": "qa-placeholder",
      "plugin": "v2ray-plugin",
      "plugin_opts": "mode=websocket;host=sip008-map.example.test",
      "network": "tcp"
    },
    "东京 SIP008 映射": {
      "name": "东京 自定义名",
      "server": "sip008-map-2.example.test",
      "server_port": 8389,
      "method": "aes-256-gcm",
      "password": "qa-placeholder"
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-256-gcm:qa-placeholder@sip008-map-2.example.test:8389#%E4%B8%9C%E4%BA%AC%20%E8%87%AA%E5%AE%9A%E4%B9%89%E5%90%8D",
		"ss://aes-128-gcm:qa-placeholder@sip008-map.example.test:8388?network=tcp&plugin=v2ray-plugin&plugin_opts=mode%3Dwebsocket%3Bhost%3Dsip008-map.example.test#%E9%A6%99%E6%B8%AF%20SIP008%20%E6%98%A0%E5%B0%84",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected SIP008 object map URI: %q", got)
	}
}

func TestNormalizeContentSSDServerObjectMapAndAliases(t *testing.T) {
	raw := `{
  "portNumber": 8388,
  "encryptMethod": "aes-128-gcm",
  "pass": "doc-placeholder",
  "servers": {
    "SSD Map A": {
      "nodeHost": "ssd-map-a.example.test"
    },
    "SSD Map B": {
      "displayName": "首尔 SSD 别名",
      "endpoint": "ssd-map-b.example.test",
      "nodePort": 8389,
      "security": "aes-256-gcm",
      "passwd": "node-placeholder"
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"ss://aes-128-gcm:doc-placeholder@ssd-map-a.example.test:8388#SSD%20Map%20A",
		"ss://aes-256-gcm:node-placeholder@ssd-map-b.example.test:8389#%E9%A6%96%E5%B0%94%20SSD%20%E5%88%AB%E5%90%8D",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected SSD object map alias URIs: %q", got)
	}
}

func TestNormalizeContentSSDFieldCaseAliases(t *testing.T) {
	raw := `{
  "Airport": "SSD QA",
  "PortNumber": 8388,
  "Encrypt-Method": "aes-128-gcm",
  "Password": "doc-placeholder",
  "Servers": [
    {
      "Display_Name": "首尔 SSD 大小写别名",
      "NodeHost": "ssd-case.example.test",
      "PluginOptions": "mode=websocket;host=ssd-case.example.test",
      "Network": "tcp"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := "ss://aes-128-gcm:doc-placeholder@ssd-case.example.test:8388?network=tcp&plugin_opts=mode%3Dwebsocket%3Bhost%3Dssd-case.example.test#%E9%A6%96%E5%B0%94%20SSD%20%E5%A4%A7%E5%B0%8F%E5%86%99%E5%88%AB%E5%90%8D"
	if got != want {
		t.Fatalf("unexpected SSD case alias URI: %q", got)
	}
}

func TestNormalizeContentSSD(t *testing.T) {
	rawJSON := `{
  "airport": "SSD QA",
  "port": 8388,
  "encryption": "aes-128-gcm",
  "password": "qa-placeholder",
  "servers": [
    {
      "remarks": "香港 SSD 01",
      "server": "ssd-hk.example.test",
      "plugin": "v2ray-plugin",
      "plugin_options": "mode=websocket;host=ssd-hk.example.test",
      "network": "tcp"
    },
    {
      "remarks": "东京 SSD 02",
      "server": "ssd-tokyo.example.test",
      "port": 8389,
      "encryption": "aes-256-gcm",
      "password": "tokyo-placeholder"
    },
    {
      "remarks": "skip me",
      "password": "missing-server-placeholder"
    }
  ]
}`
	raw := "ssd://" + base64.RawStdEncoding.EncodeToString([]byte(rawJSON))
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 SSD URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ssd-hk.example.test:8388?")
	if !strings.Contains(lines[0], "network=tcp") ||
		!strings.Contains(lines[0], "plugin=v2ray-plugin") ||
		!strings.Contains(lines[0], "plugin_opts=mode%3Dwebsocket%3Bhost%3Dssd-hk.example.test") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20SSD%2001") {
		t.Fatalf("unexpected SSD first URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-256-gcm:tokyo-placeholder@ssd-tokyo.example.test:8389#%E4%B8%9C%E4%BA%AC%20SSD%2002" {
		t.Fatalf("unexpected SSD second URI: %q", lines[1])
	}
}

func TestNormalizeContentSSDJSONAndWrappedSSDURI(t *testing.T) {
	rawSSDJSON := `{
  "port": 8388,
  "encryption": "aes-128-gcm",
  "password": "qa-placeholder",
  "servers": [
    {
      "remarks": "首尔 SSD JSON",
      "server": "ssd-json.example.test"
    }
  ]
}`
	got, err := NormalizeContent(rawSSDJSON)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "ss://aes-128-gcm:qa-placeholder@ssd-json.example.test:8388#%E9%A6%96%E5%B0%94%20SSD%20JSON" {
		t.Fatalf("unexpected SSD JSON URI: %q", got)
	}

	wrapped := `{
  "data": {
    "raw_content": "ssd://` + base64.StdEncoding.EncodeToString([]byte(rawSSDJSON)) + `"
  }
}`
	got, err = NormalizeContent(wrapped)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "ss://aes-128-gcm:qa-placeholder@ssd-json.example.test:8388#%E9%A6%96%E5%B0%94%20SSD%20JSON" {
		t.Fatalf("unexpected wrapped SSD URI: %q", got)
	}
}

func TestNormalizeContentSurgeProxyList(t *testing.T) {
	raw := `
[General]
loglevel = notify

[Proxy]
香港 Surge SS = ss, ss.surge.example.test, 8388, encrypt-method=aes-128-gcm, password=qa-placeholder, plugin=v2ray-plugin, plugin-opts=mode=websocket;host=ss.surge.example.test, network=tcp
东京 Surge Trojan = trojan, trojan.surge.example.test, 443, password=trojan-placeholder, sni=trojan.surge.example.test, skip-cert-verify=true, ws=true, ws-path=/trojan, ws-headers=Host:ws.trojan.surge.example.test
首尔 Surge HTTPS = https, http.surge.example.test, 8443, username=qa-user, password=http-placeholder, sni=http.surge.example.test
大阪 Surge SOCKS = socks5, socks.surge.example.test, 1080, username=qa-user, password=socks-placeholder, udp-relay=true

[Rule]
FINAL,DIRECT
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 Surge URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ss.surge.example.test:8388?")
	if !strings.Contains(lines[0], "network=tcp") ||
		!strings.Contains(lines[0], "plugin=v2ray-plugin") ||
		!strings.Contains(lines[0], "plugin_opts=mode%3Dwebsocket%3Bhost%3Dss.surge.example.test") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20Surge%20SS") {
		t.Fatalf("unexpected Surge SS URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@trojan.surge.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=trojan.surge.example.test",
		"insecure=1",
		"type=ws",
		"path=%2Ftrojan",
		"host=ws.trojan.surge.example.test",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Surge Trojan URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#%E4%B8%9C%E4%BA%AC%20Surge%20Trojan") {
		t.Fatalf("unexpected Surge Trojan fragment: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "https://qa-user:http-placeholder@http.surge.example.test:8443?")
	if !strings.Contains(lines[2], "sni=http.surge.example.test") ||
		!strings.HasSuffix(lines[2], "#%E9%A6%96%E5%B0%94%20Surge%20HTTPS") {
		t.Fatalf("unexpected Surge HTTPS URI: %q", lines[2])
	}
	if lines[3] != "socks5://qa-user:socks-placeholder@socks.surge.example.test:1080?network=udp#%E5%A4%A7%E9%98%AA%20Surge%20SOCKS" {
		t.Fatalf("unexpected Surge SOCKS URI: %q", lines[3])
	}
}

func TestNormalizeContentJSONWrappedSurgeProxyList(t *testing.T) {
	raw := `{
  "data": {
    "raw_content": "[Proxy]\n香港 包装 Surge = trojan, wrapped-surge.example.test, 443, password=trojan-placeholder, sni=wrapped-surge.example.test"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@wrapped-surge.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=wrapped-surge.example.test") ||
		!strings.HasSuffix(got, "#%E9%A6%99%E6%B8%AF%20%E5%8C%85%E8%A3%85%20Surge") {
		t.Fatalf("unexpected wrapped Surge URI: %q", got)
	}
}

func TestNormalizeContentSurgeProxyListSOCKSUDPFlag(t *testing.T) {
	raw := `
[Proxy]
首尔 Surge SOCKS UDP = socks5, socks-udp.surge.example.test, 1080, username=qa-user, password=socks-placeholder, udp=true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@socks-udp.surge.example.test:1080?")
	if !strings.Contains(got, "udp=1") || strings.Contains(got, "network=udp") {
		t.Fatalf("unexpected Surge SOCKS UDP URI: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20Surge%20SOCKS%20UDP") {
		t.Fatalf("unexpected Surge SOCKS UDP fragment: %q", got)
	}
}

func TestNormalizeContentSurgeProxyListProtocolAliases(t *testing.T) {
	raw := `
[Proxy]
东京 Surge Trojan-Go Alias = trojan-go, trojan-go.surge-alias.example.test, 443, trojan-placeholder, sni=trojan-go.surge-alias.example.test
香港 Surge HTTP Plus TLS Alias = http+tls, http-plus-tls.surge-alias.example.test, 443, qa-user, http-placeholder, sni=http-plus-tls.surge-alias.example.test
新加坡 Surge HTTP Dash TLS Alias = http-tls, http-dash-tls.surge-alias.example.test, 443, qa-user, http-dash-placeholder, sni=http-dash-tls.surge-alias.example.test
台北 Surge Naive HTTPS Alias = naive-https, naive-https.surge-alias.example.test, 443, qa-user, naive-https-placeholder, sni=naive-https.surge-alias.example.test
首尔 Surge SOCKS5H Alias = socks5h, socks5h.surge-alias.example.test, 1080, qa-user, socks-placeholder, udp=true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 Surge alias URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@trojan-go.surge-alias.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=trojan-go.surge-alias.example.test") ||
		!strings.HasSuffix(lines[0], "#%E4%B8%9C%E4%BA%AC%20Surge%20Trojan-Go%20Alias") {
		t.Fatalf("unexpected Surge Trojan-Go alias URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "https://qa-user:http-placeholder@http-plus-tls.surge-alias.example.test:443?")
	if !strings.Contains(lines[1], "sni=http-plus-tls.surge-alias.example.test") ||
		!strings.HasSuffix(lines[1], "#%E9%A6%99%E6%B8%AF%20Surge%20HTTP%20Plus%20TLS%20Alias") {
		t.Fatalf("unexpected Surge HTTP Plus TLS alias URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "https://qa-user:http-dash-placeholder@http-dash-tls.surge-alias.example.test:443?")
	if !strings.Contains(lines[2], "sni=http-dash-tls.surge-alias.example.test") ||
		!strings.HasSuffix(lines[2], "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20Surge%20HTTP%20Dash%20TLS%20Alias") {
		t.Fatalf("unexpected Surge HTTP Dash TLS alias URI: %q", lines[2])
	}
	assertHasPrefix(t, lines[3], "naive://qa-user:naive-https-placeholder@naive-https.surge-alias.example.test:443?")
	if !strings.Contains(lines[3], "sni=naive-https.surge-alias.example.test") ||
		!strings.HasSuffix(lines[3], "#%E5%8F%B0%E5%8C%97%20Surge%20Naive%20HTTPS%20Alias") {
		t.Fatalf("unexpected Surge Naive HTTPS alias URI: %q", lines[3])
	}
	assertHasPrefix(t, lines[4], "socks5://qa-user:socks-placeholder@socks5h.surge-alias.example.test:1080?")
	if !strings.Contains(lines[4], "udp=1") ||
		!strings.HasSuffix(lines[4], "#%E9%A6%96%E5%B0%94%20Surge%20SOCKS5H%20Alias") {
		t.Fatalf("unexpected Surge SOCKS5H alias URI: %q", lines[4])
	}
}

func TestNormalizeContentSurgeProxyListVLESSAndVMess(t *testing.T) {
	raw := `
[Proxy]
香港 Surge VLESS = vless, vless.surge.example.test, 443, uuid=00000000-0000-0000-0000-000000000089, tls=true, sni=vless.surge.example.test, obfs=ws, obfs-uri=/vless, obfs-host=ws.vless.surge.example.test, flow=xtls-rprx-vision
东京 Surge VMess = vmess, vmess.surge.example.test, 443, username=00000000-0000-0000-0000-000000000090, tls=true, sni=vmess.surge.example.test, ws=true, ws-path=/vmess, ws-host=ws.vmess.surge.example.test, encrypt-method=auto, alter-id=0
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Surge VLESS/VMess URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000089@vless.surge.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=vless.surge.example.test",
		"type=ws",
		"path=%2Fvless",
		"host=ws.vless.surge.example.test",
		"flow=xtls-rprx-vision",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Surge VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20Surge%20VLESS") {
		t.Fatalf("unexpected Surge VLESS fragment: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "vmess://")
	decoded := decodeVMessURIForTest(t, lines[1])
	for _, want := range []string{
		`"ps":"东京 Surge VMess"`,
		`"add":"vmess.surge.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000090"`,
		`"aid":"0"`,
		`"scy":"auto"`,
		`"net":"ws"`,
		`"host":"ws.vmess.surge.example.test"`,
		`"path":"/vmess"`,
		`"tls":"tls"`,
		`"sni":"vmess.surge.example.test"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected Surge VMess document to contain %s: %q", want, decoded)
		}
	}
}

func TestNormalizeContentSurgeProxyListGRPCTransport(t *testing.T) {
	raw := `
[Proxy]
香港 Surge gRPC VLESS = vless, grpc-vless.surge.example.test, 443, uuid=00000000-0000-0000-0000-000000000093, tls=true, sni=grpc-vless.surge.example.test, obfs=grpc, grpc-service-name=fluxgate-surge
东京 Surge gRPC Trojan = trojan, grpc-trojan.surge.example.test, 443, password=trojan-placeholder, sni=grpc-trojan.surge.example.test, obfs=grpc, service-name=trojan-surge
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Surge gRPC URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000093@grpc-vless.surge.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=grpc-vless.surge.example.test",
		"type=grpc",
		"service_name=fluxgate-surge",
		"#%E9%A6%99%E6%B8%AF%20Surge%20gRPC%20VLESS",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Surge gRPC VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@grpc-trojan.surge.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=grpc-trojan.surge.example.test",
		"type=grpc",
		"service_name=trojan-surge",
		"#%E4%B8%9C%E4%BA%AC%20Surge%20gRPC%20Trojan",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Surge gRPC Trojan URI to contain %q: %q", want, lines[1])
		}
	}
}

func TestNormalizeContentSurgeProxyListHysteria2AndTUIC(t *testing.T) {
	raw := `
[Proxy]
首尔 Surge Hysteria2 = hysteria2, hy2.surge.example.test, 443, password=hy2-placeholder, obfs=salamander, obfs-password=obfs-placeholder, up-mbps=80, down-mbps=160, sni=hy2.surge.example.test, alpn=h3, skip-cert-verify=true, disable-sni=true, pinSHA256=surge-hy2-pin
大阪 Surge TUIC = tuic, tuic.surge.example.test, 443, 00000000-0000-0000-0000-000000000091, tuic-placeholder, congestion-controller=bbr, udp-relay-mode=native, sni=tuic.surge.example.test, alpn=h3, skip-cert-verify=true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Surge Hysteria2/TUIC URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "hysteria2://hy2-placeholder@hy2.surge.example.test:443?")
	for _, want := range []string{
		"obfs=salamander",
		"obfs-password=obfs-placeholder",
		"up_mbps=80",
		"down_mbps=160",
		"sni=hy2.surge.example.test",
		"alpn=h3",
		"insecure=1",
		"disable_sni=1",
		"pinSHA256=surge-hy2-pin",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Surge Hysteria2 URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#%E9%A6%96%E5%B0%94%20Surge%20Hysteria2") {
		t.Fatalf("unexpected Surge Hysteria2 fragment: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "tuic://00000000-0000-0000-0000-000000000091:tuic-placeholder@tuic.surge.example.test:443?")
	for _, want := range []string{
		"congestion_control=bbr",
		"udp_relay_mode=native",
		"sni=tuic.surge.example.test",
		"alpn=h3",
		"insecure=1",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Surge TUIC URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#%E5%A4%A7%E9%98%AA%20Surge%20TUIC") {
		t.Fatalf("unexpected Surge TUIC fragment: %q", lines[1])
	}
}

func TestNormalizeContentSurgeProxyListHysteriaAndTLSHelpers(t *testing.T) {
	raw := `
[Proxy]
香港 Surge Hysteria = hysteria, hysteria.surge.example.test, 443, auth-str=hysteria-auth, up-mbps=25, down-mbps=100, obfs=obfs-placeholder, recv-window-conn=1048576, recv-window=4194304, disable-mtu-discovery=true, protocol=udp, sni=hysteria.surge.example.test, alpn=h3, skip-cert-verify=true
新加坡 Surge AnyTLS = anytls, anytls.surge.example.test, 443, password=anytls-placeholder, idle-session-check-interval=20s, idle-session-timeout=45s, min-idle-session=2, sni=anytls.surge.example.test, alpn=h2, skip-cert-verify=true
东京 Surge ShadowTLS = shadow-tls, shadowtls.surge.example.test, 443, password=shadow-placeholder, version=3, sni=shadowtls.surge.example.test, alpn=h2, skip-cert-verify=true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 Surge Hysteria/TLS helper URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "hysteria://hysteria-auth@hysteria.surge.example.test:443?")
	for _, want := range []string{
		"up_mbps=25",
		"down_mbps=100",
		"obfs=obfs-placeholder",
		"recv_window_conn=1048576",
		"recv_window=4194304",
		"disable_mtu_discovery=1",
		"network=udp",
		"sni=hysteria.surge.example.test",
		"alpn=h3",
		"insecure=1",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Surge Hysteria URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20Surge%20Hysteria") {
		t.Fatalf("unexpected Surge Hysteria fragment: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "anytls://anytls-placeholder@anytls.surge.example.test:443?")
	for _, want := range []string{
		"idle_session_check_interval=20s",
		"idle_session_timeout=45s",
		"min_idle_session=2",
		"sni=anytls.surge.example.test",
		"alpn=h2",
		"insecure=1",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Surge AnyTLS URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20Surge%20AnyTLS") {
		t.Fatalf("unexpected Surge AnyTLS fragment: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "shadowtls://shadow-placeholder@shadowtls.surge.example.test:443?")
	for _, want := range []string{
		"version=3",
		"sni=shadowtls.surge.example.test",
		"alpn=h2",
		"insecure=1",
	} {
		if !strings.Contains(lines[2], want) {
			t.Fatalf("expected Surge ShadowTLS URI to contain %q: %q", want, lines[2])
		}
	}
	if !strings.HasSuffix(lines[2], "#%E4%B8%9C%E4%BA%AC%20Surge%20ShadowTLS") {
		t.Fatalf("unexpected Surge ShadowTLS fragment: %q", lines[2])
	}
}

func TestNormalizeContentSurgeProxyListNaiveAndSSH(t *testing.T) {
	raw := `
[Proxy]
新加坡 Surge Naive = naive+quic, naive.surge.example.test, 443, qa-user, naive-placeholder, sni=naive.surge.example.test, alpn=h3, skip-cert-verify=true, quic-congestion-control=bbr, udp-over-tcp=true, insecure-concurrency=2
香港 Surge SSH = ssh, ssh.surge.example.test, 22, qa-user, ssh-placeholder, private-key-path=keys/qa_id_ed25519, host-key-algorithms="ssh-ed25519,rsa-sha2-512", client-version=SSH-2.0-FluxGateQA, cipher=aes128-gcm@openssh.com, mac=hmac-sha2-256, kex-algorithm=curve25519-sha256
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 Surge Naive/SSH URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "naive+quic://qa-user:naive-placeholder@naive.surge.example.test:443?")
	for _, want := range []string{
		"quic=1",
		"quic_congestion_control=bbr",
		"udp_over_tcp=1",
		"insecure_concurrency=2",
		"sni=naive.surge.example.test",
		"alpn=h3",
		"insecure=1",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected Surge Naive URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20Surge%20Naive") {
		t.Fatalf("unexpected Surge Naive fragment: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "ssh://qa-user:ssh-placeholder@ssh.surge.example.test:22?")
	for _, want := range []string{
		"private_key_path=keys%2Fqa_id_ed25519",
		"host_key_algorithms=ssh-ed25519%2Crsa-sha2-512",
		"client_version=SSH-2.0-FluxGateQA",
		"cipher=aes128-gcm%40openssh.com",
		"mac=hmac-sha2-256",
		"kex_algorithm=curve25519-sha256",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected Surge SSH URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#%E9%A6%99%E6%B8%AF%20Surge%20SSH") {
		t.Fatalf("unexpected Surge SSH fragment: %q", lines[1])
	}
}

func TestNormalizeContentSurgeProxyListWireGuard(t *testing.T) {
	raw := `
[Proxy]
台北 Surge WireGuard = wireguard, wg.surge.example.test, 51820, private-key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=, public-key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==, self-ip=10.66.0.3/32, self-ip-v6=fd00::3/128, pre-shared-key=cHNrLXBsYWNlaG9sZGVy, allowed-ips="0.0.0.0/0,::/0", reserved="4,5,6", mtu=1420, udp=true, interface-name=wg-surge, system-interface=true
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "wireguard://wg.surge.example.test:51820?")
	for _, want := range []string{
		"private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI",
		"peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg",
		"local_address=10.66.0.3%2F32%2Cfd00%3A%3A3%2F128",
		"pre_shared_key=cHNrLXBsYWNlaG9sZGVy",
		"allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0",
		"reserved=4%2C5%2C6",
		"mtu=1420",
		"network=udp",
		"interface_name=wg-surge",
		"system_interface=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected Surge WireGuard URI to contain %q: %q", want, got)
		}
	}
	if !strings.HasSuffix(got, "#%E5%8F%B0%E5%8C%97%20Surge%20WireGuard") {
		t.Fatalf("unexpected Surge WireGuard fragment: %q", got)
	}
}

func TestNormalizeContentSurgeProxyListInternalOutbounds(t *testing.T) {
	raw := `
[Proxy]
本地 Surge 直连 = DIRECT
广告 Surge 拦截 = REJECT
静默 Surge 拦截 = reject-drop
内部 Surge DNS = dns
`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	want := []string{
		"direct://default#%E6%9C%AC%E5%9C%B0%20Surge%20%E7%9B%B4%E8%BF%9E",
		"block://default#%E5%B9%BF%E5%91%8A%20Surge%20%E6%8B%A6%E6%88%AA",
		"block://default#%E9%9D%99%E9%BB%98%20Surge%20%E6%8B%A6%E6%88%AA",
		"dns://default#%E5%86%85%E9%83%A8%20Surge%20DNS",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d Surge internal URIs, got %d: %q", len(want), len(lines), got)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf("unexpected Surge internal URI at %d: want %q, got %q", index, expected, lines[index])
		}
	}
}

func TestNormalizeContentSingBoxJSON(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "selector",
      "tag": "auto",
      "outbounds": ["香港 03"]
    },
    {
      "type": "shadowsocks",
      "tag": "香港 03",
      "server": "ss.singbox.example.test",
      "server_port": 8388,
      "method": "aes-128-gcm",
      "password": "qa-placeholder",
      "plugin": "v2ray-plugin",
      "plugin_opts": "mode=websocket;host=ss.singbox.example.test",
      "network": "tcp"
    },
    {
      "type": "trojan",
      "tag": "东京 02",
      "server": "trojan.singbox.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "edge.singbox.example.test",
        "insecure": true,
        "disable_sni": true,
        "alpn": ["h2", "http/1.1"]
      },
      "transport": {
        "type": "grpc",
        "service_name": "trojan-flow"
      }
    },
    {
      "type": "vless",
      "tag": "首尔 05",
      "server": "vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000052",
      "flow": "xtls-rprx-vision",
      "tls": {
        "enabled": true,
        "server_name": "vless.singbox.example.test",
        "insecure": true,
        "disable_sni": true,
        "alpn": ["h3"]
      },
      "transport": {
        "type": "ws",
        "path": "/vless",
        "headers": {
          "Host": ["ws.vless.singbox.example.test", "ws-backup.vless.singbox.example.test"]
        }
      }
    },
    {
      "type": "vmess",
      "tag": "大阪 02",
      "server": "vmess.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000053",
      "security": "auto",
      "tls": {
        "enabled": true,
        "server_name": "vmess.singbox.example.test"
      },
      "transport": {
        "type": "ws",
        "path": "/ws",
        "headers": {
          "Host": "ws.singbox.example.test"
        }
      }
    },
    {
      "type": "hysteria2",
      "tag": "香港 06",
      "server": "hy2.singbox.example.test",
      "server_port": 443,
      "password": "hy2-placeholder",
      "up_mbps": 40,
      "down_mbps": 160,
      "obfs": {
        "type": "salamander",
        "password": "obfs-placeholder"
      },
      "tls": {
        "enabled": true,
        "server_name": "hy2.singbox.example.test",
        "insecure": true,
        "disable_sni": true,
        "alpn": ["h3"],
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        },
        "certificate_public_key_sha256": ["singbox-hy2-pin"]
      }
    },
    {
      "type": "tuic",
      "tag": "东京 06",
      "server": "tuic.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000054",
      "password": "tuic-placeholder",
      "congestion_control": "bbr",
      "udp_relay_mode": "native",
      "tls": {
        "enabled": true,
        "server_name": "tuic.singbox.example.test",
        "alpn": ["h3"],
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    },
    {
      "type": "anytls",
      "tag": "首尔 06",
      "server": "anytls.singbox.example.test",
      "server_port": 443,
      "password": "anytls-placeholder",
      "idle_session_check_interval": "20s",
      "idle_session_timeout": "45s",
      "min_idle_session": 2,
      "tls": {
        "enabled": true,
        "server_name": "anytls.singbox.example.test",
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    },
    {
      "type": "shadowtls",
      "tag": "东京 07",
      "server": "shadow.singbox.example.test",
      "server_port": 443,
      "version": 3,
      "password": "shadow-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "shadow.singbox.example.test",
        "insecure": true,
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    },
    {
      "type": "hysteria",
      "tag": "香港 07",
      "server": "hysteria.singbox.example.test",
      "server_port": 443,
      "auth_str": "hysteria-auth",
      "up_mbps": 20,
      "down_mbps": 80,
      "obfs": "obfs-placeholder",
      "network": "udp",
      "tls": {
        "enabled": true,
        "server_name": "hysteria.singbox.example.test",
        "insecure": true,
        "alpn": ["h3"],
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    },
    {
      "type": "http",
      "tag": "首尔 07",
      "server": "http.singbox.example.test",
      "server_port": 8443,
      "username": "qa-user",
      "password": "http-placeholder",
      "path": "/connect",
      "tls": {
        "enabled": true,
        "server_name": "http.singbox.example.test",
        "insecure": true,
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    },
    {
      "type": "socks",
      "tag": "大阪 07",
      "server": "socks.singbox.example.test",
      "server_port": 1080,
      "username": "qa-user",
      "password": "socks-placeholder",
      "version": "5",
      "network": "udp",
      "udp_over_tcp": true
    },
    {
      "type": "ssh",
      "tag": "香港 08",
      "server": "ssh.singbox.example.test",
      "server_port": 22,
      "user": "qa-user",
      "password": "ssh-placeholder",
      "private_key_path": "keys/qa_id_ed25519",
      "host_key_algorithms": ["ssh-ed25519", "rsa-sha2-512"],
      "client_version": "SSH-2.0-FluxGateQA"
    },
    {
      "type": "wireguard",
      "tag": "台北 02",
      "server": "wg.singbox.example.test",
      "server_port": 51820,
      "private_key": "cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI=",
      "local_address": ["10.66.0.2/32", "fd00::2/128"],
      "peers": [
        {
          "public_key": "cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg==",
          "allowed_ips": ["0.0.0.0/0", "::/0"],
          "pre_shared_key": "cHNrLXBsYWNlaG9sZGVy",
          "reserved": [1, 2, 3]
        }
      ],
      "mtu": 1420
    },
    {
      "type": "tor",
      "tag": "匿名 02",
      "executable_path": "/usr/bin/tor",
      "extra_args": ["--quiet", "--SocksPort", "auto"],
      "data_directory": "cache/tor",
      "torrc": {
        "ClientOnly": "1"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 14 {
		t.Fatalf("expected 14 normalized nodes, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ss.singbox.example.test:8388?")
	if !strings.Contains(lines[0], "plugin=v2ray-plugin") ||
		!strings.Contains(lines[0], "plugin_opts=mode%3Dwebsocket%3Bhost%3Dss.singbox.example.test") ||
		!strings.Contains(lines[0], "network=tcp") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%2003") {
		t.Fatalf("unexpected sing-box shadowsocks URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@trojan.singbox.example.test:443?")
	if !strings.Contains(lines[1], "sni=edge.singbox.example.test") ||
		!strings.Contains(lines[1], "insecure=1") ||
		!strings.Contains(lines[1], "disable_sni=1") ||
		!strings.Contains(lines[1], "alpn=h2%2Chttp%2F1.1") ||
		!strings.Contains(lines[1], "type=grpc") ||
		!strings.Contains(lines[1], "service_name=trojan-flow") {
		t.Fatalf("unexpected sing-box trojan URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "vless://00000000-0000-0000-0000-000000000052@vless.singbox.example.test:443?")
	if !strings.Contains(lines[2], "flow=xtls-rprx-vision") ||
		!strings.Contains(lines[2], "security=tls") ||
		!strings.Contains(lines[2], "sni=vless.singbox.example.test") ||
		!strings.Contains(lines[2], "insecure=1") ||
		!strings.Contains(lines[2], "disable_sni=1") ||
		!strings.Contains(lines[2], "alpn=h3") ||
		!strings.Contains(lines[2], "type=ws") ||
		!strings.Contains(lines[2], "path=%2Fvless") ||
		!strings.Contains(lines[2], "host=ws.vless.singbox.example.test%2Cws-backup.vless.singbox.example.test") {
		t.Fatalf("unexpected sing-box vless URI: %q", lines[2])
	}
	assertHasPrefix(t, lines[3], "vmess://")
	assertHasPrefix(t, lines[4], "hysteria2://hy2-placeholder@hy2.singbox.example.test:443?")
	if !strings.Contains(lines[4], "obfs=salamander") ||
		!strings.Contains(lines[4], "obfs-password=obfs-placeholder") ||
		!strings.Contains(lines[4], "up_mbps=40") ||
		!strings.Contains(lines[4], "down_mbps=160") ||
		!strings.Contains(lines[4], "disable_sni=1") ||
		!strings.Contains(lines[4], "fp=chrome") ||
		!strings.Contains(lines[4], "pinSHA256=singbox-hy2-pin") {
		t.Fatalf("unexpected sing-box hysteria2 URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "tuic://00000000-0000-0000-0000-000000000054:tuic-placeholder@tuic.singbox.example.test:443?")
	if !strings.Contains(lines[5], "congestion_control=bbr") || !strings.Contains(lines[5], "udp_relay_mode=native") || !strings.Contains(lines[5], "fp=chrome") {
		t.Fatalf("unexpected sing-box tuic URI: %q", lines[5])
	}
	assertHasPrefix(t, lines[6], "anytls://anytls-placeholder@anytls.singbox.example.test:443?")
	if !strings.Contains(lines[6], "idle_session_check_interval=20s") || !strings.Contains(lines[6], "min_idle_session=2") || !strings.Contains(lines[6], "fp=chrome") {
		t.Fatalf("unexpected sing-box anytls URI: %q", lines[6])
	}
	assertHasPrefix(t, lines[7], "shadowtls://shadow-placeholder@shadow.singbox.example.test:443?")
	if !strings.Contains(lines[7], "version=3") || !strings.Contains(lines[7], "sni=shadow.singbox.example.test") || !strings.Contains(lines[7], "fp=chrome") {
		t.Fatalf("unexpected sing-box shadowtls URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "hysteria://hysteria-auth@hysteria.singbox.example.test:443?")
	if !strings.Contains(lines[8], "up_mbps=20") || !strings.Contains(lines[8], "down_mbps=80") || !strings.Contains(lines[8], "network=udp") || !strings.Contains(lines[8], "fp=chrome") {
		t.Fatalf("unexpected sing-box hysteria URI: %q", lines[8])
	}
	assertHasPrefix(t, lines[9], "https://qa-user:http-placeholder@http.singbox.example.test:8443/connect?")
	if !strings.Contains(lines[9], "sni=http.singbox.example.test") || !strings.Contains(lines[9], "insecure=1") || !strings.Contains(lines[9], "fp=chrome") {
		t.Fatalf("unexpected sing-box http URI: %q", lines[9])
	}
	assertHasPrefix(t, lines[10], "socks5://qa-user:socks-placeholder@socks.singbox.example.test:1080?")
	if !strings.Contains(lines[10], "network=udp") || !strings.Contains(lines[10], "udp_over_tcp=1") {
		t.Fatalf("unexpected sing-box socks URI: %q", lines[10])
	}
	assertHasPrefix(t, lines[11], "ssh://qa-user:ssh-placeholder@ssh.singbox.example.test:22?")
	if !strings.Contains(lines[11], "private_key_path=keys%2Fqa_id_ed25519") || !strings.Contains(lines[11], "host_key_algorithms=ssh-ed25519%2Crsa-sha2-512") {
		t.Fatalf("unexpected sing-box ssh URI: %q", lines[11])
	}
	assertHasPrefix(t, lines[12], "wireguard://wg.singbox.example.test:51820?")
	if !strings.Contains(lines[12], "private_key=cHJpdmF0ZS1rZXktcGxhY2Vob2xkZXItMzI") ||
		!strings.Contains(lines[12], "peer_public_key=cHVibGljLWtleS1wbGFjZWhvbGRlci0zMg") ||
		!strings.Contains(lines[12], "local_address=10.66.0.2%2F32") ||
		!strings.Contains(lines[12], "allowed_ips=0.0.0.0%2F0") ||
		!strings.Contains(lines[12], "reserved=1%2C2%2C3") {
		t.Fatalf("unexpected sing-box wireguard URI: %q", lines[12])
	}
	assertHasPrefix(t, lines[13], "tor://default?")
	if !strings.Contains(lines[13], "executable_path=%2Fusr%2Fbin%2Ftor") ||
		!strings.Contains(lines[13], "extra_args=--quiet%2C--SocksPort%2Cauto") ||
		!strings.Contains(lines[13], "torrc.ClientOnly=1") {
		t.Fatalf("unexpected sing-box tor URI: %q", lines[13])
	}
}

func TestNormalizeContentSingBoxJSONNaive(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "naive",
      "tag": "新加坡 sing-box Naive",
      "server": "naive.singbox.example.test",
      "server_port": 443,
      "username": "qa-user",
      "password": "naive-placeholder",
      "quic": true,
      "quic_congestion_control": "bbr",
      "udp_over_tcp": true,
      "insecure_concurrency": 2,
      "tls": {
        "enabled": true,
        "server_name": "naive.singbox.example.test",
        "insecure": true,
        "disable_sni": true,
        "alpn": ["h3"],
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "naive+quic://qa-user:naive-placeholder@naive.singbox.example.test:443?")
	for _, want := range []string{
		"quic=1",
		"quic_congestion_control=bbr",
		"udp_over_tcp=1",
		"insecure_concurrency=2",
		"sni=naive.singbox.example.test",
		"insecure=1",
		"disable_sni=1",
		"alpn=h3",
		"fp=chrome",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected sing-box Naive URI to contain %q: %q", want, got)
		}
	}
	if !strings.HasSuffix(got, "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20sing-box%20Naive") {
		t.Fatalf("unexpected sing-box Naive fragment: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONProtocolAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "trojan-go",
      "tag": "sing-box Trojan-Go Alias",
      "server": "trojan-go.singbox-alias.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "trojan-go.singbox-alias.example.test"
      }
    },
    {
      "type": "vmess-aead",
      "tag": "sing-box VMess AEAD Alias",
      "server": "vmess-aead.singbox-alias.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000056",
      "tls": {
        "enabled": true,
        "server_name": "vmess-aead.singbox-alias.example.test"
      }
    },
    {
      "type": "hy2",
      "tag": "sing-box HY2 Alias",
      "server": "hy2.singbox-alias.example.test",
      "server_port": 443,
      "password": "hy2-placeholder"
    },
    {
      "type": "any-tls",
      "tag": "sing-box AnyTLS Alias",
      "server": "anytls.singbox-alias.example.test",
      "server_port": 443,
      "password": "anytls-placeholder"
    },
    {
      "type": "shadow-tls",
      "tag": "sing-box ShadowTLS Alias",
      "server": "shadowtls.singbox-alias.example.test",
      "server_port": 443,
      "version": 3,
      "password": "shadowtls-placeholder"
    },
    {
      "type": "http+tls",
      "tag": "sing-box HTTP Plus TLS Alias",
      "server": "http-plus-tls.singbox-alias.example.test",
      "server_port": 443,
      "username": "qa-user",
      "password": "http-placeholder",
      "path": "/connect"
    },
    {
      "type": "http-tls",
      "tag": "sing-box HTTP Dash TLS Alias",
      "server": "http-dash-tls.singbox-alias.example.test",
      "server_port": 443,
      "username": "qa-user",
      "password": "http-dash-placeholder"
    },
    {
      "type": "naive-quic",
      "tag": "sing-box Naive QUIC Alias",
      "server": "naive.singbox-alias.example.test",
      "server_port": 443,
      "username": "qa-user",
      "password": "naive-placeholder"
    },
    {
      "type": "naive-https",
      "tag": "sing-box Naive HTTPS Alias",
      "server": "naive-https.singbox-alias.example.test",
      "server_port": 443,
      "username": "qa-user",
      "password": "naive-https-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "naive-https.singbox-alias.example.test"
      }
    },
    {
      "type": "socks5h",
      "tag": "sing-box SOCKS5H Alias",
      "server": "socks5h.singbox-alias.example.test",
      "server_port": 1080,
      "username": "qa-user",
      "password": "socks-placeholder",
      "network": "udp"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 normalized sing-box alias nodes, got %d: %q", len(lines), got)
	}

	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@trojan-go.singbox-alias.example.test:443?")
	if !strings.Contains(lines[0], "sni=trojan-go.singbox-alias.example.test") ||
		!strings.HasSuffix(lines[0], "#sing-box%20Trojan-Go%20Alias") {
		t.Fatalf("unexpected sing-box Trojan-Go alias URI: %q", lines[0])
	}

	assertHasPrefix(t, lines[1], "vmess://")
	decodedVMess := decodeVMessURIForTest(t, lines[1])
	for _, want := range []string{
		`"ps":"sing-box VMess AEAD Alias"`,
		`"add":"vmess-aead.singbox-alias.example.test"`,
		`"id":"00000000-0000-0000-0000-000000000056"`,
		`"tls":"tls"`,
		`"sni":"vmess-aead.singbox-alias.example.test"`,
	} {
		if !strings.Contains(decodedVMess, want) {
			t.Fatalf("expected decoded VMess alias URI to contain %q: %q", want, decodedVMess)
		}
	}

	assertHasPrefix(t, lines[2], "hysteria2://hy2-placeholder@hy2.singbox-alias.example.test:443#sing-box%20HY2%20Alias")
	assertHasPrefix(t, lines[3], "anytls://anytls-placeholder@anytls.singbox-alias.example.test:443#sing-box%20AnyTLS%20Alias")
	assertHasPrefix(t, lines[4], "shadowtls://shadowtls-placeholder@shadowtls.singbox-alias.example.test:443?")
	if !strings.Contains(lines[4], "version=3") ||
		!strings.HasSuffix(lines[4], "#sing-box%20ShadowTLS%20Alias") {
		t.Fatalf("unexpected sing-box ShadowTLS alias URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "https://qa-user:http-placeholder@http-plus-tls.singbox-alias.example.test:443/connect#sing-box%20HTTP%20Plus%20TLS%20Alias")
	assertHasPrefix(t, lines[6], "https://qa-user:http-dash-placeholder@http-dash-tls.singbox-alias.example.test:443#sing-box%20HTTP%20Dash%20TLS%20Alias")
	assertHasPrefix(t, lines[7], "naive+quic://qa-user:naive-placeholder@naive.singbox-alias.example.test:443?")
	if !strings.Contains(lines[7], "quic=1") ||
		!strings.HasSuffix(lines[7], "#sing-box%20Naive%20QUIC%20Alias") {
		t.Fatalf("unexpected sing-box Naive QUIC alias URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "naive://qa-user:naive-https-placeholder@naive-https.singbox-alias.example.test:443?")
	if !strings.Contains(lines[8], "sni=naive-https.singbox-alias.example.test") ||
		!strings.HasSuffix(lines[8], "#sing-box%20Naive%20HTTPS%20Alias") {
		t.Fatalf("unexpected sing-box Naive HTTPS alias URI: %q", lines[8])
	}
	assertHasPrefix(t, lines[9], "socks5://qa-user:socks-placeholder@socks5h.singbox-alias.example.test:1080?")
	if !strings.Contains(lines[9], "network=udp") ||
		!strings.HasSuffix(lines[9], "#sing-box%20SOCKS5H%20Alias") {
		t.Fatalf("unexpected sing-box SOCKS5H alias URI: %q", lines[9])
	}
}

func TestNormalizeContentSingBoxJSONFieldCaseAliases(t *testing.T) {
	raw := `{
  "Outbounds": [
    {
      "Type": "trojan",
      "Tag": "sing-box Trojan Field Aliases",
      "Server": "trojan-field.singbox-alias.example.test",
      "Server-Port": 443,
      "Password": "trojan-placeholder",
      "TLS": {
        "Enabled": true,
        "ServerName": "trojan-field.singbox-alias.example.test",
        "Insecure": true,
        "Disable-SNI": true,
        "ALPN": ["h2", "http/1.1"],
        "UTLS": {
          "Enabled": true,
          "Fingerprint": "chrome"
        }
      },
      "Transport": {
        "Type": "grpc",
        "ServiceName": "fluxgate",
        "Idle-Timeout": "30s",
        "PingTimeout": "10s",
        "PermitWithoutStream": true
      }
    },
    {
      "Type": "vless",
      "Name": "sing-box VLESS Field Aliases",
      "Server": "vless-field.singbox-alias.example.test",
      "ServerPort": "443",
      "UUID": "00000000-0000-0000-0000-000000000057",
      "PacketEncoding": "xudp",
      "TLS": {
        "Enabled": true,
        "Server_Name": "vless-field.singbox-alias.example.test"
      },
      "Transport": {
        "Type": "ws",
        "Path": "/ws",
        "MaxEarlyData": 2048,
        "EarlyDataHeaderName": "Sec-WebSocket-Protocol",
        "Headers": {
          "Host": ["ws-field.singbox-alias.example.test", "ws-backup.singbox-alias.example.test"]
        }
      }
    }
  ],
  "Endpoints": {
    "Type": "wireguard",
    "Tag": "sing-box WireGuard Endpoint Field Aliases",
    "SystemInterface": true,
    "InterfaceName": "wg-alias",
    "Address": ["10.66.0.9/32", "fd00::9/128"],
    "PrivateKey": "endpoint-private",
    "Peers": [
      {
        "Address": "wg-field.singbox-alias.example.test",
        "ServerPort": 51820,
        "PublicKey": "endpoint-peer",
        "PreSharedKey": "endpoint-psk",
        "AllowedIPs": ["0.0.0.0/0", "::/0"],
        "Reserved": [9, 8, 7]
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 normalized sing-box field alias nodes, got %d: %q", len(lines), got)
	}

	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@trojan-field.singbox-alias.example.test:443?")
	for _, want := range []string{
		"sni=trojan-field.singbox-alias.example.test",
		"insecure=1",
		"disable_sni=1",
		"alpn=h2%2Chttp%2F1.1",
		"fp=chrome",
		"type=grpc",
		"service_name=fluxgate",
		"idle_timeout=30s",
		"ping_timeout=10s",
		"permit_without_stream=1",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected sing-box Trojan field alias URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#sing-box%20Trojan%20Field%20Aliases") {
		t.Fatalf("unexpected sing-box Trojan field alias fragment: %q", lines[0])
	}

	assertHasPrefix(t, lines[1], "vless://00000000-0000-0000-0000-000000000057@vless-field.singbox-alias.example.test:443?")
	for _, want := range []string{
		"packet_encoding=xudp",
		"security=tls",
		"sni=vless-field.singbox-alias.example.test",
		"type=ws",
		"path=%2Fws",
		"max_early_data=2048",
		"early_data_header_name=Sec-WebSocket-Protocol",
		"host=ws-field.singbox-alias.example.test%2Cws-backup.singbox-alias.example.test",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected sing-box VLESS field alias URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#sing-box%20VLESS%20Field%20Aliases") {
		t.Fatalf("unexpected sing-box VLESS field alias fragment: %q", lines[1])
	}

	assertHasPrefix(t, lines[2], "wireguard://wg-field.singbox-alias.example.test:51820?")
	for _, want := range []string{
		"private_key=endpoint-private",
		"peer_public_key=endpoint-peer",
		"pre_shared_key=endpoint-psk",
		"local_address=10.66.0.9%2F32%2Cfd00%3A%3A9%2F128",
		"allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0",
		"reserved=9%2C8%2C7",
		"system_interface=1",
		"interface_name=wg-alias",
	} {
		if !strings.Contains(lines[2], want) {
			t.Fatalf("expected sing-box WireGuard endpoint field alias URI to contain %q: %q", want, lines[2])
		}
	}
	if !strings.HasSuffix(lines[2], "#sing-box%20WireGuard%20Endpoint%20Field%20Aliases") {
		t.Fatalf("unexpected sing-box WireGuard endpoint field alias fragment: %q", lines[2])
	}
}

func TestNormalizeContentSingBoxJSONSOCKSUDPFlag(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "socks",
      "tag": "首尔 sing-box SOCKS UDP",
      "server": "socks-udp.singbox.example.test",
      "server_port": 1080,
      "username": "qa-user",
      "password": "socks-placeholder",
      "udp": true
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@socks-udp.singbox.example.test:1080?")
	if !strings.Contains(got, "udp=1") || strings.Contains(got, "network=udp") {
		t.Fatalf("unexpected sing-box SOCKS UDP URI: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20sing-box%20SOCKS%20UDP") {
		t.Fatalf("unexpected sing-box SOCKS UDP fragment: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONVLESSGRPC(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vless",
      "tag": "新加坡 gRPC",
      "server": "grpc.vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000072",
      "tls": {
        "enabled": true,
        "server_name": "grpc.vless.singbox.example.test"
      },
      "transport": {
        "type": "grpc",
        "service_name": "fluxgate"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000072@grpc.vless.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=grpc.vless.singbox.example.test") ||
		!strings.Contains(got, "type=grpc") ||
		!strings.Contains(got, "service_name=fluxgate") ||
		!strings.HasSuffix(got, "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20gRPC") {
		t.Fatalf("unexpected sing-box vless grpc URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONTrojanGRPCKeepaliveOptions(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "trojan",
      "tag": "东京 gRPC Keepalive",
      "server": "grpc.trojan.singbox.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "grpc.trojan.singbox.example.test"
      },
      "transport": {
        "type": "grpc",
        "service_name": "fluxgate",
        "idle_timeout": "30s",
        "ping_timeout": "10s",
        "permit_without_stream": true
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@grpc.trojan.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=grpc.trojan.singbox.example.test") ||
		!strings.Contains(got, "type=grpc") ||
		!strings.Contains(got, "service_name=fluxgate") ||
		!strings.Contains(got, "idle_timeout=30s") ||
		!strings.Contains(got, "ping_timeout=10s") ||
		!strings.Contains(got, "permit_without_stream=1") {
		t.Fatalf("unexpected sing-box trojan grpc keepalive URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONVLESSQUICTransport(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vless",
      "tag": "新加坡 QUIC",
      "server": "quic.vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000079",
      "tls": {
        "enabled": true,
        "server_name": "quic.vless.singbox.example.test"
      },
      "transport": {
        "type": "quic"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000079@quic.vless.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=quic.vless.singbox.example.test") ||
		!strings.Contains(got, "type=quic") ||
		!strings.HasSuffix(got, "#%E6%96%B0%E5%8A%A0%E5%9D%A1%20QUIC") {
		t.Fatalf("unexpected sing-box vless quic URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONDNS(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "dns",
      "tag": "内部 DNS"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "dns://default#%E5%86%85%E9%83%A8%20DNS" {
		t.Fatalf("unexpected sing-box dns URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONDirect(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "direct",
      "tag": "本地直连"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "direct://default#%E6%9C%AC%E5%9C%B0%E7%9B%B4%E8%BF%9E" {
		t.Fatalf("unexpected sing-box direct URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONObjectMaps(t *testing.T) {
	raw := `{
  "outbounds": {
    "ss-map": {
      "type": "shadowsocks",
      "server": "ss.map.example.test",
      "server_port": 8388,
      "method": "aes-128-gcm",
      "password": "qa-placeholder"
    },
    "direct-map": {
      "type": "direct"
    }
  },
  "endpoints": {
    "wg-map": {
      "type": "wireguard",
      "address": "10.66.0.3/32",
      "private_key": "map-private",
      "peers": [
        {
          "address": "wg.map.example.test",
          "port": 51820,
          "public_key": "map-peer",
          "allowed_ips": "0.0.0.0/0,::/0"
        }
      ]
    }
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "direct://default#direct-map" {
		t.Fatalf("unexpected direct map URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:qa-placeholder@ss.map.example.test:8388#ss-map" {
		t.Fatalf("unexpected shadowsocks map URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "wireguard://wg.map.example.test:51820?")
	if !strings.Contains(lines[2], "private_key=map-private") ||
		!strings.Contains(lines[2], "peer_public_key=map-peer") ||
		!strings.Contains(lines[2], "local_address=10.66.0.3%2F32") ||
		!strings.Contains(lines[2], "allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0") ||
		!strings.HasSuffix(lines[2], "#wg-map") {
		t.Fatalf("unexpected wireguard endpoint map URI: %q", lines[2])
	}
}

func TestNormalizeContentSingBoxJSONObjectMapGroups(t *testing.T) {
	raw := `{
  "outbounds": {
    "asia-group": [
      {
        "type": "direct"
      },
      {
        "type": "shadowsocks",
        "server": "ss.group.example.test",
        "server_port": 8388,
        "method": "aes-128-gcm",
        "password": "qa-placeholder"
      }
    ]
  },
  "endpoints": {
    "wg-group": [
      {
        "type": "wireguard",
        "address": "10.66.0.4/32",
        "private_key": "group-private",
        "peers": [
          {
            "address": "wg.group.example.test",
            "port": 51820,
            "public_key": "group-peer"
          }
        ]
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "direct://default#asia-group-1" {
		t.Fatalf("unexpected grouped direct URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:qa-placeholder@ss.group.example.test:8388#asia-group-2" {
		t.Fatalf("unexpected grouped shadowsocks URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "wireguard://wg.group.example.test:51820?")
	if !strings.Contains(lines[2], "private_key=group-private") ||
		!strings.Contains(lines[2], "peer_public_key=group-peer") ||
		!strings.Contains(lines[2], "local_address=10.66.0.4%2F32") ||
		!strings.HasSuffix(lines[2], "#wg-group") {
		t.Fatalf("unexpected grouped wireguard endpoint URI: %q", lines[2])
	}
}

func TestNormalizeContentSingBoxJSONBlock(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "block",
      "tag": "内部拦截"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "block://default#%E5%86%85%E9%83%A8%E6%8B%A6%E6%88%AA" {
		t.Fatalf("unexpected sing-box block URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONWireGuardEndpoint(t *testing.T) {
	raw := `{
  "endpoints": {
    "type": "wireguard",
    "tag": "wg-ep",
    "system": true,
    "name": "wg-endpoint",
    "mtu": 1420,
    "workers": 2,
    "address": ["10.66.0.2/32", "fd00::2/128"],
    "private_key": "endpoint-private",
    "peers": [
      {
        "address": "wg.endpoint.example.test",
        "port": 51820,
        "public_key": "endpoint-peer",
        "pre_shared_key": "endpoint-psk",
        "allowed_ips": ["0.0.0.0/0", "::/0"],
        "reserved": [7, 8, 9]
      }
    ]
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "wireguard://wg.endpoint.example.test:51820?")
	if !strings.Contains(got, "private_key=endpoint-private") ||
		!strings.Contains(got, "peer_public_key=endpoint-peer") ||
		!strings.Contains(got, "pre_shared_key=endpoint-psk") ||
		!strings.Contains(got, "local_address=10.66.0.2%2F32%2Cfd00%3A%3A2%2F128") ||
		!strings.Contains(got, "allowed_ips=0.0.0.0%2F0%2C%3A%3A%2F0") ||
		!strings.Contains(got, "reserved=7%2C8%2C9") ||
		!strings.Contains(got, "system_interface=1") ||
		!strings.Contains(got, "interface_name=wg-endpoint") ||
		!strings.Contains(got, "mtu=1420") ||
		!strings.Contains(got, "workers=2") ||
		!strings.HasSuffix(got, "#wg-ep") {
		t.Fatalf("unexpected sing-box wireguard endpoint URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONVLESSHTTPTransport(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vless",
      "tag": "新加坡 HTTP",
      "server": "http.vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000078",
      "tls": {
        "enabled": true,
        "server_name": "http.vless.singbox.example.test"
      },
      "transport": {
        "type": "http",
        "host": ["h2.singbox.example.test", "h2-backup.singbox.example.test"],
        "path": "/h2",
        "method": "GET",
        "idle_timeout": "20s",
        "ping_timeout": "10s"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000078@http.vless.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=http.vless.singbox.example.test") ||
		!strings.Contains(got, "type=http") ||
		!strings.Contains(got, "host=h2.singbox.example.test%2Ch2-backup.singbox.example.test") ||
		!strings.Contains(got, "path=%2Fh2") ||
		!strings.Contains(got, "method=GET") ||
		!strings.Contains(got, "idle_timeout=20s") ||
		!strings.Contains(got, "ping_timeout=10s") {
		t.Fatalf("unexpected sing-box vless http transport URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONTrojanWebSocketEarlyData(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "trojan",
      "tag": "东京 WS Early",
      "server": "ws.trojan.singbox.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "ws.trojan.singbox.example.test"
      },
      "transport": {
        "type": "ws",
        "path": "/ws",
        "headers": {
          "Host": "ws.singbox.example.test"
        },
        "max_early_data": 2048,
        "early_data_header_name": "Sec-WebSocket-Protocol"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@ws.trojan.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=ws.trojan.singbox.example.test") ||
		!strings.Contains(got, "type=ws") ||
		!strings.Contains(got, "path=%2Fws") ||
		!strings.Contains(got, "host=ws.singbox.example.test") ||
		!strings.Contains(got, "max_early_data=2048") ||
		!strings.Contains(got, "early_data_header_name=Sec-WebSocket-Protocol") {
		t.Fatalf("unexpected sing-box trojan websocket early data URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONTrojanHTTPUpgradeTransport(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "trojan",
      "tag": "东京 HTTPUpgrade",
      "server": "upgrade.trojan.singbox.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "upgrade.trojan.singbox.example.test"
      },
      "transport": {
        "type": "httpupgrade",
        "path": "/upgrade",
        "headers": {
          "Host": ["upgrade.singbox.example.test", "upgrade-backup.singbox.example.test"]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@upgrade.trojan.singbox.example.test:443?")
	if !strings.Contains(got, "security=tls") ||
		!strings.Contains(got, "sni=upgrade.trojan.singbox.example.test") ||
		!strings.Contains(got, "type=httpupgrade") ||
		!strings.Contains(got, "host=upgrade.singbox.example.test%2Cupgrade-backup.singbox.example.test") ||
		!strings.Contains(got, "path=%2Fupgrade") {
		t.Fatalf("unexpected sing-box trojan httpupgrade URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONVLESSReality(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vless",
      "tag": "新加坡 Reality",
      "server": "reality.vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000076",
      "flow": "xtls-rprx-vision",
      "tls": {
        "enabled": true,
        "server_name": "www.example.test",
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        },
        "reality": {
          "enabled": true,
          "public_key": "reality-public-key-placeholder",
          "short_id": "a1b2c3d4"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000076@reality.vless.singbox.example.test:443?")
	if !strings.Contains(got, "security=reality") ||
		!strings.Contains(got, "pbk=reality-public-key-placeholder") ||
		!strings.Contains(got, "sid=a1b2c3d4") ||
		!strings.Contains(got, "fp=chrome") ||
		!strings.Contains(got, "sni=www.example.test") ||
		!strings.Contains(got, "flow=xtls-rprx-vision") {
		t.Fatalf("unexpected sing-box vless reality URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONTrojanReality(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "trojan",
      "tag": "香港 Trojan Reality",
      "server": "reality.trojan.singbox.example.test",
      "server_port": 443,
      "password": "trojan-placeholder",
      "tls": {
        "enabled": true,
        "server_name": "www.example.test",
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        },
        "reality": {
          "enabled": true,
          "public_key": "trojan-reality-public-key",
          "short_id": "b1c2d3e4"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@reality.trojan.singbox.example.test:443?")
	if !strings.Contains(got, "security=reality") ||
		!strings.Contains(got, "pbk=trojan-reality-public-key") ||
		!strings.Contains(got, "sid=b1c2d3e4") ||
		!strings.Contains(got, "fp=chrome") ||
		!strings.Contains(got, "sni=www.example.test") {
		t.Fatalf("unexpected sing-box trojan reality URI: %q", got)
	}
}

func TestNormalizeContentSingBoxJSONVMessGRPC(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vmess",
      "tag": "东京 VMess gRPC",
      "server": "grpc.vmess.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000074",
      "tls": {
        "enabled": true,
        "server_name": "grpc.vmess.singbox.example.test"
      },
      "transport": {
        "type": "grpc",
        "service_name": "fluxgate-vmess"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vmess://")
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode sing-box vmess grpc URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"net":"grpc"`) ||
		!strings.Contains(decodedText, `"path":"fluxgate-vmess"`) ||
		!strings.Contains(decodedText, `"tls":"tls"`) ||
		!strings.Contains(decodedText, `"sni":"grpc.vmess.singbox.example.test"`) {
		t.Fatalf("unexpected sing-box vmess grpc document: %q", decodedText)
	}
}

func TestNormalizeContentSingBoxJSONVLESSAndVMessPacketEncoding(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vless",
      "tag": "首尔 VLESS Packet Encoding",
      "server": "packet.vless.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000097",
      "packet_encoding": "xudp"
    },
    {
      "type": "vmess",
      "tag": "东京 VMess Packet Encoding",
      "server": "packet.vmess.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000098",
      "security": "auto",
      "packetEncoding": "packetaddr",
      "tls": {
        "enabled": true,
        "server_name": "packet.vmess.singbox.example.test",
        "disable_sni": true,
        "utls": {
          "enabled": true,
          "fingerprint": "chrome"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000097@packet.vless.singbox.example.test:443?")
	if !strings.Contains(lines[0], "packet_encoding=xudp") {
		t.Fatalf("unexpected sing-box vless packet encoding URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "vmess://")
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[1], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode sing-box vmess URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"packet_encoding":"packetaddr"`) ||
		!strings.Contains(decodedText, `"tls":"tls"`) ||
		!strings.Contains(decodedText, `"disable_sni":"1"`) ||
		!strings.Contains(decodedText, `"fp":"chrome"`) ||
		!strings.Contains(decodedText, `"add":"packet.vmess.singbox.example.test"`) {
		t.Fatalf("unexpected sing-box vmess packet encoding document: %q", decodedText)
	}
}

func TestNormalizeContentSingBoxJSONVMessHTTPTransport(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "type": "vmess",
      "tag": "香港 VMess HTTP",
      "server": "http.vmess.singbox.example.test",
      "server_port": 443,
      "uuid": "00000000-0000-0000-0000-000000000089",
      "tls": {
        "enabled": true,
        "server_name": "http.vmess.singbox.example.test"
      },
      "transport": {
        "type": "http",
        "host": ["h2.vmess.singbox.example.test", "h2-backup.vmess.singbox.example.test"],
        "path": "/h2"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vmess://")
	decoded := decodeVMessURIForTest(t, got)
	for _, want := range []string{
		`"ps":"香港 VMess HTTP"`,
		`"net":"http"`,
		`"host":"h2.vmess.singbox.example.test,h2-backup.vmess.singbox.example.test"`,
		`"path":"/h2"`,
		`"tls":"tls"`,
		`"sni":"http.vmess.singbox.example.test"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected sing-box vmess http document to contain %q: %q", want, decoded)
		}
	}
}

func TestNormalizeContentV2RayJSON(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray VMess",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "vmess.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000083",
                "alterId": 0,
                "security": "auto"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "tlsSettings": {
          "serverName": "vmess.v2ray.example.test",
          "allowInsecure": true,
          "disable_sni": true,
          "clientFingerprint": "chrome",
          "alpn": ["h2", "http/1.1"]
        },
        "wsSettings": {
          "path": "/ws",
          "headers": {
            "Host": ["ws.v2ray.example.test", "ws-backup.v2ray.example.test"]
          }
        }
      }
    },
    {
      "tag": "新加坡 V2Ray VLESS",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "vless.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000084",
                "flow": "xtls-rprx-vision"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "realitySettings": {
          "serverName": "www.example.test",
          "publicKey": "v2ray-reality-public-key",
          "shortId": "aabbccdd",
          "fingerprint": "chrome"
        }
      }
    },
    {
      "tag": "东京 V2Ray Trojan",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "trojan.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "network": "grpc",
        "security": "tls",
        "tlsSettings": {
          "serverName": "trojan.v2ray.example.test",
          "clientFingerprint": "chrome",
          "skip-cert-verify": true,
          "disable_sni": true
        },
        "grpcSettings": {
          "serviceName": "fluxgate-v2ray",
          "idle_timeout": "30s",
          "health_check_timeout": "10s",
          "permit_without_stream": true,
          "multiMode": true
        }
      }
    },
    {
      "tag": "首尔 V2Ray SS",
      "protocol": "shadowsocks",
      "settings": {
        "servers": [
          {
            "address": "ss.v2ray.example.test",
            "port": 8388,
            "method": "aes-128-gcm",
            "password": "qa-placeholder"
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vmess://")
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"net":"ws"`) ||
		!strings.Contains(decodedText, `"path":"/ws"`) ||
		!strings.Contains(decodedText, `"host":"ws.v2ray.example.test,ws-backup.v2ray.example.test"`) ||
		!strings.Contains(decodedText, `"tls":"tls"`) ||
		!strings.Contains(decodedText, `"allowInsecure":"1"`) ||
		!strings.Contains(decodedText, `"disable_sni":"1"`) ||
		!strings.Contains(decodedText, `"fp":"chrome"`) ||
		!strings.Contains(decodedText, `"sni":"vmess.v2ray.example.test"`) ||
		!strings.Contains(decodedText, `"alpn":"h2,http/1.1"`) {
		t.Fatalf("unexpected v2ray vmess document: %q", decodedText)
	}
	assertHasPrefix(t, lines[1], "vless://00000000-0000-0000-0000-000000000084@vless.v2ray.example.test:443?")
	if !strings.Contains(lines[1], "security=reality") ||
		!strings.Contains(lines[1], "pbk=v2ray-reality-public-key") ||
		!strings.Contains(lines[1], "sid=aabbccdd") ||
		!strings.Contains(lines[1], "fp=chrome") ||
		!strings.Contains(lines[1], "sni=www.example.test") ||
		!strings.Contains(lines[1], "flow=xtls-rprx-vision") {
		t.Fatalf("unexpected v2ray vless URI: %q", lines[1])
	}
	assertHasPrefix(t, lines[2], "trojan://trojan-placeholder@trojan.v2ray.example.test:443?")
	if !strings.Contains(lines[2], "security=tls") ||
		!strings.Contains(lines[2], "type=grpc") ||
		!strings.Contains(lines[2], "service_name=fluxgate-v2ray") ||
		!strings.Contains(lines[2], "idle_timeout=30s") ||
		!strings.Contains(lines[2], "ping_timeout=10s") ||
		!strings.Contains(lines[2], "permit_without_stream=1") ||
		!strings.Contains(lines[2], "insecure=1") ||
		!strings.Contains(lines[2], "disable_sni=1") ||
		!strings.Contains(lines[2], "fp=chrome") ||
		!strings.Contains(lines[2], "multi_mode=1") {
		t.Fatalf("unexpected v2ray trojan URI: %q", lines[2])
	}
	if lines[3] != "ss://aes-128-gcm:qa-placeholder@ss.v2ray.example.test:8388#%E9%A6%96%E5%B0%94%20V2Ray%20SS" {
		t.Fatalf("unexpected v2ray shadowsocks URI: %q", lines[3])
	}
}

func TestNormalizeContentV2RayJSONFieldCaseAliases(t *testing.T) {
	raw := `{
  "Outbounds": [
    {
      "Tag": "香港 V2Ray Field Aliases",
      "Protocol": "vless",
      "Settings": {
        "VNext": [
          {
            "Address": "field-alias.v2ray.example.test",
            "Server-Port": "443",
            "Users": [
              {
                "ID": "00000000-0000-0000-0000-000000000095",
                "Flow": "xtls-rprx-vision",
                "PacketEncoding": "xudp"
              }
            ]
          }
        ]
      },
      "StreamSettings": {
        "Network": "ws",
        "Security": "TLS",
        "TLSSettings": {
          "ServerName": "field-alias.v2ray.example.test",
          "AllowInsecure": true,
          "Disable-SNI": true,
          "ALPN": ["h2", "http/1.1"],
          "UTLS": {
            "Fingerprint": "chrome"
          }
        },
        "WSSettings": {
          "Path": "/ws",
          "Headers": {
            "Host": ["ws-field.v2ray.example.test", "ws-backup.v2ray.example.test"]
          },
          "MaxEarlyData": 2048,
          "EarlyDataHeaderName": "Sec-WebSocket-Protocol"
        }
      }
    },
    {
      "Tag": "东京 V2Ray gRPC Field Aliases",
      "Protocol": "trojan",
      "Settings": {
        "Servers": [
          {
            "Address": "grpc-field.v2ray.example.test",
            "Port": 443,
            "Password": "trojan-placeholder"
          }
        ]
      },
      "StreamSettings": {
        "Network": "grpc",
        "Security": "TLS",
        "TLSSettings": {
          "Server-Name": "grpc-field.v2ray.example.test"
        },
        "GrpcSettings": {
          "ServiceName": "fluxgate",
          "Idle-Timeout": "30s",
          "PingTimeout": "10s",
          "PermitWithoutStream": true,
          "MultiMode": true
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 V2Ray field alias URIs, got %d: %q", len(lines), got)
	}

	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000095@field-alias.v2ray.example.test:443?")
	for _, want := range []string{
		"flow=xtls-rprx-vision",
		"packet_encoding=xudp",
		"security=tls",
		"sni=field-alias.v2ray.example.test",
		"insecure=1",
		"disable_sni=1",
		"alpn=h2%2Chttp%2F1.1",
		"fp=chrome",
		"type=ws",
		"path=%2Fws",
		"host=ws-field.v2ray.example.test%2Cws-backup.v2ray.example.test",
		"max_early_data=2048",
		"early_data_header_name=Sec-WebSocket-Protocol",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected V2Ray field alias VLESS URI to contain %q: %q", want, lines[0])
		}
	}
	if !strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20V2Ray%20Field%20Aliases") {
		t.Fatalf("unexpected V2Ray field alias VLESS fragment: %q", lines[0])
	}

	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@grpc-field.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=grpc-field.v2ray.example.test",
		"type=grpc",
		"service_name=fluxgate",
		"idle_timeout=30s",
		"ping_timeout=10s",
		"permit_without_stream=1",
		"multi_mode=1",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected V2Ray field alias Trojan URI to contain %q: %q", want, lines[1])
		}
	}
	if !strings.HasSuffix(lines[1], "#%E4%B8%9C%E4%BA%AC%20V2Ray%20gRPC%20Field%20Aliases") {
		t.Fatalf("unexpected V2Ray field alias Trojan fragment: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONDirectEndpointFields(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray VMess 扁平",
      "protocol": "vmess",
      "server": "vmess-direct.v2ray.example.test",
      "serverPort": 443,
      "uuid": "00000000-0000-0000-0000-000000000107",
      "alterId": 0,
      "security": "auto",
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "tlsSettings": {
          "serverName": "vmess-direct.v2ray.example.test"
        },
        "wsSettings": {
          "path": "/direct"
        }
      }
    },
    {
      "tag": "新加坡 V2Ray VLESS Settings 扁平",
      "protocol": "vless",
      "settings": {
        "address": "vless-direct.v2ray.example.test",
        "server-port": 8443,
        "user-id": "00000000-0000-0000-0000-000000000108",
        "flow": "xtls-rprx-vision"
      }
    },
    {
      "tag": "东京 V2Ray Trojan 扁平",
      "protocol": "trojan",
      "address": "trojan-direct.v2ray.example.test",
      "port": 443,
      "pass": "trojan-placeholder"
    },
    {
      "tag": "首尔 V2Ray SS Settings 扁平",
      "protocol": "shadowsocks",
      "settings": {
        "host": "ss-direct.v2ray.example.test",
        "portNumber": 8388,
        "encryptMethod": "aes-256-gcm",
        "passwd": "ss-placeholder"
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 V2Ray direct endpoint URIs, got %d: %q", len(lines), got)
	}
	decoded := decodeVMessURIForTest(t, lines[0])
	for _, want := range []string{
		`"ps":"香港 V2Ray VMess 扁平"`,
		`"add":"vmess-direct.v2ray.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000107"`,
		`"net":"ws"`,
		`"path":"/direct"`,
		`"tls":"tls"`,
		`"sni":"vmess-direct.v2ray.example.test"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected V2Ray direct VMess URI to contain %q: %q", want, decoded)
		}
	}
	if lines[1] != "vless://00000000-0000-0000-0000-000000000108@vless-direct.v2ray.example.test:8443?flow=xtls-rprx-vision#%E6%96%B0%E5%8A%A0%E5%9D%A1%20V2Ray%20VLESS%20Settings%20%E6%89%81%E5%B9%B3" {
		t.Fatalf("unexpected V2Ray direct VLESS URI: %q", lines[1])
	}
	if lines[2] != "trojan://trojan-placeholder@trojan-direct.v2ray.example.test:443#%E4%B8%9C%E4%BA%AC%20V2Ray%20Trojan%20%E6%89%81%E5%B9%B3" {
		t.Fatalf("unexpected V2Ray direct Trojan URI: %q", lines[2])
	}
	if lines[3] != "ss://aes-256-gcm:ss-placeholder@ss-direct.v2ray.example.test:8388#%E9%A6%96%E5%B0%94%20V2Ray%20SS%20Settings%20%E6%89%81%E5%B9%B3" {
		t.Fatalf("unexpected V2Ray direct Shadowsocks URI: %q", lines[3])
	}
}

func TestNormalizeContentV2RayJSONTCPHTTPHeader(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray TCP HTTP",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "tcp-http.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000090"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "security": "tls",
        "tlsSettings": {
          "serverName": "tcp-http.v2ray.example.test"
        },
        "tcpSettings": {
          "header": {
            "type": "http",
            "request": {
              "method": "GET",
              "path": ["/front", "/backup"],
              "headers": {
                "Host": ["front.v2ray.example.test", "front-backup.v2ray.example.test"]
              }
            }
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000090@tcp-http.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=tcp-http.v2ray.example.test",
		"type=http",
		"host=front.v2ray.example.test%2Cfront-backup.v2ray.example.test",
		"path=%2Ffront%2C%2Fbackup",
		"method=GET",
		"#%E9%A6%99%E6%B8%AF%20V2Ray%20TCP%20HTTP",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected v2ray TCP HTTP URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONHTTPUpgradeHosts(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray HTTPUpgrade Host 数组",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "upgrade-array.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000091"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "httpupgrade",
        "security": "tls",
        "httpupgradeSettings": {
          "host": ["upgrade.v2ray.example.test", "upgrade-backup.v2ray.example.test"],
          "path": "/upgrade-array"
        }
      }
    },
    {
      "tag": "东京 V2Ray HTTPUpgrade Header",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "upgrade-header.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "network": "httpupgrade",
        "security": "tls",
        "httpUpgradeSettings": {
          "path": "/upgrade-header",
          "headers": {
            "Host": ["header.v2ray.example.test", "header-backup.v2ray.example.test"]
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000091@upgrade-array.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"type=httpupgrade",
		"host=upgrade.v2ray.example.test%2Cupgrade-backup.v2ray.example.test",
		"path=%2Fupgrade-array",
	} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("expected v2ray HTTPUpgrade URI to contain %q: %q", want, lines[0])
		}
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@upgrade-header.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"type=httpupgrade",
		"host=header.v2ray.example.test%2Cheader-backup.v2ray.example.test",
		"path=%2Fupgrade-header",
	} {
		if !strings.Contains(lines[1], want) {
			t.Fatalf("expected v2ray HTTPUpgrade header URI to contain %q: %q", want, lines[1])
		}
	}
}

func TestNormalizeContentJSONWrappedV2RayJSON(t *testing.T) {
	raw := `{
  "data": {
    "raw_content": "{\"outbounds\":[{\"tag\":\"包装直连\",\"protocol\":\"freedom\"}]}"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "direct://default#%E5%8C%85%E8%A3%85%E7%9B%B4%E8%BF%9E" {
		t.Fatalf("unexpected wrapped v2ray JSON URI: %q", got)
	}
}

func TestNormalizeContentV2RayJSONInternalOutbounds(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray 直连",
      "protocol": "freedom"
    },
    {
      "tag": "V2Ray 拦截",
      "protocol": "blackhole"
    },
    {
      "tag": "V2Ray DNS",
      "protocol": "dns"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	want := strings.Join([]string{
		"direct://default#V2Ray%20%E7%9B%B4%E8%BF%9E",
		"block://default#V2Ray%20%E6%8B%A6%E6%88%AA",
		"dns://default#V2Ray%20DNS",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected v2ray internal outbounds: %q", got)
	}
}

func TestNormalizeContentV2RayJSONLegacyOutboundFields(t *testing.T) {
	raw := `{
  "Outbound": {
    "tag": "香港 V2Ray Legacy VMess",
    "protocol": "vmess",
    "settings": {
      "vnext": [
        {
          "address": "legacy-vmess.v2ray.example.test",
          "port": 443,
          "users": [
            {
              "id": "00000000-0000-0000-0000-000000000099",
              "security": "auto"
            }
          ]
        }
      ]
    }
  },
  "OutboundDetour": [
    {
      "tag": "东京 V2Ray Legacy Trojan",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "legacy-trojan.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 legacy V2Ray outbound URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	for _, want := range []string{
		`"ps":"香港 V2Ray Legacy VMess"`,
		`"add":"legacy-vmess.v2ray.example.test"`,
		`"id":"00000000-0000-0000-0000-000000000099"`,
	} {
		if !strings.Contains(decodedText, want) {
			t.Fatalf("expected legacy v2ray VMess document to contain %q: %q", want, decodedText)
		}
	}
	if lines[1] != "trojan://trojan-placeholder@legacy-trojan.v2ray.example.test:443#%E4%B8%9C%E4%BA%AC%20V2Ray%20Legacy%20Trojan" {
		t.Fatalf("unexpected legacy v2ray Trojan URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONHTTPAndSOCKS(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "东京 V2Ray HTTP",
      "protocol": "http",
      "settings": {
        "servers": [
          {
            "address": "http.v2ray.example.test",
            "port": 8080,
            "users": [
              {
                "user": "qa-user",
                "pass": "http-placeholder"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "security": "tls",
        "tlsSettings": {
          "serverName": "http.v2ray.example.test",
          "allowInsecure": true
        }
      }
    },
    {
      "tag": "首尔 V2Ray SOCKS",
      "protocol": "socks",
      "settings": {
        "servers": [
          {
            "address": "socks.v2ray.example.test",
            "port": 1080,
            "users": [
              {
                "user": "qa-user",
                "pass": "socks-placeholder"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "https://qa-user:http-placeholder@http.v2ray.example.test:8080?")
	if !strings.Contains(lines[0], "sni=http.v2ray.example.test") ||
		!strings.Contains(lines[0], "insecure=1") ||
		!strings.HasSuffix(lines[0], "#%E4%B8%9C%E4%BA%AC%20V2Ray%20HTTP") {
		t.Fatalf("unexpected v2ray http URI: %q", lines[0])
	}
	if lines[1] != "socks5://qa-user:socks-placeholder@socks.v2ray.example.test:1080#%E9%A6%96%E5%B0%94%20V2Ray%20SOCKS" {
		t.Fatalf("unexpected v2ray socks URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONProtocolAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray Trojan-Go Alias",
      "protocol": "trojan-go",
      "settings": {
        "servers": [
          {
            "address": "trojan-go.v2ray-alias.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "security": "tls",
        "tlsSettings": {
          "serverName": "trojan-go.v2ray-alias.example.test"
        }
      }
    },
    {
      "tag": "V2Ray VMess AEAD Alias",
      "protocol": "vmess-aead",
      "settings": {
        "vnext": [
          {
            "address": "vmess-aead.v2ray-alias.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000094",
                "security": "auto"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "security": "tls",
        "tlsSettings": {
          "serverName": "vmess-aead.v2ray-alias.example.test"
        }
      }
    },
    {
      "tag": "V2Ray HTTP Plus TLS Alias",
      "protocol": "http+tls",
      "settings": {
        "servers": [
          {
            "address": "http-plus-tls.v2ray-alias.example.test",
            "port": 443,
            "users": [
              {
                "user": "qa-user",
                "pass": "http-placeholder"
              }
            ]
          }
        ]
      }
    },
    {
      "tag": "V2Ray HTTP Dash TLS Alias",
      "protocol": "http-tls",
      "settings": {
        "servers": [
          {
            "address": "http-dash-tls.v2ray-alias.example.test",
            "port": 443,
            "username": "qa-user",
            "password": "http-dash-placeholder"
          }
        ]
      }
    },
    {
      "tag": "V2Ray SOCKS5H Alias",
      "protocol": "socks5h",
      "settings": {
        "servers": [
          {
            "address": "socks5h.v2ray-alias.example.test",
            "port": 1080,
            "users": [
              {
                "user": "qa-user",
                "pass": "socks-placeholder"
              }
            ]
          }
        ]
      }
    },
    {
      "tag": "V2Ray SOCKS4A Alias",
      "protocol": "socks4a",
      "settings": {
        "servers": [
          {
            "address": "socks4a.v2ray-alias.example.test",
            "port": 1081,
            "username": "qa-user",
            "password": "socks4a-placeholder"
          }
        ]
      }
    },
    {
      "tag": "V2Ray Reject Alias",
      "protocol": "reject-drop"
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 7 {
		t.Fatalf("expected 7 v2ray alias URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "trojan://trojan-placeholder@trojan-go.v2ray-alias.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=trojan-go.v2ray-alias.example.test") ||
		!strings.HasSuffix(lines[0], "#V2Ray%20Trojan-Go%20Alias") {
		t.Fatalf("unexpected v2ray Trojan-Go alias URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "vmess://")
	decodedVMessText := decodeVMessURIForTest(t, lines[1])
	for _, want := range []string{
		`"ps":"V2Ray VMess AEAD Alias"`,
		`"add":"vmess-aead.v2ray-alias.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000094"`,
		`"tls":"tls"`,
		`"sni":"vmess-aead.v2ray-alias.example.test"`,
	} {
		if !strings.Contains(decodedVMessText, want) {
			t.Fatalf("expected v2ray VMess AEAD alias document to contain %q: %q", want, decodedVMessText)
		}
	}
	assertHasPrefix(t, lines[2], "https://qa-user:http-placeholder@http-plus-tls.v2ray-alias.example.test:443#V2Ray%20HTTP%20Plus%20TLS%20Alias")
	assertHasPrefix(t, lines[3], "https://qa-user:http-dash-placeholder@http-dash-tls.v2ray-alias.example.test:443#V2Ray%20HTTP%20Dash%20TLS%20Alias")
	if lines[4] != "socks5://qa-user:socks-placeholder@socks5h.v2ray-alias.example.test:1080#V2Ray%20SOCKS5H%20Alias" {
		t.Fatalf("unexpected v2ray SOCKS5H alias URI: %q", lines[4])
	}
	if lines[5] != "socks4a://qa-user:socks4a-placeholder@socks4a.v2ray-alias.example.test:1081#V2Ray%20SOCKS4A%20Alias" {
		t.Fatalf("unexpected v2ray SOCKS4A alias URI: %q", lines[5])
	}
	if lines[6] != "block://default#V2Ray%20Reject%20Alias" {
		t.Fatalf("unexpected v2ray reject alias URI: %q", lines[6])
	}
}

func TestNormalizeContentV2RayJSONProxyAccountAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "东京 V2Ray HTTP Accounts",
      "protocol": "http",
      "settings": {
        "servers": [
          {
            "address": "http-accounts.v2ray.example.test",
            "port": 8080,
            "accounts": {
              "qa-http": {
                "password": "http-placeholder",
                "name": "qa-http"
              }
            }
          }
        ]
      }
    },
    {
      "tag": "首尔 V2Ray SOCKS Direct Account",
      "protocol": "socks",
      "settings": {
        "servers": [
          {
            "address": "socks-direct.v2ray.example.test",
            "port": 1080,
            "username": "qa-socks",
            "password": "socks-placeholder",
            "udpEnabled": true
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "http://qa-http:http-placeholder@http-accounts.v2ray.example.test:8080#qa-http" {
		t.Fatalf("unexpected v2ray http accounts URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "socks5://qa-socks:socks-placeholder@socks-direct.v2ray.example.test:1080?")
	if !strings.Contains(lines[1], "udp=1") ||
		!strings.HasSuffix(lines[1], "#%E9%A6%96%E5%B0%94%20V2Ray%20SOCKS%20Direct%20Account") {
		t.Fatalf("unexpected v2ray socks direct account URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONProxyEndpointAccountMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray HTTP Endpoint Account Map",
      "protocol": "http",
      "settings": {
        "servers": {
          "http-account-map.v2ray.example.test:8080": {
            "qa-http": "http-map-placeholder",
            "qa-http-object": {
              "pass": "http-map-object-placeholder",
              "name": "qa-http-object"
            }
          }
        }
      }
    },
    {
      "tag": "V2Ray SOCKS Endpoint Account Map",
      "protocol": "socks",
      "settings": {
        "servers": {
          "socks-account-map.v2ray.example.test:1080": {
            "qa-socks": {
              "password": "socks-map-placeholder",
              "name": "qa-socks"
            }
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "http://qa-http:http-map-placeholder@http-account-map.v2ray.example.test:8080#qa-http" {
		t.Fatalf("unexpected v2ray http scalar account map URI: %q", lines[0])
	}
	if lines[1] != "http://qa-http-object:http-map-object-placeholder@http-account-map.v2ray.example.test:8080#qa-http-object" {
		t.Fatalf("unexpected v2ray http object account map URI: %q", lines[1])
	}
	if lines[2] != "socks5://qa-socks:socks-map-placeholder@socks-account-map.v2ray.example.test:1080#qa-socks" {
		t.Fatalf("unexpected v2ray socks account map URI: %q", lines[2])
	}
}

func TestNormalizeContentV2RayJSONProxyEndpointScalarAccounts(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray HTTP Endpoint Scalar Account",
      "protocol": "http",
      "settings": {
        "servers": {
          "http-scalar-account.v2ray.example.test:8080": "qa-http:http-scalar-placeholder"
        }
      }
    },
    {
      "tag": "V2Ray SOCKS Endpoint Scalar Accounts",
      "protocol": "socks",
      "settings": {
        "servers": {
          "socks-scalar-account.v2ray.example.test:1080": [
            "qa-socks-a:socks-scalar-a-placeholder",
            "qa-socks-b:socks-scalar-b-placeholder"
          ]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "http://qa-http:http-scalar-placeholder@http-scalar-account.v2ray.example.test:8080#qa-http" {
		t.Fatalf("unexpected v2ray http scalar account URI: %q", lines[0])
	}
	if lines[1] != "socks5://qa-socks-a:socks-scalar-a-placeholder@socks-scalar-account.v2ray.example.test:1080#qa-socks-a" {
		t.Fatalf("unexpected v2ray socks scalar account URI: %q", lines[1])
	}
	if lines[2] != "socks5://qa-socks-b:socks-scalar-b-placeholder@socks-scalar-account.v2ray.example.test:1080#qa-socks-b" {
		t.Fatalf("unexpected v2ray socks scalar account URI: %q", lines[2])
	}
}

func TestNormalizeContentV2RayJSONVNextEndpointAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray VMess ServerHost Alias",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "serverHost": "vmess-alias.v2ray.example.test",
            "portNumber": 443,
            "users": [
              {
                "user_id": "00000000-0000-0000-0000-000000000087",
                "security": "auto"
              }
            ]
          }
        ]
      }
    },
    {
      "tag": "新加坡 V2Ray VLESS NodeHost Alias",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "nodeHost": "vless-alias.v2ray.example.test",
            "nodePort": 8443,
            "users": [
              {
                "user-id": "00000000-0000-0000-0000-000000000088",
                "encryption": "none",
                "flowName": "xtls-rprx-vision"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"add":"vmess-alias.v2ray.example.test"`) ||
		!strings.Contains(decodedText, `"port":"443"`) {
		t.Fatalf("unexpected v2ray vmess alias document: %q", decodedText)
	}
	if lines[1] != "vless://00000000-0000-0000-0000-000000000088@vless-alias.v2ray.example.test:8443?flow=xtls-rprx-vision#%E6%96%B0%E5%8A%A0%E5%9D%A1%20V2Ray%20VLESS%20NodeHost%20Alias" {
		t.Fatalf("unexpected v2ray vless alias URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONVNextUserMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray VMess User Map",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "vmess-user-map.v2ray.example.test",
            "port": 443,
            "users": {
              "00000000-0000-0000-0000-000000000096": {
                "security": "auto",
                "email": "香港 V2Ray VMess User Map"
              }
            }
          }
        ]
      }
    },
    {
      "tag": "V2Ray VLESS User Map",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "vless-user-map.v2ray.example.test",
            "port": 443,
            "users": {
              "00000000-0000-0000-0000-000000000097": {
                "encryption": "none",
                "flow": "xtls-rprx-vision",
                "name": "东京 V2Ray VLESS User Map"
              }
            }
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 V2Ray user map URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	for _, want := range []string{
		`"ps":"香港 V2Ray VMess User Map"`,
		`"add":"vmess-user-map.v2ray.example.test"`,
		`"id":"00000000-0000-0000-0000-000000000096"`,
	} {
		if !strings.Contains(decodedText, want) {
			t.Fatalf("expected v2ray VMess user map document to contain %q: %q", want, decodedText)
		}
	}
	if lines[1] != "vless://00000000-0000-0000-0000-000000000097@vless-user-map.v2ray.example.test:443?flow=xtls-rprx-vision#%E4%B8%9C%E4%BA%AC%20V2Ray%20VLESS%20User%20Map" {
		t.Fatalf("unexpected v2ray VLESS user map URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONEndpointObjectMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray VMess Endpoint Map",
      "protocol": "vmess",
      "settings": {
        "vnext": {
          "vmess-endpoint-map.v2ray.example.test:443": {
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000098",
                "security": "auto",
                "email": "香港 V2Ray VMess Endpoint Map"
              }
            ]
          }
        }
      }
    },
    {
      "tag": "V2Ray Trojan Endpoint Map",
      "protocol": "trojan",
      "settings": {
        "servers": {
          "trojan-endpoint-map.v2ray.example.test:443": {
            "password": "trojan-placeholder",
            "name": "东京 V2Ray Trojan Endpoint Map"
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 V2Ray endpoint map URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	for _, want := range []string{
		`"ps":"香港 V2Ray VMess Endpoint Map"`,
		`"add":"vmess-endpoint-map.v2ray.example.test"`,
		`"id":"00000000-0000-0000-0000-000000000098"`,
	} {
		if !strings.Contains(decodedText, want) {
			t.Fatalf("expected v2ray VMess endpoint map document to contain %q: %q", want, decodedText)
		}
	}
	if lines[1] != "trojan://trojan-placeholder@trojan-endpoint-map.v2ray.example.test:443#%E4%B8%9C%E4%BA%AC%20V2Ray%20Trojan%20Endpoint%20Map" {
		t.Fatalf("unexpected v2ray Trojan endpoint map URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONEndpointUserArrayMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray VMess Endpoint User Array",
      "protocol": "vmess",
      "settings": {
        "vnext": {
          "vmess-user-array.v2ray.example.test:443": [
            {
              "id": "00000000-0000-0000-0000-000000000099",
              "security": "auto",
              "email": "香港 V2Ray VMess User Array"
            }
          ]
        }
      }
    },
    {
      "tag": "V2Ray VLESS Endpoint User Array",
      "protocol": "vless",
      "settings": {
        "vnext": {
          "vless-user-array.v2ray.example.test:8443": [
            {
              "uuid": "00000000-0000-0000-0000-000000000100",
              "flow": "xtls-rprx-vision",
              "name": "东京 V2Ray VLESS User Array"
            }
          ]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 V2Ray endpoint user array URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	for _, want := range []string{
		`"ps":"香港 V2Ray VMess User Array"`,
		`"add":"vmess-user-array.v2ray.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000099"`,
	} {
		if !strings.Contains(decodedText, want) {
			t.Fatalf("expected v2ray VMess endpoint user array document to contain %q: %q", want, decodedText)
		}
	}
	if lines[1] != "vless://00000000-0000-0000-0000-000000000100@vless-user-array.v2ray.example.test:8443?flow=xtls-rprx-vision#%E4%B8%9C%E4%BA%AC%20V2Ray%20VLESS%20User%20Array" {
		t.Fatalf("unexpected v2ray VLESS endpoint user array URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONEndpointScalarUserMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray VMess Endpoint Scalar User",
      "protocol": "vmess",
      "settings": {
        "vnext": {
          "vmess-scalar-user.v2ray.example.test:443": "00000000-0000-0000-0000-000000000101"
        }
      }
    },
    {
      "tag": "V2Ray VLESS Endpoint Scalar Users",
      "protocol": "vless",
      "settings": {
        "vnext": {
          "vless-scalar-users.v2ray.example.test:8443": [
            "00000000-0000-0000-0000-000000000102",
            "00000000-0000-0000-0000-000000000103"
          ]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 V2Ray endpoint scalar user URIs, got %d: %q", len(lines), got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(lines[0], "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	for _, want := range []string{
		`"ps":"vmess-scalar-user.v2ray.example.test:443"`,
		`"add":"vmess-scalar-user.v2ray.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000101"`,
	} {
		if !strings.Contains(decodedText, want) {
			t.Fatalf("expected v2ray VMess endpoint scalar user document to contain %q: %q", want, decodedText)
		}
	}
	expected := []string{
		"vless://00000000-0000-0000-0000-000000000102@vless-scalar-users.v2ray.example.test:8443#vless-scalar-users.v2ray.example.test:8443-1",
		"vless://00000000-0000-0000-0000-000000000103@vless-scalar-users.v2ray.example.test:8443#vless-scalar-users.v2ray.example.test:8443-2",
	}
	for index, want := range expected {
		if lines[index+1] != want {
			t.Fatalf("unexpected v2ray VLESS endpoint scalar user URI at %d: %q", index+1, lines[index+1])
		}
	}
}

func TestNormalizeContentV2RayJSONEndpointMappedUserMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray VMess Endpoint Mapped Users",
      "protocol": "vmess",
      "settings": {
        "vnext": {
          "vmess-mapped-users.v2ray.example.test:443": {
            "qa-vmess": "00000000-0000-0000-0000-000000000104",
            "qa-vmess-object": {
              "uuid": "00000000-0000-0000-0000-000000000105",
              "security": "auto"
            }
          }
        }
      }
    },
    {
      "tag": "V2Ray VLESS Endpoint Mapped Users",
      "protocol": "vless",
      "settings": {
        "vnext": {
          "vless-mapped-users.v2ray.example.test:8443": {
            "qa-vless": {
              "id": "00000000-0000-0000-0000-000000000106",
              "flow": "xtls-rprx-vision"
            }
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 V2Ray endpoint mapped user URIs, got %d: %q", len(lines), got)
	}
	for index, want := range []string{
		`"ps":"qa-vmess"`,
		`"id":"00000000-0000-0000-0000-000000000104"`,
		`"add":"vmess-mapped-users.v2ray.example.test"`,
	} {
		decoded := decodeVMessURIForTest(t, lines[0])
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected v2ray VMess mapped scalar user %d to contain %q: %q", index, want, decoded)
		}
	}
	for index, want := range []string{
		`"ps":"qa-vmess-object"`,
		`"id":"00000000-0000-0000-0000-000000000105"`,
		`"scy":"auto"`,
	} {
		decoded := decodeVMessURIForTest(t, lines[1])
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected v2ray VMess mapped object user %d to contain %q: %q", index, want, decoded)
		}
	}
	if lines[2] != "vless://00000000-0000-0000-0000-000000000106@vless-mapped-users.v2ray.example.test:8443?flow=xtls-rprx-vision#qa-vless" {
		t.Fatalf("unexpected v2ray VLESS endpoint mapped user URI: %q", lines[2])
	}
}

func TestNormalizeContentV2RayJSONServerScalarPasswordMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray Trojan Scalar Server",
      "protocol": "trojan",
      "settings": {
        "servers": {
          "trojan-scalar-server.v2ray.example.test:443": "trojan-scalar-placeholder"
        }
      }
    },
    {
      "tag": "V2Ray SS Scalar Server",
      "protocol": "shadowsocks",
      "settings": {
        "method": "aes-128-gcm",
        "servers": {
          "ss-scalar-server.v2ray.example.test:8388": "ss-scalar-placeholder"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 V2Ray scalar server map URIs, got %d: %q", len(lines), got)
	}
	if lines[0] != "trojan://trojan-scalar-placeholder@trojan-scalar-server.v2ray.example.test:443#trojan-scalar-server.v2ray.example.test:443" {
		t.Fatalf("unexpected v2ray Trojan scalar server URI: %q", lines[0])
	}
	if lines[1] != "ss://aes-128-gcm:ss-scalar-placeholder@ss-scalar-server.v2ray.example.test:8388#ss-scalar-server.v2ray.example.test:8388" {
		t.Fatalf("unexpected v2ray Shadowsocks scalar server URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONServerScalarPasswordArrayMaps(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "V2Ray Trojan Scalar Server Array",
      "protocol": "trojan",
      "settings": {
        "servers": {
          "trojan-scalar-array.v2ray.example.test:443": [
            "trojan-array-one",
            "trojan-array-two"
          ]
        }
      }
    },
    {
      "tag": "V2Ray SS Scalar Server Array",
      "protocol": "shadowsocks",
      "settings": {
        "method": "aes-128-gcm",
        "servers": {
          "ss-scalar-array.v2ray.example.test:8388": [
            "ss-array-one",
            "ss-array-two"
          ]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	expected := []string{
		"trojan://trojan-array-one@trojan-scalar-array.v2ray.example.test:443#trojan-scalar-array.v2ray.example.test:443-1",
		"trojan://trojan-array-two@trojan-scalar-array.v2ray.example.test:443#trojan-scalar-array.v2ray.example.test:443-2",
		"ss://aes-128-gcm:ss-array-one@ss-scalar-array.v2ray.example.test:8388#ss-scalar-array.v2ray.example.test:8388-1",
		"ss://aes-128-gcm:ss-array-two@ss-scalar-array.v2ray.example.test:8388#ss-scalar-array.v2ray.example.test:8388-2",
	}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d V2Ray scalar server array URIs, got %d: %q", len(expected), len(lines), got)
	}
	for index, want := range expected {
		if lines[index] != want {
			t.Fatalf("unexpected v2ray scalar server array URI at %d: %q", index, lines[index])
		}
	}
}

func TestNormalizeContentV2RayJSONVLESSPacketEncoding(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray VLESS Packet Encoding",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "packet-encoding.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000094",
                "encryption": "none",
                "packetEncoding": "xudp"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000094@packet-encoding.v2ray.example.test:443?")
	if !strings.Contains(got, "packet_encoding=xudp") ||
		!strings.HasSuffix(got, "#%E9%A6%99%E6%B8%AF%20V2Ray%20VLESS%20Packet%20Encoding") {
		t.Fatalf("unexpected v2ray vless packet encoding URI: %q", got)
	}
}

func TestNormalizeContentV2RayJSONVMessPacketEncoding(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray VMess Packet Encoding",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "vmess-packet-encoding.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000095",
                "security": "auto",
                "packetEncoding": "packetaddr"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if !strings.HasPrefix(got, "vmess://") {
		t.Fatalf("expected vmess URI, got %q", got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode v2ray vmess URI: %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, `"packet_encoding":"packetaddr"`) ||
		!strings.Contains(decodedText, `"add":"vmess-packet-encoding.v2ray.example.test"`) {
		t.Fatalf("unexpected v2ray vmess packet encoding document: %q", decodedText)
	}
}

func TestNormalizeContentV2RayJSONTransportHostAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray WS Host Alias",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "ws-host-alias.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000089",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "wsSettings": {
          "path": "/edge",
          "host": "ws-direct.v2ray.example.test"
        }
      }
    },
    {
      "tag": "东京 V2Ray HTTP Header Host",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "h2-header.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "network": "http",
        "security": "tls",
        "httpSettings": {
          "path": "/h2",
          "headers": {
            "Host": ["h2-a.v2ray.example.test", "h2-b.v2ray.example.test"]
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000089@ws-host-alias.v2ray.example.test:443?")
	if !strings.Contains(lines[0], "type=ws") ||
		!strings.Contains(lines[0], "path=%2Fedge") ||
		!strings.Contains(lines[0], "host=ws-direct.v2ray.example.test") {
		t.Fatalf("unexpected v2ray websocket host alias URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@h2-header.v2ray.example.test:443?")
	if !strings.Contains(lines[1], "type=http") ||
		!strings.Contains(lines[1], "path=%2Fh2") ||
		!strings.Contains(lines[1], "host=h2-a.v2ray.example.test%2Ch2-b.v2ray.example.test") {
		t.Fatalf("unexpected v2ray http host alias URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONStreamSettingAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray Stream Hyphen",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "stream-hyphen.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000094",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "stream-settings": {
        "network": "ws",
        "security": "tls",
        "tls-settings": {
          "server-name": "stream-hyphen-sni.v2ray.example.test"
        },
        "ws-settings": {
          "path": "/stream",
          "headers": {
            "Host": "stream-host.v2ray.example.test"
          }
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000094@stream-hyphen.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=stream-hyphen-sni.v2ray.example.test",
		"type=ws",
		"path=%2Fstream",
		"host=stream-host.v2ray.example.test",
		"#%E9%A6%99%E6%B8%AF%20V2Ray%20Stream%20Hyphen",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected v2ray stream setting alias URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONWebSocketEarlyDataAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray WS Early",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "ws-early.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000093",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "wsSettings": {
          "path": "/early",
          "max-early-data": 2048,
          "early_data_header_name": "Sec-WebSocket-Protocol"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000093@ws-early.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"type=ws",
		"path=%2Fearly",
		"max_early_data=2048",
		"early_data_header_name=Sec-WebSocket-Protocol",
		"#%E9%A6%99%E6%B8%AF%20V2Ray%20WS%20Early",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected v2ray websocket early data alias URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONHTTPPathArray(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "东京 V2Ray HTTP Path Array",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "h2-path-array.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "network": "http",
        "security": "tls",
        "httpSettings": {
          "path": ["/front", "/backup"],
          "host": ["h2-path-a.v2ray.example.test", "h2-path-b.v2ray.example.test"]
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "trojan://trojan-placeholder@h2-path-array.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"type=http",
		"path=%2Ffront%2C%2Fbackup",
		"host=h2-path-a.v2ray.example.test%2Ch2-path-b.v2ray.example.test",
		"#%E4%B8%9C%E4%BA%AC%20V2Ray%20HTTP%20Path%20Array",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected v2ray HTTP path array URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONQUICSettings(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray QUIC Settings",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "quic-settings.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000093"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "security": "tls",
        "tlsSettings": {
          "serverName": "quic-settings.v2ray.example.test"
        },
        "quicSettings": {
          "security": "none"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000093@quic-settings.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"sni=quic-settings.v2ray.example.test",
		"type=quic",
		"#%E9%A6%99%E6%B8%AF%20V2Ray%20QUIC%20Settings",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected V2Ray quicSettings URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONGRPCHyphenAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray gRPC Hyphen",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "grpc-hyphen.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000092",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "grpc",
        "security": "tls",
        "grpcSettings": {
          "service-name": "fluxgate-hyphen",
          "idle-timeout": "30s",
          "health-check-timeout": "10s",
          "permit-without-stream": true,
          "multi-mode": true
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000092@grpc-hyphen.v2ray.example.test:443?")
	for _, want := range []string{
		"security=tls",
		"type=grpc",
		"service_name=fluxgate-hyphen",
		"idle_timeout=30s",
		"ping_timeout=10s",
		"permit_without_stream=1",
		"multi_mode=1",
		"#%E9%A6%99%E6%B8%AF%20V2Ray%20gRPC%20Hyphen",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected v2ray gRPC hyphen alias URI to contain %q: %q", want, got)
		}
	}
}

func TestNormalizeContentV2RayJSONRealityArrayAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "台湾 V2Ray Reality Alias",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "reality-alias.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000086",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "realitySettings": {
          "serverNames": ["www.alias.example.test", "backup.alias.example.test"],
          "public-key": "alias-public-key",
          "shortIds": ["abcd1234", "ef56"],
          "fingerprint": "chrome",
          "spider-x": "cdn"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vless://00000000-0000-0000-0000-000000000086@reality-alias.v2ray.example.test:443?")
	if !strings.Contains(got, "security=reality") ||
		!strings.Contains(got, "sni=www.alias.example.test") ||
		!strings.Contains(got, "pbk=alias-public-key") ||
		!strings.Contains(got, "sid=abcd1234") ||
		!strings.Contains(got, "fp=chrome") ||
		!strings.Contains(got, "spx=cdn") ||
		!strings.HasSuffix(got, "#%E5%8F%B0%E6%B9%BE%20V2Ray%20Reality%20Alias") {
		t.Fatalf("unexpected v2ray reality aliases URI: %q", got)
	}
}

func TestNormalizeContentV2RayJSONTLSNameAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray TLS Names",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "tls-names.v2ray.example.test",
            "port": 443,
            "users": [
              {
                "id": "00000000-0000-0000-0000-000000000090",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "security": "tls",
        "tlsSettings": {
          "serverNames": ["tls-a.v2ray.example.test", "tls-b.v2ray.example.test"]
        }
      }
    },
    {
      "tag": "东京 V2Ray TLS Hyphen",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "address": "tls-hyphen.v2ray.example.test",
            "port": 443,
            "password": "trojan-placeholder"
          }
        ]
      },
      "streamSettings": {
        "security": "tls",
        "tlsSettings": {
          "server-name": "tls-hyphen-sni.v2ray.example.test"
        }
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	assertHasPrefix(t, lines[0], "vless://00000000-0000-0000-0000-000000000090@tls-names.v2ray.example.test:443?")
	if !strings.Contains(lines[0], "security=tls") ||
		!strings.Contains(lines[0], "sni=tls-a.v2ray.example.test") ||
		!strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%20V2Ray%20TLS%20Names") {
		t.Fatalf("unexpected v2ray tls serverNames URI: %q", lines[0])
	}
	assertHasPrefix(t, lines[1], "trojan://trojan-placeholder@tls-hyphen.v2ray.example.test:443?")
	if !strings.Contains(lines[1], "security=tls") ||
		!strings.Contains(lines[1], "sni=tls-hyphen-sni.v2ray.example.test") ||
		!strings.HasSuffix(lines[1], "#%E4%B8%9C%E4%BA%AC%20V2Ray%20TLS%20Hyphen") {
		t.Fatalf("unexpected v2ray tls server-name URI: %q", lines[1])
	}
}

func TestNormalizeContentV2RayJSONShadowsocksNetwork(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "首尔 V2Ray SS UDP",
      "protocol": "shadowsocks",
      "settings": {
        "servers": [
          {
            "address": "ss-udp.v2ray.example.test",
            "port": 8388,
            "method": "aes-128-gcm",
            "password": "qa-placeholder",
            "network": "udp"
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "ss://aes-128-gcm:qa-placeholder@ss-udp.v2ray.example.test:8388?")
	if !strings.Contains(got, "network=udp") {
		t.Fatalf("expected v2ray shadowsocks network to be preserved: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20V2Ray%20SS%20UDP") {
		t.Fatalf("unexpected v2ray shadowsocks fragment: %q", got)
	}
}

func TestNormalizeContentV2RayJSONShadowsocksPluginOptions(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray SS Plugin",
      "protocol": "shadowsocks",
      "settings": {
        "plugin": "v2ray-plugin",
        "plugin_options": "mode=websocket;host=ss-plugin.v2ray.example.test",
        "servers": [
          {
            "address": "ss-plugin.v2ray.example.test",
            "port": 8388,
            "method": "aes-128-gcm",
            "password": "qa-placeholder"
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "ss://aes-128-gcm:qa-placeholder@ss-plugin.v2ray.example.test:8388?")
	if !strings.Contains(got, "plugin=v2ray-plugin") ||
		!strings.Contains(got, "plugin_opts=mode%3Dwebsocket%3Bhost%3Dss-plugin.v2ray.example.test") {
		t.Fatalf("expected v2ray shadowsocks plugin options to be preserved: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%99%E6%B8%AF%20V2Ray%20SS%20Plugin") {
		t.Fatalf("unexpected v2ray shadowsocks plugin fragment: %q", got)
	}
}

func TestNormalizeContentV2RayJSONSOCKSUDPFlag(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "首尔 V2Ray SOCKS UDP",
      "protocol": "socks",
      "settings": {
        "udp": true,
        "servers": [
          {
            "address": "socks-udp.v2ray.example.test",
            "port": 1080,
            "users": [
              {
                "user": "qa-user",
                "pass": "socks-placeholder"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@socks-udp.v2ray.example.test:1080?")
	if !strings.Contains(got, "udp=1") || strings.Contains(got, "network=udp") {
		t.Fatalf("unexpected v2ray socks udp URI: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20V2Ray%20SOCKS%20UDP") {
		t.Fatalf("unexpected v2ray socks udp fragment: %q", got)
	}
}

func TestNormalizeContentV2RayJSONSOCKSUDPOverTCPFlag(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "首尔 V2Ray SOCKS UDP over TCP",
      "protocol": "socks",
      "settings": {
        "servers": [
          {
            "address": "socks-uot.v2ray.example.test",
            "port": 1080,
            "udpOverTcp": true,
            "users": [
              {
                "user": "qa-user",
                "pass": "socks-placeholder"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "socks5://qa-user:socks-placeholder@socks-uot.v2ray.example.test:1080?")
	if !strings.Contains(got, "udp_over_tcp=1") {
		t.Fatalf("expected v2ray socks udp over tcp to be preserved: %q", got)
	}
	if !strings.HasSuffix(got, "#%E9%A6%96%E5%B0%94%20V2Ray%20SOCKS%20UDP%20over%20TCP") {
		t.Fatalf("unexpected v2ray socks udp over tcp fragment: %q", got)
	}
}

func TestNormalizeContentV2RayJSONServerFieldAliases(t *testing.T) {
	raw := `{
  "outbounds": [
    {
      "tag": "香港 V2Ray Trojan Alias",
      "protocol": "trojan",
      "settings": {
        "servers": [
          {
            "host": "trojan-alias.v2ray.example.test",
            "server-port": 443,
            "token": "trojan-placeholder"
          }
        ]
      }
    },
    {
      "tag": "首尔 V2Ray SS Alias",
      "protocol": "shadowsocks",
      "settings": {
        "servers": [
          {
            "add": "ss-alias.v2ray.example.test",
            "serverPort": 8388,
            "encryptMethod": "aes-256-gcm",
            "passwd": "ss-placeholder"
          }
        ]
      }
    },
    {
      "tag": "东京 V2Ray HTTP Alias",
      "protocol": "http",
      "settings": {
        "servers": [
          {
            "host": "http-alias.v2ray.example.test",
            "server-port": 8080,
            "userName": "qa-user",
            "pwd": "http-placeholder"
          }
        ]
      }
    },
    {
      "tag": "大阪 V2Ray SOCKS Alias",
      "protocol": "socks",
      "settings": {
        "servers": [
          {
            "add": "socks-alias.v2ray.example.test",
            "serverPort": 1080,
            "users": [
              {
                "accountName": "qa-user",
                "psk": "socks-placeholder"
              }
            ]
          }
        ]
      }
    }
  ]
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 URIs, got %d: %q", len(lines), got)
	}
	expectations := []string{
		"trojan://trojan-placeholder@trojan-alias.v2ray.example.test:443#%E9%A6%99%E6%B8%AF%20V2Ray%20Trojan%20Alias",
		"ss://aes-256-gcm:ss-placeholder@ss-alias.v2ray.example.test:8388#%E9%A6%96%E5%B0%94%20V2Ray%20SS%20Alias",
		"http://qa-user:http-placeholder@http-alias.v2ray.example.test:8080#%E4%B8%9C%E4%BA%AC%20V2Ray%20HTTP%20Alias",
		"socks5://qa-user:socks-placeholder@socks-alias.v2ray.example.test:1080#%E5%A4%A7%E9%98%AA%20V2Ray%20SOCKS%20Alias",
	}
	for index, want := range expectations {
		if lines[index] != want {
			t.Fatalf("line %d mismatch\nwant: %q\n got: %q", index, want, lines[index])
		}
	}
}

func TestNormalizeContentVMessJSON(t *testing.T) {
	raw := `{
  "v": "2",
  "ps": "香港 VMess JSON",
  "add": "vmess.raw.example.test",
  "port": "443",
  "id": "00000000-0000-0000-0000-000000000085",
  "aid": "0",
  "scy": "auto",
  "net": "ws",
  "type": "none",
  "host": "ws.raw.example.test",
  "path": "/raw",
  "tls": "tls",
  "sni": "vmess.raw.example.test",
  "alpn": ["h2", "http/1.1"],
  "packetEncoding": "packetaddr",
  "disable_sni": "1",
  "fp": "chrome",
  "allowInsecure": "1"
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	assertHasPrefix(t, got, "vmess://")
	decoded := decodeVMessURIForTest(t, got)
	for _, want := range []string{
		`"ps":"香港 VMess JSON"`,
		`"add":"vmess.raw.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000085"`,
		`"net":"ws"`,
		`"type":"none"`,
		`"host":"ws.raw.example.test"`,
		`"path":"/raw"`,
		`"tls":"tls"`,
		`"sni":"vmess.raw.example.test"`,
		`"alpn":"h2,http/1.1"`,
		`"packet_encoding":"packetaddr"`,
		`"disable_sni":"1"`,
		`"fp":"chrome"`,
		`"allowInsecure":"1"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected raw vmess document to contain %s: %q", want, decoded)
		}
	}
}

func TestNormalizeContentVMessJSONArrayAndObjectMap(t *testing.T) {
	raw := `[
  {
    "ps": "东京 VMess JSON",
    "add": "tokyo.raw.example.test",
    "port": 443,
    "id": "00000000-0000-0000-0000-000000000086"
  },
  {
    "大阪 VMess 映射": {
      "add": "osaka.raw.example.test",
      "port": "8443",
      "id": "00000000-0000-0000-0000-000000000087",
      "net": "grpc",
      "path": "fluxgate-grpc",
      "tls": true
    }
  }
]`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 URIs, got %d: %q", len(lines), got)
	}
	first := decodeVMessURIForTest(t, lines[0])
	if !strings.Contains(first, `"ps":"东京 VMess JSON"`) ||
		!strings.Contains(first, `"add":"tokyo.raw.example.test"`) ||
		!strings.Contains(first, `"port":"443"`) {
		t.Fatalf("unexpected vmess JSON array document: %q", first)
	}
	second := decodeVMessURIForTest(t, lines[1])
	if !strings.Contains(second, `"ps":"大阪 VMess 映射"`) ||
		!strings.Contains(second, `"add":"osaka.raw.example.test"`) ||
		!strings.Contains(second, `"net":"grpc"`) ||
		!strings.Contains(second, `"path":"fluxgate-grpc"`) ||
		!strings.Contains(second, `"tls":"tls"`) {
		t.Fatalf("unexpected vmess JSON object-map document: %q", second)
	}
}

func TestNormalizeContentVMessJSONFieldAliases(t *testing.T) {
	raw := `{
  "v": "2",
  "displayName": "香港 VMess 别名",
  "address": "vmess-alias.raw.example.test",
  "serverPort": 443,
  "uuid": "00000000-0000-0000-0000-000000000089",
  "alterId": 0,
  "cipher": "auto",
  "network": "ws",
  "host": "ws.vmess-alias.example.test",
  "path": "/alias",
  "security": "tls",
  "server-name": "vmess-alias.raw.example.test",
  "packet-encoding": "packetaddr",
  "disableSNI": true,
  "clientFingerprint": "chrome",
  "skip-cert-verify": true
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	decoded := decodeVMessURIForTest(t, got)
	for _, want := range []string{
		`"ps":"香港 VMess 别名"`,
		`"add":"vmess-alias.raw.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000089"`,
		`"aid":"0"`,
		`"net":"ws"`,
		`"host":"ws.vmess-alias.example.test"`,
		`"path":"/alias"`,
		`"tls":"tls"`,
		`"sni":"vmess-alias.raw.example.test"`,
		`"packet_encoding":"packetaddr"`,
		`"disable_sni":"1"`,
		`"fp":"chrome"`,
		`"allowInsecure":"1"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected raw vmess alias document to contain %s: %q", want, decoded)
		}
	}
}

func TestNormalizeContentVMessJSONFieldCaseAliases(t *testing.T) {
	raw := `{
  "V": "2",
  "Display_Name": "香港 VMess 大小写别名",
  "Address": "vmess-case.raw.example.test",
  "Server-Port": 443,
  "User_ID": "00000000-0000-0000-0000-000000000090",
  "Alter-ID": 0,
  "Cipher": "auto",
  "Network": "ws",
  "Host": "ws.vmess-case.example.test",
  "Path": "/case",
  "Security": "tls",
  "ServerName": "vmess-case.raw.example.test",
  "ALPN": ["h2", "http/1.1"],
  "Packet_Encoding": "packetaddr",
  "Disable-SNI": true,
  "Client-Fingerprint": "chrome",
  "Allow_Insecure": true
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	decoded := decodeVMessURIForTest(t, got)
	for _, want := range []string{
		`"ps":"香港 VMess 大小写别名"`,
		`"add":"vmess-case.raw.example.test"`,
		`"port":"443"`,
		`"id":"00000000-0000-0000-0000-000000000090"`,
		`"aid":"0"`,
		`"net":"ws"`,
		`"host":"ws.vmess-case.example.test"`,
		`"path":"/case"`,
		`"tls":"tls"`,
		`"sni":"vmess-case.raw.example.test"`,
		`"alpn":"h2,http/1.1"`,
		`"packet_encoding":"packetaddr"`,
		`"disable_sni":"1"`,
		`"fp":"chrome"`,
		`"allowInsecure":"1"`,
	} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("expected raw vmess case alias document to contain %s: %q", want, decoded)
		}
	}
}

func TestNormalizeContentJSONWrappedVMessJSON(t *testing.T) {
	raw := `{
  "data": {
    "raw_content": "{\"ps\":\"内嵌 VMess JSON\",\"add\":\"embedded.raw.example.test\",\"port\":\"443\",\"id\":\"00000000-0000-0000-0000-000000000088\",\"tls\":\"tls\"}"
  }
}`
	got, err := NormalizeContent(raw)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	decoded := decodeVMessURIForTest(t, got)
	if !strings.Contains(decoded, `"ps":"内嵌 VMess JSON"`) ||
		!strings.Contains(decoded, `"add":"embedded.raw.example.test"`) ||
		!strings.Contains(decoded, `"tls":"tls"`) {
		t.Fatalf("unexpected wrapped vmess JSON document: %q", decoded)
	}
}

func TestNormalizeContentRejectsUnsupportedContent(t *testing.T) {
	if _, err := NormalizeContent("not a subscription"); err == nil {
		t.Fatal("expected unsupported content error")
	}
}

func decodeVMessURIForTest(t *testing.T, value string) string {
	t.Helper()
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, "vmess://"))
	if err != nil {
		t.Fatalf("failed to decode vmess URI: %v", err)
	}
	return string(decoded)
}

func assertHasPrefix(t *testing.T, value, prefix string) {
	t.Helper()
	if !strings.HasPrefix(value, prefix) {
		t.Fatalf("expected %q to have prefix %q", value, prefix)
	}
}
