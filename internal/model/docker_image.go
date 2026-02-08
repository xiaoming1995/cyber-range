package model

import (
	"fmt"
	"time"
)

type DockerImage struct {
	ID       string `gorm:"primaryKey;size:36" json:"id"`
	Name     string `gorm:"size:255;not null" json:"name"`
	Tag      string `gorm:"size:50;not null;default:latest" json:"tag"`
	Registry string `gorm:"size:255;default:localhost:5000" json:"registry"`

	Size         int64  `gorm:"bigint" json:"size"`
	Digest       string `gorm:"size:100" json:"digest"`
	Architecture string `gorm:"size:20;default:amd64" json:"architecture"`

	RecommendedMemory int64   `gorm:"default:0;comment:推荐内存限制(字节),0表示使用默认" json:"recommended_memory"`
	RecommendedCPU    float64 `gorm:"default:0;comment:推荐CPU限制(核心数),0表示使用默认" json:"recommended_cpu"`

	IsAvailable bool       `gorm:"default:true" json:"is_available"`
	LastSyncAt  *time.Time `gorm:"type:timestamp" json:"last_sync_at"`

	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DockerImage) TableName() string {
	return "docker_images"
}

// GetFullName 返回完整镜像名 (registry/name:tag)
func (img *DockerImage) GetFullName() string {
	return fmt.Sprintf("%s/%s:%s", img.Registry, img.Name, img.Tag)
}

// GetShortName 返回简短名称 (name:tag)
func (img *DockerImage) GetShortName() string {
	return fmt.Sprintf("%s:%s", img.Name, img.Tag)
}
