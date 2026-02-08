package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Instance InstanceConfig `mapstructure:"instance"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	InstanceTTL int    `mapstructure:"instance_ttl"`
}

// DockerHostConfig Docker 主机连接配置
type DockerHostConfig struct {
	Host      string `mapstructure:"host"`
	TLSVerify bool   `mapstructure:"tls_verify"`
	CertPath  string `mapstructure:"cert_path"`
}

// DockerConfig Docker 容器引擎配置
type DockerConfig struct {
	Mode         string           `mapstructure:"mode"`           // "local" 或 "remote"
	Local        DockerHostConfig `mapstructure:"local"`          // 本地配置
	Remote       DockerHostConfig `mapstructure:"remote"`         // 远程配置
	PortRangeMin int              `mapstructure:"port_range_min"` // 端口范围最小值
	PortRangeMax int              `mapstructure:"port_range_max"` // 端口范围最大值
	MemoryLimit  int64            `mapstructure:"memory_limit"`   // 内存限制（字节）
	CPULimit     float64          `mapstructure:"cpu_limit"`      // CPU限制（核心数）
}

// GetActiveHost 根据当前模式获取激活的主机配置
func (d *DockerConfig) GetActiveHost() DockerHostConfig {
	if d.Mode == "remote" {
		return d.Remote
	}
	return d.Local
}

type InstanceConfig struct {
	MaxPerUser int `mapstructure:"max_per_user"`
	TTLHours   int `mapstructure:"ttl_hours"`
}

var AppConfig *Config

// LoadConfig 从配置文件加载配置
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("配置文件读取失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}

	AppConfig = &cfg
	return &cfg, nil
}

// DSN 返回 MySQL 数据源名称
func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

// Addr 返回 Redis 地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
