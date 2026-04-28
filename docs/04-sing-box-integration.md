# sing-box 集成方案

## 1. 职责边界

sing-box 是 FluxGate 的数据面。FluxGate 不转发真实代理流量，只生成和管理 sing-box 配置，并采集 sing-box 的统计数据。

```text
FluxGate
- 决定谁能用
- 决定能用多久
- 决定能用多少流量
- 决定能走哪些上游节点
- 生成客户端订阅
- 生成 sing-box 配置
- 采集统计并做额度判断

sing-box
- 接收客户端连接
- 识别 inbound user
- 执行路由规则
- 连接上游节点
- 统计真实流量
```

## 2. 入站协议选择

MVP 推荐优先支持 VLESS。

原因：

- sing-box VLESS inbound 支持多用户。
- 每个用户有 `name` 和 `uuid`。
- `name` 可用于 route rule 的 `auth_user` 匹配和统计。
- 客户端兼容性较好。

后续可增加：

- Shadowsocks 2022 multi-user
- Hysteria2
- TUIC
- Trojan

## 3. 用户映射

FluxGate Token 到 sing-box 用户的映射：

```text
tokens.id
  -> gateway_accounts.token_id
  -> gateway_accounts.auth_user
  -> sing-box inbound.users[].name
```

示例：

```json
{
  "type": "vless",
  "tag": "vless-in",
  "listen": "::",
  "listen_port": 443,
  "users": [
    {
      "name": "fg_u_1001_t_2001",
      "uuid": "00000000-0000-0000-0000-000000000000",
      "flow": ""
    }
  ],
  "tls": {},
  "transport": {}
}
```

规则：

- `name` 使用内部稳定标识，不直接使用真实姓名。
- `uuid` 是客户端连接凭证，可以重置。
- Token 失效后，该 user 从配置中移除，或者路由到 block。

## 4. 出站生成

每个上游节点生成一个 outbound。

```text
upstream_nodes.id -> outbound.tag
```

建议 tag 格式：

```text
up_{node_id}_{region}_{short_hash}
```

示例：

```json
{
  "type": "trojan",
  "tag": "up_42_hk_a1b2c3",
  "server": "example.com",
  "server_port": 443,
  "password": "secret",
  "tls": {
    "enabled": true,
    "server_name": "example.com"
  }
}
```

MVP 可先直接从 URI 解析成 sing-box outbound JSON。

## 5. 虚拟节点

客户端看到的不是上游节点，而是 FluxGate 虚拟节点。

示例：

```text
FluxGate-HK
FluxGate-SG
FluxGate-US
FluxGate-Premium
```

每个虚拟节点在订阅中可以体现为一个入口配置，但最终都连接同一个 sing-box 入站端口，区别通过：

- 节点名称
- path/SNI/port
- 或客户端 selector 分组

MVP 简化方案：

- 订阅中返回一组 VLESS 节点。
- 每个节点使用同一个 UUID 和入口域名。
- 节点名称代表策略组。
- 后端通过不同 inbound port 或不同 path 映射到不同 virtual node。

推荐 MVP 采用“不同 listen port 对应不同 virtual node”，实现简单、清晰、易统计。

后续可优化为同端口多 path/SNI 或更复杂的路由识别。

## 6. 路由策略

sing-box route rules 使用 `auth_user` 匹配用户，再选择 outbound。

示例：

```json
{
  "route": {
    "rules": [
      {
        "inbound": ["vless-hk"],
        "auth_user": ["fg_u_1001_t_2001"],
        "action": "route",
        "outbound": "selector_hk_team_a"
      }
    ],
    "final": "block"
  }
}
```

策略生成逻辑：

```text
Token/用户/团队策略
-> 可用 virtual nodes
-> 可用 upstream node tags
-> 生成 selector/urltest/fixed outbound
-> 生成 route rules
```

MVP 推荐策略：

- 每个 virtual node 对应一个 selector 或 urltest outbound。
- 用户不允许的 virtual node 不生成订阅节点。
- 已过期/超额用户不进入 inbound users。

## 7. 流量统计

FluxGate 通过 sing-box 的统计能力采集：

- user 维度
- inbound 维度
- outbound 维度

要求：

- sing-box 构建需要包含 V2Ray API 统计能力。
- V2Ray API 是 gRPC 接口，不是普通 HTTP REST 接口。
- Docker 镜像需要验证是否包含 `with_v2ray_api`。
- 如果官方镜像不满足，项目需要提供自定义 sing-box 镜像。

采集流程：

```text
定时轮询 sing-box stats
-> 读取累计值
-> 和上次 raw value 对比计算 delta
-> 写 traffic_samples
-> 汇总 hourly/daily
-> 更新 tokens.used_upload_bytes / used_download_bytes
-> 判断超额
```

注意：

- sing-box 重启后累计值可能归零。
- 采集器必须处理 raw value 变小的情况。
- 统计接口只绑定 Docker 内网。

## 8. 配置发布

MVP 发布流程：

```text
生成候选 config.json
-> sing-box check
-> 写入 data/sing-box/config.json
-> docker compose restart sing-box
-> 记录 config_versions
```

失败回滚：

```text
新配置 check 失败
-> 不发布
-> 保留旧配置
-> 管理后台显示错误
```

```text
新配置发布后 sing-box 启动失败
-> 回滚到上一版配置
-> restart sing-box
-> 标记发布失败
```

## 9. 端口规划

建议：

```text
443/tcp   管理后台 HTTPS 或 VLESS TLS，二选一或由反代分流
8443/tcp  VLESS-HK
8444/tcp  VLESS-SG
8445/tcp  VLESS-US
8080/tcp  FluxGate API，仅反代访问
9090/tcp  sing-box API，仅 Docker 内网
```

如果使用 Caddy/Nginx 做统一 HTTPS，需要确认代理协议和 WebSocket/gRPC/TLS 终止方式。

MVP 可以先用独立端口降低复杂度。

## 10. 客户端订阅格式

优先支持：

- Clash Meta/Mihomo YAML
- sing-box JSON

后续支持：

- Surge
- Quantumult X
- Shadowrocket

订阅接口：

```text
GET /sub/{token}
GET /sub/{token}?target=clash
GET /sub/{token}?target=sing-box
```

订阅响应头可带流量信息：

```text
subscription-userinfo: upload=...; download=...; total=...; expire=...
```

该流量来自 FluxGate 采集的真实用量，而不是客户端上报。

## 11. 官方参考

- sing-box 配置结构：https://sing-box.sagernet.org/configuration/
- VLESS inbound：https://sing-box.sagernet.org/configuration/inbound/vless/
- Route Rule `auth_user`：https://sing-box.sagernet.org/configuration/route/rule/
- V2Ray API stats：https://sing-box.sagernet.org/configuration/experimental/v2ray-api/
- Docker 部署：https://sing-box.sagernet.org/installation/docker/
