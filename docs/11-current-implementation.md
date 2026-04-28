# 当前实现状态

## 1. 已落地内容

本轮完成 Phase 1 foundation，目标是建立可运行、可测试、可继续扩展的控制面骨架。

已实现：

- Go HTTP API 服务。
- SQLite migration 自动执行。
- 嵌入式 Web 管理后台只读仪表盘。
- 团队、用户、Token 基础创建和列表。
- 上游来源创建和列表。
- 上游来源自动前缀。
- 重复来源名前缀自动编号，例如 `[机场A]`、`[机场A-2]`。
- 节点 URI 批量导入。
- 节点 `raw_name`、`display_name`、`name_mode`。
- 单节点手动改名保护。
- 单节点恢复自动命名。
- 虚拟节点基础创建和列表。
- Clash/Mihomo 订阅生成。
- sing-box 客户端订阅生成。
- sing-box 服务端配置生成骨架。
- Token hash 存储，明文只在创建时返回。
- 订阅请求日志 Token 路径脱敏。
- 结构化 JSON 服务日志。
- 本地验证脚本。
- 远程部署探测脚本。
- 磁盘清理脚本。
- 诊断采集脚本。
- 推送后远程部署脚本。
- 首次远程部署最小 sing-box config bootstrap。
- 远程 Docker build 支持 `GOPROXY`，默认优先使用 `goproxy.cn` 以避开 `proxy.golang.org` 超时。
- 页面截图验收脚本。
- 本地 QA 套件退出自动清理临时产物。

## 2. 当前 API 骨架

已实现的主要接口：

```text
GET  /healthz
GET  /readyz
GET  /api/overview

GET  /api/teams
POST /api/teams

GET  /api/users
POST /api/users

GET  /api/tokens
POST /api/tokens
POST /api/tokens/{id}/revoke

GET   /api/sources
POST  /api/sources
PATCH /api/sources/{id}
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
scripts/qa/smoke.sh
scripts/qa/api-flow.sh
scripts/qa/screenshot.sh
scripts/qa/local-suite.sh
scripts/deploy/probe-env.sh
scripts/deploy/push-and-deploy.sh
DRY_RUN=true scripts/deploy/cleanup-disk.sh
docker compose config
```

页面截图验收：

```text
KEEP_ARTIFACTS=true scripts/qa/screenshot.sh 可保留截图；默认测试退出会自动清理截图。
```

截图结论：

- 页面可打开。
- 无白屏。
- 无明显遮挡。
- 表格和卡片未出现明显溢出。
- 来源前缀、节点展示名和 Token 前缀正常展示。

## 4. 尚未完成

下一步需要继续实现：

- 管理后台登录和管理员权限。
- 管理后台写操作 UI。
- 上游订阅 URL 自动拉取和解析。
- 完整协议 URI 到 sing-box outbound 的转换。
- 策略管理。
- Token 续期、追加额度、恢复。
- 流量统计采集。
- 超额自动阻断。
- sing-box config 发布、check、回滚的 API 集成。
- GitHub Actions 远程镜像构建。
- 远程服务器 `66.10` 实际部署验证。

## 5. 当前注意事项

- `docker compose config` 已支持没有 `.env` 时做静态校验。
- 生产部署仍应由 `scripts/deploy/bootstrap-remote.sh` 生成 `.env` 后再修改密钥。
- 本地 QA 产生的 `data/`、`logs/`、`tmp/` 均被 `.gitignore` 排除，并默认在测试退出时清理。
- Playwright Chromium 已在本机安装一次，后续截图脚本会复用缓存。
