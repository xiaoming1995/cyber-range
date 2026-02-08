package service

import (
	"context"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"time"

	"gorm.io/gorm"
)

// LogCleaner 日志清理服务
type LogCleaner struct {
	db            *gorm.DB
	retentionDays int
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// NewLogCleaner 创建日志清理服务
func NewLogCleaner(db *gorm.DB, retentionDays int) *LogCleaner {
	if retentionDays <= 0 {
		retentionDays = 7 // 默认保留 7 天
	}
	return &LogCleaner{
		db:            db,
		retentionDays: retentionDays,
		ticker:        time.NewTicker(24 * time.Hour), // 每 24 小时执行一次
		stopChan:      make(chan struct{}),
	}
}

// Start 启动定时清理任务
func (c *LogCleaner) Start(ctx context.Context) {
	logger.Info(ctx, "LogCleaner started", "retention_days", c.retentionDays)

	// 启动时立即执行一次清理
	go func() {
		c.Cleanup(ctx)
	}()

	go func() {
		for {
			select {
			case <-c.ticker.C:
				count, err := c.Cleanup(ctx)
				if err != nil {
					logger.Error(ctx, "LogCleaner: cleanup failed", "error", err)
				} else if count > 0 {
					logger.Info(ctx, "LogCleaner: cleanup completed", "deleted_count", count)
				}
			case <-c.stopChan:
				logger.Info(ctx, "LogCleaner stopped")
				return
			}
		}
	}()
}

// Stop 停止清理任务
func (c *LogCleaner) Stop() {
	c.ticker.Stop()
	close(c.stopChan)
}

// Cleanup 清理过期日志
func (c *LogCleaner) Cleanup(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -c.retentionDays)

	result := c.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&model.APILog{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
