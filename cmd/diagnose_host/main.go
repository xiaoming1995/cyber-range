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

	targetID := "docker-host-local"
	var host model.DockerHost

	fmt.Printf("正在查询 Docker Host ID: %s ...\n", targetID)
	err = db.Where("id = ?", targetID).First(&host).Error
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	// 打印关键字段
	fmt.Println("--------------------------------------------------")
	fmt.Printf("ID:        %s\n", host.ID)
	fmt.Printf("Name:      %s\n", host.Name)
	fmt.Printf("Host:      %s\n", host.Host)
	fmt.Printf("IsDefault: %v\n", host.IsDefault)
	fmt.Printf("Enabled:   %v\n", host.Enabled)
	fmt.Println("--------------------------------------------------")

	// 完整 JSON
	jsonData, _ := json.MarshalIndent(host, "", "  ")
	fmt.Println("完整数据 JSON:")
	fmt.Println(string(jsonData))
}
