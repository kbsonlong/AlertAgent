# 技术栈

## 后端 (Go)
- **框架**: Gin Web 框架
- **ORM**: GORM 数据库操作
- **数据库**: MySQL 8.0+ 主存储
- **缓存**: Redis 6.0+ 缓存和队列管理
- **日志**: Zap 结构化日志，logrus 备用
- **配置**: 基于 YAML 的配置管理
- **AI 集成**: Ollama 本地大语言模型推理

### 核心依赖
- `github.com/gin-gonic/gin` - HTTP Web 框架
- `gorm.io/gorm` + `gorm.io/driver/mysql` - ORM 和 MySQL 驱动
- `github.com/redis/go-redis/v9` - Redis 客户端
- `go.uber.org/zap` - 结构化日志
- `gopkg.in/yaml.v3` - YAML 配置解析

## 前端 (React/TypeScript)
- **框架**: React 19 + TypeScript
- **UI 组件库**: Ant Design (antd) 5.24+
- **路由**: React Router DOM 7.4+
- **HTTP 客户端**: Axios API 请求
- **构建工具**: Vite 6.2+
- **Markdown**: react-markdown 渲染分析结果

### 核心依赖
- `react` + `react-dom` - React 核心框架
- `antd` + `@ant-design/icons` - UI 组件和图标
- `react-router-dom` - 客户端路由
- `axios` - API 调用的 HTTP 客户端
- `typescript` - 类型安全和开发体验

## 开发命令

### 后端
```bash
# 安装依赖
go mod download

# 运行开发服务器
go run cmd/main.go

# 构建二进制文件
go build -o bin/alert_agent cmd/main.go

# 初始化数据库
mysql -u root -p < scripts/init.sql
```

### 前端
```bash
# 安装依赖
cd web && npm install

# 开发服务器
npm run dev

# 生产构建
npm run build

# 代码检查
npm run lint

# 预览生产构建
npm run preview
```

## 配置文件
- 后端配置: `config/config.yaml`
- 前端环境: `web/.env`
- 数据库初始化: `scripts/init.sql`