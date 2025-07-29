package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Phase 实施阶段枚举
type Phase string

const (
	PhaseOne Phase = "phase_one" // 第一阶段：告警及时性优先
	PhaseTwo Phase = "phase_two" // 第二阶段：智能功能
)

// FeatureState 功能状态枚举
type FeatureState string

const (
	StateDisabled    FeatureState = "disabled"    // 禁用
	StateEnabled     FeatureState = "enabled"     // 启用
	StateCanaryTest  FeatureState = "canary_test" // 金丝雀测试
	StateGradualRoll FeatureState = "gradual_roll" // 渐进式推出
)

// FeatureName 功能名称枚举
type FeatureName string

const (
	// 第一阶段功能
	FeatureDirectRouting     FeatureName = "direct_routing"      // 直通路由
	FeatureBasicConvergence  FeatureName = "basic_convergence"   // 基础收敛
	FeatureAsyncAnalysis     FeatureName = "async_analysis"      // 异步分析
	FeatureChannelManagement FeatureName = "channel_management"  // 渠道管理
	FeatureClusterSync       FeatureName = "cluster_sync"        // 集群同步

	// 第二阶段功能
	FeatureSmartRouting      FeatureName = "smart_routing"       // 智能路由
	FeatureAdvancedConverge  FeatureName = "advanced_convergence" // 高级收敛
	FeatureAutoSuppression   FeatureName = "auto_suppression"    // 自动抑制
	FeatureAIDecisionMaking  FeatureName = "ai_decision_making"  // AI决策
	FeatureAutoRemediation   FeatureName = "auto_remediation"    // 自动修复
)

// FeatureConfig 功能配置
type FeatureConfig struct {
	Name        FeatureName  `json:"name" yaml:"name"`
	Phase       Phase        `json:"phase" yaml:"phase"`
	State       FeatureState `json:"state" yaml:"state"`
	Description string       `json:"description" yaml:"description"`
	
	// 渐进式推出配置
	RolloutConfig *RolloutConfig `json:"rollout_config,omitempty" yaml:"rollout_config,omitempty"`
	
	// AI模型成熟度要求
	AIMaturityRequirement *AIMaturityRequirement `json:"ai_maturity_requirement,omitempty" yaml:"ai_maturity_requirement,omitempty"`
	
	// 依赖的功能
	Dependencies []FeatureName `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	
	// 监控配置
	MonitoringConfig *MonitoringConfig `json:"monitoring_config,omitempty" yaml:"monitoring_config,omitempty"`
	
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

// RolloutConfig 渐进式推出配置
type RolloutConfig struct {
	Percentage    int               `json:"percentage" yaml:"percentage"`         // 推出百分比
	UserGroups    []string          `json:"user_groups" yaml:"user_groups"`       // 目标用户组
	Clusters      []string          `json:"clusters" yaml:"clusters"`             // 目标集群
	StartTime     *time.Time        `json:"start_time" yaml:"start_time"`         // 开始时间
	EndTime       *time.Time        `json:"end_time" yaml:"end_time"`             // 结束时间
	Conditions    map[string]string `json:"conditions" yaml:"conditions"`         // 启用条件
}

// AIMaturityRequirement AI模型成熟度要求
type AIMaturityRequirement struct {
	MinAccuracy     float64 `json:"min_accuracy" yaml:"min_accuracy"`         // 最小准确率
	MinConfidence   float64 `json:"min_confidence" yaml:"min_confidence"`     // 最小置信度
	MaxLatency      int     `json:"max_latency" yaml:"max_latency"`           // 最大延迟(ms)
	MinSuccessRate  float64 `json:"min_success_rate" yaml:"min_success_rate"` // 最小成功率
	EvaluationPeriod int    `json:"evaluation_period" yaml:"evaluation_period"` // 评估周期(小时)
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	MetricsEnabled    bool     `json:"metrics_enabled" yaml:"metrics_enabled"`
	AlertsEnabled     bool     `json:"alerts_enabled" yaml:"alerts_enabled"`
	LogLevel          string   `json:"log_level" yaml:"log_level"`
	SampleRate        float64  `json:"sample_rate" yaml:"sample_rate"`
	AlertThresholds   map[string]float64 `json:"alert_thresholds" yaml:"alert_thresholds"`
}

// ToggleManager 功能开关管理器
type ToggleManager struct {
	features    map[FeatureName]*FeatureConfig
	mutex       sync.RWMutex
	logger      *zap.Logger
	evaluator   *AIMaturityEvaluator
	monitor     *FeatureMonitor
	callbacks   map[FeatureName][]func(FeatureConfig)
}

// NewToggleManager 创建功能开关管理器
func NewToggleManager(logger *zap.Logger) *ToggleManager {
	return NewToggleManagerWithRegistry(logger, prometheus.DefaultRegisterer)
}

// NewToggleManagerWithRegistry 使用指定注册器创建功能开关管理器
func NewToggleManagerWithRegistry(logger *zap.Logger, registerer prometheus.Registerer) *ToggleManager {
	tm := &ToggleManager{
		features:  make(map[FeatureName]*FeatureConfig),
		logger:    logger,
		evaluator: NewAIMaturityEvaluator(logger),
		monitor:   NewFeatureMonitorWithRegistry(logger, registerer),
		callbacks: make(map[FeatureName][]func(FeatureConfig)),
	}
	
	// 初始化默认功能配置
	tm.initializeDefaultFeatures()
	
	return tm
}

// initializeDefaultFeatures 初始化默认功能配置
func (tm *ToggleManager) initializeDefaultFeatures() {
	defaultFeatures := []*FeatureConfig{
		// 第一阶段功能
		{
			Name:        FeatureDirectRouting,
			Phase:       PhaseOne,
			State:       StateEnabled,
			Description: "直通路由：告警直接通过用户定义渠道发送",
			MonitoringConfig: &MonitoringConfig{
				MetricsEnabled: true,
				AlertsEnabled:  true,
				LogLevel:      "info",
				SampleRate:    1.0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:        FeatureBasicConvergence,
			Phase:       PhaseOne,
			State:       StateDisabled, // 默认禁用，可选开关
			Description: "基础告警收敛功能",
			Dependencies: []FeatureName{FeatureDirectRouting},
			MonitoringConfig: &MonitoringConfig{
				MetricsEnabled: true,
				AlertsEnabled:  true,
				LogLevel:      "info",
				SampleRate:    0.1,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:        FeatureAsyncAnalysis,
			Phase:       PhaseOne,
			State:       StateEnabled,
			Description: "异步AI分析功能",
			Dependencies: []FeatureName{FeatureDirectRouting},
			MonitoringConfig: &MonitoringConfig{
				MetricsEnabled: true,
				AlertsEnabled:  true,
				LogLevel:      "info",
				SampleRate:    0.5,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		
		// 第二阶段功能
		{
			Name:        FeatureSmartRouting,
			Phase:       PhaseTwo,
			State:       StateDisabled,
			Description: "基于AI的智能路由决策",
			Dependencies: []FeatureName{FeatureDirectRouting, FeatureAsyncAnalysis},
			AIMaturityRequirement: &AIMaturityRequirement{
				MinAccuracy:      0.85,
				MinConfidence:    0.8,
				MaxLatency:       500,
				MinSuccessRate:   0.95,
				EvaluationPeriod: 24,
			},
			MonitoringConfig: &MonitoringConfig{
				MetricsEnabled: true,
				AlertsEnabled:  true,
				LogLevel:      "debug",
				SampleRate:    1.0,
				AlertThresholds: map[string]float64{
					"accuracy_drop":    0.1,
					"latency_increase": 0.2,
					"error_rate":       0.05,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:        FeatureAIDecisionMaking,
			Phase:       PhaseTwo,
			State:       StateDisabled,
			Description: "AI驱动的自动化决策",
			Dependencies: []FeatureName{FeatureSmartRouting, FeatureAsyncAnalysis},
			AIMaturityRequirement: &AIMaturityRequirement{
				MinAccuracy:      0.9,
				MinConfidence:    0.85,
				MaxLatency:       300,
				MinSuccessRate:   0.98,
				EvaluationPeriod: 48,
			},
			MonitoringConfig: &MonitoringConfig{
				MetricsEnabled: true,
				AlertsEnabled:  true,
				LogLevel:      "debug",
				SampleRate:    1.0,
				AlertThresholds: map[string]float64{
					"accuracy_drop":    0.05,
					"confidence_drop":  0.1,
					"error_rate":       0.02,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	for _, feature := range defaultFeatures {
		tm.features[feature.Name] = feature
	}
}

// IsEnabled 检查功能是否启用
func (tm *ToggleManager) IsEnabled(ctx context.Context, featureName FeatureName, userContext ...map[string]interface{}) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	feature, exists := tm.features[featureName]
	if !exists {
		tm.logger.Warn("Feature not found", zap.String("feature", string(featureName)))
		return false
	}
	
	// 检查功能状态
	if feature.State == StateDisabled {
		return false
	}
	
	// 检查依赖
	if !tm.checkDependencies(feature) {
		tm.logger.Debug("Feature dependencies not met", zap.String("feature", string(featureName)))
		return false
	}
	
	// 检查AI模型成熟度
	if feature.AIMaturityRequirement != nil {
		if !tm.evaluator.EvaluateMaturity(ctx, featureName, *feature.AIMaturityRequirement) {
			tm.logger.Info("AI maturity requirement not met, feature disabled",
				zap.String("feature", string(featureName)))
			return false
		}
	}
	
	// 检查渐进式推出条件
	if feature.RolloutConfig != nil {
		return tm.checkRolloutConditions(feature, userContext...)
	}
	
	return feature.State == StateEnabled
}

// checkDependencies 检查功能依赖
func (tm *ToggleManager) checkDependencies(feature *FeatureConfig) bool {
	for _, dep := range feature.Dependencies {
		depFeature, exists := tm.features[dep]
		if !exists || depFeature.State != StateEnabled {
			return false
		}
	}
	return true
}

// checkRolloutConditions 检查渐进式推出条件
func (tm *ToggleManager) checkRolloutConditions(feature *FeatureConfig, userContext ...map[string]interface{}) bool {
	rollout := feature.RolloutConfig
	
	// 检查时间窗口
	now := time.Now()
	if rollout.StartTime != nil && now.Before(*rollout.StartTime) {
		return false
	}
	if rollout.EndTime != nil && now.After(*rollout.EndTime) {
		return false
	}
	
	// 检查用户组（如果提供了用户上下文）
	if len(userContext) > 0 && len(rollout.UserGroups) > 0 {
		userGroup, exists := userContext[0]["user_group"]
		if !exists {
			return false
		}
		
		userGroupStr, ok := userGroup.(string)
		if !ok {
			return false
		}
		
		found := false
		for _, group := range rollout.UserGroups {
			if group == userGroupStr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// 检查集群（如果提供了集群上下文）
	if len(userContext) > 0 && len(rollout.Clusters) > 0 {
		cluster, exists := userContext[0]["cluster"]
		if !exists {
			return false
		}
		
		clusterStr, ok := cluster.(string)
		if !ok {
			return false
		}
		
		found := false
		for _, c := range rollout.Clusters {
			if c == clusterStr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// UpdateFeature 更新功能配置
func (tm *ToggleManager) UpdateFeature(featureName FeatureName, config *FeatureConfig) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	oldConfig := tm.features[featureName]
	config.UpdatedAt = time.Now()
	tm.features[featureName] = config
	
	tm.logger.Info("Feature configuration updated",
		zap.String("feature", string(featureName)),
		zap.String("old_state", string(oldConfig.State)),
		zap.String("new_state", string(config.State)))
	
	// 触发回调
	if callbacks, exists := tm.callbacks[featureName]; exists {
		for _, callback := range callbacks {
			go callback(*config)
		}
	}
	
	// 记录监控事件
	tm.monitor.RecordFeatureChange(featureName, oldConfig.State, config.State)
	
	return nil
}

// GetFeature 获取功能配置
func (tm *ToggleManager) GetFeature(featureName FeatureName) (*FeatureConfig, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	feature, exists := tm.features[featureName]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", featureName)
	}
	
	// 返回副本以避免并发修改
	configCopy := *feature
	return &configCopy, nil
}

// ListFeatures 列出所有功能
func (tm *ToggleManager) ListFeatures() map[FeatureName]*FeatureConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	result := make(map[FeatureName]*FeatureConfig)
	for name, config := range tm.features {
		configCopy := *config
		result[name] = &configCopy
	}
	
	return result
}

// RegisterCallback 注册功能状态变更回调
func (tm *ToggleManager) RegisterCallback(featureName FeatureName, callback func(FeatureConfig)) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	if tm.callbacks[featureName] == nil {
		tm.callbacks[featureName] = make([]func(FeatureConfig), 0)
	}
	
	tm.callbacks[featureName] = append(tm.callbacks[featureName], callback)
}

// GetPhaseFeatures 获取指定阶段的功能
func (tm *ToggleManager) GetPhaseFeatures(phase Phase) map[FeatureName]*FeatureConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	result := make(map[FeatureName]*FeatureConfig)
	for name, config := range tm.features {
		if config.Phase == phase {
			configCopy := *config
			result[name] = &configCopy
		}
	}
	
	return result
}

// EnablePhase 启用指定阶段的所有功能
func (tm *ToggleManager) EnablePhase(ctx context.Context, phase Phase) error {
	phaseFeatures := tm.GetPhaseFeatures(phase)
	
	for name, config := range phaseFeatures {
		if config.State == StateDisabled {
			newConfig := *config
			newConfig.State = StateEnabled
			if err := tm.UpdateFeature(name, &newConfig); err != nil {
				return fmt.Errorf("failed to enable feature %s: %w", name, err)
			}
		}
	}
	
	tm.logger.Info("Phase enabled", zap.String("phase", string(phase)))
	return nil
}

// DisablePhase 禁用指定阶段的所有功能
func (tm *ToggleManager) DisablePhase(ctx context.Context, phase Phase) error {
	phaseFeatures := tm.GetPhaseFeatures(phase)
	
	for name, config := range phaseFeatures {
		if config.State != StateDisabled {
			newConfig := *config
			newConfig.State = StateDisabled
			if err := tm.UpdateFeature(name, &newConfig); err != nil {
				return fmt.Errorf("failed to disable feature %s: %w", name, err)
			}
		}
	}
	
	tm.logger.Info("Phase disabled", zap.String("phase", string(phase)))
	return nil
}

// ExportConfig 导出功能配置
func (tm *ToggleManager) ExportConfig() ([]byte, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	return json.MarshalIndent(tm.features, "", "  ")
}

// ImportConfig 导入功能配置
func (tm *ToggleManager) ImportConfig(data []byte) error {
	var features map[FeatureName]*FeatureConfig
	if err := json.Unmarshal(data, &features); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.features = features
	tm.logger.Info("Feature configuration imported", zap.Int("count", len(features)))
	
	return nil
}