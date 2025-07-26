package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"

	"go.uber.org/zap"
)

// KnowledgeService 知识库服务
type KnowledgeService struct {
	ollamaService *OllamaService
	logger        *zap.Logger
}

// NewKnowledgeService 创建知识库服务实例
func NewKnowledgeService(ollamaService *OllamaService) *KnowledgeService {
	return &KnowledgeService{
		ollamaService: ollamaService,
		logger:        zap.L(),
	}
}

// CreateKnowledge 将告警转换为知识库记录
func CreateKnowledge(alert *model.Alert) (*model.Knowledge, error) {
	// 生成标题
	title := fmt.Sprintf("[%s] %s", alert.Source, alert.Title)

	// 生成内容
	content := fmt.Sprintf(`## 告警信息

- 告警标题：%s
- 告警来源：%s
- 告警级别：%s
- 告警内容：%s

## 分析结果

%s
`, alert.Title, alert.Source, alert.Level, alert.Content, alert.Analysis)

	// 检查是否已存在相同告警ID的知识库记录
	var existingKnowledge model.Knowledge
	err := database.DB.Where("source = ? AND source_id = ?", "alert", alert.ID).First(&existingKnowledge).Error

	if err == nil {
		// 记录已存在，更新现有记录（排除Vector字段）
		updateData := map[string]interface{}{
			"title":      title,
			"content":    content,
			"tags":       fmt.Sprintf("%s,%s", alert.Level, alert.Source),
			"summary":    fmt.Sprintf("%s级别告警：%s", alert.Level, alert.Title),
			"updated_at": time.Now(),
		}

		if err := database.DB.Model(&existingKnowledge).Updates(updateData).Error; err != nil {
			return nil, fmt.Errorf("更新知识库记录失败: %v", err)
		}

		// 重新获取更新后的记录
		if err := database.DB.First(&existingKnowledge, existingKnowledge.ID).Error; err != nil {
			return nil, fmt.Errorf("获取更新后的知识库记录失败: %v", err)
		}

		return &existingKnowledge, nil
	}

	// 记录不存在，创建新的知识库记录
	knowledge := &model.Knowledge{
		Title:     title,
		Content:   content,
		Category:  "告警处理",
		Tags:      fmt.Sprintf("%s,%s", alert.Level, alert.Source),
		Source:    "alert",
		SourceID:  alert.ID,
		Summary:   fmt.Sprintf("%s级别告警：%s", alert.Level, alert.Title),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存到数据库
	if err := database.DB.Create(knowledge).Error; err != nil {
		return nil, fmt.Errorf("保存知识库记录失败: %v", err)
	}

	return knowledge, nil
}

// generateSummary 生成内容摘要
func (s *KnowledgeService) generateSummary(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf("请为以下内容生成一个简短的摘要（不超过100字）：\n\n%s", content)
	return s.ollamaService.callOllamaAPI(ctx, prompt)
}

// generateVector 生成内容的向量表示
func (s *KnowledgeService) generateVector(ctx context.Context, content string) ([]float32, error) {
	// 构建请求体
	reqBody := map[string]interface{}{
		"model":  s.ollamaService.config.Model,
		"prompt": content,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 发送请求到Ollama的embeddings接口
	resp, err := s.ollamaService.client.Post(
		s.ollamaService.config.APIEndpoint+"/api/embeddings",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Embedding []float32 `json:"embedding"`
		Error     string    `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("ollama API error: %s", result.Error)
	}

	return result.Embedding, nil
}

// cosineSimilarity 计算余弦相似度
func (s *KnowledgeService) cosineSimilarity(v1, v2 []float32) float32 {
	if len(v1) != len(v2) {
		return 0
	}

	var dotProduct, norm1, norm2 float32
	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		norm1 += v1[i] * v1[i]
		norm2 += v2[i] * v2[i]
	}

	norm1 = float32(math.Sqrt(float64(norm1)))
	norm2 = float32(math.Sqrt(float64(norm2)))

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (norm1 * norm2)
}
