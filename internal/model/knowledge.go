package model

import (
	"time"
)

// Knowledge 知识库模型
type Knowledge struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Category  string    `json:"category" gorm:"type:varchar(100);not null"`
	Tags      string    `json:"tags" gorm:"type:text"`
	Source    string    `json:"source" gorm:"type:varchar(255);not null"`
	SourceID  uint      `json:"source_id" gorm:"not null"`
	Vector    string    `json:"vector" gorm:"type:json;default:null"`
	Summary   string    `json:"summary" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// KnowledgeResponse 知识库条目响应
type KnowledgeResponse struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Category   string    `json:"category"`
	Tags       string    `json:"tags,omitempty"`
	Source     string    `json:"source"`
	SourceID   uint      `json:"source_id"`
	Summary    string    `json:"summary,omitempty"`
	Similarity float32   `json:"similarity,omitempty"`
}

// ToResponse 转换为响应格式
func (k *Knowledge) ToResponse() KnowledgeResponse {
	return KnowledgeResponse{
		ID:        k.ID,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
		Title:     k.Title,
		Content:   k.Content,
		Category:  k.Category,
		Tags:      k.Tags,
		Source:    k.Source,
		SourceID:  k.SourceID,
		Summary:   k.Summary,
	}
}
