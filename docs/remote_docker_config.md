# 🌐 Docker 部署模式配置指南

本文档说明如何配置靶场系统的 Docker 容器部署模式，支持本地和远程两种模式。

## 配置说明

### 核心参数

- **`mode`**: 部署模式选择
  - `"local"`: 使用本地 Docker（默认）
  - `"remote"`: 使用远程 Docker 主机
  
- **`local`**: 本地 Docker 配置（当 mode=local 时生效）
- **`remote`**: 远程 Docker 配置（当 mode=remote 时生效）

---

## 场景1: 本地Docker（默认）

适用于开发环境或单机部署。

```yaml
docker:
  mode: "local"  # 使用本地模式
  
  local:
    host: ""  # 留空使用本地 Docker socket
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://192.168.1.100:2375"  # 备用配置（不会被使用）
    tls_verify: false
    cert_path: ""
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

## 场景2: 远程Docker（HTTP，无TLS）
⚠️ **仅用于内网测试环境**

```yaml
docker:
  mode: "remote"  # 切换到远程模式
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://192.168.1.100:2375"  # 远程主机地址
    tls_verify: false
    cert_path: ""
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

## 场景3: 远程Docker（HTTPS，启用TLS）
✅ **生产环境推荐**

适用于跨网络的远程部署，提供加密和身份验证。

```yaml
docker:
  mode: "remote"  # 使用远程模式
  
  local:
    host: ""
    tls_verify: false
    cert_path: ""
    
  remote:
    host: "tcp://remote.example.com:2376"  # HTTPS端口为2376
    tls_verify: true  # 启用TLS验证
    cert_path: "/path/to/docker/certs"  # 证书目录
    
  port_range_min: 20000
  port_range_max: 40000
  memory_limit: 134217728
  cpu_limit: 0.5
```

### TLS证书目录结构：
```
/path/to/docker/certs/
  ├── ca.pem      # CA证书
  ├── cert.pem    # 客户端证书
  └── key.pem     # 客户端私钥
```

---

## 🔄 切换部署模式

只需修改 `config.yaml` 中的 `mode` 字段即可切换：

- 切换到本地模式：`mode: "local"`
- 切换到远程模式：`mode: "remote"`

修改后重启应用即可生效。

---

## 🔒 如何配置远程Docker服务器启用TLS

### 在远程服务器上：
```bash
# 1. 生成证书（使用Docker官方脚本）
$ git clone https://github.com/docker/docker.github.io.git
$ cd docker.github.io/engine/security/https
$ ./generate-certs.sh

# 2. 配置Docker daemon
$ sudo vim /etc/docker/daemon.json
{
  "hosts": ["tcp://0.0.0.0:2376", "unix:///var/run/docker.sock"],
  "tls": true,
  "tlscert": "/etc/docker/certs/server-cert.pem",
  "tlskey": "/etc/docker/certs/server-key.pem",
  "tlscacert": "/etc/docker/certs/ca.pem",
  "tlsverify": true
}

# 3. 重启Docker
$ sudo systemctl restart docker
```

### 在应用服务器上：
将客户端证书（ca.pem, cert.pem, key.pem）复制到应用服务器，并在 `config.yaml` 中配置路径。

## ✅ 验证连接
```bash
# 本地测试远程连接
docker -H tcp://remote-server:2376 --tlsverify \
  --tlscacert=ca.pem \
  --tlscert=cert.pem \
  --tlskey=key.pem \
  ps
```
