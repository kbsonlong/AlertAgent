#!/bin/bash

# AlertAgent 认证和测试告警生成脚本

# 设置API基础URL
API_BASE="http://localhost:8080/api/v1"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AlertAgent 认证和测试告警生成器 ===${NC}"
echo "API地址: ${API_BASE}"
echo ""

# 步骤1: 获取认证token
echo -e "${BLUE}步骤1: 获取认证token${NC}"
login_response=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}')

echo "登录响应: $login_response"

# 尝试从响应中提取token
token=$(echo "$login_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$token" ]; then
    echo -e "${RED}✗ 无法获取认证token${NC}"
    echo "请检查用户名和密码是否正确"
    exit 1
fi

echo -e "${GREEN}✓ 成功获取token: ${token:0:20}...${NC}"
echo ""

# 步骤2: 创建测试告警
echo -e "${BLUE}步骤2: 创建测试告警${NC}"

# 创建CPU告警
echo -e "${YELLOW}正在创建CPU使用率告警...${NC}"
cpu_response=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
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
    }')

http_code=$(echo "$cpu_response" | tail -n 1)
response_body=$(echo "$cpu_response" | sed '$d')

echo "HTTP状态码: $http_code"
echo "响应内容: $response_body"

if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ CPU告警创建成功!${NC}"
else
    echo -e "${RED}✗ CPU告警创建失败${NC}"
fi
echo ""

# 创建内存告警
echo -e "${YELLOW}正在创建内存不足告警...${NC}"
memory_response=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
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
    }')

http_code=$(echo "$memory_response" | tail -n 1)
response_body=$(echo "$memory_response" | sed '$d')

echo "HTTP状态码: $http_code"
echo "响应内容: $response_body"

if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ 内存告警创建成功!${NC}"
else
    echo -e "${RED}✗ 内存告警创建失败${NC}"
fi
echo ""

# 创建磁盘空间告警
echo -e "${YELLOW}正在创建磁盘空间告警...${NC}"
disk_response=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/alerts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
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
    }')

http_code=$(echo "$disk_response" | tail -n 1)
response_body=$(echo "$disk_response" | sed '$d')

echo "HTTP状态码: $http_code"
echo "响应内容: $response_body"

if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ 磁盘告警创建成功!${NC}"
else
    echo -e "${RED}✗ 磁盘告警创建失败${NC}"
fi
echo ""

# 步骤3: 查看创建的告警
echo -e "${BLUE}步骤3: 查看创建的告警${NC}"
echo -e "${YELLOW}正在获取告警列表...${NC}"

list_response=$(curl -s -H "Authorization: Bearer $token" "${API_BASE}/alerts")
echo "告警列表响应: $list_response"
echo ""

echo -e "${GREEN}=== 测试完成 ===${NC}"
echo "您可以使用以下token进行后续API调用:"
echo "export AUTH_TOKEN='$token'"
echo ""
echo "示例命令:"
echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"${API_BASE}/alerts\""
echo "curl -H \"Authorization: Bearer \$AUTH_TOKEN\" \"${API_BASE}/alerts/1\""
echo "curl -X POST -H \"Authorization: Bearer \$AUTH_TOKEN\" -H \"Content-Type: application/json\" \"${API_BASE}/alerts/1/analyze\""