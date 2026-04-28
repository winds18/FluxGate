# 产品与运营闭环

## 1. 角色

### 管理员

负责：

- 添加上游机场订阅和零散节点。
- 管理团队、成员、Token、有效期和流量额度。
- 配置节点标签和团队权限。
- 查看流量统计、访问记录和异常用户。
- 发布 sing-box 配置。

### 团队成员

负责：

- 使用管理员分配的订阅地址。
- 在 Clash Meta/Mihomo、sing-box 等客户端中导入订阅。
- 通过 FluxGate 网关节点访问网络。

### sing-box 数据面

负责：

- 接收客户端代理连接。
- 根据用户身份和路由策略选择上游出口。
- 转发真实流量。
- 提供统计数据。

## 2. 资源闭环

```text
添加上游来源
-> 拉取订阅或粘贴节点 URI
-> 解析节点
-> 生成来源前缀
-> 生成节点展示名称
-> 按 URI hash 去重
-> 标记协议、地区、来源
-> 打标签
-> 保存到节点池
-> 生成 sing-box outbound
```

节点来源分两类：

- `subscription`：机场订阅 URL，支持定期刷新。
- `manual`：手动粘贴的 URI 文本，支持批量导入。

每个节点至少需要这些属性：

- 原始名称
- 展示名称
- 协议
- URI
- 来源
- 标签
- 状态
- 最近同步时间

### 2.1 节点命名前缀

很多上游订阅会使用相同或高度相似的节点名称，例如 `香港 01`、`新加坡 02`。FluxGate 必须在导入时自动给节点展示名加来源前缀，方便管理员和团队成员区分。

默认规则：

```text
节点展示名 = 来源前缀 + 原始节点名
```

示例：

```text
[机场A] 香港 01
[机场B] 香港 01
[备用源] 新加坡 02
```

前缀规则：

- 每个上游来源都有一个 `display_prefix`。
- 默认自动生成前缀，优先使用来源名称。
- 如果来源名称为空，可以从订阅域名或来源编号生成。
- 如果多个来源生成出相同前缀，系统自动追加短编号保证可区分。
- 管理员可以手动修改来源前缀。
- 手动修改前缀后，自动命名的节点展示名随之更新。
- 管理员手动改过单个节点名称后，该节点不再被来源前缀自动覆盖。

节点必须保留两个名称：

- `raw_name`：上游原始节点名。
- `display_name`：FluxGate 管理后台、订阅输出和 sing-box 配置使用的展示名。

## 3. 用户闭环

```text
创建团队
-> 创建成员
-> 创建 Token/UUID
-> 设置有效期
-> 设置流量额度
-> 绑定可用标签或策略
-> 发放订阅地址
```

Token 生命周期：

```text
active -> expired
active -> revoked
active -> over_quota
revoked -> active
expired -> active  # 续期
over_quota -> active  # 追加额度或重置周期
```

关键要求：

- Token 明文只展示一次。
- 数据库只保存 Token hash。
- 面板和日志只展示 Token 前缀。
- Token 到期或禁用后，订阅接口拒绝访问。
- Token 到期或禁用后，sing-box 入站用户也必须被移除或阻断。

## 4. 订阅闭环

```text
用户访问订阅 URL
-> FluxGate 校验 Token
-> 判断客户端目标格式
-> 读取用户可用虚拟节点
-> 生成订阅内容
-> 记录订阅访问日志
```

订阅地址建议：

```text
https://gateway.example.com/sub/{token}
https://gateway.example.com/sub/{token}?target=clash
https://gateway.example.com/sub/{token}?target=sing-box
```

返回的节点不是上游机场原始节点，而是 FluxGate 网关入口，例如：

```text
FluxGate-HK-01
FluxGate-SG-01
FluxGate-US-01
```

这些虚拟节点都指向自己的服务器，由 sing-box 再转发到上游。

## 5. 流量统计闭环

```text
客户端连接 sing-box
-> sing-box 记录 user/inbound/outbound 流量
-> FluxGate 定时采集
-> 写入小时级和天级统计表
-> 更新 Token 已用额度
-> 判断是否超额
-> 超额后刷新 sing-box 配置
```

可以统计：

- 用户上传流量
- 用户下载流量
- Token 总用量
- 虚拟节点用量
- 上游节点/出口用量
- 入站总流量
- 出站总流量
- 最近活跃时间
- 连接目标元数据，取决于 sing-box 日志和嗅探配置

不统计：

- HTTPS 页面正文
- 聊天内容
- 账号密码
- TLS 加密载荷

## 6. 风控闭环

触发条件：

- Token 到期
- 用户被禁用
- Token 被吊销
- 流量超额
- 异常高频连接
- 异常来源 IP

处理动作：

- 拒绝订阅更新。
- 从 sing-box inbound users 中移除。
- 或生成 block 路由策略。
- 重新生成 sing-box 配置。
- 校验配置。
- 重启或刷新 sing-box。
- 记录操作审计日志。

MVP 可以先采用“重新生成配置 + 重启 sing-box”的方式，后续再优化为更细粒度的动态更新。
