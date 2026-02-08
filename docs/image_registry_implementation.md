# 私有镜像仓库实施方案

## 📋 需求回顾

**用户选择**：
- 方案：私有镜像仓库（Registry）
- 部署位置：本地 Mac
- 镜像规模：1个测试镜像，未来约10个，<1GB/个
- 后台功能：基础（下拉选择镜像 + 查看列表）
- 数据库：新增 `docker_images` 表
- 预加载：系统启动时自动同步到所有 Docker 主机

---

## 🎯 方案概览

```
┌──────────────────────────────────────┐
│  本地 Mac                             │
│  ├─ Docker Registry (localhost:5000) │
│  │  └─ 镜像仓库                       │
│  └─ Cyber Range Backend              │
└──────────────────────────────────────┘
         ↓ 推送/拉取 (HTTP)
┌──────────────────────────────────────┐
│  远程 Docker 服务器                   │
│  └─ 自动拉取镜像                      │
└──────────────────────────────────────┘
```

---

## 📊 数据库设计

### 新增表：docker_images

```sql
CREATE TABLE docker_images (
    id VARCHAR(36) PRIMARY KEY COMMENT '镜像ID (UUID)',
    name VARCHAR(255) NOT NULL COMMENT '镜像名称 (例如: web-xss)',
    tag VARCHAR(50) NOT NULL DEFAULT 'latest' COMMENT '镜像标签/版本',
    registry VARCHAR(255) DEFAULT 'localhost:5000' COMMENT '仓库地址',
    
    -- 镜像元数据
    size BIGINT COMMENT '镜像大小（字节）',
    digest VARCHAR(100) COMMENT '镜像摘要 (SHA256)',
    architecture VARCHAR(20) DEFAULT 'amd64' COMMENT '架构',
    
    -- 状态管理
    is_available BOOLEAN DEFAULT TRUE COMMENT '是否可用',
    last_sync_at TIMESTAMP COMMENT '最后同步时间',
    
    -- 描述信息
    description TEXT COMMENT '镜像描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY idx_image_unique (name, tag, registry),
    INDEX idx_availability (is_available)
) COMMENT='Docker镜像管理表';
```

### 修改表：challenges

```sql
ALTER TABLE challenges 
ADD COLUMN image_id VARCHAR(36) COMMENT '关联镜像ID',
ADD INDEX idx_image_id (image_id),
ADD CONSTRAINT fk_challenge_image 
    FOREIGN KEY (image_id) REFERENCES docker_images(id) 
    ON DELETE SET NULL;

-- 保留原有 image 字段用于临时兼容
-- 后续可以删除
```

---

## 🏗️ 后端实施

### 1. Model 层

#### [NEW] internal/model/docker_image.go

```go
package model

import "time"

type DockerImage struct {
    ID           string    `gorm:"primaryKey;size:36" json:"id"`
    Name         string    `gorm:"size:255;not null" json:"name"`
    Tag          string    `gorm:"size:50;not null;default:latest" json:"tag"`
    Registry     string    `gorm:"size:255;default:localhost:5000" json:"registry"`
    
    Size         int64     `gorm:"bigint" json:"size"`
    Digest       string    `gorm:"size:100" json:"digest"`
    Architecture string    `gorm:"size:20;default:amd64" json:"architecture"`
    
    IsAvailable  bool      `gorm:"default:true" json:"is_available"`
    LastSyncAt   *time.Time `gorm:"type:timestamp" json:"last_sync_at"`
    
    Description  string    `gorm:"type:text" json:"description"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DockerImage) TableName() string {
    return "docker_images"
}

// GetFullName 返回完整镜像名 (registry/name:tag)
func (img *DockerImage) GetFullName() string {
    return fmt.Sprintf("%s/%s:%s", img.Registry, img.Name, img.Tag)
}

// GetShortName 返回简短名称 (name:tag)
func (img *DockerImage) GetShortName() string {
    return fmt.Sprintf("%s:%s", img.Name, img.Tag)
}
```

---

### 2. Repository 层

#### [MODIFY] internal/infra/db/repository.go

新增镜像管理方法：

```go
// ========== Docker Images ==========

// GetAllImages 获取所有镜像
func (r *Repository) GetAllImages(ctx context.Context) ([]*model.DockerImage, error) {
    var images []*model.DockerImage
    if err := r.db.WithContext(ctx).
        Where("is_available = ?", true).
        Order("created_at DESC").
        Find(&images).Error; err != nil {
        return nil, err
    }
    return images, nil
}

// GetImageByID 根据ID获取镜像
func (r *Repository) GetImageByID(ctx context.Context, id string) (*model.DockerImage, error) {
    var img model.DockerImage
    if err := r.db.WithContext(ctx).First(&img, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &img, nil
}

// GetImageByName 根据名称和标签获取镜像
func (r *Repository) GetImageByName(ctx context.Context, name, tag string) (*model.DockerImage, error) {
    var img model.DockerImage
    if err := r.db.WithContext(ctx).
        Where("name = ? AND tag = ?", name, tag).
        First(&img).Error; err != nil {
        return nil, err
    }
    return &img, nil
}

// CreateImage 创建镜像记录
func (r *Repository) CreateImage(ctx context.Context, img *model.DockerImage) error {
    return r.db.WithContext(ctx).Create(img).Error
}

// UpdateImage 更新镜像记录
func (r *Repository) UpdateImage(ctx context.Context, img *model.DockerImage) error {
    return r.db.WithContext(ctx).Save(img).Error
}

// DeleteImage 删除镜像记录
func (r *Repository) DeleteImage(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.DockerImage{}, "id = ?", id).Error
}
```

---

### 3. Service 层

#### [NEW] internal/service/image_service.go

```go
package service

import (
    "context"
    "cyber-range/internal/infra/db"
    "cyber-range/internal/infra/docker"
    "cyber-range/internal/model"
    "cyber-range/pkg/logger"
    "fmt"
    "time"
)

type ImageService struct {
    repo          *db.Repository
    dockerManager *docker.DockerHostManager
}

func NewImageService(repo *db.Repository, dockerManager *docker.DockerHostManager) *ImageService {
    return &ImageService{
        repo:          repo,
        dockerManager: dockerManager,
    }
}

// ListImages 获取所有可用镜像
func (s *ImageService) ListImages(ctx context.Context) ([]*model.DockerImage, error) {
    return s.repo.GetAllImages(ctx)
}

// RegisterImage 注册新镜像（管理员从本地Docker导入后调用）
func (s *ImageService) RegisterImage(ctx context.Context, name, tag, description string) (*model.DockerImage, error) {
    // 检查镜像是否已注册
    existing, _ := s.repo.GetImageByName(ctx, name, tag)
    if existing != nil {
        return nil, fmt.Errorf("镜像已存在: %s:%s", name, tag)
    }
    
    // 创建镜像记录
    img := &model.DockerImage{
        ID:          generateID(),
        Name:        name,
        Tag:         tag,
        Registry:    "localhost:5000",
        IsAvailable: true,
        Description: description,
    }
    
    if err := s.repo.CreateImage(ctx, img); err != nil {
        return nil, fmt.Errorf("注册镜像失败: %w", err)
    }
    
    logger.Info(ctx, "镜像注册成功", "image", img.GetShortName())
    return img, nil
}

// PreloadImages 预加载所有镜像到指定主机
func (s *ImageService) PreloadImages(ctx context.Context, hostID string) error {
    // 获取主机配置
    host, err := s.repo.GetDockerHostByID(ctx, hostID)
    if err != nil {
        return fmt.Errorf("主机不存在: %w", err)
    }
    
    // 获取所有镜像
    images, err := s.repo.GetAllImages(ctx)
    if err != nil {
        return fmt.Errorf("获取镜像列表失败: %w", err)
    }
    
    // 获取Docker客户端
    client, err := s.dockerManager.GetOrCreateClient(ctx, host)
    if err != nil {
        return fmt.Errorf("连接主机失败: %w", err)
    }
    
    // 逐个拉取镜像
    for _, img := range images {
        fullName := img.GetFullName()
        logger.Info(ctx, "开始拉取镜像", "host", host.Name, "image", fullName)
        
        if err := client.EnsureImage(ctx, fullName); err != nil {
            logger.Warn(ctx, "镜像拉取失败", "image", fullName, "error", err)
            continue
        }
        
        // 更新同步时间
        now := time.Now()
        img.LastSyncAt = &now
        s.repo.UpdateImage(ctx, img)
        
        logger.Info(ctx, "镜像拉取成功", "host", host.Name, "image", fullName)
    }
    
    return nil
}

// PreloadAllImages 预加载所有镜像到所有已启用主机
func (s *ImageService) PreloadAllImages(ctx context.Context) error {
    hosts, err := s.repo.GetEnabledDockerHosts(ctx)
    if err != nil {
        return err
    }
    
    logger.Info(ctx, "开始预加载镜像", "host_count", len(hosts))
    
    for _, host := range hosts {
        if err := s.PreloadImages(ctx, host.ID); err != nil {
            logger.Warn(ctx, "主机预加载失败", "host", host.Name, "error", err)
        }
    }
    
    logger.Info(ctx, "镜像预加载完成")
    return nil
}
```

---

### 4. Docker Client 扩展

#### [MODIFY] internal/infra/docker/client.go

新增镜像检查方法：

```go
// EnsureImage 确保镜像存在（不存在则拉取）
func (d *DockerClient) EnsureImage(ctx context.Context, imageName string) error {
    // 1. 检查本地是否已有
    _, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
    if err == nil {
        logger.Debug(ctx, "镜像已存在", "image", imageName)
        return nil
    }
    
    // 2. 从仓库拉取
    logger.Info(ctx, "开始拉取镜像", "image", imageName)
    
    reader, err := d.cli.ImagePull(ctx, imageName, image.PullOptions{})
    if err != nil {
        return fmt.Errorf("镜像拉取失败: %w", err)
    }
    defer reader.Close()
    
    // 等待拉取完成（带超时）
    pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
    defer cancel()
    
    _, err = io.Copy(io.Discard, reader)
    if err != nil {
        return fmt.Errorf("镜像下载失败: %w", err)
    }
    
    return nil
}

// HasLocalImage 检查本地是否有镜像
func (d *DockerClient) HasLocalImage(ctx context.Context, imageName string) bool {
    _, _, err := d.cli.ImageInspectWithRaw(ctx, imageName)
    return err == nil
}
```

#### [MODIFY] StartContainer 方法

```go
func (d *DockerClient) StartContainer(ctx context.Context, imageName string, envVars []string) (string, int, error) {
    // 🔧 优化：确保镜像存在
    if err := d.EnsureImage(ctx, imageName); err != nil {
        return "", 0, fmt.Errorf("镜像准备失败: %w", err)
    }
    
    // 2. 分配端口
    allocatedPort := d.AllocatePort()
    
    // 3. 创建并启动容器
    // ... 原有逻辑
}
```

---

### 5. API Handler

#### [NEW] internal/api/handlers/image_handler.go

```go
package handlers

import (
    "cyber-range/internal/service"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type ImageHandler struct {
    svc *service.ImageService
}

func NewImageHandler(svc *service.ImageService) *ImageHandler {
    return &ImageHandler{svc: svc}
}

// ListImages 获取镜像列表
// GET /api/admin/images
func (h *ImageHandler) List(c *gin.Context) {
    images, err := h.svc.ListImages(c.Request.Context())
    if err != nil {
        c.PureJSON(http.StatusInternalServerError, APIResponse{
            Code: 500,
            Msg:  "获取镜像列表失败",
        })
        return
    }
    
    c.PureJSON(http.StatusOK, APIResponse{
        Code: 200,
        Msg:  "success",
        Data: images,
    })
}

// RegisterImage 注册镜像
// POST /api/admin/images
func (h *ImageHandler) Register(c *gin.Context) {
    var req struct {
        Name        string `json:"name" binding:"required"`
        Tag         string `json:"tag" binding:"required"`
        Description string `json:"description"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.PureJSON(http.StatusBadRequest, APIResponse{
            Code: 400,
            Msg:  "参数错误",
        })
        return
    }
    
    img, err := h.svc.RegisterImage(c.Request.Context(), req.Name, req.Tag, req.Description)
    if err != nil {
        c.PureJSON(http.StatusBadRequest, APIResponse{
            Code: 400,
            Msg:  err.Error(),
        })
        return
    }
    
    c.PureJSON(http.StatusOK, APIResponse{
        Code: 200,
        Msg:  "镜像注册成功",
        Data: img,
    })
}

// PreloadImages 预加载镜像到所有主机
// POST /api/admin/images/preload
func (h *ImageHandler) Preload(c *gin.Context) {
    if err := h.svc.PreloadAllImages(c.Request.Context()); err != nil {
        c.PureJSON(http.StatusInternalServerError, APIResponse{
            Code: 500,
            Msg:  "预加载失败: " + err.Error(),
        })
        return
    }
    
    c.PureJSON(http.StatusOK, APIResponse{
        Code: 200,
        Msg:  "预加载任务已启动",
    })
}
```

#### [MODIFY] cmd/api/main.go

注册路由和启动预加载：

```go
func main() {
    // ... 初始化代码
    
    // 创建 ImageService
    imageService := service.NewImageService(repo, dockerManager)
    imageHandler := handlers.NewImageHandler(imageService)
    
    // 注册路由
    adminGroup := router.Group("/api/admin")
    adminGroup.Use(middleware.AdminAuth())
    {
        // 镜像管理
        adminGroup.GET("/images", imageHandler.List)
        adminGroup.POST("/images", imageHandler.Register)
        adminGroup.POST("/images/preload", imageHandler.Preload)
    }
    
    // 🔧 启动时自动预加载镜像
    go func() {
        time.Sleep(5 * time.Second) // 等待服务启动
        ctx := context.Background()
        if err := imageService.PreloadAllImages(ctx); err != nil {
            logger.Warn(ctx, "自动预加载失败", "error", err)
        }
    }()
    
    // 启动服务器
    router.Run(":8080")
}
```

---

## 🎨 前端实施

### 1. API 客户端

#### [MODIFY] web/src/api/admin.ts

```typescript
export interface DockerImage {
  id: string;
  name: string;
  tag: string;
  registry: string;
  size?: number;
  digest?: string;
  description?: string;
  created_at: string;
}

// 获取镜像列表
export const getImages = async (): Promise<DockerImage[]> => {
  const response = await adminApi.get('/images');
  return response.data.data;
};

// 注册镜像
export const registerImage = async (data: {
  name: string;
  tag: string;
  description?: string;
}) => {
  const response = await adminApi.post('/images', data);
  return response.data;
};
```

---

### 2. 新建题目页面

#### [MODIFY] web/src/pages/Admin/ChallengeNew/index.tsx

```tsx
import { getImages } from '../../../api/admin';
import type { DockerImage } from '../../../api/admin';

const ChallengeNew: React.FC = () => {
  const [images, setImages] = useState<DockerImage[]>([]);
  const [loadingImages, setLoadingImages] = useState(false);
  
  useEffect(() => {
    fetchImages();
  }, []);
  
  const fetchImages = async () => {
    setLoadingImages(true);
    try {
      const data = await getImages();
      setImages(data);
    } catch (error) {
      message.error('加载镜像列表失败');
    } finally {
      setLoadingImages(false);
    }
  };
  
  return (
    <Form>
      {/* ... 其他字段 ... */}
      
      <Form.Item
        name="image_id"
        label="Docker 镜像"
        rules={[{ required: true, message: '请选择镜像' }]}
      >
        <Select
          placeholder="选择镜像"
          loading={loadingImages}
          showSearch
          optionFilterProp="children"
        >
          {images.map(img => (
            <Select.Option key={img.id} value={img.id}>
              {img.name}:{img.tag}
              {img.description && ` - ${img.description}`}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>
    </Form>
  );
};
```

---

### 3. 镜像列表页面（可选）

#### [NEW] web/src/pages/Admin/Images/index.tsx

```tsx
import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, message } from 'antd';
import { getImages } from '../../../api/admin';
import type { DockerImage } from '../../../api/admin';

const ImagesPage: React.FC = () => {
  const [images, setImages] = useState<DockerImage[]>([]);
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    loadImages();
  }, []);
  
  const loadImages = async () => {
    setLoading(true);
    try {
      const data = await getImages();
      setImages(data);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };
  
  const columns = [
    {
      title: '镜像名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: DockerImage) => (
        <span><code>{name}:{record.tag}</code></span>
      ),
    },
    {
      title: '仓库',
      dataIndex: 'registry',
      key: 'registry',
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      render: (size?: number) => size ? `${(size / 1024 / 1024).toFixed(2)} MB` : '-',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
  ];
  
  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => message.info('请使用命令行导入镜像')}>
          导入镜像
        </Button>
        <Button onClick={loadImages}>刷新</Button>
      </Space>
      
      <Table
        dataSource={images}
        columns={columns}
        rowKey="id"
        loading={loading}
      />
    </div>
  );
};

export default ImagesPage;
```

---

## 🚀 部署配置

### 1. Registry 启动脚本

#### [NEW] scripts/start-registry.sh

```bash
#!/bin/bash

echo "🐳 启动本地 Docker Registry..."

# 创建数据目录
mkdir -p ~/cyber-range-registry

# 启动 Registry 容器
docker run -d \
  --name cyber-range-registry \
  --restart=always \
  -p 5000:5000 \
  -v ~/cyber-range-registry:/var/lib/registry \
  registry:2

echo "✅ Registry 已启动在 http://localhost:5000"
echo "📊 查看镜像列表: curl http://localhost:5000/v2/_catalog"
```

---

### 2. 镜像导入流程

#### [NEW] scripts/import-image.sh

```bash
#!/bin/bash

if [ -z "$1" ]; then
    echo "用法: ./import-image.sh <镜像tar文件路径> [镜像名称] [标签]"
    exit 1
fi

TAR_FILE=$1
IMAGE_NAME=${2:-"challenge"}
IMAGE_TAG=${3:-"latest"}

echo "📦 正在导入镜像..."

# 1. 加载到本地 Docker
docker load -i "$TAR_FILE"

# 2. 获取导入的镜像实际名称（如果未指定）
if [ "$IMAGE_NAME" == "challenge" ]; then
    LOADED_IMAGE=$(docker images --format "{{.Repository}}:{{.Tag}}" | head -n 1)
    echo "检测到镜像: $LOADED_IMAGE"
else
    LOADED_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
fi

# 3. 重新打标签
docker tag "$LOADED_IMAGE" "localhost:5000/${IMAGE_NAME}:${IMAGE_TAG}"

# 4. 推送到本地 Registry
echo "🚀 推送到 Registry..."
docker push "localhost:5000/${IMAGE_NAME}:${IMAGE_TAG}"

echo "✅ 镜像导入完成！"
echo "📝 镜像名称: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "🔗 Registry 地址: localhost:5000/${IMAGE_NAME}:${IMAGE_TAG}"
echo ""
echo "下一步："
echo "1. 在管理后台注册镜像（POST /api/admin/images）"
echo "2. 创建题目时选择该镜像"
```

---

### 3. 远程主机配置

#### [NEW] docs/remote-docker-config.md

```markdown
# 配置远程 Docker 主机信任本地 Registry

## 步骤

1. SSH 登录远程服务器
2. 编辑 Docker 配置文件 `/etc/docker/daemon.json`
3. 添加以下内容（替换 `192.168.1.100` 为您本地 Mac 的 IP）:

\`\`\`json
{
  "insecure-registries": ["192.168.1.100:5000"]
}
\`\`\`

4. 重启 Docker:
\`\`\`bash
sudo systemctl restart docker
\`\`\`

5. 测试连接:
\`\`\`bash
docker pull 192.168.1.100:5000/web-xss:v1.0
\`\`\`
```

---

## ✅ 验证计划

### 手动验证流程

1. **Registry 部署验证**
   - [ ] 运行 `start-registry.sh`
   - [ ] 访问 `http://localhost:5000/v2/_catalog`
   - [ ] 应返回空列表 `{"repositories":[]}`

2. **镜像导入验证**
   - [ ] 运行 `import-image.sh your-image.tar web-xss v1.0`
   - [ ] 检查 Registry: `curl http://localhost:5000/v2/_catalog`
   - [ ] 应看到 `{"repositories":["web-xss"]}`

3. **数据库迁移验证**
   - [ ] 运行 `go run cmd/migrate/main.go`
   - [ ] 检查 `docker_images` 表已创建
   - [ ] 检查 `challenges` 表有 `image_id` 字段

4. **API 验证**
   - [ ] POST `/api/admin/images` 注册镜像
   - [ ] GET `/api/admin/images` 查看列表
   - [ ] 应返回刚注册的镜像

5. **前端验证**
   - [ ] 打开新建题目页面
   - [ ] 镜像下拉框有数据
   - [ ] 可以选择镜像

6. **自动预加载验证**
   - [ ] 重启后端服务
   - [ ] 查看日志，应显示"开始预加载镜像"
   - [ ] 检查远程 Docker 主机，镜像应已拉取

7. **端到端验证**
   - [ ] 创建题目并选择镜像
   - [ ] 启动题目
   - [ ] 容器成功运行

---

## 📝 实施检查清单

### 数据库
- [ ] 创建 `docker_images` 表
- [ ] 修改 `challenges` 表
- [ ] 编写迁移脚本
- [ ] 编写种子数据

### 后端
- [ ] 创建 `DockerImage` Model
- [ ] 扩展 `Repository`
- [ ] 实现 `ImageService`
- [ ] 修改 `DockerClient`
- [ ] 新增 `ImageHandler`
- [ ] 修改 `main.go` 注册路由
- [ ] 添加启动预加载逻辑

### 前端
- [ ] 修改 `admin.ts` API
- [ ] 修改新建题目页面
- [ ] 修改编辑题目页面
- [ ] （可选）创建镜像列表页面

### 部署
- [ ] 编写 Registry 启动脚本
- [ ] 编写镜像导入脚本
- [ ] 编写远程主机配置文档

### 测试
- [ ] Registry 部署测试
- [ ] 镜像导入测试
- [ ] API 功能测试
- [ ] 前端界面测试
- [ ] 自动预加载测试
- [ ] 端到端测试

---

## ⏱️ 时间估算

| 阶段 | 预计时间 |
|------|---------|
| 数据库设计 | 30 分钟 |
| 后端 Model + Repo | 1 小时 |
| 后端 Service + Handler | 1.5 小时 |
| 前端界面 | 1 小时 |
| 部署脚本 | 30 分钟 |
| 测试验证 | 1 小时 |
| **总计** | **5.5 小时** |

---

需要开始实施吗？
