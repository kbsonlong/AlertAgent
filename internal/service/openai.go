package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"alert_agent/internal/model"
)

// SimilarAlert 相似告警结构体
type SimilarAlert struct {
	Alert      model.Alert `json:"alert"`
	Similarity float64     `json:"similarity"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	Endpoint   string
	Model      string
	Timeout    int
	MaxRetries int
}

// OpenAIService OpenAI服务
type OpenAIService struct {
	endpoint   string
	model      string
	timeout    time.Duration
	maxRetries int
	client     *http.Client
	logger     *zap.Logger
	// healthService *HealthService
}

// Message 消息结构体
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaRequest Ollama请求结构
type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system,omitempty"`
	Stream   bool   `json:"stream"`
	Template string `json:"template,omitempty"`
	Context  []int  `json:"context,omitempty"`
}

// OllamaResponse Ollama响应结构
type OllamaResponse struct {
	Model         string `json:"model"`
	Response      string `json:"response"`
	Done          bool   `json:"done"`
	Context       []int  `json:"context,omitempty"`
	TotalDuration int64  `json:"total_duration,omitempty"`
	LoadDuration  int64  `json:"load_duration,omitempty"`
	Error         string `json:"error,omitempty"`
}

// NewOpenAIService 创建OpenAI服务实例
func NewOpenAIService(cfg *OpenAIConfig) *OpenAIService {
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	// // 创建健康检查服务
	// healthService := NewHealthService(cfg.Endpoint, cfg.Timeout)

	// // 启动健康检查，每30秒检查一次
	// go healthService.StartHealthCheck(context.Background(), 30*time.Second)

	return &OpenAIService{
		endpoint:   cfg.Endpoint,
		model:      cfg.Model,
		timeout:    time.Duration(cfg.Timeout) * time.Second,
		maxRetries: cfg.MaxRetries,
		client:     client,
		logger:     zap.L(),
		// healthService: healthService,
	}
}

// generateResponse 生成回复
func (s *OpenAIService) generateResponse(ctx context.Context, messages []Message) (string, error) {
	// 提取系统提示和用户消息
	var systemPrompt, userPrompt string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else if msg.Role == "user" {
			userPrompt = msg.Content
		}
	}

	reqBody := OllamaRequest{
		Model:  s.model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: false,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Error("Failed to marshal request",
			zap.Error(err),
			zap.Any("request", reqBody),
		)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	s.logger.Debug("Sending request to Ollama",
		zap.String("endpoint", s.endpoint),
		zap.String("model", s.model),
		zap.Any("request", reqBody),
	)

	// 构建完整的URL
	url := fmt.Sprintf("%s/api/generate", s.endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		s.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("endpoint", url),
		)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 记录请求开始时间
	requestStart := time.Now()
	s.logger.Debug("Executing HTTP request",
		zap.String("url", url),
		zap.Duration("timeout", s.timeout),
	)

	// 执行请求
	resp, err := s.client.Do(req)

	// 检查是否是上下文取消或超时
	if ctx.Err() != nil {
		s.logger.Warn("Request canceled or timed out",
			zap.Error(ctx.Err()),
			zap.Duration("elapsed", time.Since(requestStart)),
			zap.String("endpoint", url),
		)
		return "", fmt.Errorf("request canceled or timed out: %w", ctx.Err())
	}

	// 检查其他错误
	if err != nil {
		s.logger.Error("Failed to send request",
			zap.Error(err),
			zap.String("endpoint", url),
			zap.Duration("elapsed", time.Since(requestStart)),
		)

		// 检查是否是超时错误
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("request timed out after %v: %w", s.timeout, err)
		}

		// 检查是否是连接错误
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no route to host") {
			return "", fmt.Errorf("connection to Ollama server failed: %w", err)
		}

		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	s.logger.Debug("Received response from Ollama",
		zap.Int("status", resp.StatusCode),
		zap.String("response", string(body)),
	)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("API request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		s.logger.Error("Failed to unmarshal response",
			zap.Error(err),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if ollamaResp.Error != "" {
		s.logger.Error("Ollama returned error",
			zap.String("error", ollamaResp.Error),
		)
		return "", fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// AnalyzeAlert 分析告警
func (s *OpenAIService) AnalyzeAlert(ctx context.Context, alert *model.Alert) (string, error) {
	s.logger.Info("Starting alert analysis",
		zap.Uint("alert_id", alert.ID),
		zap.String("alert_name", alert.Name),
		zap.String("alert_level", alert.Level),
	)

	// 检查Ollama服务健康状态
	// if !s.healthService.IsHealthy() {
	// 	s.logger.Warn("Ollama service is not healthy, skipping analysis",
	// 		zap.Uint("alert_id", alert.ID),
	// 		zap.String("endpoint", s.endpoint),
	// 	)
	// 	return "", fmt.Errorf("Ollama service is not available")
	// }

	systemPrompt := `你是一个专业的告警分析专家。请分析以下告警信息，并提供简要的分析结果，包括：
1. 告警的严重程度
2. 可能的原因
3. 建议的解决方案`

	alertContent := fmt.Sprintf("告警名称：%s\n告警级别：%s\n告警信息：%s\n告警时间：%s",
		alert.Name,
		alert.Level,
		alert.Content,
		alert.CreatedAt.Format("2006-01-02 15:04:05"))

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: alertContent,
		},
	}

	s.logger.Debug("Analyzing alert",
		zap.Any("alert", alert),
		zap.Any("messages", messages),
	)

	// 创建一个带有超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 添加重试机制
	var response string
	var err error
	for i := 0; i <= s.maxRetries; i++ {
		// 检查上下文是否已取消
		if timeoutCtx.Err() != nil {
			return "", fmt.Errorf("analysis canceled or timed out: %w", timeoutCtx.Err())
		}

		// 调用生成响应的方法
		response, err = s.generateResponse(timeoutCtx, messages)
		if err == nil {
			break // 成功获取响应，跳出重试循环
		}

		// 记录重试信息
		if i < s.maxRetries {
			s.logger.Warn("Retrying alert analysis",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("max_retries", s.maxRetries),
				zap.Uint("alert_id", alert.ID),
			)
			// 简单的退避策略，每次重试等待时间增加
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}

	// 检查最终结果
	if err != nil {
		s.logger.Error("Alert analysis failed after retries",
			zap.Error(err),
			zap.Int("max_retries", s.maxRetries),
			zap.Uint("alert_id", alert.ID),
		)
		return "", fmt.Errorf("failed to analyze alert after %d retries: %w", s.maxRetries, err)
	}

	s.logger.Info("Alert analysis completed successfully",
		zap.Uint("alert_id", alert.ID),
		zap.Int("response_length", len(response)),
	)

	return response, nil
}

// FindSimilarAlerts 查找相似告警
func (s *OpenAIService) FindSimilarAlerts(ctx context.Context, alert *model.Alert, alerts []*model.Alert) ([]SimilarAlert, error) {
	systemPrompt := `你是一个专业的告警分析专家。请分析给定的告警与历史告警的相似度，并返回相似度分数（0-100的数字）。
请只返回数字，不要返回其他内容。如果无法比较相似度，请返回0。`

	var result []SimilarAlert
	for _, historicalAlert := range alerts {
		content := fmt.Sprintf("告警1：\n名称：%s\n级别：%s\n信息：%s\n\n告警2：\n名称：%s\n级别：%s\n信息：%s",
			alert.Name, alert.Level, alert.Content,
			historicalAlert.Name, historicalAlert.Level, historicalAlert.Content)

		messages := []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: content,
			},
		}

		response, err := s.generateResponse(ctx, messages)
		if err != nil {
			s.logger.Error("Failed to get similarity score",
				zap.Error(err),
				zap.Any("alert", alert),
				zap.Any("historical_alert", historicalAlert),
			)
			continue
		}

		score, err := strconv.ParseFloat(strings.TrimSpace(response), 64)
		if err != nil {
			s.logger.Error("Failed to parse similarity score",
				zap.Error(err),
				zap.String("response", response),
			)
			continue
		}

		if score > 0 {
			result = append(result, SimilarAlert{
				Alert:      *historicalAlert,
				Similarity: score,
			})
		}
	}

	// 按相似度降序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Similarity > result[j].Similarity
	})

	return result, nil
}
