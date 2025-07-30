package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// WorkerMonitoringService Worker监控服务
type WorkerMonitoringService struct {
	manager     *WorkerManager
	scaler      *WorkerScaler
	alerts      map[string]*MonitoringAlert
	alertRules  map[string]*AlertRule
	httpServer  *http.Server
	mutex       sync.RWMutex
	isRunning   bool
	stopChan    chan struct{}
}

// MonitoringAlert 监控告警
type MonitoringAlert struct {
	ID          string                 `json:"id"`
	RuleName    string                 `json:"rule_name"`
	WorkerName  string                 `json:"worker_name"`
	WorkerType  string                 `json:"worker_type"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Status      string                 `json:"status"` // firing, resolved
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	Extra       map[string]interface{} `json:"extra"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string        `json:"name"`
	Expression  string        `json:"expression"`
	Duration    time.Duration `json:"duration"`
	Level       string        `json:"level"`
	Message     string        `json:"message"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool          `json:"enabled"`
}

// MonitoringDashboard 监控仪表板数据
type MonitoringDashboard struct {
	Overview      *OverviewMetrics           `json:"overview"`
	WorkerMetrics map[string]*WorkerMetrics  `json:"worker_metrics"`
	QueueMetrics  map[string]*QueueMetrics   `json:"queue_metrics"`
	Alerts        []*MonitoringAlert         `json:"alerts"`
	ScalingEvents []*ScalingEvent            `json:"scaling_events"`
	Timestamp     time.Time                  `json:"timestamp"`
}

// OverviewMetrics 概览指标
type OverviewMetrics struct {
	TotalWorkers       int     `json:"total_workers"`
	HealthyWorkers     int     `json:"healthy_workers"`
	UnhealthyWorkers   int     `json:"unhealthy_workers"`
	TotalTasksProcessed int64   `json:"total_tasks_processed"`
	TotalTasksSucceeded int64   `json:"total_tasks_succeeded"`
	TotalTasksFailed   int64   `json:"total_tasks_failed"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	AverageLatency     time.Duration `json:"average_latency"`
	ActiveAlerts       int     `json:"active_alerts"`
}

// QueueMetrics 队列指标
type QueueMetrics struct {
	QueueName       string `json:"queue_name"`
	PendingTasks    int64  `json:"pending_tasks"`
	ProcessingTasks int64  `json:"processing_tasks"`
	CompletedTasks  int64  `json:"completed_tasks"`
	FailedTasks     int64  `json:"failed_tasks"`
	DeadLetterTasks int64  `json:"dead_letter_tasks"`
}

// ScalingEvent 扩缩容事件
type ScalingEvent struct {
	ID         string    `json:"id"`
	WorkerType string    `json:"worker_type"`
	Action     string    `json:"action"`
	OldCount   int       `json:"old_count"`
	NewCount   int       `json:"new_count"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// NewWorkerMonitoringService 创建Worker监控服务
func NewWorkerMonitoringService(manager *WorkerManager, scaler *WorkerScaler, port int) *WorkerMonitoringService {
	service := &WorkerMonitoringService{
		manager:    manager,
		scaler:     scaler,
		alerts:     make(map[string]*MonitoringAlert),
		alertRules: make(map[string]*AlertRule),
		stopChan:   make(chan struct{}),
	}

	// 设置默认告警规则
	service.setDefaultAlertRules()

	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", service.metricsHandler)
	mux.HandleFunc("/dashboard", service.dashboardHandler)
	mux.HandleFunc("/alerts", service.alertsHandler)
	mux.HandleFunc("/workers", service.workersHandler)
	mux.HandleFunc("/scaling", service.scalingHandler)
	mux.HandleFunc("/health", service.healthHandler)

	service.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return service
}

// Start 启动监控服务
func (s *WorkerMonitoringService) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return fmt.Errorf("monitoring service is already running")
	}
	s.isRunning = true
	s.mutex.Unlock()

	logger.L.Info("Starting worker monitoring service")

	// 启动HTTP服务器
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Error("Monitoring HTTP server failed", zap.Error(err))
		}
	}()

	// 启动监控循环
	go s.monitoringLoop(ctx)

	// 启动告警检查循环
	go s.alertingLoop(ctx)

	return nil
}

// Stop 停止监控服务
func (s *WorkerMonitoringService) Stop(ctx context.Context) error {
	s.mutex.Lock()
	if !s.isRunning {
		s.mutex.Unlock()
		return nil
	}
	s.isRunning = false
	s.mutex.Unlock()

	logger.L.Info("Stopping worker monitoring service")

	// 停止HTTP服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.L.Error("Failed to shutdown monitoring HTTP server", zap.Error(err))
	}

	close(s.stopChan)
	return nil
}

// monitoringLoop 监控循环
func (s *WorkerMonitoringService) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectAndProcessMetrics(ctx)
		}
	}
}

// alertingLoop 告警检查循环
func (s *WorkerMonitoringService) alertingLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.evaluateAlertRules(ctx)
		}
	}
}

// collectAndProcessMetrics 收集和处理指标
func (s *WorkerMonitoringService) collectAndProcessMetrics(ctx context.Context) {
	// 这里可以添加指标收集和处理逻辑
	logger.L.Debug("Collecting and processing metrics")
}

// evaluateAlertRules 评估告警规则
func (s *WorkerMonitoringService) evaluateAlertRules(ctx context.Context) {
	s.mutex.RLock()
	rules := make(map[string]*AlertRule)
	for k, v := range s.alertRules {
		if v.Enabled {
			rules[k] = v
		}
	}
	s.mutex.RUnlock()

	workers := s.manager.ListWorkers()
	workerMetrics := s.scaler.GetMetrics()

	for ruleName, rule := range rules {
		s.evaluateRule(ctx, rule, workers, workerMetrics)
	}
}

// evaluateRule 评估单个告警规则
func (s *WorkerMonitoringService) evaluateRule(ctx context.Context, rule *AlertRule, workers map[string]*WorkerInstance, metrics map[string]*WorkerMetrics) {
	switch rule.Expression {
	case "worker_down":
		s.checkWorkerDown(rule, workers)
	case "high_error_rate":
		s.checkHighErrorRate(rule, metrics)
	case "high_latency":
		s.checkHighLatency(rule, metrics)
	case "queue_backlog":
		s.checkQueueBacklog(rule, metrics)
	case "worker_unhealthy":
		s.checkWorkerUnhealthy(rule, metrics)
	}
}

// checkWorkerDown 检查Worker宕机
func (s *WorkerMonitoringService) checkWorkerDown(rule *AlertRule, workers map[string]*WorkerInstance) {
	for name, worker := range workers {
		if !worker.IsRunning() {
			alertID := fmt.Sprintf("worker_down_%s", name)
			
			if _, exists := s.alerts[alertID]; !exists {
				alert := &MonitoringAlert{
					ID:         alertID,
					RuleName:   rule.Name,
					WorkerName: name,
					WorkerType: worker.GetConfig().Type,
					Level:      rule.Level,
					Message:    fmt.Sprintf("Worker %s is down", name),
					Labels:     rule.Labels,
					Annotations: rule.Annotations,
					Status:     "firing",
					StartsAt:   time.Now(),
					Extra: map[string]interface{}{
						"worker_config": worker.GetConfig(),
					},
				}
				
				s.mutex.Lock()
				s.alerts[alertID] = alert
				s.mutex.Unlock()
				
				logger.L.Warn("Worker down alert fired",
					zap.String("worker", name),
					zap.String("alert_id", alertID),
				)
			}
		} else {
			// 检查是否需要解决告警
			alertID := fmt.Sprintf("worker_down_%s", name)
			if alert, exists := s.alerts[alertID]; exists && alert.Status == "firing" {
				now := time.Now()
				alert.Status = "resolved"
				alert.EndsAt = &now
				
				logger.L.Info("Worker down alert resolved",
					zap.String("worker", name),
					zap.String("alert_id", alertID),
				)
			}
		}
	}
}

// checkHighErrorRate 检查高错误率
func (s *WorkerMonitoringService) checkHighErrorRate(rule *AlertRule, metrics map[string]*WorkerMetrics) {
	for workerName, metric := range metrics {
		if metric.ErrorRate > 0.1 { // 错误率超过10%
			alertID := fmt.Sprintf("high_error_rate_%s", workerName)
			
			if _, exists := s.alerts[alertID]; !exists {
				alert := &MonitoringAlert{
					ID:         alertID,
					RuleName:   rule.Name,
					WorkerName: workerName,
					WorkerType: metric.WorkerType,
					Level:      rule.Level,
					Message:    fmt.Sprintf("Worker %s has high error rate: %.2f%%", workerName, metric.ErrorRate*100),
					Labels:     rule.Labels,
					Annotations: rule.Annotations,
					Status:     "firing",
					StartsAt:   time.Now(),
					Extra: map[string]interface{}{
						"error_rate": metric.ErrorRate,
						"threshold":  0.1,
					},
				}
				
				s.mutex.Lock()
				s.alerts[alertID] = alert
				s.mutex.Unlock()
				
				logger.L.Warn("High error rate alert fired",
					zap.String("worker", workerName),
					zap.Float64("error_rate", metric.ErrorRate),
				)
			}
		}
	}
}

// checkHighLatency 检查高延迟
func (s *WorkerMonitoringService) checkHighLatency(rule *AlertRule, metrics map[string]*WorkerMetrics) {
	threshold := 30 * time.Second
	
	for workerName, metric := range metrics {
		if metric.AverageLatency > threshold {
			alertID := fmt.Sprintf("high_latency_%s", workerName)
			
			if _, exists := s.alerts[alertID]; !exists {
				alert := &MonitoringAlert{
					ID:         alertID,
					RuleName:   rule.Name,
					WorkerName: workerName,
					WorkerType: metric.WorkerType,
					Level:      rule.Level,
					Message:    fmt.Sprintf("Worker %s has high latency: %v", workerName, metric.AverageLatency),
					Labels:     rule.Labels,
					Annotations: rule.Annotations,
					Status:     "firing",
					StartsAt:   time.Now(),
					Extra: map[string]interface{}{
						"latency":   metric.AverageLatency,
						"threshold": threshold,
					},
				}
				
				s.mutex.Lock()
				s.alerts[alertID] = alert
				s.mutex.Unlock()
				
				logger.L.Warn("High latency alert fired",
					zap.String("worker", workerName),
					zap.Duration("latency", metric.AverageLatency),
				)
			}
		}
	}
}

// checkQueueBacklog 检查队列积压
func (s *WorkerMonitoringService) checkQueueBacklog(rule *AlertRule, metrics map[string]*WorkerMetrics) {
	threshold := int64(50)
	
	for workerName, metric := range metrics {
		for queueName, queueLength := range metric.QueueLengths {
			if queueLength > threshold {
				alertID := fmt.Sprintf("queue_backlog_%s_%s", workerName, queueName)
				
				if _, exists := s.alerts[alertID]; !exists {
					alert := &MonitoringAlert{
						ID:         alertID,
						RuleName:   rule.Name,
						WorkerName: workerName,
						WorkerType: metric.WorkerType,
						Level:      rule.Level,
						Message:    fmt.Sprintf("Queue %s has backlog: %d tasks", queueName, queueLength),
						Labels:     rule.Labels,
						Annotations: rule.Annotations,
						Status:     "firing",
						StartsAt:   time.Now(),
						Extra: map[string]interface{}{
							"queue_name":   queueName,
							"queue_length": queueLength,
							"threshold":    threshold,
						},
					}
					
					s.mutex.Lock()
					s.alerts[alertID] = alert
					s.mutex.Unlock()
					
					logger.L.Warn("Queue backlog alert fired",
						zap.String("worker", workerName),
						zap.String("queue", queueName),
						zap.Int64("length", queueLength),
					)
				}
			}
		}
	}
}

// checkWorkerUnhealthy 检查Worker不健康
func (s *WorkerMonitoringService) checkWorkerUnhealthy(rule *AlertRule, metrics map[string]*WorkerMetrics) {
	for workerName, metric := range metrics {
		if metric.HealthStatus == "unhealthy" || metric.HealthStatus == "degraded" {
			alertID := fmt.Sprintf("worker_unhealthy_%s", workerName)
			
			if _, exists := s.alerts[alertID]; !exists {
				alert := &MonitoringAlert{
					ID:         alertID,
					RuleName:   rule.Name,
					WorkerName: workerName,
					WorkerType: metric.WorkerType,
					Level:      rule.Level,
					Message:    fmt.Sprintf("Worker %s is %s", workerName, metric.HealthStatus),
					Labels:     rule.Labels,
					Annotations: rule.Annotations,
					Status:     "firing",
					StartsAt:   time.Now(),
					Extra: map[string]interface{}{
						"health_status": metric.HealthStatus,
					},
				}
				
				s.mutex.Lock()
				s.alerts[alertID] = alert
				s.mutex.Unlock()
				
				logger.L.Warn("Worker unhealthy alert fired",
					zap.String("worker", workerName),
					zap.String("health_status", metric.HealthStatus),
				)
			}
		}
	}
}

// setDefaultAlertRules 设置默认告警规则
func (s *WorkerMonitoringService) setDefaultAlertRules() {
	rules := map[string]*AlertRule{
		"worker_down": {
			Name:       "WorkerDown",
			Expression: "worker_down",
			Duration:   1 * time.Minute,
			Level:      "critical",
			Message:    "Worker instance is down",
			Labels:     map[string]string{"severity": "critical"},
			Annotations: map[string]string{
				"description": "Worker instance has been down for more than 1 minute",
				"runbook":     "Check worker logs and restart if necessary",
			},
			Enabled: true,
		},
		"high_error_rate": {
			Name:       "HighErrorRate",
			Expression: "high_error_rate",
			Duration:   5 * time.Minute,
			Level:      "warning",
			Message:    "Worker has high error rate",
			Labels:     map[string]string{"severity": "warning"},
			Annotations: map[string]string{
				"description": "Worker error rate is above 10%",
				"runbook":     "Check worker logs for error patterns",
			},
			Enabled: true,
		},
		"high_latency": {
			Name:       "HighLatency",
			Expression: "high_latency",
			Duration:   5 * time.Minute,
			Level:      "warning",
			Message:    "Worker has high task processing latency",
			Labels:     map[string]string{"severity": "warning"},
			Annotations: map[string]string{
				"description": "Worker average latency is above 30 seconds",
				"runbook":     "Check system resources and task complexity",
			},
			Enabled: true,
		},
		"queue_backlog": {
			Name:       "QueueBacklog",
			Expression: "queue_backlog",
			Duration:   2 * time.Minute,
			Level:      "warning",
			Message:    "Queue has significant backlog",
			Labels:     map[string]string{"severity": "warning"},
			Annotations: map[string]string{
				"description": "Queue length is above 50 tasks",
				"runbook":     "Consider scaling up workers or check for processing issues",
			},
			Enabled: true,
		},
		"worker_unhealthy": {
			Name:       "WorkerUnhealthy",
			Expression: "worker_unhealthy",
			Duration:   3 * time.Minute,
			Level:      "warning",
			Message:    "Worker is in unhealthy state",
			Labels:     map[string]string{"severity": "warning"},
			Annotations: map[string]string{
				"description": "Worker health check is failing",
				"runbook":     "Check worker health endpoint and logs",
			},
			Enabled: true,
		},
	}

	s.mutex.Lock()
	s.alertRules = rules
	s.mutex.Unlock()

	logger.L.Info("Default alert rules set",
		zap.Int("rule_count", len(rules)),
	)
}

// HTTP处理器

// metricsHandler Prometheus指标处理器
func (s *WorkerMonitoringService) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	
	workers := s.manager.ListWorkers()
	workerMetrics := s.scaler.GetMetrics()

	// 输出Worker指标
	for name, worker := range workers {
		stats, err := worker.GetStats(context.Background())
		if err != nil {
			continue
		}

		fmt.Fprintf(w, "# HELP alertagent_worker_tasks_processed_total Total number of tasks processed by worker\n")
		fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_processed_total counter\n")
		fmt.Fprintf(w, "alertagent_worker_tasks_processed_total{worker=\"%s\",type=\"%s\"} %d\n", 
			name, stats.Type, stats.TasksProcessed)

		fmt.Fprintf(w, "# HELP alertagent_worker_tasks_succeeded_total Total number of tasks succeeded by worker\n")
		fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_succeeded_total counter\n")
		fmt.Fprintf(w, "alertagent_worker_tasks_succeeded_total{worker=\"%s\",type=\"%s\"} %d\n", 
			name, stats.Type, stats.TasksSucceeded)

		fmt.Fprintf(w, "# HELP alertagent_worker_tasks_failed_total Total number of tasks failed by worker\n")
		fmt.Fprintf(w, "# TYPE alertagent_worker_tasks_failed_total counter\n")
		fmt.Fprintf(w, "alertagent_worker_tasks_failed_total{worker=\"%s\",type=\"%s\"} %d\n", 
			name, stats.Type, stats.TasksFailed)

		if metric, exists := workerMetrics[name]; exists {
			fmt.Fprintf(w, "# HELP alertagent_worker_cpu_usage Worker CPU usage percentage\n")
			fmt.Fprintf(w, "# TYPE alertagent_worker_cpu_usage gauge\n")
			fmt.Fprintf(w, "alertagent_worker_cpu_usage{worker=\"%s\",type=\"%s\"} %f\n", 
				name, stats.Type, metric.CPUUsage)

			fmt.Fprintf(w, "# HELP alertagent_worker_memory_usage Worker memory usage percentage\n")
			fmt.Fprintf(w, "# TYPE alertagent_worker_memory_usage gauge\n")
			fmt.Fprintf(w, "alertagent_worker_memory_usage{worker=\"%s\",type=\"%s\"} %f\n", 
				name, stats.Type, metric.MemoryUsage)
		}
	}

	// 输出告警指标
	s.mutex.RLock()
	activeAlerts := 0
	for _, alert := range s.alerts {
		if alert.Status == "firing" {
			activeAlerts++
		}
	}
	s.mutex.RUnlock()

	fmt.Fprintf(w, "# HELP alertagent_active_alerts Number of active alerts\n")
	fmt.Fprintf(w, "# TYPE alertagent_active_alerts gauge\n")
	fmt.Fprintf(w, "alertagent_active_alerts %d\n", activeAlerts)
}

// dashboardHandler 仪表板数据处理器
func (s *WorkerMonitoringService) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	dashboard := s.generateDashboard()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// alertsHandler 告警处理器
func (s *WorkerMonitoringService) alertsHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	alerts := make([]*MonitoringAlert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// workersHandler Worker信息处理器
func (s *WorkerMonitoringService) workersHandler(w http.ResponseWriter, r *http.Request) {
	workers := s.manager.ListWorkers()
	workerMetrics := s.scaler.GetMetrics()

	result := make(map[string]interface{})
	for name, worker := range workers {
		stats, err := worker.GetStats(context.Background())
		if err != nil {
			continue
		}

		workerInfo := map[string]interface{}{
			"config": worker.GetConfig(),
			"stats":  stats,
		}

		if metric, exists := workerMetrics[name]; exists {
			workerInfo["metrics"] = metric
		}

		result[name] = workerInfo
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// scalingHandler 扩缩容信息处理器
func (s *WorkerMonitoringService) scalingHandler(w http.ResponseWriter, r *http.Request) {
	policies := s.scaler.GetScalingPolicies()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// healthHandler 健康检查处理器
func (s *WorkerMonitoringService) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "worker-monitoring",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// generateDashboard 生成仪表板数据
func (s *WorkerMonitoringService) generateDashboard() *MonitoringDashboard {
	workers := s.manager.ListWorkers()
	workerMetrics := s.scaler.GetMetrics()

	// 计算概览指标
	overview := &OverviewMetrics{}
	overview.TotalWorkers = len(workers)

	for name, worker := range workers {
		stats, err := worker.GetStats(context.Background())
		if err != nil {
			continue
		}

		if worker.IsRunning() {
			overview.HealthyWorkers++
		} else {
			overview.UnhealthyWorkers++
		}

		overview.TotalTasksProcessed += stats.TasksProcessed
		overview.TotalTasksSucceeded += stats.TasksSucceeded
		overview.TotalTasksFailed += stats.TasksFailed
	}

	if overview.TotalTasksProcessed > 0 {
		overview.OverallSuccessRate = float64(overview.TotalTasksSucceeded) / float64(overview.TotalTasksProcessed)
	}

	// 计算活跃告警数量
	s.mutex.RLock()
	for _, alert := range s.alerts {
		if alert.Status == "firing" {
			overview.ActiveAlerts++
		}
	}
	alerts := make([]*MonitoringAlert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	s.mutex.RUnlock()

	return &MonitoringDashboard{
		Overview:      overview,
		WorkerMetrics: workerMetrics,
		QueueMetrics:  make(map[string]*QueueMetrics), // 需要实现队列指标收集
		Alerts:        alerts,
		ScalingEvents: make([]*ScalingEvent, 0), // 需要实现扩缩容事件记录
		Timestamp:     time.Now(),
	}
}