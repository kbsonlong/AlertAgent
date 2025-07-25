# 运维告警管理系统

## 项目介绍
基于 Gin 和 Vue 的运维告警管理系统，集成 Ollama 本地知识库实现智能告警分析。系统提供告警规则管理、告警记录管理、告警通知管理等功能，并通过 Ollama 实现告警智能分析和处理建议生成。

## 功能特性

### 1. 告警规则管理
- 支持多种告警规则配置
- 灵活的规则触发条件设置
- 规则启用/禁用管理
- 规则执行日志记录

### 2. 告警记录管理
- 告警事件记录与追踪
- 告警状态流转（待处理、处理中、已处理、已关闭）
- 告警处理记录
- 告警统计分析

### 3. 告警通知管理
- 支持多种通知渠道（邮件、短信、webhook等）
- 通知组管理
- 通知策略配置
- 通知发送记录

### 4. 告警通知模板管理
- 自定义模板内容
- 支持模板变量
- 多语言模板支持
- 模板测试功能

### 5. Ollama知识库集成
- 告警智能分析
- 处理建议生成
- 相似告警关联
- 告警趋势分析

## 技术架构

### 后端技术栈
- Gin: Web框架
- GORM: ORM框架
- JWT: 认证授权
- MySQL: 数据存储
- Redis: 缓存服务
- Ollama: 本地AI模型
- Swagger: API文档

### 前端技术栈
- Vue.js: 前端框架
- Element UI: UI组件库
- Axios: HTTP客户端
- Vue Router: 路由管理
- Vuex: 状态管理

## 项目结构
```
alert_agent/
├── cmd/                    # 程序入口
│   └── main.go
├── config/                 # 配置文件
│   └── config.yaml
├── internal/              # 内部代码
│   ├── api/              # API层
│   │   └── v1/
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   ├── repository/      # 数据访问
│   └── pkg/             # 公共组件
├── pkg/                  # 可导出的包
├── docs/                # 文档
└── scripts/             # 脚本
```

## 快速开始

### 🚀 一键启动开发环境

我们提供了两种开发环境启动方式：

#### 方式一：本地环境（推荐用于日常开发）
```bash
# 使用 Makefile（推荐）
make dev

# 或直接运行脚本
./scripts/dev-setup.sh
```

#### 方式二：Docker 环境（推荐用于快速体验）
```bash
# 使用 Makefile（推荐）
make docker-dev

# 或直接运行脚本
./scripts/docker-dev-setup.sh
```

### 📋 环境要求

#### 本地开发环境
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 6.0+
- Ollama（可选，用于AI功能）

#### Docker 开发环境
- Docker
- Docker Compose
- Go 1.21+（用于运行应用）
- Node.js 18+（用于前端开发）

### 🛠️ 详细安装步骤

#### 1. 克隆项目
```bash
git clone https://github.com/yourusername/alert_agent.git
cd alert_agent
```

#### 2. 检查开发环境
```bash
make check
```

#### 3. 安装开发工具（可选）
```bash
make install
```

#### 4. 启动开发环境

**本地环境：**
```bash
make dev
```

**Docker 环境：**
```bash
make docker-dev
```

#### 5. 访问应用
- 前端：http://localhost:5173
- 后端：http://localhost:8080
- API文档：http://localhost:8080/swagger/index.html

### 🔧 开发环境管理

#### 常用命令
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

# 项目管理
make deps             # 安装依赖
make build            # 构建项目
make test             # 运行测试
make lint             # 代码检查
make clean            # 清理构建文件
```

#### 脚本说明
- `scripts/dev-setup.sh` - 本地开发环境启动脚本
- `scripts/dev-stop.sh` - 本地开发环境停止脚本
- `scripts/dev-restart.sh` - 本地开发环境重启脚本
- `scripts/docker-dev-setup.sh` - Docker 开发环境启动脚本
- `scripts/docker-dev-stop.sh` - Docker 开发环境停止脚本
- `docker-compose.dev.yml` - Docker 开发环境配置

### 🐳 Docker 环境特性

Docker 环境包含以下服务：
- **MySQL 8.0** - 数据库服务
- **Redis 7** - 缓存服务
- **Ollama** - AI 模型服务
- **phpMyAdmin** - 数据库管理工具 (http://localhost:8081)
- **Redis Commander** - Redis 管理工具 (http://localhost:8082)

### 🔍 故障排除

#### 端口冲突
如果遇到端口冲突，请检查以下端口是否被占用：
- 3306 (MySQL)
- 6379 (Redis)
- 8080 (后端)
- 5173 (前端)
- 11434 (Ollama)

#### 权限问题
如果脚本无法执行，请添加执行权限：
```bash
chmod +x scripts/*.sh
```

#### 数据库连接问题
检查 `config/config.yaml` 中的数据库配置是否正确。

## API文档
访问 `http://localhost:8080/swagger/index.html` 查看API文档

详细的API接口文档请参考 [API文档](./docs/api.md)

## 开发计划
- [ ] 支持更多告警源接入
- [ ] 告警规则可视化配置
- [ ] 告警处理流程自动化
- [ ] 告警知识库建设
- [ ] 移动端适配

## 贡献指南
欢迎提交 Issue 和 Pull Request

## 许可证
MIT License