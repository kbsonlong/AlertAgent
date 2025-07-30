package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/gateway"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/pkg/types"
	"alert_agent/internal/router"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// processUnanalyzedAlerts 处理未分析的告警
func processUnanalyzedAlerts(ctx context.Context, redisQueue *queue.RedisQueue) error {
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

	// 创建任务列表
	tasks := make([]*types.AlertTask, len(alerts))
	for i, alert := range alerts {
		logger.L.Debug("Processing alert",
			zap.Uint("id", alert.ID),
		)
		tasks[i] = &types.AlertTask{
			ID:        alert.ID,
			CreatedAt: time.Now(),
		}
	}

	// 批量推送任务到队列
	if err := redisQueue.PushBatch(ctx, tasks); err != nil {
		return fmt.Errorf("failed to push tasks to queue: %w", err)
	}

	logger.L.Info("Successfully pushed tasks to queue",
		zap.Int("count", len(tasks)),
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
	redisQueue := queue.NewRedisQueue(redis.Client, "alert:queue")

	// 处理未分析的告警
	ctx := context.Background()
	if err := processUnanalyzedAlerts(ctx, redisQueue); err != nil {
		logger.L.Error("Failed to process unanalyzed alerts", zap.Error(err))
	}

	// 创建Ollama服务
	ollamaService := service.NewOllamaService()

	// 创建工作器
	worker := queue.NewWorker(redisQueue, ollamaService)

	// 创建API网关
	gateway := gateway.NewGateway()
	
	// 设置中间件
	gateway.SetupMiddleware()
	
	// 设置路由
	gateway.SetupRoutes(router.RegisterRoutes)

	// 启动工作器
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		if err := worker.Start(workerCtx); err != nil {
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
