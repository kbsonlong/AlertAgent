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
	PermissionRepository        repository.PermissionRepository
	RoleRepository              repository.RoleRepository
	UserRepository              repository.UserRepository

	// Services
	RuleService             service.RuleService
	RuleVersionService      service.RuleVersionService
	RuleDistributionService service.RuleDistributionService
	RuleValidator           service.RuleValidator
	QueueService            *service.QueueService
	PermissionService       service.PermissionService
	RoleService             service.RoleService
	UserService             service.UserService

	// Queue components
	QueueManager *queue.RedisMessageQueue
	QueueMonitor *queue.QueueMonitor

	// APIs
	RuleAPI             *v1.RuleAPI
	RuleVersionAPI      *v1.RuleVersionAPI
	QueueAPI            *v1.QueueAPI
	PermissionController *v1.PermissionController
	RoleController       *v1.RoleController
	UserController       *v1.UserController
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
	c.PermissionRepository = repository.NewPermissionRepository()
	c.RoleRepository = repository.NewRoleRepository()
	c.UserRepository = repository.NewUserRepository()
}

// initServices 初始化服务层
func (c *Container) initServices() {
	c.RuleValidator = service.NewRuleValidator()
	c.RuleVersionService = service.NewRuleVersionService(c.RuleRepository, c.RuleVersionRepository, c.RuleAuditLogRepository)
	c.RuleService = service.NewRuleService(c.RuleRepository, c.RuleValidator, c.RuleVersionService)
	c.RuleDistributionService = service.NewRuleDistributionService(c.RuleDistributionRepository, c.RuleRepository)
	c.PermissionService = service.NewPermissionService(c.PermissionRepository, c.RoleRepository)
	c.RoleService = service.NewRoleService(c.RoleRepository, c.PermissionRepository)
	c.UserService = service.NewUserService(c.UserRepository, c.RoleRepository)
	
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
	c.PermissionController = v1.NewPermissionController(c.PermissionService)
	c.RoleController = v1.NewRoleController(c.RoleService)
	c.UserController = v1.NewUserController(c.UserService)
	
	// 初始化用户控制器
	v1.InitUserController(c.UserController)
}