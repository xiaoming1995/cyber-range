package service

import (
	"context"
	"cyber-range/internal/model"
	"cyber-range/pkg/jwt"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService struct {
	db *gorm.DB
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

// Login 管理员登录
func (s *AdminService) Login(ctx context.Context, username, password string) (token string, admin *model.Admin, err error) {
	// 查找管理员
	var dbAdmin model.Admin
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&dbAdmin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户名或密码错误")
		}
		return "", nil, err
	}

	// 检查是否激活
	if !dbAdmin.IsActive {
		return "", nil, errors.New("账号已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(dbAdmin.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	dbAdmin.LastLoginAt = &now
	s.db.WithContext(ctx).Model(&dbAdmin).Update("last_login_at", now)

	// 生成 JWT Token
	token, err = jwt.GenerateAdminToken(dbAdmin.ID, dbAdmin.Username)
	if err != nil {
		return "", nil, err
	}

	return token, &dbAdmin, nil
}

// CreateAdmin 创建管理员（用于初始化或添加新管理员）
func (s *AdminService) CreateAdmin(ctx context.Context, username, email, password, name string) (*model.Admin, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.WithContext(ctx).Model(&model.Admin{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建管理员
	admin := &model.Admin{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(admin).Error; err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdminByID 根据ID获取管理员信息
func (s *AdminService) GetAdminByID(ctx context.Context, adminID string) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.WithContext(ctx).Where("id = ?", adminID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}
