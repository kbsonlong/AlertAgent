package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/pkg/types"
	"alert_agent/internal/router"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
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

	// 创建功能开关服务
	featureService, err := service.NewFeatureService(logger.L)
	if err != nil {
		logger.L.Fatal("Failed to initialize feature service", zap.Error(err))
	}
	
	// 确保在程序退出时关闭功能服务
	defer func() {
		if featureService != nil {
			featureService.Shutdown()
		}
	}()

	// 创建工作器
	worker := queue.NewWorker(redisQueue, ollamaService)

	// 创建Gin引擎
	engine := gin.Default()

	// 注册路由
	router.RegisterRoutes(engine, featureService)

	// 创建HTTP服务器
	cfg := config.GetConfig()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
	}

	// 启动工作器
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		if err := worker.Start(workerCtx); err != nil {
			logger.L.Error("Worker failed", zap.Error(err))
		}
	}()

	// 启动HTTP服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止工作器
	workerCancel()

	// 停止配置文件监听
	config.StopWatcher()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Fatal("Failed to shutdown server", zap.Error(err))
	}
}
