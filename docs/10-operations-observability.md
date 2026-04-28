# 运维与可观测性

## 1. 目标

FluxGate 的开发、部署和调试必须形成自闭环：

```text
计划
-> 实现
-> 本地验证
-> 远程构建
-> 部署前探测
-> 安全部署
-> 自动验收
-> 日志诊断
-> 截图确认
-> 文档同步
-> 提交推送
```

原则：

- 本地主要负责验证，不依赖本地完成生产镜像构建。
- 镜像构建优先远程执行。
- 部署脚本必须先做系统环境探测。
- 磁盘水位异常时必须先处理再部署。
- 所有关键动作必须留日志。
- 管理后台页面需要自行截图验收。
- 测试结束后必须清理临时产物。
- 尽量由 Codex 自主闭环，减少打扰用户。

## 2. 远程构建

推荐链路：

```text
本地测试通过
-> 推送代码
-> GitHub Actions 构建镜像
-> 推送 ghcr.io/winds18/fluxgate
-> 服务器 pull 镜像
-> docker compose up -d
```

备选链路：

```text
本地测试通过
-> ssh 66.10
-> scripts/deploy/remote-build.sh
-> 服务器构建镜像
-> scripts/deploy/deploy-remote.sh
```

本地只做：

- 单元测试
- API 测试
- 前端构建测试
- Dockerfile 和 compose 配置静态校验
- sing-box 配置生成和 check

## 3. 部署前环境探测

部署脚本必须先执行：

```text
scripts/deploy/probe-env.sh
```

探测内容：

- OS 和内核
- CPU 架构
- Docker 和 Compose 版本
- 当前容器状态
- 当前镜像版本
- 当前配置版本
- 磁盘容量和 inode
- 内存
- 端口占用
- 部署目录权限
- 必要文件是否存在
- `.env` 是否存在但不打印敏感值

探测输出：

```text
logs/deploy/probe-{timestamp}.log
logs/deploy/probe-latest.log
```

## 4. 磁盘治理

磁盘阈值：

```text
warning >= 80%
critical >= 90%
```

处理流程：

```text
检查磁盘
-> warning: 安全清理后继续
-> critical: 安全清理后复查
-> 仍 critical: 停止部署并输出诊断
```

允许清理：

- 悬空 Docker 镜像
- 已停止旧容器
- 过期 build cache
- 超过保留期的部署日志
- 超过保留期的旧备份

禁止清理：

- 当前运行镜像
- 当前 SQLite 数据库
- 最近一次数据库备份
- 当前 sing-box 配置
- 最近一次可回滚配置
- 未明确归属 FluxGate 的用户文件

## 5. 日志链路

每次部署生成一个 `deploy_id`。

日志目录：

```text
logs/
├── deploy/
├── fluxgate/
├── sing-box/
├── reverse-proxy/
├── qa/
└── diagnostics/
```

日志要求：

- 脚本日志必须有开始时间、结束时间、耗时和退出码。
- 服务日志必须有 request id。
- 配置发布必须有 config version。
- 统计采集必须记录采集周期和 delta。
- Token 状态变化必须记录原因。
- 部署失败必须自动收集诊断包。

## 6. 诊断采集

失败时执行：

```text
scripts/deploy/collect-diagnostics.sh
```

采集内容：

- Git commit 和分支
- `.env` 键名列表，不包含值
- `docker ps`
- `docker compose ps`
- FluxGate 最近日志
- sing-box 最近日志
- reverse proxy 最近日志
- 磁盘状态
- 内存状态
- sing-box config check 结果
- 健康检查响应

输出目录：

```text
logs/diagnostics/{deploy_id}/
```

## 7. 页面截图验收

涉及管理后台页面时，Codex 可以自行使用浏览器或截图工具验收。

必须检查：

- 页面是否能打开。
- 是否有白屏。
- 是否有文本溢出。
- 是否有元素遮挡。
- 是否有明显布局错位。
- 核心按钮是否可点击。
- 表单是否可提交。
- 错误和成功反馈是否可见。

截图保存：

```text
logs/qa/screenshots/{timestamp}/
```

最终说明需要写：

- 检查了哪些页面。
- 截图保存位置。
- 是否发现视觉问题。
- 是否已修复。

默认情况下，截图属于临时验收产物，测试脚本退出后自动删除。只有设置 `KEEP_ARTIFACTS=true` 时才保留截图路径用于排查。

## 8. 测试后清理

测试退出必须自动清理：

- 临时 SQLite 数据库。
- WAL/SHM 文件。
- 临时构建产物。
- 临时 API 响应。
- 临时截图。
- 本地 QA 服务进程。
- 一次性 QA 日志。

推荐入口：

```text
scripts/qa/local-suite.sh
```

需要保留产物时：

```text
KEEP_ARTIFACTS=true scripts/qa/local-suite.sh
```

即使保留产物，也必须保证：

- Token 已脱敏。
- `.env` 值不落盘。
- 产物目录在 `.gitignore` 范围内。

## 9. 少打扰原则

开发中默认 Codex 自主推进。

不需要打扰用户的情况：

- 常规代码阅读。
- 常规脚本编写。
- 本地测试。
- 远程构建。
- 页面截图。
- 日志排查。
- 文档同步。

需要用户确认的情况：

- 需要敏感信息。
- 需要生产域名或证书决策。
- 需要清理不可确认归属的大文件。
- 需要停机或可能影响现有用户。
- 验收失败且存在多个风险方案。

## 10. 自动化唤醒

长任务使用每分钟唤醒节奏：

```text
检查当前阶段
-> 查看日志
-> 判断是否阻塞
-> 推进下一步
-> 记录进度
```

适合：

- 远程构建等待
- 部署观察
- CI 等待
- 长测试
- 迁移任务

任务完成后必须停止唤醒。
