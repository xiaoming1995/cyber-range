# 🔒 方案 B：HTTPS + TLS 模式详解

本文档详细讲解 Docker TLS 双向认证的原理和实施步骤。

---

## 📊 整体架构

```
┌─────────────────────┐                  ┌──────────────────────┐
│  应用服务器          │                  │  远程 Docker 服务器   │
│  (Cyber Range)      │                  │                      │
│                     │  TLS 加密连接     │                      │
│  ┌───────────────┐  │ ←─────────────→  │  ┌────────────────┐  │
│  │ Docker Client │  │  双向证书认证     │  │ Docker Daemon  │  │
│  └───────────────┘  │                  │  └────────────────┘  │
│  证书文件：         │                  │  证书文件：          │
│  - ca.pem          │                  │  - ca.pem           │
│  - cert.pem        │                  │  - server-cert.pem  │
│  - key.pem         │                  │  - server-key.pem   │
└─────────────────────┘                  └──────────────────────┘
        ↓                                         ↓
   端口 2376                                  监听 2376
```

---

## 🔐 安全机制

### 1. TLS 加密传输
- **通信内容加密**：所有 Docker API 调用都通过 TLS 加密
- **防窃听**：中间人无法查看传输内容
- **防篡改**：数据在传输过程中不能被修改

### 2. 双向证书认证

**服务器认证（Server Authentication）**：
```
客户端验证服务器身份
↓
检查服务器证书是否由可信 CA 签发
↓
检查证书 CN/SAN 是否匹配服务器地址
↓
验证通过后建立连接
```

**客户端认证（Client Authentication）**：
```
服务器要求客户端提供证书
↓
验证客户端证书是否由可信 CA 签发
↓
验证客户端证书权限
↓
允许连接
```

---

## 📋 实施步骤详解

### 第一阶段：证书生成（在远程 Docker 服务器上）

#### 步骤 1.1 - 生成 CA（证书颁发机构）

**为什么需要 CA？**
- CA 是整个证书体系的根
- 用于签发服务器证书和客户端证书
- 只有 CA 签发的证书才会被信任

**具体操作**：
```bash
# 生成 CA 私钥（4096 位 RSA，AES256 加密保护）
openssl genrsa -aes256 -out ca-key.pem 4096
# 输入密码保护私钥（例如：MySecurePassword123!）

# 生成自签名 CA 证书（有效期 1 年）
openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem
```

**交互式信息填写**：
```
Country Name (2 letter code) [XX]: CN
State or Province Name (full name) []: Beijing
Locality Name (eg, city) []: Beijing
Organization Name (eg, company) []: Your Company
Organizational Unit Name (eg, section) []: IT Department
Common Name (eg, your name or server's hostname) []: Docker CA
Email Address []: admin@yourcompany.com
```

**生成的文件**：
- `ca-key.pem`：CA 私钥（**极其重要，必须妥善保管**）
- `ca.pem`：CA 公钥证书（可以分发）

---

#### 步骤 1.2 - 生成服务器证书

**目的**：让客户端能够验证服务器身份

**操作流程**：

```bash
# 1. 生成服务器私钥
openssl genrsa -out server-key.pem 4096

# 2. 创建证书签名请求（CSR）
# 假设服务器 IP 是 192.168.1.100
openssl req -subj "/CN=192.168.1.100" -sha256 -new -key server-key.pem -out server.csr

# 3. 配置 Subject Alternative Name（关键！）
# 这个步骤很重要，现代 TLS 要求证书必须包含 SAN
cat > extfile.cnf <<EOF
subjectAltName = IP:192.168.1.100,IP:127.0.0.1
EOF

# 如果使用域名，还可以添加 DNS：
# subjectAltName = DNS:docker.example.com,IP:192.168.1.100,IP:127.0.0.1

# 4. 使用 CA 签发服务器证书
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-cert.pem -extfile extfile.cnf
# 需要输入 CA 私钥密码
```

**关键概念 - Subject Alternative Name (SAN)**：
```
为什么需要 SAN？
├─ 传统证书只在 Common Name (CN) 中指定主机名
├─ 现代浏览器和客户端要求使用 SAN
├─ SAN 可以包含多个 IP 地址或域名
└─ 客户端连接时会验证目标地址是否在 SAN 列表中
```

**生成的文件**：
- `server-key.pem`：服务器私钥
- `server-cert.pem`：服务器证书
- `server.csr`：证书签名请求（可删除）
- `extfile.cnf`：扩展配置文件（可删除）

---

#### 步骤 1.3 - 生成客户端证书

**目的**：让服务器能够验证客户端身份（双向认证）

```bash
# 1. 生成客户端私钥
openssl genrsa -out key.pem 4096

# 2. 创建客户端 CSR
openssl req -subj '/CN=client' -new -key key.pem -out client.csr

# 3. 配置客户端证书扩展（标记为客户端认证用途）
echo "extendedKeyUsage = clientAuth" > extfile-client.cnf

# 4. 使用 CA 签发客户端证书
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out cert.pem -extfile extfile-client.cnf
# 需要输入 CA 私钥密码
```

**关键概念 - Extended Key Usage**：
```
extendedKeyUsage = clientAuth
├─ 明确标记证书用于客户端认证
├─ 服务器会检查这个标记
└─ 防止用其他用途的证书进行认证
```

**生成的文件**：
- `key.pem`：客户端私钥
- `cert.pem`：客户端证书

---

#### 步骤 1.4 - 设置安全权限

```bash
# 私钥文件：只有所有者可读（400）
chmod 0400 ca-key.pem key.pem server-key.pem

# 证书文件：所有人可读但不可写（444）
chmod 0444 ca.pem server-cert.pem cert.pem
```

**权限说明**：
```
0400 (-r--------)  只有所有者可以读取
0444 (-r--r--r--)  所有人可以读取，但没人可以写入
```

---

### 第二阶段：配置远程 Docker 服务器

#### 步骤 2.1 - 部署证书

```bash
# 创建系统证书目录
sudo mkdir -p /etc/docker/certs

# 复制服务器需要的证书
sudo cp ca.pem server-cert.pem server-key.pem /etc/docker/certs/

# 设置系统级权限
sudo chmod 0400 /etc/docker/certs/server-key.pem
sudo chmod 0444 /etc/docker/certs/ca.pem /etc/docker/certs/server-cert.pem
```

**证书用途**：
- `ca.pem`：验证客户端证书
- `server-cert.pem`：向客户端证明服务器身份
- `server-key.pem`：服务器私钥，用于 TLS 握手

---

#### 步骤 2.2 - 配置 Docker Daemon

修改 `/etc/docker/daemon.json`：

```json
{
  "hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"],
  "tls": true,
  "tlscert": "/etc/docker/certs/server-cert.pem",
  "tlskey": "/etc/docker/certs/server-key.pem",
  "tlscacert": "/etc/docker/certs/ca.pem",
  "tlsverify": true
}
```

**配置详解**：

| 参数 | 作用 | 说明 |
|------|------|------|
| `hosts` | 监听地址 | tcp://0.0.0.0:2376 + 本地 socket |
| `tls: true` | 启用 TLS | 加密通信 |
| `tlsverify: true` | 启用客户端验证 | 双向认证 |
| `tlscert` | 服务器证书 | 证明服务器身份 |
| `tlskey` | 服务器私钥 | TLS 握手使用 |
| `tlscacert` | CA 证书 | 验证客户端证书 |

---

#### 步骤 2.3 - systemd 配置

创建 `/etc/systemd/system/docker.service.d/override.conf`：

```ini
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
```

**为什么需要这个配置？**
- systemd 的默认 ExecStart 可能包含 -H 参数，与 daemon.json 中的 hosts 配置冲突
- 清空默认配置，让 Docker 完全从 daemon.json 读取

---

#### 步骤 2.4 - 重启服务

```bash
sudo systemctl daemon-reload   # 重新加载 systemd 配置
sudo systemctl restart docker  # 重启 Docker
sudo systemctl status docker   # 检查状态
```

**验证监听**：
```bash
sudo ss -tlnp | grep 2376

# 预期输出：
# LISTEN  0  128  [::]:2376  [::]:*  users:(("dockerd",pid=1234,fd=5))
```

---

### 第三阶段：配置应用服务器

#### 步骤 3.1 - 传输客户端证书

```bash
# 在应用服务器上执行
mkdir -p /Users/liujiming/web/cyber-range/certs/docker

# 从远程服务器复制证书
scp user@192.168.1.100:~/docker-certs/ca.pem \
    /Users/liujiming/web/cyber-range/certs/docker/

scp user@192.168.1.100:~/docker-certs/cert.pem \
    /Users/liujiming/web/cyber-range/certs/docker/

scp user@192.168.1.100:~/docker-certs/key.pem \
    /Users/liujiming/web/cyber-range/certs/docker/
```

**安全提醒**：
- ⚠️ 私钥文件 `key.pem` 要安全传输
- ⚠️ 不要通过不安全的渠道（如明文邮件）传输
- ✅ 推荐使用 scp、rsync 等加密传输工具

---

#### 步骤 3.2 - 设置证书权限

```bash
cd /Users/liujiming/web/cyber-range/certs/docker/

chmod 0400 key.pem          # 私钥只有所有者可读
chmod 0444 ca.pem cert.pem  # 证书所有人可读
```

---

#### 步骤 3.3 - 修改应用配置

编辑 `configs/config.yaml`：

```yaml
docker:
  mode: "remote"  # 切换到远程模式
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://192.168.1.100:2376"  # ⚠️ 注意端口是 2376
    tls_verify: true                   # 启用 TLS 验证
    cert_path: "/Users/liujiming/web/cyber-range/certs/docker"  # 证书目录
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

---

### 第四阶段：验证和测试

#### 验证 1：命令行测试

```bash
# 方法 1：使用环境变量
export DOCKER_HOST=tcp://192.168.1.100:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/Users/liujiming/web/cyber-range/certs/docker

docker ps
docker version
docker info

# 方法 2：手动指定参数
docker -H tcp://192.168.1.100:2376 \
  --tlsverify \
  --tlscacert=/Users/liujiming/web/cyber-range/certs/docker/ca.pem \
  --tlscert=/Users/liujiming/web/cyber-range/certs/docker/cert.pem \
  --tlskey=/Users/liujiming/web/cyber-range/certs/docker/key.pem \
  ps
```

**成功标志**：
- ✅ 能够列出容器（即使列表为空）
- ✅ 无 TLS 错误
- ✅ 无证书验证错误

---

#### 验证 2：测试证书认证

```bash
# 1. 测试不带证书连接（应该失败）
docker -H tcp://192.168.1.100:2376 ps
# 预期错误：需要 TLS 证书

# 2. 测试正确证书连接（应该成功）
docker -H tcp://192.168.1.100:2376 --tlsverify \
  --tlscacert=ca.pem --tlscert=cert.pem --tlskey=key.pem ps
# 预期成功
```

---

#### 验证 3：应用集成测试

```bash
cd /Users/liujiming/web/cyber-range

# 清除可能干扰的环境变量
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_PATH

# 启动应用
go run cmd/api/main.go
```

**检查日志**：
- ✅ 看到：Docker client initialized successfully
- ❌ 不应看到：TLS handshake error
- ❌ 不应看到：certificate verify failed

---

## 🔍 TLS 连接流程详解

```
┌─────────────┐                           ┌──────────────┐
│   客户端     │                           │   服务器      │
└──────┬──────┘                           └──────┬───────┘
       │                                         │
       │  1. Client Hello (支持的加密套件)       │
       │ ──────────────────────────────────────> │
       │                                         │
       │  2. Server Hello (选择加密套件)         │
       │ <────────────────────────────────────── │
       │                                         │
       │  3. 服务器证书 (server-cert.pem)        │
       │ <────────────────────────────────────── │
       │                                         │
       │  4. 请求客户端证书                      │
       │ <────────────────────────────────────── │
       │                                         │
       │  验证服务器证书：                       │
       │  - 是否由 ca.pem 签发？                 │
       │  - CN/SAN 是否匹配 192.168.1.100？     │
       │  - 是否在有效期内？                     │
       │                                         │
       │  5. 客户端证书 (cert.pem)               │
       │ ──────────────────────────────────────> │
       │                                         │
       │                        验证客户端证书： │
       │                        - 是否由 ca.pem 签发？
       │                        - 是否有 clientAuth 权限？
       │                        - 是否在有效期内？
       │                                         │
       │  6. Finished (握手完成)                 │
       │ <─────────────────────────────────────> │
       │                                         │
       │  7. 加密的应用数据                      │
       │ <─────────────────────────────────────> │
       │                                         │
```

---

## 📊 证书信任链

```
CA 证书 (ca.pem)
├─ 签发 → 服务器证书 (server-cert.pem)
│         └─ 用途：服务器身份认证
│         └─ CN: 192.168.1.100
│         └─ SAN: IP:192.168.1.100,IP:127.0.0.1
│
└─ 签发 → 客户端证书 (cert.pem)
          └─ 用途：客户端身份认证
          └─ CN: client
          └─ ExtKeyUsage: clientAuth
```

---

## 🔐 安全优势

| 安全特性 | 方案 A (HTTP) | 方案 B (TLS) |
|----------|---------------|--------------|
| **加密传输** | ❌ 明文 | ✅ TLS 1.2+ 加密 |
| **身份认证** | ❌ 无 | ✅ 双向证书认证 |
| **防中间人攻击** | ❌ 不防护 | ✅ 证书验证 |
| **防重放攻击** | ❌ 不防护 | ✅ TLS 序列号 |
| **访问控制** | ❌ 所有人可连接 | ✅ 仅持有证书者可连接 |

---

## ⚡ 性能影响

**TLS 开销**：
- 首次握手：约 50-100ms
- 后续请求：几乎无影响（会话复用）
- CPU 开销：可忽略（现代 CPU 有硬件加速）

**实际影响**：
- 启动容器：从 2 秒变为 2.1 秒
- 停止容器：几乎无差别
- 总体：**可忽略不计**

---

## 🎯 关键注意事项

### 1. 证书有效期管理
```bash
# 检查证书过期时间
openssl x509 -in cert.pem -noout -dates

# 建议：
# - 生产环境证书有效期：1-2 年
# - 提前 30 天续期
# - 设置过期提醒
```

### 2. CA 私钥安全
```
ca-key.pem 的重要性：
├─ 拥有此文件 = 可以签发任意证书
├─ 泄露后果：攻击者可以伪造客户端/服务器
└─ 防护措施：
    ├─ 使用强密码保护
    ├─ 存储在安全位置
    ├─ 限制文件权限为 400
    └─ 定期备份（加密备份）
```

### 3. IP 地址变更
```
如果远程服务器 IP 变更：
1. 需要重新生成服务器证书（包含新 IP 的 SAN）
2. 更新 Docker daemon.json
3. 重启 Docker
4. 更新应用配置中的 host

或者使用域名代替 IP（推荐）
```

### 4. 证书撤销
```
如果客户端证书泄露：
├─ 方案 1：重新生成CA和所有证书（彻底但麻烦）
├─ 方案 2：实施 CRL（证书撤销列表）
└─ 方案 3：使用短期证书（如 30 天）
```

---

## 📁 证书目录结构

**远程服务器**：
```
/etc/docker/certs/
├── ca.pem           # CA 证书
├── server-cert.pem  # 服务器证书
└── server-key.pem   # 服务器私钥
```

**应用服务器**：
```
/Users/liujiming/web/cyber-range/certs/docker/
├── ca.pem    # CA 证书
├── cert.pem  # 客户端证书
└── key.pem   # 客户端私钥
```

---

## 🔖 快速参考

### 常用命令

```bash
# 查看证书信息
openssl x509 -in cert.pem -text -noout

# 验证证书链
openssl verify -CAfile ca.pem cert.pem

# 查看证书有效期
openssl x509 -in cert.pem -noout -dates

# 测试 TLS 连接
docker -H tcp://IP:2376 --tlsverify \
  --tlscacert=ca.pem --tlscert=cert.pem --tlskey=key.pem ps
```

### 配置速查

| 场景 | 配置值 |
|------|--------|
| TLS 端口 | 2376 |
| HTTP 端口 | 2375 |
| 证书目录 | `/Users/liujiming/web/cyber-range/certs/docker` |
| 应用配置 | `configs/config.yaml` |

---

**文档版本**: 1.0  
**最后更新**: 2026-01-28
