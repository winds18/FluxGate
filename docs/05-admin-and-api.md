# 管理后台与 API

## 1. 管理后台页面

### 1.1 Dashboard

展示：

- 今日总流量
- 本月总流量
- 在线/最近活跃用户
- 即将到期 Token
- 超额 Token
- Token 额度使用率进度
- 上游节点数量
- 异常节点数量
- sing-box 当前配置版本
- 最近一次配置发布状态

### 1.2 团队管理

能力：

- 创建团队
- 编辑团队名称和备注
- 启用/禁用团队
- 查看团队成员
- 查看团队总流量
- 配置团队默认策略

### 1.3 用户管理

能力：

- 创建成员
- 绑定团队
- 编辑成员姓名、邮箱、备注和团队绑定
- 启用/禁用成员
- 查看成员 Token
- 查看成员流量
- 查看最近活跃时间
- 后台展示层统一按东八区显示时间，存储和过期判断仍以 UTC 为准

### 1.4 Token 管理

能力：

- 创建 Token
- 设置有效期
- 设置流量额度
- 续期
- 追加额度
- 禁用
- 吊销
- 重置网关连接 UUID
- 复制订阅地址
- 查看订阅访问日志

Token 状态：

```text
active
expired
revoked
over_quota
```

### 1.5 上游来源

能力：

- 添加机场订阅 URL
- 添加手动节点文本
- 自动生成节点名前缀
- 手动修改节点名前缀
- 行内编辑来源名称、类型、URL、前缀、默认标签和同步间隔
- 设置默认标签
- `default_tags` 当前会在节点导入或刷新时同步到对应上游节点
- 手动同步
- 设置同步间隔
- 查看同步结果
- 查看错误信息

### 1.6 节点池

能力：

- 查看节点列表
- 按地区聚合节点，缺失地区信息时归入“其他”
- 按来源、协议、地区、标签筛选
- 启用/禁用节点
- 修改节点名称
- 查看原始节点名和展示节点名
- 查看单节点详情，包括 URI、服务器、端口、Hash、状态和时间信息
- 将手动节点名恢复为自动命名
- 批量打标签
- 查看节点使用流量
- 查看最近检测结果

### 1.7 策略管理

能力：

- 创建团队策略
- 创建用户策略
- 创建 Token 特例策略
- 配置 include tags
- 配置 exclude tags
- 配置可见虚拟节点
- 配置最大节点数
- 修改策略名称、范围、标签限制、虚拟节点限制、最大节点数和启停状态
- `include_tags` 和 `exclude_tags` 当前按 Token > 成员 > 团队优先级限制上游出口，并生成带 `auth_user` 的 sing-box route rule
- `allowed_virtual_nodes` 当前可按虚拟节点名称或 ID 限制可见虚拟节点
- `allowed_virtual_nodes` 和 `max_nodes` 当前按 Token > 成员 > 团队优先级限制可见虚拟节点，并同步影响 sing-box 入站用户分配；没有可用用户的受限入站不会生成

### 1.8 虚拟节点

能力：

- 创建 FluxGate-HK、FluxGate-SG 等展示节点
- 设置入口协议和端口
- 绑定上游标签
- 修改名称、监听端口、标签选择器和启停状态
- `tag_selector` 当前支持按 `include` / `exclude` 标签生成专属 sing-box selector 和 route rule
- 选择策略类型
- 启用/禁用

### 1.9 sing-box 状态

展示：

- 当前配置版本
- 配置 hash
- 最后生成时间
- 最后校验结果
- 最后发布时间
- sing-box 容器状态
- 统计采集状态
- 最近错误

操作：

- 生成候选配置
- 校验配置
- 发布配置
- 回滚上一版
- 重启 sing-box

### 1.10 系统设置

能力：

- 公网订阅域名
- 默认入口协议
- 默认 Token 有效期
- 默认流量额度
- 统计采集间隔
- 备份策略
- 管理员账号管理

## 2. API 设计

### 2.1 健康检查

```text
GET /healthz
GET /readyz
```

### 2.2 管理登录

```text
GET  /api/auth/session
POST /api/auth/login
POST /api/auth/logout
```

当前实现使用管理员用户名/密码登录，登录成功后写入 HttpOnly、SameSite=Lax 的签名 Cookie。除以下接口外，`/api/*` 管理接口都要求登录：

```text
GET  /healthz
GET  /readyz
GET  /api/auth/session
POST /api/auth/login
POST /api/auth/logout
GET  /sub/{token}
```

`/sub/{token}` 使用订阅 Token 自身鉴权，不依赖管理员会话。

### 2.3 团队

```text
GET    /api/teams
POST   /api/teams
PATCH  /api/teams/{id}
```

### 2.4 用户

```text
GET    /api/users
POST   /api/users
PATCH  /api/users/{id}
```

### 2.5 Token

```text
GET   /api/tokens
POST  /api/tokens
GET   /api/tokens/{id}
PATCH /api/tokens/{id}
POST  /api/tokens/{id}/revoke
POST  /api/tokens/{id}/restore
POST  /api/tokens/{id}/rotate-subscription
POST  /api/tokens/{id}/extend
POST  /api/tokens/{id}/quota
POST  /api/tokens/{id}/rotate-gateway-credential
```

`POST /api/tokens` 创建成功时会返回明文 Token，并在 `subscriptions` 中同时给出 `default`、`clash` 和 `sing_box` 三类可分发订阅地址；兼容字段 `subscription` 等同于 `subscriptions.default`。后台不会把明文 Token 存成裸值，而是用 `TOKEN_SECRET` 派生密钥加密保存订阅密钥，因此 `GET /api/tokens` 可在管理员会话下反复返回可复制的订阅地址。

`POST /api/tokens/{id}/rotate-subscription` 会重新生成订阅密钥和三类订阅地址，旧订阅地址立即失效；适合历史旧 Token 无法恢复订阅地址，或订阅地址疑似泄露时使用。

### 2.6 上游来源

```text
GET   /api/sources
POST  /api/sources
GET   /api/sources/{id}
PATCH /api/sources/{id}
POST  /api/sources/{id}/refresh
POST  /api/sources/{id}/regenerate-node-names
```

### 2.7 节点池

```text
GET   /api/nodes
GET   /api/nodes/{id}
PATCH /api/nodes/{id}
POST  /api/nodes/{id}/reset-display-name
POST  /api/nodes/bulk-tag
POST  /api/nodes/import
```

### 2.8 标签

```text
GET    /api/tags
POST   /api/tags
PATCH  /api/tags/{id}
DELETE /api/tags/{id}
```

### 2.9 策略

```text
GET    /api/policies
POST   /api/policies
PATCH  /api/policies/{id}
```

### 2.10 虚拟节点

```text
GET    /api/virtual-nodes
POST   /api/virtual-nodes
PATCH  /api/virtual-nodes/{id}
```

### 2.11 sing-box 控制

```text
GET  /api/sing-box/status
POST /api/sing-box/config/generate
POST /api/sing-box/config/check
POST /api/sing-box/config/publish
POST /api/sing-box/config/rollback
POST /api/sing-box/restart
```

当前已落地 `config/publish`、`config/rollback` 和 `restart`。发布/回滚响应会返回 `restart_required`；当部署侧显式设置 `SING_BOX_AUTO_RESTART=true` 时，发布/回滚会自动调用重启执行器。默认 driver 为 Docker Unix socket；如需命令模式，可设置 `SING_BOX_RESTART_DRIVER=command` 并配置 `SING_BOX_RESTART_COMMAND` / `SING_BOX_RESTART_ARGS`。

### 2.12 统计

```text
GET /api/traffic/tokens
GET /api/traffic/daily?days=14
GET /api/traffic/hourly?hours=24
GET /api/traffic/outbounds?days=14

GET /api/stats/overview
GET /api/stats/users
GET /api/stats/users/{id}
GET /api/stats/tokens/{id}
GET /api/stats/outbounds
GET /api/stats/traffic/daily
GET /api/stats/traffic/hourly
```

### 2.13 订阅

```text
GET /sub/{token}
GET /sub/{token}?target=clash
GET /sub/{token}?target=sing-box
```

订阅接口必须：

- 校验 Token hash。
- 检查 Token 状态。
- 检查过期时间。
- 检查用户和团队状态。
- 成功响应返回 `subscription-userinfo`、`x-fluxgate-used-bytes`、`x-fluxgate-quota-bytes` 和 `x-fluxgate-remaining-bytes` 等流量额度头。
- 根据 target 或 User-Agent 选择格式。
- 记录访问日志。
- 返回流量信息响应头。

## 3. 权限模型

MVP 只有管理员角色：

```text
admin
```

后续可扩展：

```text
owner
operator
viewer
```

权限建议：

- `owner`：所有权限。
- `operator`：用户、Token、节点、策略管理。
- `viewer`：只读统计和状态。

## 4. 后台操作审计

所有写操作记录：

- 谁操作
- 操作类型
- 资源类型
- 资源 ID
- 修改前
- 修改后
- IP
- User-Agent
- 时间

必须审计的动作：

- 创建/禁用/吊销 Token
- 续期和追加额度
- 修改策略
- 导入节点
- 发布 sing-box 配置
- 回滚配置
- 管理员登录失败

## 5. MVP 页面优先级

第一批必须做：

1. 登录
2. Dashboard
3. 用户/团队
4. Token
5. 上游来源
6. 节点池
7. 策略
8. sing-box 状态

第二批再做：

1. 更细统计报表
2. 节点探针
3. 操作审计筛选
4. 多管理员权限
5. 系统设置可视化
