package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/service"
	"alert_agent/internal/worker"

	"go.uber.org/zap"
)

var (
	workerName   = flag.String("name", "", "Worker名称")
	workerType   = flag.String("type", "general", "Worker类型 (ai-analysis, notification, config-sync, general)")
	concurrency  = flag.Int("concurrency", 2, "并发数")
	queues       = flag.String("queues", "", "队列名称，多个用逗号分隔")
	configPath   = flag.String("config", "config/config.yaml", "配置文件路径")
	logLevel     = flag.String("log-level", "info", "日志级别")
	healthPort   = flag.Int("health-port", 8081, "健康检查端口")
)

func main() {
	flag.Parse()

	// 验证必需参数
	if *workerName == "" {
		fmt.Println("Error: worker name is required")
		flag.Usage()
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(*logLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.L.Info("Starting AlertAgent Worker",
		zap.String("name", *workerName),
		zap.String("type", *workerType),
		zap.Int("concurrency", *concurrency),
		zap.String("queues", *queues),
	)

	// 加载配置
	config.SetConfigPath(*configPath)
	if err := config.Load(); err != nil {
		logger.L.Fatal("Failed to load config", zap.Error(err))
	}

	// 初始化数据库连接
	if err := database.Init(); err != nil {
		logger.L.Fatal("Failed to initialize database", zap.Error(err))
	}

	// 初始化Redis连接
	if err := redis.Init(); err != nil {
		logger.L.Fatal("Failed to initialize redis", zap.Error(err))
	}

	// 创建Worker管理器
	workerManager := worker.NewWorkerManager()

	// 解析队列名称
	var queueNames []string
	if *queues != "" {
		queueNames = strings.Split(*queues, ",")
		for i, name := range queueNames {
			queueNames[i] = strings.TrimSpace(name)
		}
	} else {
		// 根据Worker类型设置默认队列
		queueNames = getDefaultQueues(*workerType)
	}

	// 创建Worker配置
	workerConfig := &worker.WorkerConfig{
		Name:        *workerName,
		Type:        *workerType,
		Concurrency: *concurrency,
		Queues:      queueNames,
		HealthPort:  *healthPort,
	}

	// 创建并启动Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerInstance, err := workerManager.CreateWorker(ctx, workerConfig)
	if err != nil {
		logger.L.Fatal("Failed to create worker", zap.Error(err))
	}

	// 启动Worker
	if err := workerInstance.Start(ctx); err != nil {
		logger.L.Fatal("Failed to start worker", zap.Error(err))
	}

	logger.L.Info("Worker started successfully",
		zap.String("name", *workerName),
		zap.Strings("queues", queueNames),
		zap.Int("health_port", *healthPort),
	)

	// 等待关闭信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.L.Info("Received shutdown signal, stopping worker...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := workerInstance.Stop(shutdownCtx); err != nil {
		logger.L.Error("Error during worker shutdown", zap.Error(err))
	} else {
		logger.L.Info("Worker stopped gracefully")
	}
}

// getDefaultQueues 根据Worker类型获取默认队列
func getDefaultQueues(workerType string) []string {
	switch workerType {
	case "ai-analysis":
		return []string{string(queue.TaskTypeAIAnalysis)}
	case "notification":
		return []string{string(queue.TaskTypeNotification)}
	case "config-sync":
		return []string{string(queue.TaskTypeConfigSync)}
	case "general":
		return []string{
			string(queue.TaskTypeRuleUpdate),
			string(queue.TaskTypeHealthCheck),
		}
	default:
		return []string{
			string(queue.TaskTypeAIAnalysis),
			string(queue.TaskTypeNotification),
			string(queue.TaskTypeConfigSync),
		}
	}
}