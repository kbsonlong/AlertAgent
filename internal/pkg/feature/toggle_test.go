package feature

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestToggleManager_IsEnabled(t *testing.T) {
	logger := zap.NewNop()
	tm := NewToggleManager(logger)
	
	ctx := context.Background()
	
	tests := []struct {
		name        string
		featureName FeatureName
		expected    bool
	}{
		{
			name:        "direct routing should be enabled by default",
			featureName: FeatureDirectRouting,
			expected:    true,
		},
		{
			name:        "basic convergence should be disabled by default",
			featureName: FeatureBasicConvergence,
			expected:    false,
		},
		{
			name:        "async analysis should be enabled by default",
			featureName: FeatureAsyncAnalysis,
			expected:    true,
		},
		{
			name:        "smart routing should be disabled by default",
			featureName: FeatureSmartRouting,
			expected:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.IsEnabled(ctx, tt.featureName)
			if result != tt.expected {
				t.Errorf("IsEnabled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestToggleManager_UpdateFeature(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	tm := NewToggleManagerWithRegistry(logger, registry)
	
	// 获取原始配置
	originalConfig, err := tm.GetFeature(FeatureBasicConvergence)
	if err != nil {
		t.Fatalf("Failed to get feature config: %v", err)
	}
	
	// 更新配置
	newConfig := *originalConfig
	newConfig.State = StateEnabled
	newConfig.UpdatedAt = time.Now()
	
	err = tm.UpdateFeature(FeatureBasicConvergence, &newConfig)
	if err != nil {
		t.Fatalf("Failed to update feature: %v", err)
	}
	
	// 验证更新
	updatedConfig, err := tm.GetFeature(FeatureBasicConvergence)
	if err != nil {
		t.Fatalf("Failed to get updated feature config: %v", err)
	}
	
	if updatedConfig.State != StateEnabled {
		t.Errorf("Feature state not updated: got %v, expected %v", updatedConfig.State, StateEnabled)
	}
	
	// 验证功能现在已启用
	ctx := context.Background()
	if !tm.IsEnabled(ctx, FeatureBasicConvergence) {
		t.Error("Feature should be enabled after update")
	}
}

func TestToggleManager_GetPhaseFeatures(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	tm := NewToggleManagerWithRegistry(logger, registry)
	
	phaseOneFeatures := tm.GetPhaseFeatures(PhaseOne)
	phaseTwoFeatures := tm.GetPhaseFeatures(PhaseTwo)
	
	// 验证第一阶段功能
	expectedPhaseOneFeatures := []FeatureName{
		FeatureDirectRouting,
		FeatureBasicConvergence,
		FeatureAsyncAnalysis,
	}
	
	for _, expectedFeature := range expectedPhaseOneFeatures {
		if _, exists := phaseOneFeatures[expectedFeature]; !exists {
			t.Errorf("Phase one should contain feature %s", expectedFeature)
		}
	}
	
	// 验证第二阶段功能
	expectedPhaseTwoFeatures := []FeatureName{
		FeatureSmartRouting,
		FeatureAIDecisionMaking,
	}
	
	for _, expectedFeature := range expectedPhaseTwoFeatures {
		if _, exists := phaseTwoFeatures[expectedFeature]; !exists {
			t.Errorf("Phase two should contain feature %s", expectedFeature)
		}
	}
}

func TestToggleManager_EnablePhase(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	tm := NewToggleManagerWithRegistry(logger, registry)
	ctx := context.Background()
	
	// 启用第二阶段
	err := tm.EnablePhase(ctx, PhaseTwo)
	if err != nil {
		t.Fatalf("Failed to enable phase two: %v", err)
	}
	
	// 验证第二阶段功能已启用（除了有AI成熟度要求的功能）
	phaseTwoFeatures := tm.GetPhaseFeatures(PhaseTwo)
	for featureName, config := range phaseTwoFeatures {
		// 跳过有AI成熟度要求的功能，因为它们需要满足成熟度要求才能启用
		if config.AIMaturityRequirement != nil {
			continue
		}
		
		if config.State != StateEnabled {
			t.Errorf("Feature %s should be enabled after enabling phase two", featureName)
		}
	}
}

func TestToggleManager_DependencyCheck(t *testing.T) {
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	tm := NewToggleManagerWithRegistry(logger, registry)
	ctx := context.Background()
	
	// 禁用直通路由
	directRoutingConfig, _ := tm.GetFeature(FeatureDirectRouting)
	newConfig := *directRoutingConfig
	newConfig.State = StateDisabled
	tm.UpdateFeature(FeatureDirectRouting, &newConfig)
	
	// 尝试启用依赖于直通路由的功能
	basicConvergenceConfig, _ := tm.GetFeature(FeatureBasicConvergence)
	newBasicConfig := *basicConvergenceConfig
	newBasicConfig.State = StateEnabled
	tm.UpdateFeature(FeatureBasicConvergence, &newBasicConfig)
	
	// 基础收敛应该不能启用，因为依赖的直通路由被禁用了
	if tm.IsEnabled(ctx, FeatureBasicConvergence) {
		t.Error("Basic convergence should not be enabled when its dependency is disabled")
	}
}