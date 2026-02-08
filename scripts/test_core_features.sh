#!/bin/bash

# Cyber Range Core Features - Integration Test Suite
# 用法: ./test_core_features.sh

set -e

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api"
MYSQL_USER="root"
MYSQL_PASS="123456"
MYSQL_DB="cyber_range"
MYSQL_CONTAINER="mysql"  # MySQL容器名称
REDIS_CONTAINER="redis"  # Redis容器名称

echo "========================================="
echo "  Cyber Range 核心功能测试套件"
echo "========================================="
echo ""

# 计数器
PASSED=0
FAILED=0

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

# 检查MySQL（使用docker exec）
if docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" -e "SELECT 1;" > /dev/null 2>&1; then
    test_pass "MySQL连接正常（Docker容器）"
else
    test_fail "MySQL连接失败，请确保Docker容器运行: docker ps | grep mysql"
    exit 1
fi

# 检查Redis（使用docker exec）
if docker exec -i "$REDIS_CONTAINER" redis-cli ping > /dev/null 2>&1; then
    test_pass "Redis连接正常（Docker容器）"
else
    test_fail "Redis连接失败，请确保Docker容器运行: docker ps | grep redis"
    exit 1
fi

# 检查Docker
if docker ps > /dev/null 2>&1; then
    test_pass "Docker连接正常"
else
    test_fail "Docker连接失败"
    exit 1
fi

# 检查服务器是否运行
if curl -s "$API_BASE/challenges" > /dev/null 2>&1; then
    test_pass "API服务器运行正常"
else
    test_fail "API服务器未启动，请先运行: go run cmd/api/main.go"
    exit 1
fi

echo ""

# ========================================
# 初始化测试数据
# ========================================
echo "【阶段1】初始化测试数据"
echo "-----------------------------------"

docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" "$MYSQL_DB" << 'EOF'
DELETE FROM submissions;
DELETE FROM instances;
DELETE FROM challenges;
DELETE FROM users;

INSERT INTO challenges (id, title, description, category, difficulty, image, flag, points, created_at, updated_at)
VALUES 
('test-1', 'Nginx 测试挑战', '集成测试用题目', 'Web', 'Easy', 'nginx:alpine', 'flag{static_hidden}', 100, NOW(), NOW());

INSERT INTO users (id, username, email, password_hash, role, total_points, created_at, updated_at)
VALUES ('user_mock_001', 'test_user', 'test@test.com', 'hash123', 'user', 0, NOW(), NOW());
EOF

test_pass "测试数据初始化完成"
echo ""

# ========================================
# 测试1: 获取题目列表
# ========================================
echo "【测试1】获取题目列表 (GET /api/challenges)"
echo "-----------------------------------"

RESPONSE=$(curl -s "$API_BASE/challenges")
if echo "$RESPONSE" | jq -e '.code == 200' > /dev/null && \
   echo "$RESPONSE" | jq -e '.data | length > 0' > /dev/null; then
    test_pass "题目列表获取成功"
    test_info "题目数量: $(echo "$RESPONSE" | jq '.data | length')"
else
    test_fail "题目列表获取失败"
fi

echo ""

# ========================================
# 测试2: 启动靶机实例
# ========================================
echo "【测试2】启动靶机实例 (POST /api/challenges/test-1/start)"
echo "-----------------------------------"

START_RESPONSE=$(curl -s -X POST "$API_BASE/challenges/test-1/start")
INSTANCE_ID=$(echo "$START_RESPONSE" | jq -r '.data.id')
CONTAINER_ID=$(echo "$START_RESPONSE" | jq -r '.data.container_id')
PORT=$(echo "$START_RESPONSE" | jq -r '.data.port')

if echo "$START_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "实例启动成功"
    test_info "Instance ID: $INSTANCE_ID"
    test_info "Container ID: $CONTAINER_ID"
    test_info "映射端口: $PORT"
    
    # 检查端口范围
    if [ "$PORT" -ge 20000 ] && [ "$PORT" -le 40000 ]; then
        test_pass "端口分配范围正确 (20000-40000)"
    else
        test_fail "端口分配范围错误: $PORT"
    fi
else
    test_fail "实例启动失败: $(echo "$START_RESPONSE" | jq -r '.msg')"
    exit 1
fi

# 等待容器启动
sleep 2

echo ""

# ========================================
# 测试3: 验证容器资源限制
# ========================================
echo "【测试3】验证容器资源限制 (128MB / 0.5 CPU)"
echo "-----------------------------------"

CONTAINER_STATS=$(docker stats --no-stream --format "{{.MemUsage}}" "$CONTAINER_ID")
if echo "$CONTAINER_STATS" | grep -q "128MiB"; then
    test_pass "内存限制正确 (128MB)"
else
    test_fail "内存限制未生效: $CONTAINER_STATS"
fi

echo ""

# ========================================
# 测试4: 配额检查（尝试重复启动）
# ========================================
echo "【测试4】配额检查 (尝试再次启动同一题目)"
echo "-----------------------------------"

QUOTA_RESPONSE=$(curl -s -X POST "$API_BASE/challenges/test-1/start")
if echo "$QUOTA_RESPONSE" | jq -e '.code == 400' > /dev/null && \
   echo "$QUOTA_RESPONSE" | grep -q "quota exceeded"; then
    test_pass "配额限制生效 (max 1 per user)"
else
    test_fail "配额限制未生效"
fi

echo ""

# ========================================
# 测试5: Flag验证（正确Flag）
# ========================================
echo "【测试5】Flag验证 - 提交正确Flag"
echo "-----------------------------------"

# 从Redis获取正确的Flag（使用docker exec）
CORRECT_FLAG=$(docker exec -i "$REDIS_CONTAINER" redis-cli HGET "instance:$INSTANCE_ID" flag)
test_info "从Redis获取Flag: $CORRECT_FLAG"

SUBMIT_RESPONSE=$(curl -s -X POST "$API_BASE/submit" \
    -H "Content-Type: application/json" \
    -d "{\"challenge_id\": \"test-1\", \"flag\": \"$CORRECT_FLAG\"}")

if echo "$SUBMIT_RESPONSE" | jq -e '.data.correct == true' > /dev/null; then
    test_pass "正确Flag验证通过"
    
    # 检查积分是否增加（使用docker exec）
    POINTS=$(docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" -N -B "$MYSQL_DB" \
        -e "SELECT total_points FROM users WHERE id='user_mock_001';")
    if [ "$POINTS" -eq 100 ]; then
        test_pass "积分正确增加 (100分)"
    else
        test_fail "积分未正确增加，当前: $POINTS"
    fi
else
    test_fail "正确Flag验证失败"
fi

echo ""

# ========================================
# 测试6: Flag验证（错误Flag）
# ========================================
echo "【测试6】Flag验证 - 提交错误Flag"
echo "-----------------------------------"

WRONG_SUBMIT=$(curl -s -X POST "$API_BASE/submit" \
    -H "Content-Type: application/json" \
    -d '{"challenge_id": "test-1", "flag": "flag{wrong_answer}"}')

if echo "$WRONG_SUBMIT" | jq -e '.data.correct == false' > /dev/null; then
    test_pass "错误Flag正确拒绝"
else
    test_fail "错误Flag验证逻辑有误"
fi

echo ""

# ========================================
# 测试7: 停止实例
# ========================================
echo "【测试7】停止靶机实例 (POST /api/challenges/test-1/stop)"
echo "-----------------------------------"

STOP_RESPONSE=$(curl -s -X POST "$API_BASE/challenges/test-1/stop")
if echo "$STOP_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    test_pass "实例停止成功"
    
    # 验证容器已删除
    sleep 1
    if ! docker ps | grep -q "$CONTAINER_ID"; then
        test_pass "容器已成功删除"
    else
        test_fail "容器未被删除"
    fi
    
    # 验证Redis已清理（使用docker exec）
    REDIS_CHECK=$(docker exec -i "$REDIS_CONTAINER" redis-cli EXISTS "instance:$INSTANCE_ID")
    if [ "$REDIS_CHECK" -eq 0 ]; then
        test_pass "Redis状态已清理"
    else
        test_fail "Redis状态未清理"
    fi
else
    test_fail "实例停止失败"
fi

echo ""

# ========================================
# 测试8: The Reaper（可选，需要等待）
# ========================================
echo "【测试8】The Reaper自动清理 (跳过，需要1小时等待)"
echo "-----------------------------------"
test_info "此测试需修改config.yaml的ttl_hours为0.017并重启服务器"
test_info "建议手动验证，或查看日志观察Reaper输出"

echo ""

# ========================================
# 测试报告
# ========================================
echo "========================================="
echo "  测试报告"
echo "========================================="
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有核心功能测试通过！${NC}"
    exit 0
else
    echo -e "${RED}⚠️  有 $FAILED 个测试失败，请检查日志${NC}"
    exit 1
fi
