#!/bin/bash

# 批量发布所有题目的脚本

echo "🚀 开始批量发布题目..."

# 1. 登录获取 token
echo "📝 Step 1: 登录管理后台..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ 登录失败！请检查后端服务是否运行"
  exit 1
fi

echo "✅ 登录成功"

# 2. 获取所有题目
echo ""
echo "📋 Step 2: 获取所有题目..."
CHALLENGES=$(curl -s "http://localhost:8080/api/admin/challenges" \
  -H "Authorization: Bearer $TOKEN")

# 提取所有题目 ID
CHALLENGE_IDS=$(echo $CHALLENGES | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$CHALLENGE_IDS" ]; then
  echo "❌ 未找到任何题目"
  exit 1
fi

TOTAL=$(echo "$CHALLENGE_IDS" | wc -l | tr -d ' ')
echo "✅ 找到 $TOTAL 个题目"

# 3. 批量更新状态为 published
echo ""
echo "🔄 Step 3: 批量更新题目状态..."
SUCCESS_COUNT=0

for CHALLENGE_ID in $CHALLENGE_IDS; do
  # 更新题目状态为 published
  RESPONSE=$(curl -s -X PUT "http://localhost:8080/api/admin/challenges/${CHALLENGE_ID}/status" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":"published"}')
  
  CODE=$(echo $RESPONSE | grep -o '"code":[0-9]*' | cut -d':' -f2)
  
  if [ "$CODE" = "200" ]; then
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    echo "  ✅ 题目 $CHALLENGE_ID 已发布"
  else
    echo "  ❌ 题目 $CHALLENGE_ID 发布失败"
  fi
done

echo ""
echo "========================================="
echo "🎉 批量发布完成！"
echo "  成功: $SUCCESS_COUNT/$TOTAL"
echo "========================================="
echo ""
echo "💡 现在可以刷新用户前台，查看题目列表"
