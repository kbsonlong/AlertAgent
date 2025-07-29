package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"alert_agent/internal/domain/analysis"
	"go.uber.org/zap"
)

// HTTPClientConfig HTTP 客户端配置
type HTTPClientConfig struct {
	BaseURL        string        `json:"base_url"`
	APIKey         string        `json:"api_key"`
	Timeout        time.Duration `json:"timeout"`
	RetryAttempts  int           `json:"retry_attempts"`
	RetryDelay     time.Duration `json:"retry_delay"`
	MaxRetryDelay  time.Duration `json:"max_retry_delay"`
	CallbackURL    string        `json:"callback_url"`
	CallbackSecret string        `json:"callback_secret"`
}

// HTTPClient n8n HTTP 客户端实现
type HTTPClient struct {
	config    *HTTPClientConfig
	httpClient *http.Client
	logger    *zap.Logger
	callbacks map[string]func(ctx context.Context, data map[string]interface{}) error
	callbackMu sync.RWMutex
}

// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient(config *HTTPClientConfig, logger *zap.Logger) *HTTPClient {
	return &HTTPClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:    logger,
		callbacks: make(map[string]func(ctx context.Context, data map[string]interface{}) error),
	}
}

// TriggerWorkflow 触发工作流
func (c *HTTPClient) TriggerWorkflow(ctx context.Context, req *analysis.N8NWorkflowTriggerRequest) (*analysis.N8NWorkflowExecution, error) {
	url := fmt.Sprintf("%s/api/v1/workflows/%s/execute", c.config.BaseURL, req.WorkflowID)
	
	// 准备请求数据
	requestData := map[string]interface{}{
		"data": req.InputData,
	}
	
	if req.Metadata != nil {
		requestData["metadata"] = req.Metadata
	}
	
	// 如果有回调配置，添加回调信息
	if req.Callback != nil {
		requestData["callback"] = map[string]interface{}{
			"url":     req.Callback.URL,
			"method":  req.Callback.Method,
			"headers": req.Callback.Headers,
			"secret":  req.Callback.Secret,
		}
	}
	
	body, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data: %w", err)
	}
	
	resp, err := c.doRequestWithRetry(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("trigger workflow: %w", err)
	}
	
	var execution analysis.N8NWorkflowExecution
	if err := json.Unmarshal(resp, &execution); err != nil {
		return nil, fmt.Errorf("unmarshal execution response: %w", err)
	}
	
	c.logger.Info("workflow triggered successfully",
		zap.String("workflow_id", req.WorkflowID),
		zap.String("execution_id", execution.ID),
		zap.String("status", string(execution.Status)))
	
	return &execution, nil
}

// GetWorkflowExecution 获取工作流执行状态
func (c *HTTPClient) GetWorkflowExecution(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	url := fmt.Sprintf("%s/api/v1/executions/%s", c.config.BaseURL, executionID)
	
	resp, err := c.doRequestWithRetry(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("get workflow execution: %w", err)
	}
	
	var execution analysis.N8NWorkflowExecution
	if err := json.Unmarshal(resp, &execution); err != nil {
		return nil, fmt.Errorf("unmarshal execution response: %w", err)
	}
	
	return &execution, nil
}

// CancelWorkflowExecution 取消工作流执行
func (c *HTTPClient) CancelWorkflowExecution(ctx context.Context, executionID string) error {
	url := fmt.Sprintf("%s/api/v1/executions/%s/cancel", c.config.BaseURL, executionID)
	
	_, err := c.doRequestWithRetry(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("cancel workflow execution: %w", err)
	}
	
	c.logger.Info("workflow execution canceled", zap.String("execution_id", executionID))
	return nil
}

// RetryWorkflowExecution 重试工作流执行
func (c *HTTPClient) RetryWorkflowExecution(ctx context.Context, executionID string) (*analysis.N8NWorkflowExecution, error) {
	url := fmt.Sprintf("%s/api/v1/executions/%s/retry", c.config.BaseURL, executionID)
	
	resp, err := c.doRequestWithRetry(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("retry workflow execution: %w", err)
	}
	
	var execution analysis.N8NWorkflowExecution
	if err := json.Unmarshal(resp, &execution); err != nil {
		return nil, fmt.Errorf("unmarshal execution response: %w", err)
	}
	
	c.logger.Info("workflow execution retried",
		zap.String("execution_id", executionID),
		zap.String("new_execution_id", execution.ID))
	
	return &execution, nil
}

// ListWorkflowExecutions 列出工作流执行历史
func (c *HTTPClient) ListWorkflowExecutions(ctx context.Context, workflowID string, limit, offset int) ([]*analysis.N8NWorkflowExecution, error) {
	params := url.Values{}
	params.Set("workflowId", workflowID)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	
	url := fmt.Sprintf("%s/api/v1/executions?%s", c.config.BaseURL, params.Encode())
	
	resp, err := c.doRequestWithRetry(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("list workflow executions: %w", err)
	}
	
	var response struct {
		Data []*analysis.N8NWorkflowExecution `json:"data"`
	}
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("unmarshal executions response: %w", err)
	}
	
	return response.Data, nil
}

// GetWorkflowTemplate 获取工作流模板
func (c *HTTPClient) GetWorkflowTemplate(ctx context.Context, workflowID string) (*analysis.N8NWorkflowTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/workflows/%s", c.config.BaseURL, workflowID)
	
	resp, err := c.doRequestWithRetry(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("get workflow template: %w", err)
	}
	
	var template analysis.N8NWorkflowTemplate
	if err := json.Unmarshal(resp, &template); err != nil {
		return nil, fmt.Errorf("unmarshal template response: %w", err)
	}
	
	return &template, nil
}

// CreateWorkflowTemplate 创建工作流模板
func (c *HTTPClient) CreateWorkflowTemplate(ctx context.Context, template *analysis.N8NWorkflowTemplate) (*analysis.N8NWorkflowTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/workflows", c.config.BaseURL)
	
	body, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("marshal template: %w", err)
	}
	
	resp, err := c.doRequestWithRetry(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create workflow template: %w", err)
	}
	
	var createdTemplate analysis.N8NWorkflowTemplate
	if err := json.Unmarshal(resp, &createdTemplate); err != nil {
		return nil, fmt.Errorf("unmarshal created template response: %w", err)
	}
	
	c.logger.Info("workflow template created",
		zap.String("template_id", createdTemplate.ID),
		zap.String("template_name", createdTemplate.Name))
	
	return &createdTemplate, nil
}

// UpdateWorkflowTemplate 更新工作流模板
func (c *HTTPClient) UpdateWorkflowTemplate(ctx context.Context, workflowID string, template *analysis.N8NWorkflowTemplate) (*analysis.N8NWorkflowTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/workflows/%s", c.config.BaseURL, workflowID)
	
	body, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("marshal template: %w", err)
	}
	
	resp, err := c.doRequestWithRetry(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("update workflow template: %w", err)
	}
	
	var updatedTemplate analysis.N8NWorkflowTemplate
	if err := json.Unmarshal(resp, &updatedTemplate); err != nil {
		return nil, fmt.Errorf("unmarshal updated template response: %w", err)
	}
	
	c.logger.Info("workflow template updated",
		zap.String("template_id", updatedTemplate.ID),
		zap.String("template_name", updatedTemplate.Name))
	
	return &updatedTemplate, nil
}

// DeleteWorkflowTemplate 删除工作流模板
func (c *HTTPClient) DeleteWorkflowTemplate(ctx context.Context, workflowID string) error {
	url := fmt.Sprintf("%s/api/v1/workflows/%s", c.config.BaseURL, workflowID)
	
	_, err := c.doRequestWithRetry(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("delete workflow template: %w", err)
	}
	
	c.logger.Info("workflow template deleted", zap.String("template_id", workflowID))
	return nil
}

// ListWorkflowTemplates 列出工作流模板
func (c *HTTPClient) ListWorkflowTemplates(ctx context.Context, limit, offset int) ([]*analysis.N8NWorkflowTemplate, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	
	url := fmt.Sprintf("%s/api/v1/workflows?%s", c.config.BaseURL, params.Encode())
	
	resp, err := c.doRequestWithRetry(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("list workflow templates: %w", err)
	}
	
	var response struct {
		Data []*analysis.N8NWorkflowTemplate `json:"data"`
	}
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("unmarshal templates response: %w", err)
	}
	
	return response.Data, nil
}

// GetHealth 获取 n8n 健康状态
func (c *HTTPClient) GetHealth(ctx context.Context) (*analysis.N8NHealthStatus, error) {
	url := fmt.Sprintf("%s/api/v1/health", c.config.BaseURL)
	
	resp, err := c.doRequestWithRetry(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("get health status: %w", err)
	}
	
	var health analysis.N8NHealthStatus
	if err := json.Unmarshal(resp, &health); err != nil {
		return nil, fmt.Errorf("unmarshal health response: %w", err)
	}
	
	return &health, nil
}

// RegisterCallback 注册回调处理器
func (c *HTTPClient) RegisterCallback(ctx context.Context, callbackID string, handler func(ctx context.Context, data map[string]interface{}) error) error {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	
	c.callbacks[callbackID] = handler
	c.logger.Info("callback handler registered", zap.String("callback_id", callbackID))
	return nil
}

// UnregisterCallback 注销回调处理器
func (c *HTTPClient) UnregisterCallback(ctx context.Context, callbackID string) error {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	
	delete(c.callbacks, callbackID)
	c.logger.Info("callback handler unregistered", zap.String("callback_id", callbackID))
	return nil
}

// HandleCallback 处理回调请求
func (c *HTTPClient) HandleCallback(ctx context.Context, callbackID string, data map[string]interface{}) error {
	c.callbackMu.RLock()
	handler, exists := c.callbacks[callbackID]
	c.callbackMu.RUnlock()
	
	if !exists {
		return fmt.Errorf("callback handler not found: %s", callbackID)
	}
	
	return handler(ctx, data)
}

// doRequestWithRetry 执行带重试的 HTTP 请求
func (c *HTTPClient) doRequestWithRetry(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// 计算退避延迟
			delay := time.Duration(attempt) * c.config.RetryDelay
			if delay > c.config.MaxRetryDelay {
				delay = c.config.MaxRetryDelay
			}
			
			c.logger.Warn("retrying request",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.Error(lastErr))
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		
		resp, err := c.doRequest(ctx, method, url, body)
		if err != nil {
			lastErr = err
			continue
		}
		
		return resp, nil
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.RetryAttempts, lastErr)
}

// doRequest 执行单次 HTTP 请求
func (c *HTTPClient) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}