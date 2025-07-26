package queue

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/types"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// Worker 队列工作器
type Worker struct {
	queue         Queue
	ollamaService *service.OllamaService
	shutdown      chan struct{}
	isRunning     bool
}

// NewWorker 创建工作器
func NewWorker(queue Queue, ollamaService *service.OllamaService) *Worker {
	return &Worker{
		queue:         queue,
		ollamaService: ollamaService,
		shutdown:      make(chan struct{}),
		isRunning:     false,
	}
}

// Start 启动工作器
func (w *Worker) Start(ctx context.Context) error {
	logger.L.Info("Starting worker...")
	if w.isRunning {
		logger.L.Warn("Worker is already running")
		return nil
	}

	w.isRunning = true
	logger.L.Info("Starting analysis worker")

	go func() {
		for {
			select {
			case <-w.shutdown:
				logger.L.Info("Worker shutting down")
				return
			case <-ctx.Done():
				logger.L.Info("Context canceled, worker shutting down")
				return
			default:
				task, err := w.queue.Pop(ctx)
				if err != nil {
					logger.L.Error("Failed to pop task from queue",
						zap.Error(err),
					)
					time.Sleep(time.Second)
					continue
				}

				if task == nil {
					logger.L.Debug("No tasks in queue, waiting...")
					time.Sleep(time.Second)
					continue
				}

				logger.L.Info("Processing task",
					zap.Uint("task_id", task.ID),
					zap.Time("created_at", task.CreatedAt),
				)

				// 处理任务
				result, err := w.processTask(ctx, task)
				if err != nil {
					logger.L.Error("Failed to process task",
						zap.Uint("task_id", task.ID),
						zap.Error(err),
					)
					continue
				}

				// 保存结果
				if err := w.queue.Complete(ctx, result); err != nil {
					logger.L.Error("Failed to save task result",
						zap.Uint("task_id", task.ID),
						zap.Error(err),
					)
					continue
				}

				logger.L.Info("Task completed successfully",
					zap.Uint("task_id", task.ID),
					zap.Duration("duration", time.Since(task.CreatedAt)),
				)
			}
		}
	}()

	return nil
}

// Stop 停止工作器
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	logger.L.Info("Stopping analysis worker")
	close(w.shutdown)
	w.isRunning = false
}

// processTask 处理任务
func (w *Worker) processTask(ctx context.Context, task *types.AlertTask) (*types.AlertResult, error) {
	// 创建一个带有超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	logger.L.Info("Processing analysis task",
		zap.Uint("task_id", task.ID),
	)

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.First(&alert, task.ID).Error; err != nil {
		logger.L.Error("Failed to get alert",
			zap.Error(err),
			zap.Uint("alert_id", task.ID),
		)

		// 记录失败结果
		result := &types.AlertResult{
			TaskID:    task.ID,
			Status:    "failed",
			Message:   fmt.Sprintf("Failed to get alert: %v", err),
			CreatedAt: time.Now(),
		}
		return result, nil
	}

	// 调用Ollama服务进行分析
	analysis, err := w.ollamaService.AnalyzeAlert(timeoutCtx, &alert)

	// 创建分析结果
	result := &types.AlertResult{
		TaskID:    task.ID,
		CreatedAt: time.Now(),
	}

	if err != nil {
		logger.L.Error("Failed to analyze alert",
			zap.Error(err),
			zap.Uint("alert_id", task.ID),
		)
		result.Status = "failed"
		result.Message = err.Error()
	} else {
		result.Status = "completed"
		result.Message = "Analysis completed successfully"

		// 更新数据库中的分析结果
		alert.Analysis = analysis
		if err := database.DB.Save(&alert).Error; err != nil {
			logger.L.Error("Failed to update alert analysis in database",
				zap.Error(err),
				zap.Uint("alert_id", task.ID),
			)
			result.Status = "warning"
			result.Message = fmt.Sprintf("Analysis completed but failed to update database: %v", err)
		}
	}

	return result, nil
}
