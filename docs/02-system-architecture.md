# 系统架构

## 1. 总体架构

```text
Client
  |
  | subscription update
  v
FluxGate API/UI
  |
  | generate client profile
  v
Client
  |
  | proxy traffic
  v
sing-box Gateway
  |
  | route by user/policy
  v
Upstream Airport / Proxy Node
  |
  v
Internet
```

控制链路和数据链路分离：

```text
控制链路：管理员/客户端订阅 -> FluxGate
数据链路：客户端真实代理流量 -> sing-box -> 上游节点
```

## 2. 服务组成

### fluxgate-api

职责：

- 提供管理后台 API。
- 提供客户端订阅接口。
- 管理用户、团队、Token、额度、策略。
- 导入和解析上游节点。
- 生成 sing-box 配置。
- 校验配置并发布到数据面。
- 采集流量统计。
- 写入审计日志。

建议技术：

- Go 1.22+
- Fiber 或 Gin
- SQLite
- sqlc 或 GORM，MVP 可先用 GORM 加快开发
- 嵌入式前端静态文件

### fluxgate-web

职责：

- 管理后台 UI。
- 仪表盘。
- 用户/团队/Token 管理。
- 节点池管理。
- 策略管理。
- 统计报表。
- sing-box 状态页。

建议技术：

- React
- Vite
- TypeScript
- shadcn/ui 或轻量自建组件
- Recharts/ECharts 用于统计图表

MVP 可以把构建产物嵌入 Go 服务，部署时只运行一个 `fluxgate` 容器。

### sing-box

职责：

- 暴露 VLESS 或 Shadowsocks 入站。
- 根据 inbound user 识别团队成员。
- 根据 route rules 选择上游 outbound。
- 真实转发流量。
- 通过统计 API 提供流量计数。

关键点：

- sing-box 是数据面，不直接暴露管理 API 到公网。
- 统计 API 只允许 Docker 内网访问。
- 代理入口端口按协议暴露到公网。

### SQLite

职责：

- 保存控制面全部状态。
- 保存用户、Token、节点、策略、统计和审计日志。

部署方式：

- SQLite 文件放在 Docker volume 或宿主机 `data/fluxgate.db`。
- 定时备份到 `data/backups/`。

## 3. 模块划分

```text
cmd/server
  main.go

internal/config
  环境变量与配置加载

internal/http
  路由、中间件、响应封装

internal/auth
  管理后台登录、管理员 session、API token

internal/users
  团队、成员、Token、额度、有效期

internal/nodes
  上游来源、节点解析、节点池、标签

internal/policies
  团队策略、标签权限、出口策略

internal/subscription
  Clash/Mihomo/sing-box 订阅生成

internal/singbox
  sing-box config.json 生成、校验、发布、重启、统计采集

internal/stats
  流量计数、小时/天汇总、额度判断

internal/audit
  操作日志、访问日志

internal/store
  SQLite 访问层

web
  React 管理后台
```

## 4. 关键运行流程

### 4.1 发布 sing-box 配置

```text
管理员修改用户/节点/策略
-> FluxGate 写数据库
-> 生成 config candidate
-> 执行 sing-box check
-> 备份旧配置
-> 写入新配置
-> 重启 sing-box 容器
-> 记录发布版本
```

MVP 采用重启方式，保证实现简单可靠。

后续可优化：

- 拆分多个 sing-box 实例，减少单次重启影响。
- 对 Shadowsocks 使用 SSM API 做部分用户动态管理。
- 探索 sing-box 可用的热更新能力。

### 4.2 用户到期

```text
定时任务扫描即将到期/已到期 Token
-> 标记 expired
-> 重新生成 sing-box 配置
-> 过期用户从 inbound users 移除
-> 重启 sing-box
-> 订阅接口拒绝该 Token
```

请求时也必须实时检查过期时间，不能只依赖定时任务。

### 4.3 用户超额

```text
统计采集任务更新 used_bytes
-> 判断 used_bytes >= quota_bytes
-> 标记 over_quota
-> 重新生成 sing-box 配置
-> 移除或阻断该用户
```

MVP 默认策略：

- 超额后拒绝新连接。
- 管理员可以追加额度或重置周期。

## 5. 网络暴露面

公网暴露：

- 管理后台 HTTPS，建议限制访问 IP 或开启强密码/2FA。
- 客户端订阅 HTTPS。
- sing-box 代理入口端口。

Docker 内网暴露：

- sing-box V2Ray API 或 Clash API。
- FluxGate 调用 sing-box 的内部控制/统计接口。

严禁公网暴露：

- sing-box 统计 API。
- SQLite 文件。
- 原始上游订阅和节点 URI。

## 6. 安全原则

- Token 明文只显示一次。
- 数据库保存 Token hash。
- 客户端连接 UUID 可重置。
- 管理员密码 hash 存储。
- 管理后台操作写审计日志。
- sing-box API 设置 secret，并绑定 Docker 内网或 localhost。
- `.env` 不入 Git。
- 上游节点 URI 视为敏感信息。
- 备份文件需要限制权限。
