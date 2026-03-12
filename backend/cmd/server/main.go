package main

import (
	"log"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/router"
)

func main() {
	log.Println("启动排班系统...")

	// 检查是否已安装
	isInstalled := config.IsInstalled()

	if isInstalled {
		log.Println("系统已安装，加载配置...")
		cfg, err := config.LoadConfig(config.ConfigFilePath)
		if err != nil {
			log.Fatalf("加载配置失败: %v", err)
		}

		if err := db.InitDB(&cfg.Database); err != nil {
			log.Fatalf("数据库初始化失败: %v", err)
		}
		defer db.Close()
	} else {
		log.Println("系统未安装，请先完成安装向导")
	}

	// 统一使用同一个路由
	r := router.SetupRouter()

	log.Println("服务器启动，监听端口: 8080")
	log.Println("访问地址: http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
