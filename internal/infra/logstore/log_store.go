package logstore

import (
	"context"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Filter 日志查询筛选条件
type Filter struct {
	Status    *int      // 状态码筛选
	StatusMin *int      // 状态码范围下限（如 400）
	StatusMax *int      // 状态码范围上限（如 599）
	Path      string    // 路径模糊匹配
	Method    string    // HTTP 方法
	StartTime time.Time // 时间范围开始
	EndTime   time.Time // 时间范围结束
	TraceID   string    // Trace ID 精确匹配
}

// LogStats 日志统计信息
type LogStats struct {
	TotalRequests int64   `json:"total_requests"`
	ErrorRequests int64   `json:"error_requests"` // status >= 400
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	TodayRequests int64   `json:"today_requests"`
	TodayErrors   int64   `json:"today_errors"`
}

// LogStore 日志存储接口
type LogStore interface {
	Store(log *model.APILog)
	Query(ctx context.Context, filter Filter, page, pageSize int) ([]model.APILog, int64, error)
	GetStats(ctx context.Context) (*LogStats, error)
	Shutdown()
}

// MySQLLogStore MySQL 日志存储实现
type MySQLLogStore struct {
	db            *gorm.DB
	logChan       chan *model.APILog
	batchSize     int
	flushInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewMySQLLogStore 创建 MySQL 日志存储
func NewMySQLLogStore(db *gorm.DB) *MySQLLogStore {
	store := &MySQLLogStore{
		db:            db,
		logChan:       make(chan *model.APILog, 1000), // 缓冲 1000 条
		batchSize:     100,                            // 达到 100 条触发写入
		flushInterval: 5 * time.Second,                // 每 5 秒强制刷新
		stopChan:      make(chan struct{}),
	}

	// 启动批量写入 Goroutine
	store.wg.Add(1)
	go store.batchWriter()

	return store
}

// Store 非阻塞写入日志（写入 Channel）
func (s *MySQLLogStore) Store(log *model.APILog) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	select {
	case s.logChan <- log:
		// DEBUG
		// logger.Info(context.Background(), "LogStore: log pushed to channel")
	default:
		// Channel 已满，丢弃日志（避免阻塞请求）
		logger.Warn(context.Background(), "LogStore: channel full, dropping log", "chan_len", len(s.logChan))
	}
}

// batchWriter 批量写入 Goroutine
func (s *MySQLLogStore) batchWriter() {
	defer s.wg.Done()

	batch := make([]*model.APILog, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		// DEBUG: 打印批量写入信息
		logger.Info(context.Background(), "LogStore: flushing batch", "count", len(batch))
		if err := s.db.CreateInBatches(batch, len(batch)).Error; err != nil {
			logger.Error(context.Background(), "LogStore: batch insert failed", "error", err, "count", len(batch))
		}
		batch = batch[:0]
	}

	for {
		select {
		case log := <-s.logChan:
			batch = append(batch, log)
			if len(batch) >= s.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-s.stopChan:
			// 关闭前刷新剩余日志
			close(s.logChan)
			for log := range s.logChan {
				batch = append(batch, log)
			}
			flush()
			return
		}
	}
}

// Query 查询日志
func (s *MySQLLogStore) Query(ctx context.Context, filter Filter, page, pageSize int) ([]model.APILog, int64, error) {
	var logs []model.APILog
	var total int64

	query := s.db.WithContext(ctx).Model(&model.APILog{})

	// 应用筛选条件
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.StatusMin != nil {
		query = query.Where("status >= ?", *filter.StatusMin)
	}
	if filter.StatusMax != nil {
		query = query.Where("status <= ?", *filter.StatusMax)
	}
	if filter.Path != "" {
		query = query.Where("path LIKE ?", "%"+filter.Path+"%")
	}
	if filter.Method != "" {
		query = query.Where("method = ?", filter.Method)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}
	if filter.TraceID != "" {
		query = query.Where("trace_id = ?", filter.TraceID)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetStats 获取统计信息
func (s *MySQLLogStore) GetStats(ctx context.Context) (*LogStats, error) {
	var stats LogStats

	// 总请求数
	s.db.WithContext(ctx).Model(&model.APILog{}).Count(&stats.TotalRequests)

	// 错误请求数 (status >= 400)
	s.db.WithContext(ctx).Model(&model.APILog{}).Where("status >= 400").Count(&stats.ErrorRequests)

	// 平均延迟
	var avgResult struct{ Avg float64 }
	s.db.WithContext(ctx).Model(&model.APILog{}).Select("AVG(latency_ms) as avg").Scan(&avgResult)
	stats.AvgLatencyMs = avgResult.Avg

	// 今日统计
	today := time.Now().Truncate(24 * time.Hour)
	s.db.WithContext(ctx).Model(&model.APILog{}).Where("created_at >= ?", today).Count(&stats.TodayRequests)
	s.db.WithContext(ctx).Model(&model.APILog{}).Where("created_at >= ? AND status >= 400", today).Count(&stats.TodayErrors)

	return &stats, nil
}

// Shutdown 优雅关闭
func (s *MySQLLogStore) Shutdown() {
	close(s.stopChan)
	s.wg.Wait()
}
