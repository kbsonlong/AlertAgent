package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/infrastructure/config"
	"alert_agent/internal/infrastructure/database"
	"alert_agent/internal/infrastructure/di"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/types"
	"alert_agent/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// processUnanalyzedAlerts 处理未分析的告警
func processUnanalyzedAlerts(ctx context.Context, container *di.Container) error {
	logger := container.GetLogger()
	db := container.GetDB()
	// TODO: 实现队列获取方法

	var alerts []model.Alert
	logger.Debug("Processing unanalyzed alerts...")
	
	// 获取未分析的告警
	if err := db.Where("analysis = ? OR analysis IS NULL", "").Find(&alerts).Error; err != nil {
		return fmt.Errorf("failed to get unanalyzed alerts: %w", err)
	}

	if len(alerts) == 0 {
		logger.Info("No unanalyzed alerts found")
		return nil
	}

	logger.Info("Found unanalyzed alerts", zap.Int("count", len(alerts)))

	// 创建任务列表
	tasks := make([]*types.AlertTask, len(alerts))
	for i, alert := range alerts {
		logger.Debug("Processing alert", zap.Uint("id", alert.ID))
		tasks[i] = &types.AlertTask{
			ID:        alert.ID,
			CreatedAt: time.Now(),
		}
	}

	// TODO: 实现队列推送逻辑
	// 批量推送任务到队列
	// if err := queue.PushBatch(ctx, tasks); err != nil {
	//	return fmt.Errorf("failed to push tasks to queue: %w", err)
	// }

	logger.Info("Successfully processed tasks", zap.Int("count", len(tasks)))
	return nil
}

func main() {
	// 初始化日志器
	logger := logger.NewLogger()
	defer logger.Sync()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// 初始化数据库
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// 运行数据库迁移
	if err := database.Migrate(db); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 测试 Redis 连接
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := redisClient.Ping(ctx2).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}

	// 创建依赖注入容器
	container := di.NewContainer(db, redisClient, logger, cfg)
	defer container.Close()

	// 处理未分析的告警
	ctx := context.Background()
	if err := processUnanalyzedAlerts(ctx, container); err != nil {
		logger.Error("Failed to process unanalyzed alerts", zap.Error(err))
	}

	// TODO: 实现工作器获取和启动逻辑
	// 获取工作器
	// worker := container.GetWorker()

	// 启动工作器
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		logger.Info("Starting worker...")
		// TODO: 实现工作器启动逻辑
		// if err := worker.Start(workerCtx); err != nil {
		//	logger.Error("Worker failed", zap.Error(err))
		// }
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-time.After(10 * time.Second):
				logger.Debug("Worker heartbeat")
			}
		}
	}()

	logger.Info("Worker started successfully")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止工作器
	workerCancel()

	// 等待工作器完全停止
	select {
	case <-ctx.Done():
		logger.Warn("Worker shutdown timeout")
	case <-time.After(5 * time.Second):
		logger.Info("Worker shutdown completed")
	}

	logger.Info("Worker exited")
}