# 📚 Cyber Range 脚本与文档索引

本文档整理了项目中所有脚本和文档资源，便于快速查阅和使用。

---

## 📁 目录

- [脚本索引 (scripts/)](#脚本索引)
- [文档索引 (docs/)](#文档索引)
- [命令与工具入口 (cmd/)](#命令与工具入口)

---

## 🔧 脚本索引

> 路径：`/scripts/`

### 运维脚本

| 脚本名称 | 用途 | 使用场景 |
|----------|------|----------|
| [start-registry.sh](../scripts/start-registry.sh) | 启动本地 Docker Registry | 首次部署/Registry 挂掉后恢复 |
| [import-image.sh](../scripts/import-image.sh) | 导入镜像到 Registry | 上传新的挑战镜像 |
| [publish_all_challenges.sh](../scripts/publish_all_challenges.sh) | 批量发布所有题目 | 初始化后批量上架题目 |

### Docker TLS 配置脚本

| 脚本名称 | 用途 | 使用场景 |
|----------|------|----------|
| [generate-docker-tls-certs.sh](../scripts/generate-docker-tls-certs.sh) | 自动生成 TLS 证书 | 配置远程 Docker 安全连接 |
| [configure-docker-server.sh](../scripts/configure-docker-server.sh) | 配置远程 Docker 服务器 | 在远程服务器上部署 TLS 证书 |

### 测试脚本

| 脚本名称 | 用途 | 使用场景 |
|----------|------|----------|
| [test_admin_api.sh](../scripts/test_admin_api.sh) | 管理后台 API 测试 | 验证后台登录、题目CRUD、搜索筛选 |
| [test_core_features.sh](../scripts/test_core_features.sh) | 核心功能集成测试 | 验证实例启动/停止、Flag验证、积分 |

### 调试脚本

| 脚本名称 | 用途 | 使用场景 |
|----------|------|----------|
| [debug_docker_hosts.sh](../scripts/debug_docker_hosts.sh) | 排查 Docker 主机问题 | 前端下拉框无数据时排查 |

---

### 脚本详细说明

#### 🐳 start-registry.sh
```bash
# 启动本地 Docker Registry
./scripts/start-registry.sh

# 功能：
# - 创建数据目录 ~/cyber-range-registry
# - 启动 registry:2 容器，映射 5000 端口
# - 自动检测已运行的 Registry
```

#### 📦 import-image.sh
```bash
# 导入镜像到 Registry
./scripts/import-image.sh <镜像tar文件> [镜像名称] [标签]

# 示例：
./scripts/import-image.sh my-challenge.tar web-xss v1.0

# 流程：
# 1. docker load 加载 tar 文件
# 2. docker tag 重新打标签
# 3. docker push 推送到 localhost:5000
```

#### 📢 publish_all_challenges.sh
```bash
# 批量发布所有题目
./scripts/publish_all_challenges.sh

# 功能：
# - 自动登录获取 Token
# - 遍历所有题目并设置状态为 published
```

#### 🔐 generate-docker-tls-certs.sh
```bash
# 生成 Docker TLS 证书
./scripts/generate-docker-tls-certs.sh --ip 192.168.1.100

# 参数：
#   -i, --ip       服务器 IP（必填）
#   -d, --domain   服务器域名（可选）
#   -o, --output   输出目录（默认: ./docker-certs）
#   --no-password  不使用密码保护 CA 私钥
#   --days         证书有效期（默认: 365天）

# 生成的文件：
# - ca.pem, ca-key.pem        → CA 证书和私钥
# - server-cert.pem, server-key.pem → 服务器证书
# - cert.pem, key.pem         → 客户端证书
```

#### 🖥️ configure-docker-server.sh
```bash
# 在远程服务器上执行，配置 Docker TLS
scp /tmp/ca.pem /tmp/server-cert.pem /tmp/server-key.pem user@remote:/tmp/
ssh user@remote "sudo ./configure-docker-server.sh"

# 功能：
# - 部署证书到 /etc/docker/certs/
# - 配置 daemon.json 启用 TLS
# - 配置防火墙开放 2376 端口
# - 重启 Docker 服务
```

---

## 📖 文档索引

> 路径：`/docs/`

### 核心文档

| 文档名称 | 内容概述 | 重要程度 |
|----------|----------|----------|
| [需求文档.md](需求文档.md) | 项目完整需求、功能规划、验收标准 | ⭐⭐⭐ |
| [API.md](API.md) | 用户端 API 接口文档 | ⭐⭐⭐ |
| [DATABASE_COMMENTS.md](DATABASE_COMMENTS.md) | 数据库表结构和字段说明 | ⭐⭐ |

### Docker 配置文档

| 文档名称 | 内容概述 | 重要程度 |
|----------|----------|----------|
| [REMOTE_DOCKER_SETUP.md](REMOTE_DOCKER_SETUP.md) | 远程 Docker 配置完整指南（HTTP/TLS） | ⭐⭐⭐ |
| [DOCKER_TLS_DETAILED_GUIDE.md](DOCKER_TLS_DETAILED_GUIDE.md) | TLS 双向认证原理和详细步骤 | ⭐⭐ |
| [DOCKER_SETUP_NO_SSH.md](DOCKER_SETUP_NO_SSH.md) | 无 SSH 情况下的 Docker 配置方案 | ⭐ |
| [remote_docker_config.md](remote_docker_config.md) | 远程 Docker 快速配置参考 | ⭐ |

### 镜像管理文档

| 文档名称 | 内容概述 | 重要程度 |
|----------|----------|----------|
| [image_registry_implementation.md](image_registry_implementation.md) | 镜像仓库实现方案和技术细节 | ⭐⭐ |
| [image_optimization_plan.md](image_optimization_plan.md) | 镜像优化计划 | ⭐ |
| [IMAGE_OPTIMIZATION_INDEX.md](IMAGE_OPTIMIZATION_INDEX.md) | 镜像优化索引 | ⭐ |
| [REMOTE_DOCKER_REGISTRY_SETUP.md](REMOTE_DOCKER_REGISTRY_SETUP.md) | 远程 Registry 配置 | ⭐ |

### 前端与系统文档

| 文档名称 | 内容概述 | 重要程度 |
|----------|----------|----------|
| [FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md) | 前后端集成说明 | ⭐⭐ |
| [admin-frontend-design.md](admin-frontend-design.md) | 管理后台前端设计 | ⭐⭐ |
| [LOGGING_SYSTEM_PLAN.md](LOGGING_SYSTEM_PLAN.md) | 日志系统设计方案 | ⭐ |

### 测试与调试文档

| 文档名称 | 内容概述 | 重要程度 |
|----------|----------|----------|
| [golang_testing_guide.md](golang_testing_guide.md) | Go 测试指南 | ⭐⭐ |
| [testing_instance_creation.md](testing_instance_creation.md) | 实例创建测试文档 | ⭐ |
| [SEED_DATA.md](SEED_DATA.md) | 种子数据说明 | ⭐ |

---

## 🚀 命令与工具入口

> 路径：`/cmd/`

| 命令目录 | 用途 | 运行方式 |
|----------|------|----------|
| `cmd/api/` | **主服务入口** | `go run cmd/api/main.go` |
| `cmd/migrate/` | 数据库迁移 | `go run cmd/migrate/main.go` |
| `cmd/seed/` | 种子数据初始化 | `go run cmd/seed/main.go` |
| `cmd/diagnose/` | 系统诊断 | `go run cmd/diagnose/main.go` |
| `cmd/diagnose_all_hosts/` | 诊断所有 Docker 主机 | `go run cmd/diagnose_all_hosts/main.go` |
| `cmd/diagnose_host/` | 诊断单个 Docker 主机 | `go run cmd/diagnose_host/main.go` |
| `cmd/enable_privileged/` | 启用特权模式 | `go run cmd/enable_privileged/main.go` |
| `cmd/disable_remote_host/` | 禁用远程主机 | `go run cmd/disable_remote_host/main.go` |

---

## 📋 快速参考

### 常用命令速查

```bash
# 🚀 启动服务
go run cmd/api/main.go              # 启动后端
cd web && npm run dev               # 启动前端

# 📦 镜像管理
./scripts/start-registry.sh         # 启动 Registry
./scripts/import-image.sh xxx.tar   # 导入镜像

# 🔐 TLS 配置
./scripts/generate-docker-tls-certs.sh --ip YOUR_IP  # 生成证书

# 📢 题目管理
./scripts/publish_all_challenges.sh  # 批量发布题目

# 🧪 测试验证
./scripts/test_admin_api.sh          # 管理后台API测试
./scripts/test_core_features.sh      # 核心功能集成测试
```

---

**文档版本**: 1.0  
**创建时间**: 2026-02-09  
**维护者**: Cyber Range Team
