package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

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

// CreateKnowledge 从告警创建知识条目
func (s *KnowledgeService) CreateKnowledge(ctx context.Context, alert *model.Alert) error {
	// 生成知识条目标题
	title := fmt.Sprintf("%s - %s", alert.Name, alert.Level)

	// 构建知识内容
	content := fmt.Sprintf("告警名称：%s\n告警级别：%s\n告警来源：%s\n告警内容：%s\n\n分析结果：\n%s",
		alert.Name, alert.Level, alert.Source, alert.Content, alert.Analysis)

	// 生成知识条目摘要
	summary, err := s.generateSummary(ctx, content)
	if err != nil {
		s.logger.Error("Failed to generate summary", zap.Error(err))
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// 生成向量表示
	vector, err := s.generateVector(ctx, content)
	if err != nil {
		s.logger.Error("Failed to generate vector", zap.Error(err))
		return fmt.Errorf("failed to generate vector: %w", err)
	}

	// 创建知识条目
	knowledge := &model.Knowledge{
		Title:    title,
		Content:  content,
		Category: "alert",
		Tags:     strings.Join([]string{alert.Level, alert.Source}, ","),
		Source:   "alert",
		SourceID: alert.ID,
		Vector:   vector,
		Summary:  summary,
	}

	// 保存到数据库
	if err := database.DB.Create(knowledge).Error; err != nil {
		s.logger.Error("Failed to save knowledge", zap.Error(err))
		return fmt.Errorf("failed to save knowledge: %w", err)
	}

	return nil
}

// FindSimilarKnowledge 查找相似知识
func (s *KnowledgeService) FindSimilarKnowledge(ctx context.Context, content string, limit int) ([]*model.Knowledge, error) {
	// 生成查询内容的向量表示
	vector, err := s.generateVector(ctx, content)
	if err != nil {
		s.logger.Error("Failed to generate vector", zap.Error(err))
		return nil, fmt.Errorf("failed to generate vector: %w", err)
	}

	// 从数据库中查找所有知识条目
	var allKnowledge []*model.Knowledge
	if err := database.DB.Find(&allKnowledge).Error; err != nil {
		s.logger.Error("Failed to get knowledge from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get knowledge from database: %w", err)
	}

	// 计算相似度并排序
	for _, k := range allKnowledge {
		k.Similarity = s.cosineSimilarity(vector, k.Vector)
	}

	// 按相似度降序排序
	sort.Slice(allKnowledge, func(i, j int) bool {
		return allKnowledge[i].Similarity > allKnowledge[j].Similarity
	})

	// 返回相似度最高的N个结果
	if len(allKnowledge) > limit {
		allKnowledge = allKnowledge[:limit]
	}

	return allKnowledge, nil
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
