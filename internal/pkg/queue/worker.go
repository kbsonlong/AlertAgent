package queue

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// Worker 分析工作器
type Worker struct {
	openAIService *service.OpenAIService
	shutdown      chan struct{}
	isRunning     bool
}

// NewWorker 创建新的工作器
func NewWorker(openAIService *service.OpenAIService) *Worker {
	return &Worker{
		openAIService: openAIService,
		shutdown:      make(chan struct{}),
		isRunning:     false,
	}
}

// Start 启动工作器
func (w *Worker) Start(ctx context.Context) {
	if w.isRunning {
		logger.Warn("Worker is already running")
		return
	}

	w.isRunning = true
	logger.Info("Starting analysis worker")

	go func() {
		for {
			select {
			case <-w.shutdown:
				logger.Info("Worker shutting down")
				return
			case <-ctx.Done():
				logger.Info("Context canceled, worker shutting down")
				return
			default:
				// 处理任务
				w.processTask(ctx)
				// 短暂休眠，避免CPU占用过高
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop 停止工作器
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	logger.Info("Stopping analysis worker")
	close(w.shutdown)
	w.isRunning = false
}

// processTask 处理单个分析任务
func (w *Worker) processTask(ctx context.Context) {
	// 创建一个带有超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// 从队列获取任务
	task, err := DequeueAnalysisTask(timeoutCtx)
	if err != nil {
		logger.Error("Failed to dequeue analysis task", zap.Error(err))
		return
	}

	// 队列为空
	if task == nil {
		return
	}

	logger.Info("Processing analysis task",
		zap.Uint("alert_id", task.AlertID),
	)

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.First(&alert, task.AlertID).Error; err != nil {
		logger.Error("Failed to get alert",
			zap.Error(err),
			zap.Uint("alert_id", task.AlertID),
		)

		// 记录失败结果
		result := &AnalysisResult{
			AlertID:   task.AlertID,
			Analysis:  "",
			CreatedAt: time.Now(),
			Error:     fmt.Sprintf("Failed to get alert: %v", err),
		}
		if err := CompleteAnalysisTask(ctx, result); err != nil {
			logger.Error("Failed to complete analysis task",
				zap.Error(err),
				zap.Uint("alert_id", task.AlertID),
			)
		}
		return
	}

	// 调用OpenAI服务进行分析
	analysis, err := w.openAIService.AnalyzeAlert(timeoutCtx, &alert)

	// 创建分析结果
	result := &AnalysisResult{
		AlertID:   task.AlertID,
		CreatedAt: time.Now(),
	}

	if err != nil {
		logger.Error("Failed to analyze alert",
			zap.Error(err),
			zap.Uint("alert_id", task.AlertID),
		)
		result.Error = err.Error()
	} else {
		result.Analysis = analysis

		// 更新数据库中的分析结果
		alert.Analysis = analysis
		if err := database.DB.Save(&alert).Error; err != nil {
			logger.Error("Failed to update alert analysis in database",
				zap.Error(err),
				zap.Uint("alert_id", task.AlertID),
			)
			result.Error = fmt.Sprintf("Analysis completed but failed to update database: %v", err)
		}
	}

	// 完成任务，保存结果
	if err := CompleteAnalysisTask(ctx, result); err != nil {
		logger.Error("Failed to complete analysis task",
			zap.Error(err),
			zap.Uint("alert_id", task.AlertID),
		)
	}
}
