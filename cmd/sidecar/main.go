package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/sidecar"

	"go.uber.org/zap"
)

func main() {
	// 命令行参数
	var (
		alertAgentEndpoint = flag.String("endpoint", "http://localhost:8080", "AlertAgent API endpoint")
		clusterID          = flag.String("cluster-id", "", "Cluster ID for this sidecar")
		configType         = flag.String("type", "", "Config type: prometheus, alertmanager, vmalert")
		configPath         = flag.String("config-path", "", "Path to write config file")
		reloadURL          = flag.String("reload-url", "", "URL to trigger reload")
		syncInterval       = flag.Duration("sync-interval", 30*time.Second, "Config sync interval")
		healthPort         = flag.Int("health-port", 8081, "Health check port")
		logLevel           = flag.String("log-level", "info", "Log level")
	)
	flag.Parse()

	// 验证必需参数
	if *clusterID == "" {
		fmt.Fprintf(os.Stderr, "Error: cluster-id is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *configType == "" {
		fmt.Fprintf(os.Stderr, "Error: type is required (prometheus, alertmanager, vmalert)\n")
		flag.Usage()
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: config-path is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *reloadURL == "" {
		fmt.Fprintf(os.Stderr, "Error: reload-url is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(*logLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.L.Info("Starting AlertAgent Sidecar",
		zap.String("endpoint", *alertAgentEndpoint),
		zap.String("cluster_id", *clusterID),
		zap.String("config_type", *configType),
		zap.String("config_path", *configPath),
		zap.String("reload_url", *reloadURL),
		zap.Duration("sync_interval", *syncInterval),
		zap.Int("health_port", *healthPort),
	)

	// 创建配置同步器
	syncer := sidecar.NewConfigSyncer(&sidecar.Config{
		AlertAgentEndpoint: *alertAgentEndpoint,
		ClusterID:          *clusterID,
		ConfigType:         *configType,
		ConfigPath:         *configPath,
		ReloadURL:          *reloadURL,
		SyncInterval:       *syncInterval,
	})

	// 创建上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动配置同步器
	go func() {
		if err := syncer.Start(ctx); err != nil && err != context.Canceled {
			logger.L.Error("Config syncer failed", zap.Error(err))
			cancel()
		}
	}()

	// 等待退出信号
	select {
	case sig := <-sigCh:
		logger.L.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	case <-ctx.Done():
		logger.L.Info("Context cancelled, shutting down")
	}

	// 优雅关闭
	cancel()
	logger.L.Info("Sidecar stopped")
}
