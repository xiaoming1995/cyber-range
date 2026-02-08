# Cyber Range API 文档

## 📌 概览

Cyber Range 平台提供RESTful API，用于管理CTF挑战题目、容器实例和Flag验证。所有API均返回标准JSON格式。

**Base URL:** `http://localhost:8080/api`

**认证方式:** 当前版本使用Mock用户ID（`user_mock_001`），未来将实现JWT认证。

---

## 📋 标准响应格式

所有API统一返回以下JSON结构：

```json
{
  "code": 200,
  "msg": "操作描述",
  "data": {}
}
```

### HTTP状态码映射

| HTTP状态码 | code字段 | 说明 |
|:----------|:---------|:-----|
| 200 | 200 | 成功 |
| 400 | 400 | 客户端错误（参数错误、配额超限等） |
| 500 | 500 | 服务器内部错误 |

---

## 🎯 API端点列表

### 1. 获取题目列表

获取所有可用的挑战题目。

**请求:**
```http
GET /api/challenges
```

**响应示例:**
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "id": "1",
      "title": "Nginx 基础挑战",
      "description": "找到隐藏的Flag",
      "category": "Web",
      "difficulty": "Easy",
      "image": "nginx:alpine",
      "points": 100,
      "created_at": "2026-01-27T00:00:00Z",
      "updated_at": "2026-01-27T00:00:00Z"
    }
  ]
}
```

**字段说明:**
- `id`: 题目唯一标识
- `title`: 题目标题
- `description`: 题目描述
- `category`: 类别（Web/Pwn/Reverse/Crypto）
- `difficulty`: 难度（Easy/Medium/Hard）
- `image`: Docker镜像名称
- `points`: 题目分值
- `flag`: 隐藏字段，不返回给客户端

---

### 2. 启动挑战实例

为用户创建并启动一个挑战容器实例。

**请求:**
```http
POST /api/challenges/:id/start
```

**路径参数:**
- `id`: 题目ID（例如：`1`）

**响应示例（成功）:**
```json
{
  "code": 200,
  "msg": "Instance started successfully",
  "data": {
    "id": "abc-123-def-456",
    "user_id": "user_mock_001",
    "challenge_id": "1",
    "container_id": "a1b2c3d4e5f6",
    "port": 23456,
    "status": "running",
    "expires_at": "2026-01-27T05:00:00Z",
    "created_at": "2026-01-27T04:00:00Z"
  }
}
```

**响应示例（配额超限）:**
```json
{
  "code": 400,
  "msg": "quota exceeded: max 1 active instance per user"
}
```

**字段说明:**
- `id`: 实例唯一标识
- `user_id`: 用户ID
- `challenge_id`: 题目ID
- `container_id`: Docker容器ID
- `port`: 映射到宿主机的端口（范围：20000-40000）
- `status`: 实例状态（running/stopped/expired）
- `expires_at`: 过期时间（默认1小时后）
- `flag`: 隐藏字段，用户需通过攻击容器获取

**资源限制:**
- 内存：128MB
- CPU：0.5核心
- 存活时间：1小时（由The Reaper自动清理）
- 每用户同时最多：1个实例

**访问实例:**
```
http://localhost:{port}
```

---

### 3. 停止挑战实例

停止并删除用户的挑战容器实例。

**请求:**
```http
POST /api/challenges/:id/stop
```

**路径参数:**
- `id`: 题目ID

**响应示例:**
```json
{
  "code": 200,
  "msg": "Instance stopped successfully",
  "data": {
    "status": "stopped"
  }
}
```

**说明:**
- 强制停止容器并立即删除
- 清理Redis状态
- 更新数据库记录状态为 `stopped`

---

### 4. 提交Flag验证

提交Flag答案进行验证。

**请求:**
```http
POST /api/submit
Content-Type: application/json

{
  "challenge_id": "1",
  "flag": "flag{user_mock_001_1738024000_a1b2c3d4}"
}
```

**请求体:**
```json
{
  "challenge_id": "string (必填)",
  "flag": "string (必填)"
}
```

**响应示例（正确）:**
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "correct": true,
    "message": "回答正确！你获得了积分。"
  }
}
```

**响应示例（错误）:**
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "correct": false,
    "message": "Flag 错误，请重试。"
  }
}
```

**说明:**
- Flag必须与用户当前运行实例的动态Flag完全匹配
- 正确提交后自动加分（题目points值）
- 记录所有提交历史（correct/incorrect）
- 如果用户没有运行实例，返回提示信息

**Flag格式:**
```
flag{userID_timestamp_random}
示例: flag{user_mock_001_1738024567_a1b2c3d4}
```

---

## 🔐 安全机制

### 1. 资源隔离
- ✅ 每个容器严格限制为128MB内存和0.5 CPU
- ✅ 端口随机分配（20000-40000），避免冲突
- ✅ 容器自动过期（1小时），由The Reaper清理

### 2. 配额控制
- ✅ 每个用户同时只能运行1个实例
- ✅ 防止资源耗尽攻击

### 3. Flag安全
- ✅ 每个用户的Flag动态生成，包含用户ID和时间戳
- ✅ 防止Flag重复使用
- ✅ Flag不在API响应中返回

---

## 📊 错误码参考

| code | msg示例 | 说明 |
|:-----|:--------|:-----|
| 200 | success | 操作成功 |
| 400 | Invalid request format | 请求格式错误 |
| 400 | quota exceeded: max 1 active instance per user | 配额超限 |
| 400 | challenge not found | 题目不存在 |
| 400 | no active instance found | 用户没有运行中的实例 |
| 500 | Failed to fetch challenges | 服务器内部错误 |
| 500 | Verification failed | Flag验证失败 |

---

## 🧪 测试示例

### 完整流程示例

```bash
# 1. 获取题目列表
curl http://localhost:8080/api/challenges

# 2. 启动实例
curl -X POST http://localhost:8080/api/challenges/1/start

# 返回示例:
# {
#   "code": 200,
#   "data": {
#     "id": "inst-123",
#     "port": 23456,
#     ...
#   }
# }

# 3. 访问靶机（浏览器或curl）
curl http://localhost:23456

# 4. 获取Flag后提交
curl -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{"challenge_id": "1", "flag": "flag{...}"}'

# 5. 停止实例
curl -X POST http://localhost:8080/api/challenges/1/stop
```

---

## 🚀 未来功能（待实现）

以下API在需求文档中定义，但当前版本未实现：

- [ ] `POST /api/register` - 用户注册
- [ ] `POST /api/login` - 用户登录
- [ ] `GET /api/me` - 获取当前用户信息
- [ ] `GET /api/leaderboard` - 排行榜
- [ ] `GET /api/submissions` - 提交历史记录
- [ ] `GET /api/challenges/:id` - 获取单个题目详情

管理员API（未实现）：
- [ ] `POST /api/admin/challenges` - 创建题目
- [ ] `PUT /api/admin/challenges/:id` - 更新题目
- [ ] `DELETE /api/admin/challenges/:id` - 删除题目
- [ ] `POST /api/admin/instances/:id/stop` - 强制停止任意实例

---

## 📝 变更日志

### v1.0.0 (2026-01-27)
- ✅ 实现核心4个API接口
- ✅ 添加配额限制（max 1 per user）
- ✅ 实现动态Flag生成
- ✅ 添加容器资源限制（128MB/0.5CPU）
- ✅ 实现The Reaper自动清理过期实例
- ✅ 标准化JSON响应格式

---

## 📞 技术支持

如有问题，请参考：
- [测试脚本](../test_core_features.sh) - 完整的API测试示例
- [需求文档](../docs/需求文档.md) - 详细的功能需求
- [实施计划](../.gemini/antigravity/brain/*/implementation_plan.md)
