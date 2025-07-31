package types

import (
	"time"
)

// Metadata 通用元数据结构
type Metadata struct {
	CreatedBy   string            `json:"created_by" gorm:"type:varchar(255)"`
	UpdatedBy   string            `json:"updated_by" gorm:"type:varchar(255)"`
	Version     string            `json:"version" gorm:"type:varchar(50)"`
	Tags        []string          `json:"tags" gorm:"type:json"`
	Labels      map[string]string `json:"labels" gorm:"type:json"`
	Annotations map[string]string `json:"annotations" gorm:"type:json"`
	Extra       map[string]interface{} `json:"extra" gorm:"type:json"`
}

// RetryConfig 重试配置结构
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay" yaml:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay" yaml:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor" yaml:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors" yaml:"retryable_errors"`
	Enabled         bool          `json:"enabled" yaml:"enabled"`
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialDelay:    time.Second,
		MaxDelay:        time.Minute,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"timeout", "connection refused", "network unreachable"},
		Enabled:         true,
	}
}

// ShouldRetry 判断是否应该重试
func (rc *RetryConfig) ShouldRetry(err error, attempt int) bool {
	if !rc.Enabled || attempt >= rc.MaxRetries {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range rc.RetryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// GetRetryDelay 获取重试延迟时间
func (rc *RetryConfig) GetRetryDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rc.InitialDelay) * pow(rc.BackoffFactor, float64(attempt)))
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}
	return delay
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring 查找子字符串的位置
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// pow 计算幂次方
func pow(base float64, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return base
	}
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}