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

## 计划部署位置

```text
服务器 SSH 别名：66.10
部署路径：/home/wings/docker/FluxGate
远程仓库：git@github.com:winds18/FluxGate.git
```

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

基础验收：

```bash
scripts/qa/smoke.sh
scripts/qa/api-flow.sh
scripts/qa/screenshot.sh
```

停止本地服务：

```bash
scripts/dev/stop.sh
```
