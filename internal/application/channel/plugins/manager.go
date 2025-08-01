package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PluginManager 插件管理器，提供高级插件管理功能
type PluginManager struct {
	registry       *PluginRegistry
	healthChecker  *HealthChecker
	statsCollector *StatsCollector
	hotLoader      *HotLoader
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewPluginManager 创建插件管理器
func NewPluginManager(logger *zap.Logger) *PluginManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	registry := NewPluginRegistry(logger)
	
	return &PluginManager{
		registry:       registry,
		healthChecker:  NewHealthChecker(registry, logger),
		statsCollector: NewStatsCollector(registry, logger),
		hotLoader:      NewHotLoader(registry, logger),
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动插件管理器
func (pm *PluginManager) Start() error {
	pm.logger.Info("Starting plugin manager")

	// 启动健康检查器
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.healthChecker.Start(pm.ctx)
	}()

	// 启动统计收集器
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.statsCollector.Start(pm.ctx)
	}()

	// 启动热加载器
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.hotLoader.Start(pm.ctx)
	}()

	// 注册内置插件
	if err := pm.registerBuiltinPlugins(); err != nil {
		return fmt.Errorf("failed to register builtin plugins: %w", err)
	}

	pm.logger.Info("Plugin manager started successfully")
	return nil
}

// Stop 停止插件管理器
func (pm *PluginManager) Stop() error {
	pm.logger.Info("Stopping plugin manager")
	
	pm.cancel()
	pm.wg.Wait()
	
	pm.logger.Info("Plugin manager stopped")
	return nil
}

// GetRegistry 获取插件注册表
func (pm *PluginManager) GetRegistry() *PluginRegistry {
	return pm.registry
}

// GetHealthChecker 获取健康检查器
func (pm *PluginManager) GetHealthChecker() *HealthChecker {
	return pm.healthChecker
}

// GetStatsCollector 获取统计收集器
func (pm *PluginManager) GetStatsCollector() *StatsCollector {
	return pm.statsCollector
}

// GetHotLoader 获取热加载器
func (pm *PluginManager) GetHotLoader() *HotLoader {
	return pm.hotLoader
}

// registerBuiltinPlugins 注册内置插件
func (pm *PluginManager) registerBuiltinPlugins() error {
	// 注册钉钉插件
	if err := pm.registry.RegisterPlugin(NewDingTalkPlugin()); err != nil {
		pm.logger.Error("Failed to register DingTalk plugin", zap.Error(err))
	}

	// 注册企业微信插件
	if err := pm.registry.RegisterPlugin(NewWeChatWorkPlugin()); err != nil {
		pm.logger.Error("Failed to register WeChat Work plugin", zap.Error(err))
	}

	// 注册邮件插件
	if err := pm.registry.RegisterPlugin(NewEmailPlugin()); err != nil {
		pm.logger.Error("Failed to register Email plugin", zap.Error(err))
	}

	return nil
}

// HealthChecker 健康检查器
type HealthChecker struct {
	registry      *PluginRegistry
	logger        *zap.Logger
	checkInterval time.Duration
	healthStatus  map[string]*HealthStatus
	mutex         sync.RWMutex
}

// HealthStatus 健康状态
type HealthStatus struct {
	PluginName    string    `json:"plugin_name"`
	IsHealthy     bool      `json:"is_healthy"`
	LastCheck     time.Time `json:"last_check"`
	LastError     string    `json:"last_error,omitempty"`
	CheckCount    int64     `json:"check_count"`
	FailureCount  int64     `json:"failure_count"`
	SuccessRate   float64   `json:"success_rate"`
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(registry *PluginRegistry, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		registry:      registry,
		logger:        logger,
		checkInterval: 5 * time.Minute, // 默认5分钟检查一次
		healthStatus:  make(map[string]*HealthStatus),
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	// 启动时立即检查一次
	hc.checkAllPlugins(ctx)

	for {
		select {
		case <-ctx.Done():
			hc.logger.Info("Health checker stopped")
			return
		case <-ticker.C:
			hc.checkAllPlugins(ctx)
		}
	}
}

// checkAllPlugins 检查所有插件健康状态
func (hc *HealthChecker) checkAllPlugins(ctx context.Context) {
	plugins := hc.registry.GetAvailablePlugins()
	
	for _, pluginInfo := range plugins {
		if pluginInfo.Status == "active" {
			hc.checkPlugin(ctx, pluginInfo.Name)
		}
	}
}

// checkPlugin 检查单个插件健康状态
func (hc *HealthChecker) checkPlugin(ctx context.Context, pluginName string) {
	hc.mutex.Lock()
	status, exists := hc.healthStatus[pluginName]
	if !exists {
		status = &HealthStatus{
			PluginName: pluginName,
		}
		hc.healthStatus[pluginName] = status
	}
	hc.mutex.Unlock()

	status.CheckCount++
	status.LastCheck = time.Now()

	err := hc.registry.HealthCheckPlugin(ctx, pluginName)
	if err != nil {
		status.IsHealthy = false
		status.LastError = err.Error()
		status.FailureCount++
		
		hc.logger.Warn("Plugin health check failed",
			zap.String("plugin", pluginName),
			zap.Error(err))
	} else {
		status.IsHealthy = true
		status.LastError = ""
		
		hc.logger.Debug("Plugin health check passed",
			zap.String("plugin", pluginName))
	}

	// 计算成功率
	if status.CheckCount > 0 {
		successCount := status.CheckCount - status.FailureCount
		status.SuccessRate = float64(successCount) / float64(status.CheckCount) * 100
	}
}

// GetHealthStatus 获取插件健康状态
func (hc *HealthChecker) GetHealthStatus(pluginName string) (*HealthStatus, bool) {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	status, exists := hc.healthStatus[pluginName]
	if !exists {
		return nil, false
	}
	
	// 返回副本
	statusCopy := *status
	return &statusCopy, true
}

// GetAllHealthStatus 获取所有插件健康状态
func (hc *HealthChecker) GetAllHealthStatus() map[string]*HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	result := make(map[string]*HealthStatus)
	for name, status := range hc.healthStatus {
		statusCopy := *status
		result[name] = &statusCopy
	}
	
	return result
}

// StatsCollector 统计收集器
type StatsCollector struct {
	registry        *PluginRegistry
	logger          *zap.Logger
	collectInterval time.Duration
	usageStats      map[string]*UsageStats
	mutex           sync.RWMutex
}

// UsageStats 使用统计
type UsageStats struct {
	PluginName       string        `json:"plugin_name"`
	TotalRequests    int64         `json:"total_requests"`
	SuccessRequests  int64         `json:"success_requests"`
	FailureRequests  int64         `json:"failure_requests"`
	AvgResponseTime  time.Duration `json:"avg_response_time" swaggertype:"integer" example:"2000" format:"int64"`
	LastUsed         time.Time     `json:"last_used"`
	RequestsPerHour  float64       `json:"requests_per_hour"`
	SuccessRate      float64       `json:"success_rate"`
}

// NewStatsCollector 创建统计收集器
func NewStatsCollector(registry *PluginRegistry, logger *zap.Logger) *StatsCollector {
	return &StatsCollector{
		registry:        registry,
		logger:          logger,
		collectInterval: 1 * time.Minute, // 每分钟收集一次统计
		usageStats:      make(map[string]*UsageStats),
	}
}

// Start 启动统计收集
func (sc *StatsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(sc.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sc.logger.Info("Stats collector stopped")
			return
		case <-ticker.C:
			sc.collectStats()
		}
	}
}

// collectStats 收集统计信息
func (sc *StatsCollector) collectStats() {
	allStats := sc.registry.GetAllPluginStats()
	
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	for pluginName, pluginStats := range allStats {
		usage, exists := sc.usageStats[pluginName]
		if !exists {
			usage = &UsageStats{
				PluginName: pluginName,
			}
			sc.usageStats[pluginName] = usage
		}
		
		// 更新统计信息
		usage.TotalRequests = pluginStats.TotalSent
		usage.SuccessRequests = pluginStats.SuccessCount
		usage.FailureRequests = pluginStats.FailureCount
		usage.AvgResponseTime = pluginStats.AvgDuration
		usage.LastUsed = pluginStats.LastSent
		
		// 计算成功率
		if usage.TotalRequests > 0 {
			usage.SuccessRate = float64(usage.SuccessRequests) / float64(usage.TotalRequests) * 100
		}
		
		// 计算每小时请求数（简化计算）
		if !usage.LastUsed.IsZero() {
			hoursSinceLastUse := time.Since(usage.LastUsed).Hours()
			if hoursSinceLastUse > 0 {
				usage.RequestsPerHour = float64(usage.TotalRequests) / hoursSinceLastUse
			}
		}
	}
}

// GetUsageStats 获取使用统计
func (sc *StatsCollector) GetUsageStats(pluginName string) (*UsageStats, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	stats, exists := sc.usageStats[pluginName]
	if !exists {
		return nil, false
	}
	
	// 返回副本
	statsCopy := *stats
	return &statsCopy, true
}

// GetAllUsageStats 获取所有使用统计
func (sc *StatsCollector) GetAllUsageStats() map[string]*UsageStats {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	result := make(map[string]*UsageStats)
	for name, stats := range sc.usageStats {
		statsCopy := *stats
		result[name] = &statsCopy
	}
	
	return result
}

// HotLoader 热加载器
type HotLoader struct {
	registry     *PluginRegistry
	logger       *zap.Logger
	watchInterval time.Duration
	pluginPaths  []string
}

// NewHotLoader 创建热加载器
func NewHotLoader(registry *PluginRegistry, logger *zap.Logger) *HotLoader {
	return &HotLoader{
		registry:      registry,
		logger:        logger,
		watchInterval: 30 * time.Second, // 30秒检查一次
		pluginPaths:   []string{}, // TODO: 从配置文件读取插件路径
	}
}

// Start 启动热加载
func (hl *HotLoader) Start(ctx context.Context) {
	if len(hl.pluginPaths) == 0 {
		hl.logger.Info("No plugin paths configured, hot loading disabled")
		return
	}

	ticker := time.NewTicker(hl.watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hl.logger.Info("Hot loader stopped")
			return
		case <-ticker.C:
			hl.scanPlugins()
		}
	}
}

// scanPlugins 扫描插件目录
func (hl *HotLoader) scanPlugins() {
	// TODO: 实现插件目录扫描和热加载逻辑
	// 这里需要实现：
	// 1. 扫描插件目录
	// 2. 检测新插件或插件更新
	// 3. 动态加载/卸载插件
	// 4. 处理插件依赖关系
	
	hl.logger.Debug("Scanning plugins for hot loading")
}

// LoadPlugin 动态加载插件
func (hl *HotLoader) LoadPlugin(pluginPath string) error {
	// TODO: 实现动态插件加载
	return fmt.Errorf("dynamic plugin loading not implemented yet")
}

// UnloadPlugin 动态卸载插件
func (hl *HotLoader) UnloadPlugin(pluginName string) error {
	return hl.registry.UnregisterPlugin(pluginName)
}

// ReloadPlugin 重新加载插件
func (hl *HotLoader) ReloadPlugin(pluginName string) error {
	// TODO: 实现插件重新加载
	return fmt.Errorf("plugin reloading not implemented yet")
}