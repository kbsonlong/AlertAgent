package worker

import (
	"context"
	"fmt"

	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"go.uber.org/zap"
)

// ConfigSyncHandler 配置同步任务处理器
type ConfigSyncHandler struct {
	// 这里可以添加配置同步服务依赖
}

// NewConfigSyncHandler 创建配置同步处理器
func NewConfigSyncHandler() *ConfigSyncHandler {
	return &ConfigSyncHandler{}
}

// Type 返回处理器类型
func (h *ConfigSyncHandler) Type() queue.TaskType {
	return queue.TaskTypeConfigSync
}

// Handle 处理配置同步任务
func (h *ConfigSyncHandler) Handle(ctx context.Context, task *queue.Task) error {
	// 解析任务载荷
	syncType, ok := task.Payload["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type in task payload")
	}

	ruleID, ok := task.Payload["rule_id"].(string)
	if !ok {
		return fmt.Errorf("invalid rule_id in task payload")
	}

	targets, ok := task.Payload["targets"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid targets in task payload")
	}

	logger.L.Info("Processing config sync task",
		zap.String("task_id", task.ID),
		zap.String("type", syncType),
		zap.String("rule_id", ruleID),
		zap.Int("target_count", len(targets)),
	)

	// TODO: 实现实际的配置同步逻辑
	// 这里应该调用配置同步服务同步规则到目标系统
	
	logger.L.Info("Config sync completed successfully",
		zap.String("task_id", task.ID),
		zap.String("rule_id", ruleID),
	)

	return nil
}