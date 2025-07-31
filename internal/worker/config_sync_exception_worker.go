package worker

import (
	"context"
	"time"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// ConfigSyncExceptionWorker 配置同步异常处理Worker
type ConfigSyncExceptionWorker struct {
	exceptionHandler *service.ConfigSyncExceptionHandler
	interval         time.Duration
	stopCh           chan struct{}
}

// NewConfigSyncExceptionWorker 创建配置同步异常处理Worker
func NewConfigSyncExceptionWorker(interval time.Duration) *ConfigSyncExceptionWorker {
	if interval <= 0 {
		interval = 5 * time.Minute // 默认5分钟检查一次
	}

	return &ConfigSyncExceptionWorker{
		exceptionHandler: service.NewConfigSyncExceptionHandler(),
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

// Start 启动Worker
func (w *ConfigSyncExceptionWorker) Start(ctx context.Context) error {
	logger.L.Info("Starting config sync exception worker",
		zap.Duration("interval", w.interval),
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 启动时立即执行一次检测
	if err := w.detectAndHandleExceptions(ctx); err != nil {
		logger.L.Error("Initial exception detection failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			logger.L.Info("Config sync exception worker stopped due to context cancellation")
			return ctx.Err()
		case <-w.stopCh:
			logger.L.Info("Config sync exception worker stopped")
			return nil
		case <-ticker.C:
			if err := w.detectAndHandleExceptions(ctx); err != nil {
				logger.L.Error("Exception detection failed", zap.Error(err))
			}
		}
	}
}

// Stop 停止Worker
func (w *ConfigSyncExceptionWorker) Stop() {
	close(w.stopCh)
}

// detectAndHandleExceptions 检测和处理异常
func (w *ConfigSyncExceptionWorker) detectAndHandleExceptions(ctx context.Context) error {
	logger.L.Debug("Detecting sync exceptions")

	startTime := time.Now()
	
	// 检测同步异常
	if err := w.exceptionHandler.DetectSyncExceptions(ctx); err != nil {
		logger.L.Error("Failed to detect sync exceptions", zap.Error(err))
		return err
	}

	duration := time.Since(startTime)
	logger.L.Debug("Exception detection completed",
		zap.Duration("duration", duration),
	)

	return nil
}

// GetStatus 获取Worker状态
func (w *ConfigSyncExceptionWorker) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":     "config_sync_exception_worker",
		"interval": w.interval.String(),
		"running":  true,
	}
}