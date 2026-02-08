package main

import (
	"cyber-range/internal/model"
	"cyber-range/pkg/config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("🔄 开始表结构迁移（添加中文注释）...")
	fmt.Println("⚠️  警告：此操作会删除所有现有数据！")
	fmt.Println("")

	// 删除旧表
	fmt.Println("【1/2】删除旧表...")
	db.Exec("DROP TABLE IF EXISTS submissions")
	db.Exec("DROP TABLE IF EXISTS instances")
	db.Exec("DROP TABLE IF EXISTS challenges")
	db.Exec("DROP TABLE IF EXISTS docker_images") // 新增
	db.Exec("DROP TABLE IF EXISTS docker_hosts")
	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec("DROP TABLE IF EXISTS admins")
	fmt.Println("✓ 旧表已删除")

	// 重新创建表（带中文注释）
	fmt.Println("\n【2/2】创建新表（带中文注释）...")
	if err := db.AutoMigrate(
		&model.DockerHost{},
		&model.DockerImage{}, // 新增镜像管理表
		&model.Challenge{},
		&model.Instance{},
		&model.User{},
		&model.Submission{},
		&model.Admin{},
	); err != nil {
		log.Fatalf("表创建失败: %v", err)
	}
	fmt.Println("✓ 新表创建完成")

	// 验证表结构
	fmt.Println("\n" + repeat("=", 70))
	fmt.Println("📊 验证表结构")
	fmt.Println(repeat("=", 70))

	tables := []string{"docker_hosts", "challenges", "instances", "users", "submissions", "admins"}
	for _, table := range tables {
		var createSQL string
		db.Raw(fmt.Sprintf("SHOW CREATE TABLE %s", table)).Scan(&createSQL)
		fmt.Printf("\n表: %s ✓\n", table)
	}

	fmt.Println("\n🎉 迁移完成！现在表和字段都有中文注释了。")
	fmt.Println("💡 提示：请运行 go run cmd/seed/main.go 重新填充数据")
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
