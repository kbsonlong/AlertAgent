# Alert Agent API 文档

## 功能概述

Alert Agent 是一个智能告警分析系统，主要功能包括：

1. 告警规则管理
2. 告警记录管理
3. 告警分析（基于 Ollama AI）
4. 异步告警分析
5. 通知模板管理
6. 通知组管理
7. 系统设置

## API 接口

### 1. 告警规则管理

#### 1.1 获取告警规则列表
- **接口**: `GET /api/v1/rules`
- **描述**: 获取所有告警规则
- **响应**: 告警规则列表

#### 1.2 创建告警规则
- **接口**: `POST /api/v1/rules`
- **描述**: 创建新的告警规则
- **请求体**: 告警规则信息
- **响应**: 创建的告警规则

#### 1.3 获取单个告警规则
- **接口**: `GET /api/v1/rules/{id}`
- **描述**: 获取指定ID的告警规则
- **参数**: 
  - `id`: 告警规则ID
- **响应**: 告警规则详情

#### 1.4 更新告警规则
- **接口**: `PUT /api/v1/rules/{id}`
- **描述**: 更新指定ID的告警规则
- **参数**: 
  - `id`: 告警规则ID
- **请求体**: 更新的告警规则信息
- **响应**: 更新后的告警规则

#### 1.5 删除告警规则
- **接口**: `DELETE /api/v1/rules/{id}`
- **描述**: 删除指定ID的告警规则
- **参数**: 
  - `id`: 告警规则ID
- **响应**: 删除结果

### 2. 告警记录管理

#### 2.1 获取告警记录列表
- **接口**: `GET /api/v1/alerts`
- **描述**: 获取所有告警记录
- **响应**: 告警记录列表

#### 2.2 创建告警记录
- **接口**: `POST /api/v1/alerts`
- **描述**: 创建新的告警记录
- **请求体**: 告警记录信息
- **响应**: 创建的告警记录

#### 2.3 获取单个告警记录
- **接口**: `GET /api/v1/alerts/{id}`
- **描述**: 获取指定ID的告警记录
- **参数**: 
  - `id`: 告警记录ID
- **响应**: 告警记录详情

#### 2.4 更新告警记录
- **接口**: `PUT /api/v1/alerts/{id}`
- **描述**: 更新指定ID的告警记录
- **参数**: 
  - `id`: 告警记录ID
- **请求体**: 更新的告警记录信息
- **响应**: 更新后的告警记录

#### 2.5 处理告警
- **接口**: `POST /api/v1/alerts/{id}/handle`
- **描述**: 处理指定ID的告警
- **参数**: 
  - `id`: 告警记录ID
- **请求体**: 处理信息
- **响应**: 处理结果

#### 2.6 查找相似告警
- **接口**: `GET /api/v1/alerts/{id}/similar`
- **描述**: 查找与指定告警相似的告警记录
- **参数**: 
  - `id`: 告警记录ID
- **响应**: 相似告警列表

### 3. 异步告警分析

#### 3.1 异步分析告警
- **接口**: `POST /api/v1/alerts/{id}/async-analyze`
- **描述**: 异步分析指定ID的告警
- **参数**: 
  - `id`: 告警记录ID
- **响应**: 任务ID

#### 3.2 获取分析结果
- **接口**: `GET /api/v1/alerts/{id}/analysis-result`
- **描述**: 获取告警分析结果
- **参数**: 
  - `id`: 告警记录ID
- **响应**: 分析结果

### 4. 通知模板管理

#### 4.1 获取通知模板列表
- **接口**: `GET /api/v1/templates`
- **描述**: 获取所有通知模板
- **响应**: 通知模板列表

#### 4.2 创建通知模板
- **接口**: `POST /api/v1/templates`
- **描述**: 创建新的通知模板
- **请求体**: 通知模板信息
- **响应**: 创建的通知模板

#### 4.3 获取单个通知模板
- **接口**: `GET /api/v1/templates/{id}`
- **描述**: 获取指定ID的通知模板
- **参数**: 
  - `id`: 通知模板ID
- **响应**: 通知模板详情

#### 4.4 更新通知模板
- **接口**: `PUT /api/v1/templates/{id}`
- **描述**: 更新指定ID的通知模板
- **参数**: 
  - `id`: 通知模板ID
- **请求体**: 更新的通知模板信息
- **响应**: 更新后的通知模板

#### 4.5 删除通知模板
- **接口**: `DELETE /api/v1/templates/{id}`
- **描述**: 删除指定ID的通知模板
- **参数**: 
  - `id`: 通知模板ID
- **响应**: 删除结果

### 5. 通知组管理

#### 5.1 获取通知组列表
- **接口**: `GET /api/v1/groups`
- **描述**: 获取所有通知组
- **响应**: 通知组列表

#### 5.2 创建通知组
- **接口**: `POST /api/v1/groups`
- **描述**: 创建新的通知组
- **请求体**: 通知组信息
- **响应**: 创建的通知组

#### 5.3 获取单个通知组
- **接口**: `GET /api/v1/groups/{id}`
- **描述**: 获取指定ID的通知组
- **参数**: 
  - `id`: 通知组ID
- **响应**: 通知组详情

#### 5.4 更新通知组
- **接口**: `PUT /api/v1/groups/{id}`
- **描述**: 更新指定ID的通知组
- **参数**: 
  - `id`: 通知组ID
- **请求体**: 更新的通知组信息
- **响应**: 更新后的通知组

#### 5.5 删除通知组
- **接口**: `DELETE /api/v1/groups/{id}`
- **描述**: 删除指定ID的通知组
- **参数**: 
  - `id`: 通知组ID
- **响应**: 删除结果

### 6. 系统设置

#### 6.1 获取系统设置
- **接口**: `GET /api/v1/settings`
- **描述**: 获取系统设置
- **响应**: 系统设置信息

#### 6.2 更新系统设置
- **接口**: `PUT /api/v1/settings`
- **描述**: 更新系统设置
- **请求体**: 系统设置信息
- **响应**: 更新后的系统设置

## 数据模型

### AlertTask
```go
type AlertTask struct {
    ID        uint      `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}
```

### AlertResult
```go
type AlertResult struct {
    TaskID    uint      `json:"task_id"`
    Status    string    `json:"status"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
}
```

## 配置说明

### 服务器配置
```yaml
server:
  port: 8080
  mode: debug
```

### 数据库配置
```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: password
  dbname: alert_agent
```

### Redis配置
```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

### Ollama配置
```yaml
ollama:
  api_endpoint: http://localhost:11434
  model: llama2
  timeout: 30s
  max_retries: 3
```

## 使用说明

1. 启动服务：
```bash
go run cmd/main.go
```

2. 服务启动时会自动：
   - 加载配置文件
   - 初始化日志系统
   - 连接数据库
   - 连接Redis
   - 处理未分析的告警
   - 启动异步分析工作器
   - 启动HTTP服务器

3. 监控队列状态：
```bash
redis-cli LLEN alert:queue  # 查看队列长度
redis-cli LRANGE alert:queue 0 -1  # 查看队列中的所有任务
```

4. 查看分析结果：
```bash
redis-cli GET alert:result:{task_id}  # 查看特定任务的结果
``` 