# AlertAgent Docker 部署指南

本文档详细介绍了如何使用Docker容器化部署AlertAgent系统。

## 目录

- [系统架构](#系统架构)
- [容器组件](#容器组件)
- [快速开始](#快速开始)
- [环境配置](#环境配置)
- [部署方式](#部署方式)
- [监控和维护](#监控和维护)
- [故障排查](#故障排查)

## 系统架构

AlertAgent采用微服务架构，包含以下容器组件：

```
┌─────────────────────────────────────────────────────────────┐
│                    AlertAgent 容器架构                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Nginx     │  │AlertAgent   │  │   Worker    │         │
│  │  (反向代理)  │  │    Core     │  │   Cluster   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   MySQL     │  │    Redis    │  │   Sidecar   │         │
│  │   (数据库)   │  │   (缓存)    │  │  (配置同步)  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐                                           │
│  │   Ollama    │                                           │
│  │  (AI服务)   │                                           │
│  └─────────────┘                                           │
└─────────────────────────────────────────────────────────────┘
```

## 容器组件

### 1. AlertAgent Core
- **镜像**: `alertagent-core:latest`
- **功能**: API网关、业务逻辑处理
- **端口**: 8080
- **健康检查**: `/api/v1/health`

### 2. AlertAgent Worker
- **镜像**: `alertagent-worker:latest`
- **功能**: 异步任务处理
- **类型**: AI分析、通知发送、配置同步
- **端口**: 8081 (健康检查)

### 3. AlertAgent Sidecar
- **镜像**: `alertagent-sidecar:latest`
- **功能**: 配置同步到监控系统
- **支持**: Prometheus、Alertmanager、vmalert
- **端口**: 8081 (健康检查)

### 4. 基础服务
- **MySQL 8.0**: 数据持久化
- **Redis 7**: 缓存和消息队列
- **Ollama**: AI分析服务 (可选)
- **Nginx**: 反向代理 (生产环境)

## 快速开始

### 1. 环境准备

确保已安装以下软件：
- Docker 20.10+
- Docker Compose 2.0+

### 2. 克隆项目

```bash
git clone <repository-url>
cd AlertAgent
```

### 3. 开发环境部署

```bash
# 启动开发环境 (仅基础服务)
./scripts/docker-deploy.sh dev up

# 查看服务状态
./scripts/docker-deploy.sh dev status
```

### 4. 完整环境部署

```bash
# 构建镜像
./scripts/docker-build.sh

# 启动完整环境
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 环境配置

### 开发环境 (.env.dev)

```bash
# 复制环境配置
cp .env.example .env.dev

# 编辑配置
vim .env.dev
```

开发环境特点：
- 仅启动MySQL、Redis、Ollama基础服务
- 应用在本地运行，便于调试
- 包含管理工具 (phpMyAdmin, Redis Commander)

### 测试环境 (.env.test)

```bash
cp .env.example .env.test
# 修改为测试环境配置
```

测试环境特点：
- 完整的容器化部署
- 使用测试数据库
- 启用所有监控功能

### 生产环境 (.env.prod)

```bash
cp .env.example .env.prod
# 修改为生产环境配置，注意安全性
```

生产环境特点：
- 高可用配置
- 资源限制和自动扩缩容
- 完整的监控和日志收集
- HTTPS和安全加固

## 部署方式

### 1. 单机部署

适用于开发、测试或小规模生产环境：

```bash
# 开发环境
./scripts/docker-deploy.sh dev up

# 生产环境
./scripts/docker-deploy.sh prod up
```

### 2. 集群部署

适用于大规模生产环境，参考 [Kubernetes部署文档](./KUBERNETES_DEPLOYMENT.md)。

### 3. 自定义部署

```bash
# 使用自定义compose文件
docker-compose -f docker-compose.custom.yml up -d

# 指定环境文件
docker-compose --env-file .env.custom up -d
```

## 监控和维护

### 1. 健康检查

```bash
# 检查所有服务健康状态
./scripts/docker-deploy.sh prod health

# 手动健康检查
curl http://localhost:8080/api/v1/health
```

### 2. 日志管理

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f alertagent-core

# 查看最近100行日志
docker-compose logs --tail=100 alertagent-core
```

### 3. 性能监控

```bash
# 查看容器资源使用
docker stats

# 查看服务状态
docker-compose ps
```

### 4. 数据备份

```bash
# 数据库备份
docker exec alertagent-mysql mysqldump -u root -p alert_agent > backup.sql

# Redis备份
docker exec alertagent-redis redis-cli BGSAVE
```

## 故障排查

### 1. 常见问题

#### 服务启动失败
```bash
# 查看容器日志
docker-compose logs <service-name>

# 检查容器状态
docker-compose ps

# 重启服务
docker-compose restart <service-name>
```

#### 数据库连接失败
```bash
# 检查MySQL容器状态
docker-compose ps mysql

# 检查数据库连接
docker exec -it alertagent-mysql mysql -u root -p

# 查看数据库日志
docker-compose logs mysql
```

#### Redis连接失败
```bash
# 检查Redis容器状态
docker-compose ps redis

# 测试Redis连接
docker exec -it alertagent-redis redis-cli ping

# 查看Redis日志
docker-compose logs redis
```

### 2. 性能问题

#### 内存不足
```bash
# 查看内存使用
docker stats --no-stream

# 调整容器内存限制
# 编辑 docker-compose.yml 中的 deploy.resources.limits.memory
```

#### CPU使用率高
```bash
# 查看CPU使用
docker stats --no-stream

# 调整Worker并发数
# 修改环境变量 AI_WORKER_CONCURRENCY 等
```

### 3. 网络问题

#### 端口冲突
```bash
# 查看端口占用
netstat -tulpn | grep :8080

# 修改端口映射
# 编辑 docker-compose.yml 中的 ports 配置
```

#### 容器间通信失败
```bash
# 检查网络配置
docker network ls
docker network inspect alertagent-network

# 测试容器间连接
docker exec -it alertagent-core ping mysql
```

## 安全建议

### 1. 生产环境安全

- 使用强密码和密钥
- 启用HTTPS
- 限制网络访问
- 定期更新镜像
- 使用非root用户运行容器

### 2. 数据安全

- 定期备份数据
- 加密敏感数据
- 使用安全的存储卷
- 监控数据访问

### 3. 网络安全

- 使用防火墙规则
- 限制容器间通信
- 使用VPN或专用网络
- 监控网络流量

## 扩展和定制

### 1. 添加新的Worker类型

1. 修改 `cmd/worker/main.go`
2. 更新 Docker Compose 配置
3. 重新构建镜像

### 2. 集成新的监控系统

1. 创建新的Sidecar配置
2. 修改 `internal/sidecar/` 代码
3. 更新部署配置

### 3. 自定义配置

1. 修改配置文件模板
2. 更新环境变量
3. 重新部署服务

## 参考资料

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [AlertAgent API文档](./API.md)
- [Kubernetes部署文档](./KUBERNETES_DEPLOYMENT.md)