package main

import (
	"flag"
	"fmt"
	"log"

	"alert_agent/internal/config"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/router"

	"github.com/gin-gonic/gin"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "config/config.yaml", "path to config file")
}

func main() {
	flag.Parse()

	// 加载配置
	if err := config.Init(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis连接
	if err := redis.Init(); err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	// 设置gin模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 创建gin引擎
	r := gin.Default()

	// 注册路由
	router.RegisterRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
