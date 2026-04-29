# Docker 部署方案

## 1. 目标环境

```text
SSH 别名：<REMOTE_HOST>
部署目录：<REMOTE_DIR>
远程仓库：<REMOTE_URL>
```

公开仓库不记录真实服务器别名、主机路径、生产域名、密钥、Token 或部署探测日志；这些值只允许放在本机未跟踪配置或服务器 `.env` 中。

## 2. 服务器目录结构

```text
<REMOTE_DIR>
├── docker-compose.yml
├── .env
├── data/
│   ├── fluxgate.db
│   ├── backups/
│   └── sing-box/
│       ├── config.json
│       ├── config.previous.json
│       └── versions/
├── logs/
│   ├── fluxgate/
│   └── sing-box/
└── Caddyfile 或 nginx.conf
```

## 3. Docker 服务

### fluxgate

职责：

- Go API
- Web 管理后台
- SQLite 访问
- 订阅接口
- sing-box 配置生成
- 统计采集

端口：

```text
18080/tcp  宿主机 HTTP API，仅反代或内网访问，容器内仍为 8080
```

默认绑定 `127.0.0.1`，适合只通过反代访问。需要局域网直接访问管理后台时，在服务器 `.env` 或本机未跟踪部署配置中设置：

```text
FLUXGATE_HOST_BIND=0.0.0.0
FLUXGATE_HTTP_PORT=18080
```

公网环境不要裸露管理后台；如果必须开放端口，应配合防火墙、访问控制和 HTTPS 反代。

挂载：

```text
./data:/app/data
./logs/fluxgate:/app/logs
/var/run/docker.sock:/var/run/docker.sock  # 可选，用于重启 sing-box
```

### sing-box

职责：

- 数据面入口
- 上游出站
- 统计 API

端口示例：

```text
8443/tcp  VLESS-HK
8444/tcp  VLESS-SG
8445/tcp  VLESS-US
9090/tcp  内部统计/API，不映射公网
```

挂载：

```text
./data/sing-box:/etc/sing-box
./logs/sing-box:/var/log/sing-box
```

### reverse-proxy

可选服务：

- Caddy
- Nginx
- Traefik

职责：

- 管理后台 HTTPS
- 订阅接口 HTTPS
- 可选的 VLESS TLS 分流

MVP 可以先不做复杂分流，代理入口端口独立暴露。

## 4. docker-compose 草案

```yaml
services:
  fluxgate:
    image: ghcr.io/winds18/fluxgate:${FLUXGATE_IMAGE_TAG:-latest}
    container_name: fluxgate
    restart: unless-stopped
    env_file:
      - path: .env
        required: false
    volumes:
      - ./data:/app/data
      - ./logs/fluxgate:/app/logs
      - /var/run/docker.sock:/var/run/docker.sock
    ports:
      - "${FLUXGATE_HOST_BIND:-127.0.0.1}:${FLUXGATE_HTTP_PORT:-18080}:8080"
    depends_on:
      - sing-box

  sing-box:
    image: ghcr.io/sagernet/sing-box:latest
    container_name: fluxgate-sing-box
    restart: unless-stopped
    volumes:
      - ./data/sing-box:/etc/sing-box
      - ./logs/sing-box:/var/log/sing-box
    command: -D /var/lib/sing-box -C /etc/sing-box run
    ports:
      - "8443:8443/tcp"
      - "8444:8444/tcp"
      - "8445:8445/tcp"
    networks:
      - fluxgate

networks:
  fluxgate:
    name: fluxgate
```

注意：

- sing-box 统计 API 不应映射公网端口。
- 如果需要 V2Ray API 统计能力，必须确认镜像包含对应构建标签。
- 如果官方镜像不包含，需要改为自定义 sing-box Dockerfile。
- 部署环境默认拉取远程已构建镜像，不在服务器上散落执行临时构建命令。
- 本地开发只负责验证、测试和生成可复现构建结果。
- 镜像构建优先在远程构建环境或 GitHub Actions 中执行。

## 5. 环境变量

`.env.example` 规划：

```text
APP_ENV=production
HTTP_ADDR=0.0.0.0:8080
PUBLIC_BASE_URL=https://gateway.example.com
FLUXGATE_HOST_BIND=127.0.0.1
FLUXGATE_HTTP_PORT=18080

DB_PATH=/app/data/fluxgate.db

ADMIN_BOOTSTRAP_USERNAME=admin
ADMIN_BOOTSTRAP_PASSWORD=change-me

TOKEN_SECRET=change-me
SESSION_SECRET=change-me

SING_BOX_CONFIG_PATH=/app/data/sing-box/config.json
SING_BOX_PREVIOUS_CONFIG_PATH=/app/data/sing-box/config.previous.json
SING_BOX_CONTAINER_NAME=fluxgate-sing-box
SING_BOX_V2RAY_API_ADDR=sing-box:9090
SING_BOX_API_SECRET=change-me

SOURCE_SYNC_POLL_SECONDS=60
SOURCE_SYNC_BATCH_LIMIT=20
STATS_POLL_INTERVAL_SECONDS=30
BACKUP_RETENTION_DAYS=14

DISK_WARN_PERCENT=80
DISK_CRITICAL_PERCENT=90
DEPLOY_ALLOW_DISK_CLEANUP=true
REMOTE_BUILD_ENABLED=true
GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
```

## 6. 部署脚本要求

部署必须通过脚本执行，不能散落手动执行写操作。

部署脚本至少包含：

```text
scripts/deploy/probe-env.sh
scripts/deploy/bootstrap-remote.sh
scripts/deploy/configure-access.sh
scripts/deploy/cleanup-disk.sh
scripts/deploy/push-and-deploy.sh
scripts/deploy/remote-build.sh
scripts/deploy/deploy-remote.sh
scripts/deploy/verify-remote.sh
scripts/deploy/collect-diagnostics.sh
```

### 6.1 系统环境探测

`probe-env.sh` 必须检测：

- 操作系统和内核版本
- CPU 架构
- Docker 版本
- Docker Compose 版本
- 可用内存
- 磁盘总量、已用量、可用量
- 部署目录是否存在
- 关键端口是否被占用
- 当前容器状态
- 当前镜像和配置版本
- 防火墙或安全组需要开放的端口提示

探测结果必须输出为：

```text
logs/deploy/probe-YYYYMMDD-HHMMSS.log
```

### 6.2 磁盘水位处理

部署前必须检查磁盘水位。

建议阈值：

```text
DISK_WARN_PERCENT=80
DISK_CRITICAL_PERCENT=90
```

处理规则：

- 低于 warning：继续部署。
- 达到 warning：输出警告，清理可安全删除的构建缓存和旧日志。
- 达到 critical：先执行安全清理；清理后仍然超限则停止部署。

允许自动清理：

- 悬空 Docker 镜像
- 停止状态的旧容器
- 过期构建缓存
- 超过保留期的 FluxGate 日志
- 超过保留期的 SQLite 备份

禁止自动清理：

- 当前正在运行的镜像
- 当前配置版本
- 最近一次可回滚配置
- 当前 SQLite 数据库
- 未超过保留期的备份

### 6.3 远程镜像构建

镜像构建可以远程执行。

优先级：

```text
GitHub Actions / 远程构建机
-> 服务器 docker buildx
-> 本地验证构建
```

本地只要求：

- 单元测试通过
- 前端构建通过
- Dockerfile 语法和构建上下文验证
- 生成镜像标签和发布说明

部署服务器默认执行：

```text
docker compose pull
docker compose up -d
```

只有在明确需要服务器构建时，才通过 `scripts/deploy/remote-build.sh` 执行。

FluxGate 当前默认使用远程构建闭环：

```text
scripts/deploy/push-and-deploy.sh
```

该脚本会：

- 推送当前分支到 GitHub。
- SSH 到 `REMOTE_HOST` 指定的服务器。
- 在 `REMOTE_DIR` 指定目录 clone 或更新同名分支。服务器默认使用 `REMOTE_CLONE_URL`，避免依赖服务器 GitHub SSH key。
- 执行远程环境探测。
- 检查并按策略清理磁盘。
- 在远程构建 FluxGate 镜像。
- 启动 Docker Compose。
- 执行远程健康检查。

首次部署时如果还没有 sing-box 配置，`bootstrap-remote.sh` 会生成一个最小可启动配置；后续由 FluxGate 控制面发布正式配置。

## 7. 首次部署流程

```bash
ssh <REMOTE_HOST>
cd <REMOTE_DIR>
scripts/deploy/bootstrap-remote.sh
scripts/deploy/probe-env.sh
scripts/deploy/deploy-remote.sh
scripts/deploy/verify-remote.sh
```

初始化后：

```text
1. 登录管理后台。
2. 修改默认管理员密码。
3. 添加上游来源。
4. 创建虚拟节点。
5. 创建团队和用户。
6. 创建 Token。
7. 生成并发布 sing-box 配置。
8. 复制订阅地址给成员。
```

首次部署前如果目录还不存在，应通过仓库外层的统一 bootstrap 入口完成 clone 和目录创建；该入口也要脚本化，不能散落手动执行写操作。

## 8. 更新流程

```bash
ssh <REMOTE_HOST>
cd <REMOTE_DIR>
git pull
scripts/deploy/push-and-deploy.sh
```

升级前建议：

```text
scripts/db/backup.sh
scripts/sing-box/backup-config.sh
```

## 9. 备份策略

MVP：

- 每天备份 SQLite。
- 每次发布 sing-box 配置前备份上一版。
- 保留最近 14 天备份。

备份目录：

```text
data/backups/
```

后续：

- 增加远程备份。
- 增加一键恢复。
- 增加备份完整性检查。

## 10. 安全建议

- 管理后台必须走 HTTPS。
- 管理后台建议限制 IP 或使用强密码。
- `.env` 权限设为 `600`。
- 不公开 sing-box API 端口。
- 不公开 SQLite 和备份文件。
- 上游订阅 URL 和节点 URI 仅管理员可见。
- 服务器防火墙只开放必要端口。

## 11. 部署可观测性

部署脚本必须保存完整日志链路，便于后续调试。

至少保留：

```text
logs/deploy/
logs/fluxgate/
logs/sing-box/
logs/reverse-proxy/
```

每次部署生成一个 `deploy_id`，并贯穿：

- 环境探测日志
- 磁盘清理日志
- 镜像拉取或构建日志
- compose 更新日志
- 健康检查日志
- 回滚日志

部署失败时，脚本必须自动收集：

- `docker ps`
- `docker compose ps`
- 最近容器日志
- 磁盘状态
- sing-box 配置校验结果
- FluxGate 健康检查结果

## 12. 待确认部署项

需要在正式实现前确认：

- 是否已有域名。
- 是否已有 Caddy/Nginx。
- 管理后台和代理入口是否共用 443。
- 第一版入口端口数量。
- 服务器 CPU 架构和 Docker 版本。
- 官方 sing-box 镜像是否包含所需统计 API。
