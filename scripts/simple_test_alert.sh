#!/bin/bash

# 简单的测试告警生成脚本
# 基于 swagger.json API 定义

# 设置API基础URL
API_BASE="http://localhost:8080/api/v1"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AlertAgent 简单测试告警生成器 ===${NC}"
echo "API地址: ${API_BASE}"
echo ""

# 创建一个简单的测试告警
echo -e "${YELLOW}正在创建测试告警...${NC}"

response=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/alerts" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "CPU使用率过高",
        "title": "服务器CPU使用率告警",
        "content": "服务器CPU使用率持续5分钟超过90%，当前使用率95.6%",
        "source": "node-exporter",
        "level": "critical",
        "severity": "high",
        "status": "firing",
        "labels": "{\"instance\": \"server-01\", \"job\": \"node-exporter\", \"cpu\": \"total\"}",
        "fingerprint": "cpu-high-usage-server-01"
    }')

# 分离HTTP状态码和响应体
http_code=$(echo "$response" | tail -n 1)
response_body=$(echo "$response" | sed '$d')

echo "HTTP状态码: $http_code"
echo "响应内容: $response_body"
echo ""

if [ "$http_code" -eq 201 ] || [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ 告警创建成功!${NC}"
else
    echo -e "${RED}✗ 告警创建失败${NC}"
    
    case $http_code in
        401)
            echo -e "${YELLOW}错误: 需要认证 (401 Unauthorized)${NC}"
            ;;
        400)
            echo -e "${YELLOW}错误: 请求参数错误 (400 Bad Request)${NC}"
            ;;
        500)
            echo -e "${YELLOW}错误: 服务器内部错误 (500 Internal Server Error)${NC}"
            ;;
        *)
            echo -e "${YELLOW}错误: 未知错误 (HTTP $http_code)${NC}"
            ;;
    esac
fi

echo ""
echo "=== 查看所有告警 ==="
echo "执行以下命令查看创建的告警:"
echo "curl -X GET \"${API_BASE}/alerts\" -H \"Content-Type: application/json\""
echo ""

# 尝试获取告警列表
echo -e "${YELLOW}正在获取告警列表...${NC}"
list_response=$(curl -s "${API_BASE}/alerts")
echo "告警列表响应: $list_response"

echo -e "${GREEN}=== 测试完成 ===${NC}"