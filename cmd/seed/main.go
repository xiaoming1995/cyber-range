package main

import (
	"context"
	"cyber-range/internal/model"
	"cyber-range/pkg/config"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()

	// 1. 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 3. 清空旧数据
	fmt.Println("【1/5】清空旧数据...")
	db.Exec("DELETE FROM submissions")
	db.Exec("DELETE FROM instances")
	db.Exec("DELETE FROM challenges")
	db.Exec("DELETE FROM docker_hosts")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM admins")
	fmt.Println("✓ 旧数据已清除")

	// 4. 插入管理员
	fmt.Println("\n【2/4】插入管理员账号...")
	admins := getAdmins()
	if err := db.CreateInBatches(admins, 5).Error; err != nil {
		log.Fatalf("管理员插入失败: %v", err)
	}
	fmt.Printf("✓ 已插入 %d 个管理员账号\n", len(admins))

	// 5. 插入 Docker 主机配置
	fmt.Println("\n【3/5】插入 Docker 主机配置...")
	dockerHosts := getDockerHosts(cfg)
	if err := db.CreateInBatches(dockerHosts, 5).Error; err != nil {
		log.Fatalf("Docker 主机插入失败: %v", err)
	}
	fmt.Printf("✓ 已插入 %d 个 Docker 主机\n", len(dockerHosts))

	// 6. 插入测试用户
	fmt.Println("\n【4/5】插入测试用户...")
	users := getTestUsers()
	if err := db.CreateInBatches(users, 10).Error; err != nil {
		log.Fatalf("用户插入失败: %v", err)
	}
	fmt.Printf("✓ 已插入 %d 个测试用户\n", len(users))

	// 7. 插入挑战题目
	fmt.Println("\n【5/5】插入挑战题目...")
	challenges := getChallenges()
	if err := db.CreateInBatches(challenges, 20).Error; err != nil {
		log.Fatalf("题目插入失败: %v", err)
	}
	fmt.Printf("✓ 已插入 %d 个挑战题目\n", len(challenges))

	// 8. 显示统计信息
	printStats(ctx, db)

	fmt.Println("\n🎉 数据填充完成！")
}

// getAdmins 返回管理员数据
func getAdmins() []model.Admin {
	now := time.Now()
	// 密码: admin123 的 bcrypt hash
	passwordHash := "$2a$10$9ZlPBt1K9LDtFbC/Qvh8GeTndMNNBZOQjhzvFH5q73NaRxoZm1aeO"

	return []model.Admin{
		{
			ID:           "admin-001",
			Username:     "admin",
			Email:        "admin@cyber-range.com",
			PasswordHash: passwordHash,
			Name:         "系统管理员",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

// getDockerHosts 返回 Docker 主机配置数据（从 config.yaml 迁移）
func getDockerHosts(cfg *config.Config) []model.DockerHost {
	now := time.Now()
	hosts := []model.DockerHost{}

	// 根据 config.yaml 中的模式创建默认主机
	if cfg.Docker.Mode == "local" || cfg.Docker.Mode == "" {
		// 本地 Docker 主机
		hosts = append(hosts, model.DockerHost{
			ID:           "docker-host-local",
			Name:         "本地 Docker",
			Host:         cfg.Docker.Local.Host,
			TLSVerify:    cfg.Docker.Local.TLSVerify,
			CertPath:     cfg.Docker.Local.CertPath,
			PortRangeMin: cfg.Docker.PortRangeMin,
			PortRangeMax: cfg.Docker.PortRangeMax,
			MemoryLimit:  cfg.Docker.MemoryLimit,
			CPULimit:     cfg.Docker.CPULimit,
			Enabled:      true,
			IsDefault:    true, // 设为默认主机
			Description:  "本地 Docker 主机（从配置文件迁移）",
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	// 如果配置中有远程主机配置，也添加进来（但不设为默认）
	if cfg.Docker.Remote.Host != "" {
		hosts = append(hosts, model.DockerHost{
			ID:           "docker-host-remote-1",
			Name:         "远程 Docker 服务器 1",
			Host:         cfg.Docker.Remote.Host,
			TLSVerify:    cfg.Docker.Remote.TLSVerify,
			CertPath:     cfg.Docker.Remote.CertPath,
			PortRangeMin: cfg.Docker.PortRangeMin,
			PortRangeMax: cfg.Docker.PortRangeMax,
			MemoryLimit:  cfg.Docker.MemoryLimit,
			CPULimit:     cfg.Docker.CPULimit,
			Enabled:      false, // 默认禁用，等待管理员启用
			IsDefault:    false,
			Description:  "远程 Docker 主机（从配置文件迁移，请测试连接后启用）",
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	// 如果两者都为空，创建一个使用环境变量的默认主机
	if len(hosts) == 0 {
		hosts = append(hosts, model.DockerHost{
			ID:           "docker-host-default",
			Name:         "默认 Docker 主机",
			Host:         "", // 留空使用环境变量
			TLSVerify:    false,
			CertPath:     "",
			PortRangeMin: cfg.Docker.PortRangeMin,
			PortRangeMax: cfg.Docker.PortRangeMax,
			MemoryLimit:  cfg.Docker.MemoryLimit,
			CPULimit:     cfg.Docker.CPULimit,
			Enabled:      true,
			IsDefault:    true,
			Description:  "默认 Docker 主机（使用环境变量配置）",
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	return hosts
}

// getTestUsers 返回测试用户数据
func getTestUsers() []model.User {
	now := time.Now()
	return []model.User{
		{
			ID:           "admin",
			Username:     "admin",
			Email:        "admin@cyber-range.com",
			PasswordHash: "$2a$10$dummyhash", // 实际应用中需要真实的bcrypt hash
			Role:         "admin",
			TotalPoints:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "user_alice",
			Username:     "alice",
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$dummyhash",
			Role:         "user",
			TotalPoints:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "user_bob",
			Username:     "bob",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$dummyhash",
			Role:         "user",
			TotalPoints:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "user_carol",
			Username:     "carol",
			Email:        "carol@example.com",
			PasswordHash: "$2a$10$dummyhash",
			Role:         "user",
			TotalPoints:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

// getChallenges 返回挑战题目数据
func getChallenges() []model.Challenge {
	now := time.Now()
	return []model.Challenge{
		// ==================== Web 题目 ====================
		{
			ID:          "web-nginx-001",
			Title:       "Nginx 配置泄露",
			Description: "在Nginx容器中找到泄露的配置文件。熟悉Linux基础命令（ls, cat, find）即可完成此挑战。",
			Category:    "Web",
			Difficulty:  "Easy",
			Image:       "nginx:alpine",
			Flag:        "flag{nginx_config_exposed}",
			Points:      100,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-apache-001",
			Title:       "Apache 目录遍历",
			Description: "利用Apache的目录浏览功能找到隐藏的Flag文件。考察对Web服务器配置的理解。",
			Category:    "Web",
			Difficulty:  "Easy",
			Image:       "httpd:2.4",
			Flag:        "flag{apache_indexing}",
			Points:      120,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-python-001",
			Title:       "Flask 调试模式",
			Description: "Python Flask应用开启了调试模式，利用此漏洞获取敏感信息。",
			Category:    "Web",
			Difficulty:  "Easy",
			Image:       "python:3.9-slim",
			Flag:        "flag{flask_debug_leak}",
			Points:      150,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-sqli-001",
			Title:       "SQL 注入登录绕过",
			Description: "经典的登录框SQL注入漏洞。使用简单的payload即可绕过身份验证，获取管理员权限。",
			Category:    "Web",
			Difficulty:  "Medium",
			Image:       "vulnerables/web-dvwa",
			Flag:        "flag{sqli_auth_bypass}",
			Points:      200,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-xss-001",
			Title:       "XSS 反射型漏洞",
			Description: "网站存在反射型XSS漏洞。构造payload窃取Cookie或执行任意JavaScript代码。",
			Category:    "Web",
			Difficulty:  "Medium",
			Image:       "vulnerables/web-dvwa",
			Flag:        "flag{xss_reflected}",
			Points:      250,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-ssrf-001",
			Title:       "SSRF 内网探测",
			Description: "利用SSRF漏洞访问内网服务，获取敏感数据。考察对HTTP协议和内网渗透的理解。",
			Category:    "Web",
			Difficulty:  "Medium",
			Image:       "python:3.9-slim",
			Flag:        "flag{ssrf_internal_access}",
			Points:      300,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-deserialization-001",
			Title:       "PHP 反序列化漏洞",
			Description: "PHP应用存在不安全的反序列化操作。构造恶意序列化数据实现RCE（远程代码执行）。",
			Category:    "Web",
			Difficulty:  "Hard",
			Image:       "php:7.4-apache",
			Flag:        "flag{php_unserialize_rce}",
			Points:      400,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "web-jwt-001",
			Title:       "JWT 加密算法混淆",
			Description: "应用使用JWT进行认证，但存在算法混淆漏洞（none/HS256）。绕过认证机制。",
			Category:    "Web",
			Difficulty:  "Hard",
			Image:       "node:16-alpine",
			Flag:        "flag{jwt_algo_confusion}",
			Points:      450,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// ==================== Pwn 题目 ====================
		{
			ID:          "pwn-suid-001",
			Title:       "SUID 程序提权",
			Description: "系统中存在配置错误的SUID二进制文件。利用此漏洞提升权限到root。",
			Category:    "Pwn",
			Difficulty:  "Medium",
			Image:       "ubuntu:20.04",
			Flag:        "flag{suid_privilege_escalation}",
			Points:      300,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "pwn-buffer-001",
			Title:       "栈溢出入门",
			Description: "简单的栈缓冲区溢出漏洞。覆盖返回地址，劫持程序执行流。",
			Category:    "Pwn",
			Difficulty:  "Medium",
			Image:       "ubuntu:20.04",
			Flag:        "flag{buffer_overflow_basic}",
			Points:      350,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "pwn-rop-001",
			Title:       "ROP 链构造",
			Description: "开启了NX保护的C程序。构造ROP链实现任意代码执行。需要掌握汇编和栈帧知识。",
			Category:    "Pwn",
			Difficulty:  "Hard",
			Image:       "ubuntu:20.04",
			Flag:        "flag{rop_chain_exploit}",
			Points:      500,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "pwn-kernel-001",
			Title:       "Linux 内核提权",
			Description: "利用内核模块漏洞实现从普通用户到root的提权。高难度挑战，需要深入理解Linux内核。",
			Category:    "Pwn",
			Difficulty:  "Hard",
			Image:       "ubuntu:20.04",
			Flag:        "flag{kernel_privilege_escalation}",
			Points:      600,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// ==================== Crypto 题目 ====================
		{
			ID:          "crypto-base64-001",
			Title:       "Base64 多重编码",
			Description: "Flag经过多次Base64编码。逐层解码即可获取明文。适合密码学入门。",
			Category:    "Crypto",
			Difficulty:  "Easy",
			Image:       "alpine:latest",
			Flag:        "flag{base64_layered_encoding}",
			Points:      100,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "crypto-caesar-001",
			Title:       "凯撒密码变种",
			Description: "使用改进的凯撒密码加密Flag。需要暴力破解或频率分析。",
			Category:    "Crypto",
			Difficulty:  "Medium",
			Image:       "python:3.9-slim",
			Flag:        "flag{caesar_cipher_variant}",
			Points:      250,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "crypto-rsa-001",
			Title:       "RSA 低加密指数攻击",
			Description: "RSA加密使用了较小的公钥指数e=3。利用低加密指数攻击恢复明文。",
			Category:    "Crypto",
			Difficulty:  "Hard",
			Image:       "python:3.9-slim",
			Flag:        "flag{rsa_low_exponent_attack}",
			Points:      500,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "crypto-aes-001",
			Title:       "AES ECB 模式缺陷",
			Description: "AES使用了不安全的ECB模式。利用ECB模式的特性（相同明文块产生相同密文块）破解加密。",
			Category:    "Crypto",
			Difficulty:  "Hard",
			Image:       "python:3.9-slim",
			Flag:        "flag{aes_ecb_pattern_attack}",
			Points:      550,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		// ==================== Reverse 题目 ====================
		{
			ID:          "reverse-strings-001",
			Title:       "字符串隐写",
			Description: "二进制文件中隐藏了Flag字符串。使用strings命令即可找到。",
			Category:    "Reverse",
			Difficulty:  "Easy",
			Image:       "alpine:latest",
			Flag:        "flag{strings_command_find}",
			Points:      100,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "reverse-upx-001",
			Title:       "UPX 加壳程序",
			Description: "可执行文件被UPX加壳。脱壳后逆向分析获取Flag验证逻辑。",
			Category:    "Reverse",
			Difficulty:  "Medium",
			Image:       "ubuntu:20.04",
			Flag:        "flag{upx_unpacked_binary}",
			Points:      300,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "reverse-obfuscation-001",
			Title:       "代码混淆与反混淆",
			Description: "Python代码经过重度混淆。需要理解混淆技术并还原原始逻辑。",
			Category:    "Reverse",
			Difficulty:  "Hard",
			Image:       "python:3.9-slim",
			Flag:        "flag{deobfuscation_master}",
			Points:      450,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// printStats 打印统计信息
func printStats(ctx context.Context, db *gorm.DB) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 数据统计")
	fmt.Println(strings.Repeat("=", 50))

	// 按分类统计
	var categoryStats []struct {
		Category string
		Count    int64
	}
	db.Model(&model.Challenge{}).Select("category, COUNT(*) as count").
		Group("category").Find(&categoryStats)

	fmt.Println("\n分类分布：")
	for _, stat := range categoryStats {
		fmt.Printf("  %s: %d 个题目\n", stat.Category, stat.Count)
	}

	// 按难度统计
	var difficultyStats []struct {
		Difficulty string
		Count      int64
	}
	db.Model(&model.Challenge{}).Select("difficulty, COUNT(*) as count").
		Group("difficulty").Find(&difficultyStats)

	fmt.Println("\n难度分布：")
	for _, stat := range difficultyStats {
		fmt.Printf("  %s: %d 个题目\n", stat.Difficulty, stat.Count)
	}

	// 总分值
	var totalPoints int64
	db.Model(&model.Challenge{}).Select("SUM(points)").Scan(&totalPoints)
	fmt.Printf("\n总分值：%d 分\n", totalPoints)

	fmt.Println(strings.Repeat("=", 50))
}
