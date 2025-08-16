package main

import (
	"log"

	"github.com/DullJZ/docker-manager/api"
	"github.com/DullJZ/docker-manager/auth"
	"github.com/DullJZ/docker-manager/config"
	"github.com/DullJZ/docker-manager/service"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化Token管理器
	tokenManager := auth.NewTokenManager()
	if err := tokenManager.LoadFromFile(cfg.TokenAuthFile); err != nil {
		log.Fatalf("Failed to load token whitelist: %v", err)
	}

	// 初始化Docker服务
	dockerService := service.NewDockerService("")

	// 初始化路由器
	router := api.NewRouter(tokenManager, dockerService)

	// 启动服务器
	log.Printf("Docker Manager is starting on %s", cfg.GetAddress())
	if err := router.Run(cfg.GetAddress()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
