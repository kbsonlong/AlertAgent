package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// DifyConfig Dify配置 (local copy to avoid circular dependency)
type DifyConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
	APIKey      string `yaml:"api_key"`
	Model       string `yaml:"model"`
	Timeout     int    `yaml:"timeout"`
	MaxRetries  int    `yaml:"max_retries"`
}

// DifyService Dify AI服务
type DifyService struct {
	config     *DifyConfig
	client     *http.Client
	logger     *zap.Logger
	mutex      sync.RWMutex
}



// DifyChatRequest Dify聊天请求
type DifyChatRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	Query        string                 `json:"query"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
	ConversationID string               `json:"conversation_id,omitempty"`
}

// DifyChatResponse Dify聊天响应
type DifyChatResponse struct {
	Event          string                 `json:"event"`
	MessageID      string                 `json:"message_id"`
	ConversationID string                 `json:"conversation_id"`
	Mode           string                 `json:"mode"`
	Answer         string                 `json:"answer"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      int64                  `json:"created_at"`
}

// DifyWorkflowRequest Dify工作流请求
type DifyWorkflowRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	ResponseMode   string                 `json:"response_mode"`
	User           string                 `json:"user"`
}

// DifyWorkflowResponse Dify工作流响应
type DifyWorkflowResponse struct {
	WorkflowRunID string                 `json:"workflow_run_id"`
	TaskID        string                 `json:"task_id"`
	Data          map[string]interface{} `json:"data"`
	Status        string                 `json:"status"`
	Error         string                 `json:"error,omitempty"`
}

// NewDifyService 创建Dify服务
func NewDifyService() *DifyService {
	// 从全局配置获取Dify配置
	// cfg := config.GetConfig()
	
	// 创建默认Dify配置（如果配置文件中没有）
	difyConfig := &DifyConfig{
		Enabled:     false,
		APIEndpoint: "http://localhost:5001",
		APIKey:      "",
		Timeout:     30,
		Model:       "gpt-3.5-turbo",
		MaxRetries:  3,
	}

	// 如果配置中有Dify配置，使用配置的值
	// 这里假设配置结构中有Dify字段，如果没有需要添加到config包中

	logger := logger.L
	if logger == nil {
		logger = zap.L()
	}

	service := &DifyService{
		config: difyConfig,
		client: &http.Client{
			Timeout: time.Duration(difyConfig.Timeout) * time.Second,
		},
		logger: logger,
	}

	return service
}

// UpdateConfig 更新配置
func (s *DifyService) UpdateConfig(newConfig *DifyConfig) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config = newConfig
	s.client = &http.Client{
		Timeout: time.Duration(newConfig.Timeout) * time.Second,
	}

	s.logger.Info("Dify configuration updated",
		zap.Bool("enabled", newConfig.Enabled),
		zap.String("api_endpoint", newConfig.APIEndpoint),
		zap.String("model", newConfig.Model),
		zap.Int("timeout", newConfig.Timeout))
}

// getConfig 安全地获取当前配置
func (s *DifyService) getConfig() *DifyConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config
}

// AnalyzeAlert 分析告警
func (s *DifyService) AnalyzeAlert(ctx context.Context, alert *model.Alert) (string, error) {
	currentConfig := s.getConfig()
	
	if !currentConfig.Enabled {
		return "", fmt.Errorf("dify analysis is disabled")
	}

	if currentConfig.APIKey == "" {
		return "", fmt.Errorf("dify API key is not configured")
	}

	// 构建分析提示词
	query := s.buildAnalysisQuery(alert)
	
	// 构建输入数据
	inputs := map[string]interface{}{
		"alert_title":   alert.Title,
		"alert_level":   alert.Level,
		"alert_source":  alert.Source,
		"alert_content": alert.Content,
		"alert_labels":  alert.Labels,
		"alert_time":    alert.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 调用Dify Chat API
	response, err := s.callChatAPI(ctx, query, inputs)
	if err != nil {
		return "", fmt.Errorf("failed to call dify chat API: %w", err)
	}

	return response.Answer, nil
}

// AnalyzeAlertWithWorkflow 使用工作流分析告警
func (s *DifyService) AnalyzeAlertWithWorkflow(ctx context.Context, alert *model.Alert) (map[string]interface{}, error) {
	currentConfig := s.getConfig()
	
	if !currentConfig.Enabled {
		return nil, fmt.Errorf("dify analysis is disabled")
	}

	if currentConfig.APIKey == "" {
		return nil, fmt.Errorf("dify API key is not configured")
	}

	// 构建工作流输入
	inputs := map[string]interface{}{
		"alert_id":      alert.ID,
		"alert_title":   alert.Title,
		"alert_level":   alert.Level,
		"alert_source":  alert.Source,
		"alert_content": alert.Content,
		"alert_labels":  alert.Labels,
		"alert_time":    alert.CreatedAt.Format("2006-01-02 15:04:05"),
		"analysis_type": "comprehensive",
	}

	// 调用Dify Workflow API
	response, err := s.callWorkflowAPI(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to call dify workflow API: %w", err)
	}

	return response.Data, nil
}

// FindSimilarAlerts 查找相似告警
func (s *DifyService) FindSimilarAlerts(ctx context.Context, alert *model.Alert) ([]string, error) {
	currentConfig := s.getConfig()
	
	if !currentConfig.Enabled {
		return nil, fmt.Errorf("dify analysis is disabled")
	}

	query := fmt.Sprintf(`请根据以下告警信息，从历史告警中查找相似的告警：

告警标题：%s
告警级别：%s
告警来源：%s
告警内容：%s

请返回相似告警的ID列表，格式为JSON数组。`, 
		alert.Title, alert.Level, alert.Source, alert.Content)

	inputs := map[string]interface{}{
		"alert_title":   alert.Title,
		"alert_level":   alert.Level,
		"alert_source":  alert.Source,
		"alert_content": alert.Content,
		"task_type":     "find_similar",
	}

	response, err := s.callChatAPI(ctx, query, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar alerts: %w", err)
	}

	// 解析响应中的告警ID列表
	var alertIDs []string
	if err := json.Unmarshal([]byte(response.Answer), &alertIDs); err != nil {
		// 如果解析失败，返回空列表
		s.logger.Warn("Failed to parse similar alert IDs", 
			zap.String("response", response.Answer),
			zap.Error(err))
		return []string{}, nil
	}

	return alertIDs, nil
}

// GenerateActionPlan 生成处理方案
func (s *DifyService) GenerateActionPlan(ctx context.Context, alert *model.Alert, analysisResult string) (string, error) {
	currentConfig := s.getConfig()
	
	if !currentConfig.Enabled {
		return "", fmt.Errorf("dify analysis is disabled")
	}

	query := fmt.Sprintf(`基于以下告警信息和分析结果，生成详细的处理方案：

告警信息：
- 标题：%s
- 级别：%s
- 来源：%s
- 内容：%s

分析结果：
%s

请生成包含以下内容的处理方案：
1. 紧急处理步骤
2. 根本原因分析
3. 长期解决方案
4. 预防措施建议

请用中文回答，并保持专业和实用。`, 
		alert.Title, alert.Level, alert.Source, alert.Content, analysisResult)

	inputs := map[string]interface{}{
		"alert_title":     alert.Title,
		"alert_level":     alert.Level,
		"alert_source":    alert.Source,
		"alert_content":   alert.Content,
		"analysis_result": analysisResult,
		"task_type":       "action_plan",
	}

	response, err := s.callChatAPI(ctx, query, inputs)
	if err != nil {
		return "", fmt.Errorf("failed to generate action plan: %w", err)
	}

	return response.Answer, nil
}

// buildAnalysisQuery 构建分析查询
func (s *DifyService) buildAnalysisQuery(alert *model.Alert) string {
	return fmt.Sprintf(`请分析以下告警信息，并提供详细的分析和建议：

告警标题：%s
告警级别：%s
告警来源：%s
告警内容：%s
告警时间：%s

请从以下几个方面进行分析：
1. 告警的严重程度和影响范围
2. 可能的原因分析
3. 建议的处理方案
4. 预防措施建议

请用中文回答，并保持专业和客观。`, 
		alert.Title, 
		alert.Level, 
		alert.Source, 
		alert.Content,
		alert.CreatedAt.Format("2006-01-02 15:04:05"))
}

// callChatAPI 调用Dify Chat API
func (s *DifyService) callChatAPI(ctx context.Context, query string, inputs map[string]interface{}) (*DifyChatResponse, error) {
	currentConfig := s.getConfig()
	
	request := &DifyChatRequest{
		Inputs:       inputs,
		Query:        query,
		ResponseMode: "blocking",
		User:         "alertagent-worker",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat-messages", currentConfig.APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", currentConfig.APIKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dify API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response DifyChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// callWorkflowAPI 调用Dify Workflow API
func (s *DifyService) callWorkflowAPI(ctx context.Context, inputs map[string]interface{}) (*DifyWorkflowResponse, error) {
	currentConfig := s.getConfig()
	
	request := &DifyWorkflowRequest{
		Inputs:       inputs,
		ResponseMode: "blocking",
		User:         "alertagent-worker",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/workflows/run", currentConfig.APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", currentConfig.APIKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dify API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response DifyWorkflowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// HealthCheck 健康检查
func (s *DifyService) HealthCheck(ctx context.Context) error {
	currentConfig := s.getConfig()
	
	if !currentConfig.Enabled {
		return fmt.Errorf("dify service is disabled")
	}

	if currentConfig.APIKey == "" {
		return fmt.Errorf("dify API key is not configured")
	}

	// 发送简单的测试请求
	testInputs := map[string]interface{}{
		"test": "health_check",
	}

	_, err := s.callChatAPI(ctx, "健康检查测试", testInputs)
	if err != nil {
		return fmt.Errorf("dify health check failed: %w", err)
	}

	return nil
}