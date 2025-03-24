package model

import (
	"time"

	"gorm.io/gorm"
)

// Knowledge 知识库条目
type Knowledge struct {
	ID         uint           `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Title      string         `json:"title" gorm:"type:varchar(255);not null"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	Category   string         `json:"category" gorm:"type:varchar(100);not null"`
	Tags       string         `json:"tags" gorm:"type:text"`
	Source     string         `json:"source" gorm:"type:varchar(255);not null"`
	SourceID   uint           `json:"source_id" gorm:"not null"`
	Vector     []float32      `json:"vector" gorm:"type:json"`
	Summary    string         `json:"summary" gorm:"type:text"`
	Similarity float32        `json:"similarity" gorm:"-"`
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
