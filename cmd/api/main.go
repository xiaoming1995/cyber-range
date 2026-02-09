package main

import (
	"context"
	"cyber-range/internal/api/handlers"
	"cyber-range/internal/api/middleware"
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/infra/logstore"
	"cyber-range/internal/infra/redis"
	"cyber-range/internal/model"
	"cyber-range/internal/service"
	"cyber-range/pkg/config"
	"cyber-range/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
	}

	// 1. Load Configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. Initialize Logger
	logger.InitLogger(cfg.Server.Env)
	ctx := context.Background()
	logger.Info(ctx, "Starting Cyber Range Platform", "env", cfg.Server.Env)

	// 3. Initialize MySQL
	gormDB, err := db.InitDB(ctx, &cfg.MySQL)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		panic(err)
	}
	defer db.Close()

	// 4. Initialize Redis
	_, err = redis.InitRedis(ctx, &cfg.Redis)
	if err != nil {
		logger.Error(ctx, "Failed to initialize Redis", "error", err)
		panic(err)
	}
	defer redis.Close()

	// 5. Initialize Docker Host Manager
	dockerManager := docker.NewDockerHostManager()
	logger.Info(ctx, "Docker Host Manager initialized")

	// 6. Initialize Repository
	repository := db.NewRepository(gormDB)
	logger.Info(ctx, "Repository initialized")

	// 7. Auto-migrate APILog table
	if err := gormDB.AutoMigrate(&model.APILog{}); err != nil {
		logger.Warn(ctx, "Failed to auto-migrate APILog table", "error", err)
	}

	// 8. Initialize LogStore and LogCleaner
	logStore := logstore.NewMySQLLogStore(gormDB)
	middleware.SetLogStore(logStore)
	logCleaner := service.NewLogCleaner(gormDB, 7) // 保留 7 天
	logCleaner.Start(ctx)
	logger.Info(ctx, "LogStore and LogCleaner initialized")

	// 9. Initialize Services
	challengeSvc := service.NewChallengeService(dockerManager, repository, gormDB, cfg)
	adminSvc := service.NewAdminService(gormDB)
	imageSvc := service.NewImageService(repository, dockerManager)

	// 11. 启动时自动同步 Registry 并预加载镜像
	go func() {
		time.Sleep(3 * time.Second) // 等待服务启动

		// 步骤1：同步 Registry 镜像到数据库
		logger.Info(ctx, "开始自动同步 Registry 镜像...")
		count, err := imageSvc.SyncFromRegistry(ctx, "http://localhost:5000")
		if err != nil {
			logger.Warn(ctx, "自动同步失败", "error", err)
		} else {
			logger.Info(ctx, "自动同步完成", "synced_count", count)
		}

		// 步骤2：预加载镜像到各主机
		time.Sleep(2 * time.Second)
		logger.Info(ctx, "开始自动预加载镜像...")
		if err := imageSvc.PreloadAllImages(ctx); err != nil {
			logger.Warn(ctx, "自动预加载失败", "error", err)
		} else {
			logger.Info(ctx, "镜像预加载任务已启动")
		}
	}()

	// 12. Start HTTP Server (background job)
	reaper := service.NewReaper(dockerManager, repository, gormDB)
	reaper.Start(ctx)
	defer reaper.Stop()

	// 10. Initialize Handlers
	challengeHandler := handlers.NewChallengeHandler(challengeSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc, challengeSvc, gormDB)
	dockerHostHandler := handlers.NewDockerHostHandler(repository, dockerManager)
	imageHandler := handlers.NewImageHandler(imageSvc)
	instanceHandler := handlers.NewInstanceHandler(repository, dockerManager)
	logHandler := handlers.NewLogHandler(logStore)

	// 10. Setup Router
	gin.SetMode(gin.ReleaseMode)
	if cfg.Server.Env == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())

	// 配置JSON编码器：禁用HTML转义，保持中文字符原样输出
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Trace-ID", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Proxy trust settings
	r.SetTrustedProxies(nil)

	// API Routes (User)
	api := r.Group("/api")
	{
		api.GET("/challenges", challengeHandler.List)
		api.POST("/challenges/:id/start", challengeHandler.Start)
		api.POST("/challenges/:id/stop", challengeHandler.Stop)
		api.POST("/submit", challengeHandler.Verify)
	}

	// Admin API Routes
	admin := r.Group("/api/admin")
	{
		// Public routes
		admin.POST("/login", adminHandler.Login)

		// Protected routes (require admin auth)
		protected := admin.Group("")
		protected.Use(middleware.AdminAuth())
		{
			// 题库管理
			protected.GET("/challenges", adminHandler.ListChallenges)
			protected.GET("/challenges/:id", adminHandler.GetChallenge)
			protected.POST("/challenges", adminHandler.CreateChallenge)
			protected.PUT("/challenges/:id", adminHandler.UpdateChallenge)
			protected.DELETE("/challenges/:id", adminHandler.DeleteChallenge)
			protected.PUT("/challenges/:id/status", adminHandler.UpdateChallengeStatus)

			// 实例管理
			protected.GET("/instances", adminHandler.ListInstances)
			protected.GET("/instances/:id/stats", instanceHandler.GetInstanceStats)
			protected.GET("/instances/:id/logs", instanceHandler.GetInstanceLogs)

			// 提交记录
			protected.GET("/submissions", adminHandler.ListSubmissions)

			// 总览统计
			protected.GET("/overview/stats", adminHandler.GetOverviewStats)

			// Docker 主机管理
			protected.GET("/docker-hosts", dockerHostHandler.ListDockerHosts)
			protected.POST("/docker-hosts", dockerHostHandler.CreateDockerHost)
			protected.PUT("/docker-hosts/:id", dockerHostHandler.UpdateDockerHost)
			protected.DELETE("/docker-hosts/:id", dockerHostHandler.DeleteDockerHost)
			protected.POST("/docker-hosts/:id/test", dockerHostHandler.TestDockerHost)
			protected.POST("/docker-hosts/:id/toggle", dockerHostHandler.ToggleDockerHost)

			// Docker 镜像管理
			protected.GET("/images", imageHandler.List)
			protected.POST("/images", imageHandler.Register)
			protected.DELETE("/images/:id", imageHandler.Delete)
			protected.POST("/images/sync", imageHandler.Sync)
			protected.POST("/images/preload", imageHandler.Preload)
			protected.POST("/images/upload", imageHandler.Upload)

			// API 日志管理
			protected.GET("/logs", logHandler.List)
			protected.GET("/logs/stats", logHandler.GetStats)
		}
	}

	// 11. Graceful Shutdown
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info(ctx, "Server starting", "addr", addr)

	// Start server in goroutine
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Error(ctx, "Server failed to run", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down server gracefully...")
	reaper.Stop()
	logCleaner.Stop()
	logStore.Shutdown()
	logger.Info(ctx, "Server exited")
}
