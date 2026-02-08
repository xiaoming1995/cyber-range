package model

import "time"

// APILog API请求日志表 - 存储所有API请求记录
type APILog struct {
	ID           string    `gorm:"primaryKey;size:36;comment:日志唯一标识" json:"id"`
	TraceID      string    `gorm:"size:36;index;comment:链路追踪ID" json:"trace_id"`
	Method       string    `gorm:"size:10;comment:HTTP方法" json:"method"`
	Path         string    `gorm:"size:500;index;comment:请求路径" json:"path"`
	Status       int       `gorm:"index;comment:响应状态码" json:"status"`
	LatencyMs    int64     `gorm:"comment:响应延迟(毫秒)" json:"latency_ms"`
	IP           string    `gorm:"size:50;comment:客户端IP" json:"ip"`
	UserAgent    string    `gorm:"size:500;comment:用户代理" json:"user_agent"`
	UserID       string    `gorm:"size:36;index;comment:登录用户ID(可选)" json:"user_id,omitempty"`
	ErrorMessage string    `gorm:"type:text;comment:错误信息" json:"error_message,omitempty"`
	RequestBody  string    `gorm:"type:text;comment:请求体" json:"request_body,omitempty"`
	ResponseBody string    `gorm:"type:text;comment:响应体" json:"response_body,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (APILog) TableName() string { return "api_logs" }
