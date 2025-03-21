package model

import "gorm.io/gorm"

// Settings 系统设置
type Settings struct {
	gorm.Model
	OllamaEndpoint string `json:"ollama_endpoint" gorm:"type:varchar(255);not null"`
	OllamaModel    string `json:"ollama_model" gorm:"type:varchar(100);not null"`
}
