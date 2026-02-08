package model

import "time"

// DockerHost Docker主机配置表 - 存储所有Docker主机连接信息
type DockerHost struct {
	ID        string `gorm:"primaryKey;size:36;comment:Docker主机唯一标识" json:"id"`
	Name      string `gorm:"size:100;not null;comment:主机名称(如:本地Docker,远程服务器1)" json:"name"`
	Host      string `gorm:"size:255;not null;comment:Docker连接地址(如:tcp://192.168.1.100:2376)" json:"host"`
	TLSVerify bool   `gorm:"default:false;comment:是否启用TLS加密" json:"tls_verify"`
	CertPath  string `gorm:"size:500;comment:TLS证书路径" json:"cert_path"`

	// 端口分配范围
	PortRangeMin int `gorm:"not null;default:20000;comment:端口范围最小值" json:"port_range_min"`
	PortRangeMax int `gorm:"not null;default:40000;comment:端口范围最大值" json:"port_range_max"`

	// 资源限制（可选，允许每个主机有不同的限制）
	MemoryLimit int64   `gorm:"default:134217728;comment:默认内存限制(字节)" json:"memory_limit"`
	CPULimit    float64 `gorm:"type:decimal(3,2);default:0.50;comment:默认CPU限制(核心数)" json:"cpu_limit"`

	// 状态控制
	Enabled   bool `gorm:"default:true;comment:是否启用(管理员可手动禁用)" json:"enabled"`
	IsDefault bool `gorm:"default:false;index;comment:是否为默认主机" json:"is_default"`

	// 元数据
	Description string    `gorm:"type:text;comment:主机描述" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定自定义表名
func (DockerHost) TableName() string {
	return "docker_hosts"
}
