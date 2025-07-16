# 项目结构

## 根目录布局
```
alert_agent/
├── cmd/                    # 应用程序入口
├── config/                 # 配置文件
├── internal/              # 私有应用代码
├── web/                   # 前端 React 应用
├── docs/                  # 文档
├── scripts/               # 数据库和工具脚本
└── go.mod/go.sum          # Go 模块文件
```

## 后端结构 (`internal/`)

### 清洁架构模式
- **API 层** (`internal/api/v1/`) - HTTP 处理器和请求/响应逻辑
- **服务层** (`internal/service/`) - 业务逻辑和编排
- **模型层** (`internal/model/`) - 数据模型和验证
- **基础设施** (`internal/pkg/`) - 外部依赖和工具

### 关键目录
```
internal/
├── api/v1/                # REST API 处理器
│   ├── alert.go          # 告警管理端点
│   ├── rule.go           # 规则管理端点
│   └── ...
├── service/              # 业务逻辑服务
│   ├── alert.go          # 告警处理逻辑
│   ├── ollama.go         # AI 分析服务
│   └── ...
├── model/                # 数据模型
│   ├── alert.go          # 告警实体和验证
│   ├── rule.go           # 规则实体
│   └── ...
├── pkg/                  # 共享包
│   ├── database/         # 数据库连接
│   ├── redis/            # Redis 连接
│   ├── logger/           # 日志工具
│   ├── queue/            # 后台任务处理
│   └── types/            # 共享类型定义
├── config/               # 配置管理
└── router/               # 路由注册
```

## 前端结构 (`web/src/`)

### 组件组织
```
web/src/
├── pages/                # 按功能分组的页面组件
│   ├── alerts/          # 告警管理页面
│   ├── knowledge/       # 知识库页面
│   ├── notifications/   # 通知管理
│   └── settings/        # 系统设置
├── services/            # API 服务层
│   ├── alert.ts         # 告警 API 调用
│   ├── knowledge.ts     # 知识库 API 调用
│   └── notification.ts  # 通知 API 调用
├── utils/               # 工具函数
│   ├── requests.ts      # HTTP 客户端配置
│   └── datetime.ts      # 日期格式化工具
├── layouts/             # 布局组件
└── assets/              # 静态资源
```

## 命名约定

### Go 代码
- **文件**: snake_case (例如: `alert_service.go`)
- **类型**: PascalCase (例如: `AlertService`)
- **函数**: 导出用 PascalCase，私有用 camelCase
- **常量**: 导出用 UPPER_SNAKE_CASE 或 PascalCase

### TypeScript/React
- **文件**: 组件用 PascalCase (例如: `AlertList.tsx`)，工具用 camelCase
- **组件**: PascalCase (例如: `AlertList`)
- **变量/函数**: camelCase
- **类型/接口**: PascalCase

## API 结构
- **基础路径**: `/api/v1/`
- **RESTful 端点**: 遵循 REST 约定
- **响应格式**: 统一的 JSON 结构，包含 `code`、`msg`、`data` 字段
- **错误处理**: 标准化错误响应

## 数据库约定
- **表名**: snake_case，复数形式 (例如: `alerts`、`notify_templates`)
- **列名**: snake_case (例如: `created_at`、`notify_count`)
- **外键**: `{table}_id` 格式 (例如: `rule_id`、`template_id`)
- **软删除**: 使用 GORM 的 `DeletedAt` 字段