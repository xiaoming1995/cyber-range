package handlers

import (
	"cyber-range/internal/service"
	"net/http"
	"os"
	"path/filepath"

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

// SyncFromRegistry 从 Registry 同步镜像
// POST /api/admin/images/sync
func (h *ImageHandler) Sync(c *gin.Context) {
	var req struct {
		RegistryURL string `json:"registry_url"`
	}

	// 使用默认 Registry 如果未指定
	registryURL := "http://localhost:5000"
	if err := c.ShouldBindJSON(&req); err == nil && req.RegistryURL != "" {
		registryURL = req.RegistryURL
	}

	count, err := h.svc.SyncFromRegistry(c.Request.Context(), registryURL)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "同步失败: " + err.Error(),
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "同步完成",
		Data: gin.H{
			"synced_count": count,
			"registry_url": registryURL,
		},
	})
}

// Upload 上传并导入镜像
// POST /api/admin/images/upload
// 接收 multipart/form-data，字段名: file
func (h *ImageHandler) Upload(c *gin.Context) {
	// 1. 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "请上传镜像文件",
		})
		return
	}

	// 2. 检查文件扩展名
	ext := filepath.Ext(file.Filename)
	if ext != ".tar" && ext != ".gz" {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "仅支持 .tar 或 .tar.gz 格式",
		})
		return
	}

	// 3. 检查文件大小 (限制 2GB)
	const maxSize = 2 << 30 // 2GB
	if file.Size > maxSize {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "文件大小超过 2GB 限制",
		})
		return
	}

	// 4. 保存到项目本地临时目录（避免 macOS 系统临时目录权限问题）
	tempDir := "./tmp/uploads"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "创建临时目录失败: " + err.Error(),
		})
		return
	}
	tempFile := filepath.Join(tempDir, "cyberrange_upload_"+file.Filename)
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "保存文件失败: " + err.Error(),
		})
		return
	}

	// 5. 调用 Service 导入镜像
	result, err := h.svc.ImportFromTar(c.Request.Context(), tempFile)
	if err != nil {
		// 清理临时文件
		os.Remove(tempFile)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "导入失败: " + err.Error(),
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "镜像导入成功",
		Data: result,
	})
}
