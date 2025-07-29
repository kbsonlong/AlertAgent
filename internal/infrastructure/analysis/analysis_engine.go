package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/domain/analysis"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
)

// AnalysisEngineImpl AI分析引擎实现
type AnalysisEngineImpl struct {
	aiService    AIService
	templateRepo TemplateRepository
	logger       *zap.Logger
	config       *EngineConfig
}

// AIService AI服务接口
type AIService interface {
	Analyze(ctx context.Context, prompt string, data interface{}) (*AIResponse, error)
	IsHealthy() bool
	GetModelInfo() *ModelInfo
}

// TemplateRepository 模板仓库接口
type TemplateRepository interface {
	GetTemplate(analysisType analysis.AnalysisType) (*AnalysisTemplate, error)
	GetDefaultTemplate() (*AnalysisTemplate, error)
}

// AIResponse AI服务响应
type AIResponse struct {
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	ModelUsed string                 `json:"model_used"`
	Tokens    *TokenUsage            `json:"tokens,omitempty"`
}

// TokenUsage 令牌使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Provider    string `json:"provider"`
	Capabilities []string `json:"capabilities"`
}

// AnalysisTemplate 分析模板
type AnalysisTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        analysis.AnalysisType  `json:"type"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EngineConfig 引擎配置
type EngineConfig struct {
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	MaxPromptLength int           `json:"max_prompt_length"`
	EnableCache     bool          `json:"enable_cache"`
	CacheTTL        time.Duration `json:"cache_ttl"`
}

// DefaultEngineConfig 默认引擎配置
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		Timeout:         30 * time.Second,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		MaxPromptLength: 8000,
		EnableCache:     true,
		CacheTTL:        10 * time.Minute,
	}
}

// NewAnalysisEngine 创建分析引擎
func NewAnalysisEngine(
	aiService AIService,
	templateRepo TemplateRepository,
	config *EngineConfig,
) analysis.AnalysisEngine {
	if config == nil {
		config = DefaultEngineConfig()
	}

	return &AnalysisEngineImpl{
		aiService:    aiService,
		templateRepo: templateRepo,
		logger:       logger.L.Named("analysis-engine"),
		config:       config,
	}
}

// Analyze 执行分析
func (e *AnalysisEngineImpl) Analyze(ctx context.Context, request *analysis.AnalysisRequest) (*analysis.AnalysisResult, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if request.Alert == nil {
		return nil, fmt.Errorf("request alert cannot be nil")
	}

	// 创建分析任务
	task := &analysis.AnalysisTask{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		AlertID:   fmt.Sprintf("%d", request.Alert.ID),
		Type:      request.Type,
		Status:    analysis.AnalysisStatusProcessing,
		Priority:  request.Priority,
		Timeout:   request.Timeout,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  request.Options,
	}

	e.logger.Info("Starting alert analysis",
		zap.String("task_id", task.ID),
		zap.String("alert_id", task.AlertID),
		zap.String("analysis_type", string(task.Type)))

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// 获取分析模板
	template, err := e.getAnalysisTemplate(task.Type)
	if err != nil {
		e.logger.Error("Failed to get analysis template",
			zap.String("task_id", task.ID),
			zap.String("analysis_type", string(task.Type)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis template: %w", err)
	}

	// 构建分析提示
	prompt, err := e.buildAnalysisPrompt(template, request.Alert, request.Options)
	if err != nil {
		e.logger.Error("Failed to build analysis prompt",
			zap.String("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to build analysis prompt: %w", err)
	}

	// 执行AI分析（带重试）
	aiResponse, err := e.executeAnalysisWithRetry(ctx, prompt, request.Alert)
	if err != nil {
		e.logger.Error("Failed to execute AI analysis",
			zap.String("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to execute AI analysis: %w", err)
	}

	// 解析分析结果
	result, err := e.parseAnalysisResult(task, aiResponse, template)
	if err != nil {
		e.logger.Error("Failed to parse analysis result",
			zap.String("task_id", task.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse analysis result: %w", err)
	}

	e.logger.Info("Alert analysis completed",
		zap.String("task_id", task.ID),
		zap.String("result_id", result.ID),
		zap.Float64("confidence", result.ConfidenceScore))

	return result, nil
}

// IsHealthy 检查引擎健康状态
func (e *AnalysisEngineImpl) IsHealthy() bool {
	return e.aiService.IsHealthy()
}

// getAnalysisTemplate 获取分析模板
func (e *AnalysisEngineImpl) getAnalysisTemplate(analysisType analysis.AnalysisType) (*AnalysisTemplate, error) {
	template, err := e.templateRepo.GetTemplate(analysisType)
	if err != nil {
		e.logger.Warn("Failed to get specific template, using default",
			zap.String("analysis_type", string(analysisType)),
			zap.Error(err))
		
		// 尝试获取默认模板
		defaultTemplate, defaultErr := e.templateRepo.GetDefaultTemplate()
		if defaultErr != nil {
			return nil, fmt.Errorf("failed to get default template: %w", defaultErr)
		}
		return defaultTemplate, nil
	}
	return template, nil
}

// buildAnalysisPrompt 构建分析提示
func (e *AnalysisEngineImpl) buildAnalysisPrompt(
	template *AnalysisTemplate,
	alert *model.Alert,
	parameters map[string]interface{},
) (string, error) {
	// 准备模板变量
	vars := map[string]interface{}{
		"alert":      alert,
		"parameters": parameters,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// 添加告警详细信息
	vars["alert_name"] = alert.Name
	vars["alert_level"] = alert.Level
	vars["alert_status"] = alert.Status
	vars["alert_content"] = alert.Content
	vars["alert_source"] = alert.Source
	vars["alert_created_at"] = alert.CreatedAt.Format(time.RFC3339)

	// 替换模板中的变量
	prompt := template.Prompt
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := e.formatValue(value)
		prompt = strings.ReplaceAll(prompt, placeholder, valueStr)
	}

	// 检查提示长度
	if len(prompt) > e.config.MaxPromptLength {
		e.logger.Warn("Prompt length exceeds maximum, truncating",
			zap.Int("length", len(prompt)),
			zap.Int("max_length", e.config.MaxPromptLength))
		prompt = prompt[:e.config.MaxPromptLength]
	}

	return prompt, nil
}

// formatValue 格式化值为字符串
func (e *AnalysisEngineImpl) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	case map[string]interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%+v", v)
		}
		return string(data)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// executeAnalysisWithRetry 执行AI分析（带重试）
func (e *AnalysisEngineImpl) executeAnalysisWithRetry(
	ctx context.Context,
	prompt string,
	alert *model.Alert,
) (*AIResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			e.logger.Info("Retrying AI analysis",
				zap.Uint("alert_id", alert.ID),
				zap.Int("attempt", attempt))
			
			// 等待重试延迟
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(e.config.RetryDelay * time.Duration(attempt)):
			}
		}

		response, err := e.aiService.Analyze(ctx, prompt, alert)
		if err == nil {
			return response, nil
		}

		lastErr = err
		e.logger.Warn("AI analysis attempt failed",
			zap.Uint("alert_id", alert.ID),
			zap.Int("attempt", attempt),
			zap.Error(err))
	}

	return nil, fmt.Errorf("AI analysis failed after %d attempts: %w", e.config.MaxRetries+1, lastErr)
}

// parseAnalysisResult 解析分析结果
func (e *AnalysisEngineImpl) parseAnalysisResult(
	task *analysis.AnalysisTask,
	aiResponse *AIResponse,
	template *AnalysisTemplate,
) (*analysis.AnalysisResult, error) {
	// 尝试解析结构化结果
	var structuredResult map[string]interface{}
	if err := json.Unmarshal([]byte(aiResponse.Content), &structuredResult); err != nil {
		// 如果不是JSON，作为纯文本处理
		structuredResult = map[string]interface{}{
			"analysis": aiResponse.Content,
			"type":     "text",
		}
	}

	// 提取置信度
	confidence := 0.5 // 默认置信度
	if conf, ok := structuredResult["confidence"].(float64); ok {
		confidence = conf
	} else if conf, ok := structuredResult["confidence"].(int); ok {
		confidence = float64(conf) / 100.0
	}

	// 提取严重程度
	severity := "medium"
	if sev, ok := structuredResult["severity"].(string); ok {
		severity = sev
	}

	// 提取建议
	recommendations := []string{}
	if recs, ok := structuredResult["recommendations"].([]interface{}); ok {
		for _, rec := range recs {
			if recStr, ok := rec.(string); ok {
				recommendations = append(recommendations, recStr)
			}
		}
	}

	// 构建分析结果
	result := &analysis.AnalysisResult{
		ID:              fmt.Sprintf("result_%s_%d", task.ID, time.Now().Unix()),
		TaskID:          task.ID,
		AlertID:         task.AlertID,
		Type:            task.Type,
		Status:          analysis.AnalysisStatusCompleted,
		Result:          structuredResult,
		Summary:         aiResponse.Content,
		ConfidenceScore: confidence,
		Recommendations: recommendations,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"template_id":      template.ID,
			"template_version": template.Version,
			"model_used":       aiResponse.ModelUsed,
			"analysis_time":    time.Now().Format(time.RFC3339),
			"severity":         severity,
		},
	}

	// 添加令牌使用信息
	if aiResponse.Tokens != nil {
		result.Metadata["tokens"] = aiResponse.Tokens
	}

	// 添加AI响应元数据
	if aiResponse.Metadata != nil {
		for k, v := range aiResponse.Metadata {
			result.Metadata[k] = v
		}
	}

	return result, nil
}

// ValidateRequest 验证分析请求
func (e *AnalysisEngineImpl) ValidateRequest(request *analysis.AnalysisRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	if request.Type == "" {
		return fmt.Errorf("analysis type cannot be empty")
	}

	// 验证分析类型是否支持
	supportedTypes := e.GetSupportedTypes()
	supported := false
	for _, supportedType := range supportedTypes {
		if request.Type == supportedType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported analysis type: %s", request.Type)
	}

	return nil
}

// GetSupportedTypes 获取支持的分析类型
func (e *AnalysisEngineImpl) GetSupportedTypes() []analysis.AnalysisType {
	return []analysis.AnalysisType{
		analysis.AnalysisTypeRootCause,
		analysis.AnalysisTypeImpactAssess,
		analysis.AnalysisTypeSolution,
		analysis.AnalysisTypeClassification,
		analysis.AnalysisTypePriority,
	}
}

// EstimateProcessingTime 估算处理时间
func (e *AnalysisEngineImpl) EstimateProcessingTime(request *analysis.AnalysisRequest) time.Duration {
	// 基础处理时间
	baseTime := 10 * time.Second
	
	// 根据分析类型调整
	switch request.Type {
	case analysis.AnalysisTypeRootCause:
		return baseTime * 2
	case analysis.AnalysisTypeImpactAssess:
		return baseTime * 3
	case analysis.AnalysisTypeSolution:
		return baseTime * 4
	default:
		return baseTime
	}
}

// GetEngineInfo 获取引擎信息
func (e *AnalysisEngineImpl) GetEngineInfo() map[string]interface{} {
	modelInfo := e.aiService.GetModelInfo()
	return map[string]interface{}{
		"engine_type":       "ai_analysis_engine",
		"version":           "1.0.0",
		"supported_types":   e.GetSupportedTypes(),
		"model_info":        modelInfo,
		"healthy":           e.IsHealthy(),
		"max_prompt_length": e.config.MaxPromptLength,
		"timeout":           e.config.Timeout.String(),
		"max_retries":       e.config.MaxRetries,
	}
}