# AlertAgent

智能告警管理系统 - 一个基于Vue.js和Go的现代化告警管理平台

## 项目概述

AlertAgent是一个功能完整的告警管理系统，提供告警监控、规则管理、知识库、用户管理等核心功能。系统采用前后端分离架构，前端使用Vue 3 + TypeScript + Ant Design Vue，后端使用Go语言开发。

## 🎉 最新更新 - 渠道管理功能已添加

✅ **渠道管理功能已完成**，包括：
- 渠道管理页面路由配置
- 侧边栏导航菜单项添加
- 通知渠道和通知组管理界面
- 完整的渠道管理功能集成
- 路由映射和导航逻辑完善

✅ **登录功能和路由系统已完成**，包括：
- 完整的登录页面组件
- 路由认证守卫机制
- 自动重定向逻辑
- Vue Router警告修复
- 用户认证状态管理

✅ **后端API集成已完成**，包括：
- 完整的API服务层架构
- 统一的HTTP客户端配置
- 模拟API数据用于开发测试
- 响应式数据展示组件
- 错误处理和用户反馈机制

### 登录功能特性

1. **登录页面** (`/web/src/views/Login.vue`)
   - 用户名/密码登录表单
   - 记住我功能
   - 表单验证和错误提示
   - 响应式设计

2. **路由认证守卫** (`/web/src/router/index.ts`)
   - 自动检查用户认证状态
   - 未认证用户重定向到登录页
   - 已登录用户访问登录页自动跳转首页
   - 页面标题动态设置

3. **认证状态管理**
   - localStorage token存储
   - 自动token验证
   - 401状态自动登出

### API集成特性

1. **统一API客户端** (`/web/src/services/api.ts`)
   - Axios实例配置
   - 请求/响应拦截器
   - 自动token认证
   - 统一错误处理

2. **模块化API服务**
   - `KnowledgeService` - 知识库管理API
   - `AlertService` - 告警管理API
   - `RuleService` - 规则管理API
   - `ProviderService` - 数据源管理API
   - `SystemService` - 系统管理API
   - `UserService` - 用户管理API
   - `NotificationService` - 通知管理API

3. **模拟API数据** (`/web/src/services/mockApi.ts`)
   - 完整的模拟数据响应
   - 真实的数据结构
   - 异步延迟模拟
   - 便于前端开发和测试

4. **API示例组件** (`/web/src/components/ApiExample.vue`)
   - 实时API调用演示
   - 数据展示和交互
   - 错误处理示例
   - 用户友好的反馈机制

## 核心功能

### 1. 告警管理
- 告警列表查看和筛选
- 告警详情展示
- 告警处理和状态更新
- 告警分析和统计
- 告警转知识库功能

### 2. 规则管理
- 告警规则的创建、编辑、删除
- 规则版本控制和历史记录
- 规则测试和验证
- 规则激活/停用管理
- 规则导入导出功能

### 3. 知识库管理
- 知识条目的CRUD操作
- 知识分类和标签管理
- 知识搜索和筛选
- 知识点赞和收藏
- 知识导入导出功能

### 4. 数据源管理
- 多种数据源支持（Prometheus、MySQL、ElasticSearch等）
- 数据源连接测试
- 数据源健康检查
- 指标查询和监控

### 5. 渠道管理
- 通知渠道配置和管理
- 通知组的创建和维护
- 通知模板的设计和编辑
- 渠道测试和验证
- 通知发送历史和统计

### 6. 用户和权限管理
- 用户账户管理
- 用户组和角色管理
- 权限控制和访问管理
- 用户活动日志

### 7. 系统管理
- 系统配置管理
- 系统监控和统计
- 日志管理
- 备份和恢复
- 通知模板管理

## 技术栈

### 前端技术
- **Vue 3**: 使用Composition API和setup语法糖
- **TypeScript**: 提供类型安全和更好的开发体验
- **Ant Design Vue**: 企业级UI组件库
- **Vue Router**: 单页面应用路由管理
- **Pinia**: 现代化的状态管理
- **Axios**: HTTP客户端库
- **Vite**: 快速的构建工具

### 后端技术
- **Go**: 高性能的后端开发语言
- **Gin**: 轻量级Web框架
- **GORM**: ORM数据库操作库
- **MySQL**: 主数据库
- **Redis**: 缓存和会话存储
- **Prometheus**: 监控指标收集

## 项目结构

```
AlertAgent/
├── web/                    # 前端项目目录
│   ├── src/
│   │   ├── components/     # Vue组件
│   │   │   └── ApiExample.vue  # API集成示例组件
│   │   ├── views/         # 页面视图
│   │   ├── services/      # API服务层
│   │   │   ├── api.ts     # 统一API客户端
│   │   │   ├── mockApi.ts # 模拟API数据
│   │   │   ├── knowledgeService.ts
│   │   │   ├── alertService.ts
│   │   │   ├── ruleService.ts
│   │   │   ├── providerService.ts
│   │   │   ├── systemService.ts
│   │   │   ├── userService.ts
│   │   │   ├── notificationService.ts
│   │   │   └── index.ts   # 服务统一导出
│   │   ├── router/        # 路由配置
│   │   ├── stores/        # Pinia状态管理
│   │   ├── utils/         # 工具函数
│   │   └── types/         # TypeScript类型定义
│   ├── public/            # 静态资源
│   └── package.json       # 前端依赖配置
├── cmd/                   # Go应用入口
├── internal/              # 内部包
│   ├── api/              # API路由和处理器
│   ├── service/          # 业务逻辑层
│   ├── model/            # 数据模型
│   └── pkg/              # 公共包
├── configs/              # 配置文件
└── docs/                 # 项目文档
```

## 快速开始

### 前端开发

1. 进入前端目录：
```bash
cd web
```

2. 安装依赖：
```bash
npm install
```

3. 启动开发服务器：
```bash
npm run dev
```

4. 访问应用：
打开浏览器访问 `http://localhost:5174`

### 后端开发

1. 确保Go环境已安装（版本 >= 1.19）

2. 安装依赖：
```bash
go mod tidy
```

3. 启动后端服务：
```bash
go run cmd/main.go
```

4. API服务将在 `http://localhost:8080` 启动

## API文档

系统提供完整的RESTful API，主要包括：

- `/api/v1/auth/*` - 用户认证相关
- `/api/v1/users/*` - 用户管理
- `/api/v1/groups/*` - 用户组管理
- `/api/v1/alerts/*` - 告警管理
- `/api/v1/rules/*` - 规则管理
- `/api/v1/knowledge/*` - 知识库管理
- `/api/v1/providers/*` - 数据源管理
- `/api/v1/system/*` - 系统管理
- `/api/v1/notifications/*` - 通知管理

### API集成使用示例

```typescript
// 导入API服务
import { KnowledgeService, AlertService } from '@/services'

// 获取知识库列表
const knowledgeList = await KnowledgeService.getKnowledgeList({
  page: 1,
  pageSize: 10,
  keyword: '搜索关键词'
})

// 获取告警列表
const alertList = await AlertService.getAlertList({
  page: 1,
  pageSize: 10,
  status: 'active'
})
```

详细的API文档请参考 `/docs/api.md`

## 开发规范

### 前端开发规范

1. **组件命名**: 使用PascalCase命名组件文件
2. **变量命名**: 使用camelCase命名变量和函数
3. **类型定义**: 为所有API响应和组件props定义TypeScript类型
4. **代码注释**: 为复杂逻辑添加详细注释
5. **错误处理**: 统一使用try-catch处理异步操作
6. **API调用**: 使用统一的API服务层，避免直接调用axios

### 后端开发规范

1. **包命名**: 使用小写字母和下划线
2. **函数命名**: 使用驼峰命名法
3. **错误处理**: 统一的错误响应格式
4. **日志记录**: 关键操作添加日志记录
5. **API设计**: 遵循RESTful设计原则

## 部署说明

### 前端部署

1. 构建生产版本：
```bash
npm run build
```

2. 将dist目录部署到Web服务器

### 后端部署

1. 编译Go应用：
```bash
go build -o alertagent cmd/main.go
```

2. 配置环境变量和配置文件

3. 启动应用：
```bash
./alertagent
```

## 开发进度

- [x] 项目初始化和基础架构
- [x] Vue 3 + TypeScript + Ant Design Vue 集成
- [x] Vue Router 路由配置
- [x] 基础组件和页面创建
- [x] API服务层架构设计
- [x] 后端API集成
- [x] 模拟数据和API示例
- [x] 错误处理和用户反馈
- [ ] 真实后端API对接
- [ ] 用户认证和权限管理
- [ ] 数据持久化和状态管理
- [ ] 性能优化和代码分割
- [ ] 单元测试和集成测试
- [ ] 生产环境部署配置

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

本项目采用MIT许可证，详情请参考LICENSE文件。

## 联系方式

如有问题或建议，请通过以下方式联系：

- 项目Issues: [GitHub Issues](https://github.com/your-repo/AlertAgent/issues)
- 邮箱: your-email@example.com

---

**注意**: 这是一个演示项目，用于展示Vue.js和Go的集成开发。当前使用模拟API数据进行前端开发，在生产环境中使用前，请确保进行充分的安全性和性能测试。