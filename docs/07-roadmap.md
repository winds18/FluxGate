# 实施路线图

## Phase 0: 文档和技术验证

目标：

- 确认控制面/数据面边界。
- 确认 sing-box 统计能力和 Docker 镜像。
- 确认部署端口、域名、反代方案。

交付：

- 项目设计文档。
- sing-box 最小可运行配置样例。
- VLESS 用户统计验证。
- 配置发布和重启验证。

验收：

- 一个测试用户可以通过 sing-box 代理。
- 可以按 user 采集到上传/下载流量。
- 用户从配置中移除后不能继续新建连接。

## Phase 1: MVP 控制面

目标：

- 做出可用管理后台。
- 打通用户、Token、节点、策略、订阅、配置生成闭环。

当前已落地 Phase 1 foundation：

- Go API 服务骨架。
- SQLite migration。
- 嵌入式管理后台只读仪表盘。
- 团队、用户、Token、上游来源、节点导入、虚拟节点基础 API。
- 策略列表和创建 API。
- 策略 `allowed_virtual_nodes` 和 `max_nodes` 按 Token > 成员 > 团队优先级限制可见虚拟节点，并同步作用于 sing-box 入站用户分配。
- 策略 `include_tags` 和 `exclude_tags` 按 Token > 成员 > 团队优先级限制上游出口，并生成带 `auth_user` 的 sing-box route rule。
- Token 续期、追加额度、撤销和恢复操作。
- 来源前缀自动生成和重复前缀编号。
- 来源默认标签会同步到导入的上游节点。
- 节点手动改名保护和恢复自动命名。
- 虚拟节点 `tag_selector` 可按上游标签生成专属 sing-box selector 和 route rule。
- Clash/Mihomo 与 sing-box 基础订阅生成。
- sing-box config.json 生成骨架。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2/Hysteria/TUIC/AnyTLS/ShadowTLS/Naive/HTTP/SOCKS/SSH/WireGuard/Tor 上游节点会进入 sing-box outbound，并由默认 selector 承接出口。
- Direct URI 和 sing-box JSON `direct` outbound 可进入默认上游 selector 和 outbound stats 列表。
- Block URI、Clash `reject`/`reject-drop` 和 sing-box JSON `block` outbound 可进入 sing-box 配置，但不会进入普通上游 selector 或 outbound stats 列表。
- DNS URI 和 sing-box JSON `dns` outbound 可进入 sing-box 配置，但不会进入普通上游 selector 或 outbound stats 列表。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan/VMess TLS 参数并同步进入 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS WebSocket 传输参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan WebSocket early data 参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS gRPC 传输参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan gRPC keepalive 参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan QUIC transport 并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTP transport 参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTPUpgrade transport 参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS REALITY 的 public key、short id 和 uTLS fingerprint 参数并同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan REALITY 的 public key、short id 和 uTLS fingerprint 参数并同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan WebSocket/gRPC 传输参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VMess gRPC 传输参数并同步为 sing-box outbound transport。
- Clash YAML 和 URI 导入链路会保留 Hysteria 窗口和 MTU 发现参数并同步为 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria2 上下行带宽、TLS 标志和证书 pin 参数并同步为 sing-box outbound。
- sing-box 服务端配置生成会过滤不可用 Token，避免已撤销、已过期或已超额用户继续进入网关配置。
- 本地 smoke、API flow、页面截图验收脚本。
- 远程部署探测、磁盘清理、诊断采集脚本骨架。

范围：

- Go API 项目骨架。
- React/Vite 管理后台。
- SQLite migration。
- 管理员登录。
- 团队/用户/Token CRUD。
- 上游节点手动导入。
- subscription 类型来源手动刷新导入。
- subscription 来源刷新可识别当前已支持 outbound 的 URI 协议列表。
- subscription 来源可解析 JSON URI 数组和常见 `nodes`/`links`/`subscription`/`raw_content` 包装对象，包装字段中的内嵌 Clash YAML、SIP008、sing-box JSON 和 base64 订阅内容也会归一化。
- JSON 包装订阅可解析 `payload`、`result`、`response`、`body`、`text`、`sub` 等常见接口外层字段。
- JSON 包装订阅可解析 `nodes`/`proxies` 中的 Clash 风格结构化节点对象。
- JSON 结构化节点可识别 `protocol`、`host`、`address`、`server_port`、`method`、`pass` 和 `remarks` 等常见别名字段。
- subscription 来源可解析 Quantumult X `[server_local]`/`[server_remote]` 常见 SS、Trojan、VLESS、VMess、HTTP 和 SOCKS 节点。
- subscription 来源可解析以节点名称为 key 的 JSON URI 对象映射。
- subscription 来源可解析裸 VMess JSON 单对象、数组、包装字符串和按名称映射对象，并转换为标准 `vmess://` URI。
- subscription 来源可解析 SSD/ShadowsocksD `ssd://` 订阅和裸 SSD JSON，并展开为标准 `ss://` URI。
- subscription 来源可解析 Surge `[Proxy]` 代理段中的 SS、Trojan、VLESS、VMess、Hysteria2、TUIC、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、HTTP/HTTPS、SOCKS、Direct、Reject 和 DNS 节点，并转换为标准 URI。
- subscription 来源支持按刷新间隔自动同步到期订阅。
- subscription 来源可解析 Clash YAML `proxies` 中的常见 SS/Trojan/VLESS/VMess/Hysteria2/TUIC/Hysteria/HTTP/SOCKS/AnyTLS/ShadowTLS/Naive/SSH/WireGuard 节点，以及 Direct/Reject 变体/DNS 内置出站。
- Clash YAML 解析支持嵌套 `proxies` 列表，可兼容带内嵌 provider 节点清单的订阅结构。
- Clash YAML 解析支持 `proxies: [{ ... }]` 内联节点数组，可兼容 provider 或顶层节点的紧凑写法。
- Clash YAML 解析支持 `proxies: &anchor` 这类带 YAML anchor 的块状节点列表。
- Clash YAML 解析支持 `- &anchor { ... }` 和 `proxies: [&anchor { ... }]` 这类节点条目级 anchor 的内联写法。
- Clash YAML 解析支持常见块状和内联 `ws-opts`、`grpc-opts` 嵌套写法，能保留 WebSocket path/host 和 gRPC service name。
- Clash YAML 解析支持行内和块状数组标量，可保留 ALPN、WireGuard 地址、允许 IP 和 reserved 字节列表。
- subscription 来源可解析 SIP008 Shadowsocks 订阅，`servers` 支持数组和按名称分组的对象映射。
- SIP008、Clash YAML 和 sing-box JSON 的 Shadowsocks 节点会保留 SIP003 插件参数并同步进入 sing-box outbound。
- subscription 来源可解析 sing-box JSON 顶层 `endpoints` 中的 WireGuard endpoint 模型，并转换为当前 Phase 1 WireGuard outbound 兼容 URI。
- sing-box JSON `outbounds` 和 `endpoints` 支持数组、单个对象、按名称分组的对象映射和分组数组映射。
- subscription 来源可解析 V2Ray/Xray JSON `outbounds` 中的 VMess、VLESS、Trojan、Shadowsocks、HTTP、SOCKS、freedom、blackhole 和 DNS 出站，并保留常见 TLS、REALITY、WebSocket、gRPC、HTTP/H2 和 HTTPUpgrade 传输参数。
- 节点来源前缀和展示名规则。
- 标签管理。
- 虚拟节点管理。
- 策略管理基础列表和创建。
- Clash/Mihomo 订阅生成。
- sing-box config.json 生成。
- `sing-box check` 校验。
- 管理后台可触发 sing-box config check 并查看配置摘要。
- 管理后台可发布 sing-box config 文件并保存上一版。
- 管理后台可回滚到上一版 sing-box config 文件。
- 管理后台可触发受配置保护的 sing-box 重启执行器，发布/回滚可按部署配置自动重启。
- 发布配置并重启容器。

验收：

- 管理员能创建用户和 Token。
- 用户能通过订阅导入客户端。
- 客户端能连上 FluxGate 网关。
- 流量通过 sing-box 转发到上游。
- 禁用用户后新连接失败。

## Phase 2: 真实流量统计和额度

目标：

- 建立可靠的统计采集和额度控制。

当前已开始落地：

- V2Ray stats user/inbound/outbound 计数名称解析。
- `traffic_samples` 增量样本落库、counter reset 处理、Token 用量累加。
- 采集入库后自动标记 `over_quota`，追加足够额度后自动恢复 Token 和 gateway account。
- 用户小时/天级流量汇总表和 Token 今日、本月、累计流量摘要 API。
- 管理后台基于小时/日汇总 API 展示最近 24 小时和最近 14 天流量图，并展示近 14 天上游出口流量摘要。
- 订阅响应头返回标准 `subscription-userinfo` 和 FluxGate 用量、额度、剩余额度头。
- stats 可插拔轮询调度器，可把采集器返回的 V2Ray counters 转换为流量样本并写入现有汇总链路。
- 真实 sing-box V2Ray gRPC stats 采集器已接入主进程，可从 Docker 内网的 V2Ray API 轮询 counters。
- sing-box 配置生成会把 active 上游 outbound tag 纳入 stats 配置，支撑上游出口维度流量汇总。
- 部署侧默认远程构建带 `with_v2ray_api` 的自定义 sing-box 镜像，保证统计 API 可用。

范围：

- sing-box stats 采集器。
- user/inbound/outbound 维度原始样本。
- 小时/天级汇总。
- Dashboard 流量图。
- Token 已用流量。
- 流量额度判断。
- 超额自动阻断。
- 订阅响应头返回流量信息。

验收：

- 能按 Token 查看今日、本月、累计流量。
- 超额 Token 自动变为 `over_quota`。
- 超额用户不能继续新建连接。
- 追加额度后可以恢复。

## Phase 3: 上游订阅自动同步

目标：

- 管理多个机场订阅。
- 自动解析、去重、更新节点池。

范围：

- 添加 subscription 类型来源。
- 定时拉取上游订阅。
- 解析多协议 URI。
- URI hash 去重。
- 节点失效标记。
- 同步日志。
- 默认标签。

验收：

- 可以添加多个机场订阅。
- 定时同步后节点池自动更新。
- 重复节点不会重复入库。
- 禁用来源后对应节点不再出现在新配置中。

## Phase 4: 节点探针和路由优化

目标：

- 提升可用性和出口质量。

范围：

- 节点连通性检测。
- 延迟检测。
- 失败次数统计。
- 自动禁用连续失败节点。
- urltest/fallback/selector 策略生成。
- 按地区和标签自动生成出口组。

验收：

- 不可用节点不会被分配给用户。
- 出口组可以自动选择延迟较低节点。
- Dashboard 能展示节点健康状态。

## Phase 5: 运维增强

目标：

- 提升部署和长期运行稳定性。

范围：

- 自动备份。
- 一键恢复。
- 配置版本回滚。
- 远程镜像构建。
- 部署前系统环境探测。
- 磁盘水位检查和安全清理。
- 部署诊断日志采集。
- 页面截图验收链路。
- 操作审计查询。
- 多管理员角色。
- 限速策略。
- 异常流量告警。
- 到期提醒。

验收：

- 误发布配置可以回滚。
- 数据库可恢复。
- 磁盘 critical 时部署会停止或先安全清理。
- 部署失败时可以自动生成诊断包。
- 管理后台页面变更有截图验收记录。
- 管理员操作可追踪。
- 异常用量可发现。

## 优先级建议

先做：

```text
Phase 0 -> Phase 1 -> Phase 2
```

这三阶段完成后，FluxGate 就具备真实运营价值。

后做：

```text
Phase 3 -> Phase 4 -> Phase 5
```

这些阶段提升效率和稳定性，但不阻塞最小闭环。
