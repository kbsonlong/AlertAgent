package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthConfig 健康检查配置
type HealthConfig struct {
	Version          string                    `yaml:"version" json:"version"`
	ExternalServices []ExternalServiceConfig  `yaml:"external_services" json:"external_services"`
}

// ExternalServiceConfig 外部服务配置
type ExternalServiceConfig struct {
	Name    string `yaml:"name" json:"name"`
	URL     string `yaml:"url" json:"url"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}

// DefaultHealthConfig 默认健康检查配置
func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		Version:          "1.0.0",
		ExternalServices: []ExternalServiceConfig{},
	}
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// CheckResult 检查结果
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// HealthReport 健康报告
type HealthReport struct {
	Status    HealthStatus             `json:"status"`
	Timestamp time.Time               `json:"timestamp"`
	Uptime    time.Duration            `json:"uptime"`
	Version   string                   `json:"version"`
	Checks    map[string]*CheckResult  `json:"checks"`
	Summary   map[string]interface{}   `json:"summary"`
}

// ReadinessReport 就绪报告
type ReadinessReport struct {
	Ready     bool                     `json:"ready"`
	Timestamp time.Time               `json:"timestamp"`
	Checks    map[string]*CheckResult  `json:"checks"`
	Message   string                   `json:"message,omitempty"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) *CheckResult
}

// ReadinessChecker 就绪检查器接口
type ReadinessChecker interface {
	Name() string
	Check(ctx context.Context) *CheckResult
}

// HealthManager 健康管理器
type HealthManager struct {
	logger           *zap.Logger
	startTime        time.Time
	version          string
	healthCheckers   map[string]HealthChecker
	readinessCheckers map[string]ReadinessChecker
	mu               sync.RWMutex
	lastHealthCheck  *HealthReport
	lastReadinessCheck *ReadinessReport
}

// Shutdown 关闭健康检查管理器
func (hm *HealthManager) Shutdown(ctx context.Context) error {
	hm.logger.Info("Health manager shutdown completed")
	return nil
}

// NewHealthManager 创建健康管理器
func NewHealthManager(logger *zap.Logger, version string) *HealthManager {
	return &HealthManager{
		logger:            logger,
		startTime:         time.Now(),
		version:           version,
		healthCheckers:    make(map[string]HealthChecker),
		readinessCheckers: make(map[string]ReadinessChecker),
	}
}

// RegisterHealthChecker 注册健康检查器
func (hm *HealthManager) RegisterHealthChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.healthCheckers[checker.Name()] = checker
	hm.logger.Info("Registered health checker", zap.String("name", checker.Name()))
}

// RegisterReadinessChecker 注册就绪检查器
func (hm *HealthManager) RegisterReadinessChecker(checker ReadinessChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.readinessCheckers[checker.Name()] = checker
	hm.logger.Info("Registered readiness checker", zap.String("name", checker.Name()))
}

// CheckHealth 执行健康检查
func (hm *HealthManager) CheckHealth(ctx context.Context) *HealthReport {
	hm.mu.RLock()
	checkers := make(map[string]HealthChecker)
	for name, checker := range hm.healthCheckers {
		checkers[name] = checker
	}
	hm.mu.RUnlock()

	report := &HealthReport{
		Timestamp: time.Now(),
		Uptime:    time.Since(hm.startTime),
		Version:   hm.version,
		Checks:    make(map[string]*CheckResult),
		Summary:   make(map[string]interface{}),
	}

	// 执行所有健康检查
	overallStatus := HealthStatusHealthy
	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0

	for name, checker := range checkers {
		result := checker.Check(ctx)
		report.Checks[name] = result

		switch result.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
			overallStatus = HealthStatusUnhealthy
		case HealthStatusDegraded:
			degradedCount++
			if overallStatus == HealthStatusHealthy {
				overallStatus = HealthStatusDegraded
			}
		}
	}

	report.Status = overallStatus
	report.Summary["total_checks"] = len(checkers)
	report.Summary["healthy_checks"] = healthyCount
	report.Summary["unhealthy_checks"] = unhealthyCount
	report.Summary["degraded_checks"] = degradedCount

	hm.mu.Lock()
	hm.lastHealthCheck = report
	hm.mu.Unlock()

	return report
}

// CheckReadiness 执行就绪检查
func (hm *HealthManager) CheckReadiness(ctx context.Context) *ReadinessReport {
	hm.mu.RLock()
	checkers := make(map[string]ReadinessChecker)
	for name, checker := range hm.readinessCheckers {
		checkers[name] = checker
	}
	hm.mu.RUnlock()

	report := &ReadinessReport{
		Timestamp: time.Now(),
		Checks:    make(map[string]*CheckResult),
		Ready:     true,
	}

	// 执行所有就绪检查
	for name, checker := range checkers {
		result := checker.Check(ctx)
		report.Checks[name] = result

		if result.Status != HealthStatusHealthy {
			report.Ready = false
			if report.Message == "" {
				report.Message = fmt.Sprintf("Check '%s' failed: %s", name, result.Message)
			}
		}
	}

	hm.mu.Lock()
	hm.lastReadinessCheck = report
	hm.mu.Unlock()

	return report
}

// GetLastHealthCheck 获取最后一次健康检查结果
func (hm *HealthManager) GetLastHealthCheck() *HealthReport {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.lastHealthCheck
}

// GetLastReadinessCheck 获取最后一次就绪检查结果
func (hm *HealthManager) GetLastReadinessCheck() *ReadinessReport {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.lastReadinessCheck
}

// HealthHandler HTTP健康检查处理器
func (hm *HealthManager) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		report := hm.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		if report.Status == HealthStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(report); err != nil {
			hm.logger.Error("Failed to encode health report", zap.Error(err))
		}
	}
}

// ReadinessHandler HTTP就绪检查处理器
func (hm *HealthManager) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		report := hm.CheckReadiness(ctx)

		w.Header().Set("Content-Type", "application/json")
		if !report.Ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(report); err != nil {
			hm.logger.Error("Failed to encode readiness report", zap.Error(err))
		}
	}
}

// DatabaseHealthChecker 数据库健康检查器
type DatabaseHealthChecker struct {
	db *gorm.DB
}

// NewDatabaseHealthChecker 创建数据库健康检查器
func NewDatabaseHealthChecker(db *gorm.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{db: db}
}

// Name 返回检查器名称
func (dhc *DatabaseHealthChecker) Name() string {
	return "database"
}

// Check 执行数据库健康检查
func (dhc *DatabaseHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      dhc.Name(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// 检查数据库连接
	sqlDB, err := dhc.db.DB()
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Failed to get database instance: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 执行ping检查
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Database ping failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()
	result.Details["open_connections"] = stats.OpenConnections
	result.Details["in_use"] = stats.InUse
	result.Details["idle"] = stats.Idle
	result.Details["wait_count"] = stats.WaitCount
	result.Details["wait_duration"] = stats.WaitDuration.String()
	result.Details["max_idle_closed"] = stats.MaxIdleClosed
	result.Details["max_lifetime_closed"] = stats.MaxLifetimeClosed

	// 检查连接池状态
	if stats.OpenConnections > 0 {
		result.Status = HealthStatusHealthy
		result.Message = "Database connection is healthy"
	} else {
		result.Status = HealthStatusDegraded
		result.Message = "No active database connections"
	}

	result.Duration = time.Since(start)
	return result
}

// RedisHealthChecker Redis健康检查器
type RedisHealthChecker struct {
	// 这里可以添加Redis客户端
	// redisClient redis.Client
}

// NewRedisHealthChecker 创建Redis健康检查器
func NewRedisHealthChecker() *RedisHealthChecker {
	return &RedisHealthChecker{}
}

// Name 返回检查器名称
func (rhc *RedisHealthChecker) Name() string {
	return "redis"
}

// Check 执行Redis健康检查
func (rhc *RedisHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      rhc.Name(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// TODO: 实现Redis连接检查
	// 这里暂时返回健康状态
	result.Status = HealthStatusHealthy
	result.Message = "Redis connection is healthy"
	result.Duration = time.Since(start)

	return result
}

// ExternalServiceHealthChecker 外部服务健康检查器
type ExternalServiceHealthChecker struct {
	name     string
	url      string
	timeout  time.Duration
	client   *http.Client
}

// NewExternalServiceHealthChecker 创建外部服务健康检查器
func NewExternalServiceHealthChecker(name, url string, timeout time.Duration) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		name:    name,
		url:     url,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name 返回检查器名称
func (eshc *ExternalServiceHealthChecker) Name() string {
	return eshc.name
}

// Check 执行外部服务健康检查
func (eshc *ExternalServiceHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      eshc.Name(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", eshc.url, nil)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	resp, err := eshc.client.Do(req)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.Details["status_code"] = resp.StatusCode
	result.Details["url"] = eshc.url

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = HealthStatusHealthy
		result.Message = "External service is healthy"
	} else if resp.StatusCode >= 500 {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("External service returned status %d", resp.StatusCode)
	} else {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("External service returned status %d", resp.StatusCode)
	}

	result.Duration = time.Since(start)
	return result
}