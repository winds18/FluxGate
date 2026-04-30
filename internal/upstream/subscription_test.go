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
	if !strings.Contains(lines[6], "sni=http.clash.example.test") || !strings.Contains(lines[6], "insecure=1") || !strings.HasSuffix(lines[6], "#%E4%B8%9C%E4%BA%AC%20HTTP") {
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
        "insecure": true
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
        "insecure": true
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
	if !strings.Contains(lines[7], "version=3") || !strings.Contains(lines[7], "sni=shadow.singbox.example.test") {
		t.Fatalf("unexpected sing-box shadowtls URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "hysteria://hysteria-auth@hysteria.singbox.example.test:443?")
	if !strings.Contains(lines[8], "up_mbps=20") || !strings.Contains(lines[8], "down_mbps=80") || !strings.Contains(lines[8], "network=udp") || !strings.Contains(lines[8], "fp=chrome") {
		t.Fatalf("unexpected sing-box hysteria URI: %q", lines[8])
	}
	assertHasPrefix(t, lines[9], "https://qa-user:http-placeholder@http.singbox.example.test:8443/connect?")
	if !strings.Contains(lines[9], "sni=http.singbox.example.test") || !strings.Contains(lines[9], "insecure=1") {
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
