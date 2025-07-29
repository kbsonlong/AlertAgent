package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"alert_agent/internal/domain/analysis"
	"go.uber.org/zap"
)

// DifyClientImpl Dify 客户端实现
type DifyClientImpl struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewDifyClient 创建新的 Dify 客户端
func NewDifyClient(baseURL, apiKey string, logger *zap.Logger) analysis.DifyClient {
	return &DifyClientImpl{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ChatMessage 发送聊天消息
func (c *DifyClientImpl) ChatMessage(ctx context.Context, request *analysis.DifyChatRequest) (*analysis.DifyChatResponse, error) {
	url := fmt.Sprintf("%s/v1/chat-messages", c.baseURL)
	
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	resp, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response analysis.DifyChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &response, nil
}

// RunWorkflow 运行工作流
func (c *DifyClientImpl) RunWorkflow(ctx context.Context, request *analysis.DifyWorkflowRequest) (*analysis.DifyWorkflowResponse, error) {
	url := fmt.Sprintf("%s/v1/workflows/run", c.baseURL)
	
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	resp, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response analysis.DifyWorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &response, nil
}

// GetConversation 获取对话信息
func (c *DifyClientImpl) GetConversation(ctx context.Context, conversationID string) (*analysis.DifyConversation, error) {
	url := fmt.Sprintf("%s/v1/conversations/%s", c.baseURL, conversationID)
	
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var conversation analysis.DifyConversation
	if err := json.NewDecoder(resp.Body).Decode(&conversation); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &conversation, nil
}

// HealthCheck 健康检查
func (c *DifyClientImpl) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/parameters", c.baseURL)
	
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// GetModelInfo 获取模型信息
func (c *DifyClientImpl) GetModelInfo(ctx context.Context) (*analysis.DifyModelInfo, error) {
	url := fmt.Sprintf("%s/v1/parameters", c.baseURL)
	
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var modelInfo analysis.DifyModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&modelInfo); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &modelInfo, nil
}

// GetWorkflowStatus 获取工作流状态
func (c *DifyClientImpl) GetWorkflowStatus(ctx context.Context, workflowRunID string) (*analysis.DifyWorkflowResponse, error) {
	url := fmt.Sprintf("%s/v1/workflows/run/%s", c.baseURL, workflowRunID)
	
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var status analysis.DifyWorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &status, nil
}

// SearchKnowledge 搜索知识库
func (c *DifyClientImpl) SearchKnowledge(ctx context.Context, query string, datasetIDs []string) (*analysis.KnowledgeSearchResult, error) {
	url := fmt.Sprintf("%s/v1/datasets/search", c.baseURL)
	
	request := map[string]interface{}{
		"query": query,
		"dataset_ids": datasetIDs,
	}
	
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}
	
	resp, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response analysis.KnowledgeSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	
	return &response, nil
}

// CancelWorkflow 取消工作流
func (c *DifyClientImpl) CancelWorkflow(ctx context.Context, workflowRunID string) error {
	url := fmt.Sprintf("%s/v1/workflows/run/%s/stop", c.baseURL, workflowRunID)
	
	resp, err := c.doRequest(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel workflow failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// doRequest 执行 HTTP 请求
func (c *DifyClientImpl) doRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlertAgent/1.0")
	
	// 记录请求日志
	c.logger.Debug("Dify API request",
		zap.String("method", method),
		zap.String("url", url),
		zap.Bool("has_body", body != nil),
	)
	
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		c.logger.Error("Dify API request failed",
			zap.String("method", method),
			zap.String("url", url),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	// 记录响应日志
	c.logger.Debug("Dify API response",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", duration),
	)
	
	// 检查响应状态
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		c.logger.Error("Dify API error response",
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	return resp, nil
}

// SetTimeout 设置超时时间
func (c *DifyClientImpl) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetHTTPClient 设置自定义 HTTP 客户端
func (c *DifyClientImpl) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}