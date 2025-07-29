package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"alert_agent/internal/pkg/feature"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// 创建功能开关服务
	featureService, err := service.NewFeatureService(logger)
	if err != nil {
		logger.Fatal("Failed to create feature service", zap.Error(err))
	}
	defer featureService.Shutdown()

	ctx := context.Background()

	// 示例1: 检查功能是否启用
	fmt.Println("=== 功能状态检查 ===")
	features := []feature.FeatureName{
		feature.FeatureDirectRouting,
		feature.FeatureBasicConvergence,
		feature.FeatureAsyncAnalysis,
		feature.FeatureSmartRouting,
		feature.FeatureAIDecisionMaking,
	}

	for _, featureName := range features {
		enabled := featureService.IsEnabled(ctx, featureName)
		fmt.Printf("功能 %s: %s\n", featureName, getStatusText(enabled))
	}

	// 示例2: 启用基础收敛功能
	fmt.Println("\n=== 启用基础收敛功能 ===")
	basicConvergenceConfig, _ := featureService.GetFeature(feature.FeatureBasicConvergence)
	if basicConvergenceConfig != nil {
		newConfig := *basicConvergenceConfig
		newConfig.State = feature.StateEnabled
		newConfig.UpdatedAt = time.Now()
		
		if err := featureService.UpdateFeature(feature.FeatureBasicConvergence, &newConfig); err != nil {
			logger.Error("Failed to enable basic convergence", zap.Error(err))
		} else {
			fmt.Printf("基础收敛功能已启用\n")
			enabled := featureService.IsEnabled(ctx, feature.FeatureBasicConvergence)
			fmt.Printf("验证状态: %s\n", getStatusText(enabled))
		}
	}

	// 示例3: 记录AI指标
	fmt.Println("\n=== 记录AI指标 ===")
	aiMetrics := feature.AIMetrics{
		Accuracy:    0.92,
		Confidence:  0.88,
		Latency:     250,
		SuccessRate: 0.96,
		ErrorRate:   0.04,
		SampleCount: 100,
	}
	
	featureService.RecordAIMetrics(feature.FeatureSmartRouting, aiMetrics)
	fmt.Printf("已记录智能路由AI指标: 准确率=%.2f, 置信度=%.2f, 延迟=%dms\n", 
		aiMetrics.Accuracy, aiMetrics.Confidence, aiMetrics.Latency)

	// 示例4: 获取阶段功能
	fmt.Println("\n=== 阶段功能列表 ===")
	phaseOneFeatures := featureService.GetPhaseFeatures(feature.PhaseOne)
	fmt.Printf("第一阶段功能 (%d个):\n", len(phaseOneFeatures))
	for name, config := range phaseOneFeatures {
		fmt.Printf("  - %s: %s (%s)\n", name, config.State, config.Description)
	}

	phaseTwoFeatures := featureService.GetPhaseFeatures(feature.PhaseTwo)
	fmt.Printf("\n第二阶段功能 (%d个):\n", len(phaseTwoFeatures))
	for name, config := range phaseTwoFeatures {
		fmt.Printf("  - %s: %s (%s)\n", name, config.State, config.Description)
	}

	// 示例5: 用户上下文检查
	fmt.Println("\n=== 用户上下文检查 ===")
	userContext := map[string]interface{}{
		"user_group": "beta_testers",
		"cluster":    "staging",
	}
	
	smartRoutingEnabled := featureService.IsEnabled(ctx, feature.FeatureSmartRouting, userContext)
	fmt.Printf("智能路由功能 (beta测试用户): %s\n", getStatusText(smartRoutingEnabled))

	// 示例6: 导出配置
	fmt.Println("\n=== 导出配置 ===")
	configData, err := featureService.ExportConfig()
	if err != nil {
		logger.Error("Failed to export config", zap.Error(err))
	} else {
		fmt.Printf("配置已导出 (%d 字节)\n", len(configData))
		// 可以保存到文件或发送到其他系统
	}

	fmt.Println("\n=== 功能开关演示完成 ===")
}

func getStatusText(enabled bool) string {
	if enabled {
		return "✅ 启用"
	}
	return "❌ 禁用"
}