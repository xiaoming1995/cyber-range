package handlers

import (
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DockerHostHandler struct {
	repo          *db.Repository
	dockerManager *docker.DockerHostManager
}

func NewDockerHostHandler(repo *db.Repository, dockerManager *docker.DockerHostManager) *DockerHostHandler {
	return &DockerHostHandler{
		repo:          repo,
		dockerManager: dockerManager,
	}
}

// ListDockerHosts 获取 Docker 主机列表
// GET /api/admin/docker-hosts
func (h *DockerHostHandler) ListDockerHosts(c *gin.Context) {
	ctx := c.Request.Context()

	hosts, err := h.repo.ListDockerHosts(ctx, false) // 获取所有主机
	if err != nil {
		logger.Error(ctx, "Failed to list Docker hosts", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取 Docker 主机列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": hosts,
	})
}

// CreateDockerHost 创建新的 Docker 主机
// POST /api/admin/docker-hosts
func (h *DockerHostHandler) CreateDockerHost(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Host         string  `json:"host" binding:"required"`
		TLSVerify    bool    `json:"tls_verify"`
		CertPath     string  `json:"cert_path"`
		PortRangeMin int     `json:"port_range_min" binding:"required,min=1024,max=65535"`
		PortRangeMax int     `json:"port_range_max" binding:"required,min=1024,max=65535"`
		MemoryLimit  int64   `json:"memory_limit" binding:"required,min=67108864"` // 最小 64MB
		CPULimit     float64 `json:"cpu_limit" binding:"required,min=0.1,max=128"` // 最小 0.1 核
		Enabled      bool    `json:"enabled"`
		IsDefault    bool    `json:"is_default"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证端口范围
	if req.PortRangeMin >= req.PortRangeMax {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "端口范围最小值必须小于最大值",
		})
		return
	}

	host := &model.DockerHost{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Host:         req.Host,
		TLSVerify:    req.TLSVerify,
		CertPath:     req.CertPath,
		PortRangeMin: req.PortRangeMin,
		PortRangeMax: req.PortRangeMax,
		MemoryLimit:  req.MemoryLimit,
		CPULimit:     req.CPULimit,
		Enabled:      req.Enabled,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.repo.CreateDockerHost(ctx, host); err != nil {
		logger.Error(ctx, "Failed to create Docker host", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建 Docker 主机失败: " + err.Error(),
		})
		return
	}

	logger.Info(ctx, "Docker host created", "host_id", host.ID, "name", host.Name)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": host,
	})
}

// UpdateDockerHost 更新 Docker 主机配置
// PUT /api/admin/docker-hosts/:id
func (h *DockerHostHandler) UpdateDockerHost(c *gin.Context) {
	ctx := c.Request.Context()
	hostID := c.Param("id")

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Host         string  `json:"host" binding:"required"`
		TLSVerify    bool    `json:"tls_verify"`
		CertPath     string  `json:"cert_path"`
		PortRangeMin int     `json:"port_range_min" binding:"required,min=1024,max=65535"`
		PortRangeMax int     `json:"port_range_max" binding:"required,min=1024,max=65535"`
		MemoryLimit  int64   `json:"memory_limit" binding:"required,min=67108864"`
		CPULimit     float64 `json:"cpu_limit" binding:"required,min=0.1,max=128"`
		Enabled      bool    `json:"enabled"`
		IsDefault    bool    `json:"is_default"`
		Description  string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证端口范围
	if req.PortRangeMin >= req.PortRangeMax {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "端口范围最小值必须小于最大值",
		})
		return
	}

	// 获取现有主机信息
	existingHost, err := h.repo.GetDockerHostByID(ctx, hostID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Docker 主机不存在",
		})
		return
	}

	// 更新字段
	existingHost.Name = req.Name
	existingHost.Host = req.Host
	existingHost.TLSVerify = req.TLSVerify
	existingHost.CertPath = req.CertPath
	existingHost.PortRangeMin = req.PortRangeMin
	existingHost.PortRangeMax = req.PortRangeMax
	existingHost.MemoryLimit = req.MemoryLimit
	existingHost.CPULimit = req.CPULimit
	existingHost.Enabled = req.Enabled
	existingHost.IsDefault = req.IsDefault
	existingHost.Description = req.Description
	existingHost.UpdatedAt = time.Now()

	if err := h.repo.UpdateDockerHost(ctx, existingHost); err != nil {
		logger.Error(ctx, "Failed to update Docker host", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新 Docker 主机失败: " + err.Error(),
		})
		return
	}

	// 清除客户端缓存，强制下次使用时重新创建
	h.dockerManager.RemoveClient(hostID)

	logger.Info(ctx, "Docker host updated", "host_id", hostID, "name", req.Name)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": existingHost,
	})
}

// DeleteDockerHost 删除 Docker 主机
// DELETE /api/admin/docker-hosts/:id
func (h *DockerHostHandler) DeleteDockerHost(c *gin.Context) {
	ctx := c.Request.Context()
	hostID := c.Param("id")

	if err := h.repo.DeleteDockerHost(ctx, hostID); err != nil {
		logger.Error(ctx, "Failed to delete Docker host", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 清除客户端缓存
	h.dockerManager.RemoveClient(hostID)

	logger.Info(ctx, "Docker host deleted", "host_id", hostID)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// TestDockerHost 测试 Docker 主机连接
// POST /api/admin/docker-hosts/:id/test
func (h *DockerHostHandler) TestDockerHost(c *gin.Context) {
	ctx := c.Request.Context()
	hostID := c.Param("id")

	host, err := h.repo.GetDockerHostByID(ctx, hostID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "Docker 主机不存在",
		})
		return
	}

	// 测试连接
	if err := h.dockerManager.Ping(ctx, host); err != nil {
		logger.Warn(ctx, "Docker host connection test failed", "host_id", hostID, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "连接测试失败",
			"data": gin.H{
				"success": false,
				"error":   err.Error(),
			},
		})
		return
	}

	logger.Info(ctx, "Docker host connection test succeeded", "host_id", hostID)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "连接测试成功",
		"data": gin.H{
			"success": true,
		},
	})
}

// ToggleDockerHost 启用/禁用 Docker 主机
// POST /api/admin/docker-hosts/:id/toggle
func (h *DockerHostHandler) ToggleDockerHost(c *gin.Context) {
	ctx := c.Request.Context()
	hostID := c.Param("id")

	if err := h.repo.ToggleDockerHostEnabled(ctx, hostID); err != nil {
		logger.Error(ctx, "Failed to toggle Docker host", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "切换状态失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的主机信息
	host, _ := h.repo.GetDockerHostByID(ctx, hostID)

	logger.Info(ctx, "Docker host toggled", "host_id", hostID, "enabled", host.Enabled)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "状态已更新",
		"data": host,
	})
}
