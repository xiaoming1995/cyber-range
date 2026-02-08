package main

import (
	"cyber-range/internal/model"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 硬编码简单配置，用于快速诊断
	dsn := "root:123456@tcp(localhost:3306)/cyber_range?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	targetID := "b6aadbf0-523b-4fd8-aad3-85a8677452d3"
	var challenge model.Challenge

	fmt.Printf("正在查询 ID: %s ...\n", targetID)
	err = db.Where("id = ?", targetID).First(&challenge).Error
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	// 打印关键字段
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Title:      %s\n", challenge.Title)
	fmt.Printf("Category:   %s\n", challenge.Category)
	fmt.Printf("Difficulty: %s\n", challenge.Difficulty)
	fmt.Printf("Image:      %s\n", challenge.Image)
	fmt.Printf("ImageID:    %s\n", challenge.ImageID)
	fmt.Printf("Flag:       %s\n", challenge.Flag)
	fmt.Printf("Desc HTML (len): %d\n", len(challenge.Description))
	fmt.Println("--------------------------------------------------")

	// 完整 JSON
	jsonData, _ := json.MarshalIndent(challenge, "", "  ")
	fmt.Println("完整数据 JSON:")
	fmt.Println(string(jsonData))
}
