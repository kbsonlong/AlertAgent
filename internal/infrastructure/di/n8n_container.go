package di

import (
	"time"

	"alert_agent/internal/application/analysis"
	"alert_agent/internal/domain/alert"
	domainAnalysis "alert_agent/internal/domain/analysis"
	alertInfra "alert_agent/internal/infrastructure/alert"
	"alert_agent/internal/infrastructure/n8n"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// N8NContainer n8n 依赖注入容器
type N8NContainer struct {
	db     *gorm.DB
	logger *zap.Logger

	// Repositories
	alertRepo           alert.AlertRepository
	executionRepo       domainAnalysis.N8NWorkflowExecutionRepository

	// Services
	n8nClient           domainAnalysis.N8NClient
	workflowManager     domainAnalysis.N8NWorkflowManager
	n8nAnalysisService  *analysis.N8NAnalysisService
}

// NewN8NContainer 创建 n8n 依赖注入容器
func NewN8NContainer(db *gorm.DB, logger *zap.Logger) *N8NContainer {
	return &N8NContainer{
		db:     db,
		logger: logger,
	}
}

// InitializeRepositories 初始化仓储层
func (c *N8NContainer) InitializeRepositories() {
	// 初始化告警仓储
	c.alertRepo = alertInfra.NewGORMAlertRepository(c.db)

	// 初始化执行仓储
	c.executionRepo = n8n.NewGORMExecutionRepository(c.db)
}

// InitializeServices 初始化服务层
func (c *N8NContainer) InitializeServices(baseURL, apiKey string) {
	// 创建 HTTP 客户端配置
	httpConfig := &n8n.HTTPClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	// 初始化 n8n 客户端
	c.n8nClient = n8n.NewHTTPClient(httpConfig, c.logger)

	// 创建工作流管理器配置
	workflowConfig := &n8n.WorkflowManagerConfig{
		MonitorInterval:   5 * time.Second,
		MaxRetryAttempts:  3,
		RetryDelay:        10 * time.Second,
		ExecutionTimeout:  300 * time.Second,
		CallbackTimeout:   30 * time.Second,
		MaxConcurrentJobs: 10,
		CleanupInterval:   60 * time.Second,
		RetentionPeriod:   24 * time.Hour,
	}

	// 初始化工作流管理器
	c.workflowManager = n8n.NewWorkflowManager(
		workflowConfig,
		c.n8nClient,
		c.executionRepo,
		c.alertRepo,
		c.logger,
	)

	// 初始化 n8n 分析服务
	c.n8nAnalysisService = analysis.NewN8NAnalysisService(
		c.workflowManager,
		c.alertRepo,
		c.executionRepo,
	)
}

// GetAlertRepository 获取告警仓储
func (c *N8NContainer) GetAlertRepository() alert.AlertRepository {
	return c.alertRepo
}

// GetExecutionRepository 获取执行仓储
func (c *N8NContainer) GetExecutionRepository() domainAnalysis.N8NWorkflowExecutionRepository {
	return c.executionRepo
}

// GetN8NClient 获取 n8n 客户端
func (c *N8NContainer) GetN8NClient() domainAnalysis.N8NClient {
	return c.n8nClient
}

// GetWorkflowManager 获取工作流管理器
func (c *N8NContainer) GetWorkflowManager() domainAnalysis.N8NWorkflowManager {
	return c.workflowManager
}

// GetN8NAnalysisService 获取 n8n 分析服务
func (c *N8NContainer) GetN8NAnalysisService() *analysis.N8NAnalysisService {
	return c.n8nAnalysisService
}

// Initialize 初始化所有组件
func (c *N8NContainer) Initialize(baseURL, apiKey string) {
	c.InitializeRepositories()
	c.InitializeServices(baseURL, apiKey)
}

// Cleanup 清理资源
func (c *N8NContainer) Cleanup() {
	if c.n8nAnalysisService != nil {
		// 停止自动分析
		// 注意：这里需要实现停止机制
		c.logger.Info("Stopping n8n analysis service")
	}

	c.logger.Info("N8N container cleanup completed")
}