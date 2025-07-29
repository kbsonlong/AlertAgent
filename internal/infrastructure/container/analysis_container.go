package container

import (
	"context"
	"fmt"
	"time"

	analysisApp "alert_agent/internal/application/analysis"
	analysisDomain "alert_agent/internal/domain/analysis"
	"alert_agent/internal/infrastructure/queue"
	"alert_agent/internal/infrastructure/repository"
	"alert_agent/internal/infrastructure/worker"
	"alert_agent/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AnalysisContainer 分析模块依赖注入容器
type AnalysisContainer struct {
	// 基础设施
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger

	// 仓库层
	taskRepo        analysisDomain.AnalysisTaskRepository
	resultRepo      analysisDomain.AnalysisResultRepository
	progressTracker analysisDomain.AnalysisProgressTracker

	// 队列
	taskQueue analysisDomain.AnalysisTaskQueue

	// 分析引擎
	analysisEngine analysisDomain.AnalysisEngine

	// 指标收集器
	metricsCollector analysisDomain.AnalysisMetricsCollector

	// 通知器
	notifier analysisDomain.AnalysisNotifier

	// 重试策略
	retryPolicy analysisDomain.AnalysisRetryPolicy

	// 工作器相关
	workerFactory  worker.WorkerFactory
	workerManager  analysisDomain.AnalysisWorkerManager

	// 应用服务
	analysisService analysisDomain.AnalysisService
}

// NewAnalysisContainer 创建分析容器
func NewAnalysisContainer(db *gorm.DB, redisClient *redis.Client) *AnalysisContainer {
	container := &AnalysisContainer{
		db:          db,
		redisClient: redisClient,
		logger:      logger.L.Named("analysis-container"),
	}

	// 初始化所有依赖
	container.initRepositories()
	container.initQueue()
	container.initEngine()
	container.initMetrics()
	container.initNotifier()
	container.initRetryPolicy()
	container.initWorkers()
	container.initServices()

	return container
}

// initRepositories 初始化仓库层
func (c *AnalysisContainer) initRepositories() {
	c.taskRepo = repository.NewAnalysisTaskRepository(c.db)
	c.resultRepo = repository.NewAnalysisResultRepository(c.db)
	c.progressTracker = repository.NewAnalysisProgressTracker(c.redisClient)
}

// initQueue 初始化队列
func (c *AnalysisContainer) initQueue() {
	c.taskQueue = queue.NewAnalysisTaskQueue(c.redisClient)
}

// initEngine 初始化分析引擎
func (c *AnalysisContainer) initEngine() {
	// TODO: 实现分析引擎
	// c.analysisEngine = engine.NewAnalysisEngine(...)
	c.analysisEngine = &MockAnalysisEngine{}
}

// initMetrics 初始化指标收集器
func (c *AnalysisContainer) initMetrics() {
	// TODO: 实现指标收集器
	// c.metricsCollector = metrics.NewAnalysisMetricsCollector(...)
	c.metricsCollector = &MockMetricsCollector{}
}

// initNotifier 初始化通知器
func (c *AnalysisContainer) initNotifier() {
	// TODO: 实现通知器
	// c.notifier = notifier.NewAnalysisNotifier(...)
	c.notifier = &MockNotifier{}
}

// initRetryPolicy 初始化重试策略
func (c *AnalysisContainer) initRetryPolicy() {
	// TODO: 实现重试策略
	// c.retryPolicy = retry.NewAnalysisRetryPolicy(...)
	c.retryPolicy = &MockRetryPolicy{}
}

// initWorkers 初始化工作器
func (c *AnalysisContainer) initWorkers() {
	c.workerFactory = worker.NewDefaultWorkerFactory(
		c.taskQueue,
		c.taskRepo,
		c.resultRepo,
		c.progressTracker,
		c.analysisEngine,
		c.metricsCollector,
	)
	c.workerManager = worker.NewAnalysisWorkerManager(c.workerFactory)
}

// initServices 初始化应用服务
func (c *AnalysisContainer) initServices() {
	c.analysisService = analysisApp.NewAnalysisService(
		c.taskQueue,
		c.taskRepo,
		c.resultRepo,
		c.progressTracker,
		c.workerManager,
		c.notifier,
		c.metricsCollector,
		c.retryPolicy,
	)
}

// GetAnalysisService 获取分析服务
func (c *AnalysisContainer) GetAnalysisService() analysisDomain.AnalysisService {
	return c.analysisService
}

// GetWorkerManager 获取工作器管理器
func (c *AnalysisContainer) GetWorkerManager() analysisDomain.AnalysisWorkerManager {
	return c.workerManager
}

// GetTaskQueue 获取任务队列
func (c *AnalysisContainer) GetTaskQueue() analysisDomain.AnalysisTaskQueue {
	return c.taskQueue
}

// GetTaskRepository 获取任务仓库
func (c *AnalysisContainer) GetTaskRepository() analysisDomain.AnalysisTaskRepository {
	return c.taskRepo
}

// GetResultRepository 获取结果仓库
func (c *AnalysisContainer) GetResultRepository() analysisDomain.AnalysisResultRepository {
	return c.resultRepo
}

// GetProgressTracker 获取进度跟踪器
func (c *AnalysisContainer) GetProgressTracker() analysisDomain.AnalysisProgressTracker {
	return c.progressTracker
}

// Cleanup 清理资源
func (c *AnalysisContainer) Cleanup() error {
	c.logger.Info("Cleaning up analysis container")

	// 停止工作器
	if err := c.workerManager.StopWorkers(context.Background()); err != nil {
		c.logger.Error("Failed to stop workers", zap.Error(err))
		return err
	}

	c.logger.Info("Analysis container cleanup completed")
	return nil
}

// Mock implementations for testing and development

// MockAnalysisEngine 模拟分析引擎
type MockAnalysisEngine struct{}

func (m *MockAnalysisEngine) Analyze(ctx context.Context, request *analysisDomain.AnalysisRequest) (*analysisDomain.AnalysisResult, error) {
	// 模拟分析过程
	time.Sleep(100 * time.Millisecond)

	return &analysisDomain.AnalysisResult{
		ID:              uuid.New().String(),
		TaskID:          uuid.New().String(), // 模拟任务ID
		AlertID:         fmt.Sprintf("%d", request.Alert.ID),
		Type:            request.Type,
		Status:          analysisDomain.AnalysisStatusCompleted,
		Result:          map[string]interface{}{"mock": "result"},
		ConfidenceScore: 0.8,
		ProcessingTime:  100 * time.Millisecond,
		Summary:         "Mock analysis result",
		Recommendations: []string{"Mock recommendation"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (m *MockAnalysisEngine) GetSupportedTypes() []analysisDomain.AnalysisType {
	return []analysisDomain.AnalysisType{
		analysisDomain.AnalysisTypeRootCause,
		analysisDomain.AnalysisTypeImpactAssess,
		analysisDomain.AnalysisTypeSolution,
	}
}

func (m *MockAnalysisEngine) ValidateRequest(request *analysisDomain.AnalysisRequest) error {
	return nil
}

func (m *MockAnalysisEngine) EstimateProcessingTime(request *analysisDomain.AnalysisRequest) time.Duration {
	return 100 * time.Millisecond
}

func (m *MockAnalysisEngine) GetEngineInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "MockEngine",
		"version": "1.0.0",
	}
}

// MockMetricsCollector 模拟指标收集器
type MockMetricsCollector struct{}

func (m *MockMetricsCollector) RecordTaskSubmitted(ctx context.Context, analysisType analysisDomain.AnalysisType) {}
func (m *MockMetricsCollector) RecordTaskStarted(ctx context.Context, taskID string, analysisType analysisDomain.AnalysisType) {}
func (m *MockMetricsCollector) RecordTaskCompleted(ctx context.Context, taskID string, analysisType analysisDomain.AnalysisType, duration time.Duration) {}
func (m *MockMetricsCollector) RecordTaskFailed(ctx context.Context, taskID string, analysisType analysisDomain.AnalysisType, err error) {}
func (m *MockMetricsCollector) RecordQueueSize(ctx context.Context, size int64) {}
func (m *MockMetricsCollector) RecordWorkerStatus(ctx context.Context, workerID string, status string) {}
func (m *MockMetricsCollector) GetMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{}
}

// MockNotifier 模拟通知器
type MockNotifier struct{}

func (m *MockNotifier) NotifyTaskCompleted(ctx context.Context, task *analysisDomain.AnalysisTask, result *analysisDomain.AnalysisResult) error {
	return nil
}
func (m *MockNotifier) NotifyTaskFailed(ctx context.Context, task *analysisDomain.AnalysisTask, err error) error {
	return nil
}
func (m *MockNotifier) NotifyProgressUpdate(ctx context.Context, progress *analysisDomain.AnalysisProgress) error {
	return nil
}
func (m *MockNotifier) RegisterCallback(taskID string, callback func(*analysisDomain.AnalysisResult, error)) error {
	return nil
}
func (m *MockNotifier) UnregisterCallback(taskID string) error {
	return nil
}

// MockRetryPolicy 模拟重试策略
type MockRetryPolicy struct{}

func (m *MockRetryPolicy) ShouldRetry(task *analysisDomain.AnalysisTask, err error) bool {
	return task.RetryCount < 3
}
func (m *MockRetryPolicy) GetRetryDelay(task *analysisDomain.AnalysisTask) time.Duration {
	return time.Duration(task.RetryCount) * time.Second
}
func (m *MockRetryPolicy) GetMaxRetries(analysisType analysisDomain.AnalysisType) int {
	return 3
}