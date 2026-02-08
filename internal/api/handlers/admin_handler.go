package handlers

import (
	"cyber-range/internal/model"
	"cyber-range/internal/service"
	"cyber-range/pkg/logger"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminSvc     *service.AdminService
	challengeSvc *service.ChallengeService
	db           interface{} // 用于直接查询实例和提交记录
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(adminSvc *service.AdminService, challengeSvc *service.ChallengeService, db interface{}) *AdminHandler {
	return &AdminHandler{
		adminSvc:     adminSvc,
		challengeSvc: challengeSvc,
		db:           db,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
// POST /api/admin/login
func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid request format",
		})
		return
	}

	token, admin, err := h.adminSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.PureJSON(http.StatusUnauthorized, APIResponse{
			Code: 401,
			Msg:  err.Error(),
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: gin.H{
			"token": token,
			"admin": gin.H{
				"id":       admin.ID,
				"username": admin.Username,
				"email":    admin.Email,
				"name":     admin.Name,
			},
		},
	})
}

// CreateChallengeRequest 创建题目请求
type CreateChallengeRequest struct {
	Title           string  `json:"title" binding:"required"`
	DescriptionHtml string  `json:"descriptionHtml"` // 富文本 HTML (允许为空)
	HintHtml        string  `json:"hintHtml"`        // 提示 HTML
	Category        string  `json:"category" binding:"required,oneof=Web Pwn Crypto Reverse Misc web pwn crypto reverse misc"`
	Difficulty      string  `json:"difficulty" binding:"required,oneof=Easy Medium Hard easy medium hard"`
	Image           string  `json:"image"`          // 兼容旧字段，逻辑校验
	ImageID         string  `json:"image_id"`       // 关联镜像ID
	DockerHostID    string  `json:"docker_host_id"` // Docker主机ID
	Port            int     `json:"port" binding:"required"`
	MemoryLimit     int64   `json:"memory_limit"` // 内存限制
	CPULimit        float64 `json:"cpu_limit"`    // CPU限制
	Privileged      bool    `json:"privileged"`   // 特权模式
	Flag            string  `json:"flag"`         // 逻辑校验 (Create 必填, Update 选填)
	Points          int     `json:"points" binding:"required"`
	Status          string  `json:"status"` // published/unpublished
}

// CreateChallenge 创建题目
// POST /api/admin/challenges
func (h *AdminHandler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid request format: " + err.Error(),
		})
		return
	}

	// 手动校验必需字段
	if req.Flag == "" {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Flag 不能为空",
		})
		return
	}

	// 校验镜像：Image 和 ImageID 必须有一个
	if req.Image == "" && req.ImageID == "" {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "必须指定镜像或镜像ID",
		})
		return
	}

	// 输入验证
	if req.Points < 1 || req.Points > 10000 {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "分值必须在 1-10000 之间",
		})
		return
	}

	if req.Port < 1 || req.Port > 65535 {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "端口必须在 1-65535 之间",
		})
		return
	}

	// 默认状态
	if req.Status == "" {
		req.Status = "unpublished"
	}

	// 验证状态值
	if req.Status != "published" && req.Status != "unpublished" {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "状态必须是 published 或 unpublished",
		})
		return
	}

	// 创建题目
	challenge := &model.Challenge{
		ID:           uuid.New().String(),
		Title:        req.Title,
		Description:  req.DescriptionHtml,
		Hint:         req.HintHtml,
		Category:     req.Category,
		Difficulty:   req.Difficulty,
		Image:        req.Image,
		ImageID:      req.ImageID,
		DockerHostID: req.DockerHostID,
		Port:         req.Port,
		MemoryLimit:  req.MemoryLimit,
		CPULimit:     req.CPULimit,
		Privileged:   req.Privileged,
		Flag:         req.Flag,
		Points:       req.Points,
		Status:       req.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	db, ok := h.db.(*gorm.DB)
	if !ok {
		logger.Error(c.Request.Context(), "Database type assertion failed")
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	if err := db.WithContext(c.Request.Context()).Create(challenge).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to create challenge", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "创建失败",
		})
		return
	}

	logger.Info(c.Request.Context(), "Created challenge", "id", challenge.ID, "title", challenge.Title)

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Challenge created successfully",
		Data: challenge,
	})
}

// UpdateChallenge 更新题目
// PUT /api/admin/challenges/:id
func (h *AdminHandler) UpdateChallenge(c *gin.Context) {
	challengeID := c.Param("id")

	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid request format: " + err.Error(),
		})
		return
	}

	// 输入验证
	if req.Points < 1 || req.Points > 10000 {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "分值必须在 1-10000 之间",
		})
		return
	}

	if req.Port < 1 || req.Port > 65535 {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "端口必须在 1-65535 之间",
		})
		return
	}

	db, ok := h.db.(*gorm.DB)
	if !ok {
		logger.Error(c.Request.Context(), "Database type assertion failed")
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	// 检查题目是否存在
	var existing model.Challenge
	if err := db.WithContext(c.Request.Context()).Where("id = ?", challengeID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.PureJSON(http.StatusNotFound, APIResponse{
				Code: 404,
				Msg:  "题目不存在",
			})
			return
		}
		logger.Error(c.Request.Context(), "Failed to query challenge", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "查询失败",
		})
		return
	}

	// 自动填充镜像名称
	if req.ImageID != "" {
		var dockerImage model.DockerImage
		if err := db.WithContext(c.Request.Context()).First(&dockerImage, "id = ?", req.ImageID).Error; err == nil {
			req.Image = dockerImage.GetShortName()
		}
	}

	// 确保 Image 不为空 (如果是更新，且 req.Image 为空，应保留原值还是？)
	// 这里假设如果前端传了 ImageID，req.Image 已经被填充。
	// 如果前端既没传 ImageID 也没传 Image，且是要清除 Image？通常题目必须有 Image。
	// 简单起见，如果 req.Image 为空且没有 ImageID，则不更新 Image 字段（或者报错）
	// 但鉴于 updates map 的构造方式，如果 Image 为空字符串，它会被更新为空。
	// 所以我们得确保 req.Image 有值，或者如果为空则沿用旧值（但这在 PUT 全量更新语义下通常不推荐，不过这里是 partial update 吗？结构体绑定是全量的）

	if req.Image == "" {
		// 如果尝试更新为空，且原本有值，这可能是个错误，或者是前端只传了部分字段？
		// 暂且认为 Image 是必填的
		if existing.Image != "" && req.ImageID == "" {
			req.Image = existing.Image // 保持原值
		} else if req.Image == "" {
			c.PureJSON(http.StatusBadRequest, APIResponse{
				Code: 400,
				Msg:  "镜像名称不能为空",
			})
			return
		}
	}

	// 更新字段
	updates := map[string]interface{}{
		"title":          req.Title,
		"description":    req.DescriptionHtml,
		"hint":           req.HintHtml,
		"category":       req.Category,
		"difficulty":     req.Difficulty,
		"image":          req.Image,
		"image_id":       req.ImageID,
		"docker_host_id": req.DockerHostID,
		"port":           req.Port,
		"memory_limit":   req.MemoryLimit,
		"cpu_limit":      req.CPULimit,
		"privileged":     req.Privileged,
		"flag":           req.Flag,
		"points":         req.Points,
		"updated_at":     time.Now(),
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := db.WithContext(c.Request.Context()).Model(&existing).Updates(updates).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to update challenge", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "更新失败",
		})
		return
	}

	logger.Info(c.Request.Context(), "Updated challenge", "id", challengeID)

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Challenge updated successfully",
	})
}

// DeleteChallenge 删除题目
// DELETE /api/admin/challenges/:id
func (h *AdminHandler) DeleteChallenge(c *gin.Context) {
	challengeID := c.Param("id")

	db, ok := h.db.(*gorm.DB)
	if !ok {
		logger.Error(c.Request.Context(), "Database type assertion failed")
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	// 检查是否有正在运行的实例
	var runningCount int64
	if err := db.WithContext(c.Request.Context()).
		Model(&model.Instance{}).
		Where("challenge_id = ? AND status = ?", challengeID, "running").
		Count(&runningCount).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to check running instances", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "检查失败",
		})
		return
	}

	if runningCount > 0 {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "该题目有正在运行的实例，无法删除",
		})
		return
	}

	// 删除题目
	if err := db.WithContext(c.Request.Context()).Where("id = ?", challengeID).Delete(&model.Challenge{}).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to delete challenge", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "删除失败",
		})
		return
	}

	logger.Info(c.Request.Context(), "Deleted challenge", "id", challengeID)

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Challenge deleted successfully",
	})
}

// UpdateChallengeStatus 更新题目状态（上架/下架）
// PUT /api/admin/challenges/:id/status
func (h *AdminHandler) UpdateChallengeStatus(c *gin.Context) {
	challengeID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=published unpublished"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid request format",
		})
		return
	}

	db, ok := h.db.(*gorm.DB)
	if !ok {
		logger.Error(c.Request.Context(), "Database type assertion failed")
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	// 准备更新字段，包含时间戳
	now := time.Now()
	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_at": now,
	}

	// 根据状态记录时间
	if req.Status == "published" {
		updates["published_at"] = now
	} else {
		updates["unpublished_at"] = now
	}

	// 更新状态
	result := db.WithContext(c.Request.Context()).
		Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(updates)

	if result.Error != nil {
		logger.Error(c.Request.Context(), "Failed to update status", "error", result.Error)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "状态更新失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.PureJSON(http.StatusNotFound, APIResponse{
			Code: 404,
			Msg:  "题目不存在",
		})
		return
	}

	logger.Info(c.Request.Context(), "Updated challenge status", "id", challengeID, "status", req.Status)

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Status updated successfully",
	})
}

// ListInstances 获取实例列表
// GET /api/admin/instances
func (h *AdminHandler) ListInstances(c *gin.Context) {
	// 查询参数
	status := c.Query("status")       // running/stopped/expired
	challenge := c.Query("challenge") // 题目ID
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)

	logger.Info(c.Request.Context(), "Listing instances", "status", status, "challenge", challenge, "page", pageNum)

	db, ok := h.db.(*gorm.DB)
	if !ok {
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	// 构建查询
	query := db.WithContext(c.Request.Context()).Model(&model.Instance{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if challenge != "" {
		query = query.Where("challenge_id = ?", challenge)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页查询
	var instances []model.Instance
	offset := (pageNum - 1) * pageSizeNum
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSizeNum).Find(&instances).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to list instances", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "查询失败",
		})
		return
	}

	// 关联题目信息
	type InstanceWithChallenge struct {
		model.Instance
		ChallengeTitle string `json:"challenge_title"`
	}

	var result []InstanceWithChallenge
	for _, inst := range instances {
		item := InstanceWithChallenge{Instance: inst}
		// 查询题目标题
		var chal model.Challenge
		if db.First(&chal, "id = ?", inst.ChallengeID).Error == nil {
			item.ChallengeTitle = chal.Title
		}
		result = append(result, item)
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: gin.H{
			"list":     result,
			"total":    total,
			"page":     pageNum,
			"pageSize": pageSizeNum,
		},
	})
}

// ListSubmissions 获取提交记录列表
// GET /api/admin/submissions
func (h *AdminHandler) ListSubmissions(c *gin.Context) {
	// 查询参数
	user := c.Query("user")
	challenge := c.Query("challenge")
	result := c.Query("result") // correct/wrong
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)

	logger.Info(c.Request.Context(), "Listing submissions", "user", user, "challenge", challenge, "result", result)

	// TODO: 实现提交记录查询逻辑
	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: gin.H{
			"list":     []interface{}{},
			"total":    0,
			"page":     pageNum,
			"pageSize": pageSizeNum,
		},
	})
}

// GetOverviewStats 获取总览统计数据
// GET /api/admin/overview/stats
func (h *AdminHandler) GetOverviewStats(c *gin.Context) {
	db, ok := h.db.(*gorm.DB)
	if !ok {
		c.PureJSON(http.StatusInternalServerError, APIResponse{Code: 500, Msg: "系统错误"})
		return
	}

	ctx := c.Request.Context()
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 1. 统计实例数据
	var todayInstances int64
	var runningInstances int64
	if err := db.WithContext(ctx).Model(&model.Instance{}).Where("created_at >= ?", todayStart).Count(&todayInstances).Error; err != nil {
		logger.Error(ctx, "Failed to count stats", "error", err)
	}
	if err := db.WithContext(ctx).Model(&model.Instance{}).Where("status = ?", "running").Count(&runningInstances).Error; err != nil {
		logger.Error(ctx, "Failed to count stats", "error", err)
	}

	// 2. 统计提交数据
	var todaySubmissions int64
	var todayCorrect int64
	if err := db.WithContext(ctx).Model(&model.Submission{}).Where("submitted_at >= ?", todayStart).Count(&todaySubmissions).Error; err != nil {
		logger.Error(ctx, "Failed to count stats", "error", err)
	}
	if err := db.WithContext(ctx).Model(&model.Submission{}).Where("submitted_at >= ? AND is_correct = ?", todayStart, true).Count(&todayCorrect).Error; err != nil {
		logger.Error(ctx, "Failed to count stats", "error", err)
	}

	todayCorrectRate := 0
	if todaySubmissions > 0 {
		todayCorrectRate = int((float64(todayCorrect) / float64(todaySubmissions)) * 100)
	}

	// 3. 最近提交记录 (联表查询获取用户名和题目名)
	type SubmissionView struct {
		ID              string    `json:"id"`
		UserDisplayName string    `json:"userDisplayName"`
		ChallengeTitle  string    `json:"challengeTitle"`
		Result          string    `json:"result"` // correct/wrong
		CreatedAt       time.Time `json:"createdAt"`
	}
	var recentSubmissions []SubmissionView

	// 处理 Result 字段转换 (bool -> string)
	type TempSubmission struct {
		ID              string
		UserDisplayName string
		ChallengeTitle  string
		IsCorrect       bool
		CreatedAt       time.Time
	}
	var tempRecent []TempSubmission
	if err := db.WithContext(ctx).Table("submissions").
		Select("submissions.id, users.username as user_display_name, challenges.title as challenge_title, submissions.is_correct, submissions.submitted_at as created_at").
		Joins("LEFT JOIN users ON users.id = submissions.user_id").
		Joins("LEFT JOIN challenges ON challenges.id = submissions.challenge_id").
		Order("submissions.submitted_at DESC").
		Limit(8).
		Scan(&tempRecent).Error; err != nil {
		logger.Error(ctx, "Failed to list recent submissions", "error", err)
	}

	recentSubmissions = make([]SubmissionView, len(tempRecent))
	for i, item := range tempRecent {
		res := "wrong"
		if item.IsCorrect {
			res = "correct"
		}
		recentSubmissions[i] = SubmissionView{
			ID:              item.ID,
			UserDisplayName: item.UserDisplayName,
			ChallengeTitle:  item.ChallengeTitle,
			Result:          res,
			CreatedAt:       item.CreatedAt,
		}
	}

	// 4. 热门题目 Top5
	type HotChallenge struct {
		Title string `json:"title"`
		Count int64  `json:"count"`
	}
	var hotChallenges []HotChallenge
	db.WithContext(ctx).Table("submissions").
		Select("challenges.title, count(*) as count").
		Joins("LEFT JOIN challenges ON challenges.id = submissions.challenge_id").
		Group("submissions.challenge_id").
		Order("count DESC").
		Limit(5).
		Scan(&hotChallenges)

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: gin.H{
			"todayInstances":    todayInstances,
			"runningInstances":  runningInstances,
			"todaySubmissions":  todaySubmissions,
			"todayCorrectRate":  todayCorrectRate,
			"recentSubmissions": recentSubmissions,
			"hotChallenges":     hotChallenges,
		},
	})
}
