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
  - { name: "东京 01", type: trojan, server: trojan.example.test, port: 443, password: "trojan-placeholder", sni: edge.example.test, skip-cert-verify: true, disable-sni: true, alpn: "h2,http/1.1", network: ws, ws-path: /trojan, ws-headers.host: ws.trojan.example.test }
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
    sni: hy2.clash.example.test
    skip-cert-verify: true
  - { name: "大阪 03", type: tuic, server: tuic.clash.example.test, port: 443, uuid: "00000000-0000-0000-0000-000000000055", password: "tuic-placeholder", congestion-controller: bbr, udp-relay-mode: native, sni: tuic.clash.example.test }
  - name: "香港 04"
    type: hysteria
    server: hysteria.clash.example.test
    port: 443
    auth-str: "hysteria-auth"
    up-mbps: 20
    down-mbps: 80
    obfs: "obfs-placeholder"
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
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ss.example.test:8388#")
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
	if !strings.Contains(lines[3], "obfs=salamander") || !strings.Contains(lines[3], "obfs-password=obfs-placeholder") || !strings.Contains(lines[3], "insecure=1") {
		t.Fatalf("unexpected hysteria2 URI: %q", lines[3])
	}
	assertHasPrefix(t, lines[4], "tuic://00000000-0000-0000-0000-000000000055:tuic-placeholder@tuic.clash.example.test:443?")
	if !strings.Contains(lines[4], "congestion_control=bbr") || !strings.Contains(lines[4], "udp_relay_mode=native") {
		t.Fatalf("unexpected tuic URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "hysteria://hysteria-auth@hysteria.clash.example.test:443?")
	if !strings.Contains(lines[5], "up_mbps=20") || !strings.Contains(lines[5], "down_mbps=80") || !strings.Contains(lines[5], "network=udp") || !strings.Contains(lines[5], "insecure=1") {
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
      "password": "qa-placeholder"
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
	assertHasPrefix(t, lines[0], "ss://aes-256-gcm:qa-placeholder@sip008.example.test:8388#")
	if !strings.HasSuffix(lines[0], "#%E9%A6%99%E6%B8%AF%2002") {
		t.Fatalf("unexpected SIP008 URI fragment: %q", lines[0])
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
      "password": "qa-placeholder"
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
          "Host": "ws.vless.singbox.example.test"
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
      "obfs": {
        "type": "salamander",
        "password": "obfs-placeholder"
      },
      "tls": {
        "enabled": true,
        "server_name": "hy2.singbox.example.test",
        "insecure": true,
        "alpn": ["h3"]
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
        "alpn": ["h3"]
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
        "server_name": "anytls.singbox.example.test"
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
        "alpn": ["h3"]
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
	assertHasPrefix(t, lines[0], "ss://aes-128-gcm:qa-placeholder@ss.singbox.example.test:8388#")
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
		!strings.Contains(lines[2], "host=ws.vless.singbox.example.test") {
		t.Fatalf("unexpected sing-box vless URI: %q", lines[2])
	}
	assertHasPrefix(t, lines[3], "vmess://")
	assertHasPrefix(t, lines[4], "hysteria2://hy2-placeholder@hy2.singbox.example.test:443?")
	if !strings.Contains(lines[4], "obfs=salamander") || !strings.Contains(lines[4], "obfs-password=obfs-placeholder") {
		t.Fatalf("unexpected sing-box hysteria2 URI: %q", lines[4])
	}
	assertHasPrefix(t, lines[5], "tuic://00000000-0000-0000-0000-000000000054:tuic-placeholder@tuic.singbox.example.test:443?")
	if !strings.Contains(lines[5], "congestion_control=bbr") || !strings.Contains(lines[5], "udp_relay_mode=native") {
		t.Fatalf("unexpected sing-box tuic URI: %q", lines[5])
	}
	assertHasPrefix(t, lines[6], "anytls://anytls-placeholder@anytls.singbox.example.test:443?")
	if !strings.Contains(lines[6], "idle_session_check_interval=20s") || !strings.Contains(lines[6], "min_idle_session=2") {
		t.Fatalf("unexpected sing-box anytls URI: %q", lines[6])
	}
	assertHasPrefix(t, lines[7], "shadowtls://shadow-placeholder@shadow.singbox.example.test:443?")
	if !strings.Contains(lines[7], "version=3") || !strings.Contains(lines[7], "sni=shadow.singbox.example.test") {
		t.Fatalf("unexpected sing-box shadowtls URI: %q", lines[7])
	}
	assertHasPrefix(t, lines[8], "hysteria://hysteria-auth@hysteria.singbox.example.test:443?")
	if !strings.Contains(lines[8], "up_mbps=20") || !strings.Contains(lines[8], "down_mbps=80") || !strings.Contains(lines[8], "network=udp") {
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

func TestNormalizeContentRejectsUnsupportedContent(t *testing.T) {
	if _, err := NormalizeContent("not a subscription"); err == nil {
		t.Fatal("expected unsupported content error")
	}
}

func assertHasPrefix(t *testing.T, value, prefix string) {
	t.Helper()
	if !strings.HasPrefix(value, prefix) {
		t.Fatalf("expected %q to have prefix %q", value, prefix)
	}
}
