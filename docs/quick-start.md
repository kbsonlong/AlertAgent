# AlertAgent 快速开始指南

本指南将帮助您快速搭建 AlertAgent 的开发环境并开始开发。

## 📋 前置要求

### 必需工具
- **Git** - 版本控制
- **Go 1.21+** - 后端开发语言
- **Node.js 18+** - 前端开发环境
- **npm** - Node.js 包管理器

### 开发环境选择

我们提供两种开发环境：

1. **本地环境** - 适合日常开发，性能更好
2. **Docker 环境** - 适合快速体验，环境隔离

## 🚀 方式一：本地开发环境

### 1. 安装依赖服务

#### macOS (使用 Homebrew)
```bash
# 安装 MySQL
brew install mysql
brew services start mysql

# 安装 Redis
brew install redis
brew services start redis

# 安装 Ollama (可选)
brew install ollama
```

#### Ubuntu/Debian
```bash
# 更新包列表
sudo apt update

# 安装 MySQL
sudo apt install mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql

# 安装 Redis
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis

# 安装 Ollama (可选)
curl -fsSL https://ollama.ai/install.sh | sh
```

#### CentOS/RHEL
```bash
# 安装 MySQL
sudo yum install mysql-server
sudo systemctl start mysqld
sudo systemctl enable mysqld

# 安装 Redis
sudo yum install redis
sudo systemctl start redis
sudo systemctl enable redis

# 安装 Ollama (可选)
curl -fsSL https://ollama.ai/install.sh | sh
```

### 2. 配置数据库

```bash
# 登录 MySQL（首次可能需要重置密码）
mysql -u root -p

# 创建数据库用户（可选）
CREATE USER 'alertagent'@'localhost' IDENTIFIED BY 'alertagent123';
GRANT ALL PRIVILEGES ON alert_agent.* TO 'alertagent'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. 克隆并启动项目

```bash
# 克隆项目
git clone <your-repo-url>
cd AlertAgent

# 检查环境
make check

# 安装开发工具（可选）
make install

# 一键启动开发环境
make dev
```

## 🐳 方式二：Docker 开发环境

### 1. 安装 Docker

#### macOS
下载并安装 [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/)

#### Windows
下载并安装 [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/install/)

#### Linux
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 重新登录或运行
newgrp docker
```

### 2. 启动 Docker 环境

```bash
# 克隆项目
git clone <your-repo-url>
cd AlertAgent

# 一键启动 Docker 开发环境
make docker-dev
```

## 🔧 开发环境验证

启动完成后，访问以下地址验证环境：

- **前端应用**: http://localhost:5173
- **后端 API**: http://localhost:8080
- **API 文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

### Docker 环境额外服务

- **phpMyAdmin**: http://localhost:8081 (用户名: root, 密码: along123)
- **Redis Commander**: http://localhost:8082

## 📝 开发工作流

### 日常开发命令

```bash
# 查看所有可用命令
make help

# 启动开发环境
make dev              # 本地环境
make docker-dev       # Docker 环境

# 停止开发环境
make dev-stop         # 停止本地环境
make docker-dev-stop  # 停止 Docker 环境

# 重启开发环境
make dev-restart      # 重启本地环境
make docker-dev-restart # 重启 Docker 环境

# 查看服务状态
make status

# 查看日志
make logs             # 应用日志
make docker-logs      # Docker 服务日志
```

### 项目管理命令

```bash
# 安装依赖
make deps

# 构建项目
make build

# 运行测试
make test

# 代码检查
make lint

# 清理构建文件
make clean
```

### 代码热重载

开发环境支持代码热重载：

- **后端**: 修改 Go 代码后自动重启
- **前端**: 修改 React 代码后自动刷新浏览器

## 🔍 故障排除

### 常见问题

#### 1. 端口被占用

```bash
# 检查端口占用
lsof -i :8080  # 后端端口
lsof -i :5173  # 前端端口
lsof -i :3306  # MySQL 端口
lsof -i :6379  # Redis 端口

# 杀死占用端口的进程
kill -9 <PID>
```

#### 2. 数据库连接失败

```bash
# 检查 MySQL 服务状态
# macOS
brew services list | grep mysql

# Linux
sudo systemctl status mysql

# 检查配置文件
cat config/config.yaml
```

#### 3. Redis 连接失败

```bash
# 检查 Redis 服务状态
# macOS
brew services list | grep redis

# Linux
sudo systemctl status redis

# 测试 Redis 连接
redis-cli ping
```

#### 4. 前端依赖安装失败

```bash
# 清理 npm 缓存
npm cache clean --force

# 删除 node_modules 重新安装
cd web
rm -rf node_modules package-lock.json
npm install
```

#### 5. Go 模块下载失败

```bash
# 设置 Go 代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct

# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download
```

### 获取帮助

如果遇到其他问题：

1. 查看项目 [Issues](https://github.com/your-repo/issues)
2. 查看详细日志：`make logs`
3. 检查服务状态：`make status`
4. 重启开发环境：`make dev-restart`

## 🎯 下一步

环境搭建完成后，您可以：

1. 查看 [API 文档](./api.md) 了解接口设计
2. 查看 [开发文档](./docs.md) 了解项目架构
3. 运行示例：`./scripts/create_sample_alerts.sh`
4. 开始开发您的功能

## 📚 相关文档

- [API 文档](./api.md)
- [项目文档](./docs.md)
- [cURL 示例](./curl_examples.md)
- [项目结构说明](../README.md)

---

**提示**: 建议使用 Docker 环境进行快速体验，使用本地环境进行日常开发。