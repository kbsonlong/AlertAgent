package main

import (
	"context"
	"fmt"
	"os"

	"alert_agent/internal/config"
	"alert_agent/internal/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/router"
	"alert_agent/internal/service"
	"alert_agent/internal/worker"

	"go.uber.org/zap"
)

// processUnanalyzedAlerts 处理未分析的告警
func processUnanalyzedAlerts(ctx context.Context, messageQueue queue.MessageQueue) error {
	var alerts []model.Alert
	logger.L.Debug("Processing unanalyzed alerts...")
	// 获取未分析的告警
	if err := database.DB.Where("analysis = ?", "").Find(&alerts).Error; err != nil {
		return fmt.Errorf("failed to get unanalyzed alerts: %w", err)
	}

	if len(alerts) == 0 {
		logger.L.Info("No unanalyzed alerts found")
		return nil
	}

	logger.L.Info("Found unanalyzed alerts",
		zap.Int("count", len(alerts)),
	)

	// 创建任务生产者
	producer := queue.NewTaskProducer(messageQueue)

	// 为每个告警创建AI分析任务
	for _, alert := range alerts {
		logger.L.Debug("Processing alert",
			zap.Uint("id", alert.ID),
		)
		
		alertData := map[string]interface{}{
			"title":   alert.Title,
			"level":   alert.Level,
			"source":  alert.Source,
			"content": alert.Content,
		}
		
		if err := producer.PublishAIAnalysisTask(ctx, fmt.Sprintf("%d", alert.ID), alertData); err != nil {
			logger.L.Error("Failed to publish AI analysis task",
				zap.Uint("alert_id", alert.ID),
				zap.Error(err),
			)
			continue
		}
	}

	logger.L.Info("Successfully pushed tasks to queue",
		zap.Int("count", len(alerts)),
	)

	return nil
}

func main() {
	// 初始化日志
	if err := logger.Init("info"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
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
	if err := processUnanalyzedAlerts(ctx, redisQueue); err != nil {
		logger.L.Error("Failed to process unanalyzed alerts", zap.Error(err))
	}

	// 创建Ollama服务
	ollamaService := service.NewOllamaService()

	// 创建队列监控器
	monitor := queue.NewQueueMonitor(redisQueue, redis.Client, "alert:queue")

	// 创建工作器
	queueWorker := queue.NewWorker(redisQueue, monitor, 2)

	// 注册任务处理器
	aiHandler := worker.NewAIAnalysisHandler(ollamaService)
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

	// 启动工作器
	workerCtx, workerCancel := context.WithCancel(context.Background())
	queueNames := []string{
		string(queue.TaskTypeAIAnalysis),
		string(queue.TaskTypeNotification),
		string(queue.TaskTypeConfigSync),
	}
	
	go func() {
		logger.L.Info("Starting queue worker...")
		if err := queueWorker.Start(workerCtx, queueNames); err != nil {
			logger.L.Error("Worker failed", zap.Error(err))
		}
	}()

	// 启动API网关
	if err := gateway.Start(); err != nil {
		logger.L.Fatal("Failed to start API Gateway", zap.Error(err))
	}

	// 等待关闭信号并优雅关闭
	gateway.WaitForShutdown()

	// 停止工作器
	workerCancel()

	// 停止配置文件监听
	config.StopWatcher()
}
