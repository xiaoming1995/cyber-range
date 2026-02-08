package middleware

import (
	"bytes"
	"context"
	"cyber-range/internal/infra/logstore"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"io"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// logStoreInstance 全局 LogStore 实例（由 main.go 注入）
var logStoreInstance logstore.LogStore

// SetLogStore 设置 LogStore 实例
func SetLogStore(ls logstore.LogStore) {
	logStoreInstance = ls
}

// bodyLogWriter 自定义 ResponseWriter 以截获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// LoggerMiddleware adds structured logger and trace ID to each request
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Generate or Extract Trace ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 2. Set Trace ID in Response Header (for frontend debugging)
		c.Header("X-Trace-ID", traceID)

		// 3. Inject Trace ID into Context (to be passed to Service/Core layers)
		ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
		c.Request = c.Request.WithContext(ctx)

		// --- [NEW] Capture Request Body ---
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Truncate if too long
			if len(bodyBytes) > 4096 {
				reqBody = string(bodyBytes[:4096]) + "...(truncated)"
			} else {
				reqBody = string(bodyBytes)
			}
		}

		// --- [NEW] Capture Response Body ---
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 4. Process Request
		c.Next()

		// 5. Log Request Completion
		latency := time.Since(start)
		status := c.Writer.Status()

		// Determine log level based on status code
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		// Log using our structured logger with context
		logger.Log.Log(
			ctx,
			level,
			"request handled",
			"status", status,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
			"latency_ms", latency.Milliseconds(),
			"user_agent", c.Request.UserAgent(),
			"trace_id", traceID,
		)

		// 6. Store to LogStore (async, non-blocking)
		if logStoreInstance != nil {
			// 获取错误信息（如果有 c.Error）
			var errorMsg string
			if len(c.Errors) > 0 {
				errorMsg = c.Errors.String()
			}

			// 获取响应体 (Truncate if too long)
			respBody := blw.body.String()
			if len(respBody) > 4096 {
				respBody = respBody[:4096] + "...(truncated)"
			}

			// 获取用户ID（如果已登录）
			userID, _ := c.Get("user_id")
			userIDStr, _ := userID.(string)

			logStoreInstance.Store(&model.APILog{
				TraceID:      traceID,
				Method:       c.Request.Method,
				Path:         c.Request.URL.Path,
				Status:       status,
				LatencyMs:    latency.Milliseconds(),
				IP:           c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
				UserID:       userIDStr,
				ErrorMessage: errorMsg,
				RequestBody:  reqBody,  // [NEW]
				ResponseBody: respBody, // [NEW]
			})
		}
	}
}
