#!/bin/bash

# Cyber Range 管理员 API 测试脚本
# 用法: ./test_admin_api.sh

set -e

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api/admin"
PASSED=0
FAILED=0

echo "========================================"
echo "  Cyber Range 管理员 API 测试"
echo "========================================"
echo ""

# 测试辅助函数
test_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    PASSED=$((PASSED + 1))
}

test_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    FAILED=$((FAILED + 1))
}

test_info() {
    echo -e "${YELLOW}ℹ INFO${NC}: $1"
}

# ========================================
# 前置检查
# ========================================
echo "【阶段0】前置条件检查"
echo "-----------------------------------"

# 检查服务器是否运行
if curl -s "http://localhost:8080/api/challenges" > /dev/null 2>&1; then
    test_pass "API服务器运行正常"
else
    test_fail "API服务器未启动，请先运行: go run cmd/api/main.go"
    exit 1
fi

echo ""

# ========================================
# 测试1: 管理员登录
# ========================================
echo "【测试1】管理员登录 (POST /api/admin/login)"
echo "-----------------------------------"

LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if echo "$LOGIN_RESPONSE" | jq -e '.code == 200' > /dev/null && [ "$TOKEN" != "null" ]; then
    test_pass "管理员登录成功"
    test_info "Token: ${TOKEN:0:50}..."
else
    test_fail "管理员登录失败"
    test_info "响应: $LOGIN_RESPONSE"
    exit 1
fi

echo ""

# ========================================
# 测试2: 未授权访问（无Token）
# ========================================
echo "【测试2】未授权访问测试（无Token）"
echo "-----------------------------------"

UNAUTH_RESPONSE=$(curl -s -X GET "$API_BASE/challenges")

if echo "$UNAUTH_RESPONSE" | jq -e '.code == 401' > /dev/null; then
    test_pass "未授权访问正确拒绝"
else
    test_fail "未授权访问应该返回 401"
fi

echo ""

# ========================================
# 测试3: 获取题目列表（带认证）
# ========================================
echo "【测试3】获取题目列表 (GET /api/admin/challenges)"
echo "-----------------------------------"

LIST_RESPONSE=$(curl -s -X GET "$API_BASE/challenges?page=1&pageSize=10" \
    -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "题目列表获取成功"
    TOTAL=$(echo "$LIST_RESPONSE" | jq '.data.total')
    PAGE_SIZE=$(echo "$LIST_RESPONSE" | jq '.data.pageSize')
    test_info "总题目数: $TOTAL"
    test_info "每页: $PAGE_SIZE"
else
    test_fail "题目列表获取失败"
fi

echo ""

# ========================================
# 测试4: 创建题目
# ========================================
echo "【测试4】创建题目 (POST /api/admin/challenges)"
echo "-----------------------------------"

CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/challenges" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "自动化测试题目",
        "descriptionHtml": "<p>这是自动化测试创建的题目</p>",
        "hintHtml": "<p>提示：仔细观察</p>",
        "category": "Web",
        "difficulty": "Easy",
        "image": "nginx:alpine",
        "port": 80,
        "flag": "flag{auto_test_123}",
        "points": 200,
        "status": "unpublished"
    }')

NEW_CHALLENGE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id')

if echo "$CREATE_RESPONSE" | jq -e '.code == 200' > /dev/null && [ "$NEW_CHALLENGE_ID" != "null" ]; then
    test_pass "题目创建成功"
    test_info "新题目 ID: $NEW_CHALLENGE_ID"
else
    test_fail "题目创建失败"
    test_info "响应: $CREATE_RESPONSE"
fi

echo ""

# ========================================
# 测试5: 获取单个题目详情
# ========================================
echo "【测试5】获取题目详情 (GET /api/admin/challenges/$NEW_CHALLENGE_ID)"
echo "-----------------------------------"

DETAIL_RESPONSE=$(curl -s -X GET "$API_BASE/challenges/$NEW_CHALLENGE_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$DETAIL_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "题目详情获取成功"
    TITLE=$(echo "$DETAIL_RESPONSE" | jq -r '.data.title')
    test_info "题目标题: $TITLE"
else
    test_fail "题目详情获取失败"
fi

echo ""

# ========================================
# 测试6: 更新题目
# ========================================
echo "【测试6】更新题目 (PUT /api/admin/challenges/$NEW_CHALLENGE_ID)"
echo "-----------------------------------"

UPDATE_RESPONSE=$(curl -s -X PUT "$API_BASE/challenges/$NEW_CHALLENGE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "自动化测试题目（已更新）",
        "descriptionHtml": "<p>更新后的描述</p>",
        "hintHtml": "<p>更新后的提示</p>",
        "category": "Web",
        "difficulty": "Medium",
        "image": "nginx:alpine",
        "port": 80,
        "flag": "flag{updated_test}",
        "points": 300,
        "status": "published"
    }')

if echo "$UPDATE_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "题目更新成功"
else
    test_fail "题目更新失败"
fi

echo ""

# ========================================
# 测试7: 题目搜索
# ========================================
echo "【测试7】题目搜索 (GET /api/admin/challenges?search=自动化)"
echo "-----------------------------------"

SEARCH_RESPONSE=$(curl -s "$API_BASE/challenges?search=自动化" \
    -H "Authorization: Bearer $TOKEN")

if echo "$SEARCH_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    SEARCH_COUNT=$(echo "$SEARCH_RESPONSE" | jq '.data.list | length')
    if [ "$SEARCH_COUNT" -gt 0 ]; then
        test_pass "搜索功能正常（找到 $SEARCH_COUNT 个结果）"
    else
        test_fail "搜索未找到结果"
    fi
else
    test_fail "搜索功能失败"
fi

echo ""

# ========================================
# 测试8: 题目筛选（分类+难度）
# ========================================
echo "【测试8】题目筛选 (GET /api/admin/challenges?category=Web&difficulty=Medium)"
echo "-----------------------------------"

FILTER_RESPONSE=$(curl -s "$API_BASE/challenges?category=Web&difficulty=Medium" \
    -H "Authorization: Bearer $TOKEN")

if echo "$FILTER_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "筛选功能正常"
    FILTER_COUNT=$(echo "$FILTER_RESPONSE" | jq '.data.total')
    test_info "筛选结果: $FILTER_COUNT 个题目"
else
    test_fail "筛选功能失败"
fi

echo ""

# ========================================
# 测试9: 更新题目状态（上架/下架）
# ========================================
echo "【测试9】更新题目状态 (PUT /api/admin/challenges/$NEW_CHALLENGE_ID/status)"
echo "-----------------------------------"

STATUS_RESPONSE=$(curl -s -X PUT "$API_BASE/challenges/$NEW_CHALLENGE_ID/status" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status": "unpublished"}')

if echo "$STATUS_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "题目状态更新成功"
else
    test_fail "题目状态更新失败"
fi

echo ""

# ========================================
# 测试10: 分页测试
# ========================================
echo "【测试10】分页功能 (page=1&pageSize=5)"
echo "-----------------------------------"

PAGE_RESPONSE=$(curl -s "$API_BASE/challenges?page=1&pageSize=5" \
    -H "Authorization: Bearer $TOKEN")

if echo "$PAGE_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    PAGE_SIZE=$(echo "$PAGE_RESPONSE" | jq '.data.pageSize')
    PAGE=$(echo "$PAGE_RESPONSE" | jq '.data.page')
    if [ "$PAGE_SIZE" -eq 5 ] && [ "$PAGE" -eq 1 ]; then
        test_pass "分页功能正常"
    else
        test_fail "分页参数错误"
    fi
else
    test_fail "分页功能失败"
fi

echo ""

# ========================================
# 测试11: 删除题目
# ========================================
echo "【测试11】删除题目 (DELETE /api/admin/challenges/$NEW_CHALLENGE_ID)"
echo "-----------------------------------"

DELETE_RESPONSE=$(curl -s -X DELETE "$API_BASE/challenges/$NEW_CHALLENGE_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$DELETE_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "题目删除成功"
    
    # 验证删除后确实不存在
    VERIFY_RESPONSE=$(curl -s -X GET "$API_BASE/challenges/$NEW_CHALLENGE_ID" \
        -H "Authorization: Bearer $TOKEN")
    if echo "$VERIFY_RESPONSE" | jq -e '.code == 404' > /dev/null; then
        test_pass "删除验证通过（题目不存在）"
    else
        test_fail "删除后题目仍然存在"
    fi
else
    test_fail "题目删除失败"
fi

echo ""

# ========================================
# 测试12: 获取统计数据（如果API已实现）
# ========================================
echo "【测试12】获取总览统计 (GET /api/admin/overview/stats)"
echo "-----------------------------------"

STATS_RESPONSE=$(curl -s -X GET "$API_BASE/overview/stats" \
    -H "Authorization: Bearer $TOKEN")

if echo "$STATS_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "统计数据获取成功"
    test_info "响应: $(echo $STATS_RESPONSE | jq -c '.data')"
else
    test_info "统计API未完全实现（预期行为）"
fi

echo ""

# ========================================
# 测试报告
# ========================================
echo "========================================"
echo "  测试报告"
echo "========================================"
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有管理员 API 测试通过！${NC}"
    exit 0
else
    echo -e "${RED}⚠️  有 $FAILED 个测试失败，请检查日志${NC}"
    exit 1
fi
