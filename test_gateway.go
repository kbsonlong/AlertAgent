package main

import (
	"alert_agent/internal/config"
	"alert_agent/internal/gateway"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/router"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	if err := logger.Init("info"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	// 加载配置
	if err := config.Load(); err != nil {
		logger.L.Fatal("Failed to load config", zap.Error(err))
		return
	}

	// 创建API网关
	gw := gateway.NewGateway()
	
	// 设置中间件
	gw.SetupMiddleware()
	
	// 设置路由
	gw.SetupRoutes(router.RegisterRoutes)

	// 启动网关
	if err := gw.Start(); err != nil {
		logger.L.Fatal("Failed to start gateway", zap.Error(err))
		return
	}

	logger.L.Info("Gateway started successfully")

	// 运行5秒后停止
	time.Sleep(5 * time.Second)

	// 停止网关
	if err := gw.Stop(); err != nil {
		logger.L.Error("Failed to stop gateway", zap.Error(err))
	}

	logger.L.Info("Gateway stopped successfully")
}