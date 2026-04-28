# 数据模型

## 1. 设计原则

- 用户身份、订阅 Token 和 sing-box 入站凭证分开保存。
- Token 只保存 hash，明文只展示一次。
- sing-box 用户名使用稳定、可读、不可泄露敏感信息的内部标识。
- 统计数据保留原始采样，同时汇总小时级和天级报表。
- 节点池保留原始 URI，但仅管理员可见。

## 2. 核心表

### admins

管理后台账号。

```text
id
username
password_hash
status
last_login_at
created_at
updated_at
```

### teams

团队。

```text
id
name
description
status
created_at
updated_at
```

### users

团队成员。

```text
id
team_id
name
email
remark
status              # active / disabled
created_at
updated_at
```

### tokens

订阅 Token 和管理状态。

```text
id
user_id
token_hash
token_prefix
name
status              # active / expired / revoked / over_quota
expire_at
quota_bytes
used_upload_bytes
used_download_bytes
last_used_at
created_at
updated_at
revoked_at
```

说明：

- `quota_bytes = 0` 表示不限制。
- `used_upload_bytes + used_download_bytes` 用于判断是否超额。
- Token 到期或超额后，订阅接口和 sing-box 数据面都要生效。

### gateway_accounts

映射到 sing-box inbound user 的真实连接凭证。

```text
id
token_id
protocol            # vless / shadowsocks
auth_user           # sing-box route/auth stats 使用的用户名
uuid                # VLESS
password            # Shadowsocks 可选
status
created_at
updated_at
rotated_at
```

说明：

- 一个 Token 默认对应一个 gateway account。
- 重置连接凭证时更新 `uuid` 或 `password`。
- `auth_user` 应保持稳定，便于历史统计归属。

## 3. 节点池表

### upstream_sources

上游来源。

```text
id
name
type                # subscription / manual
url
raw_content
prefix_mode         # auto / manual
display_prefix
default_tags
refresh_interval_minutes
status              # active / disabled / failed
last_sync_at
last_error
created_at
updated_at
```

### upstream_nodes

解析后的上游节点。

```text
id
source_id
raw_name
display_name
name_mode           # auto / manual
uri
uri_hash
protocol            # vmess / vless / trojan / shadowsocks / hysteria2 / ...
server
server_port
region
status              # active / disabled / failed
last_seen_at
last_checked_at
last_error
created_at
updated_at
```

说明：

- `uri_hash` 用于去重。
- `uri` 是敏感信息，只在管理端显示。
- `raw_name` 保存上游原始节点名，不随前缀变化。
- `display_name` 是 FluxGate 对外展示和订阅输出使用的名称。
- `name_mode = auto` 时，`display_name` 由来源前缀和 `raw_name` 自动生成。
- `name_mode = manual` 时，管理员手动修改的节点名不会被后续同步覆盖。
- 后续可以把 URI 解析后的结构化参数单独拆表或放 JSON 字段。

### 节点命名规则

上游来源默认启用自动前缀：

```text
display_prefix = "[" + source.name + "] "
display_name = display_prefix + raw_name
```

当订阅 URL 或节点名称相同、相似时，仍然通过来源前缀区分：

```text
[机场A] 香港 01
[机场B] 香港 01
```

当多个来源自动生成出相同前缀时，系统保存最终可区分的 `display_prefix`：

```text
[机场A] 香港 01
[机场A-2] 香港 01
```

如果管理员修改来源前缀：

- `prefix_mode` 变为 `manual`。
- 所有 `name_mode = auto` 的节点重新计算 `display_name`。
- 所有 `name_mode = manual` 的节点保持不变。

如果管理员修改单个节点展示名：

- 该节点 `name_mode` 变为 `manual`。
- 后续来源同步、来源前缀变化不会覆盖该名称。

### tags

节点标签。

```text
id
name                # HK / SG / US / Premium / Gaming
color
created_at
updated_at
```

### upstream_node_tags

节点和标签多对多关系。

```text
node_id
tag_id
```

## 4. 虚拟节点和策略表

### virtual_nodes

面向客户端展示的 FluxGate 虚拟节点。

```text
id
name                # FluxGate-HK-01
listen_protocol     # vless / shadowsocks
listen_port
tag_selector        # JSON: include/exclude tags
strategy            # fixed / selector / urltest / fallback
status
created_at
updated_at
```

说明：

- 客户端订阅看到的是 virtual node。
- virtual node 背后可以对应一个或多个 upstream node。

### policies

团队或 Token 可用策略。

```text
id
name
scope_type          # team / user / token
scope_id
include_tags        # JSON array
exclude_tags        # JSON array
allowed_virtual_nodes # JSON array
max_nodes
status
created_at
updated_at
```

策略优先级建议：

```text
token policy > user policy > team policy > default policy
```

### routing_bindings

虚拟节点到上游出站的绑定关系。

```text
id
virtual_node_id
upstream_node_id
weight
priority
status
created_at
updated_at
```

MVP 可以先做固定绑定或按标签生成 selector/urltest。

## 5. 统计表

### traffic_samples

从 sing-box 采集的原始增量样本。

```text
id
sampled_at
metric_type         # user / inbound / outbound
metric_name
upload_bytes_delta
download_bytes_delta
raw_value_upload
raw_value_download
created_at
```

说明：

- sing-box API 可能返回累计值，FluxGate 需要计算 delta。
- 保存 raw value 方便排查重启或计数回绕。

### traffic_user_hourly

用户小时级汇总。

```text
id
hour
user_id
token_id
gateway_account_id
upload_bytes
download_bytes
created_at
updated_at
```

### traffic_user_daily

用户天级汇总。

```text
id
day
user_id
token_id
gateway_account_id
upload_bytes
download_bytes
created_at
updated_at
```

### traffic_outbound_daily

上游出站天级汇总。

```text
id
day
upstream_node_id
outbound_tag
upload_bytes
download_bytes
created_at
updated_at
```

## 6. 日志表

### subscription_access_logs

订阅访问日志。

```text
id
token_id
ip
user_agent
target
status_code
message
response_bytes
created_at
```

### admin_audit_logs

管理后台操作日志。

```text
id
admin_id
action
resource_type
resource_id
before_json
after_json
ip
user_agent
created_at
```

### config_versions

sing-box 配置发布记录。

```text
id
version
config_hash
config_path
status              # generated / checked / published / failed
check_output
published_at
created_at
```

## 7. 初始索引

建议索引：

```text
tokens(token_hash)
tokens(user_id)
tokens(status, expire_at)
gateway_accounts(auth_user)
upstream_nodes(uri_hash)
upstream_nodes(source_id)
traffic_samples(sampled_at)
traffic_user_hourly(hour, token_id)
traffic_user_daily(day, token_id)
subscription_access_logs(token_id, created_at)
admin_audit_logs(admin_id, created_at)
```
