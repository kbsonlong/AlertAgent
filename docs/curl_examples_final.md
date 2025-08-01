# AlertAgent API 测试告警生成指南

基于 `swagger.json` API 接口定义的测试告警生成示例。

## 1. 获取认证Token

```bash
# 登录获取token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}'

# 提取token (使用jq)
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' | jq -r '.data.access_token')

# 提取token (不使用jq)
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' | \
    grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

echo "Token: $TOKEN"
```

## 2. 创建测试告警

### CPU使用率告警
```bash
curl -X POST "http://localhost:8080/api/v1/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "CPU使用率过高",
        "title": "服务器CPU使用率告警",
        "content": "服务器CPU使用率持续5分钟超过90%，当前使用率95.6%",
        "source": "node-exporter",
        "level": "critical",
        "severity": "high",
        "status": "new",
        "labels": "{\"instance\": \"server-01\", \"job\": \"node-exporter\", \"cpu\": \"total\"}",
        "fingerprint": "cpu-high-usage-server-01",
        "rule_id": 1,
        "analysis_result": "{}",
        "similar_alerts": "[]"
    }'
```

### 内存不足告警
```bash
curl -X POST "http://localhost:8080/api/v1/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "内存不足",
        "title": "服务器内存不足告警",
        "content": "服务器可用内存低于10%，当前可用内存仅8.5%",
        "source": "node-exporter",
        "level": "high",
        "severity": "medium",
        "status": "new",
        "labels": "{\"instance\": \"server-02\", \"job\": \"node-exporter\", \"memory\": \"available\"}",
        "fingerprint": "memory-low-server-02",
        "rule_id": 1,
        "analysis_result": "{}",
        "similar_alerts": "[]"
    }'
```

### 磁盘空间告警
```bash
curl -X POST "http://localhost:8080/api/v1/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "磁盘空间不足",
        "title": "服务器磁盘空间告警",
        "content": "服务器/var分区磁盘使用率超过85%，当前使用率88.3%",
        "source": "node-exporter",
        "level": "medium",
        "severity": "medium",
        "status": "new",
        "labels": "{\"instance\": \"server-03\", \"job\": \"node-exporter\", \"device\": \"/dev/sda1\", \"mountpoint\": \"/var\"}",
        "fingerprint": "disk-space-low-server-03",
        "rule_id": 1,
        "analysis_result": "{}",
        "similar_alerts": "[]"
    }'
```

### 服务不可用告警
```bash
curl -X POST "http://localhost:8080/api/v1/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "name": "服务不可用",
        "title": "关键服务下线告警",
        "content": "Web服务无法访问，HTTP状态码500，服务可能已下线",
        "source": "blackbox-exporter",
        "level": "critical",
        "severity": "high",
        "status": "new",
        "labels": "{\"instance\": \"web-server-01\", \"job\": \"blackbox-exporter\", \"service\": \"web-api\"}",
        "fingerprint": "service-down-web-server-01",
        "rule_id": 1,
        "analysis_result": "{}",
        "similar_alerts": "[]"
    }'
```

## 3. 查看告警

### 获取告警列表
```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/alerts"
```

### 获取特定告警详情
```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/alerts/1"
```

### 获取告警统计
```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/alerts/stats"
```

## 4. 告警操作

### 分析告警
```bash
curl -X POST "http://localhost:8080/api/v1/alerts/1/analyze" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json"
```

### 处理告警
```bash
curl -X PUT "http://localhost:8080/api/v1/alerts/1" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "status": "acknowledged",
        "handler": "admin",
        "handle_note": "正在处理中"
    }'
```

### 解决告警
```bash
curl -X PUT "http://localhost:8080/api/v1/alerts/1" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "status": "resolved",
        "handler": "admin",
        "handle_note": "问题已解决"
    }'
```

### 查找相似告警
```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/alerts/1/similar"
```

## 5. 字段说明

根据 `swagger.json` 中的 `model.Alert` 定义，创建告警时的必需字段：

- **name**: 告警名称 (必需)
- **title**: 告警标题 (必需)
- **content**: 告警内容 (必需)
- **source**: 告警来源 (必需)
- **level**: 告警级别 (critical/high/medium/low)
- **severity**: 严重程度 (high/medium/low)
- **status**: 告警状态 (new/acknowledged/resolved)
- **rule_id**: 规则ID (必需)
- **labels**: 标签 (JSON字符串)
- **fingerprint**: 指纹 (用于去重)
- **analysis_result**: 分析结果 (JSON字符串，可为空对象 "{}")
- **similar_alerts**: 相似告警 (JSON数组字符串，可为空数组 "[]")

## 6. 认证信息

- **用户名**: admin
- **密码**: password
- **API地址**: http://localhost:8080/api/v1

## 7. 一键测试脚本

使用提供的脚本快速生成测试告警：

```bash
# 完整测试脚本
./auth_and_test.sh

# 快速测试脚本
./quick_test_alert.sh
```

这些脚本会自动：
1. 获取认证token
2. 创建多种类型的测试告警
3. 显示创建结果
4. 提供后续操作的curl命令示例