# 产品概述

## 运维告警管理系统

基于 Go/Gin 后端和 React/TypeScript 前端构建的综合告警管理系统，集成 Ollama 本地知识库实现智能告警分析。

### 核心功能

- **告警管理**: 创建、追踪和管理告警，支持状态流转（新建 → 已确认 → 已解决）
- **规则引擎**: 配置告警规则，支持灵活的触发条件和通知策略
- **通知系统**: 多渠道通知（邮件、短信、webhook），支持模板管理和分组路由
- **AI 分析**: 使用本地 Ollama 模型进行智能告警分析：
  - 告警分析和处理建议
  - 相似告警检测
  - 知识库集成
- **知识库**: 将处理过的告警转换为可复用的知识条目

### 关键组件

- 告警记录，支持严重程度分级（严重、高、中、低）
- 支持变量替换的通知模板
- 针对性告警的通知组
- 基于 Redis 的异步处理队列
- MySQL 数据持久化和 Redis 缓存