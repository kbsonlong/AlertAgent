// Package main AlertAgent API服务
// @title AlertAgent API
// @version 1.0
// @description AlertAgent告警管理系统API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"fmt"
	"os"

	"alert_agent/internal/config"
	"alert_agent/internal/gateway"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/router"
	"alert_agent/internal/service"
	"alert_agent/internal/worker"
	"alert_agent/pkg/utils"

	"go.uber.org/zap"

	// Swagger docs
	_ "alert_agent/docs"
)




func main() {
	// 初始化日志
	if err := logger.Init("info"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 重新生成Swagger文档
	if err := utils.RegenerateSwaggerDocs(); err != nil {
		logger.L.Error("Failed to regenerate Swagger docs", zap.Error(err))
		// 不中断启动，继续执行
	}

	// 加载配置
	if err := config.Load(); err != nil {
		logger.L.Fatal("Failed to load config", zap.Error(err))
		os.Exit(1)
	}

	// 注册配置重载回调函数
	config.RegisterReloadCallback(func(newConfig config.Config) {
		logger.L.Debug("Configuration reloaded",
			zap.Int("port", newConfig.Server.Port),
			zap.String("mode", newConfig.Server.Mode),
		)
		// 这里可以添加其他需要在配置更新时执行的逻辑
		// 例如：重新初始化某些服务、更新日志级别等
	})

	// 初始化数据库连接
	if err := database.Init(); err != nil {
		logger.L.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 初始化Redis连接
	if err := redis.Init(); err != nil {
		logger.L.Fatal("Failed to initialize redis", zap.Error(err))
	}

	// 创建Redis队列
	redisQueue := queue.NewRedisMessageQueue(redis.Client, "alert:queue")

	// 处理未分析的告警
	ctx := context.Background()
	if err := utils.ProcessUnanalyzedAlerts(ctx, redisQueue); err != nil {
		logger.L.Error("Failed to process unanalyzed alerts", zap.Error(err))
	}

	// 创建Ollama服务
	ollamaService := service.NewOllamaService()

	// 创建Dify服务
	difyService := service.NewDifyService()

	// 创建队列监控器
	monitor := queue.NewQueueMonitor(redisQueue, redis.Client, "alert:queue")

	// 创建工作器
	queueWorker := queue.NewWorker(redisQueue, monitor, config.GetConfig().Worker.Concurrency)

	// 注册任务处理器
	aiHandler := worker.NewAIAnalysisHandler(ollamaService, difyService)
	queueWorker.RegisterHandler(aiHandler)
	
	notificationHandler := worker.NewNotificationHandler()
	queueWorker.RegisterHandler(notificationHandler)
	
	configSyncHandler := worker.NewConfigSyncHandler()
	queueWorker.RegisterHandler(configSyncHandler)

	logger.L.Info("AlertAgent Core initialized successfully")

	// 创建API网关
	gateway := gateway.NewGateway()
	
	// 设置中间件
	gateway.SetupMiddleware()
	
	// 设置路由
	gateway.SetupRoutes(router.RegisterRoutes)

	// 根据配置决定是否启动工作器
	var workerCtx context.Context
	var workerCancel context.CancelFunc
	
	if config.GetConfig().Worker.Enabled {
		logger.L.Info("Worker is enabled, starting queue worker...", 
			zap.Int("concurrency", config.GetConfig().Worker.Concurrency))
		
		workerCtx, workerCancel = context.WithCancel(context.Background())
		queueNames := []string{
			string(queue.TaskTypeAIAnalysis),
			string(queue.TaskTypeNotification),
			string(queue.TaskTypeConfigSync),
		}
		
		go func() {
			if err := queueWorker.Start(workerCtx, queueNames); err != nil {
				logger.L.Error("Worker failed", zap.Error(err))
			}
		}()
	} else {
		logger.L.Info("Worker is disabled, skipping queue worker startup")
	}

	// 启动API网关
	if err := gateway.Start(); err != nil {
		logger.L.Fatal("Failed to start API Gateway", zap.Error(err))
	}

	// 等待关闭信号并优雅关闭
	gateway.WaitForShutdown()

	// 停止工作器（如果已启动）
	if workerCancel != nil {
		logger.L.Info("Stopping queue worker...")
		workerCancel()
	}

	// 停止配置文件监听
	config.StopWatcher()
}
