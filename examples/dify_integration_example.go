package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"go.uber.org/zap"
	
	"alert_agent/internal/infrastructure/dify"
)

// SimpleDifyExample 简单的Dify集成示例
type SimpleDifyExample struct {
	logger *zap.Logger
}

// NewSimpleDifyExample 创建简单的Dify集成示例
func NewSimpleDifyExample(logger *zap.Logger) *SimpleDifyExample {
	return &SimpleDifyExample{
		logger: logger,
	}
}

// RunDifyClientExample 运行Dify客户端示例
func (e *SimpleDifyExample) RunDifyClientExample(ctx context.Context) error {
	e.logger.Info("开始运行Dify客户端示例")
	
	// 创建Dify客户端
	client := dify.NewDifyClient(
		"https://api.dify.ai",
		"your-api-key",
		e.logger,
	)
	
	// 执行健康检查
	err := client.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	
	e.logger.Info("Dify服务健康检查成功")
	
	return nil
}

// RunConfigExample 运行配置示例
func (e *SimpleDifyExample) RunConfigExample() {
	e.logger.Info("开始运行配置示例")
	
	// 创建默认配置
	config := dify.DefaultDifyConfig()
	
	// 自定义配置
	config.BaseURL = "https://api.dify.ai"
	config.APIKey = "your-api-key"
	config.Timeout = 30 * time.Second
	
	// 验证配置
	if err := config.Validate(); err != nil {
		e.logger.Error("配置验证失败", zap.Error(err))
		return
	}
	
	e.logger.Info("配置验证成功",
		zap.String("base_url", config.BaseURL),
		zap.Duration("timeout", config.Timeout),
		zap.Int("max_concurrent_tasks", config.AnalysisConfig.MaxConcurrentTasks),
	)
}

// RunCompleteExample 运行完整示例
func (e *SimpleDifyExample) RunCompleteExample() {
	ctx := context.Background()
	
	e.logger.Info("开始运行Dify集成完整示例")
	
	// 1. 配置示例
	e.RunConfigExample()
	
	// 2. 客户端示例
	if err := e.RunDifyClientExample(ctx); err != nil {
		log.Printf("客户端示例失败: %v", err)
	}
	
	e.logger.Info("Dify集成完整示例运行完成")
}

// RunExample 运行示例的函数
func RunExample() {
	// 创建日志器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// 创建示例
	example := NewSimpleDifyExample(logger)
	
	// 运行示例
	example.RunCompleteExample()
}

// 使用说明:
// 1. 设置环境变量 DIFY_API_KEY
// 2. 运行: go run examples/dify_integration_example.go
// 3. 查看日志输出了解集成状态
//
// 生产环境集成步骤:
// 1. 在配置文件中设置Dify相关配置
// 2. 使用依赖注入初始化Dify服务
// 3. 在告警处理流程中调用Dify分析服务
// 4. 监控分析任务状态和结果
// 5. 将分析结果回写到告警记录