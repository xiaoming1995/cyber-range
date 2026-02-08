package service

import (
	"context"
	"crypto/rand"
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	redisRepo "cyber-range/internal/infra/redis"
	"cyber-range/internal/model"
	"cyber-range/pkg/config"
	"cyber-range/pkg/logger"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ChallengeService struct {
	dockerManager *docker.DockerHostManager // Docker 主机管理器
	repo          *db.Repository            // 数据访问层
	gormDB        *gorm.DB                  // 保留用于兼容现有代码
	cfg           *config.Config
}

func NewChallengeService(dockerManager *docker.DockerHostManager, repo *db.Repository, gormDB *gorm.DB, cfg *config.Config) *ChallengeService {
	return &ChallengeService{
		dockerManager: dockerManager,
		repo:          repo,
		gormDB:        gormDB,
		cfg:           cfg,
	}
}

// ListChallenges returns all published challenges for users
func (s *ChallengeService) ListChallenges(ctx context.Context) ([]model.Challenge, error) {
	var challenges []model.Challenge
	// 只返回已发布的题目给用户
	if err := s.gormDB.WithContext(ctx).
		Where("status = ?", "published").
		Find(&challenges).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch challenges: %w", err)
	}
	return challenges, nil
}

// GetChallenge returns a single challenge by ID
func (s *ChallengeService) GetChallenge(ctx context.Context, id string) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := s.gormDB.First(&challenge, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("challenge not found")
		}
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	return &challenge, nil
}

// StartInstance 创建并启动带资源限制的容器
// 返回：容器ID, 分配的端口, 错误
func (s *ChallengeService) StartInstance(ctx context.Context, userID, challengeID string) (*model.Instance, error) {
	// 1. 检查题目是否存在
	challenge, err := s.GetChallenge(ctx, challengeID)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	// 2. 检查是否已经有该题目的运行实例（每个题目只能有1个实例）
	existingInstance, err := redisRepo.GetInstanceByUserAndChallenge(ctx, userID, challengeID)
	if err == nil && existingInstance != nil {
		return nil, fmt.Errorf("你已经启动了该题目的实例，请先停止后再重新启动")
	}

	// 注意：允许用户同时运行多个不同题目的实例，不限制总数

	// 3. 确定 Docker 主机
	dockerHostID := challenge.DockerHostID
	if dockerHostID == "" {
		// 使用默认主机
		defaultHost, err := s.repo.GetDefaultDockerHost(ctx)
		if err != nil {
			return nil, fmt.Errorf("未找到默认 Docker 主机: %w", err)
		}
		dockerHostID = defaultHost.ID
	}

	// 4. 加载 Docker 主机配置
	dockerHost, err := s.repo.GetDockerHostByID(ctx, dockerHostID)
	if err != nil {
		return nil, fmt.Errorf("Docker 主机配置不存在: %w", err)
	}

	// 5. 检查主机是否启用
	if !dockerHost.Enabled {
		return nil, fmt.Errorf("Docker 主机已禁用: %s", dockerHost.Name)
	}

	// 6. 获取 Docker 客户端
	dockerClient, err := s.dockerManager.GetOrCreateClient(ctx, dockerHost)
	if err != nil {
		return nil, fmt.Errorf("连接 Docker 主机失败: %w", err)
	}

	// 7. 生成唯一Flag
	flag := s.generateFlag(userID)
	logger.Debug(ctx, "Generated flag for user", "user_id", userID, "flag", flag)

	// 8. 启动 Docker 容器
	imageName := challenge.Image
	if challenge.ImageID != "" {
		var dockerImg model.DockerImage
		if err := s.gormDB.WithContext(ctx).First(&dockerImg, "id = ?", challenge.ImageID).Error; err == nil {
			imageName = dockerImg.GetFullName()
		} else {
			logger.Warn(ctx, "关联镜像不存在，降级使用 challenge.Image", "image_id", challenge.ImageID, "error", err)
		}
	}

	envVars := []string{fmt.Sprintf("FLAG=%s", flag)}
	containerID, port, err := dockerClient.StartContainer(ctx, imageName, envVars, challenge.Port, challenge.Privileged, challenge.MemoryLimit, challenge.CPULimit)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// 9. 创建实例记录
	instance := &model.Instance{
		ID:           generateID(),
		UserID:       userID,
		ChallengeID:  challengeID,
		ContainerID:  containerID,
		DockerHostID: dockerHost.ID,
		Flag:         flag,
		Port:         port,
		Status:       "running",
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.Instance.TTLHours) * time.Hour),
		CreatedAt:    time.Now(),
	}

	// 10. 存储到 Redis (with TTL) and DB (for history)
	if err := redisRepo.StoreInstance(ctx, instance.ID, userID, challengeID, containerID, flag, port, instance.ExpiresAt); err != nil {
		// Rollback: kill container if Redis fails
		dockerClient.StopContainer(ctx, containerID)
		return nil, fmt.Errorf("failed to store instance in Redis: %w", err)
	}

	if err := s.gormDB.Create(instance).Error; err != nil {
		logger.Warn(ctx, "Failed to save instance to DB (non-critical)", "error", err)
	}

	logger.Info(ctx, "Instance started successfully",
		"instance_id", instance.ID,
		"user_id", userID,
		"challenge_id", challengeID,
		"docker_host", dockerHost.Name,
		"port", port)

	return instance, nil
}

// StopInstance forcefully stops and cleans up an instance
func (s *ChallengeService) StopInstance(ctx context.Context, userID, challengeID string) error {
	// Get instance from Redis
	activeInstances, err := redisRepo.GetUserActiveInstances(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user instances: %w", err)
	}

	var targetInstanceID string
	for _, instID := range activeInstances {
		instData, _ := redisRepo.GetInstance(ctx, instID)
		if instData["challenge_id"] == challengeID {
			targetInstanceID = instID
			break
		}
	}

	if targetInstanceID == "" {
		return errors.New("no active instance found for this challenge")
	}

	instData, err := redisRepo.GetInstance(ctx, targetInstanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance data: %w", err)
	}

	containerID := instData["container_id"]

	// 从数据库读取完整的实例信息（包含 docker_host_id）
	var instance model.Instance
	if err := s.gormDB.WithContext(ctx).First(&instance, "id = ?", targetInstanceID).Error; err != nil {
		logger.Warn(ctx, "Instance not found in DB", "instance_id", targetInstanceID, "error", err)
		// 即使数据库查询失败，仍然清理 Redis
		redisRepo.DeleteInstance(ctx, targetInstanceID, userID)
		return fmt.Errorf("instance not found in database: %w", err)
	}

	// 获取 Docker 主机配置
	dockerHost, err := s.repo.GetDockerHostByID(ctx, instance.DockerHostID)
	if err != nil {
		logger.Warn(ctx, "Docker host not found", "docker_host_id", instance.DockerHostID, "error", err)
		// 清理 Redis
		redisRepo.DeleteInstance(ctx, targetInstanceID, userID)
		s.gormDB.Model(&model.Instance{}).Where("id = ?", targetInstanceID).Update("status", "stopped")
		return fmt.Errorf("Docker 主机配置不存在: %w", err)
	}

	// 获取 Docker 客户端
	dockerClient, err := s.dockerManager.GetOrCreateClient(ctx, dockerHost)
	if err != nil {
		logger.Warn(ctx, "Failed to get Docker client", "docker_host", dockerHost.Name, "error", err)
		// 清理 Redis
		redisRepo.DeleteInstance(ctx, targetInstanceID, userID)
		s.gormDB.Model(&model.Instance{}).Where("id = ?", targetInstanceID).Update("status", "stopped")
		return fmt.Errorf("连接 Docker 主机失败: %w", err)
	}

	// Force kill Docker container
	if err := dockerClient.StopContainer(ctx, containerID); err != nil {
		logger.Warn(ctx, "Failed to stop container (may already be stopped)", "error", err)
	}

	// Clean up Redis
	if err := redisRepo.DeleteInstance(ctx, targetInstanceID, userID); err != nil {
		return fmt.Errorf("failed to delete instance from Redis: %w", err)
	}

	// Update DB status
	s.gormDB.Model(&model.Instance{}).Where("id = ?", targetInstanceID).Update("status", "stopped")

	logger.Info(ctx, "Instance stopped successfully",
		"instance_id", targetInstanceID,
		"docker_host", dockerHost.Name)
	return nil
}

// VerifyFlag checks if submitted flag matches user's instance flag
func (s *ChallengeService) VerifyFlag(ctx context.Context, userID, challengeID, submittedFlag string) (bool, string, error) {
	// Get user's active instance
	activeInstances, err := redisRepo.GetUserActiveInstances(ctx, userID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get user instances: %w", err)
	}

	var correctFlag string
	for _, instID := range activeInstances {
		instData, _ := redisRepo.GetInstance(ctx, instID)
		if instData["challenge_id"] == challengeID {
			correctFlag = instData["flag"]
			break
		}
	}

	if correctFlag == "" {
		return false, "No active instance found. Please start the challenge first.", nil
	}

	isCorrect := correctFlag == submittedFlag

	// Record submission in DB
	challenge, _ := s.GetChallenge(ctx, challengeID)
	points := 0
	if isCorrect {
		points = challenge.Points
		// Award points to user
		s.gormDB.Model(&model.User{}).Where("id = ?", userID).Update("total_points", gorm.Expr("total_points + ?", points))
	}

	submission := &model.Submission{
		ID:          generateID(),
		UserID:      userID,
		ChallengeID: challengeID,
		Flag:        submittedFlag,
		IsCorrect:   isCorrect,
		Points:      points,
		SubmittedAt: time.Now(),
	}
	s.gormDB.Create(submission)

	if isCorrect {
		return true, "回答正确！你获得了积分。", nil
	}
	return false, "Flag 错误，请重试。", nil
}

// generateFlag creates a unique flag: flag{userID_timestamp_random}
func (s *ChallengeService) generateFlag(userID string) string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("flag{%s_%d_%s}", userID, timestamp, randomStr)
}

// generateID creates a random UUID-like ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
