# API 文档

本文档详细描述了告警管理系统的API接口。所有API均以JSON格式返回数据，返回格式统一为：

```json
{
  "code": 200,  // 状态码，200表示成功，其他值表示失败
  "msg": "success",  // 状态消息
  "data": {}  // 返回数据，可能是对象、数组或null
}
```

## 目录

- [告警管理](#告警管理)
- [告警规则管理](#告警规则管理)
- [通知模板管理](#通知模板管理)
- [通知组管理](#通知组管理)
- [系统设置](#系统设置)

## 告警管理

### 获取告警列表

```
GET /api/v1/alerts
```

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "created_at": "2023-01-01 12:00:00",
      "updated_at": "2023-01-01 12:00:00",
      "name": "CPU使用率过高",
      "level": "warning",
      "status": "active",
      "source": "prometheus",
      "content": "服务器CPU使用率超过90%",
      "title": "CPU告警",
      "rule_id": 1,
      "notify_count": 0
    }
  ]
}
```

### 创建告警

```
POST /api/v1/alerts
```

**请求参数：**

| 参数名 | 类型 | 必填 | 描述 |
| ------ | ---- | ---- | ---- |
| name | string | 是 | 告警名称 |
| level | string | 是 | 告警级别 |
| source | string | 是 | 告警来源 |
| content | string | 是 | 告警内容 |
| title | string | 是 | 告警标题 |
| rule_id | number | 是 | 关联的规则ID |
| labels | string | 否 | 告警标签，JSON格式 |
| template_id | number | 否 | 通知模板ID |
| group_id | number | 否 | 通知组ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "created_at": "2023-01-01 12:00:00",
    "updated_at": "2023-01-01 12:00:00",
    "name": "CPU使用率过高",
    "level": "warning",
    "status": "active",
    "source": "prometheus",
    "content": "服务器CPU使用率超过90%",
    "title": "CPU告警",
    "rule_id": 1,
    "notify_count": 0
  }
}
```

### 获取单个告警

```
GET /api/v1/alerts/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 告警ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "created_at": "2023-01-01 12:00:00",
    "updated_at": "2023-01-01 12:00:00",
    "name": "CPU使用率过高",
    "level": "warning",
    "status": "active",
    "source": "prometheus",
    "content": "服务器CPU使用率超过90%",
    "title": "CPU告警",
    "rule_id": 1,
    "notify_count": 0
  }
}
```

### 更新告警

```
PUT /api/v1/alerts/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 告警ID |

**请求参数：** 同创建告警

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": null
}
```

### 处理告警

```
PUT /api/v1/alerts/:id/handle
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 告警ID |

**请求参数：**

| 参数名 | 类型 | 必填 | 描述 |
| ------ | ---- | ---- | ---- |
| handler | string | 是 | 处理人 |
| note | string | 是 | 处理备注 |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": null
}
```

### 分析告警

```
POST /api/v1/alerts/:id/analyze
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 告警ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "analysis": "根据告警内容分析，可能是由于应用程序内存泄漏导致CPU使用率过高..."
  }
}
```

### 查找相似告警

```
GET /api/v1/alerts/:id/similar
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 告警ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "alert": {
        "id": 2,
        "name": "CPU使用率过高",
        "level": "warning",
        "status": "handled",
        "source": "prometheus",
        "content": "服务器CPU使用率超过95%"
      },
      "similarity": 0.92
    }
  ]
}
```

## 告警规则管理

### 获取规则列表

```
GET /api/v1/rules
```

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "CPU使用率告警",
      "description": "CPU使用率超过阈值",
      "level": "warning",
      "enabled": true,
      "condition_expr": "cpu_usage > 90",
      "notify_type": "email",
      "notify_group": "运维组",
      "template": "默认邮件模板"
    }
  ]
}
```

### 创建规则

```
POST /api/v1/rules
```

**请求参数：**

| 参数名 | 类型 | 必填 | 描述 |
| ------ | ---- | ---- | ---- |
| name | string | 是 | 规则名称 |
| description | string | 是 | 规则描述 |
| level | string | 是 | 告警级别 |
| enabled | boolean | 是 | 是否启用 |
| condition_expr | string | 是 | 条件表达式 |
| notify_type | string | 是 | 通知类型 |
| notify_group | string | 是 | 通知组 |
| template | string | 是 | 通知模板 |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "CPU使用率告警",
    "description": "CPU使用率超过阈值",
    "level": "warning",
    "enabled": true,
    "condition_expr": "cpu_usage > 90",
    "notify_type": "email",
    "notify_group": "运维组",
    "template": "默认邮件模板"
  }
}
```

### 获取单个规则

```
GET /api/v1/rules/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 规则ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "CPU使用率告警",
    "description": "CPU使用率超过阈值",
    "level": "warning",
    "enabled": true,
    "condition_expr": "cpu_usage > 90",
    "notify_type": "email",
    "notify_group": "运维组",
    "template": "默认邮件模板"
  }
}
```

### 更新规则

```
PUT /api/v1/rules/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 规则ID |

**请求参数：** 同创建规则

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": null
}
```

### 删除规则

```
DELETE /api/v1/rules/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |
| ------ | ---- | ---- |
| id | number | 规则ID |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": null
}
```

## 通知模板管理

### 获取模板列表

```
GET /api/v1/templates
```

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "默认邮件模板",
      "type": "email",
      "content": "告警标题：${title}\n告警级别：${level}\n告警内容：${content}\n发生时间：${time}",
      "description": "默认邮件通知模板",
      "enabled": true
    }
  ]
}
```

### 创建模板

```
POST /api/v1/templates
```

**请求参数：**

| 参数名 | 类型 | 必填 | 描述 |
| ------ | ---- | ---- | ---- |
| name | string | 是 | 模板名称 |
| type | string | 是 | 模板类型 |
| content | string | 是 | 模板内容 |
| description | string | 否 | 模板描述 |
| enabled | boolean | 否 | 是否启用 |

**返回示例：**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "默认邮件模板",
    "type": "email",
    "content": "告警标题：${title}\n告警级别：${level}\n告警内容：${content}\n发生时间：${time}",
    "description": "默认邮件通知模板",
    "enabled": true
  }
}
```

### 获取单个模板

```
GET /api/v1/templates/:id
```

**路径参数：**

| 参数名 | 类型 | 描述 |