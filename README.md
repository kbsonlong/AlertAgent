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

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Ollama

### 安装步骤
1. 克隆项目
```bash
git clone https://github.com/yourusername/alert_agent.git
cd alert_agent
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境
- 复制 `config/config.yaml.example` 到 `config/config.yaml`
- 修改配置文件中的数据库、Redis和Ollama配置

4. 初始化数据库
```bash
mysql -u root -p < scripts/init.sql
```

5. 启动服务
```bash
go run cmd/main.go
```

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