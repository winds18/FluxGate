# 当前实现状态

## 1. 已落地内容

本轮完成 Phase 1 foundation，目标是建立可运行、可测试、可继续扩展的控制面骨架。

已实现：

- Go HTTP API 服务。
- SQLite migration 自动执行。
- 嵌入式 Web 管理后台登录页、仪表盘和基础写操作表单。
- 管理员引导账号。
- PBKDF2-SHA256 密码哈希。
- 管理员签名 Cookie 会话。
- 管理 API 登录保护。
- 会话 Cookie 按实际请求协议设置 Secure，支持 HTTPS 反代和 HTTP 局域网调试。
- 团队、用户、Token 基础页面创建和列表。
- Token 支持续期、追加额度、撤销和恢复，并同步 gateway account 状态。
- 上游来源页面创建和列表。
- subscription 类型上游来源可保存 URL 或 raw content，并可手动刷新导入节点。
- subscription 来源刷新时，本次订阅中消失的旧节点会标记为 `inactive`。
- 上游来源自动前缀。
- 重复来源名前缀自动编号，例如 `[机场A]`、`[机场A-2]`。
- 节点 URI 页面批量导入。
- 节点 `raw_name`、`display_name`、`name_mode`。
- 单节点手动改名保护。
- 单节点恢复自动命名。
- 虚拟节点页面基础创建和列表。
- Clash/Mihomo 订阅生成。
- sing-box 客户端订阅生成。
- sing-box 服务端配置生成骨架。
- active VLESS/Trojan/Shadowsocks/VMess/Hysteria2 上游节点会转换为 sing-box outbound，并通过默认 selector 承接网关出口。
- sing-box 服务端配置生成会过滤已撤销、已过期、已超额或 gateway account 不可用的 Token。
- Token hash 存储，明文只在创建时返回。
- 订阅请求日志 Token 路径脱敏。
- 结构化 JSON 服务日志。
- 本地验证脚本。
- 公开仓库脱敏扫描脚本。
- 远程部署探测脚本。
- 磁盘清理脚本。
- 诊断采集脚本。
- 推送后远程部署脚本。
- 远程验收健康检查重试。
- 首次远程部署最小 sing-box config bootstrap。
- 远程 Docker build 支持 `GOPROXY`，默认优先使用 `goproxy.cn` 以避开 `proxy.golang.org` 超时。
- 部署时宿主机 HTTP 端口默认使用 `127.0.0.1:18080`，避免和服务器已有 8080 服务冲突。
- 部署侧可通过未跟踪配置打开局域网访问，不把真实环境信息提交到公开仓库。
- 页面截图验收脚本。
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
POST  /api/nodes/import
PATCH /api/nodes/{id}
POST  /api/nodes/{id}/reset-display-name

GET  /api/virtual-nodes
POST /api/virtual-nodes

POST /api/sing-box/config/generate
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
scripts/qa/browser-login.sh http://<lan-host>:<port> 可做真实浏览器登录验收。
```

截图结论：

- 页面可打开。
- 未登录时展示管理员登录页。
- 登录后管理员登录表单必须不可见，后台视图必须可见，并可加载仪表盘数据。
- 无白屏。
- 无明显遮挡。
- 表格和卡片未出现明显溢出。
- 来源前缀、节点展示名和 Token 前缀正常展示。

## 4. 尚未完成

下一步需要继续实现：

- 上游订阅定时同步。
- 上游订阅更多协议格式解析。
- TUIC 等更多协议 URI 到 sing-box outbound 的转换。
- 策略管理。
- 流量统计采集。
- 流量采集后的超额自动标记、阻断和配置发布触发。
- sing-box config 发布、check、回滚的 API 集成。
- GitHub Actions 远程镜像构建。
- 远程服务器实际部署验证，相关连接信息仅保存在本机未跟踪配置中。

## 5. 当前注意事项

- `docker compose config` 已支持没有 `.env` 时做静态校验。
- 生产部署仍应由 `scripts/deploy/bootstrap-remote.sh` 生成 `.env` 后再修改密钥；管理员引导账号只在没有管理员记录时创建。
- 管理 API 需要管理员会话；订阅接口 `/sub/{token}` 继续使用订阅 Token 鉴权，不依赖管理员登录。
- 需要局域网访问管理后台时，通过部署侧配置 `FLUXGATE_HOST_BIND=0.0.0.0` 和 `FLUXGATE_HTTP_PORT`，公开仓库不保存真实访问地址。
- 本地 QA 产生的 `data/`、`logs/`、`tmp/` 均被 `.gitignore` 排除，并默认在测试退出时清理。
- Playwright Chromium 已在本机安装一次，后续截图脚本会复用缓存。
