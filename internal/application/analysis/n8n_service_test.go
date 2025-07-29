package analysis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"alert_agent/internal/domain/alert"
	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkflowManager mock implementation of N8NWorkflowManager
type MockWorkflowManager struct {
	mock.Mock
}

func (m *MockWorkflowManager) TriggerAnalysisWorkflow(ctx context.Context, alertID string, analysisType string, metadata map[string]interface{}) (*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, alertID, analysisType, metadata)
	return args.Get(0).(*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockWorkflowManager) MonitorExecution(ctx context.Context, executionID string) (<-chan *analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, executionID)
	return args.Get(0).(<-chan *analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockWorkflowManager) HandleCallback(ctx context.Context, executionID string, data map[string]interface{}) error {
	args := m.Called(ctx, executionID, data)
	return args.Error(0)
}

func (m *MockWorkflowManager) RetryFailedExecution(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, executionID)
	return args.Get(0).(*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockWorkflowManager) GetExecutionLogs(ctx context.Context, executionID string) ([]string, error) {
	args := m.Called(ctx, executionID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockWorkflowManager) GetWorkflowMetrics(ctx context.Context, workflowID string, timeRange time.Duration) (map[string]interface{}, error) {
	args := m.Called(ctx, workflowID, timeRange)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockAlertRepository mock implementation of AlertRepository
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id uint) (*model.Alert, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) List(ctx context.Context, filter alert.AlertFilter) ([]*model.Alert, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) Count(ctx context.Context, filter alert.AlertFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlertRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) BatchUpdate(ctx context.Context, ids []uint, updates map[string]interface{}) error {
	args := m.Called(ctx, ids, updates)
	return args.Error(0)
}

func (m *MockAlertRepository) GetSimilarAlerts(ctx context.Context, alert *model.Alert, limit int) ([]*model.Alert, error) {
	args := m.Called(ctx, alert, limit)
	return args.Get(0).([]*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetStatistics(ctx context.Context, filter alert.AlertFilter) (*alert.AlertStatistics, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*alert.AlertStatistics), args.Error(1)
}

func (m *MockAlertRepository) GetRecentAlerts(ctx context.Context, duration time.Duration, limit int) ([]*model.Alert, error) {
	args := m.Called(ctx, duration, limit)
	return args.Get(0).([]*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) UpdateAnalysisResult(ctx context.Context, alertID uint, analysis string) error {
	args := m.Called(ctx, alertID, analysis)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAlertsForAnalysis(ctx context.Context, limit int) ([]*model.Alert, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) MarkAsAnalyzed(ctx context.Context, alertID uint) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

// MockExecutionRepository mock implementation of N8NWorkflowExecutionRepository
type MockExecutionRepository struct {
	mock.Mock
}

func (m *MockExecutionRepository) Create(ctx context.Context, execution *analysis.N8NWorkflowExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockExecutionRepository) GetByID(ctx context.Context, id string) (*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockExecutionRepository) Update(ctx context.Context, execution *analysis.N8NWorkflowExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockExecutionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockExecutionRepository) ListByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, workflowID, limit, offset)
	return args.Get(0).([]*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockExecutionRepository) ListByStatus(ctx context.Context, status analysis.N8NWorkflowStatus, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockExecutionRepository) ListByDateRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0).([]*analysis.N8NWorkflowExecution), args.Error(1)
}

func (m *MockExecutionRepository) GetStatistics(ctx context.Context, workflowID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	args := m.Called(ctx, workflowID, startTime, endTime)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestN8NAnalysisService_AnalyzeAlert 测试单个告警分析
func TestN8NAnalysisService_AnalyzeAlert(t *testing.T) {
	// 创建mock对象
	mockWorkflowManager := &MockWorkflowManager{}
	mockAlertRepo := &MockAlertRepository{}
	mockExecutionRepo := &MockExecutionRepository{}

	// 创建服务实例
	service := NewN8NAnalysisService(mockWorkflowManager, mockAlertRepo, mockExecutionRepo)

	// 测试数据
	alertID := uint(123)
	workflowTemplateID := "test-workflow-123"

	// 设置mock期望
	mockAlert := &model.Alert{
		ID:    alertID,
		Title: "Test Alert",
		Level: "critical",
	}
	mockAlertRepo.On("GetByID", mock.Anything, alertID).Return(mockAlert, nil)

	// 设置执行记录查询的mock
	mockExecutions := []*analysis.N8NWorkflowExecution{}
	mockExecutionRepo.On("ListByWorkflowID", mock.Anything, workflowTemplateID, 1, 0).Return(mockExecutions, nil)

	// 设置工作流触发的mock
	mockExecution := &analysis.N8NWorkflowExecution{
		ID:         "exec-123",
		WorkflowID: workflowTemplateID,
		Status:     analysis.N8NWorkflowStatusRunning,
		Metadata: map[string]interface{}{
			"alert_id": fmt.Sprintf("%d", alertID),
			"type":     "alert_analysis",
		},
	}
	mockWorkflowManager.On("TriggerAnalysisWorkflow", mock.Anything, fmt.Sprintf("%d", alertID), workflowTemplateID, mock.Anything).Return(mockExecution, nil)

	// 执行测试
	result, err := service.AnalyzeAlert(context.Background(), alertID, workflowTemplateID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "exec-123", result.ID)
	assert.Equal(t, workflowTemplateID, result.WorkflowID)
	assert.Equal(t, analysis.N8NWorkflowStatusRunning, result.Status)

	// 验证mock调用
	mockAlertRepo.AssertExpectations(t)
	mockExecutionRepo.AssertExpectations(t)
	mockWorkflowManager.AssertExpectations(t)
}

// TestN8NAnalysisService_BatchAnalyzeAlerts 测试批量告警分析
func TestN8NAnalysisService_BatchAnalyzeAlerts(t *testing.T) {
	// 创建mock对象
	mockWorkflowManager := &MockWorkflowManager{}
	mockAlertRepo := &MockAlertRepository{}
	mockExecutionRepo := &MockExecutionRepository{}

	// 创建服务实例
	service := NewN8NAnalysisService(mockWorkflowManager, mockAlertRepo, mockExecutionRepo)

	ctx := context.Background()
	config := N8NAnalysisConfig{
		DefaultWorkflowTemplateID: "workflow-template-1",
		BatchSize:                 2,
		ProcessInterval:           time.Second,
		MaxRetries:                3,
		Timeout:                   time.Minute,
		AutoAnalysisEnabled:       true,
	}

	// 设置mock期望
	alerts := []*model.Alert{
		{ID: 1, Title: "Alert 1", Level: "high"},
		{ID: 2, Title: "Alert 2", Level: "critical"},
	}

	// Mock GetAlertsForAnalysis
	mockAlertRepo.On("GetAlertsForAnalysis", ctx, config.BatchSize).Return(alerts, nil)
	
	// Mock GetByID for each alert
	mockAlertRepo.On("GetByID", ctx, uint(1)).Return(alerts[0], nil)
	mockAlertRepo.On("GetByID", ctx, uint(2)).Return(alerts[1], nil)
	
	// Mock ListByWorkflowID for each alert
	mockExecutions := []*analysis.N8NWorkflowExecution{}
	mockExecutionRepo.On("ListByWorkflowID", ctx, config.DefaultWorkflowTemplateID, 1, 0).Return(mockExecutions, nil).Times(2)
	
	// Mock TriggerAnalysisWorkflow for each alert
	mockExecution1 := &analysis.N8NWorkflowExecution{ID: "exec-1", WorkflowID: config.DefaultWorkflowTemplateID, Status: analysis.N8NWorkflowStatusRunning}
	mockExecution2 := &analysis.N8NWorkflowExecution{ID: "exec-2", WorkflowID: config.DefaultWorkflowTemplateID, Status: analysis.N8NWorkflowStatusRunning}
	mockWorkflowManager.On("TriggerAnalysisWorkflow", ctx, "1", config.DefaultWorkflowTemplateID, mock.Anything).Return(mockExecution1, nil)
	mockWorkflowManager.On("TriggerAnalysisWorkflow", ctx, "2", config.DefaultWorkflowTemplateID, mock.Anything).Return(mockExecution2, nil)

	// 执行测试
	err := service.BatchAnalyzeAlerts(ctx, config)

	// 验证结果
	assert.NoError(t, err)

	// 验证mock调用
	mockAlertRepo.AssertExpectations(t)
	mockWorkflowManager.AssertExpectations(t)
	mockExecutionRepo.AssertExpectations(t)
}

// TestN8NAnalysisService_GetAnalysisStatus 测试获取分析状态
func TestN8NAnalysisService_GetAnalysisStatus(t *testing.T) {
	// 创建mock对象
	mockWorkflowManager := &MockWorkflowManager{}
	mockAlertRepo := &MockAlertRepository{}
	mockExecutionRepo := &MockExecutionRepository{}

	// 创建服务实例
	service := NewN8NAnalysisService(mockWorkflowManager, mockAlertRepo, mockExecutionRepo)

	ctx := context.Background()
	executionID := "exec-123"

	// 设置mock期望
	execution := &analysis.N8NWorkflowExecution{
		ID:         executionID,
		WorkflowID: "workflow-1",
		Status:     analysis.N8NWorkflowStatusCompleted,
		StartedAt:  time.Now(),
	}

	mockExecutionRepo.On("GetByID", ctx, executionID).Return(execution, nil)

	// 执行测试
	result, err := service.GetAnalysisStatus(ctx, executionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, executionID, result.ID)
	assert.Equal(t, analysis.N8NWorkflowStatusCompleted, result.Status)

	// 验证mock调用
	mockExecutionRepo.AssertExpectations(t)
}

// TestN8NAnalysisService_CancelAnalysis 测试取消分析
func TestN8NAnalysisService_CancelAnalysis(t *testing.T) {
	// 创建mock对象
	mockWorkflowManager := &MockWorkflowManager{}
	mockAlertRepo := &MockAlertRepository{}
	mockExecutionRepo := &MockExecutionRepository{}

	// 创建服务实例
	service := NewN8NAnalysisService(mockWorkflowManager, mockAlertRepo, mockExecutionRepo)

	ctx := context.Background()
	executionID := "exec-123"

	// 设置mock期望
	execution := &analysis.N8NWorkflowExecution{
		ID:         executionID,
		WorkflowID: "workflow-1",
		Status:     analysis.N8NWorkflowStatusRunning,
		StartedAt:  time.Now(),
	}

	updatedExecution := &analysis.N8NWorkflowExecution{
		ID:         executionID,
		WorkflowID: "workflow-1",
		Status:     analysis.N8NWorkflowStatusCanceled,
		StartedAt:  time.Now(),
	}

	mockExecutionRepo.On("GetByID", ctx, executionID).Return(execution, nil)
	mockExecutionRepo.On("Update", ctx, mock.AnythingOfType("*analysis.N8NWorkflowExecution")).Return(nil)

	// 执行测试
	err := service.CancelAnalysis(ctx, executionID)

	// 验证结果
	assert.NoError(t, err)

	// 验证mock调用
	mockExecutionRepo.AssertExpectations(t)

	// 验证状态更新
	mockExecutionRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(exec *analysis.N8NWorkflowExecution) bool {
		return exec.Status == analysis.N8NWorkflowStatusFailed
	}))

	_ = updatedExecution // 避免未使用变量警告
}

// BenchmarkN8NAnalysisService_AnalyzeAlert 性能测试
func BenchmarkN8NAnalysisService_AnalyzeAlert(b *testing.B) {
	// 创建mock对象
	mockWorkflowManager := &MockWorkflowManager{}
	mockAlertRepo := &MockAlertRepository{}
	mockExecutionRepo := &MockExecutionRepository{}

	// 创建服务实例
	service := NewN8NAnalysisService(mockWorkflowManager, mockAlertRepo, mockExecutionRepo)

	ctx := context.Background()
	alertID := uint(123)
	workflowTemplateID := "workflow-template-1"

	// 设置mock期望
	alert := &model.Alert{
		ID:    alertID,
		Title: "Test Alert",
		Level: "high",
	}

	execution := &analysis.N8NWorkflowExecution{
		ID:         "exec-123",
		WorkflowID: workflowTemplateID,
		Status:     analysis.N8NWorkflowStatusRunning,
		Metadata: map[string]interface{}{
			"alert_id": fmt.Sprintf("%d", alertID),
			"type":     "alert_analysis",
		},
	}

	// Mock所有需要的方法调用
	mockAlertRepo.On("GetByID", ctx, alertID).Return(alert, nil)
	mockExecutions := []*analysis.N8NWorkflowExecution{}
	mockExecutionRepo.On("ListByWorkflowID", ctx, workflowTemplateID, 1, 0).Return(mockExecutions, nil)
	mockWorkflowManager.On("TriggerAnalysisWorkflow", ctx, fmt.Sprintf("%d", alertID), workflowTemplateID, mock.Anything).Return(execution, nil)

	// 重置计时器
	b.ResetTimer()

	// 执行基准测试
	for i := 0; i < b.N; i++ {
		_, err := service.AnalyzeAlert(ctx, alertID, workflowTemplateID)
		if err != nil {
			b.Fatal(err)
		}
	}
}