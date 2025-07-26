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
	mutex  sync.RWMutex
}

// NewOllamaService 创建Ollama服务
func NewOllamaService() *OllamaService {
	// 使用全局配置
	cfg := config.GetConfig()
	ollamaConfig := &cfg.Ollama

	// 获取全局logger实例
	logger := logger.L
	if logger == nil {
		logger = zap.L()
	}

	service := &OllamaService{
		config: ollamaConfig,
		client: &http.Client{
			Timeout: time.Duration(ollamaConfig.Timeout) * time.Second,
		},
		logger: logger,
	}

	// 注册配置更新回调
	config.RegisterReloadCallback(service.onConfigReload)

	return service
}

// onConfigReload 配置重载回调
func (s *OllamaService) onConfigReload(newConfig config.Config) {
	s.updateConfig(&newConfig.Ollama)
}

// updateConfig 更新配置
func (s *OllamaService) updateConfig(newConfig *config.OllamaConfig) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 更新配置
	s.config = newConfig

	// 重新创建HTTP客户端以应用新的超时设置
	s.client = &http.Client{
		Timeout: time.Duration(newConfig.Timeout) * time.Second,
	}

	s.logger.Info("Ollama configuration updated",
		zap.Bool("enabled", newConfig.Enabled),
		zap.String("endpoint", newConfig.APIEndpoint),
		zap.String("model", newConfig.Model),
		zap.Int("timeout", newConfig.Timeout))
}

// getConfig 安全地获取当前配置
func (s *OllamaService) getConfig() *config.OllamaConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config
}

// AnalyzeAlert 分析告警
func (s *OllamaService) AnalyzeAlert(ctx context.Context, alert *model.Alert) (string, error) {
	// 获取当前配置
	currentConfig := s.getConfig()
	
	// 检查是否启用Ollama功能
	if !currentConfig.Enabled {
		return "", fmt.Errorf("ollama analysis is disabled")
	}

	// 构建提示词
	prompt := fmt.Sprintf(`请分析以下告警信息，并提供详细的分析和建议：

告警标题：%s
告警级别：%s
告警来源：%s
告警内容：%s

请从以下几个方面进行分析：
1. 告警的严重程度和影响范围
2. 可能的原因分析
3. 建议的处理方案
4. 预防措施建议

请用中文回答，并保持专业和客观。`, alert.Title, alert.Level, alert.Source, alert.Content)

	// 调用Ollama API
	analysis, err := s.callOllamaAPI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to analyze alert: %w", err)
	}

	return analysis, nil
}

// FindSimilarAlerts 查找相似告警
func (s *OllamaService) FindSimilarAlerts(ctx context.Context, alert *model.Alert) ([]*model.Alert, error) {
	// 获取当前配置
	currentConfig := s.getConfig()
	
	// 检查是否启用Ollama功能
	if !currentConfig.Enabled {
		return nil, fmt.Errorf("ollama analysis is disabled")
	}

	// 构建提示词
	prompt := fmt.Sprintf(`请根据以下告警信息，查找相似的告警：

告警标题：%s
告警级别：%s
告警来源：%s
告警内容：%s

请从数据库中查找相似的告警，并返回告警ID列表。`, alert.Title, alert.Level, alert.Source, alert.Content)

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
	// 获取当前配置
	currentConfig := s.getConfig()
	
	// 构建请求体
	reqBody := map[string]interface{}{
		"model":  currentConfig.Model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建请求
	fmt.Println("jsonData", string(jsonData))
	fmt.Println("currentConfig.APIEndpoint", currentConfig.APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", currentConfig.APIEndpoint+"/api/generate", bytes.NewBuffer(jsonData))
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
