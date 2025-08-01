#!/bin/bash

# AlertAgent 测试告警生成脚本
# 基于 swagger.json API 定义生成符合 model.Alert 结构的测试告警

# 设置API基础URL
API_BASE="http://localhost:8080/api/v1"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AlertAgent 测试告警生成器 ===${NC}"
echo "API地址: ${API_BASE}"

# 检查是否设置了认证令牌
if [ -n "$AUTH_TOKEN" ]; then
    echo -e "${GREEN}✓ 已设置认证令牌${NC}"
else
    echo -e "${YELLOW}⚠ 未设置认证令牌，如果API需要认证可能会失败${NC}"
    echo -e "${YELLOW}  可以通过以下方式设置: export AUTH_TOKEN=your_token${NC}"
fi
echo ""

# 认证令牌（如果需要的话）
AUTH_TOKEN=""

# 函数：创建告警
create_alert() {
    local alert_data="$1"
    local alert_name="$2"
    
    echo -e "${YELLOW}正在创建告警: ${alert_name}${NC}"
    
    # 构建curl命令
    local curl_cmd="curl -s -w \"\\n%{http_code}\" -X POST \"${API_BASE}/alerts\" -H \"Content-Type: application/json\""
    
    # 如果有认证令牌，添加Authorization头
    if [ -n "$AUTH_TOKEN" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: Bearer $AUTH_TOKEN\""
    fi
    
    curl_cmd="$curl_cmd -d '$alert_data'"
    
    # 执行curl命令
    response=$(eval $curl_cmd)
    
    # 分离HTTP状态码和响应体
    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✓ 告警创建成功 (HTTP $http_code)${NC}"
        echo "响应: $response_body"
    else
        echo -e "${RED}✗ 告警创建失败 (HTTP $http_code)${NC}"
        echo "错误响应: $response_body"
        
        # 如果是401错误，提示认证问题
        if [ "$http_code" -eq 401 ]; then
            echo -e "${YELLOW}提示: 可能需要认证令牌，请设置 AUTH_TOKEN 环境变量${NC}"
        fi
    fi
    echo ""
}

# 1. CPU使用率过高告警
echo "=== 1. CPU使用率过高告警 ==="
cpu_alert='{
    "name": "CPU使用率过高",
    "title": "服务器CPU使用率告警",
    "content": "服务器CPU使用率持续5分钟超过90%，当前使用率95.6%",
    "source": "node-exporter",
    "level": "critical",
    "severity": "high",
    "status": "firing",
    "labels": "{\"instance\": \"server-01\", \"job\": \"node-exporter\", \"cpu\": \"total\"}",
    "fingerprint": "cpu-high-usage-server-01",
    "rule_id": 1,
    "group_id": 1
}'
create_alert "$cpu_alert" "CPU使用率过高"

# 2. 内存不足告警
echo "=== 2. 内存不足告警 ==="
memory_alert='{
    "name": "内存不足",
    "title": "服务器内存不足告警",
    "content": "服务器可用内存低于10%，当前可用内存仅8.5%",
    "source": "node-exporter",
    "level": "warning",
    "severity": "medium",
    "status": "firing",
    "labels": "{\"instance\": \"server-02\", \"job\": \"node-exporter\", \"memory\": \"available\"}",
    "fingerprint": "memory-low-server-02",
    "rule_id": 2,
    "group_id": 1
}'
create_alert "$memory_alert" "内存不足"

# 3. 磁盘空间告警
echo "=== 3. 磁盘空间告警 ==="
disk_alert='{
    "name": "磁盘空间不足",
    "title": "磁盘空间使用率告警",
    "content": "根分区使用率超过95%，当前使用率96.8%",
    "source": "node-exporter",
    "level": "critical",
    "severity": "high",
    "status": "firing",
    "labels": "{\"instance\": \"server-03\", \"job\": \"node-exporter\", \"device\": \"/dev/sda1\", \"mountpoint\": \"/\"}",
    "fingerprint": "disk-full-server-03",
    "rule_id": 3,
    "group_id": 2
}'
create_alert "$disk_alert" "磁盘空间不足"

# 4. 网络延迟告警
echo "=== 4. 网络延迟告警 ==="
network_alert='{
    "name": "网络延迟过高",
    "title": "网络连接延迟告警",
    "content": "网络延迟超过阈值，当前延迟150ms",
    "source": "blackbox-exporter",
    "level": "warning",
    "severity": "medium",
    "status": "firing",
    "labels": "{\"instance\": \"api.example.com\", \"job\": \"blackbox\", \"module\": \"http_2xx\"}",
    "fingerprint": "network-latency-api-example",
    "rule_id": 4,
    "group_id": 2
}'
create_alert "$network_alert" "网络延迟过高"

# 5. 服务宕机告警
echo "=== 5. 服务宕机告警 ==="
service_alert='{
    "name": "服务不可用",
    "title": "关键服务宕机告警",
    "content": "Web服务无法访问，连续3次健康检查失败",
    "source": "blackbox-exporter",
    "level": "critical",
    "severity": "critical",
    "status": "firing",
    "labels": "{\"instance\": \"web.example.com\", \"job\": \"blackbox\", \"service\": \"web\"}",
    "fingerprint": "service-down-web-example",
    "rule_id": 5,
    "group_id": 3
}'
create_alert "$service_alert" "服务不可用"

# 6. 数据库连接告警
echo "=== 6. 数据库连接告警 ==="
db_alert='{
    "name": "数据库连接异常",
    "title": "数据库连接池告警",
    "content": "数据库连接池使用率超过90%，当前活跃连接95/100",
    "source": "mysql-exporter",
    "level": "warning",
    "severity": "medium",
    "status": "firing",
    "labels": "{\"instance\": \"mysql-01\", \"job\": \"mysql\", \"database\": \"production\"}",
    "fingerprint": "db-connection-pool-mysql-01",
    "rule_id": 6,
    "group_id": 3
}'
create_alert "$db_alert" "数据库连接异常"

# 7. SSL证书过期告警
echo "=== 7. SSL证书过期告警 ==="
ssl_alert='{
    "name": "SSL证书即将过期",
    "title": "SSL证书过期告警",
    "content": "SSL证书将在7天内过期，请及时更新证书",
    "source": "blackbox-exporter",
    "level": "warning",
    "severity": "medium",
    "status": "firing",
    "labels": "{\"instance\": \"secure.example.com\", \"job\": \"blackbox\", \"cert_type\": \"ssl\"}",
    "fingerprint": "ssl-cert-expiry-secure-example",
    "rule_id": 7,
    "group_id": 4
}'
create_alert "$ssl_alert" "SSL证书即将过期"

# 8. API错误率告警
echo "=== 8. API错误率告警 ==="
api_alert='{
    "name": "API错误率过高",
    "title": "API服务错误率告警",
    "content": "API 5xx错误率超过5%，当前错误率8.2%",
    "source": "prometheus",
    "level": "critical",
    "severity": "high",
    "status": "firing",
    "labels": "{\"instance\": \"api-gateway\", \"job\": \"api\", \"status_code\": \"5xx\"}",
    "fingerprint": "api-error-rate-gateway",
    "rule_id": 8,
    "group_id": 4
}'
create_alert "$api_alert" "API错误率过高"

echo -e "${GREEN}=== 测试告警生成完成 ===${NC}"
echo "您可以通过以下命令查看创建的告警:"
echo "curl -X GET \"${API_BASE}/alerts\" -H \"Content-Type: application/json\""
echo ""
echo "或访问 Web 界面查看告警列表"