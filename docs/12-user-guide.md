# FluxGate 使用说明

这份说明面向真实体验和日常运维，按“从空实例到可分发订阅”的顺序走。敏感信息只放在服务器 `.env` 或本机未跟踪的 `.env.deploy.local`，不要提交到公开仓库。

## 1. 登录前检查

本地开发：

```bash
scripts/dev/bootstrap.sh
scripts/dev/run.sh
```

远程部署：

```bash
scripts/deploy/push-and-deploy.sh
```

部署后先做只读就绪检查：

```bash
scripts/qa/readiness.sh http://<lan-host>:<port>
scripts/qa/browser-login.sh http://<lan-host>:<port>
```

如果想把缺少体验数据也当作失败：

```bash
STRICT=true scripts/qa/readiness.sh http://<lan-host>:<port>
```

就绪检查只读取健康状态、登录状态、概览、团队、成员、Token、来源、节点、虚拟节点、策略和流量摘要，不会修改线上数据。

如果实例已经有团队、成员、Token 和上游节点，但还缺少虚拟节点、默认策略或已发布的 sing-box 配置，可以执行交付收口脚本：

```bash
scripts/deploy/ensure-usable.sh http://<lan-host>:<port>
STRICT=true scripts/qa/readiness.sh http://<lan-host>:<port>
scripts/qa/usable-probe.sh http://<lan-host>:<port>
scripts/qa/browser-login.sh http://<lan-host>:<port>
```

这个脚本会创建缺失的默认虚拟节点和默认可用策略，校验并发布 sing-box 配置；部署配置里有远程主机信息时，还会重启远程 sing-box 容器让配置生效。

`usable-probe` 是只读验收：它会确认后台已有团队、成员、Token、来源、节点、虚拟节点和策略，检查 sing-box 配置至少包含一个入站、一个可用用户和一个上游出口，并探测当前虚拟节点监听端口是否可从局域网连通。后台会为新创建或已重置订阅的 Token 加密保存订阅密钥，所以 Token 列表可以反复显示和复制订阅地址。历史旧 Token 如果没有加密订阅密钥，列表会提示“旧 Token 无法反复显示”，可以点“重置订阅”生成新地址；重置后旧订阅地址会失效。

如果要连同客户端订阅内容一起验收，可以把后台 Token 列表里复制出来的订阅地址临时传入：

```bash
FLUXGATE_QA_SUBSCRIPTION_URL='http://<lan-host>:<port>/sub/<token>' \
  scripts/qa/usable-probe.sh http://<lan-host>:<port>
```

## 2. 管理员登录

1. 打开管理后台地址。
2. 输入管理员用户名和密码。
3. 登录成功后确认顶部状态为“已连接”。
4. 如果登录后仍停在登录页，先运行：

```bash
scripts/qa/browser-login.sh http://<lan-host>:<port>
scripts/deploy/remote-logs.sh
```

## 3. 添加上游来源

上游来源有两种常用方式：

- `subscription`：粘贴机场订阅 URL，或把订阅内容粘到 raw content。
- `manual`：手动粘贴一批节点 URI。

操作步骤：

1. 在“上游来源”创建来源。
2. 名称建议写机场或供应商名称。
3. 前缀默认自动生成，遇到相同来源名会自动编号；也可以手动修改。
4. 默认标签可写 `HK, Premium` 或 JSON 数组格式。
5. 创建后点“刷新”，观察同步时间、错误列和节点数量变化。

前缀修改后会同步刷新自动命名节点；手动改名过的节点会保留人工名称。

## 4. 检查节点池

节点池默认按地区聚合：

1. 先看地区类别。
2. 点击地区进入节点卡片列表。
3. 点击节点卡片查看详情。
4. 点“编辑”可修改展示名。
5. 点“恢复自动命名”可重新使用来源前缀和原始名称。

地区规则：

- 香港统一归入 `🇨🇳中国|香港`。
- 台湾统一归入 `🇨🇳中国|台湾`。
- 没有地区信息归入“其他”。
- 节点名写成“香港-美国”这类路径时，按目标地区归到美国。

## 5. 创建团队、成员和 Token

推荐顺序：

1. 创建团队。
2. 创建成员并绑定团队。
3. 为成员创建 Token，设置有效期和额度。
4. 复制创建结果里返回的三类订阅地址：
   - 默认订阅
   - Clash/Mihomo 订阅
   - sing-box 订阅

Token 列表里可以继续操作：

- 反复查看和复制订阅地址。
- 自定义续期天数。
- 追加额度 MiB。
- 撤销。
- 恢复。
- 重置订阅地址。

重置订阅地址会让旧订阅地址失效，适合旧 Token 没有可恢复订阅地址、或订阅地址已经泄露的情况。撤销、过期、超额的 Token 不会进入新生成的 sing-box 配置。

## 6. 创建虚拟节点

虚拟节点是团队伙伴最终看到的网关入口，例如：

```text
名称：FluxGate-HK
协议：vless
监听端口：8443
标签选择器：{"include":["HK"]}
策略：selector
状态：active
```

标签选择器支持：

```json
{"include":["HK"],"exclude":["Backup"]}
```

创建后可以在虚拟节点列表行内编辑名称、端口、标签选择器、策略和状态。

## 7. 配置策略

策略可以挂在团队、成员或 Token 上，优先级是：

```text
Token > 成员 > 团队
```

常用字段：

- `include_tags`：允许走哪些上游标签。
- `exclude_tags`：排除哪些上游标签。
- `allowed_virtual_nodes`：允许看到哪些虚拟节点，可写名称或 ID。
- `max_nodes`：最多输出几个虚拟节点。

例子：

```json
include_tags: ["HK"]
exclude_tags: ["Backup"]
allowed_virtual_nodes: ["FluxGate-HK"]
max_nodes: 5
```

策略会同时影响客户端订阅输出和 sing-box 入站用户分配。

## 8. 发布 sing-box 配置

推荐顺序：

1. 点“生成配置”。
2. 点“校验配置”。
3. 确认上游数量、hash 和校验结果。
4. 点“发布配置”。
5. 如果部署侧开启了自动重启，发布后会自动重启 sing-box；否则按页面提示手动重启。

回滚：

1. 点“回滚上一版”。
2. 确认返回的 hash。
3. 必要时重启 sing-box。

脚本方式：

```bash
scripts/sing-box/check-config.sh
scripts/sing-box/publish-config.sh
scripts/sing-box/rollback-config.sh
scripts/sing-box/reload-config.sh
```

## 9. 分发订阅并测试

1. 在 Token 列表中复制订阅地址。
2. 按客户端选择目标：
   - Clash/Mihomo 用 `?target=clash`
   - sing-box 用 `?target=sing-box`
   - 未带 target 使用默认格式
3. 在客户端导入订阅。
4. 确认客户端看到的是 FluxGate 虚拟节点，而不是全部上游机场节点。
5. 连接后访问测试网站。
6. 回到管理后台查看 Token 流量摘要和额度进度。

FluxGate 统计的是经过 sing-box 网关的流量，不做 HTTPS 内容解密，也不审计网页内容。

## 10. 运维与排查

常用命令：

```bash
scripts/deploy/probe-env.sh
scripts/deploy/remote-logs.sh
scripts/deploy/collect-diagnostics.sh
DRY_RUN=true scripts/deploy/cleanup-disk.sh
scripts/qa/public-scan.sh
```

常见问题：

- 页面无法访问：先确认 `FLUXGATE_HOST_BIND=0.0.0.0` 和端口映射。
- 登录失败：确认管理员引导账号只在没有管理员记录时生效。
- 订阅为空：检查 Token 状态、策略限制、虚拟节点状态和上游节点状态。
- 配置校验失败：先看配置校验返回，再采集远程日志。
- 没有流量数据：确认自定义 sing-box 镜像带 `with_v2ray_api`，并配置 `SING_BOX_V2RAY_API_ADDR`。

## 11. 验收清单

一个实例可以交给真实体验前，应至少满足：

- 管理后台能从局域网登录。
- 上游来源能创建、刷新和编辑。
- 节点能按地区聚合、进入卡片、查看详情和编辑名称。
- 团队、成员、Token 能创建和编辑。
- Token 能续期、加额、撤销和恢复。
- 虚拟节点能创建和编辑。
- 策略能创建和编辑，并能限制订阅输出。
- 订阅地址能被客户端下载。
- sing-box 配置能生成、校验、发布和回滚。
- Token 流量摘要和额度进度能展示。

自动化验收：

```bash
scripts/qa/local-suite.sh
scripts/deploy/ensure-usable.sh http://<lan-host>:<port>
scripts/qa/readiness.sh http://<lan-host>:<port>
scripts/qa/usable-probe.sh http://<lan-host>:<port>
scripts/qa/browser-login.sh http://<lan-host>:<port>
```

测试退出后执行：

```bash
scripts/qa/cleanup.sh
```
