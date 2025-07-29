package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	configsyncer "alert_agent/pkg/config-syncer"
	"alert_agent/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 解析环境变量
	alertAgentEndpoint := os.Getenv("ALERTAGENT_ENDPOINT")
	if alertAgentEndpoint == "" {
		panic("ALERTAGENT_ENDPOINT environment variable is required")
	}

	clusterID := os.Getenv("CLUSTER_ID")
	if clusterID == "" {
		panic("CLUSTER_ID environment variable is required")
	}

	configType := os.Getenv("CONFIG_TYPE")
	if configType == "" {
		panic("CONFIG_TYPE environment variable is required")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH environment variable is required")
	}

	reloadURL := os.Getenv("RELOAD_URL")
	if reloadURL == "" {
		panic("RELOAD_URL environment variable is required")
	}

	syncIntervalStr := os.Getenv("SYNC_INTERVAL")
	if syncIntervalStr == "" {
		syncIntervalStr = "30s"
	}
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		panic("Invalid SYNC_INTERVAL: " + err.Error())
	}

	// 解析HTTP服务器端口
	httpPortStr := os.Getenv("HTTP_PORT")
	if httpPortStr == "" {
		httpPortStr = "8080"
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		panic("Invalid HTTP_PORT: " + err.Error())
	}

	// 解析HTTP超时
	httpTimeoutStr := os.Getenv("HTTP_TIMEOUT")
	if httpTimeoutStr == "" {
		httpTimeoutStr = "30s"
	}
	httpTimeout, err := time.ParseDuration(httpTimeoutStr)
	if err != nil {
		panic("Invalid HTTP_TIMEOUT: " + err.Error())
	}

	// 解析最大重试次数
	maxRetriesStr := os.Getenv("MAX_RETRIES")
	if maxRetriesStr == "" {
		maxRetriesStr = "3"
	}
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		panic("Invalid MAX_RETRIES: " + err.Error())
	}

	// 解析重试退避时间
	retryBackoffStr := os.Getenv("RETRY_BACKOFF")
	if retryBackoffStr == "" {
		retryBackoffStr = "5s"
	}
	retryBackoff, err := time.ParseDuration(retryBackoffStr)
	if err != nil {
		panic("Invalid RETRY_BACKOFF: " + err.Error())
	}

	// 初始化日志器
	logger := logger.NewLogger()

	// 创建配置同步器
	syncer := configsyncer.NewConfigSyncer(&configsyncer.Config{
		AlertAgentEndpoint: alertAgentEndpoint,
		ClusterID:         clusterID,
		ConfigType:        configType,
		ConfigPath:        configPath,
		ReloadURL:         reloadURL,
		SyncInterval:      syncInterval,
		HTTPTimeout:       httpTimeout,
		MaxRetries:        maxRetries,
		RetryBackoff:      retryBackoff,
		Logger:            logger,
	})

	// 验证配置
	if err := syncer.ValidateConfig(); err != nil {
		panic("Invalid config: " + err.Error())
	}

	// 创建HTTP服务器
	httpServer := configsyncer.NewHTTPServer(syncer, httpPort)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// 启动HTTP服务器
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// 启动配置同步器
	logger.Info("Starting config syncer",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
		zap.Duration("sync_interval", syncInterval),
		zap.Int("http_port", httpPort))

	if err := syncer.Start(ctx); err != nil {
		logger.Error("Config syncer failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Config syncer stopped")
}