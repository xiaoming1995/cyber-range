package handlers

import (
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InstanceHandler 实例管理处理器
type InstanceHandler struct {
	repo          *db.Repository
	dockerManager *docker.DockerHostManager
}

// NewInstanceHandler 创建实例处理器
func NewInstanceHandler(repo *db.Repository, dockerManager *docker.DockerHostManager) *InstanceHandler {
	return &InstanceHandler{
		repo:          repo,
		dockerManager: dockerManager,
	}
}

// GetInstanceStats 获取容器实时资源统计
// GET /api/admin/instances/:id/stats
func (h *InstanceHandler) GetInstanceStats(c *gin.Context) {
	ctx := c.Request.Context()
	instanceID := c.Param("id")

	// 1. 从数据库获取实例信息
	var instance model.Instance
	if err := h.repo.DB().WithContext(ctx).First(&instance, "id = ?", instanceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "实例不存在",
		})
		return
	}

	// 2. 检查实例状态
	if instance.Status != "running" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "实例未在运行状态",
		})
		return
	}

	// 3. 获取 Docker 主机配置
	dockerHost, err := h.repo.GetDockerHostByID(ctx, instance.DockerHostID)
	if err != nil {
		logger.Error(ctx, "Docker host not found", "host_id", instance.DockerHostID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Docker 主机配置获取失败",
		})
		return
	}

	// 4. 获取 Docker 客户端
	dockerClient, err := h.dockerManager.GetOrCreateClient(ctx, dockerHost)
	if err != nil {
		logger.Error(ctx, "Failed to get Docker client", "host", dockerHost.Name, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Docker 客户端连接失败",
		})
		return
	}

	// 5. 获取容器统计
	stats, err := dockerClient.GetContainerStats(ctx, instance.ContainerID)
	if err != nil {
		logger.Error(ctx, "Failed to get container stats", "container_id", instance.ContainerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取容器统计失败: " + err.Error(),
		})
		return
	}

	logger.Info(ctx, "Got container stats", "instance_id", instanceID, "cpu", stats.CPUPercent)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}

// GetInstanceLogs 获取容器日志
// GET /api/admin/instances/:id/logs?tail=200
func (h *InstanceHandler) GetInstanceLogs(c *gin.Context) {
	ctx := c.Request.Context()
	instanceID := c.Param("id")
	tailStr := c.DefaultQuery("tail", "200")
	tail, _ := strconv.Atoi(tailStr)

	// 1. 从数据库获取实例信息
	var instance model.Instance
	if err := h.repo.DB().WithContext(ctx).First(&instance, "id = ?", instanceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "实例不存在",
		})
		return
	}

	// 2. 检查实例状态（stopped 状态也可以查看日志）
	if instance.Status == "expired" && instance.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "容器已删除，无法获取日志",
		})
		return
	}

	// 3. 获取 Docker 主机配置
	dockerHost, err := h.repo.GetDockerHostByID(ctx, instance.DockerHostID)
	if err != nil {
		logger.Error(ctx, "Docker host not found", "host_id", instance.DockerHostID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Docker 主机配置获取失败",
		})
		return
	}

	// 4. 获取 Docker 客户端
	dockerClient, err := h.dockerManager.GetOrCreateClient(ctx, dockerHost)
	if err != nil {
		logger.Error(ctx, "Failed to get Docker client", "host", dockerHost.Name, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "Docker 客户端连接失败",
		})
		return
	}

	// 5. 获取容器日志
	logs, err := dockerClient.GetContainerLogs(ctx, instance.ContainerID, tail)
	if err != nil {
		logger.Error(ctx, "Failed to get container logs", "container_id", instance.ContainerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取容器日志失败: " + err.Error(),
		})
		return
	}

	logger.Info(ctx, "Got container logs", "instance_id", instanceID, "tail", tail, "length", len(logs))

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"logs":         logs,
			"container_id": instance.ContainerID,
		},
	})
}
