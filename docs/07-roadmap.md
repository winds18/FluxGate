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
- 管理后台已切换为模块化控制台布局，包含带模块符号的桌面侧边导航、7 模块移动底部 Dock、概览/接入/节点/身份/策略/流量/运维视图，并纳入浏览器导航、符号数量和溢出验收。
- 管理后台工作区标题已补齐当前模块符号，与侧边导航和移动 Dock 统一，并纳入浏览器逐模块切换验收。
- 管理后台桌面侧边导航和移动 Dock 已补齐模块数量徽标，把来源、节点、Token、策略、流量样本和配置就绪状态前置到导航层，并纳入浏览器数量、填充值和移动端溢出验收。
- 管理后台创建/导入表单默认收纳为模块抽屉，抽屉标题补齐统一功能符号，减少首屏视觉负担，并纳入浏览器符号顺序和溢出验收。
- 管理后台概览页顶部指标卡已补齐模块符号，团队、用户、Token、来源、节点、虚拟节点和策略指标具备统一视觉锚点，并纳入浏览器数量和值溢出验收。
- 管理后台 HTML 转义工具已兼容数值 ID 等真实数据字段，登录浏览器 QA 会等待完整数据加载完成，并在异常时输出失败接口与页面诊断，避免“登录成功但后台空白/异常”难以定位。
- 管理后台概览页已加入真实测试闭环看板，可直接查看来源、节点池、虚拟网关、团队 Token 和访问策略是否齐备；每个闭环步骤都有序号、符号化状态单元和就绪/待补徽标，并纳入浏览器序号、符号、状态徽标和溢出验收。
- 管理后台概览页快速入口卡已加入模块符号、实时数量和就绪徽标，来源、地区、Token 和发布状态在概览首屏可直接扫描，并纳入浏览器符号、数量和溢出验收。
- 管理后台概览页下一步入口已改为结构化动作卡，目标模块符号、状态说明和主操作按钮分层展示，并纳入浏览器溢出验收。
- 管理后台模块标题区已加入随视图变化的上下文摘要条，帮助在进入每个模块前先确认关键状态；摘要胶囊会把标签和值约束在自身内部，并纳入浏览器溢出验收。
- 管理后台模块标题区已加入随视图变化的快速定位条，可直接跳转到模块内列表、详情和创建抽屉；创建入口会自动展开对应表单。
- 管理后台模块快速定位条已补齐功能符号、数量和新增徽标，让入口既能导航也能展示当前模块规模，且符号、文字和徽标会被约束在按钮内部，并纳入浏览器符号顺序、徽标数量和溢出验收。
- 管理后台各模块列表/详情面板标题已统一展示模块符号、数量或就绪状态徽标，减少进入模块后的二次扫读成本，并纳入浏览器符号顺序和标题溢出验收。
- 管理后台团队和成员列表已改为卡片式信息层级，新增团队/成员摘要芯片，团队归属、邮箱、备注和状态分层展示，编辑/保存/取消动作已统一为符号按钮，并纳入视觉溢出与符号数量验收。
- 管理后台上游来源列表已改为卡片式信息层级，新增类型、前缀、刷新方式和同步状态摘要芯片，编辑、刷新和同步命名动作已统一为符号按钮，并纳入视觉溢出与符号数量验收。
- 管理后台虚拟网关列表已改为卡片式信息层级，新增监听、策略、状态和标签筛选范围摘要芯片，监听信息、策略、状态和标签选择器分层展示，并纳入视觉溢出验收。
- 管理后台策略列表已改为卡片式信息层级，新增作用域、节点上限、状态和虚拟网关范围摘要芯片，标签限制、允许虚拟网关、最大节点数和状态分层展示，并纳入视觉溢出验收。
- 管理后台 Token 卡片已加入状态、归属成员、到期状态和额度状态摘要芯片，并把订阅地址拆成默认、Clash/Mihomo 和 sing-box 的可复制订阅卡，长 URL 省略在卡片内部，避免挤压操作按钮。
- 管理后台 Token 订阅卡已补齐“打开”入口，可直接预览或下载三类订阅内容；打开和复制入口已改为紧凑符号按钮，并纳入浏览器数量、符号和溢出验收。
- 管理后台 Token 订阅卡已补齐通用、Mihomo 和 sing-box 类型徽标，以及 `URI · 通用`、`YAML · Mihomo`、`JSON · sing-box` 格式画像和“当前访问域名/配置域名/相对地址”来源徽标，降低真实分发时选错客户端格式或域名来源的概率，并纳入浏览器数量、格式画像、地址来源和溢出验收。
- 管理后台 Token 卡片已加入额度使用率进度条，帮助在身份模块里直接扫到已用比例，并纳入浏览器数量和溢出验收。
- 管理后台订阅复制动作已加入按钮级“已复制”反馈，并纳入浏览器点击验收。
- 管理后台创建/重置 Token 后的订阅结果框已复用订阅卡布局，避免结果区长 URL 造成横向溢出。
- 管理后台 Token 续期、加额和状态控制已整理为分区动作面板，补齐字段标签、单位、动作符号和无横向滚动验收，减少真实测试时的输入歧义与按钮挤压。
- 管理后台流量模块 Token 用量和上游出口摘要已改为卡片式信息层级，新增 Token 状态/成员/今日/本月与出口来源/上传/下载/总量摘要芯片，额度进度、今日/本月/累计用量和出口上传下载分层展示，并纳入视觉溢出验收。
- 管理后台流量趋势图已补齐小时/日摘要头，展示趋势符号、总量、峰值和活跃样本数，并纳入浏览器趋势摘要与视觉溢出验收。
- 管理后台运维模块已改为带操作符号、状态胶囊的配置操作卡和结构化结果卡，配置检查结果可分层展示结果符号、状态胶囊、Hash、入站、出口、上游和用户数量，并纳入浏览器检查与视觉溢出验收。
- 管理后台空状态已统一为带模块符号、标题和提示文案的结构化空状态，避免真实测试时无数据区域只显示一行“暂无数据”，流量模块空状态已纳入浏览器符号与溢出验收。
- 管理后台节点池地区聚合卡已补齐地区符号、可用数量徽标和协议/来源状态胶囊，地区首屏具备更清晰的控制台扫描层级，并纳入浏览器数量和溢出验收。
- 管理后台节点池地区内节点卡片已补齐协议符号、状态/协议/命名模式信息芯片、服务端端点摘要和统一动作符号按钮，长标题、来源、标签、服务器地址和操作区继续限制在卡片内部，并纳入浏览器数量和溢出验收。
- 管理后台节点池地区展开页已加入地区概况条，先展示地区符号、本区节点数、可用数、协议数和来源数，并纳入浏览器长地区名称和摘要芯片溢出验收。
- 管理后台节点池已加入地区/节点/协议/来源/标签/服务器搜索和一键清空，返回地区与清空筛选控件也统一为符号按钮，便于在大量节点和地区中快速定位目标节点，并纳入浏览器符号验收。
- 管理后台节点详情已加入协议符号摘要头、状态/协议/地区/服务器芯片和 URI 复制动作，并纳入浏览器摘要溢出与点击反馈验收。
- 团队、用户、Token、上游来源、节点导入、虚拟节点基础 API；团队、用户和虚拟节点支持行内编辑。
- 策略列表、创建 API 和行内编辑。
- 策略 `allowed_virtual_nodes` 和 `max_nodes` 按 Token > 成员 > 团队优先级限制可见虚拟节点，并同步作用于 sing-box 入站用户分配。
- 策略 `include_tags` 和 `exclude_tags` 按 Token > 成员 > 团队优先级限制上游出口，并生成带 `auth_user` 的 sing-box route rule。
- Token 续期、追加额度、撤销和恢复操作。
- 管理后台 Token 列表可按行输入自定义续期天数和追加额度 MiB。
- Token 创建响应同时返回默认、Clash/Mihomo 和 sing-box 三类订阅地址，后台创建结果可直接复制分发。
- 来源前缀自动生成和重复前缀编号。
- 来源默认标签会同步到导入的上游节点。
- 节点手动改名保护和恢复自动命名。
- 虚拟节点 `tag_selector` 可按上游标签生成专属 sing-box selector 和 route rule。
- Clash/Mihomo 与 sing-box 基础订阅生成。
- sing-box config.json 生成骨架。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2/Hysteria/TUIC/Juicity/AnyTLS/ShadowTLS/Naive/HTTP/SOCKS/SSH/WireGuard/Tor 上游节点会进入 sing-box outbound，并由默认 selector 承接出口。
- 配置生成侧兼容导入层已支持的 `shadowsocks`、`trojan-go`、`vmess-aead`、`tuic-v5`、`tuic5`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https` 和 `naive-quic` 协议别名。
- Direct URI、`freedom://` URI 和 sing-box JSON `direct` outbound 可进入默认上游 selector 和 outbound stats 列表，配置生成侧兼容手动保存的 `freedom` 协议节点。
- Block URI、`blackhole://` URI、`reject://` URI、Clash `reject`/`reject-drop` 和 sing-box JSON `block` outbound 可进入 sing-box 配置，但不会进入普通上游 selector 或 outbound stats 列表，配置生成侧兼容手动保存的 `blackhole`/`reject`/`reject-drop`/`reject-no-drop`/`reject-tinygif` 协议节点。
- `reject-drop://`、`reject-no-drop://` 和 `reject-tinygif://` 裸 URI 会归一化为标准 `block://`，和 Clash 结构化 reject 变体保持一致。
- DNS URI 和 sing-box JSON `dns` outbound 可进入 sing-box 配置，但不会进入普通上游 selector 或 outbound stats 列表。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan/VMess 的跳过证书校验、禁用 SNI、ALPN 和 uTLS fingerprint TLS 参数并同步进入 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS WebSocket 传输参数并同步为 sing-box outbound transport，兼容 URI 中 `wsPath` 和 `wsHost` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan WebSocket early data 参数并同步为 sing-box outbound transport，兼容 URI 中 `maxEarlyData` 和 `earlyDataHeaderName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS gRPC 传输参数并同步为 sing-box outbound transport，兼容 URI 中 `serviceName` 和 `grpcServiceName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan gRPC keepalive 参数并同步为 sing-box outbound transport，兼容 URI 中 `grpcIdleTimeout`、`grpcPingTimeout`、`permitWithoutStream`、`multiMode` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan QUIC transport 并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTP transport 参数并同步为 sing-box outbound transport，兼容 URI 中 `httpHost`、`httpPath`、`httpMethod`、`httpIdleTimeout` 和 `httpPingTimeout` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTPUpgrade transport 参数并同步为 sing-box outbound transport，兼容 URI 中 `httpUpgradeHost` 和 `httpUpgradePath` 查询别名。
- VLESS、Trojan 和 VMess URI 生成 sing-box outbound 时，WebSocket、HTTP/H2 和 HTTPUpgrade transport Host 兼容 `authority`、`:authority`、`headers.Host`、`headers.:authority`、`headerHost` 和协议专属 Host 查询别名。
- VLESS URI 可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，Trojan URI 可从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息。
- VLESS 和 Trojan URI 可通过 `tls=tls`、`tls=true`、`tlsEnabled`、`enableTLS` 或 `overTLS` 查询参数显式启用普通 TLS，兼容未写 `security=tls` 的订阅写法。
- VLESS 和 Trojan URI 的 SNI 兼容 `sni`、`servername`、`server_name`、`server-name`、`serverName`、`tlsHost`、`tls_host`、`tls-host`、`tlsServerName`、`tls_server_name` 和 `tls-server-name` 查询别名。
- VLESS 和 Trojan URI 的跳过证书校验兼容 `insecure`、`skip-cert-verify`、`skip_cert_verify`、`skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allow-insecure`、`allowInsecure`、`tlsAllowInsecure` 和 `tlsSkipVerify` 查询别名。
- VLESS 和 Trojan URI 的禁用 SNI 兼容 `disable_sni`、`disable-sni`、`disableSNI`、`disableSni`、`tlsDisableSNI`、`tlsDisableSni` 和 `tls-disable-sni` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure` 和 `disableSNI` 等常见查询参数别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure`、`skip_cert_verify` 和 `disableSNI` 等常见查询参数别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan WebSocket/gRPC 传输参数并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VMess gRPC 传输参数并同步为 sing-box outbound transport，兼容 `serviceName`、`grpcServiceName`、`multiMode` 和 `grpcMultiMode` 写法。
- VMess TCP HTTP 伪装和 HTTP/H2 传输参数会同步为 sing-box HTTP transport。
- VMess HTTPUpgrade 传输参数会同步为 sing-box HTTPUpgrade transport，兼容 VMess JSON `net=httpupgrade` 以及 userinfo URI 中 `httpUpgradeHost`/`httpUpgradePath` 查询别名。
- sing-box JSON VMess 节点的 HTTP transport 会保留 host 数组和 path，避免导入后丢失 HTTP 伪装参数。
- VMess 可兼容 `vmess://uuid@host:port?...` userinfo 直连 URI，也可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，并保留 TLS、WebSocket、gRPC、SNI、ALPN、跳过证书校验、uTLS fingerprint 和 `packet_encoding`/`packet-encoding`/`packetEncoding` 参数同步为 sing-box outbound；userinfo URI 兼容 `serverName`、`tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`tlsDisableSNI`、`tlsEnabled`、`enableTLS`、`overTLS`、`grpcServiceName` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria 窗口、MTU 发现和 `tls.utls.fingerprint`/`fp` 参数并同步为 sing-box outbound，兼容 `authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMTUDiscovery` 和 `clientFingerprint` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria2 上下行带宽、TLS 标志、证书 pin、URI 客户端指纹和 sing-box `tls.utls.fingerprint` 参数并同步为 sing-box outbound，兼容 `authStr`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`obfsPassword`、`pinSHA256`、`pin_sha256`、`certificatePublicKeySHA256`、`certificate_public_key_sha256` 和 `clientFingerprint` 查询别名。
- Hysteria 和 Hysteria2 URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Hysteria/Hysteria2 结构化节点会保留 `token`/`pass`/`passwd` credential 别名、`obfsPassword`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMtuDiscovery`、`disableMTUDiscovery`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 等 camelCase 参数，并归一化到 Hysteria/Hysteria2 URI 到 sing-box outbound 生成链路。
- TUIC URI 可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，并兼容 `congestionControl`、`udpOverStream`、`udpRelayMode`、`zeroRttHandshake`、`heartbeatInterval`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 等常见查询参数别名。
- TUIC URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML TUIC 结构化节点会保留 credential 别名、`udp-over-stream`/`udpOverStream`、`zero-rtt-handshake`/`zeroRttHandshake`、`heartbeat-interval`/`heartbeatInterval`、`disable-sni`/`disable_sni` 和 `client-fingerprint`/`clientFingerprint` 等参数，并归一化到 TUIC URI 到 sing-box outbound 生成链路。
- sing-box JSON 和 URI 导入链路会保留 TUIC 跳过证书校验、禁用 SNI、ALPN 和 `tls.utls.fingerprint`/`fp` 客户端指纹并同步为 sing-box outbound。
- Clash YAML、Quantumult X、Surge、sing-box JSON 和 URI 导入链路会保留 Juicity UUID、密码、`congestion_control`、SNI、跳过证书校验、禁用 SNI、ALPN 和 `tls.utls.fingerprint`/`fp` 客户端指纹并同步为 sing-box outbound；URI 生成侧兼容 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询别名。
- Juicity URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Juicity 结构化节点会保留 `uuid`/`id`/`user-id` credential 别名、`congestionControl`/`congestion-controller`、`skip-cert-verify`、`disable-sni`/`disable_sni`、ALPN 和 `client-fingerprint`/`clientFingerprint` 客户端指纹，并归一化到 Juicity URI 到 sing-box outbound 生成链路。
- AnyTLS 和 ShadowTLS URI 可从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息。
- sing-box JSON 和 URI 导入链路会保留 AnyTLS 会话空闲参数和 `tls.utls.fingerprint`/`fp` 客户端指纹并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`idleSessionCheckInterval`、`idleSessionTimeout`、`minIdleSession` 和 `clientFingerprint` 查询别名。
- sing-box JSON 和 URI 导入链路会保留 ShadowTLS `tls.utls.fingerprint`/`fp` 客户端指纹并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 查询别名。
- AnyTLS 和 ShadowTLS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML AnyTLS/ShadowTLS 结构化节点会保留 `password`/`pass`/`passwd`/`psk`/`token` credential 别名；AnyTLS 会同步 `idleSessionCheckInterval`、`idleSessionTimeout` 和 `minIdleSession` 等会话空闲参数；两者都会通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- sing-box JSON 和 URI 导入链路会保留 Naive/Naive+QUIC 跳过证书校验、禁用 SNI、ALPN、QUIC 控制项和 `tls.utls.fingerprint`/`fp` 参数并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`insecureConcurrency`、`udpOverTcp`、`quicCongestionControl` 和 `clientFingerprint` 查询别名。
- Clash YAML Naive 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `insecureConcurrency`、`udpOverTcp` 和 `quicCongestionControl` 等 camelCase QUIC 控制参数，并通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 HTTP/HTTPS 代理跳过证书校验、禁用 SNI、ALPN 和 `tls.utls.fingerprint`/`fp` 参数并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`tlsEnabled`、`enableTLS` 和 `overTLS` 查询别名。
- Naive 和 HTTP/HTTPS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- SOCKS URI 会按 `socks4://`、`socks4a://`、`socks5://` 或 `version=4/4a/5` 查询参数保留版本，并同步为 sing-box SOCKS outbound 的 `version` 字段。
- SOCKS URI 支持 `udp`、`udpEnabled`、`udp_relay` 和 `udp-relay` 标志并归一为 sing-box outbound 的 `network=udp`，并兼容 `udpOverTcp` 查询别名。
- `socks5h://` 裸 URI 会归一化为标准 `socks5://`，兼容 curl/requests 生态里常见的 SOCKS5H 订阅写法；配置生成侧也兼容手动保存的 `socks5h` 协议节点。
- Clash YAML SOCKS 节点可保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化订阅导入后会继续进入 SOCKS UDP 转换链路。
- Surge SOCKS 节点可保留 `udp=true` 标志，结构化订阅导入后会继续进入 SOCKS UDP 转换链路。
- Quantumult X SOCKS 节点可保留 `udp=true` 标志，结构化订阅导入后会继续进入 SOCKS UDP 转换链路。
- sing-box JSON SOCKS outbound 可保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化订阅导入后会继续进入 SOCKS UDP 转换链路。
- V2Ray/Xray JSON SOCKS outbound 可保留 `udp`、`udpEnabled`、`udp_enabled`、`udp_relay`、`udp-relay`、`udp_over_tcp`、`udpOverTcp`、`udp-over-tcp` 和 `uot` 标志，结构化订阅导入后会继续进入 SOCKS UDP 转换链路。
- Shadowsocks URI 可从 `method`/`cipher`/`encrypt-method`/`encrypt_method`/`encryptMethod`/`encryption`/`security` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`account_password`/`account-password`/`accountPassword` 查询参数读取认证信息，并兼容 `pluginOptions` 查询别名。
- Naive、HTTP/HTTPS 和 SOCKS URI 可从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息。
- SSH URI 可从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息，并兼容 `identity_file`/`key_path`、`key`、`privateKey`、`privateKeyPath`、`privateKeyPassphrase`、`passphrase`、`hostKey`、`hostKeyAlgorithms`、`clientVersion`、`cipher`/`ciphers`、`mac`/`macAlgorithms` 和 `kexAlgorithm`/`kexAlgorithms` 等常见查询参数别名。
- Clash YAML SSH 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `privateKeyPath`、`privateKeyPassphrase`、`hostKey`、`hostKeyAlgorithms`、`clientVersion` 和 `kexAlgorithm` 等 camelCase 参数，并归一化到 SSH URI 到 sing-box outbound 生成链路。
- WireGuard URI 可兼容 `privateKey`/`publicKey`/`preSharedKey`、`ip`/`ipv6`/`localAddresses`、`allowedIPs`/`allowedIps`/`peerAllowedIps`、`systemInterface`、`interfaceName`、`reservedBytes` 和 `peerReserved` 等常见查询参数别名。
- Clash YAML WireGuard 结构化节点会保留 `privateKey`、`peerPublicKey`、`localAddress`/`localAddresses`、`preSharedKey`、`allowedIPs`/`peerAllowedIps`、`reservedBytes`/`peerReserved`、`systemInterface` 和 `interfaceName` 等 camelCase 参数，并归一化到 WireGuard URI 到 sing-box outbound 生成链路。
- Tor URI 可兼容 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/重复 `arg` 和 `torrc[Option]` 等常见查询参数别名。
- Clash YAML Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和嵌套 `torrc` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Surge 和 Quantumult X Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和 `torrc.*`/`torrc[Option]` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Hysteria2/Hy2 URI 可从查询参数读取 `password`、`auth`、`auth_str`、`authStr`、`pass`、`passwd`、`pwd`、`psk`、`token`、`secret`、`credential` 或 `accountPassword` 认证密码。
- Hysteria v1 URI 可从查询参数读取 `token`、`auth`、`auth_str`、`authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`psk`、`secret`、`credential` 或 `accountPassword` 认证字符串。
- sing-box 服务端配置生成会过滤不可用 Token，避免已撤销、已过期或已超额用户继续进入网关配置。
- 本地 smoke、API flow、页面截图验收脚本。
- 远程部署探测、磁盘清理、诊断采集脚本骨架。
- 自定义 sing-box 镜像运行阶段使用 `scratch`，减少远程构建对额外镜像仓库的依赖。

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
- JSON 包装订阅会忽略 `meta`、`pagination`、`errors`、`message`、`traceId` 等常见接口元信息字段，避免把接口文档或分页 URL 误导入为 HTTP 节点。
- JSON 包装订阅会忽略 Clash 风格 `proxy-providers`、`rule-providers`、`proxy-groups`、`health-check` 和 `rules` 中的下载 URL、健康检查 URL 和规则源 URL，避免误导入为 HTTP 节点；其中 `proxy-providers` 内嵌的 `proxy`/`proxies`、`node`/`nodes`、`server`/`servers`、`proxy-list`/`proxies-list`、`nodeList`/`nodesList`、`serverList`/`serversList` 节点清单会继续按结构化节点解析。
- JSON 包装订阅可解析 `list`、`records`、`rows`、`entries`、`node`/`nodes`、`proxy`/`proxies`、`server`/`servers`、`nodeList`/`nodesList`、`proxyList`/`proxiesList`、`serverList`/`serversList`、`subscriptionList`/`subscriptionsList`、`urlList`/`urlsList`、`linkList`/`linksList` 及其 snake_case/kebab-case 变体等常见集合字段，容器名不会污染缺少 fragment 的节点 URI 展示名。
- JSON 包装订阅字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Data.NodeList`、`Proxy_List` 和 `Payload.LinkList`。
- JSON 包装订阅节点名称字段也兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Name`、`Display_Name` 和 `NodeName`，缺少 fragment 的链接会继续保留真实展示名。
- JSON 包装订阅可解析 `shareUrl`、`shareLink`、`subscriptionUrl`、`nodeUrl` 及其 snake_case/kebab-case 变体；对象内带名称的 `subscribeUrl`、`subUrl` 和 `downloadUrl` 分享字段也会按节点链接解析，分享链接缺少 fragment 时会优先使用对象内名称。
- JSON 包装订阅可解析 `nodes`/`proxies` 中的 Clash 风格结构化节点对象。
- JSON 结构化节点可识别 `protocol`/`proto`/`scheme`/`protocolName`/`nodeType`/`serverType`/`protocolType`/`proxyType`/`proxyProtocol`、`host`/`hostname`/`address`/`serverAddress`/`serverHost`/`remoteHost`/`nodeHost`/`endpoint`/`add`、`server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber`、`method`/`encryptMethod`/`encrypt-method`、`pass`/`passwd`/`pwd`/`secret`/`credential`/`accountPassword`、`username`/`userName`/`user_name`/`accountName`/`account`、`id`/`user-id` 和 `displayName`/`nodeName`/`label`/`title`/`remarks`/`remark` 等常见别名字段；`server`/`host`/`address`/`endpoint` 写成 `host:port` 时会自动拆分端口。
- JSON 结构化节点可归一化 `skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allowInsecure`、`allow_insecure`、`allow-insecure`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`clientFingerprint`、`flowName`、`xtlsFlow`、`networkType`、`transportType`、`tlsEnabled`、`enableTLS`、`overTLS`、`tlsHost`、`tlsServerName`、`tlsSettings`、`tlsOptions`、`pluginOpts`、`wsOpts`、`wsSettings`、`wsHeaders`、`grpcSettings`、`httpSettings`、`h2Settings`、`httpUpgradeSettings`、`quicSettings`、`tcpSettings`、`grpcServiceName`、`privateKeyPath`、`publicKey`、`shortId`、`shortIds`、`spiderX`、`serverNames`、`server_names`、`server-name` 等常见 camelCase/snake_case/kebab-case 参数，并支持 `tls.enabled`、`tls.enable`、`tls.serverName`、`tls.serverNames`、`tls.server_names`、`tls.server-name`、`tls.tls_server_name`、`tls.utls.fingerprint`、`reality.publicKey` 和 `reality.serverNames` 等嵌套字段，保留 TLS、REALITY、插件和传输配置。
- JSON 结构化节点可解析通用 `transport` 容器，保留 `transport.type`、`transport.path`、`transport.host`、`transport.headers.Host`、`transport.authority`、`transport.headers.authority`、`transport.headers.:authority`、`transport.serviceName`、`transport.idleTimeout` 等 WebSocket、HTTP、HTTPUpgrade 或 gRPC 参数。
- subscription 来源可解析 Quantumult X `[server_local]`/`[server_remote]` 常见 SS、SSR、Hysteria2/Hy2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、Trojan、VLESS、VMess、HTTP 和 SOCKS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名。
- Quantumult X `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Quantumult X、Surge、Clash YAML 和 sing-box JSON 的 `tuic-v5`/`tuic5` 类型会按 TUIC 兼容链路归一化为 `tuic://`，保留 UUID、密码、拥塞控制、UDP relay、SNI 和节点名。
- Quantumult X 和 Surge 结构化 VLESS/Trojan 节点可保留 gRPC transport 和 service name。
- Quantumult X HTTP/SOCKS 节点可兼容键值和位置参数两种认证写法。
- subscription 来源可解析以节点名称为 key 的 JSON URI 对象映射。
- subscription 来源可解析裸 VMess JSON 单对象、数组、包装字符串和按名称映射对象，并转换为标准 `vmess://` URI；裸 VMess JSON 可识别 `address`/`server`/`serverHost`/`nodeHost`/`endpoint`、`serverPort`/`nodePort`/`portNumber`、`uuid`/`user_id` 等字段别名，并保留 `packetEncoding`/`packet_encoding`/`packet-encoding`、`disable_sni` 和 `fp`/`fingerprint`/`clientFingerprint` 参数。
- 裸 VMess JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Display_Name`、`Server-Port`、`User_ID`、`Packet_Encoding`、`Disable-SNI` 和 `Client-Fingerprint`。
- subscription 来源可解析 SSD/ShadowsocksD `ssd://` 订阅和裸 SSD JSON，并展开为标准 `ss://` URI；SSD `servers` 支持数组和按名称分组的对象映射，并兼容 SIP008 同款 Shadowsocks 字段别名。
- SIP008 和 SSD JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Version`、`Servers`、`Display_Name`、`Server-Port`、`Encrypt_Method`、`PluginOpts` 和 `PluginOptions`。
- subscription 来源可解析 Surge `[Proxy]` 代理段中的 SS、SSR、Trojan、VLESS、VMess、Hysteria2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、HTTP/HTTPS、SOCKS、Direct、Reject 和 DNS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名，转换为标准 URI。
- Surge `[Proxy]` `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- subscription 来源支持按刷新间隔自动同步到期订阅。
- subscription 来源可解析 Clash YAML `proxies` 中的常见 SS/Trojan/VLESS/VMess/Hysteria2/TUIC/Juicity/Hysteria/HTTP/SOCKS/AnyTLS/ShadowTLS/Naive/SSH/WireGuard 节点，以及 Direct/Reject 变体/DNS 内置出站；Clash YAML 结构化 `type` 兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 别名。
- Clash YAML `type: ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Clash YAML 解析支持嵌套 `proxy`/`proxies`/`proxy-list`/`proxies-list`/`node`/`nodes`/`node-list`/`server`/`servers`/`server-list` 列表，并兼容 `proxyList`、`nodeList`、`serverList` 等 camelCase/PascalCase 写法，可兼容带内嵌 provider 节点清单的订阅结构。
- Clash YAML 解析支持 `proxy: [{ ... }]`、`proxies: [{ ... }]`、`proxy-list: [{ ... }]`、`proxies-list: [{ ... }]`、`nodes: [{ ... }]`、`servers: [{ ... }]` 和 `node-list: [{ ... }]` 等内联节点数组，并兼容 camelCase/PascalCase 容器名，可兼容 provider 或顶层节点的紧凑写法。
- Clash YAML 解析支持 `proxies: &anchor` 这类带 YAML anchor 的块状节点列表。
- Clash YAML 解析支持 `- &anchor { ... }` 和 `proxies: [&anchor { ... }]` 这类节点条目级 anchor 的内联写法。
- Clash YAML 解析支持常见块状和内联 `ws-opts`、`grpc-opts` 嵌套写法，能保留 WebSocket path/host 和 gRPC service name。
- Clash YAML 解析可归一化 `skipCertVerify`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`wsPath`、`wsHost`、`wsOpts`、`maxEarlyData`、`earlyDataHeaderName`、`grpcServiceName` 等常见 camelCase 参数，块状和内联节点都会保留 TLS 与传输配置。
- Clash YAML 解析支持行内和块状数组标量，可保留 ALPN、WireGuard 地址、允许 IP 和 reserved 字节列表。
- subscription 来源可解析 SIP008 Shadowsocks 订阅，`servers` 支持数组和按名称分组的对象映射；SIP008 Shadowsocks 节点可识别 `address`/`serverAddress`/`nodeHost`/`endpoint`、`portNumber`/`nodePort`、`encryptMethod`/`security` 和 `pass`/`passwd` 等字段别名。
- SIP008、Clash YAML 和 sing-box JSON 的 Shadowsocks 节点会保留 SIP003 插件参数并同步进入 sing-box outbound。
- `shadowsocks://` URI 会归一化为标准 `ss://` URI，避免长 scheme 导入后无法进入 Shadowsocks outbound 转换链路。
- `ssr://` ShadowsocksR URI 会解码为标准 `ss://` URI，保留服务端、端口、加密方法、密码和 remarks 节点名；SSR 专有 protocol/obfs 参数暂按 Shadowsocks 兼容链路导入。
- `http+tls://` 和 `http-tls://` 裸 URI 会归一化为标准 `https://`，复用现有 HTTP outbound TLS 配置生成链路。
- `any-tls://`、`shadow-tls://`、`naive-https://` 和 `naive-quic://` 裸 URI 会归一化为标准 `anytls://`、`shadowtls://`、`naive+https://` 和 `naive+quic://`，和结构化订阅解析保持一致。
- `hy2://` 和 `wg://` 裸 URI 会归一化为标准 `hysteria2://` 和 `wireguard://`，让导入后的协议字段和配置生成链路保持一致。
- `trojan-go://` 裸 URI 会归一化为标准 `trojan://`，复用现有 Trojan TLS、REALITY 和传输参数转换链路。
- `vmess-aead://` 裸 URI 会归一化为标准 `vmess://`，复用现有 VMess TLS、WebSocket、gRPC、HTTP 和 packet encoding 转换链路。
- `tuic-v5://` 和 `tuic5://` 裸 URI 会归一化为标准 `tuic://`，复用现有 TUIC outbound 转换链路；手动保存 `tuic-v5`/`tuic5` 协议节点时配置生成侧也会按 TUIC 输出。
- sing-box JSON `outbounds` 可解析 Juicity 节点并兼容 `trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 结构化 `type` 别名。
- subscription 来源可解析 sing-box JSON `outbounds` 中的 Naive/Naive+QUIC 节点，并保留 QUIC、UDP over TCP、并发、TLS 和 uTLS fingerprint 参数。
- subscription 来源可解析 sing-box JSON 顶层 `endpoints` 中的 WireGuard endpoint 模型，并转换为当前 Phase 1 WireGuard outbound 兼容 URI。
- sing-box JSON 顶层可兼容单数 `outbound`/`endpoint` 和 `outboundList`/`outboundsList`、`endpointList`/`endpointsList` 写法；也兼容顶层直接给出 outbound/endpoint 数组、单个对象或按名称映射对象；单对象、数组和对象映射都会进入同一归一化链路。
- sing-box JSON 顶层、outbound、endpoint、TLS、uTLS、REALITY、transport、headers 和 WireGuard peer 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Server-Port`、`ServerPort`、`ServerName`、`Disable-SNI`、`ServiceName`、`PermitWithoutStream`、`PrivateKey`、`AllowedIPs`。
- sing-box JSON `outbounds` 和 `endpoints` 支持数组、单个对象、按名称分组的对象映射和分组数组映射。
- sing-box JSON transport 的 `host`、`authority`、`headers.Host`、`headers.authority` 和 `headers.:authority` 支持字符串或数组写法，WebSocket、HTTP、HTTPUpgrade 和 VMess WebSocket/HTTP/HTTPUpgrade transport 会保留为逗号分隔 Host。
- subscription 来源可解析 V2Ray/Xray JSON `outbounds` 中的 VMess、VLESS、Trojan、Shadowsocks、HTTP、SOCKS、freedom、blackhole、reject 变体和 DNS 出站，并兼容 `trojan-go`、`vmess-aead`、`http+tls`、`http-tls`、`socks4`、`socks4a`、`socks5`、`socks5h` 协议别名，保留常见 TLS、REALITY、WebSocket、gRPC、QUIC、HTTP/H2 和 HTTPUpgrade 传输参数；REALITY 兼容 `serverName`/`serverNames`、`shortId`/`shortIds` 和 hyphen/snake-case 别名。
- V2Ray/Xray JSON `streamSettings`、TLS/REALITY 设置和各 transport 设置块支持 camelCase、snake_case 和 hyphen-case 容器字段别名。
- V2Ray/Xray JSON 顶层、outbound、settings、vnext/server、user/account、streamSettings、TLS、uTLS、REALITY、WebSocket、gRPC、HTTP、HTTPUpgrade 和 TCP header 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Settings`、`VNext`、`Server-Port`、`StreamSettings`、`TLSSettings`、`ServerName`、`Disable-SNI`、`WSSettings`、`MaxEarlyData`、`GrpcSettings` 和 `PermitWithoutStream`。
- V2Ray/Xray JSON Shadowsocks outbound 可识别 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`security`/`encryption`/`encryptMethod` 字段别名，并保留 SIP003 `plugin`、`plugin_opts`、`plugin_options`、`network` 和 `protocol` 标志，结构化订阅导入后会继续进入 Shadowsocks 转换链路。
- V2Ray/Xray JSON 普通 TLS 会兼容 `serverName`/`serverNames` 以及 hyphen/snake-case SNI 别名，数组写法会选取首个有效值同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会兼容 `allowInsecure`、`skip-cert-verify`、`insecure` 和 `disable_sni` 写法，并同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会保留 `fingerprint`、`clientFingerprint` 或 `utls.fingerprint` 客户端指纹，并同步为 sing-box uTLS fingerprint。
- V2Ray/Xray JSON 兼容旧配置中的单个 `outbound`、`outboundDetour`、复数 `outboundDetours` 和 `outboundList`/`outboundsList` 出站入口，会和现代 `outbounds` 一起归一化导入。
- V2Ray/Xray JSON 也兼容顶层直接给出 outbound 数组、单个 outbound 对象或按名称映射的 outbound 对象，映射 key 写成 `host:port` 时会同步补齐节点服务端、端口和默认节点名。
- V2Ray/Xray JSON `wsSettings.host`、`wsSettings.headers.Host` 和 `authority` 支持字符串或数组写法，数组会保留为逗号分隔的 WebSocket Host。
- V2Ray/Xray JSON `wsSettings` 支持 `maxEarlyData`、`max_early_data`、`max-early-data`、`earlyDataHeaderName`、`early_data_header_name` 和 `early-data-header-name` WebSocket early data 别名。
- V2Ray/Xray JSON `tcpSettings.header.type=http` 的 HTTP 伪装 Host、path 和 method 会保留为 HTTP transport 参数。
- V2Ray/Xray JSON `httpSettings`/`h2Settings` 可从 `host`、`headers.Host` 和 `authority` 读取 HTTP/H2 transport Host，并支持数组写法的 path。
- V2Ray/Xray JSON `httpupgradeSettings.host` 支持字符串或数组写法，也可从 `headers.Host` 读取 Host 列表。
- V2Ray/Xray JSON `grpcSettings` 的 `idle_timeout`、`health_check_timeout` 和 `permit_without_stream` 会保留为 gRPC transport keepalive 参数。
- V2Ray/Xray JSON `grpcSettings.multiMode` 会保留为 sing-box gRPC transport 的 `multi_mode` 参数。
- V2Ray/Xray JSON `grpcSettings` 支持 `service-name`、`idle-timeout`、`health-check-timeout`、`permit-without-stream` 和 `multi-mode` 这类 hyphen 别名。
- V2Ray/Xray JSON 存在 `quicSettings`/`quic_settings`/`quic-settings` 时会保留为 QUIC transport。
- V2Ray/Xray JSON VMess/VLESS `vnext` endpoint 可识别 `address`、`server`、`host`、`add`、`serverAddress`、`serverHost`、`remoteHost`、`nodeHost`、`endpoint` 和 `server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber` 别名；VMess/VLESS user 可识别 `id`、`uuid`、`userId`、`user_id` 和 `user-id` 认证字段，VMess user 可识别 `alter-id`/`aid` 和 `cipher` 别名，VMess/VLESS user 可识别 `flow`/`flowName`/`xtlsFlow` 等 flow 别名，并可识别 `packetEncoding`/`packet_encoding`/`packet-encoding` 同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON VMess/VLESS `vnext.users` 支持以 UUID 为 key 的对象映射，缺少 `id` 时会使用映射 key 作为用户 ID，并保留映射值中的名称、加密、flow 和 packet encoding 等字段。
- V2Ray/Xray JSON `vnext` 和 `servers` 支持以服务端域名或 IP 为 key 的对象映射，缺少服务端字段时会使用映射 key 作为节点服务端，映射 key 写成 `host:port` 或 `[ipv6]:port` 时会同步拆出端口；VMess/VLESS 的 `vnext` 映射值也支持直接写用户对象、用户数组、UUID 标量、UUID 数组或账号名到 UUID/用户对象的映射，Trojan/Shadowsocks 的 `servers` 映射值支持直接写密码字符串或密码数组。
- V2Ray/Xray JSON VMess、VLESS、Trojan 和 Shadowsocks 支持缺省 `settings.vnext`/`settings.servers` 的扁平 outbound 写法；服务端、端口、认证和加密字段可直接写在 outbound 顶层或 `settings` 顶层。
- Clash YAML 和 sing-box JSON VMess/VLESS 节点可识别 `packetEncoding`、`packet_encoding` 和 `packet-encoding` 写法，并同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON Trojan/HTTP/SOCKS outbound 可兼容 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`userName`/`accountName`、`users`、`accounts` 数组、`accounts` 用户名到密码或对象映射和 server 层认证写法；HTTP/SOCKS 的 `servers` 端点映射值也支持直接写账号名到密码或账号对象的映射，并兼容 `user:password` 标量或标量数组。
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
- V2Ray stats 计数名称解析可容忍分段空白和 `traffic`/方向字段大小写差异。
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
- 真实实例只读就绪检查。
- 部署后真实可用性收口脚本和订阅正文自动探测；未显式提供订阅 URL 时，可从后台 Token 列表选取可恢复订阅地址验证默认、Clash/Mihomo 和 sing-box 输出。
- 面向实际使用者的手把手使用说明。
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
- 可通过只读脚本判断实例是否具备真实体验数据。
- 可通过部署后收口脚本补齐默认虚拟节点、默认策略、发布配置并重启数据面。
- 手把手使用说明覆盖从部署初始化到日常管理、订阅分发、流量审计和故障排查的完整流程。
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
