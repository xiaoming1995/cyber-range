package service

import (
	"context"
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	redisRepo "cyber-range/internal/infra/redis"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"time"

	"gorm.io/gorm"
)

// Reaper manages automatic cleanup of expired instances
type Reaper struct {
	dockerManager *docker.DockerHostManager
	repo          *db.Repository
	gormDB        *gorm.DB
	ticker        *time.Ticker
	stopChan      chan struct{}
}

func NewReaper(dockerManager *docker.DockerHostManager, repo *db.Repository, gormDB *gorm.DB) *Reaper {
	return &Reaper{
		dockerManager: dockerManager,
		repo:          repo,
		gormDB:        gormDB,
		ticker:        time.NewTicker(1 * time.Minute),
		stopChan:      make(chan struct{}),
	}
}

// Start launches The Reaper goroutine
func (r *Reaper) Start(ctx context.Context) {
	logger.Info(ctx, "The Reaper started: scanning for expired instances every 1 minute")

	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.reapExpiredInstances(ctx)
			case <-r.stopChan:
				logger.Info(ctx, "The Reaper stopped")
				return
			}
		}
	}()
}

// Stop gracefully shuts down The Reaper
func (r *Reaper) Stop() {
	r.ticker.Stop()
	close(r.stopChan)
}

// reapExpiredInstances scans Redis for expired instances and forcefully kills them
func (r *Reaper) reapExpiredInstances(ctx context.Context) {
	logger.Debug(ctx, "Reaper: starting scan for expired instances...")

	expiredIDs, err := redisRepo.GetExpiredInstances(ctx)
	if err != nil {
		logger.Error(ctx, "Reaper failed to get expired instances", "error", err)
		return
	}

	logger.Debug(ctx, "Reaper: scan complete", "expired_count", len(expiredIDs), "expired_ids", expiredIDs)

	if len(expiredIDs) == 0 {
		return
	}

	logger.Info(ctx, "Reaper found expired instances", "count", len(expiredIDs))

	for _, instanceID := range expiredIDs {
		logger.Info(ctx, "Reaper: processing expired instance", "instance_id", instanceID)
		r.killInstance(ctx, instanceID)
	}
}

// killInstance forcefully stops a container and cleans up state
func (r *Reaper) killInstance(ctx context.Context, instanceID string) {
	// 从 Redis 获取实例数据（可能已过期被清除）
	instData, _ := redisRepo.GetInstance(ctx, instanceID)

	// 从数据库读取权威实例信息（包含 docker_host_id）
	var instance model.Instance
	if err := r.gormDB.WithContext(ctx).First(&instance, "id = ?", instanceID).Error; err != nil {
		logger.Warn(ctx, "Reaper: instance not found in DB, cleaning up ZSET only",
			"instance_id", instanceID, "error", err)
		// 数据库无记录，仅清理 ZSET 中的残留
		redisRepo.RemoveFromExpiredSet(ctx, instanceID)
		return
	}

	// 优先使用数据库中的数据，Redis 作为备用校验
	containerID := instance.ContainerID
	userID := instance.UserID
	if len(instData) > 0 && instData["container_id"] != "" {
		containerID = instData["container_id"]
	}

	// 获取 Docker 主机配置
	dockerHost, err := r.repo.GetDockerHostByID(ctx, instance.DockerHostID)
	if err != nil {
		logger.Warn(ctx, "Reaper: Docker host not found",
			"instance_id", instanceID,
			"docker_host_id", instance.DockerHostID,
			"error", err)
		// 清理 Redis，即使无法停止容器
		redisRepo.DeleteInstance(ctx, instanceID, userID)
		r.updateInstanceStatus(ctx, instanceID, "expired")
		return
	}

	// 获取 Docker 客户端
	dockerClient, err := r.dockerManager.GetOrCreateClient(ctx, dockerHost)
	if err != nil {
		logger.Warn(ctx, "Reaper: failed to get Docker client",
			"docker_host", dockerHost.Name,
			"error", err)
		// 清理 Redis，即使无法停止容器
		redisRepo.DeleteInstance(ctx, instanceID, userID)
		r.updateInstanceStatus(ctx, instanceID, "expired")
		return
	}

	// Force kill container
	if err := dockerClient.StopContainer(ctx, containerID); err != nil {
		logger.Warn(ctx, "Reaper: failed to stop container",
			"container_id", containerID,
			"docker_host", dockerHost.Name,
			"error", err)
	}

	// Clean up Redis
	if err := redisRepo.DeleteInstance(ctx, instanceID, userID); err != nil {
		logger.Error(ctx, "Reaper: failed to delete from Redis", "instance_id", instanceID, "error", err)
	}

	// Update DB
	r.updateInstanceStatus(ctx, instanceID, "expired")

	logger.Info(ctx, "Reaper: successfully killed expired instance",
		"instance_id", instanceID,
		"container_id", containerID,
		"docker_host", dockerHost.Name)
}

// updateInstanceStatus 更新实例状态
func (r *Reaper) updateInstanceStatus(ctx context.Context, instanceID, status string) {
	r.gormDB.WithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]interface{}{
			"status": status,
		})
}
