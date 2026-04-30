# 当前实现状态

## 1. 已落地内容

本轮完成 Phase 1 foundation，目标是建立可运行、可测试、可继续扩展的控制面骨架。

已实现：

- Go HTTP API 服务。
- SQLite migration 自动执行。
- 嵌入式 Web 管理后台登录页、仪表盘和基础写操作表单。
- 管理后台已做基础视觉统一：统一顶部栏、卡片、表单、按钮、表格、状态徽标和响应式栅格，输入控件高度保持一致。
- 管理员引导账号。
- PBKDF2-SHA256 密码哈希。
- 管理员签名 Cookie 会话。
- 管理 API 登录保护。
- 会话 Cookie 按实际请求协议设置 Secure，支持 HTTPS 反代和 HTTP 局域网调试。
- 团队、用户、Token 基础页面创建和列表。
- 策略基础页面创建和列表。
- 策略 `allowed_virtual_nodes` 和 `max_nodes` 已按 Token > 成员 > 团队优先级生效，用于限制订阅输出中的可见虚拟节点，并同步约束 sing-box 入站里的用户分配；被策略挡住且没有可用用户的虚拟节点不会生成入站。
- 策略 `include_tags` 和 `exclude_tags` 已按 Token > 成员 > 团队优先级生效，会为匹配 Token 生成带 `auth_user` 的 sing-box route rule，限制该 Token 在虚拟节点下可走的上游出口。
- Token 支持续期、追加额度、撤销和恢复，并同步 gateway account 状态。
- 上游来源页面创建和列表。
- 管理后台上游来源支持行内编辑名称、类型、URL、前缀、默认标签和刷新间隔；前缀变更会同步刷新自动命名节点。
- subscription 类型上游来源可保存 URL 或 raw content，并可手动刷新导入节点。
- 上游来源 `default_tags` 会在节点导入或刷新时同步到节点标签。
- subscription 来源刷新时，本次订阅中消失的旧节点会标记为 `inactive`。
- subscription 来源刷新可识别当前已支持 outbound 的 URI 协议列表。
- subscription 来源支持解析 JSON URI 数组和常见 `nodes`/`links`/`subscription`/`raw_content` 包装对象，包装字段内的 URI 列表、Clash YAML、SIP008、sing-box JSON 和 base64 内嵌订阅内容也会归一化。
- JSON 包装订阅支持 `payload`、`result`、`response`、`body`、`text`、`sub` 等常见接口外层字段，不会把这些包装字段误当成节点名称。
- JSON 包装订阅支持解析 `nodes`/`proxies` 中的 Clash 风格结构化节点对象，例如 `type`、`server`、`port`、`cipher`、`password` 和嵌套 `ws-opts`。
- JSON 结构化节点支持常见字段别名，例如 `protocol`、`host`、`address`、`server_port`、`method`、`pass` 和 `remarks`。
- subscription 来源支持解析 Quantumult X `[server_local]`/`[server_remote]` 常见 SS、Hysteria2/Hy2、TUIC、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、Trojan、VLESS、VMess、HTTP 和 SOCKS 节点，并转换为标准 URI。
- Quantumult X 和 Surge 结构化 VLESS/Trojan 节点支持保留 gRPC transport 和 service name，并同步到 sing-box outbound。
- Quantumult X HTTP/SOCKS 节点支持键值和位置参数两种认证写法。
- subscription 来源支持解析以节点名称为 key 的 JSON URI 对象映射；当 URI 缺少 fragment 时会使用映射 key 或对象内 `name` 作为节点名。
- subscription 来源支持解析裸 VMess JSON 单对象、数组、包装字符串和按名称映射对象，并转换为标准 `vmess://` URI。
- subscription 来源支持解析 SSD/ShadowsocksD `ssd://` 订阅和裸 SSD JSON，并展开为标准 `ss://` URI。
- subscription 来源支持解析 Surge `[Proxy]` 代理段中的 SS、Trojan、VLESS、VMess、Hysteria2、TUIC、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、HTTP/HTTPS、SOCKS、Direct、Reject 和 DNS 节点，并转换为标准 URI。
- subscription 来源支持解析 Clash YAML `proxies` 中的常见 SS/Trojan/VLESS/VMess/Hysteria2/TUIC/Hysteria/HTTP/SOCKS/AnyTLS/ShadowTLS/Naive/SSH/WireGuard 节点，以及 Direct/Reject 变体/DNS 内置出站。
- Clash YAML 解析支持嵌套 `proxies` 列表，可兼容带内嵌 provider 节点清单的订阅结构。
- Clash YAML 解析支持 `proxies: [{ ... }]` 内联节点数组，可兼容 provider 或顶层节点的紧凑写法。
- Clash YAML 解析支持 `proxies: &anchor` 这类带 YAML anchor 的块状节点列表。
- Clash YAML 解析支持 `- &anchor { ... }` 和 `proxies: [&anchor { ... }]` 这类节点条目级 anchor 的内联写法。
- Clash YAML 解析支持常见块状和内联 `ws-opts`、`grpc-opts` 嵌套写法，能保留 WebSocket path/host 和 gRPC service name。
- Clash YAML 解析支持行内和块状数组标量，例如 `alpn: [h2, http/1.1]`、`alpn: ... - h2` 和 WireGuard `allowed-ips`/`reserved` 列表。
- subscription 来源支持解析 SIP008 Shadowsocks 订阅，`servers` 支持数组和按名称分组的对象映射。
- SIP008、Clash YAML 和 sing-box JSON 的 Shadowsocks 节点会保留 SIP003 `plugin`、`plugin_opts` 和 `network` 参数，并同步到 sing-box outbound。
- subscription 来源支持解析 sing-box JSON `outbounds` 中的常见 Shadowsocks/Trojan/VLESS/VMess/Hysteria2/TUIC/AnyTLS/ShadowTLS/Naive/Hysteria/HTTP/SOCKS/SSH/WireGuard/Tor 节点。
- subscription 来源支持解析 sing-box JSON 顶层 `endpoints` 中的 WireGuard endpoint 模型，并转换为当前 Phase 1 WireGuard outbound 兼容 URI。
- sing-box JSON `outbounds` 和 `endpoints` 支持数组、单个对象、按名称分组的对象映射和分组数组映射；映射对象缺少 `tag`/`name` 时会使用映射键或分组序号作为节点名。
- sing-box JSON transport 的 `headers.Host` 支持字符串或数组写法，WebSocket 和 HTTPUpgrade 会保留为逗号分隔 Host。
- subscription 来源支持解析 V2Ray/Xray JSON `outbounds` 中的 VMess、VLESS、Trojan、Shadowsocks、HTTP、SOCKS、freedom、blackhole 和 DNS 出站，并保留常见 TLS、REALITY、WebSocket、gRPC、QUIC、HTTP/H2 和 HTTPUpgrade 传输参数；REALITY 兼容 `serverName`/`serverNames`、`shortId`/`shortIds` 和 hyphen/snake-case 别名。
- V2Ray/Xray JSON `streamSettings`、TLS/REALITY 设置和各 transport 设置块支持 camelCase、snake_case 和 hyphen-case 容器字段别名。
- V2Ray/Xray JSON Shadowsocks outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`、`security`/`encryption` 字段别名，并保留 SIP003 `plugin`、`plugin_opts`、`plugin_options`、`network` 和 `protocol` 标志，结构化导入后会继续进入 Shadowsocks 转换链路。
- V2Ray/Xray JSON 普通 TLS 会兼容 `serverName`/`serverNames` 以及 hyphen/snake-case SNI 别名，数组写法会选取首个有效值同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会兼容 `allowInsecure`、`skip-cert-verify`、`insecure` 和 `disable_sni` 写法，并同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会保留 `fingerprint`、`clientFingerprint` 或 `utls.fingerprint` 客户端指纹，并同步为 sing-box uTLS fingerprint。
- V2Ray/Xray JSON `wsSettings.host`、`wsSettings.headers.Host` 和 `authority` 支持字符串或数组写法，数组会保留为逗号分隔的 WebSocket Host。
- V2Ray/Xray JSON `wsSettings` 支持 `maxEarlyData`、`max_early_data`、`max-early-data`、`earlyDataHeaderName`、`early_data_header_name` 和 `early-data-header-name` WebSocket early data 别名。
- V2Ray/Xray JSON `tcpSettings.header.type=http` 的 HTTP 伪装 Host、path 和 method 会保留为 HTTP transport 参数。
- V2Ray/Xray JSON `httpSettings`/`h2Settings` 支持从 `host`、`headers.Host` 和 `authority` 读取 HTTP/H2 transport Host，并支持数组写法的 path。
- V2Ray/Xray JSON `httpupgradeSettings.host` 支持字符串或数组写法，也可从 `headers.Host` 读取 Host 列表。
- V2Ray/Xray JSON `grpcSettings` 的 `idle_timeout`、`health_check_timeout` 和 `permit_without_stream` 会保留为 gRPC transport keepalive 参数。
- V2Ray/Xray JSON `grpcSettings.multiMode` 会保留为 sing-box gRPC transport 的 `multi_mode` 参数。
- V2Ray/Xray JSON `grpcSettings` 支持 `service-name`、`idle-timeout`、`health-check-timeout`、`permit-without-stream` 和 `multi-mode` 这类 hyphen 别名。
- V2Ray/Xray JSON 存在 `quicSettings`/`quic_settings`/`quic-settings` 时会保留为 QUIC transport。
- V2Ray/Xray JSON VMess/VLESS `vnext` endpoint 支持 `address`、`server`、`host`、`add` 和 `server_port`/`serverPort`/`server-port` 别名；VMess user 支持 `alter-id`/`aid` 和 `cipher` 别名，VMess/VLESS user 支持 `packetEncoding`/`packet_encoding`/`packet-encoding` 并同步为 sing-box `packet_encoding`。
- Clash YAML 和 sing-box JSON VMess/VLESS 节点支持 `packetEncoding`、`packet_encoding` 和 `packet-encoding` 写法，并同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON Trojan/HTTP/SOCKS outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`、`users`、`accounts` 数组、`accounts` 用户名到密码映射和 server 层 `username`/`password` 认证写法。
- subscription 来源支持按 `refresh_interval_minutes` 定时同步；后台调度默认每 60 秒检查一批到期来源。
- 上游来源自动前缀。
- 重复来源名前缀自动编号，例如 `[机场A]`、`[机场A-2]`。
- 节点 URI 页面批量导入。
- 节点 `raw_name`、`display_name`、`name_mode`。
- 管理 API 和管理后台可查看单个节点详情，包括 URI、服务器、端口、Hash、状态、标签和时间信息。
- 管理后台所有时间字段统一按东八区展示，后端仍以 UTC 作为存储和计算基准。
- 管理后台节点池默认按地区聚合展示，缺失地区的节点归入“其他”；进入地区后展示节点卡片，再点击卡片查看节点详情。
- 节点导入、读取和历史数据迁移会识别常见地区并统一带国旗展示；中国香港/中国台湾统一归为 `🇨🇳中国|香港`、`🇨🇳中国|台湾`，`香港-美国` 这类路径命名会按右侧目的地归为 `🇺🇸美国`。
- 单节点手动改名保护。
- 单节点恢复自动命名。
- 管理后台节点池支持行内编辑节点展示名，并可恢复来源前缀驱动的自动命名。
- 虚拟节点页面基础创建和列表。
- 虚拟节点 `tag_selector` 可按上游节点标签生成专属 sing-box selector，并生成入站到该 selector 的 route rule；无匹配上游时会路由到 `block`。
- Clash/Mihomo 订阅生成。
- sing-box 客户端订阅生成。
- sing-box 服务端配置生成骨架。
- sing-box config check API 和管理后台检查入口，可返回配置 hash、入站/出站/用户摘要。
- sing-box config publish API 和管理后台发布入口，可写入当前配置并保存上一版文件。
- sing-box config rollback API 和管理后台回滚入口，可恢复上一版配置文件。
- sing-box config publish/rollback API 会返回 `restart_required`，管理后台会提示发布或回滚后需要重启 sing-box。
- sing-box restart API 和管理后台重启入口已落地；发布/回滚可在 `SING_BOX_AUTO_RESTART=true` 时自动调用受配置保护的重启命令。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2/Hysteria/TUIC/AnyTLS/ShadowTLS/Naive/HTTP/SOCKS/SSH/WireGuard/Tor 上游节点会转换为 sing-box outbound，并通过默认 selector 承接网关出口。
- Direct URI 和 sing-box JSON `direct` outbound 可导入为直连上游 outbound，并进入默认上游 selector 和 V2Ray outbound stats 列表。
- Block URI、Clash `reject`/`reject-drop` 和 sing-box JSON `block` outbound 可导入为内部拦截 outbound；为避免误承接普通代理流量，Block outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表。
- DNS URI 和 sing-box JSON `dns` outbound 可导入为内部 DNS outbound；为避免误承接普通代理流量，DNS outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan/VMess 的 `insecure`/`skip-cert-verify`、`disable_sni` 和 `alpn` TLS 参数，并同步到 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS WebSocket 的 `type=ws`、`path` 和 `host` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan WebSocket 的 `max_early_data` 和 `early_data_header_name` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS gRPC 的 `type=grpc` 和 `service_name` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan gRPC 的 `idle_timeout`、`ping_timeout` 和 `permit_without_stream` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan QUIC transport，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTP transport 的 `host`、`path`、`method`、`idle_timeout` 和 `ping_timeout` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTPUpgrade transport 的 `host` 和 `path` 参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS REALITY 的 public key、short id 和 uTLS fingerprint 参数，并同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan REALITY 的 public key、short id 和 uTLS fingerprint 参数，并同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan WebSocket/gRPC 传输参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VMess gRPC 传输参数，并同步为 sing-box outbound transport。
- VMess TCP HTTP 伪装和 HTTP/H2 传输参数会同步为 sing-box HTTP transport。
- sing-box JSON VMess 节点的 HTTP transport 会保留 host 数组和 path，避免导入后丢失 HTTP 伪装参数。
- VMess 可兼容 `vmess://uuid@host:port?...` userinfo 直连 URI，并保留 TLS、WebSocket、SNI、ALPN、跳过证书校验、uTLS fingerprint 和 `packet_encoding`/`packet-encoding`/`packetEncoding` 参数。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria `recv_window_conn`、`recv_window`、`disable_mtu_discovery` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria2 `up_mbps`、`down_mbps`、`insecure`、`disable_sni`、`alpn`、证书 pin、`fp`/`client-fingerprint` 和 `tls.utls.fingerprint` 参数，并同步为 sing-box outbound。
- sing-box JSON 和 URI 导入链路会保留 TUIC `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound。
- sing-box JSON 和 URI 导入链路会保留 AnyTLS `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound。
- sing-box JSON 和 URI 导入链路会保留 ShadowTLS `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound。
- sing-box JSON 和 URI 导入链路会保留 Naive/Naive+QUIC `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound。
- sing-box JSON 和 URI 导入链路会保留 HTTP/HTTPS 代理 `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound。
- SOCKS URI 支持 `udp`、`udp_relay` 和 `udp-relay` 标志，并会归一为 sing-box outbound 的 `network=udp`。
- Clash YAML SOCKS 节点支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Surge SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Quantumult X SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- sing-box JSON SOCKS outbound 支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- V2Ray/Xray JSON SOCKS outbound 支持保留 `udp`、`udpEnabled`、`udp_enabled`、`udp_relay`、`udp-relay`、`udp_over_tcp`、`udpOverTcp`、`udp-over-tcp` 和 `uot` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Hysteria2/Hy2 URI 支持从 `password`、`auth`、`auth_str` 或 `token` 查询参数读取认证密码，兼容缺少 userinfo 的订阅写法。
- Hysteria v1 URI 支持从 `token` 查询参数读取认证字符串，兼容部分上游订阅的 token 写法。
- sing-box 服务端配置生成会过滤已撤销、已过期、已超额或 gateway account 不可用的 Token。
- 流量统计入库骨架已落地：可解析 V2Ray stats 的 user/inbound/outbound 计数名称，写入 `traffic_samples`，计算 counter delta，并把 user 维度增量累加到 Token 用量和小时/天汇总表。
- user 维度流量入库后会自动判断 Token 额度；超额时将 Token 和 gateway account 标记为 `over_quota`，追加足够额度后自动恢复为 `active`。
- stats 轮询发现 Token 因真实流量进入 `over_quota` 时，会自动发布新的 sing-box 配置；如果部署侧开启 `SING_BOX_AUTO_RESTART`，会继续走受配置保护的重启执行器。
- 管理 API 和后台页面可查看 Token 今日、本月和累计流量用量摘要，展示最近 24 小时/最近 14 天流量图，并查看近 14 天上游出口流量摘要。
- 管理后台 Token 流量摘要可展示额度使用率进度条，并按接近超额和已超额状态变色。
- 订阅响应头返回标准 `subscription-userinfo`，并提供 FluxGate 专属的已用、总额和剩余额度头。
- stats 可插拔轮询调度器已落地，能将采集器返回的 V2Ray counters 转换为流量样本并写入现有用量汇总链路。
- 真实 sing-box V2Ray gRPC stats 采集器已接入主进程；配置 `SING_BOX_V2RAY_API_ADDR` 后会按 `STATS_POLL_INTERVAL_SECONDS` 轮询并写入现有统计链路。
- sing-box 服务端配置生成会把 active 上游 outbound tag 写入 V2Ray stats 配置，支持上游出口流量汇总。
- Token hash 存储，明文只在创建时返回。
- 订阅请求日志 Token 路径脱敏。
- 结构化 JSON 服务日志。
- 本地验证脚本。
- 公开仓库脱敏扫描脚本。
- 远程部署探测脚本。
- 磁盘清理脚本。
- 诊断采集脚本。
- sing-box 当前配置校验并重载脚本，适用于配置文件已由控制面写入但需要数据面重新加载的场景。
- 推送后远程部署脚本。
- GitHub Actions Docker 镜像构建工作流，支持 PR 构建验证和分支/tag 推送 GHCR。
- 自定义 sing-box Dockerfile 已落地，默认远程构建带 `with_v2ray_api` 的本地数据面镜像，避免官方镜像缺少 V2Ray API 导致统计采集不可用。
- Docker 构建上下文脱敏，默认排除 `.env`、数据库、日志、测试产物和本机部署配置。
- 远程验收健康检查重试。
- 首次远程部署最小 sing-box config bootstrap。
- 远程 Docker build 支持 `GOPROXY`，默认优先使用 `goproxy.cn` 以避开 `proxy.golang.org` 超时。
- 推送后远程部署支持远程构建优先；当外部 registry 网络不可用时，会回退为本地交叉编译、scratch 镜像构建并流式加载到服务器。
- 部署时宿主机 HTTP 端口默认使用 `127.0.0.1:18080`，避免和服务器已有 8080 服务冲突。
- 部署侧可通过未跟踪配置打开局域网访问，不把真实环境信息提交到公开仓库。
- 页面截图验收脚本，支持指定目标 URL、保留截图和自定义输出目录。
- 浏览器登录验收脚本。
- 本地 QA 套件退出自动清理临时产物。

## 2. 当前 API 骨架

已实现的主要接口：

```text
GET  /healthz
GET  /readyz
GET  /api/auth/session
POST /api/auth/login
POST /api/auth/logout
GET  /api/overview

GET  /api/teams
POST /api/teams

GET  /api/users
POST /api/users

GET  /api/tokens
POST /api/tokens
POST /api/tokens/{id}/revoke
POST /api/tokens/{id}/restore
POST /api/tokens/{id}/extend
POST /api/tokens/{id}/quota

GET   /api/sources
POST  /api/sources
PATCH /api/sources/{id}
POST  /api/sources/{id}/refresh
POST  /api/sources/{id}/regenerate-node-names

GET   /api/nodes
GET   /api/nodes/{id}
POST  /api/nodes/import
PATCH /api/nodes/{id}
POST  /api/nodes/{id}/reset-display-name

GET  /api/virtual-nodes
POST /api/virtual-nodes

GET  /api/policies
POST /api/policies
GET  /api/traffic/tokens
GET  /api/traffic/daily?days=14
GET  /api/traffic/hourly?hours=24
GET  /api/traffic/outbounds?days=14

POST /api/sing-box/config/generate
POST /api/sing-box/config/check
POST /api/sing-box/config/publish
POST /api/sing-box/config/rollback
POST /api/sing-box/restart
GET  /sub/{token}
```

## 3. 验证结果

已通过：

```text
scripts/dev/bootstrap.sh
scripts/dev/format.sh
scripts/dev/test.sh
scripts/dev/lint.sh
scripts/dev/build.sh
scripts/db/migrate.sh
scripts/qa/public-scan.sh
scripts/qa/smoke.sh
scripts/qa/api-flow.sh
scripts/qa/browser-login.sh
scripts/qa/screenshot.sh
scripts/qa/local-suite.sh
scripts/deploy/remote-logs.sh
scripts/deploy/probe-env.sh
scripts/deploy/push-and-deploy.sh
DRY_RUN=true scripts/deploy/cleanup-disk.sh
docker compose config
```

页面截图验收：

```text
KEEP_ARTIFACTS=true scripts/qa/screenshot.sh 可保留截图；默认测试退出会自动清理截图。
scripts/qa/screenshot.sh --keep http://<lan-host>:<port> 可对局域网部署页面保留人工复核截图。
scripts/qa/browser-login.sh http://<lan-host>:<port> 可做真实浏览器登录验收。
```

截图结论：

- 页面可打开。
- 未登录时展示管理员登录页。
- 登录后管理员登录表单必须不可见，后台视图必须可见，并可加载仪表盘数据。
- 无白屏。
- 无明显遮挡。
- 表格、卡片、按钮和输入控件未出现明显溢出或尺寸错位。
- 来源前缀、节点展示名和 Token 前缀正常展示。
- 有节点数据时，浏览器验收会确认节点池先展示地区聚合，进入地区后出现节点卡片。
- 有节点数据时，浏览器验收会确认节点卡片“编辑”入口可打开行内编辑表单。
- 有来源数据时，浏览器验收会确认上游来源“编辑”入口可打开行内编辑字段。
- 有节点数据时，浏览器验收会确认节点池“详情”入口可展开单节点详情。
- 浏览器验收会确认 RFC3339 和 SQLite 时间字符串都按东八区展示。
- 有 Token 流量行时，浏览器验收会确认额度使用率进度条可见。

## 4. 尚未完成

下一步需要继续实现：

- 上游订阅更多结构化格式解析。
- 更多特殊协议 URI 到 sing-box outbound 的转换。
- 远程服务器实际部署验证，相关连接信息仅保存在本机未跟踪配置中。
- 当前大版本收尾时交付一份手把手使用说明，覆盖部署初始化、管理员登录、添加上游订阅、节点地区聚合与详情查看、节点/来源编辑、团队/用户/Token 管理、订阅地址分发、流量额度查看、配置发布/回滚、远程运维和常见问题排查。

## 5. 当前注意事项

- `docker compose config` 已支持没有 `.env` 时做静态校验。
- 生产部署仍应由 `scripts/deploy/bootstrap-remote.sh` 生成 `.env` 后再修改密钥；管理员引导账号只在没有管理员记录时创建。
- 管理 API 需要管理员会话；订阅接口 `/sub/{token}` 继续使用订阅 Token 鉴权，不依赖管理员登录。
- 需要局域网访问管理后台时，通过部署侧配置 `FLUXGATE_HOST_BIND=0.0.0.0` 和 `FLUXGATE_HTTP_PORT`，公开仓库不保存真实访问地址。
- `SING_BOX_AUTO_RESTART` 默认关闭；如果要让控制面重启 sing-box，需要在部署侧显式挂载 Docker socket 或改用 `command` driver，并提供对应容器/宿主机权限。
- `SING_BOX_V2RAY_API_ADDR` 为空时 stats 轮询器不会启动；生产部署应绑定 Docker 内网地址，不应暴露到公网。
- sing-box 官方文档已标注 WireGuard outbound 废弃；当前已能导入 WireGuard endpoint 模型，但配置生成侧仍按 Phase 1 既有 outbound 骨架兼容输出，后续应评估完整迁移到 endpoint 配置。
- 本地 QA 产生的 `data/`、`logs/`、`tmp/` 均被 `.gitignore` 排除，并默认在测试退出时清理。
- Playwright Chromium 已在本机安装一次，后续截图脚本会复用缓存。
