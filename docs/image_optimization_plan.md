# 本地镜像源优化方案

## 📋 需求背景

**现状问题**：
- 当前系统从公共 Docker Hub 拉取镜像，网络依赖高且速度慢
- 用户已本地打包好镜像，希望直接使用以提升性能和可控性

**优化目标**：
- ✅ 使用本地镜像，无需外网拉取
- ✅ 支持多 Docker 主机的镜像同步
- ✅ 镜像预加载机制，消除首次启动等待
- ✅ 镜像版本管理和更新机制

---

## 🎯 方案设计

### 方案 A：直接使用本地镜像（最简单）✅ 推荐

**适用场景**：单机部署 或 Docker 主机可以共享存储

#### 实施步骤

**1. 镜像导入到 Docker**

```bash
# 假设您的镜像文件为 ctf-web-challenge.tar
docker load -i /path/to/ctf-web-challenge.tar

# 验证导入成功
docker images | grep ctf-web-challenge
# 输出: ctf-web-challenge   v1.0   abc123   2 days ago   500MB
```

**2. 修改题目配置使用本地镜像**

```sql
-- 直接在数据库中更新镜像名
UPDATE challenges 
SET image = 'ctf-web-challenge:v1.0'  -- 使用导入的本地镜像
WHERE id = 'web-xss-001';
```

**3. 修改代码逻辑：跳过镜像拉取**

在 `internal/infra/docker/client.go` 中：

```go
func (d *DockerClient) StartContainer(ctx context.Context, imageName string, envVars []string) (string, int, error) {
    // 🔧 优化点1：检查本地镜像是否存在
    _, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
    if err != nil {
        // 镜像不存在，记录错误但不拉取
        return "", 0, fmt.Errorf("镜像不存在: %s，请先导入镜像", imageName)
    }
    
    // 🔧 优化点2：移除 ImagePull 逻辑
    // 不再尝试从外网拉取，直接使用本地镜像
    
    // 2. 分配端口...
    allocatedPort := d.AllocatePort()
    
    // 3. 创建容器（使用本地镜像）
    // ...
}
```

**优点**：
- ✅ 实施简单，改动最小
- ✅ 启动速度快（无网络拉取）
- ✅ 离线可用

**缺点**：
- ❌ 需要手动在每个 Docker 主机上导入镜像
- ❌ 镜像更新需要手动操作

---

### 方案 B：搭建私有镜像仓库（推荐多主机）

**适用场景**：多 Docker 主机环境，需要集中管理镜像

#### 架构图

```
┌─────────────────────────────────────────┐
│  本地私有 Docker Registry (端口 5000)    │
│  └─ ctf-web-challenge:v1.0              │
│  └─ ctf-pwn-buffer:v2.1                 │
└─────────────────────────────────────────┘
            ↓ ↓ ↓ (局域网)
┌──────────────┐  ┌──────────────┐
│ Docker 主机1  │  │ Docker 主机2  │
│ (本地 Mac)   │  │ (远程服务器)  │
└──────────────┘  └──────────────┘
```

#### 实施步骤

**1. 启动私有镜像仓库**

```bash
# 在本地 Mac 或专用服务器运行
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry \
  -v /data/registry:/var/lib/registry \
  registry:2

# 验证运行状态
curl http://localhost:5000/v2/_catalog
```

**2. 推送镜像到私有仓库**

```bash
# 加载本地镜像
docker load -i ctf-web-challenge.tar

# 重新打标签
docker tag ctf-web-challenge:v1.0 localhost:5000/ctf-web-challenge:v1.0

# 推送到私有仓库
docker push localhost:5000/ctf-web-challenge:v1.0
```

**3. 配置 Docker 主机信任私有仓库**

在远程 Docker 主机的 `/etc/docker/daemon.json`：

```json
{
  "insecure-registries": ["192.168.1.100:5000"]
}
```

```bash
# 重启 Docker
sudo systemctl restart docker
```

**4. 修改题目配置使用私有仓库镜像**

```sql
UPDATE challenges 
SET image = '192.168.1.100:5000/ctf-web-challenge:v1.0'
WHERE id = 'web-xss-001';
```

**5. 代码优化：智能拉取**

```go
func (d *DockerClient) StartContainer(ctx context.Context, imageName string, envVars []string) (string, int, error) {
    // 1. 检查本地是否已有镜像
    _, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
    
    if err != nil {
        // 本地没有，从私有仓库拉取（局域网速度快）
        logger.Info(ctx, "本地无镜像，从私有仓库拉取", "image", imageName)
        
        reader, pullErr := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
        if pullErr != nil {
            return "", 0, fmt.Errorf("镜像拉取失败: %w", pullErr)
        }
        defer reader.Close()
        
        // 等待拉取完成（加超时控制）
        ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()
        io.Copy(io.Discard, reader)
    }
    
    // 2. 创建容器...
}
```

**优点**：
- ✅ 集中管理镜像
- ✅ 多主机自动同步
- ✅ 支持版本控制
- ✅ 局域网拉取速度快

**缺点**：
- ❌ 需要额外维护 Registry 服务
- ❌ 配置相对复杂

---

### 方案 C：混合模式（灵活推荐）

**核心思想**：优先使用本地镜像，回退到私有仓库

#### 配置文件扩展

在 `configs/config.yaml` 中新增：

```yaml
docker:
  # 新增镜像源配置
  image_registry:
    mode: "hybrid"  # local | registry | hybrid
    local_only: false  # 强制只用本地镜像
    private_registry: "192.168.1.100:5000"  # 私有仓库地址
    pull_timeout: 300  # 拉取超时（秒）
```

#### 智能镜像加载逻辑

```go
type ImageStrategy struct {
    mode            string
    localOnly       bool
    privateRegistry string
    pullTimeout     int
}

func (d *DockerClient) EnsureImage(ctx context.Context, imageName string) error {
    // 1. 检查本地
    if d.hasLocalImage(ctx, imageName) {
        logger.Info(ctx, "使用本地镜像", "image", imageName)
        return nil
    }
    
    // 2. 如果是 local_only 模式，拒绝拉取
    if d.strategy.localOnly {
        return fmt.Errorf("镜像不存在且配置为仅使用本地镜像: %s", imageName)
    }
    
    // 3. 尝试从私有仓库拉取
    if d.strategy.mode == "hybrid" || d.strategy.mode == "registry" {
        registryImage := d.strategy.privateRegistry + "/" + imageName
        if err := d.pullImage(ctx, registryImage); err == nil {
            // 拉取成功，重新打标签为原名
            d.cli.ImageTag(ctx, registryImage, imageName)
            return nil
        }
    }
    
    // 4. 最后尝试公共仓库（可选）
    return d.pullImage(ctx, imageName)
}
```

---

## 🔧 数据库层面优化

### 新增镜像管理表（可选）

```sql
CREATE TABLE docker_images (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '镜像名称',
    tag VARCHAR(50) NOT NULL DEFAULT 'latest',
    registry VARCHAR(255) COMMENT '仓库地址',
    size BIGINT COMMENT '镜像大小（字节）',
    digest VARCHAR(100) COMMENT '镜像摘要（SHA256）',
    is_preloaded BOOLEAN DEFAULT FALSE COMMENT '是否已预加载',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_image (name, tag, registry)
);
```

### 题目表关联镜像

```sql
-- 修改 challenges 表
ALTER TABLE challenges 
ADD COLUMN image_id VARCHAR(36) COMMENT '关联镜像ID',
ADD FOREIGN KEY (image_id) REFERENCES docker_images(id);

-- 迁移数据
INSERT INTO docker_images (id, name, tag)
SELECT UUID(), image, 'v1.0' FROM challenges;

UPDATE challenges c
SET image_id = (SELECT id FROM docker_images WHERE name = c.image LIMIT 1);
```

---

## 🚀 镜像预加载机制

### 方案：启动时预加载

**实施位置**：`cmd/api/main.go`

```go
func main() {
    // ... 初始化数据库、Redis、DockerManager ...
    
    // 🔧 新增：镜像预加载
    if cfg.Docker.PreloadImages {
        logger.Info(ctx, "开始预加载镜像...")
        preloadImages(ctx, dockerManager, repo)
    }
    
    // 启动服务器...
}

func preloadImages(ctx context.Context, mgr *docker.DockerHostManager, repo *db.Repository) {
    // 1. 获取所有已发布题目的镜像列表
    challenges, _ := repo.GetPublishedChallenges(ctx)
    
    imageSet := make(map[string]bool)
    for _, ch := range challenges {
        imageSet[ch.Image] = true
    }
    
    // 2. 获取所有 Docker 主机
    hosts, _ := repo.GetEnabledDockerHosts(ctx)
    
    // 3. 并发预加载到每个主机
    for _, host := range hosts {
        go func(h *model.DockerHost) {
            client, _ := mgr.GetOrCreateClient(ctx, h)
            
            for image := range imageSet {
                if err := client.EnsureImage(ctx, image); err != nil {
                    logger.Warn(ctx, "镜像预加载失败", 
                        "host", h.Name, "image", image, "error", err)
                } else {
                    logger.Info(ctx, "镜像预加载成功", 
                        "host", h.Name, "image", image)
                }
            }
        }(host)
    }
}
```

---

## 📦 镜像打包和分发流程

### 1. 镜像打包

```bash
# 构建镜像
docker build -t ctf-web-xss:v1.0 ./challenges/web-xss/

# 保存为 tar 文件
docker save -o ctf-web-xss-v1.0.tar ctf-web-xss:v1.0

# 压缩（可选）
gzip ctf-web-xss-v1.0.tar
```

### 2. 批量导入脚本

创建 `scripts/import_images.sh`：

```bash
#!/bin/bash

IMAGES_DIR="/path/to/images"

for tar_file in $IMAGES_DIR/*.tar; do
    echo "正在导入: $tar_file"
    docker load -i "$tar_file"
done

echo "✅ 所有镜像导入完成"
docker images
```

### 3. 远程主机同步

```bash
#!/bin/bash

REMOTE_HOST="192.168.1.100"
IMAGES_DIR="/path/to/images"

# 传输镜像文件
scp $IMAGES_DIR/*.tar root@$REMOTE_HOST:/tmp/

# 远程导入
ssh root@$REMOTE_HOST << 'EOF'
for tar_file in /tmp/*.tar; do
    docker load -i "$tar_file"
    rm "$tar_file"
done
EOF
```

---

## ✅ 推荐实施路径

### 阶段 1：快速优化（1-2 小时）

1. ✅ 使用 `docker load` 导入本地镜像
2. ✅ 修改题目数据库 `image` 字段为本地镜像名
3. ✅ 代码修改：添加本地镜像检查，移除自动拉取

### 阶段 2：中期优化（1 天）

1. ✅ 搭建私有 Docker Registry
2. ✅ 推送所有镜像到私有仓库
3. ✅ 配置远程 Docker 主机信任私有仓库
4. ✅ 实现智能拉取逻辑（优先本地，回退私有仓库）

### 阶段 3：长期优化（1-2 天）

1. ✅ 新增 `docker_images` 表
2. ✅ 管理后台支持镜像管理界面
3. ✅ 实现启动时镜像预加载
4. ✅ 镜像版本管理和更新机制

---

## 💡 最佳实践建议

### 1. 镜像命名规范

```
格式: <registry>/<category>-<name>:<version>

示例:
- localhost:5000/web-xss:v1.0
- localhost:5000/pwn-buffer:v2.3
- localhost:5000/crypto-rsa:latest
```

### 2. 镜像清理策略

```bash
# 定期清理未使用的镜像（避免磁盘占满）
docker image prune -a --filter "until=168h"  # 清理7天前的镜像
```

### 3. 镜像安全扫描

```bash
# 使用 Trivy 扫描镜像漏洞
trivy image ctf-web-xss:v1.0
```

---

## 📊 性能对比

| 指标 | 公网拉取 | 私有仓库 | 本地镜像 |
|------|---------|---------|---------|
| 首次启动时间 | 30-120秒 | 5-15秒 | 2-3秒 ✅ |
| 网络依赖 | ❌ 必须 | ⚠️ 局域网 | ✅ 无 |
| 镜像可控性 | ❌ 低 | ✅ 高 | ✅ 高 |
| 维护成本 | ✅ 低 | ⚠️ 中 | ✅ 低 |

---

需要我帮您选择具体方案并提供实施步骤吗？
