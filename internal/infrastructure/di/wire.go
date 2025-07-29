//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	analysisapp "alert_agent/internal/application/analysis"
	channelapp "alert_agent/internal/application/channel"
	clusterapp "alert_agent/internal/application/cluster"
	gatewayapp "alert_agent/internal/application/gateway"
	"alert_agent/internal/config"
	analysisdomain "alert_agent/internal/domain/analysis"
	channeldomain "alert_agent/internal/domain/channel"
	clusterdomain "alert_agent/internal/domain/cluster"
	gatewaydomain "alert_agent/internal/domain/gateway"
	"alert_agent/internal/infrastructure/database"
	"alert_agent/internal/infrastructure/repository"
	"alert_agent/internal/shared/logger"
)

// ProviderSet 全局依赖注入提供者集合
var ProviderSet = wire.NewSet(
	// 基础设施提供者
	ProvideConfig,
	ProvideLogger,
	ProvideDatabase,
	ProvideRedis,

	// 仓储提供者
	ProvideChannelRepository,
	ProvideClusterRepository,
	ProvideAnalysisRepository,
	ProvideGatewayRepository,

	// 应用服务提供者
	ProvideChannelService,
	ProvideClusterService,
	ProvideAnalysisService,
	ProvideGatewayService,

	// 接口绑定
	wire.Bind(new(channeldomain.ChannelRepository), new(*repository.ChannelRepositoryImpl)),
	wire.Bind(new(clusterdomain.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	wire.Bind(new(analysisdomain.AnalysisRepository), new(*repository.AnalysisRepositoryImpl)),
	wire.Bind(new(gatewaydomain.GatewayRepository), new(*repository.GatewayRepositoryImpl)),
)

// ProvideConfig 提供配置
func ProvideConfig() *config.Config {
	cfg := config.GetConfig()
	return &cfg
}

// ProvideLogger 提供日志器
func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	loggerConfig := logger.Config{
		Level:            "info",
		Format:           "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EnableCaller:     true,
		EnableStacktrace: false,
	}

	if err := logger.Init(loggerConfig); err != nil {
		return nil, err
	}

	return logger.GetLogger(), nil
}

// ProvideDatabase 提供数据库连接
func ProvideDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	return database.InitDB(cfg, logger)
}

// ProvideRedis 提供Redis连接
func ProvideRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	return database.InitRedis(cfg, logger)
}

// ProvideChannelRepository 提供渠道仓储
func ProvideChannelRepository(db *gorm.DB, logger *zap.Logger) channeldomain.ChannelRepository {
	return repository.NewChannelRepository(db, logger)
}

// ProvideClusterRepository 提供集群仓储
func ProvideClusterRepository(db *gorm.DB, logger *zap.Logger) clusterdomain.ClusterRepository {
	return repository.NewClusterRepository(db, logger)
}

// ProvideAnalysisRepository 提供分析仓储
func ProvideAnalysisRepository(db *gorm.DB, logger *zap.Logger) analysisdomain.AnalysisRepository {
	return repository.NewAnalysisRepository(db, logger)
}

// ProvideGatewayRepository 提供网关仓储
func ProvideGatewayRepository(db *gorm.DB, redis *redis.Client, logger *zap.Logger) gatewaydomain.GatewayRepository {
	return repository.NewGatewayRepository(db, redis, logger)
}

// ProvideChannelService 提供渠道服务
func ProvideChannelService(
	repo channeldomain.ChannelRepository,
	logger *zap.Logger,
) channelapp.ChannelService {
	return channelapp.NewChannelService(repo, logger)
}

// ProvideClusterService 提供集群服务
func ProvideClusterService(
	repo clusterdomain.ClusterRepository,
	logger *zap.Logger,
) clusterapp.ClusterService {
	return clusterapp.NewClusterService(repo, logger)
}

// ProvideAnalysisService 提供分析服务
func ProvideAnalysisService(
	repo analysisdomain.AnalysisRepository,
	logger *zap.Logger,
) analysisapp.AnalysisService {
	return analysisapp.NewAnalysisService(repo, logger)
}

// ProvideGatewayService 提供网关服务
func ProvideGatewayService(
	repo gatewaydomain.GatewayRepository,
	channelService channelapp.ChannelService,
	logger *zap.Logger,
) gatewayapp.GatewayService {
	return gatewayapp.NewGatewayService(repo, channelService, logger)
}

// InitializeApplication 初始化应用程序
func InitializeApplication(ctx context.Context) (*Application, func(), error) {
	panic(wire.Build(ProviderSet, NewApplication))
}

// Application 应用程序结构
type Application struct {
	Config         *config.Config
	Logger         *zap.Logger
	DB             *gorm.DB
	Redis          *redis.Client
	ChannelService channelapp.ChannelService
	ClusterService clusterapp.ClusterService
	AnalysisService analysisapp.AnalysisService
	GatewayService gatewayapp.GatewayService
}

// NewApplication 创建应用程序实例
func NewApplication(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	redis *redis.Client,
	channelService channelapp.ChannelService,
	clusterService clusterapp.ClusterService,
	analysisService analysisapp.AnalysisService,
	gatewayService gatewayapp.GatewayService,
) *Application {
	return &Application{
		Config:          cfg,
		Logger:          logger,
		DB:              db,
		Redis:           redis,
		ChannelService:  channelService,
		ClusterService:  clusterService,
		AnalysisService: analysisService,
		GatewayService:  gatewayService,
	}
}

// Cleanup 清理资源
func (app *Application) Cleanup() {
	if app.Logger != nil {
		app.Logger.Sync()
	}
}