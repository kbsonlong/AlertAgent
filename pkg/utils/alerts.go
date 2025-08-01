package utils

import (
	"context"
	"fmt"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"

	"go.uber.org/zap"
)

// ProcessUnanalyzedAlerts 处理未分析的告警
// 该函数会查找所有未分析的告警并为它们创建AI分析任务
func ProcessUnanalyzedAlerts(ctx context.Context, messageQueue queue.MessageQueue) error {
	var alerts []model.Alert
	logger.L.Debug("Processing unanalyzed alerts...")
	// 获取未分析的告警
	if err := database.DB.Where("analysis = ?", "").Find(&alerts).Error; err != nil {
		return fmt.Errorf("failed to get unanalyzed alerts: %w", err)
	}

	if len(alerts) == 0 {
		logger.L.Info("No unanalyzed alerts found")
		return nil
	}

	logger.L.Info("Found unanalyzed alerts",
		zap.Int("count", len(alerts)),
	)

	// 创建任务生产者
	producer := queue.NewTaskProducer(messageQueue)

	// 为每个告警创建AI分析任务
	for _, alert := range alerts {
		logger.L.Debug("Processing alert",
			zap.Uint("id", alert.ID),
		)
		
		alertData := map[string]interface{}{
			"title":   alert.Title,
			"level":   alert.Level,
			"source":  alert.Source,
			"content": alert.Content,
		}
		
		if err := producer.PublishAIAnalysisTask(ctx, fmt.Sprintf("%d", alert.ID), alertData); err != nil {
			logger.L.Error("Failed to publish AI analysis task",
				zap.Uint("alert_id", alert.ID),
				zap.Error(err),
			)
			continue
		}
	}

	logger.L.Info("Successfully pushed tasks to queue",
		zap.Int("count", len(alerts)),
	)

	return nil
}