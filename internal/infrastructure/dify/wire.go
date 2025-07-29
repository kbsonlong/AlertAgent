//go:build wireinject
// +build wireinject

package dify

import (
	"context"
	"database/sql"
	
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
	
	"alert_agent/internal/application/analysis"
	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/infrastructure/repository"
)

// ProviderSet Dify相关依赖注入提供者集合
var ProviderSet = wire.NewSet(
	// 配置提供者
	ProvideConfig,
	
	// 客户端提供者
	ProvideDifyClient,
	
	// 仓储提供者
	ProvideDifyAnalysisRepository,
	
	// 服务提供者
	ProvideDifyAnalysisService,
	
	// 接口绑定
	wire.Bind(new(analysis.DifyClient), new(*DifyClientImpl)),
	wire.Bind(new(analysis.DifyAnalysisRepository), new(*repository.DifyAnalysisRepositoryImpl)),
	wire.Bind(new(analysis.DifyAnalysisService), new(*analysis.DifyAnalysisServiceImpl)),
)

// ProvideConfig 提供Dify配置
func ProvideConfig() *DifyConfig {
	// 在实际应用中，这里应该从配置文件或环境变量加载
	config := DefaultDifyConfig()
	
	// 示例配置
	config.BaseURL = "https://api.dify.ai"
	config.APIKey = "your-api-key"
	config.AppToken = "your-app-token"
	config.UserID = "alert-agent"
	
	// 工作流配置
	config.WorkflowConfig.DefaultWorkflowID = "default-analysis-workflow"
	config.WorkflowConfig.WorkflowMapping = map[string]string{
		"cpu_high":    "cpu-analysis-workflow",
		"memory_high": "memory-analysis-workflow",
		"disk_full":   "disk-analysis-workflow",
		"network":     "network-analysis-workflow",
	}
	
	// 知识库配置
	config.KnowledgeConfig.DefaultDatasetIDs = []string{"general-kb", "troubleshooting-kb"}
	config.KnowledgeConfig.DatasetMapping = map[string][]string{
		"cpu_high":    {"cpu-kb", "performance-kb"},
		"memory_high": {"memory-kb", "performance-kb"},
		"disk_full":   {"storage-kb", "capacity-kb"},
		"network":     {"network-kb", "connectivity-kb"},
	}
	
	return config
}

// ProvideDifyClient 提供Dify客户端
func ProvideDifyClient(config *DifyConfig, logger *zap.Logger) analysis.DifyClient {
	return NewDifyClient(config.BaseURL, config.APIKey, config.AppToken, config.UserID, logger)
}

// ProvideDifyAnalysisRepository 提供Dify分析仓储
func ProvideDifyAnalysisRepository(db *gorm.DB, logger *zap.Logger) analysis.DifyAnalysisRepository {
	return repository.NewDifyAnalysisRepository(db, logger)
}

// ProvideDifyAnalysisService 提供Dify分析服务
func ProvideDifyAnalysisService(
	client analysis.DifyClient,
	repo analysis.DifyAnalysisRepository,
	logger *zap.Logger,
) analysis.DifyAnalysisService {
	config := &analysis.DifyAnalysisConfig{
		MaxConcurrentTasks: 10,
		TaskQueueSize:      100,
		AnalysisTimeout:    600, // 10分钟
		ContextTimeout:     30,  // 30秒
		MaxRetries:         3,
		RetryInterval:      5, // 5秒
		ResultTTL:          86400, // 24小时
		EnableMetrics:      true,
		EnableTracing:      false,
	}
	
	return analysis.NewDifyAnalysisService(client, repo, config, logger)
}

// InitializeDifyAnalysisService 初始化Dify分析服务
func InitializeDifyAnalysisService(
	ctx context.Context,
	db *gorm.DB,
	logger *zap.Logger,
) (analysis.DifyAnalysisService, func(), error) {
	panic(wire.Build(ProviderSet))
}

// InitializeDifyClient 初始化Dify客户端
func InitializeDifyClient(
	logger *zap.Logger,
) (analysis.DifyClient, func(), error) {
	panic(wire.Build(
		ProvideConfig,
		ProvideDifyClient,
		wire.Bind(new(analysis.DifyClient), new(*DifyClientImpl)),
	))
}

// InitializeDifyAnalysisRepository 初始化Dify分析仓储
func InitializeDifyAnalysisRepository(
	db *gorm.DB,
	logger *zap.Logger,
) (analysis.DifyAnalysisRepository, func(), error) {
	panic(wire.Build(
		ProvideDifyAnalysisRepository,
		wire.Bind(new(analysis.DifyAnalysisRepository), new(*repository.DifyAnalysisRepositoryImpl)),
	))
}