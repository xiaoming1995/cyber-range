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
	dsn := "root:123456@tcp(localhost:3306)/cyber_range?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	var hosts []model.DockerHost

	fmt.Println("正在查询所有 Docker Hosts...")
	err = db.Find(&hosts).Error
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	fmt.Printf("共找到 %d 个 Docker Host:\n", len(hosts))
	fmt.Println("--------------------------------------------------")

	for _, host := range hosts {
		jsonData, _ := json.MarshalIndent(host, "", "  ")
		fmt.Println(string(jsonData))
		fmt.Println("--------------------------------------------------")
	}
}
