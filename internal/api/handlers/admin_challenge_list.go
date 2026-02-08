package handlers

import (
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListChallengesRequest 题目列表查询参数
type ListChallengesRequest struct {
	Page       int    `form:"page"`       // 页码，默认 1
	PageSize   int    `form:"pageSize"`   // 每页数量，默认 20
	Category   string `form:"category"`   // 分类筛选
	Difficulty string `form:"difficulty"` // 难度筛选
	Status     string `form:"status"`     // 状态筛选
	Search     string `form:"search"`     // 搜索关键词（标题）
}

// ListChallengesResponse 题目列表响应
type ListChallengesResponse struct {
	List     []model.Challenge `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// AdminChallengeView 包含 Flag 的完整题目视图
type AdminChallengeView struct {
	model.Challenge
	Flag string `json:"flag"`
}

// ListChallenges 获取题目列表（管理员）
// GET /api/admin/challenges
func (h *AdminHandler) ListChallenges(c *gin.Context) {
	var req ListChallengesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid query parameters",
		})
		return
	}

	// 默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	// 限制最大每页数量
	if req.PageSize > 100 {
		req.PageSize = 100
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

	// 构建查询
	query := db.WithContext(c.Request.Context()).Model(&model.Challenge{})

	// 筛选条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Difficulty != "" {
		query = query.Where("difficulty = ?", req.Difficulty)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 搜索
	if req.Search != "" {
		searchPattern := "%" + strings.TrimSpace(req.Search) + "%"
		query = query.Where("title LIKE ?", searchPattern)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to count challenges", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "查询失败",
		})
		return
	}

	// 分页查询
	challenges := make([]model.Challenge, 0)
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Order("created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&challenges).Error; err != nil {
		logger.Error(c.Request.Context(), "Failed to list challenges", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "查询失败",
		})
		return
	}

	// 转换为 Admin 视图（暴露 Flag）
	adminChallenges := make([]AdminChallengeView, len(challenges))
	for i, ch := range challenges {
		adminChallenges[i] = AdminChallengeView{
			Challenge: ch,
			Flag:      ch.Flag,
		}
	}

	// 返回结果
	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: gin.H{ // 保持与前端 ListChallengesResponse 结构一致，但 List 类型变了
			"list":     adminChallenges,
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	})
}

// GetChallenge 获取单个题目详情（管理员）
// GET /api/admin/challenges/:id
func (h *AdminHandler) GetChallenge(c *gin.Context) {
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

	var challenge model.Challenge
	if err := db.WithContext(c.Request.Context()).
		Where("id = ?", challengeID).
		First(&challenge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.PureJSON(http.StatusNotFound, APIResponse{
				Code: 404,
				Msg:  "题目不存在",
			})
			return
		}
		logger.Error(c.Request.Context(), "Failed to get challenge", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "查询失败",
		})
		return
	}

	// 返回完整视图
	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: AdminChallengeView{
			Challenge: challenge,
			Flag:      challenge.Flag,
		},
	})
}
