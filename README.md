# FluxGate

FluxGate 是一个面向团队使用的代理网关控制面板。

项目定位已经从“机场订阅聚合器”升级为“sing-box 数据面控制平台”：FluxGate 负责用户、Token、有效期、额度、上游节点池、订阅生成、配置生成和流量统计；sing-box 负责真实代理入口、路由转发和流量统计数据来源。

## 核心闭环

```text
管理员导入上游机场/节点
-> FluxGate 解析、去重、打标签、入库
-> 管理员创建团队、成员、Token、有效期和额度
-> FluxGate 生成客户端订阅
-> 客户端连接 FluxGate 网关节点
-> sing-box 根据用户身份转发到上游节点
-> FluxGate 采集 sing-box 统计数据
-> 到期、禁用、超额后自动刷新 sing-box 配置
```

## 架构边界

```text
FluxGate Control Plane
- Web 管理后台
- Go API
- SQLite 数据库
- 上游节点导入
- Token/用户/团队/额度管理
- 客户端订阅生成
- sing-box 配置生成与发布
- 流量统计采集与账单汇总

sing-box Data Plane
- 真实代理入口
- 上游代理出站
- 用户身份识别
- 路由选择
- 真实流量转发
- 流量统计数据源
```

FluxGate 不做真实代理转发，不解密 HTTPS 内容。它可以统计用户、节点和出口维度的流量、连接元数据和订阅访问情况，但不做 MITM 内容审计。

## 部署配置

```text
服务器 SSH 别名：通过本机未跟踪的 .env.deploy.local 配置
部署路径：通过本机未跟踪的 .env.deploy.local 配置
远程仓库：使用当前 Git remote 或显式 REMOTE_URL
```

公开仓库不保存真实服务器别名、主机路径、生产域名、Token、密钥或部署探测日志。

## 文档目录

- [项目总览](docs/00-project-overview.md)
- [产品与运营闭环](docs/01-product-loop.md)
- [系统架构](docs/02-system-architecture.md)
- [数据模型](docs/03-data-model.md)
- [sing-box 集成方案](docs/04-sing-box-integration.md)
- [管理后台与 API](docs/05-admin-and-api.md)
- [Docker 部署方案](docs/06-deployment.md)
- [实施路线图](docs/07-roadmap.md)
- [审阅清单](docs/08-review-checklist.md)
- [开发规范](docs/09-development-standard.md)
- [运维与可观测性](docs/10-operations-observability.md)
- [当前实现状态](docs/11-current-implementation.md)
- [手把手使用说明](docs/12-user-guide.md)

## 本地验证

```bash
scripts/dev/bootstrap.sh
scripts/dev/format.sh
scripts/dev/test.sh
scripts/dev/lint.sh
scripts/dev/build.sh
```

启动本地控制面：

```bash
scripts/dev/run.sh
```

本地开发默认会引导管理员账号：

```text
admin / dev-admin-change-me
```

生产部署应在服务器 `.env` 或本机未跟踪部署配置中设置 `ADMIN_BOOTSTRAP_USERNAME`、`ADMIN_BOOTSTRAP_PASSWORD` 和 `SESSION_SECRET`；公开仓库不保存真实密码。

基础验收：

```bash
scripts/qa/smoke.sh
scripts/qa/api-flow.sh
scripts/qa/readiness.sh
scripts/qa/screenshot.sh
```

推荐使用完整本地验收套件，退出时会自动清理临时数据库、日志、截图和构建产物：

```bash
scripts/qa/local-suite.sh
```

推送或部署前执行公开仓库脱敏检查：

```bash
scripts/qa/public-scan.sh
```

如需保留截图或 API 响应用于排查：

```bash
KEEP_ARTIFACTS=true scripts/qa/local-suite.sh
```

远程实例只读就绪检查：

```bash
scripts/qa/readiness.sh http://<lan-host>:<port>
STRICT=true scripts/qa/readiness.sh http://<lan-host>:<port>
```

停止本地服务：

```bash
scripts/dev/stop.sh
```

默认 Docker Compose 只绑定 `127.0.0.1`。如果需要在局域网访问管理后台，在部署侧未跟踪配置中设置：

```bash
REMOTE_FLUXGATE_HOST_BIND=0.0.0.0
REMOTE_FLUXGATE_HTTP_PORT=18080
```
