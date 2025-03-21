package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HealthService 健康检查服务
type HealthService struct {
	endpoint   string
	timeout    time.Duration
	client     *http.Client
	logger     *zap.Logger
	statusLock sync.RWMutex
	isHealthy  bool
	lastCheck  time.Time
}

// NewHealthService 创建健康检查服务实例
func NewHealthService(endpoint string, timeout int) *HealthService {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return &HealthService{
		endpoint:  endpoint,
		timeout:   time.Duration(timeout) * time.Second,
		client:    client,
		logger:    zap.L(),
		isHealthy: false,
	}
}

// StartHealthCheck 启动健康检查
func (s *HealthService) StartHealthCheck(ctx context.Context, interval time.Duration) {
	s.logger.Info("Starting health check service",
		zap.String("endpoint", s.endpoint),
		zap.Duration("interval", interval),
	)

	// 立即执行一次健康检查
	s.checkHealth(ctx)

	// 定期执行健康检查
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.checkHealth(ctx)
			case <-ctx.Done():
				s.logger.Info("Health check service stopped")
				return
			}
		}
	}()
}

// checkHealth 执行健康检查
func (s *HealthService) checkHealth(ctx context.Context) {
	// 创建一个带有超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 构建请求URL（Ollama健康检查端点）
	url := fmt.Sprintf("%s/api/tags", s.endpoint)
	req, err := http.NewRequestWithContext(timeoutCtx, "GET", url, nil)
	if err != nil {
		s.setUnhealthy(fmt.Errorf("failed to create request: %w", err))
		return
	}

	// 执行请求
	resp, err := s.client.Do(req)

	// 检查是否是上下文取消或超时
	if timeoutCtx.Err() != nil {
		s.setUnhealthy(fmt.Errorf("request timed out: %w", timeoutCtx.Err()))
		return
	}

	// 检查其他错误
	if err != nil {
		s.setUnhealthy(fmt.Errorf("request failed: %w", err))
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		s.setUnhealthy(fmt.Errorf("unhealthy status code: %d", resp.StatusCode))
		return
	}

	// 服务健康
	s.setHealthy()
}

// setHealthy 设置服务为健康状态
func (s *HealthService) setHealthy() {
	s.statusLock.Lock()
	defer s.statusLock.Unlock()

	prevStatus := s.isHealthy
	s.isHealthy = true
	s.lastCheck = time.Now()

	if !prevStatus {
		s.logger.Info("Ollama service is now healthy",
			zap.String("endpoint", s.endpoint),
		)
	}
}

// setUnhealthy 设置服务为不健康状态
func (s *HealthService) setUnhealthy(err error) {
	s.statusLock.Lock()
	defer s.statusLock.Unlock()

	prevStatus := s.isHealthy
	s.isHealthy = false
	s.lastCheck = time.Now()

	if prevStatus {
		s.logger.Warn("Ollama service is now unhealthy",
			zap.String("endpoint", s.endpoint),
			zap.Error(err),
		)
	} else {
		s.logger.Debug("Ollama service remains unhealthy",
			zap.String("endpoint", s.endpoint),
			zap.Error(err),
		)
	}
}

// IsHealthy 检查服务是否健康
func (s *HealthService) IsHealthy() bool {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()
	return s.isHealthy
}

// GetLastCheckTime 获取最后一次检查时间
func (s *HealthService) GetLastCheckTime() time.Time {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()
	return s.lastCheck
}
