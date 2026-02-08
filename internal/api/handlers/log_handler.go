package handlers

import (
	"cyber-range/internal/infra/logstore"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志查询处理器
type LogHandler struct {
	logStore logstore.LogStore
}

// NewLogHandler 创建日志处理器
func NewLogHandler(logStore logstore.LogStore) *LogHandler {
	return &LogHandler{logStore: logStore}
}

// List 分页查询日志
// GET /api/admin/logs
func (h *LogHandler) List(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析筛选参数
	filter := logstore.Filter{
		Path:    c.Query("path"),
		Method:  c.Query("method"),
		TraceID: c.Query("trace_id"),
	}

	// 状态码筛选
	if status := c.Query("status"); status != "" {
		s, _ := strconv.Atoi(status)
		filter.Status = &s
	}
	if statusMin := c.Query("status_min"); statusMin != "" {
		s, _ := strconv.Atoi(statusMin)
		filter.StatusMin = &s
	}
	if statusMax := c.Query("status_max"); statusMax != "" {
		s, _ := strconv.Atoi(statusMax)
		filter.StatusMax = &s
	}

	// 时间范围
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = t
		}
	}

	// 查询
	logs, total, err := h.logStore.Query(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "查询日志失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetStats 获取日志统计
// GET /api/admin/logs/stats
func (h *LogHandler) GetStats(c *gin.Context) {
	stats, err := h.logStore.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取统计失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": stats,
	})
}
