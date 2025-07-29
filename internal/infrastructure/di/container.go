package di

import (
	"time"
	
	"alert_agent/internal/application/analysis"
	"alert_agent/internal/application/channel"
	"alert_agent/internal/application/cluster"
	"alert_agent/internal/infrastructure/alert"
	"alert_agent/internal/infrastructure/container"
	"alert_agent/internal/infrastructure/dify"
	"alert_agent/internal/infrastructure/repository"
	"alert_agent/internal/interfaces/http"

	analysisDomain "alert_agent/internal/domain/analysis"
	alertDomain "alert_agent/internal/domain/alert"
	channelDomain "alert_agent/internal/domain/channel"
	clusterDomain "alert_agent/internal/domain/cluster"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger

	// Repositories
	clusterRepo         clusterDomain.Repository
	channelRepo         channelDomain.Repository
	alertRepo           alertDomain.AlertRepository
	difyAnalysisRepo    analysisDomain.DifyAnalysisRepository

	// Services
	clusterService      clusterDomain.Service
	channelService      channelDomain.Service
	channelManager      channelDomain.ChannelManager
	analysisService     analysisDomain.AnalysisService
	difyAnalysisService analysisDomain.DifyAnalysisService

	// Dify Components
	difyClient analysisDomain.DifyClient
	difyConfig *analysis.DifyAnalysisConfig

	// Analysis Container
	analysisContainer *container.AnalysisContainer

	// HTTP Router
	router *http.Router
}

// NewContainer 创建依赖注入容器
func NewContainer(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *Container {
	c := &Container{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}

	c.initRepositories()
	c.initServices()
	c.initAnalysisContainer()
	c.initHTTPRouter()

	return c
}

// initRepositories 初始化仓储层
func (c *Container) initRepositories() {
	c.clusterRepo = repository.NewClusterRepository(c.db)
	c.channelRepo = repository.NewChannelRepository(c.db)
	c.alertRepo = alert.NewGORMAlertRepository(c.db)
	c.difyAnalysisRepo = repository.NewDifyAnalysisRepository(c.db, c.logger)
}

// initServices 初始化服务层
func (c *Container) initServices() {
	c.clusterService = cluster.NewClusterService(c.clusterRepo)
	c.channelService = channel.NewChannelService(c.channelRepo)
	c.channelManager = channel.NewDefaultChannelManager(c.channelRepo, c.channelService, c.logger)
	
	// 初始化 Dify 配置和客户端
	c.initDifyComponents()
	
	// 初始化 Dify 分析服务
	c.difyAnalysisService = analysis.NewDifyAnalysisService(
		c.difyClient,
		c.difyAnalysisRepo,
		c.alertRepo,
		c.logger,
		c.difyConfig,
	)
}

// initDifyComponents 初始化 Dify 组件
func (c *Container) initDifyComponents() {
	// 初始化 Dify 配置
	c.difyConfig = &analysis.DifyAnalysisConfig{
		// 这里应该从配置文件或环境变量加载
		DefaultTimeout:        30 * time.Second,
		MaxRetries:           3,
		RetryInterval:        5 * time.Second,
		DefaultAgentID:       "default-agent",
		DefaultWorkflowID:    "default-workflow",
		ConcurrencyLimit:     10,
		TaskCleanupInterval:  1 * time.Hour,
		TaskRetentionTime:    24 * time.Hour,
	}
	
	// 创建 Dify 客户端
	c.difyClient = dify.NewDifyClient(
		"http://dify:5001", // baseURL
		"your-api-key",     // apiKey
		c.logger,
	)
}

// initAnalysisContainer 初始化分析容器
func (c *Container) initAnalysisContainer() {
	c.analysisContainer = container.NewAnalysisContainer(c.db, c.redisClient)
	c.analysisService = c.analysisContainer.GetAnalysisService()
}

// initHTTPRouter 初始化HTTP路由
func (c *Container) initHTTPRouter() {
	// 注意：这里需要根据实际的 NewRouter 函数签名来调整参数
	// 暂时使用 nil 作为占位符，实际使用时需要传入正确的参数
	c.router = http.NewRouter(
		c.clusterService,
		c.channelService,
		c.channelManager,
		c.analysisService,
		nil, // n8nService - 需要根据实际情况传入
		nil, // workflowManager - 需要根据实际情况传入
		c.logger,
	)
}

// GetDB 获取数据库连接
func (c *Container) GetDB() *gorm.DB {
	return c.db
}

// GetLogger 获取日志器
func (c *Container) GetLogger() *zap.Logger {
	return c.logger
}

// GetClusterRepository 获取集群仓储
func (c *Container) GetClusterRepository() clusterDomain.Repository {
	return c.clusterRepo
}

// GetChannelRepository 获取通道仓储
func (c *Container) GetChannelRepository() channelDomain.Repository {
	return c.channelRepo
}

// GetClusterService 获取集群服务
func (c *Container) GetClusterService() clusterDomain.Service {
	return c.clusterService
}

// GetChannelService 获取通道服务
func (c *Container) GetChannelService() channelDomain.Service {
	return c.channelService
}

// GetHTTPRouter 获取HTTP路由器
func (c *Container) GetHTTPRouter() *http.Router {
	return c.router
}

// Close 关闭容器，释放资源
func (c *Container) Close() error {
	// 关闭数据库连接
	if sqlDB, err := c.db.DB(); err == nil {
		return sqlDB.Close()
	}
	return nil
}