package main

import (
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

	result := db.Exec("UPDATE docker_hosts SET enabled = false WHERE id = 'docker-host-remote-1'")
	if result.Error != nil {
		log.Fatalf("更新失败: %v", result.Error)
	}
	fmt.Printf("已禁用远程 Docker 主机，影响行数: %d\n", result.RowsAffected)
}
