package main

import (
	"cyber-range/internal/model"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:123456@tcp(localhost:3306)/cyber_range?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	fmt.Println("正在执行数据库迁移...")
	if err := db.AutoMigrate(&model.Challenge{}); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	fmt.Println("数据库迁移完成！")

	// 同时更新题目为特权模式
	challengeID := "b6aadbf0-523b-4fd8-aad3-85a8677452d3"
	result := db.Exec("UPDATE challenges SET privileged = true WHERE id = ?", challengeID)
	if result.Error != nil {
		log.Fatalf("更新失败: %v", result.Error)
	}
	fmt.Printf("已将题目 %s 设为特权模式，影响行数: %d\n", challengeID, result.RowsAffected)
}
