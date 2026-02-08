package service

import (
	"bufio"
	"context"
	"crypto/rand"
	"cyber-range/internal/infra/db"
	"cyber-range/internal/infra/docker"
	"cyber-range/internal/model"
	"cyber-range/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ImageService struct {
	repo          *db.Repository
	dockerManager *docker.DockerHostManager
}

func NewImageService(repo *db.Repository, dockerManager *docker.DockerHostManager) *ImageService {
	return &ImageService{
		repo:          repo,
		dockerManager: dockerManager,
	}
}

// ListImages 获取所有可用镜像
func (s *ImageService) ListImages(ctx context.Context) ([]*model.DockerImage, error) {
	return s.repo.GetAllImages(ctx)
}

// RegisterImage 注册新镜像（管理员从本地Docker导入后调用）
func (s *ImageService) RegisterImage(ctx context.Context, name, tag, description string) (*model.DockerImage, error) {
	// 检查镜像是否已注册
	existing, _ := s.repo.GetImageByName(ctx, name, tag)
	if existing != nil {
		return nil, fmt.Errorf("镜像已存在: %s:%s", name, tag)
	}

	// 创建镜像记录
	img := &model.DockerImage{
		ID:          generateImageID(),
		Name:        name,
		Tag:         tag,
		Registry:    "localhost:5000",
		IsAvailable: true,
		Description: description,
	}

	if err := s.repo.CreateImage(ctx, img); err != nil {
		return nil, fmt.Errorf("注册镜像失败: %w", err)
	}

	logger.Info(ctx, "镜像注册成功", "image", img.GetShortName())
	return img, nil
}

// Registry API 响应结构
type registryCatalog struct {
	Repositories []string `json:"repositories"`
}

type registryTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// SyncFromRegistry 从 Registry 同步镜像到数据库
func (s *ImageService) SyncFromRegistry(ctx context.Context, registryURL string) (int, error) {
	logger.Info(ctx, "开始从 Registry 同步镜像", "registry", registryURL)

	// 1. 获取 Registry 中的所有仓库
	catalogURL := fmt.Sprintf("%s/v2/_catalog", registryURL)
	resp, err := http.Get(catalogURL)
	if err != nil {
		return 0, fmt.Errorf("无法连接到 Registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("Registry 返回错误: %s - %s", resp.Status, string(body))
	}

	var catalog registryCatalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return 0, fmt.Errorf("解析 Registry 响应失败: %w", err)
	}

	logger.Info(ctx, "发现镜像仓库", "count", len(catalog.Repositories))

	syncedCount := 0

	// 2. 遍历每个仓库，获取标签
	for _, repo := range catalog.Repositories {
		tagsURL := fmt.Sprintf("%s/v2/%s/tags/list", registryURL, repo)
		resp, err := http.Get(tagsURL)
		if err != nil {
			logger.Warn(ctx, "获取镜像标签失败", "repo", repo, "error", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			logger.Warn(ctx, "镜像标签返回错误", "repo", repo, "status", resp.Status)
			continue
		}

		var tags registryTags
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			resp.Body.Close()
			logger.Warn(ctx, "解析标签失败", "repo", repo, "error", err)
			continue
		}
		resp.Body.Close()

		// 3. 为每个标签检查是否已存在，不存在则创建
		for _, tag := range tags.Tags {
			// 检查是否已存在
			existing, err := s.repo.GetImageByName(ctx, repo, tag)
			if err == nil && existing != nil {
				logger.Debug(ctx, "镜像已存在，跳过", "name", repo, "tag", tag)
				continue
			}

			// 创建新镜像记录
			registryHost := "localhost:5000" // 默认值
			image := &model.DockerImage{
				ID:          uuid.New().String(),
				Name:        repo,
				Tag:         tag,
				Registry:    registryHost,
				IsAvailable: true,
				Description: fmt.Sprintf("从 Registry 自动同步: %s", time.Now().Format("2006-01-02 15:04:05")),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := s.repo.CreateImage(ctx, image); err != nil {
				logger.Warn(ctx, "创建镜像记录失败", "name", repo, "tag", tag, "error", err)
				continue
			}

			logger.Info(ctx, "同步镜像成功", "name", repo, "tag", tag)
			syncedCount++
		}
	}

	logger.Info(ctx, "Registry 同步完成", "synced", syncedCount, "total_repos", len(catalog.Repositories))
	return syncedCount, nil
}

// PreloadImages 预加载所有镜像到指定主机
func (s *ImageService) PreloadImages(ctx context.Context, hostID string) error {
	// 获取主机配置
	host, err := s.repo.GetDockerHostByID(ctx, hostID)
	if err != nil {
		return fmt.Errorf("主机不存在: %w", err)
	}

	// 获取所有镜像
	images, err := s.repo.GetAllImages(ctx)
	if err != nil {
		return fmt.Errorf("获取镜像列表失败: %w", err)
	}

	// 获取Docker客户端
	client, err := s.dockerManager.GetOrCreateClient(ctx, host)
	if err != nil {
		return fmt.Errorf("连接主机失败: %w", err)
	}

	// 逐个拉取镜像
	for _, img := range images {
		fullName := img.GetFullName()
		logger.Info(ctx, "开始拉取镜像", "host", host.Name, "image", fullName)

		if err := client.EnsureImage(ctx, fullName); err != nil {
			logger.Warn(ctx, "镜像拉取失败", "image", fullName, "error", err)
			continue
		}

		// 更新同步时间
		now := time.Now()
		img.LastSyncAt = &now
		s.repo.UpdateImage(ctx, img)

		logger.Info(ctx, "镜像拉取成功", "host", host.Name, "image", fullName)
	}

	return nil
}

// PreloadAllImages 预加载所有镜像到所有已启用主机
func (s *ImageService) PreloadAllImages(ctx context.Context) error {
	hosts, err := s.repo.GetEnabledDockerHosts(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "开始预加载镜像", "host_count", len(hosts))

	for _, host := range hosts {
		// 异步预加载每个主机
		go func(h *model.DockerHost) {
			hostCtx := context.Background() // 使用新的 context 避免超时
			if err := s.PreloadImages(hostCtx, h.ID); err != nil {
				logger.Warn(hostCtx, "主机预加载失败", "host", h.Name, "error", err)
			}
		}(host)
	}

	logger.Info(ctx, "镜像预加载任务已启动")
	return nil
}

// ImportResult 镜像导入结果
type ImportResult struct {
	ImageName   string `json:"image_name"`   // 原始镜像名
	RegistryTag string `json:"registry_tag"` // Registry 中的完整标签
	Pushed      bool   `json:"pushed"`       // 是否推送成功
}

// ImportFromTar 从 tar 文件导入镜像并推送到 Registry
func (s *ImageService) ImportFromTar(ctx context.Context, tarFilePath string) (*ImportResult, error) {
	logger.Info(ctx, "开始导入镜像", "file", tarFilePath)

	// 1. 执行 docker load
	loadCmd := exec.CommandContext(ctx, "docker", "load", "-i", tarFilePath)
	loadOutput, err := loadCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker load 失败: %s - %w", string(loadOutput), err)
	}

	// 2. 解析 docker load 输出，获取镜像名
	// 输出格式: "Loaded image: nginx:latest" 或 "Loaded image ID: sha256:xxx"
	loadedImage := ""
	scanner := bufio.NewScanner(strings.NewReader(string(loadOutput)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Loaded image:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) >= 2 {
				loadedImage = strings.TrimSpace(parts[1])
			}
		}
	}

	if loadedImage == "" {
		return nil, fmt.Errorf("无法解析镜像名，docker load 输出: %s", string(loadOutput))
	}

	logger.Info(ctx, "镜像加载成功", "image", loadedImage)

	// 3. 解析镜像名和标签
	imageName := loadedImage
	imageTag := "latest"
	if strings.Contains(loadedImage, ":") {
		parts := strings.SplitN(loadedImage, ":", 2)
		imageName = parts[0]
		imageTag = parts[1]
	}

	// 移除可能的 registry 前缀（如果源镜像带有其他 registry）
	if strings.Contains(imageName, "/") {
		parts := strings.Split(imageName, "/")
		imageName = parts[len(parts)-1]
	}

	// 4. 打标签为本地 Registry
	registryHost := "localhost:5000"
	registryTag := fmt.Sprintf("%s/%s:%s", registryHost, imageName, imageTag)

	tagCmd := exec.CommandContext(ctx, "docker", "tag", loadedImage, registryTag)
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker tag 失败: %s - %w", string(output), err)
	}

	logger.Info(ctx, "镜像打标签成功", "tag", registryTag)

	// 5. 推送到 Registry
	pushCmd := exec.CommandContext(ctx, "docker", "push", registryTag)
	pushOutput, err := pushCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker push 失败: %s - %w", string(pushOutput), err)
	}

	logger.Info(ctx, "镜像推送成功", "registry_tag", registryTag)

	// 6. 注册到数据库
	img := &model.DockerImage{
		ID:          uuid.New().String(),
		Name:        imageName,
		Tag:         imageTag,
		Registry:    registryHost,
		IsAvailable: true,
		Description: fmt.Sprintf("通过上传导入: %s", time.Now().Format("2006-01-02 15:04:05")),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 检查是否已存在
	existing, _ := s.repo.GetImageByName(ctx, imageName, imageTag)
	if existing == nil {
		if err := s.repo.CreateImage(ctx, img); err != nil {
			logger.Warn(ctx, "创建镜像记录失败", "error", err)
		}
	} else {
		// 更新已存在的记录
		existing.UpdatedAt = time.Now()
		s.repo.UpdateImage(ctx, existing)
		img = existing
	}

	// 7. 清理临时文件
	os.Remove(tarFilePath)

	logger.Info(ctx, "镜像导入完成", "name", imageName, "tag", imageTag)

	return &ImportResult{
		ImageName:   loadedImage,
		RegistryTag: registryTag,
		Pushed:      true,
	}, nil
}

// generateImageID 生成镜像 ID
func generateImageID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
