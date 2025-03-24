package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// OllamaService Ollama服务
type OllamaService struct {
	config *config.OllamaConfig
	client *http.Client
	logger *zap.Logger
}

// NewOllamaService 创建Ollama服务
func NewOllamaService() *OllamaService {
	// 创建默认配置
	config := &config.OllamaConfig{
		APIEndpoint: "http://localhost:11434", // 默认地址
		Model:       "llama2",                 // 默认模型
		Timeout:     300,                      // 默认30秒超时
		MaxRetries:  3,                        // 默认最大重试3次
	}

	// 获取全局logger实例
	logger := logger.L
	if logger == nil {
		logger = zap.L()
	}

	// 尝试从数据库获取配置
	if database.DB != nil {
		var settings model.Settings
		if err := database.DB.Order("updated_at desc").First(&settings).Error; err == nil {
			config.APIEndpoint = settings.OllamaEndpoint
			config.Model = settings.OllamaModel
			logger.Info("loaded settings from database",
				zap.String("endpoint", config.APIEndpoint),
				zap.String("model", config.Model))
		} else {
			logger.Warn("failed to get settings from database, using default config", zap.Error(err))
		}
	} else {
		logger.Warn("database connection not initialized, using default config")
	}

	return &OllamaService{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// AnalyzeAlert 分析告警
func (s *OllamaService) AnalyzeAlert(ctx context.Context, alert *model.Alert) (string, error) {
	// 构建提示词
	prompt := fmt.Sprintf(`请分析以下告警信息，并提供详细的分析和建议：

告警名称：%s
告警级别：%s
告警来源：%s
告警内容：%s

请从以下几个方面进行分析：
1. 告警的严重程度和影响范围
2. 可能的原因分析
3. 建议的处理方案
4. 预防措施建议

请用中文回答，并保持专业和客观。`, alert.Name, alert.Level, alert.Source, alert.Content)

	// 调用Ollama API
	analysis, err := s.callOllamaAPI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to analyze alert: %w", err)
	}

	return analysis, nil
}

// FindSimilarAlerts 查找相似告警
func (s *OllamaService) FindSimilarAlerts(ctx context.Context, alert *model.Alert) ([]*model.Alert, error) {
	// 构建提示词
	prompt := fmt.Sprintf(`请根据以下告警信息，查找相似的告警：

告警名称：%s
告警级别：%s
告警来源：%s
告警内容：%s

请从数据库中查找相似的告警，并返回告警ID列表。`, alert.Name, alert.Level, alert.Source, alert.Content)

	// 调用Ollama API
	similarIDs, err := s.callOllamaAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar alerts: %w", err)
	}

	// 从数据库获取相似告警
	var similarAlerts []*model.Alert
	if err := database.DB.Where("id IN ?", similarIDs).Find(&similarAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get similar alerts from database: %w", err)
	}

	return similarAlerts, nil
}

// callOllamaAPI 调用Ollama API
func (s *OllamaService) callOllamaAPI(ctx context.Context, prompt string) (string, error) {
	// 构建请求体
	reqBody := map[string]interface{}{
		"model":  s.config.Model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建请求
	fmt.Println("jsonData", string(jsonData))
	fmt.Println("s.config.APIEndpoint", s.config.APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.APIEndpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var result struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("ollama API error: %s", result.Error)
	}

	return result.Response, nil
}
