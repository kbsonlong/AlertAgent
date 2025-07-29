package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/pkg/feature"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// FeatureService 功能开关服务
type FeatureService struct {
	toggleManager *feature.ToggleManager
	monitor       *feature.FeatureMonitor
	evaluator     *feature.AIMaturityEvaluator
	logger        *zap.Logger
	configPath    string
	watcher       *fsnotify.Watcher
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// FeatureConfigFile 功能配置文件结构
type FeatureConfigFile struct {
	PhaseOneFeatures map[string]*feature.FeatureConfig `yaml:"phase_one_features"`
	PhaseTwoFeatures map[string]*feature.FeatureConfig `yaml:"phase_two_features"`
	GlobalSettings   GlobalSettings                    `yaml:"global_settings"`
}

// GlobalSettings 全局设置
type GlobalSettings struct {
	PhaseTransition    PhaseTransitionConfig    `yaml:"phase_transition"`
	DegradationPolicy  DegradationPolicyConfig  `yaml:"degradation_policy"`
	Monitoring         MonitoringConfig         `yaml:"monitoring"`
	Security           SecurityConfig           `yaml:"security"`
	Experimentation    ExperimentationConfig    `yaml:"experimentation"`
}

// PhaseTransitionConfig 阶段转换配置
type PhaseTransitionConfig struct {
	AutoPromotion     bool                           `yaml:"auto_promotion"`
	PromotionCriteria map[string]PromotionCriteria   `yaml:"promotion_criteria"`
}

// PromotionCriteria 晋升标准
type PromotionCriteria struct {
	MinUptimeDays           int     `yaml:"min_uptime_days"`
	MaxErrorRate            float64 `yaml:"max_error_rate"`
	MinAIMaturityScore      float64 `yaml:"min_ai_maturity_score"`
	UserSatisfactionScore   float64 `yaml:"user_satisfaction_score"`
}

// DegradationPolicyConfig 降级策略配置
type DegradationPolicyConfig struct {
	AutoDegradation      bool                    `yaml:"auto_degradation"`
	DegradationTriggers  DegradationTriggers     `yaml:"degradation_triggers"`
	RecoveryCriteria     RecoveryCriteria        `yaml:"recovery_criteria"`
}

// DegradationTriggers 降级触发器
type DegradationTriggers struct {
	ErrorRateThreshold    float64 `yaml:"error_rate_threshold"`
	LatencyThreshold      int     `yaml:"latency_threshold"`
	AIAccuracyDrop        float64 `yaml:"ai_accuracy_drop"`
	ConsecutiveFailures   int     `yaml:"consecutive_failures"`
}

// RecoveryCriteria 恢复标准
type RecoveryCriteria struct {
	StabilizationPeriod    int     `yaml:"stabilization_period"`
	SuccessRateThreshold   float64 `yaml:"success_rate_threshold"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	MetricsCollectionInterval int               `yaml:"metrics_collection_interval"`
	AlertEvaluationInterval   int               `yaml:"alert_evaluation_interval"`
	RetentionPolicy          RetentionPolicy    `yaml:"retention_policy"`
}

// RetentionPolicy 保留策略
type RetentionPolicy struct {
	Metrics string `yaml:"metrics"`
	Logs    string `yaml:"logs"`
	Alerts  string `yaml:"alerts"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	FeatureModificationRequiresApproval bool     `yaml:"feature_modification_requires_approval"`
	AuditLogging                       bool     `yaml:"audit_logging"`
	SensitiveFeatures                  []string `yaml:"sensitive_features"`
}

// ExperimentationConfig 实验配置
type ExperimentationConfig struct {
	ABTestingEnabled        bool `yaml:"a_b_testing_enabled"`
	CanaryDeploymentEnabled bool `yaml:"canary_deployment_enabled"`
	FeatureFlagsUIEnabled   bool `yaml:"feature_flags_ui_enabled"`
}

// NewFeatureService 创建功能开关服务
func NewFeatureService(logger *zap.Logger) (*FeatureService, error) {
	cfg := config.GetConfig()
	
	if !cfg.Features.Enabled {
		logger.Info("Feature toggle system is disabled")
		return nil, nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &FeatureService{
		toggleManager: feature.NewToggleManager(logger),
		monitor:       feature.NewFeatureMonitor(logger),
		evaluator:     feature.NewAIMaturityEvaluator(logger),
		logger:        logger,
		configPath:    cfg.Features.ConfigPath,
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// 加载功能配置
	if err := service.loadFeatureConfig(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to load feature config: %w", err)
	}
	
	// 启动配置文件监听
	if err := service.startConfigWatcher(); err != nil {
		logger.Warn("Failed to start feature config watcher", zap.Error(err))
	}
	
	// 注册回调函数
	service.registerCallbacks()
	
	// 启动监控
	if cfg.Features.MonitoringEnabled {
		go func() {
			if err := service.monitor.StartMonitoring(ctx); err != nil && err != context.Canceled {
				logger.Error("Feature monitoring failed", zap.Error(err))
			}
		}()
	}
	
	logger.Info("Feature service initialized",
		zap.String("config_path", service.configPath),
		zap.Bool("monitoring_enabled", cfg.Features.MonitoringEnabled),
		zap.Bool("ai_maturity_enabled", cfg.Features.AIMaturityEnabled))
	
	return service, nil
}

// loadFeatureConfig 加载功能配置
func (fs *FeatureService) loadFeatureConfig() error {
	// 获取配置文件的绝对路径
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	configPath := filepath.Join(workDir, fs.configPath)
	
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read feature config file: %w", err)
	}
	
	// 解析配置
	var configFile FeatureConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return fmt.Errorf("failed to parse feature config: %w", err)
	}
	
	// 更新功能配置
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	// 加载第一阶段功能
	for _, featureConfig := range configFile.PhaseOneFeatures {
		if err := fs.toggleManager.UpdateFeature(featureConfig.Name, featureConfig); err != nil {
			fs.logger.Error("Failed to update phase one feature",
				zap.String("feature", string(featureConfig.Name)),
				zap.Error(err))
		}
	}
	
	// 加载第二阶段功能
	for _, featureConfig := range configFile.PhaseTwoFeatures {
		if err := fs.toggleManager.UpdateFeature(featureConfig.Name, featureConfig); err != nil {
			fs.logger.Error("Failed to update phase two feature",
				zap.String("feature", string(featureConfig.Name)),
				zap.Error(err))
		}
	}
	
	fs.logger.Info("Feature configuration loaded successfully",
		zap.Int("phase_one_features", len(configFile.PhaseOneFeatures)),
		zap.Int("phase_two_features", len(configFile.PhaseTwoFeatures)))
	
	return nil
}

// startConfigWatcher 启动配置文件监听
func (fs *FeatureService) startConfigWatcher() error {
	var err error
	fs.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	
	// 获取配置文件的绝对路径
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	configPath := filepath.Join(workDir, fs.configPath)
	
	// 添加配置文件到监听列表
	if err := fs.watcher.Add(configPath); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}
	
	// 启动监听协程
	go func() {
		defer fs.watcher.Close()
		
		var lastReload time.Time
		const debounceInterval = 2 * time.Second
		
		for {
			select {
			case <-fs.ctx.Done():
				return
			case event, ok := <-fs.watcher.Events:
				if !ok {
					return
				}
				
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// 防抖动检查
					if time.Since(lastReload) < debounceInterval {
						continue
					}
					
					fs.logger.Info("Feature config file changed", zap.String("file", event.Name))
					
					// 重新加载配置
					if err := fs.loadFeatureConfig(); err != nil {
						fs.logger.Error("Failed to reload feature config", zap.Error(err))
					} else {
						lastReload = time.Now()
						fs.logger.Info("Feature configuration reloaded successfully")
					}
				}
				
			case err, ok := <-fs.watcher.Errors:
				if !ok {
					return
				}
				fs.logger.Error("Feature config watcher error", zap.Error(err))
			}
		}
	}()
	
	fs.logger.Info("Started watching feature config file", zap.String("file", configPath))
	return nil
}

// registerCallbacks 注册回调函数
func (fs *FeatureService) registerCallbacks() {
	// 注册AI成熟度降级回调
	allFeatures := fs.toggleManager.ListFeatures()
	for featureName := range allFeatures {
		fs.evaluator.RegisterDegradeCallback(featureName, func(assessment feature.MaturityAssessment) {
			fs.logger.Warn("AI maturity degradation detected",
				zap.String("feature", string(assessment.FeatureName)),
				zap.String("level", string(assessment.Level)),
				zap.Float64("score", assessment.Score))
			
			// 自动降级功能
			cfg := config.GetConfig()
			if cfg.Features.AutoDegradationEnabled {
				if err := fs.degradeFeature(assessment.FeatureName, "ai_maturity_low"); err != nil {
					fs.logger.Error("Failed to auto-degrade feature",
						zap.String("feature", string(assessment.FeatureName)),
						zap.Error(err))
				}
			}
		})
	}
	
	// 注册功能状态变更回调
	for featureName := range allFeatures {
		fs.toggleManager.RegisterCallback(featureName, func(config feature.FeatureConfig) {
			fs.logger.Info("Feature configuration changed",
				zap.String("feature", string(config.Name)),
				zap.String("state", string(config.State)))
			
			// 记录监控事件
			fs.monitor.RecordFeatureChange(config.Name, feature.StateDisabled, config.State)
		})
	}
}

// IsEnabled 检查功能是否启用
func (fs *FeatureService) IsEnabled(ctx context.Context, featureName feature.FeatureName, userContext ...map[string]interface{}) bool {
	if fs == nil {
		return false // 功能开关系统未启用
	}
	
	enabled := fs.toggleManager.IsEnabled(ctx, featureName, userContext...)
	
	// 记录使用情况
	result := "enabled"
	if !enabled {
		result = "disabled"
	}
	
	featureConfig, _ := fs.toggleManager.GetFeature(featureName)
	phase := feature.PhaseOne
	if featureConfig != nil {
		phase = featureConfig.Phase
	}
	
	fs.monitor.RecordFeatureUsage(featureName, phase, result)
	
	return enabled
}

// UpdateFeature 更新功能配置
func (fs *FeatureService) UpdateFeature(featureName feature.FeatureName, config *feature.FeatureConfig) error {
	if fs == nil {
		return fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.UpdateFeature(featureName, config)
}

// GetFeature 获取功能配置
func (fs *FeatureService) GetFeature(featureName feature.FeatureName) (*feature.FeatureConfig, error) {
	if fs == nil {
		return nil, fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.GetFeature(featureName)
}

// ListFeatures 列出所有功能
func (fs *FeatureService) ListFeatures() map[feature.FeatureName]*feature.FeatureConfig {
	if fs == nil {
		return make(map[feature.FeatureName]*feature.FeatureConfig)
	}
	
	return fs.toggleManager.ListFeatures()
}

// GetPhaseFeatures 获取指定阶段的功能
func (fs *FeatureService) GetPhaseFeatures(phase feature.Phase) map[feature.FeatureName]*feature.FeatureConfig {
	if fs == nil {
		return make(map[feature.FeatureName]*feature.FeatureConfig)
	}
	
	return fs.toggleManager.GetPhaseFeatures(phase)
}

// EnablePhase 启用指定阶段
func (fs *FeatureService) EnablePhase(ctx context.Context, phase feature.Phase) error {
	if fs == nil {
		return fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.EnablePhase(ctx, phase)
}

// DisablePhase 禁用指定阶段
func (fs *FeatureService) DisablePhase(ctx context.Context, phase feature.Phase) error {
	if fs == nil {
		return fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.DisablePhase(ctx, phase)
}

// RecordAIMetrics 记录AI指标
func (fs *FeatureService) RecordAIMetrics(featureName feature.FeatureName, metrics feature.AIMetrics) {
	if fs == nil {
		return
	}
	
	fs.evaluator.RecordMetrics(featureName, metrics)
	fs.monitor.RecordAIMaturityScore(featureName, metrics)
}

// GetAIMaturityAssessment 获取AI成熟度评估
func (fs *FeatureService) GetAIMaturityAssessment(featureName feature.FeatureName) (*feature.MaturityAssessment, error) {
	if fs == nil {
		return nil, fmt.Errorf("feature service not initialized")
	}
	
	return fs.evaluator.GetAssessment(featureName)
}

// degradeFeature 降级功能
func (fs *FeatureService) degradeFeature(featureName feature.FeatureName, reason string) error {
	featureConfig, err := fs.toggleManager.GetFeature(featureName)
	if err != nil {
		return fmt.Errorf("failed to get feature config: %w", err)
	}
	
	// 如果功能已经禁用，无需降级
	if featureConfig.State == feature.StateDisabled {
		return nil
	}
	
	// 降级到禁用状态
	newConfig := *featureConfig
	newConfig.State = feature.StateDisabled
	newConfig.UpdatedAt = time.Now()
	
	if err := fs.toggleManager.UpdateFeature(featureName, &newConfig); err != nil {
		return fmt.Errorf("failed to degrade feature: %w", err)
	}
	
	// 记录降级事件
	fs.monitor.RecordFeatureDegradation(featureName, reason)
	
	fs.logger.Warn("Feature degraded",
		zap.String("feature", string(featureName)),
		zap.String("reason", reason))
	
	return nil
}

// GetMonitoringReport 获取监控报告
func (fs *FeatureService) GetMonitoringReport(featureName feature.FeatureName, duration time.Duration) (*feature.MonitoringReport, error) {
	if fs == nil {
		return nil, fmt.Errorf("feature service not initialized")
	}
	
	return fs.monitor.GenerateReport(featureName, duration)
}

// GetActiveAlerts 获取活跃告警
func (fs *FeatureService) GetActiveAlerts() map[string]*feature.Alert {
	if fs == nil {
		return make(map[string]*feature.Alert)
	}
	
	return fs.monitor.GetActiveAlerts()
}

// ExportConfig 导出功能配置
func (fs *FeatureService) ExportConfig() ([]byte, error) {
	if fs == nil {
		return nil, fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.ExportConfig()
}

// ImportConfig 导入功能配置
func (fs *FeatureService) ImportConfig(data []byte) error {
	if fs == nil {
		return fmt.Errorf("feature service not initialized")
	}
	
	return fs.toggleManager.ImportConfig(data)
}

// Shutdown 关闭服务
func (fs *FeatureService) Shutdown() {
	if fs == nil {
		return
	}
	
	fs.cancel()
	
	if fs.watcher != nil {
		fs.watcher.Close()
	}
	
	fs.logger.Info("Feature service shutdown completed")
}

// GetToggleManager 获取功能开关管理器（用于测试）
func (fs *FeatureService) GetToggleManager() *feature.ToggleManager {
	if fs == nil {
		return nil
	}
	return fs.toggleManager
}

// GetMonitor 获取功能监控器（用于测试）
func (fs *FeatureService) GetMonitor() *feature.FeatureMonitor {
	if fs == nil {
		return nil
	}
	return fs.monitor
}

// GetEvaluator 获取AI成熟度评估器（用于测试）
func (fs *FeatureService) GetEvaluator() *feature.AIMaturityEvaluator {
	if fs == nil {
		return nil
	}
	return fs.evaluator
}