package channel

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"alert_agent/internal/domain/channel"
)

// HealthMonitor 健康监控器
type HealthMonitor struct {
	logger    *zap.Logger
	statuses  map[string]*channel.HealthStatus
	mutex     sync.RWMutex
	stopChan  chan struct{}
	running   bool
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(logger *zap.Logger) *HealthMonitor {
	return &HealthMonitor{
		logger:   logger,
		statuses: make(map[string]*channel.HealthStatus),
		stopChan: make(chan struct{}),
	}
}

// CheckChannel 检查单个渠道健康状态
func (hm *HealthMonitor) CheckChannel(ctx context.Context, channelID string, manager channel.ChannelManager) (*channel.HealthStatus, error) {
	start := time.Now()

	// 获取渠道信息
	ch, err := manager.GetChannel(ctx, channelID)
	if err != nil {
		status := &channel.HealthStatus{
			ChannelID:     channelID,
			Status:        channel.HealthStatusUnhealthy,
			Message:       "Channel not found",
			LastCheck:     time.Now(),
			ResponseTime:  time.Since(start),
			LastError:     err.Error(),
			LastErrorTime: &start,
		}
		hm.updateStatus(channelID, status)
		return status, nil
	}

	// 获取插件
	plugin, err := manager.GetPlugin(ch.Type)
	if err != nil {
		status := &channel.HealthStatus{
			ChannelID:     channelID,
			Status:        channel.HealthStatusUnhealthy,
			Message:       "Plugin not available",
			LastCheck:     time.Now(),
			ResponseTime:  time.Since(start),
			LastError:     err.Error(),
			LastErrorTime: &start,
		}
		hm.updateStatus(channelID, status)
		return status, nil
	}

	// 执行健康检查
	status := &channel.HealthStatus{
		ChannelID:    channelID,
		Status:       channel.HealthStatusTesting,
		Message:      "Checking...",
		LastCheck:    time.Now(),
		ResponseTime: 0,
	}

	if err := plugin.HealthCheck(ctx); err != nil {
		status.Status = channel.HealthStatusUnhealthy
		status.Message = "Health check failed"
		status.LastError = err.Error()
		status.LastErrorTime = &start
		status.ResponseTime = time.Since(start)
		
		// 增加错误计数
		if existingStatus := hm.getStatus(channelID); existingStatus != nil {
			status.ErrorCount = existingStatus.ErrorCount + 1
		} else {
			status.ErrorCount = 1
		}
	} else {
		status.Status = channel.HealthStatusHealthy
		status.Message = "Healthy"
		status.ResponseTime = time.Since(start)
		status.ErrorCount = 0
	}

	hm.updateStatus(channelID, status)
	return status, nil
}

// BatchCheck 批量健康检查
func (hm *HealthMonitor) BatchCheck(ctx context.Context, channelIDs []string, manager channel.ChannelManager) (map[string]*channel.HealthStatus, error) {
	results := make(map[string]*channel.HealthStatus)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, channelID := range channelIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			status, err := hm.CheckChannel(ctx, id, manager)
			if err != nil {
				status = &channel.HealthStatus{
					ChannelID:     id,
					Status:        channel.HealthStatusUnknown,
					Message:       "Check failed",
					LastCheck:     time.Now(),
					LastError:     err.Error(),
					LastErrorTime: func() *time.Time { t := time.Now(); return &t }(),
				}
			}
			mutex.Lock()
			results[id] = status
			mutex.Unlock()
		}(channelID)
	}

	wg.Wait()
	return results, nil
}

// Start 启动健康监控
func (hm *HealthMonitor) Start(ctx context.Context, interval time.Duration, manager channel.ChannelManager) {
	hm.running = true
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	hm.logger.Info("Health monitor started", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			hm.logger.Info("Health monitor stopped due to context cancellation")
			return
		case <-hm.stopChan:
			hm.logger.Info("Health monitor stopped")
			return
		case <-ticker.C:
			hm.performHealthCheck(ctx, manager)
		}
	}
}

// Stop 停止健康监控
func (hm *HealthMonitor) Stop() {
	if hm.running {
		close(hm.stopChan)
		hm.running = false
	}
}

// GetStatus 获取渠道健康状态
func (hm *HealthMonitor) GetStatus(channelID string) *channel.HealthStatus {
	return hm.getStatus(channelID)
}

// GetAllStatuses 获取所有渠道健康状态
func (hm *HealthMonitor) GetAllStatuses() map[string]*channel.HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	result := make(map[string]*channel.HealthStatus)
	for id, status := range hm.statuses {
		result[id] = status
	}
	return result
}

// performHealthCheck 执行健康检查
func (hm *HealthMonitor) performHealthCheck(ctx context.Context, manager channel.ChannelManager) {
	// 获取所有活跃渠道
	channels, err := manager.GetActiveChannels(ctx)
	if err != nil {
		hm.logger.Error("Failed to get active channels for health check", zap.Error(err))
		return
	}

	// 并发检查所有渠道
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(channelID string) {
			defer wg.Done()
			_, err := hm.CheckChannel(ctx, channelID, manager)
			if err != nil {
				hm.logger.Warn("Health check failed for channel",
					zap.String("channel_id", channelID),
					zap.Error(err))
			}
		}(ch.ID)
	}

	wg.Wait()
	hm.logger.Debug("Health check completed", zap.Int("channels_checked", len(channels)))
}

// updateStatus 更新状态
func (hm *HealthMonitor) updateStatus(channelID string, status *channel.HealthStatus) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hm.statuses[channelID] = status
}

// getStatus 获取状态
func (hm *HealthMonitor) getStatus(channelID string) *channel.HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.statuses[channelID]
}