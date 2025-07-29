package dify

import (
	"fmt"
	"time"
)

// DifyConfig Dify客户端配置
type DifyConfig struct {
	// 基础配置
	BaseURL    string `mapstructure:"base_url" json:"base_url"`
	APIKey     string `mapstructure:"api_key" json:"api_key"`
	AppToken   string `mapstructure:"app_token" json:"app_token"`
	UserID     string `mapstructure:"user_id" json:"user_id"`
	
	// HTTP客户端配置
	Timeout         time.Duration `mapstructure:"timeout" json:"timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts" json:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay" json:"retry_delay"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns"`
	MaxConnsPerHost int           `mapstructure:"max_conns_per_host" json:"max_conns_per_host"`
	
	// 工作流配置
	WorkflowConfig WorkflowConfig `mapstructure:"workflow" json:"workflow"`
	
	// 知识库配置
	KnowledgeConfig KnowledgeConfig `mapstructure:"knowledge" json:"knowledge"`
	
	// 分析配置
	AnalysisConfig AnalysisConfig `mapstructure:"analysis" json:"analysis"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	// 默认工作流ID
	DefaultWorkflowID string `mapstructure:"default_workflow_id" json:"default_workflow_id"`
	
	// 工作流映射（告警类型 -> 工作流ID）
	WorkflowMapping map[string]string `mapstructure:"workflow_mapping" json:"workflow_mapping"`
	
	// 执行配置
	MaxExecutionTime time.Duration `mapstructure:"max_execution_time" json:"max_execution_time"`
	PollingInterval  time.Duration `mapstructure:"polling_interval" json:"polling_interval"`
	MaxPollingTime   time.Duration `mapstructure:"max_polling_time" json:"max_polling_time"`
}

// KnowledgeConfig 知识库配置
type KnowledgeConfig struct {
	// 默认数据集ID列表
	DefaultDatasetIDs []string `mapstructure:"default_dataset_ids" json:"default_dataset_ids"`
	
	// 数据集映射（告警类型 -> 数据集ID列表）
	DatasetMapping map[string][]string `mapstructure:"dataset_mapping" json:"dataset_mapping"`
	
	// 搜索配置
	SearchLimit     int     `mapstructure:"search_limit" json:"search_limit"`
	SimilarityScore float64 `mapstructure:"similarity_score" json:"similarity_score"`
}

// AnalysisConfig 分析配置
type AnalysisConfig struct {
	// 并发配置
	MaxConcurrentTasks int `mapstructure:"max_concurrent_tasks" json:"max_concurrent_tasks"`
	TaskQueueSize      int `mapstructure:"task_queue_size" json:"task_queue_size"`
	
	// 超时配置
	AnalysisTimeout time.Duration `mapstructure:"analysis_timeout" json:"analysis_timeout"`
	ContextTimeout  time.Duration `mapstructure:"context_timeout" json:"context_timeout"`
	
	// 重试配置
	MaxRetries    int           `mapstructure:"max_retries" json:"max_retries"`
	RetryInterval time.Duration `mapstructure:"retry_interval" json:"retry_interval"`
	
	// 结果配置
	ResultTTL        time.Duration `mapstructure:"result_ttl" json:"result_ttl"`
	MaxResultSize    int64         `mapstructure:"max_result_size" json:"max_result_size"`
	EnableMetrics    bool          `mapstructure:"enable_metrics" json:"enable_metrics"`
	EnableTracing    bool          `mapstructure:"enable_tracing" json:"enable_tracing"`
}

// DefaultDifyConfig 返回默认配置
func DefaultDifyConfig() *DifyConfig {
	return &DifyConfig{
		BaseURL:         "https://api.dify.ai",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		MaxIdleConns:    10,
		MaxConnsPerHost: 10,
		
		WorkflowConfig: WorkflowConfig{
			MaxExecutionTime: 5 * time.Minute,
			PollingInterval:  2 * time.Second,
			MaxPollingTime:   10 * time.Minute,
			WorkflowMapping:  make(map[string]string),
		},
		
		KnowledgeConfig: KnowledgeConfig{
			SearchLimit:     10,
			SimilarityScore: 0.7,
			DatasetMapping:  make(map[string][]string),
		},
		
		AnalysisConfig: AnalysisConfig{
			MaxConcurrentTasks: 10,
			TaskQueueSize:      100,
			AnalysisTimeout:    10 * time.Minute,
			ContextTimeout:     30 * time.Second,
			MaxRetries:         3,
			RetryInterval:      5 * time.Second,
			ResultTTL:          24 * time.Hour,
			MaxResultSize:      1024 * 1024, // 1MB
			EnableMetrics:      true,
			EnableTracing:      false,
		},
	}
}

// Validate 验证配置
func (c *DifyConfig) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	
	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.AnalysisConfig.MaxConcurrentTasks <= 0 {
		return fmt.Errorf("max_concurrent_tasks must be positive")
	}
	
	if c.AnalysisConfig.TaskQueueSize <= 0 {
		return fmt.Errorf("task_queue_size must be positive")
	}
	
	return nil
}