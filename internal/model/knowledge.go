package model

import (
	"time"
)

// Knowledge 知识库模型
type Knowledge struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Source    string    `json:"source" gorm:"type:varchar(50);not null"`
	AlertID   uint      `json:"alert_id" gorm:"not null"`
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
