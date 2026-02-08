# 前后端集成说明

## 🔗 已完成的工作

已将前端的Mock API替换为真实的后端API调用！

### 修改的文件

1. **`web/src/api/challenges.ts`** - API层
   - ✅ 移除Mock数据
   - ✅ 使用axios调用真实后端API
   - ✅ 添加响应拦截器处理标准格式 `{code, msg, data}`
   - ✅ 完善错误处理

2. **`web/vite.config.ts`** - 开发服务器配置
   - ✅ 添加代理配置，将 `/api` 请求代理到 `http://localhost:8080`

---

## 🚀 如何运行

### 1. 启动后端服务（Go）

```bash
# 在项目根目录
go run cmd/api/main.go
```

后端将运行在 `http://localhost:8080`

### 2. 启动前端服务（React + Vite）

```bash
# 进入前端目录
cd web

# 安装依赖（首次运行）
npm install

# 启动开发服务器
npm run dev
```

前端将运行在 `http://localhost:5173`

### 3. 访问应用

打开浏览器访问：`http://localhost:5173`

---

## 📡 API对接情况

| 前端功能 | API端点 | 方法 | 状态 |
|:---------|:--------|:-----|:-----|
| 获取题目列表 | `/api/challenges` | GET | ✅ 已对接 |
| 启动实例 | `/api/challenges/:id/start` | POST | ✅ 已对接 |
| 停止实例 | `/api/challenges/:id/stop` | POST | ✅ 已对接 |
| 提交Flag | `/api/submit` | POST | ✅ 已对接 |

---

## 🔧 技术细节

### API配置

```typescript
// web/src/api/challenges.ts
const API_BASE_URL = 'http://localhost:8080/api';
```

**开发环境：** Vite代理会自动将 `/api` 请求转发到后端
**生产环境：** 需要配置Nginx反向代理或修改 `API_BASE_URL`

### 响应格式处理

后端统一返回格式：
```json
{
  "code": 200,
  "msg": "success",
  "data": {...}
}
```

前端axios拦截器自动提取 `data` 字段，业务代码直接使用数据。

### 错误处理

- 配额超限（400）→ 显示 "配额超限：每个用户最多同时运行1个实例"
- 服务器错误（500）→ 显示具体错误信息
- 网络错误 → 显示 "请检查Docker是否运行"

---

## 🧪 测试步骤

### 1. 验证后端运行

```bash
curl http://localhost:8080/api/challenges
```

应返回题目列表（JSON格式）。

### 2. 前端操作流程

1. 打开 `http://localhost:5173`
2. 点击任意题目的"启动"按钮
3. 等待实例启动（约2-3秒）
4. 点击"查看详情"
5. 访问显示的实例URL（例如 `http://localhost:23456`）
6. 找到Flag后提交

### 3. 检查浏览器控制台

打开浏览器开发者工具（F12）→ Network标签页，查看API请求：
- 应该看到对 `/api/challenges` 等的请求
- 响应应该是 `{code: 200, msg: "success", data: [...]}`

---

## ⚠️ 常见问题

### 问题1: CORS错误

**症状：** 浏览器控制台显示 "Access-Control-Allow-Origin"错误

**解决：** 后端已配置CORS允许 `http://localhost:5173`，检查：
```go
// cmd/api/main.go
AllowOrigins: []string{"http://localhost:5173"}
```

### 问题2: 404 Not Found

**症状：** API请求返回404

**检查：**
1. 后端服务是否运行？`curl http://localhost:8080/api/challenges`
2. 端口是否正确？后端默认8080，前端默认5173

### 问题3: 实例启动失败

**症状：** 点击"启动"后报错

**检查：**
1. Docker是否运行？`docker ps`
2. MySQL是否运行？`docker ps | grep mysql`
3. Redis是否运行？`docker ps | grep redis`

---

## 📦 生产部署

### 方案1: 前后端同源部署

使用Nginx反向代理：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    # 前端静态文件
    location / {
        root /path/to/web/dist;
        try_files $uri /index.html;
    }
    
    # 后端API代理
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
    }
}
```

### 方案2: 分离部署

修改前端API地址：
```typescript
const API_BASE_URL = 'https://api.your-domain.com/api';
```

确保后端CORS配置允许前端域名。

---

## 📝 后续优化建议

- [ ] 添加JWT Token认证
- [ ] 实现请求重试机制
- [ ] 添加Loading骨架屏
- [ ] 实现WebSocket实时状态更新
- [ ] 添加错误边界组件
