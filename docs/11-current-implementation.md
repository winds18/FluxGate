# 当前实现状态

## 1. 已落地内容

本轮完成 Phase 1 foundation，目标是建立可运行、可测试、可继续扩展的控制面骨架。

已实现：

- Go HTTP API 服务。
- SQLite migration 自动执行。
- 嵌入式 Web 管理后台登录页、仪表盘和基础写操作表单。
- 管理后台已参考 zashboard 的模块化控制台信息架构重构为桌面侧边导航、移动底部 Dock、概览/接入/节点/身份/策略/流量/运维模块视图；桌面侧边导航和移动 Dock 已补齐统一模块符号，写操作表单收敛到对应模块内，避免单页堆叠导致扫描困难。
- 管理后台导航符号已纳入浏览器 QA，校验桌面 7 个侧边模块和移动 7 个 Dock 模块都有符号且移动端无横向溢出。
- 管理后台工作区标题已补齐当前模块符号，和侧边导航、移动 Dock 使用同一套视觉锚点；浏览器 QA 会逐模块校验标题符号跟随切换更新。
- 管理后台桌面侧边导航和移动 Dock 已补齐模块数量徽标，来源、节点、Token、策略、流量和配置就绪状态可在导航层直接扫描；浏览器 QA 会校验桌面/移动徽标数量、填充值和移动端无溢出。
- 管理后台各模块创建/导入表单已默认收纳为可展开抽屉，抽屉标题、展开/收起控件和提交控件都补齐统一功能符号，列表、卡片和详情成为模块首屏重点，减少常驻表单压缩内容区。
- 管理后台概览页顶部指标卡已补齐模块符号，团队、用户、Token、来源、节点、虚拟节点和策略指标使用统一符号、标签和数字层级；浏览器 QA 会校验 7 个指标符号和值都存在且无视觉溢出。
- 管理后台基础 HTML 转义工具已兼容数值 ID 等非字符串字段，避免真实数据加载到流量、Token 或节点卡片时触发前端渲染异常；浏览器 QA 会在登录后等待完整数据加载，并在加载异常时输出失败接口、页面状态和控制台信息。
- 管理后台概览页已提供真实测试闭环看板，按来源、节点池、虚拟网关、团队 Token 和访问策略展示准备状态，并给出下一步模块入口；每个闭环步骤都有序号、符号化状态单元和就绪/待补徽标，浏览器 QA 会校验序号、符号、状态徽标和无视觉溢出。
- 管理后台概览页快速入口卡已补齐模块符号、实时数量和就绪徽标，接入来源、节点地区、Token 和发布状态能在概览首屏直接扫描；浏览器 QA 会校验 4 个入口符号顺序、徽标填充值和视觉溢出。
- 管理后台概览页下一步入口已改为结构化动作卡，展示目标模块符号、补齐/可测试状态说明和主操作按钮；浏览器 QA 会校验动作卡存在且无视觉溢出。
- 管理后台模块标题区已加入随当前视图变化的数据摘要条，进入接入、节点、身份、策略、流量和运维模块时能先看到来源数、地区数、有效 Token、策略限制和流量样本等关键上下文；摘要胶囊会约束标签和值在自身内部省略。
- 管理后台模块标题区已加入当前模块快速定位条，可在接入、节点、身份、策略、流量和运维视图内直接跳转到列表、详情或创建抽屉；点击创建入口会自动展开对应表单，减少长页面反复滚动。
- 管理后台模块快速定位条已补齐功能符号、数量和新增徽标，列表入口直接展示当前规模，创建入口统一展示 `+`，并限制符号、文字和徽标始终收纳在按钮内；浏览器 QA 会校验每个定位按钮都有符号与徽标且无横向溢出。
- 管理后台各模块列表/详情面板标题已统一为模块符号、标题和数量徽标，来源、节点池、虚拟节点、团队、成员、Token、策略、流量和配置发布区都能在面板头部直接扫到所属模块和当前规模或就绪状态；浏览器 QA 会校验 9 个面板符号顺序和标题无溢出。
- 管理后台团队和成员列表已改为卡片列表，团队卡增加类型、状态、成员数和备注状态摘要芯片，成员卡增加所属团队、状态、邮箱和备注状态摘要芯片；名称、团队归属、邮箱、备注和状态继续分层展示，编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证身份卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台上游来源已改为卡片列表，来源名称下方增加类型、前缀、刷新方式和同步状态摘要芯片，URL、默认标签、上次同步和异常状态继续分层展示；编辑、刷新和同步命名动作已统一为符号按钮，浏览器 QA 会验证来源卡片数量、摘要芯片、动作符号、编辑入口和卡片无视觉溢出。
- 管理后台虚拟网关列表已改为卡片列表，名称下方增加监听、策略、状态和标签筛选范围摘要芯片，监听协议/端口、策略、状态和标签选择器继续分层展示；编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证虚拟网关卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台策略列表已改为卡片列表，策略名称下方增加作用域、节点上限、状态和虚拟网关范围摘要芯片，标签限制、允许虚拟网关、最大节点数和状态继续分层展示；编辑/保存/取消动作已统一为符号按钮，浏览器 QA 会验证策略卡片数量、摘要芯片、动作符号、编辑入口和无视觉溢出。
- 管理后台 Token 卡片已在标题下增加状态、归属成员、到期状态和额度状态摘要芯片，订阅可用性可以先扫读；同时会把默认、Clash/Mihomo 和 sing-box 订阅地址拆成可复制订阅卡，长 URL 只在卡内省略显示，并用 `URI · 通用`、`YAML · Mihomo`、`JSON · sing-box` 标明客户端格式画像，同时显示“当前访问域名/配置域名/相对地址”的地址来源；浏览器 QA 会校验摘要芯片、订阅卡数量、格式画像、地址来源和无视觉溢出。
- 管理后台 Token 订阅卡除复制外已提供“打开”入口，可直接在新标签页预览或下载默认、Clash/Mihomo 和 sing-box 订阅内容；打开和复制入口已改为紧凑符号按钮，浏览器 QA 会校验入口数量、操作符号、格式画像、地址来源和无视觉溢出。
- 管理后台 Token 订阅卡已补齐通用、Mihomo 和 sing-box 类型徽标，浏览器 QA 会校验徽标数量和无视觉溢出。
- 管理后台 Token 卡片已加入额度使用率进度条，和流量模块使用同一套状态色；浏览器 QA 会校验每张 Token 卡都有额度条且无视觉溢出。
- 管理后台复制订阅地址后会在对应按钮上显示“已复制”并同步顶部状态，浏览器 QA 会实际点击复制入口确认反馈可见。
- 管理后台创建或重置 Token 后的订阅结果框已复用同一套订阅卡样式，避免结果区长 URL 回到旧式两列表格布局；浏览器 QA 会构造结果预览并验证无视觉溢出。
- 管理后台 Token 续期、加额和状态控制已整理为分区动作面板，续期/加额保留可见字段标签与单位，恢复、重置订阅和撤销收纳为同一状态控制区，避免底部按钮条横向挤压；浏览器 QA 会校验每张 Token 卡都有两组带标签的操作字段、状态控制区、动作符号和无横向滚动溢出。
- 管理后台运维模块已把配置检查、发布、回滚和重启拆为带符号和状态胶囊的操作卡；配置结果使用带结果符号、状态胶囊和结构化字段的结果卡展示 Hash、入站、出口、上游和用户等关键字段，浏览器 QA 会点击检查配置并验证操作卡、按钮符号、结果卡和无视觉溢出。
- 管理后台空状态已统一为带模块符号、标题和提示文案的结构化空状态，来源、节点、身份、策略、Token、流量图和出口摘要在无数据时不会只剩一行“暂无数据”；浏览器 QA 会验证流量模块空状态符号和无视觉溢出。
- 管理后台已做基础视觉统一：统一顶部栏、卡片、表单、按钮、表格、状态徽标和响应式栅格，输入控件高度保持一致；顶部退出和刷新动作都使用统一符号按钮，节点卡片长标题、来源名和标签会限制在父卡片内换行/截断，并通过浏览器 QA 检测桌面和移动视图是否发生视觉溢出。
- 节点池地区聚合卡已补齐地区符号、可用数量徽标和协议/来源状态小胶囊，地区首屏更接近控制台扫描节奏；浏览器 QA 会校验每个地区卡都有符号、徽标、状态胶囊且无视觉溢出。
- 节点池地区展开后的节点卡片已改为协议符号、标题/来源、状态/协议/命名模式芯片和服务端端点摘要分层展示，编辑/保存/恢复自动操作使用统一符号按钮；长节点名、来源、标签、服务端地址和动作按钮会在卡片内收束，浏览器 QA 会校验每张节点卡都有协议符号、三枚信息芯片、端点摘要、编辑保存符号和两枚动作符号且无视觉溢出。
- 节点池地区展开页已加入地区概况条，先展示地区符号、本区节点数、可用数、协议数和来源数；浏览器 QA 会用长地区名称验证概况条摘要芯片和标题不溢出。
- 节点池支持在地区聚合页和地区节点卡片页按地区、节点名、协议、来源、标签或服务器搜索，并可一键清空筛选；浏览器 QA 会验证搜索可见、筛选有效、清空可恢复且不产生横向溢出。
- 节点详情面板已加入摘要头，先展示协议符号、节点名、来源、状态、协议、地区和服务器地址，再展开字段网格；URI 复制动作保留，复制后按钮显示“已复制”并同步顶部状态；浏览器 QA 会实际展开节点详情，校验摘要芯片无溢出并点击复制入口。
- 管理员引导账号。
- PBKDF2-SHA256 密码哈希。
- 管理员签名 Cookie 会话。
- 管理 API 登录保护。
- 会话 Cookie 按实际请求协议设置 Secure，支持 HTTPS 反代和 HTTP 局域网调试。
- 团队和用户基础页面支持创建、列表和行内编辑名称、归属、备注及启停状态；Token 支持创建、列表、续期、追加额度、撤销和恢复。
- 策略基础页面支持创建、列表和行内编辑名称、范围、标签限制、虚拟节点限制、最大节点数和启停状态。
- 策略 `allowed_virtual_nodes` 和 `max_nodes` 已按 Token > 成员 > 团队优先级生效，用于限制订阅输出中的可见虚拟节点，并同步约束 sing-box 入站里的用户分配；被策略挡住且没有可用用户的虚拟节点不会生成入站。
- 策略 `include_tags` 和 `exclude_tags` 已按 Token > 成员 > 团队优先级生效，会为匹配 Token 生成带 `auth_user` 的 sing-box route rule，限制该 Token 在虚拟节点下可走的上游出口。
- Token 支持续期、追加额度、撤销和恢复，并同步 gateway account 状态。
- 管理后台 Token 列表支持按行输入自定义续期天数和追加额度 MiB，避免只能使用固定续期/加额步长。
- 创建 Token 时会返回默认、Clash/Mihomo 和 sing-box 三类订阅地址；地址优先基于当前管理后台访问 Host / `X-Forwarded-*` 头生成，管理后台创建结果同步展示，Token 列表会加密恢复并反复展示订阅地址，便于后续复制分发给不同客户端。
- 订阅内容里的 Clash/Mihomo `server` 和 sing-box outbound `server` 会优先使用当前订阅请求 Host 并去掉管理端口，端口仍来自虚拟节点监听端口；请求 Host 不安全时才回退到 `GATEWAY_HOST` 或 `PUBLIC_BASE_URL` 主机名，避免真实客户端拿到 `gateway.example.com` 模板占位域名。
- 管理后台 Token 列表支持重置订阅地址；历史旧 Token 如果没有可恢复订阅密钥，会提示重置后重新获取，新地址生成后旧订阅地址立即失效。
- 上游来源页面创建和列表。
- 管理后台上游来源支持行内编辑名称、类型、URL、前缀、默认标签和刷新间隔；前缀变更会同步刷新自动命名节点。
- subscription 类型上游来源可保存 URL 或 raw content，并可手动刷新导入节点。
- 上游来源 `default_tags` 会在节点导入或刷新时同步到节点标签。
- subscription 来源刷新时，本次订阅中消失的旧节点会标记为 `inactive`。
- subscription 来源刷新可识别当前已支持 outbound 的 URI 协议列表。
- subscription 来源支持解析 JSON URI 数组和常见 `nodes`/`links`/`subscription`/`raw_content` 包装对象，包装字段内的 URI 列表、Clash YAML、SIP008、sing-box JSON 和 base64 内嵌订阅内容也会归一化。
- JSON 包装订阅支持 `payload`、`result`、`response`、`body`、`text`、`sub` 等常见接口外层字段，不会把这些包装字段误当成节点名称。
- JSON 包装订阅会忽略 `meta`、`pagination`、`errors`、`message`、`traceId` 等常见接口元信息字段，避免把接口文档或分页 URL 误导入为 HTTP 节点。
- JSON 包装订阅会忽略 Clash 风格 `proxy-providers`、`rule-providers`、`proxy-groups`、`health-check` 和 `rules` 中的下载 URL、健康检查 URL 和规则源 URL，避免误导入为 HTTP 节点；其中 `proxy-providers` 内嵌的 `proxy`/`proxies`、`node`/`nodes`、`server`/`servers`、`proxy-list`/`proxies-list`、`nodeList`/`nodesList`、`serverList`/`serversList` 节点清单会继续按结构化节点解析。
- JSON 包装订阅支持 `list`、`records`、`rows`、`entries`、`node`/`nodes`、`proxy`/`proxies`、`server`/`servers`、`nodeList`/`nodesList`、`proxyList`/`proxiesList`、`serverList`/`serversList`、`subscriptionList`/`subscriptionsList`、`urlList`/`urlsList`、`linkList`/`linksList` 及其 snake_case/kebab-case 变体等常见集合字段；这些容器名不会污染缺少 fragment 的节点 URI 展示名。
- JSON 包装订阅字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Data.NodeList`、`Proxy_List` 和 `Payload.LinkList`。
- JSON 包装订阅节点名称字段也兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Name`、`Display_Name` 和 `NodeName`，缺少 fragment 的链接会继续保留真实展示名。
- JSON 包装订阅支持 `shareUrl`、`shareLink`、`subscriptionUrl`、`nodeUrl` 及其 snake_case/kebab-case 变体；对象内带名称的 `subscribeUrl`、`subUrl` 和 `downloadUrl` 分享字段也会按节点链接解析，分享链接缺少 fragment 时会优先使用对象内名称而不是字段名。
- JSON 包装订阅支持解析 `nodes`/`proxies` 中的 Clash 风格结构化节点对象，例如 `type`、`server`、`port`、`cipher`、`password` 和嵌套 `ws-opts`。
- JSON 结构化节点支持常见字段别名，例如 `protocol`/`proto`/`scheme`/`protocolName`/`nodeType`/`serverType`/`protocolType`/`proxyType`/`proxyProtocol`、`host`/`hostname`/`address`/`serverAddress`/`serverHost`/`remoteHost`/`nodeHost`/`endpoint`/`add`、`server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber`、`method`/`encryptMethod`/`encrypt-method`、`pass`/`passwd`/`pwd`/`secret`/`credential`/`accountPassword`、`username`/`userName`/`user_name`/`accountName`/`account`、`id`/`user-id` 和 `displayName`/`nodeName`/`label`/`title`/`remarks`/`remark`；`server`/`host`/`address`/`endpoint` 写成 `host:port` 时会自动拆分端口。
- JSON 结构化节点会归一化常见 camelCase/snake_case/kebab-case 参数，例如 `skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allowInsecure`、`allow_insecure`、`allow-insecure`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`clientFingerprint`、`flowName`、`xtlsFlow`、`networkType`、`transportType`、`tlsEnabled`、`enableTLS`、`overTLS`、`tlsHost`、`tlsServerName`、`tlsSettings`、`tlsOptions`、`pluginOpts`、`wsOpts`、`wsSettings`、`wsHeaders`、`grpcSettings`、`httpSettings`、`h2Settings`、`httpUpgradeSettings`、`quicSettings`、`tcpSettings`、`grpcServiceName`、`privateKeyPath`、`publicKey`、`shortId`、`shortIds`、`spiderX`、`serverNames`、`server_names`、`server-name` 等，并支持 `tls.enabled`、`tls.enable`、`tls.serverName`、`tls.serverNames`、`tls.server_names`、`tls.server-name`、`tls.tls_server_name`、`tls.utls.fingerprint`、`reality.publicKey` 和 `reality.serverNames` 等嵌套字段，避免导入后丢失 TLS、REALITY、插件和传输配置。
- JSON 结构化节点支持通用 `transport` 容器，可从 `transport.type`、`transport.path`、`transport.host`、`transport.headers.Host`、`transport.authority`、`transport.headers.authority`、`transport.headers.:authority`、`transport.serviceName`、`transport.idleTimeout` 等字段生成 WebSocket、HTTP、HTTPUpgrade 或 gRPC 传输参数。
- subscription 来源支持解析 Quantumult X `[server_local]`/`[server_remote]` 常见 SS、SSR、Hysteria2/Hy2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、Trojan、VLESS、VMess、HTTP 和 SOCKS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名，转换为标准 URI。
- Quantumult X `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Quantumult X、Surge、Clash YAML 和 sing-box JSON 的 `tuic-v5`/`tuic5` 类型会按 TUIC 兼容链路归一化为 `tuic://`，保留 UUID、密码、拥塞控制、UDP relay、SNI 和节点名。
- Quantumult X 和 Surge 结构化 VLESS/Trojan 节点支持保留 gRPC transport 和 service name，并同步到 sing-box outbound。
- Quantumult X HTTP/SOCKS 节点支持键值和位置参数两种认证写法。
- subscription 来源支持解析以节点名称为 key 的 JSON URI 对象映射；当 URI 缺少 fragment 时会使用映射 key 或对象内 `name` 作为节点名。
- subscription 来源支持解析裸 VMess JSON 单对象、数组、包装字符串和按名称映射对象，并转换为标准 `vmess://` URI；裸 VMess JSON 支持 `address`/`server`/`serverHost`/`nodeHost`/`endpoint`、`serverPort`/`nodePort`/`portNumber`、`uuid`/`user_id` 等字段别名，会保留 `packetEncoding`/`packet_encoding`/`packet-encoding`、`disable_sni` 和 `fp`/`fingerprint`/`clientFingerprint` 并同步为 sing-box `packet_encoding` 和 uTLS fingerprint。
- 裸 VMess JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Display_Name`、`Server-Port`、`User_ID`、`Packet_Encoding`、`Disable-SNI` 和 `Client-Fingerprint`。
- subscription 来源支持解析 SSD/ShadowsocksD `ssd://` 订阅和裸 SSD JSON，并展开为标准 `ss://` URI；SSD `servers` 支持数组和按名称分组的对象映射，并兼容 SIP008 同款 Shadowsocks 字段别名。
- SIP008 和 SSD JSON 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Version`、`Servers`、`Display_Name`、`Server-Port`、`Encrypt_Method`、`PluginOpts` 和 `PluginOptions`。
- subscription 来源支持解析 Surge `[Proxy]` 代理段中的 SS、SSR、Trojan、VLESS、VMess、Hysteria2、TUIC、Juicity、Hysteria、AnyTLS、ShadowTLS、Naive、SSH、WireGuard、HTTP/HTTPS、SOCKS、Direct、Reject 和 DNS 节点，并兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https`、`naive-quic` 和 `socks5h` 协议别名，转换为标准 URI。
- Surge `[Proxy]` `ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- subscription 来源支持解析 Clash YAML `proxies` 中的常见 SS/Trojan/VLESS/VMess/Hysteria2/TUIC/Juicity/Hysteria/HTTP/SOCKS/AnyTLS/ShadowTLS/Naive/SSH/WireGuard 节点，以及 Direct/Reject 变体/DNS 内置出站；Clash YAML 结构化 `type` 兼容 `ssr`/`shadowsocksr`、`trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 别名。
- Clash YAML `type: ssr`/`shadowsocksr` 节点会按 Shadowsocks 兼容链路归一化为 `ss://`，保留服务端、端口、加密方法、密码和节点名，并避免 SSR 专有 `protocol`/`obfs` 参数误写入 sing-box Shadowsocks `network`。
- Clash YAML 解析支持嵌套 `proxy`/`proxies`/`proxy-list`/`proxies-list`/`node`/`nodes`/`node-list`/`server`/`servers`/`server-list` 列表，并兼容 `proxyList`、`nodeList`、`serverList` 等 camelCase/PascalCase 写法，可兼容带内嵌 provider 节点清单的订阅结构。
- Clash YAML 解析支持 `proxy: [{ ... }]`、`proxies: [{ ... }]`、`proxy-list: [{ ... }]`、`proxies-list: [{ ... }]`、`nodes: [{ ... }]`、`servers: [{ ... }]` 和 `node-list: [{ ... }]` 等内联节点数组，并兼容 camelCase/PascalCase 容器名，可兼容 provider 或顶层节点的紧凑写法。
- Clash YAML 解析支持 `proxies: &anchor` 这类带 YAML anchor 的块状节点列表。
- Clash YAML 解析支持 `- &anchor { ... }` 和 `proxies: [&anchor { ... }]` 这类节点条目级 anchor 的内联写法。
- Clash YAML 解析支持常见块状和内联 `ws-opts`、`grpc-opts` 嵌套写法，能保留 WebSocket path/host 和 gRPC service name。
- Clash YAML 解析会归一化常见 camelCase 参数，例如 `skipCertVerify`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`wsPath`、`wsHost`、`wsOpts`、`maxEarlyData`、`earlyDataHeaderName`、`grpcServiceName` 等，块状和内联节点都会保留 TLS 与传输配置。
- Clash YAML 解析支持行内和块状数组标量，例如 `alpn: [h2, http/1.1]`、`alpn: ... - h2` 和 WireGuard `allowed-ips`/`reserved` 列表。
- subscription 来源支持解析 SIP008 Shadowsocks 订阅，`servers` 支持数组和按名称分组的对象映射；SIP008 Shadowsocks 节点可识别 `address`/`serverAddress`/`nodeHost`/`endpoint`、`portNumber`/`nodePort`、`encryptMethod`/`security` 和 `pass`/`passwd` 等字段别名。
- SIP008、Clash YAML 和 sing-box JSON 的 Shadowsocks 节点会保留 SIP003 `plugin`、`plugin_opts` 和 `network` 参数，并同步到 sing-box outbound。
- `shadowsocks://` URI 会归一化为标准 `ss://` URI，避免长 scheme 导入后无法进入 Shadowsocks outbound 转换链路。
- `ssr://` ShadowsocksR URI 会解码为标准 `ss://` URI，保留服务端、端口、加密方法、密码和 remarks 节点名；SSR 专有 protocol/obfs 参数暂按 Shadowsocks 兼容链路导入。
- `http+tls://` 和 `http-tls://` 裸 URI 会归一化为标准 `https://`，复用现有 HTTP outbound TLS 配置生成链路。
- `any-tls://`、`shadow-tls://`、`naive-https://` 和 `naive-quic://` 裸 URI 会归一化为标准 `anytls://`、`shadowtls://`、`naive+https://` 和 `naive+quic://`，和结构化订阅解析保持一致。
- `hy2://` 和 `wg://` 裸 URI 会归一化为标准 `hysteria2://` 和 `wireguard://`，让导入后的协议字段和配置生成链路保持一致。
- `trojan-go://` 裸 URI 会归一化为标准 `trojan://`，复用现有 Trojan TLS、REALITY 和传输参数转换链路。
- `vmess-aead://` 裸 URI 会归一化为标准 `vmess://`，复用现有 VMess TLS、WebSocket、gRPC、HTTP 和 packet encoding 转换链路。
- `tuic-v5://` 和 `tuic5://` 裸 URI 会归一化为标准 `tuic://`，复用现有 TUIC outbound 转换链路；手动保存 `tuic-v5`/`tuic5` 协议节点时配置生成侧也会按 TUIC 输出。
- subscription 来源支持解析 sing-box JSON `outbounds` 中的常见 Shadowsocks/Trojan/VLESS/VMess/Hysteria2/TUIC/Juicity/AnyTLS/ShadowTLS/Naive/Hysteria/HTTP/SOCKS/SSH/WireGuard/Tor 节点，并兼容 `trojan-go`、`vmess-aead`、`hy2`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive+https`、`naive-https`、`naive-quic` 和 `socks5h` 结构化 `type` 别名。
- subscription 来源支持解析 sing-box JSON 顶层 `endpoints` 中的 WireGuard endpoint 模型，并转换为当前 Phase 1 WireGuard outbound 兼容 URI。
- sing-box JSON 顶层可兼容单数 `outbound`/`endpoint` 和 `outboundList`/`outboundsList`、`endpointList`/`endpointsList` 写法；也兼容顶层直接给出 outbound/endpoint 数组、单个对象或按名称映射对象；单对象、数组和对象映射都会进入同一归一化链路。
- sing-box JSON 顶层、outbound、endpoint、TLS、uTLS、REALITY、transport、headers 和 WireGuard peer 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Server-Port`、`ServerPort`、`ServerName`、`Disable-SNI`、`ServiceName`、`PermitWithoutStream`、`PrivateKey`、`AllowedIPs`。
- sing-box JSON `outbounds` 和 `endpoints` 支持数组、单个对象、按名称分组的对象映射和分组数组映射；映射对象缺少 `tag`/`name` 时会使用映射键或分组序号作为节点名。
- sing-box JSON transport 的 `host`、`authority`、`headers.Host`、`headers.authority` 和 `headers.:authority` 支持字符串或数组写法，WebSocket、HTTP、HTTPUpgrade 和 VMess WebSocket/HTTP/HTTPUpgrade transport 会保留为逗号分隔 Host。
- subscription 来源支持解析 V2Ray/Xray JSON `outbounds` 中的 VMess、VLESS、Trojan、Shadowsocks、HTTP、SOCKS、freedom、blackhole、reject 变体和 DNS 出站，并兼容 `trojan-go`、`vmess-aead`、`http+tls`、`http-tls`、`socks4`、`socks4a`、`socks5`、`socks5h` 协议别名，保留常见 TLS、REALITY、WebSocket、gRPC、QUIC、HTTP/H2 和 HTTPUpgrade 传输参数；REALITY 兼容 `serverName`/`serverNames`、`shortId`/`shortIds` 和 hyphen/snake-case 别名。
- V2Ray/Xray JSON `streamSettings`、TLS/REALITY 设置和各 transport 设置块支持 camelCase、snake_case 和 hyphen-case 容器字段别名。
- V2Ray/Xray JSON 顶层、outbound、settings、vnext/server、user/account、streamSettings、TLS、uTLS、REALITY、WebSocket、gRPC、HTTP、HTTPUpgrade 和 TCP header 字段匹配兼容大小写差异和 snake_case/kebab-case/PascalCase 变体，例如 `Outbounds`、`Settings`、`VNext`、`Server-Port`、`StreamSettings`、`TLSSettings`、`ServerName`、`Disable-SNI`、`WSSettings`、`MaxEarlyData`、`GrpcSettings` 和 `PermitWithoutStream`。
- V2Ray/Xray JSON Shadowsocks outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`security`/`encryption`/`encryptMethod` 字段别名，并保留 SIP003 `plugin`、`plugin_opts`、`plugin_options`、`network` 和 `protocol` 标志，结构化导入后会继续进入 Shadowsocks 转换链路。
- V2Ray/Xray JSON 普通 TLS 会兼容 `serverName`/`serverNames` 以及 hyphen/snake-case SNI 别名，数组写法会选取首个有效值同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会兼容 `allowInsecure`、`skip-cert-verify`、`insecure` 和 `disable_sni` 写法，并同步到生成的节点 URI。
- V2Ray/Xray JSON 普通 TLS 会保留 `fingerprint`、`clientFingerprint` 或 `utls.fingerprint` 客户端指纹，并同步为 sing-box uTLS fingerprint。
- V2Ray/Xray JSON 兼容旧配置中的单个 `outbound`、`outboundDetour`、复数 `outboundDetours` 和 `outboundList`/`outboundsList` 出站入口，会和现代 `outbounds` 一起归一化导入。
- V2Ray/Xray JSON 也兼容顶层直接给出 outbound 数组、单个 outbound 对象或按名称映射的 outbound 对象，映射 key 写成 `host:port` 时会同步补齐节点服务端、端口和默认节点名。
- V2Ray/Xray JSON `wsSettings.host`、`wsSettings.headers.Host` 和 `authority` 支持字符串或数组写法，数组会保留为逗号分隔的 WebSocket Host。
- V2Ray/Xray JSON `wsSettings` 支持 `maxEarlyData`、`max_early_data`、`max-early-data`、`earlyDataHeaderName`、`early_data_header_name` 和 `early-data-header-name` WebSocket early data 别名。
- V2Ray/Xray JSON `tcpSettings.header.type=http` 的 HTTP 伪装 Host、path 和 method 会保留为 HTTP transport 参数。
- V2Ray/Xray JSON `httpSettings`/`h2Settings` 支持从 `host`、`headers.Host` 和 `authority` 读取 HTTP/H2 transport Host，并支持数组写法的 path。
- V2Ray/Xray JSON `httpupgradeSettings.host` 支持字符串或数组写法，也可从 `headers.Host` 读取 Host 列表。
- V2Ray/Xray JSON `grpcSettings` 的 `idle_timeout`、`health_check_timeout` 和 `permit_without_stream` 会保留为 gRPC transport keepalive 参数。
- V2Ray/Xray JSON `grpcSettings.multiMode` 会保留为 sing-box gRPC transport 的 `multi_mode` 参数。
- V2Ray/Xray JSON `grpcSettings` 支持 `service-name`、`idle-timeout`、`health-check-timeout`、`permit-without-stream` 和 `multi-mode` 这类 hyphen 别名。
- V2Ray/Xray JSON 存在 `quicSettings`/`quic_settings`/`quic-settings` 时会保留为 QUIC transport。
- V2Ray/Xray JSON VMess/VLESS `vnext` endpoint 支持 `address`、`server`、`host`、`add`、`serverAddress`、`serverHost`、`remoteHost`、`nodeHost`、`endpoint` 和 `server_port`/`serverPort`/`server-port`/`remotePort`/`nodePort`/`portNumber` 别名；VMess/VLESS user 支持 `id`、`uuid`、`userId`、`user_id` 和 `user-id` 认证字段，VMess user 支持 `alter-id`/`aid` 和 `cipher` 别名，VMess/VLESS user 支持 `flow`/`flowName`/`xtlsFlow` 等 flow 别名，并支持 `packetEncoding`/`packet_encoding`/`packet-encoding` 同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON VMess/VLESS `vnext.users` 支持以 UUID 为 key 的对象映射；映射值可补充 `email`/`name`、`security`、`flow` 等字段，缺少 `id` 时会使用映射 key 作为用户 ID。
- V2Ray/Xray JSON `vnext` 和 `servers` 支持以服务端域名或 IP 为 key 的对象映射；对象缺少 `address`/`server` 时会使用映射 key 作为节点服务端，映射 key 写成 `host:port` 或 `[ipv6]:port` 时会同步拆出端口；VMess/VLESS 的 `vnext` 映射值也支持直接写用户对象、用户数组、UUID 标量、UUID 数组或账号名到 UUID/用户对象的映射，Trojan/Shadowsocks 的 `servers` 映射值支持直接写密码字符串或密码数组。
- V2Ray/Xray JSON VMess、VLESS、Trojan 和 Shadowsocks 支持缺省 `settings.vnext`/`settings.servers` 的扁平 outbound 写法；服务端、端口、认证和加密字段可直接写在 outbound 顶层或 `settings` 顶层。
- Clash YAML 和 sing-box JSON VMess/VLESS 节点支持 `packetEncoding`、`packet_encoding` 和 `packet-encoding` 写法，并同步为 sing-box `packet_encoding`。
- V2Ray/Xray JSON Trojan/HTTP/SOCKS outbound 支持 `host`/`add`、`serverPort`/`server-port`、`pass`/`passwd`/`pwd`/`psk`/`token`、`userName`/`accountName`、`users`、`accounts` 数组、`accounts` 用户名到密码或对象映射和 server 层认证写法；HTTP/SOCKS 的 `servers` 端点映射值也支持直接写账号名到密码或账号对象的映射，并兼容 `user:password` 标量或标量数组。
- subscription 来源支持按 `refresh_interval_minutes` 定时同步；后台调度默认每 60 秒检查一批到期来源。
- 上游来源自动前缀。
- 重复来源名前缀自动编号，例如 `[机场A]`、`[机场A-2]`。
- 节点 URI 页面批量导入。
- 节点 `raw_name`、`display_name`、`name_mode`。
- 管理 API 和管理后台可查看单个节点详情，包括 URI、服务器、端口、Hash、状态、标签和时间信息。
- 管理后台所有时间字段统一按东八区展示，后端仍以 UTC 作为存储和计算基准。
- 管理后台节点池默认按地区聚合展示，缺失地区的节点归入“其他”；进入地区后展示节点卡片，再点击卡片查看节点详情。
- 管理后台节点池地区返回和筛选清空控件已统一为符号按钮，和节点卡片操作区保持一致。
- 节点导入、读取和历史数据迁移会识别常见地区并统一带国旗展示；中国香港/中国台湾统一归为 `🇨🇳中国|香港`、`🇨🇳中国|台湾`，`香港-美国` 这类路径命名会按右侧目的地归为 `🇺🇸美国`。
- 单节点手动改名保护。
- 单节点恢复自动命名。
- 管理后台节点池支持行内编辑节点展示名，并可恢复来源前缀驱动的自动命名。
- 虚拟节点页面支持创建、列表和行内编辑名称、监听协议、监听端口、标签选择器、出口策略和启停状态。
- 虚拟节点 `tag_selector` 可按上游节点标签生成专属 sing-box selector，并生成入站到该 selector 的 route rule；无匹配上游时会路由到 `block`。
- Clash/Mihomo 订阅生成。
- sing-box 客户端订阅生成。
- sing-box 服务端配置生成骨架。
- sing-box config check API 和管理后台检查入口，可返回配置 hash、入站/出站/用户摘要。
- sing-box config publish API 和管理后台发布入口，可写入当前配置并保存上一版文件。
- sing-box config rollback API 和管理后台回滚入口，可恢复上一版配置文件。
- sing-box config publish/rollback API 会返回 `restart_required`，管理后台会提示发布或回滚后需要重启 sing-box。
- sing-box restart API 和管理后台重启入口已落地；发布/回滚可在 `SING_BOX_AUTO_RESTART=true` 时自动调用受配置保护的重启命令。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2/Hysteria/TUIC/Juicity/AnyTLS/ShadowTLS/Naive/HTTP/SOCKS/SSH/WireGuard/Tor 上游节点会转换为 sing-box outbound，并通过默认 selector 承接网关出口。
- 配置生成侧兼容导入层已支持的 URI 协议别名，包括 `shadowsocks`、`trojan-go`、`vmess-aead`、`tuic-v5`、`tuic5`、`http+tls`、`http-tls`、`any-tls`、`shadow-tls`、`naive-https` 和 `naive-quic`，手动保存这些协议名时也会归一为对应 sing-box outbound。
- Direct URI、`freedom://` URI 和 sing-box JSON `direct` outbound 可导入为直连上游 outbound，并进入默认上游 selector 和 V2Ray outbound stats 列表；配置生成侧也兼容手动保存的 `freedom` 协议节点。
- Block URI、`blackhole://` URI、`reject://` URI、Clash `reject`/`reject-drop` 和 sing-box JSON `block` outbound 可导入为内部拦截 outbound；为避免误承接普通代理流量，Block outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表；配置生成侧也兼容手动保存的 `blackhole`/`reject`/`reject-drop`/`reject-no-drop`/`reject-tinygif` 协议节点。
- `reject-drop://`、`reject-no-drop://` 和 `reject-tinygif://` 裸 URI 会归一化为标准 `block://`，和 Clash 结构化 reject 变体保持一致。
- DNS URI 和 sing-box JSON `dns` outbound 可导入为内部 DNS outbound；为避免误承接普通代理流量，DNS outbound 不进入默认上游 selector 和 V2Ray outbound stats 列表。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan/VMess 的 `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 uTLS fingerprint TLS 参数，并同步到 sing-box outbound。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS WebSocket 的 `type=ws`、`path` 和 `host` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `wsPath` 和 `wsHost` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan WebSocket 的 `max_early_data` 和 `early_data_header_name` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `maxEarlyData` 和 `earlyDataHeaderName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS gRPC 的 `type=grpc` 和 `service_name` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `serviceName` 和 `grpcServiceName` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan gRPC 的 `idle_timeout`、`ping_timeout`、`permit_without_stream` 和 `multi_mode` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `grpcIdleTimeout`、`grpcPingTimeout`、`permitWithoutStream`、`multiMode` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan QUIC transport，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTP transport 的 `host`、`path`、`method`、`idle_timeout` 和 `ping_timeout` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `httpHost`、`httpPath`、`httpMethod`、`httpIdleTimeout` 和 `httpPingTimeout` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS/Trojan HTTPUpgrade transport 的 `host` 和 `path` 参数，并同步为 sing-box outbound transport，兼容 URI 中 `httpUpgradeHost` 和 `httpUpgradePath` 查询别名。
- VLESS、Trojan 和 VMess URI 生成 sing-box outbound 时，WebSocket、HTTP/H2 和 HTTPUpgrade transport Host 兼容 `authority`、`:authority`、`headers.Host`、`headers.:authority`、`headerHost` 和协议专属 Host 查询别名。
- VLESS URI 可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，Trojan URI 可从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- VLESS 和 Trojan URI 可通过 `tls=tls`、`tls=true`、`tlsEnabled`、`enableTLS` 或 `overTLS` 查询参数显式启用普通 TLS，兼容未写 `security=tls` 的订阅写法。
- VLESS 和 Trojan URI 的 SNI 兼容 `sni`、`servername`、`server_name`、`server-name`、`serverName`、`tlsHost`、`tls_host`、`tls-host`、`tlsServerName`、`tls_server_name` 和 `tls-server-name` 查询别名。
- VLESS 和 Trojan URI 的跳过证书校验兼容 `insecure`、`skip-cert-verify`、`skip_cert_verify`、`skipCertVerify`、`skipCertificateVerify`、`skipVerify`、`allow-insecure`、`allowInsecure`、`tlsAllowInsecure` 和 `tlsSkipVerify` 查询别名。
- VLESS 和 Trojan URI 的禁用 SNI 兼容 `disable_sni`、`disable-sni`、`disableSNI`、`disableSni`、`tlsDisableSNI`、`tlsDisableSni` 和 `tls-disable-sni` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VLESS REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure` 和 `disableSNI` 查询别名，同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan REALITY 的 public key、short id 和 uTLS fingerprint 参数，并兼容 URI 中 `serverName`、`publicKey`、`shortId`、`clientFingerprint`、`allowInsecure`、`skip_cert_verify` 和 `disableSNI` 查询别名，同步为 sing-box outbound TLS 配置。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Trojan WebSocket/gRPC 传输参数，并同步为 sing-box outbound transport。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 VMess gRPC 传输参数，并同步为 sing-box outbound transport，兼容 `serviceName`、`grpcServiceName`、`multiMode` 和 `grpcMultiMode` 写法。
- VMess TCP HTTP 伪装和 HTTP/H2 传输参数会同步为 sing-box HTTP transport。
- VMess HTTPUpgrade 传输参数会同步为 sing-box HTTPUpgrade transport，兼容 VMess JSON `net=httpupgrade` 以及 userinfo URI 中 `httpUpgradeHost`/`httpUpgradePath` 查询别名。
- sing-box JSON VMess 节点的 HTTP transport 会保留 host 数组和 path，避免导入后丢失 HTTP 伪装参数。
- VMess 可兼容 `vmess://uuid@host:port?...` userinfo 直连 URI，也可从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 查询参数读取认证信息，并保留 TLS、WebSocket、gRPC、SNI、ALPN、跳过证书校验、uTLS fingerprint 和 `packet_encoding`/`packet-encoding`/`packetEncoding` 参数；userinfo URI 兼容 `serverName`、`tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`disableSNI`、`tlsDisableSNI`、`tlsEnabled`、`enableTLS`、`overTLS`、`grpcServiceName` 和 `grpcMultiMode` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria `recv_window_conn`、`recv_window`、`disable_mtu_discovery` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMTUDiscovery` 和 `clientFingerprint` 查询别名。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 Hysteria2 `up_mbps`、`down_mbps`、`insecure`、`disable_sni`、`alpn`、证书 pin、`fp`/`client-fingerprint` 和 `tls.utls.fingerprint` 参数，并同步为 sing-box outbound，兼容 URI 中 `authStr`、`pass`、`passwd`、`pwd`、`secret`、`credential`、`accountPassword`、`serverName`、`allowInsecure`、`disableSNI`、`upMbps`、`downMbps`、`obfsPassword`、`pinSHA256`、`pin_sha256`、`certificatePublicKeySHA256`、`certificate_public_key_sha256` 和 `clientFingerprint` 查询别名。
- Hysteria 和 Hysteria2 URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Hysteria/Hysteria2 结构化节点会保留 `token`/`pass`/`passwd` credential 别名、`obfsPassword`、`upMbps`、`downMbps`、`recvWindowConn`、`recvWindow`、`disableMtuDiscovery`、`disableMTUDiscovery`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 等 camelCase 参数，并归一化到 Hysteria/Hysteria2 URI 到 sing-box outbound 生成链路。
- TUIC URI 支持从 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，并兼容 `congestionControl`、`udpOverStream`、`udpRelayMode`、`zeroRttHandshake`、`heartbeatInterval`、`serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 查询别名，兼容缺少 userinfo 的订阅写法。
- TUIC URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML TUIC 结构化节点会保留 credential 别名、`udp-over-stream`/`udpOverStream`、`zero-rtt-handshake`/`zeroRttHandshake`、`heartbeat-interval`/`heartbeatInterval`、`disable-sni`/`disable_sni` 和 `client-fingerprint`/`clientFingerprint` 等参数，并归一化到 TUIC URI 到 sing-box outbound 生成链路。
- sing-box JSON 和 URI 导入链路会保留 TUIC `insecure`/`skip-cert-verify`、`disable_sni`、`alpn`、`tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound。
- Clash YAML、Quantumult X、Surge、sing-box JSON 和 URI 导入链路会保留 Juicity UUID、密码、`congestion_control`、SNI、跳过证书校验、`disable_sni`、ALPN 和 `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound；URI 生成侧兼容 `uuid`/`id`/`user_id`/`user-id`/`userId`/`userID`/`userid` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询别名。
- Juicity URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML Juicity 结构化节点会保留 `uuid`/`id`/`user-id` credential 别名、`congestionControl`/`congestion-controller`、`skip-cert-verify`、`disable-sni`/`disable_sni`、ALPN 和 `client-fingerprint`/`clientFingerprint` 客户端指纹，并归一化到 Juicity URI 到 sing-box outbound 生成链路。
- AnyTLS 和 ShadowTLS URI 支持从 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- sing-box JSON 和 URI 导入链路会保留 AnyTLS 会话空闲参数和 `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`idleSessionCheckInterval`、`idleSessionTimeout`、`minIdleSession` 和 `clientFingerprint` 查询别名。
- sing-box JSON 和 URI 导入链路会保留 ShadowTLS `tls.utls.fingerprint`/`fp` 客户端指纹，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI` 和 `clientFingerprint` 查询别名。
- AnyTLS 和 ShadowTLS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Clash YAML AnyTLS/ShadowTLS 结构化节点会保留 `password`/`pass`/`passwd`/`psk`/`token` credential 别名；AnyTLS 会同步 `idleSessionCheckInterval`、`idleSessionTimeout` 和 `minIdleSession` 等会话空闲参数；两者都会通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- sing-box JSON 和 URI 导入链路会保留 Naive/Naive+QUIC `insecure`/`skip-cert-verify`、`disable_sni`、`alpn`、QUIC 控制项和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`insecureConcurrency`、`udpOverTcp`、`quicCongestionControl` 和 `clientFingerprint` 查询别名。
- Clash YAML Naive 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `insecureConcurrency`、`udpOverTcp` 和 `quicCongestionControl` 等 camelCase QUIC 控制参数，并通过公共 TLS helper 保留 `serverName`、`allowInsecure`、`disableSNI`、ALPN 和 `clientFingerprint` 客户端指纹。
- Clash YAML、sing-box JSON 和 URI 导入链路会保留 HTTP/HTTPS 代理 `insecure`/`skip-cert-verify`、`disable_sni`、`alpn` 和 `tls.utls.fingerprint`/`fp` 参数，并同步为 sing-box outbound，兼容 URI 中 `serverName`、`allowInsecure`、`disableSNI`、`clientFingerprint`、`tlsEnabled`、`enableTLS` 和 `overTLS` 查询别名。
- Naive 和 HTTP/HTTPS URI 的 SNI、跳过证书校验和禁用 SNI 查询参数复用公共 TLS 别名解析，兼容 `tlsHost`、`tlsServerName`、`skipVerify`、`tlsAllowInsecure`、`tlsSkipVerify`、`tlsDisableSNI` 和 `tlsDisableSni` 等写法。
- Naive、HTTP/HTTPS 和 SOCKS URI 支持从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息，兼容缺少 userinfo 的订阅写法。
- SSH URI 支持从 `username`/`user`/`userName`/`accountName`/`login` 与 `password`/`pass`/`passwd`/`pwd`/`accountPassword` 查询参数读取认证信息，并兼容 `identity_file`、`key_path`、`key`、`privateKey`、`privateKeyPath`、`privateKeyPassphrase`、passphrase、`hostKey`、`hostKeyAlgorithms`、`clientVersion`、`cipher`/`ciphers`、`mac`/`macAlgorithms` 和 `kexAlgorithm`/`kexAlgorithms` 查询别名，兼容缺少 userinfo 的订阅写法。
- Clash YAML SSH 结构化节点会保留 `pass`/`passwd` credential 别名，以及 `privateKeyPath`、`privateKeyPassphrase`、`hostKey`、`hostKeyAlgorithms`、`clientVersion` 和 `kexAlgorithm` 等 camelCase 参数，并归一化到 SSH URI 到 sing-box outbound 生成链路。
- SOCKS URI 会按 `socks4://`、`socks4a://`、`socks5://` 或 `version=4/4a/5` 查询参数保留版本，并同步为 sing-box SOCKS outbound 的 `version` 字段。
- SOCKS URI 支持 `udp`、`udpEnabled`、`udp_relay` 和 `udp-relay` 标志，并会归一为 sing-box outbound 的 `network=udp`，同时兼容 `udpOverTcp` 查询别名。
- `socks5h://` 裸 URI 会归一化为标准 `socks5://`，兼容 curl/requests 生态里常见的 SOCKS5H 订阅写法；配置生成侧也兼容手动保存的 `socks5h` 协议节点并按 SOCKS5 outbound 输出。
- Clash YAML SOCKS 节点支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Surge SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- Quantumult X SOCKS 节点支持保留 `udp=true` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- sing-box JSON SOCKS outbound 支持保留 `udp`、`udp_relay` 和 `udp-relay` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- V2Ray/Xray JSON SOCKS outbound 支持保留 `udp`、`udpEnabled`、`udp_enabled`、`udp_relay`、`udp-relay`、`udp_over_tcp`、`udpOverTcp`、`udp-over-tcp` 和 `uot` 标志，结构化导入后会继续进入 SOCKS UDP 转换链路。
- WireGuard URI 支持 `privateKey`/`secretKey`、`publicKey`/`peerPublicKey`、`preSharedKey`、`ip`/`ipv6`/`localAddresses`、`allowedIPs`/`allowedIps`/`peerAllowedIps`、`systemInterface`、`interfaceName`、`reservedBytes` 和 `peerReserved` 查询参数别名，并同步为 sing-box outbound。
- Clash YAML WireGuard 结构化节点会保留 `privateKey`、`peerPublicKey`、`localAddress`/`localAddresses`、`preSharedKey`、`allowedIPs`/`peerAllowedIps`、`reservedBytes`/`peerReserved`、`systemInterface` 和 `interfaceName` 等 camelCase 参数，并归一化到 WireGuard URI 到 sing-box outbound 生成链路。
- Tor URI 支持 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments`/重复 `arg` 和 `torrc[Option]` 查询参数别名，并同步为 sing-box outbound。
- Clash YAML Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和嵌套 `torrc` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Surge 和 Quantumult X Tor 结构化节点会保留 `executablePath`、`dataDirectory`/`dataDir`、`extraArgs`/`arguments` 和 `torrc.*`/`torrc[Option]` 选项，并归一化到 Tor URI 到 sing-box outbound 生成链路。
- Shadowsocks URI 支持从 `method`/`cipher`/`encrypt-method`/`encrypt_method`/`encryptMethod`/`encryption`/`security` 与 `password`/`pass`/`passwd`/`pwd`/`psk`/`token`/`secret`/`credential`/`credentials`/`account_password`/`account-password`/`accountPassword` 查询参数读取认证信息，并兼容 `pluginOptions` 查询别名，适配缺少 userinfo 的订阅写法。
- Hysteria2/Hy2 URI 支持从 `password`、`auth`、`auth_str`、`authStr`、`pass`、`passwd`、`pwd`、`psk`、`token`、`secret`、`credential` 或 `accountPassword` 查询参数读取认证密码，兼容缺少 userinfo 的订阅写法。
- Hysteria v1 URI 支持从 `token`、`auth`、`auth_str`、`authStr`、`authBase64`、`pass`、`passwd`、`pwd`、`psk`、`secret`、`credential` 或 `accountPassword` 查询参数读取认证字符串，兼容部分上游订阅的 token 写法。
- sing-box 服务端配置生成会过滤已撤销、已过期、已超额或 gateway account 不可用的 Token。
- 流量统计入库骨架已落地：可解析 V2Ray stats 的 user/inbound/outbound 计数名称，写入 `traffic_samples`，计算 counter delta，并把 user 维度增量累加到 Token 用量和小时/天汇总表。
- V2Ray stats 计数名称解析会容忍分段前后空白和 `traffic`/方向字段大小写差异，降低不同实现输出格式抖动导致样本被丢弃的概率。
- user 维度流量入库后会自动判断 Token 额度；超额时将 Token 和 gateway account 标记为 `over_quota`，追加足够额度后自动恢复为 `active`。
- stats 轮询发现 Token 因真实流量进入 `over_quota` 时，会自动发布新的 sing-box 配置；如果部署侧开启 `SING_BOX_AUTO_RESTART`，会继续走受配置保护的重启执行器。
- 管理 API 和后台页面可查看 Token 今日、本月和累计流量用量摘要，展示最近 24 小时/最近 14 天流量图，并查看近 14 天上游出口流量摘要。
- 管理后台流量模块的 Token 用量和上游出口摘要已改为卡片式信息层级，两个分区不管有无数据都会展示统一符号、标题、数量徽标和说明；Token 用量卡增加状态、成员、今日和本月摘要芯片，上游出口卡增加来源、上传、下载和总量摘要芯片；Token 用量卡展示额度使用率进度条，并按接近超额和已超额状态变色；浏览器 QA 会验证流量分区头、卡片数量、摘要芯片、额度条和无视觉溢出。
- 管理后台流量趋势图已补齐小时/日摘要头，展示趋势符号、总量、峰值和活跃样本数，便于真实测试时先扫读流量是否进入统计链路；浏览器 QA 会验证两个趋势图摘要头、摘要芯片和无视觉溢出。
- 订阅响应头返回标准 `subscription-userinfo`，并提供 FluxGate 专属的已用、总额和剩余额度头。
- stats 可插拔轮询调度器已落地，能将采集器返回的 V2Ray counters 转换为流量样本并写入现有用量汇总链路。
- 真实 sing-box V2Ray gRPC stats 采集器已接入主进程；配置 `SING_BOX_V2RAY_API_ADDR` 后会按 `STATS_POLL_INTERVAL_SECONDS` 轮询并写入现有统计链路。
- sing-box 服务端配置生成会把 active 上游 outbound tag 写入 V2Ray stats 配置，支持上游出口流量汇总。
- Token 仍以 hash 作为鉴权主键；订阅密钥额外使用 `TOKEN_SECRET` 派生密钥加密保存，支持管理员后台按当前访问域名反复复制订阅地址。
- 订阅请求日志 Token 路径脱敏。
- 结构化 JSON 服务日志。
- 本地验证脚本。
- 公开仓库脱敏扫描脚本。
- 远程部署探测脚本。
- 磁盘清理脚本。
- 诊断采集脚本。
- 部署后真实可用性收口脚本，可基于已有团队、成员、Token 和上游节点创建缺失的默认虚拟节点与默认策略，校验并发布 sing-box 配置，远程配置存在时会重启 sing-box 数据面。
- 部署后真实可用探测脚本，可只读检查控制面核心资源、sing-box 配置入站/用户/上游摘要、局域网网关端口连通性；脚本会优先使用 `FLUXGATE_QA_SUBSCRIPTION_URL`，未提供时自动从管理员 Token 列表中选取可恢复订阅地址，验证默认、Clash/Mihomo 和 sing-box 客户端订阅正文包含当前虚拟节点。
- sing-box 当前配置校验并重载脚本，适用于配置文件已由控制面写入但需要数据面重新加载的场景。
- 推送后远程部署脚本。
- GitHub Actions Docker 镜像构建工作流，支持 PR 构建验证和分支/tag 推送 GHCR。
- 自定义 sing-box Dockerfile 已落地，默认远程构建带 `with_v2ray_api` 的本地数据面镜像，避免官方镜像缺少 V2Ray API 导致统计采集不可用。
- 自定义 sing-box Dockerfile 运行阶段使用 `scratch` 并从 builder 复制 CA 证书，减少远程构建对额外镜像仓库的依赖，降低 registry 元数据超时导致部署回退的概率。
- Docker 构建上下文脱敏，默认排除 `.env`、数据库、日志、测试产物和本机部署配置。
- 远程验收健康检查重试。
- 首次远程部署最小 sing-box config bootstrap。
- 远程 Docker build 支持 `GOPROXY`，默认优先使用 `goproxy.cn` 以避开 `proxy.golang.org` 超时。
- 推送后远程部署支持远程构建优先；当外部 registry 网络不可用时，会回退为本地交叉编译、scratch 镜像构建并流式加载到服务器。
- 部署时宿主机 HTTP 端口默认使用 `127.0.0.1:18080`，避免和服务器已有 8080 服务冲突。
- 部署侧可通过未跟踪配置打开局域网访问，不把真实环境信息提交到公开仓库。
- 页面截图验收脚本，支持指定目标 URL、保留截图和自定义输出目录。
- 浏览器登录验收脚本。
- 只读就绪检查脚本，可在本地或远程实例上检查健康状态、登录、概览、团队、成员、Token、来源、节点、虚拟节点、策略和流量摘要；`STRICT=true` 时缺少体验前置数据会使检查失败。
- 本地 QA 套件退出自动清理临时产物。
- 手把手使用说明已落地，覆盖部署初始化、管理员登录、添加上游订阅、节点地区聚合与详情查看、节点/来源编辑、团队/用户/Token 管理、订阅地址分发、流量额度查看、配置发布/回滚、远程运维和常见问题排查。

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
PATCH /api/teams/{id}

GET  /api/users
POST /api/users
PATCH /api/users/{id}

GET  /api/tokens
POST /api/tokens
POST /api/tokens/{id}/revoke
POST /api/tokens/{id}/restore
POST /api/tokens/{id}/rotate-subscription
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
PATCH /api/virtual-nodes/{id}

GET  /api/policies
POST /api/policies
PATCH /api/policies/{id}
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
scripts/qa/readiness.sh
scripts/qa/usable-probe.sh
scripts/qa/browser-login.sh
scripts/qa/screenshot.sh
scripts/qa/local-suite.sh
scripts/deploy/remote-logs.sh
scripts/deploy/probe-env.sh
scripts/deploy/ensure-usable.sh
scripts/deploy/push-and-deploy.sh
DRY_RUN=true scripts/deploy/cleanup-disk.sh
docker compose config
```

页面截图验收：

```text
KEEP_ARTIFACTS=true scripts/qa/screenshot.sh 可保留截图；默认测试退出会自动清理截图。
scripts/qa/screenshot.sh --keep http://<lan-host>:<port> 可对局域网部署页面保留人工复核截图。
scripts/qa/browser-login.sh http://<lan-host>:<port> 可做真实浏览器登录验收。
scripts/qa/readiness.sh http://<lan-host>:<port> 可做真实实例只读就绪检查；STRICT=true 会把缺少体验数据视为失败。
scripts/qa/usable-probe.sh http://<lan-host>:<port> 可做真实实例只读可用性检查，包括 sing-box 配置摘要、局域网网关端口连通，以及从后台 Token 列表自动取订阅地址后的正文探测。
```

截图结论：

- 页面可打开。
- 未登录时展示管理员登录页。
- 登录后管理员登录表单必须不可见，后台视图必须可见，并可加载仪表盘数据。
- 浏览器验收会确认概览、接入、节点、身份、策略、流量和运维 7 个模块视图可通过后台导航切换。
- 浏览器验收会确认 7 个创建/导入表单抽屉默认收纳、标题符号顺序正确且标题区无视觉溢出。
- 浏览器验收会确认每个模块标题区至少展示 3 个上下文摘要项，且摘要胶囊内部无视觉溢出。
- 浏览器验收会确认模块快速定位条的符号顺序、徽标数量、按钮内部无溢出和创建入口展开能力。
- 浏览器验收会确认 9 个模块面板符号和数量/状态徽标都已渲染且非空。
- 浏览器验收会确认概览页包含 7 个带符号的顶部指标卡、5 个带序号/符号/状态徽标的真实测试闭环项、1 个下一步动作卡和 4 个带模块符号与实时徽标的快速入口。
- 浏览器验收会等待管理后台数据加载完成后再进入模块视觉断言；若加载异常，会输出状态、指标区错误、API 响应列表和控制台消息，便于直接定位真实页面问题。
- 无白屏。
- 无明显遮挡。
- 表格、卡片、按钮和输入控件未出现明显溢出或尺寸错位；每个模块视图都必须没有页面级横向滚动。
- 浏览器验收会在 390px 移动宽度确认底部 Dock 可见、桌面侧边栏隐藏，且概览、策略、节点和运维视图均可从 Dock 进入并不产生横向溢出。
- 来源前缀、节点展示名和 Token 前缀正常展示。
- 有团队和成员数据时，浏览器验收会确认团队/成员动作按钮带有符号，且“编辑”入口可打开行内编辑字段。
- 有来源数据时，浏览器验收会确认上游来源编辑、刷新和同步命名动作都带有符号且不溢出。
- 有节点数据时，浏览器验收会确认节点池先展示带地区符号、可用徽标和协议/来源胶囊的地区聚合，进入地区后出现节点卡片。
- 有节点数据时，浏览器验收会确认节点池地区返回和筛选清空控件都带有动作符号。
- 有节点数据时，浏览器验收会确认节点卡片“编辑”入口可打开行内编辑表单。
- 有来源数据时，浏览器验收会确认上游来源“编辑”入口可打开行内编辑字段。
- 有策略数据时，浏览器验收会确认策略“编辑”入口可打开行内编辑字段。
- 有节点数据时，浏览器验收会确认节点池“详情”入口可展开单节点详情。
- 浏览器验收会确认 RFC3339 和 SQLite 时间字符串都按东八区展示。
- 有 Token 流量数据时，浏览器验收会确认流量分区头、卡片、额度使用率进度条、小时/日趋势图摘要头和出口摘要卡片无视觉溢出；没有出口流量时会确认出口摘要分区头和结构化空状态符号、文案、布局不溢出。
- 浏览器验收会确认运维模块 4 个带符号和状态胶囊的操作卡可见，并点击“检查配置”验证结果卡符号、状态胶囊、结构化字段和无视觉溢出。
- 有 Token 数据时，浏览器验收会确认每行都展示带可见标签和单位的自定义续期天数、追加额度 MiB 输入控件、状态控制动作区和带格式画像、地址来源的可复制订阅卡，实际点击复制按钮确认反馈与符号按钮恢复，并确认创建/重置后的订阅结果卡、Token 动作区可见且无溢出或横向滚动。
- 有节点数据时，浏览器验收会展开节点编辑表单确认保存按钮符号化，再展开节点详情、确认 URI 不横向溢出，并实际点击复制 URI 按钮确认反馈。

## 4. 尚未完成

下一步需要继续实现：

- 上游订阅更多结构化格式解析。
- 更多特殊协议 URI 到 sing-box outbound 的转换。
- 更多真实订阅样本和特殊协议兼容性验证。
- 手把手使用说明需要随后续功能持续更新。
- 后续需要把“首次可用性收口”做成管理后台内的初始化向导，减少脚本和页面之间的切换。

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
