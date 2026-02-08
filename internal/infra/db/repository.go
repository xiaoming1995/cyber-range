package db

import (
	"context"
	"cyber-range/internal/model"
	"fmt"

	"gorm.io/gorm"
)

// Repository 数据访问层接口实现
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB 返回底层 gorm.DB 实例（用于直接查询）
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// ===== Docker 主机管理 =====

// GetDockerHostByID 根据 ID 获取 Docker 主机
func (r *Repository) GetDockerHostByID(ctx context.Context, id string) (*model.DockerHost, error) {
	var host model.DockerHost
	if err := r.db.WithContext(ctx).First(&host, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Docker 主机不存在: %s", id)
		}
		return nil, fmt.Errorf("获取 Docker 主机失败: %w", err)
	}
	return &host, nil
}

// GetDefaultDockerHost 获取默认 Docker 主机
func (r *Repository) GetDefaultDockerHost(ctx context.Context) (*model.DockerHost, error) {
	var host model.DockerHost
	if err := r.db.WithContext(ctx).Where("is_default = ? AND enabled = ?", true, true).First(&host).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到启用的默认 Docker 主机")
		}
		return nil, fmt.Errorf("获取默认 Docker 主机失败: %w", err)
	}
	return &host, nil
}

// ListDockerHosts 获取 Docker 主机列表
func (r *Repository) ListDockerHosts(ctx context.Context, enabledOnly bool) ([]*model.DockerHost, error) {
	var hosts []*model.DockerHost
	query := r.db.WithContext(ctx)

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	if err := query.Order("is_default DESC, created_at ASC").Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("获取 Docker 主机列表失败: %w", err)
	}
	return hosts, nil
}

// CreateDockerHost 创建新的 Docker 主机
func (r *Repository) CreateDockerHost(ctx context.Context, host *model.DockerHost) error {
	// 如果设置为默认主机，先取消其他主机的默认状态
	if host.IsDefault {
		if err := r.db.WithContext(ctx).Model(&model.DockerHost{}).
			Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("更新原默认主机状态失败: %w", err)
		}
	}

	if err := r.db.WithContext(ctx).Create(host).Error; err != nil {
		return fmt.Errorf("创建 Docker 主机失败: %w", err)
	}
	return nil
}

// UpdateDockerHost 更新 Docker 主机配置
func (r *Repository) UpdateDockerHost(ctx context.Context, host *model.DockerHost) error {
	// 如果设置为默认主机，先取消其他主机的默认状态
	if host.IsDefault {
		if err := r.db.WithContext(ctx).Model(&model.DockerHost{}).
			Where("is_default = ? AND id != ?", true, host.ID).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("更新原默认主机状态失败: %w", err)
		}
	}

	if err := r.db.WithContext(ctx).Save(host).Error; err != nil {
		return fmt.Errorf("更新 Docker 主机失败: %w", err)
	}
	return nil
}

// DeleteDockerHost 删除 Docker 主机
func (r *Repository) DeleteDockerHost(ctx context.Context, id string) error {
	// 检查是否有题目关联此主机
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Challenge{}).
		Where("docker_host_id = ?", id).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查关联题目失败: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("无法删除：有 %d 个题目关联此 Docker 主机", count)
	}

	if err := r.db.WithContext(ctx).Delete(&model.DockerHost{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除 Docker 主机失败: %w", err)
	}
	return nil
}

// ToggleDockerHostEnabled 切换 Docker 主机的启用状态
func (r *Repository) ToggleDockerHostEnabled(ctx context.Context, id string) error {
	var host model.DockerHost
	if err := r.db.WithContext(ctx).First(&host, "id = ?", id).Error; err != nil {
		return fmt.Errorf("Docker 主机不存在: %w", err)
	}

	newStatus := !host.Enabled
	if err := r.db.WithContext(ctx).Model(&host).Update("enabled", newStatus).Error; err != nil {
		return fmt.Errorf("更新启用状态失败: %w", err)
	}
	return nil
}

// ===== 题目管理 =====

// GetChallengeByID 根据 ID 获取题目
func (r *Repository) GetChallengeByID(ctx context.Context, id string) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.db.WithContext(ctx).First(&challenge, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("题目不存在: %s", id)
		}
		return nil, fmt.Errorf("获取题目失败: %w", err)
	}
	return &challenge, nil
}

// GetEnabledDockerHosts 获取所有已启用的 Docker 主机
func (r *Repository) GetEnabledDockerHosts(ctx context.Context) ([]*model.DockerHost, error) {
	return r.ListDockerHosts(ctx, true)
}

// GetPublishedChallenges 获取所有已发布的题目
func (r *Repository) GetPublishedChallenges(ctx context.Context) ([]*model.Challenge, error) {
	var challenges []*model.Challenge
	if err := r.db.WithContext(ctx).
		Where("status = ?", "published").
		Find(&challenges).Error; err != nil {
		return nil, fmt.Errorf("获取已发布题目失败: %w", err)
	}
	return challenges, nil
}

// ===== Docker 镜像管理 =====

// GetAllImages 获取所有镜像
func (r *Repository) GetAllImages(ctx context.Context) ([]*model.DockerImage, error) {
	var images []*model.DockerImage
	if err := r.db.WithContext(ctx).
		Where("is_available = ?", true).
		Order("created_at DESC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}
	return images, nil
}

// GetImageByID 根据 ID 获取镜像
func (r *Repository) GetImageByID(ctx context.Context, id string) (*model.DockerImage, error) {
	var img model.DockerImage
	if err := r.db.WithContext(ctx).First(&img, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("镜像不存在: %s", id)
		}
		return nil, fmt.Errorf("获取镜像失败: %w", err)
	}
	return &img, nil
}

// GetImageByName 根据名称和标签获取镜像
func (r *Repository) GetImageByName(ctx context.Context, name, tag string) (*model.DockerImage, error) {
	var img model.DockerImage
	if err := r.db.WithContext(ctx).
		Where("name = ? AND tag = ?", name, tag).
		First(&img).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("镜像不存在: %s:%s", name, tag)
		}
		return nil, fmt.Errorf("获取镜像失败: %w", err)
	}
	return &img, nil
}

// CreateImage 创建镜像记录
func (r *Repository) CreateImage(ctx context.Context, img *model.DockerImage) error {
	if err := r.db.WithContext(ctx).Create(img).Error; err != nil {
		return fmt.Errorf("创建镜像失败: %w", err)
	}
	return nil
}

// UpdateImage 更新镜像记录
func (r *Repository) UpdateImage(ctx context.Context, img *model.DockerImage) error {
	if err := r.db.WithContext(ctx).Save(img).Error; err != nil {
		return fmt.Errorf("更新镜像失败: %w", err)
	}
	return nil
}

// DeleteImage 删除镜像记录
func (r *Repository) DeleteImage(ctx context.Context, id string) error {
	// 检查是否有题目关联此镜像
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Challenge{}).
		Where("image_id = ?", id).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查关联题目失败: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("无法删除：有 %d 个题目关联此镜像", count)
	}

	if err := r.db.WithContext(ctx).Delete(&model.DockerImage{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除镜像失败: %w", err)
	}
	return nil
}
