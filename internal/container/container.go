package container

import (
	"alert_agent/internal/api/v1"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/repository"
	"alert_agent/internal/service"
)

// Container 依赖注入容器
type Container struct {
	// Repositories
	RuleRepository              repository.RuleRepository
	RuleVersionRepository       repository.RuleVersionRepository
	RuleAuditLogRepository      repository.RuleAuditLogRepository
	RuleDistributionRepository  repository.RuleDistributionRepository

	// Services
	RuleService             service.RuleService
	RuleVersionService      service.RuleVersionService
	RuleDistributionService service.RuleDistributionService
	RuleValidator           service.RuleValidator
	QueueService            *service.QueueService

	// Queue components
	QueueManager *queue.RedisMessageQueue
	QueueMonitor *queue.QueueMonitor

	// APIs
	RuleAPI        *v1.RuleAPI
	RuleVersionAPI *v1.RuleVersionAPI
	QueueAPI       *v1.QueueAPI
}

// NewContainer 创建新的容器实例
func NewContainer() *Container {
	container := &Container{}
	container.initRepositories()
	container.initServices()
	container.initAPIs()
	return container
}

// initRepositories 初始化仓库层
func (c *Container) initRepositories() {
	c.RuleRepository = repository.NewRuleRepository(database.DB)
	c.RuleVersionRepository = repository.NewRuleVersionRepository(database.DB)
	c.RuleAuditLogRepository = repository.NewRuleAuditLogRepository(database.DB)
	c.RuleDistributionRepository = repository.NewRuleDistributionRepository(database.DB)
}

// initServices 初始化服务层
func (c *Container) initServices() {
	c.RuleValidator = service.NewRuleValidator()
	c.RuleVersionService = service.NewRuleVersionService(c.RuleRepository, c.RuleVersionRepository, c.RuleAuditLogRepository)
	c.RuleService = service.NewRuleService(c.RuleRepository, c.RuleValidator, c.RuleVersionService)
	c.RuleDistributionService = service.NewRuleDistributionService(c.RuleDistributionRepository, c.RuleRepository)
	
	// 初始化队列组件
	c.QueueManager = queue.NewRedisMessageQueue(redis.Client, "alert_agent")
	c.QueueMonitor = queue.NewQueueMonitor(c.QueueManager, redis.Client, "alert_agent")
	c.QueueService = service.NewQueueService(c.QueueManager, c.QueueMonitor, redis.Client, "alert_agent")
}

// initAPIs 初始化API层
func (c *Container) initAPIs() {
	c.RuleAPI = v1.NewRuleAPI(c.RuleService, c.RuleDistributionService)
	c.RuleVersionAPI = v1.NewRuleVersionAPI(c.RuleService, c.RuleVersionService)
	c.QueueAPI = v1.NewQueueAPI(c.QueueManager, c.QueueMonitor)
}