package db

import (
	"cyber-range/internal/model"
	"cyber-range/pkg/config"
	"cyber-range/pkg/logger"
	"fmt"
	"time"

	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes MySQL database connection
func InitDB(ctx context.Context, cfg *config.MySQLConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	// Configure GORM logger
	gormLog := gormlogger.Default.LogMode(gormlogger.Silent)
	if config.AppConfig.Server.Env == "dev" {
		gormLog = gormlogger.Default.LogMode(gormlogger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&model.DockerHost{},
		&model.Challenge{},
		&model.Instance{},
		&model.User{},
		&model.Submission{},
		&model.Admin{},
		&model.DockerImage{}, // 添加 DockerImage 表
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	DB = db
	logger.Info(ctx, "MySQL database initialized successfully")
	return db, nil
}

// Close closes database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
