package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAsyncTaskProcessing 测试异步任务处理端到端流程
func TestAsyncTaskProcessing(t *testing.T) {
	// 清理测试数据
	require.NoError(t, clearTestData())

	t.Run("AI分析任务处理", func(t *testing.T) {
		testAIAnalysisTask(t)
	})

	t.Run("通知任务处理", func(t *testing.T) {
		testNotificationTask(t)
	})

	t.Run("配置同步任务处理", func(t *testing.T) {
		testConfigSyncTask(t)
	})

	t.Run("任务重试机制", func(t *testing.T) {
		testTaskRetryMechanism(t)
	})

	t.Run("并发任务处理", func(t *testing.T) {
		testConcurrentTaskProcessing(t)
	})

	t.Run("任务优先级处理", func(t *testing.T) {
		testTaskPriorityProcessing(t)
	})
}

func testAIAnalysisTask(t *testing.T) {
	// 创建任务生产者和消费者
	producer := NewTaskProducer(testRedis)
	worker := NewTestWorker("ai_analysis", testDB, testRedis)

	// 启动Worker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go worker.Start(ctx)

	// 发布AI分析任务
	alertID := uuid.New().String()
	alertData := map[string]interface{}{
		"alertname": "HighCPUUsage",
		"instance":  "server-01",
		"severity":  "critical",
		"summary":   "CPU usage is above 90%",
	}

	err := producer.PublishAIAnalysisTask(ctx, alertID, alertData)
	require.NoError(t, err)

	// 等待任务处理完成
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'ai_analysis'").Scan(&count)
		return count > 0
	}, 20*time.Second, 500*time.Millisecond)

	assert.True(t, success, "AI analysis task should be completed")

	// 验证任务结果
	var task TaskRecord
	err = testDB.Raw(`
		SELECT id, task_type, payload, status, error_message 
		FROM task_queue 
		WHERE task_type = 'ai_analysis' AND status = 'completed'
		LIMIT 1
	`).Scan(&task).Error

	require.NoError(t, err)
	assert.Equal(t, "ai_analysis", task.TaskType)
	assert.Equal(t, "completed", task.Status)
	assert.Empty(t, task.ErrorMessage)

	// 验证payload包含正确的告警数据
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(task.Payload), &payload)
	require.NoError(t, err)
	assert.Equal(t, alertID, payload["alert_id"])
}

func testNotificationTask(t *testing.T) {
	producer := NewTaskProducer(testRedis)
	worker := NewTestWorker("notification", testDB, testRedis)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go worker.Start(ctx)

	// 发布通知任务
	alertID := uuid.New().String()
	channels := []string{"dingtalk", "email"}
	message := map[string]interface{}{
		"title":    "告警通知",
		"content":  "系统出现高CPU使用率告警",
		"severity": "critical",
	}

	err := producer.PublishNotificationTask(ctx, alertID, channels, message)
	require.NoError(t, err)

	// 等待任务处理完成
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'notification'").Scan(&count)
		return count > 0
	}, 20*time.Second, 500*time.Millisecond)

	assert.True(t, success, "Notification task should be completed")

	// 验证任务结果
	var task TaskRecord
	err = testDB.Raw(`
		SELECT id, task_type, payload, status, error_message 
		FROM task_queue 
		WHERE task_type = 'notification' AND status = 'completed'
		LIMIT 1
	`).Scan(&task).Error

	require.NoError(t, err)
	assert.Equal(t, "notification", task.TaskType)
	assert.Equal(t, "completed", task.Status)
}

func testConfigSyncTask(t *testing.T) {
	producer := NewTaskProducer(testRedis)
	worker := NewTestWorker("config_sync", testDB, testRedis)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go worker.Start(ctx)

	// 发布配置同步任务
	task := &ConfigSyncTask{
		Type:     "rule_create",
		RuleID:   uuid.New().String(),
		Targets:  []string{"cluster-1", "cluster-2"},
		Priority: "normal",
	}

	err := producer.PublishConfigSyncTask(ctx, task)
	require.NoError(t, err)

	// 等待任务处理完成
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'config_sync'").Scan(&count)
		return count > 0
	}, 20*time.Second, 500*time.Millisecond)

	assert.True(t, success, "Config sync task should be completed")
}

func testTaskRetryMechanism(t *testing.T) {
	producer := NewTaskProducer(testRedis)
	worker := NewTestWorker("failing_task", testDB, testRedis)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go worker.Start(ctx)

	// 发布一个会失败的任务
	taskData := map[string]interface{}{
		"should_fail": true,
		"fail_count":  2, // 前2次失败，第3次成功
	}

	err := producer.PublishTask(ctx, "test_queue", &Task{
		Type:     "failing_task",
		Payload:  taskData,
		Priority: 1,
		MaxRetry: 3,
	})
	require.NoError(t, err)

	// 等待任务最终成功
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'failing_task'").Scan(&count)
		return count > 0
	}, 25*time.Second, 1*time.Second)

	assert.True(t, success, "Failing task should eventually succeed after retries")

	// 验证重试次数
	var task TaskRecord
	err = testDB.Raw(`
		SELECT retry_count FROM task_queue 
		WHERE task_type = 'failing_task' AND status = 'completed'
		LIMIT 1
	`).Scan(&task).Error

	require.NoError(t, err)
	assert.Equal(t, 2, task.RetryCount) // 应该重试了2次
}

func testConcurrentTaskProcessing(t *testing.T) {
	producer := NewTaskProducer(testRedis)
	
	// 启动多个Worker
	workerCount := 3
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < workerCount; i++ {
		worker := NewTestWorker("concurrent_task", testDB, testRedis)
		go worker.Start(ctx)
	}

	// 发布多个任务
	taskCount := 10
	var wg sync.WaitGroup
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		go func(taskIndex int) {
			defer wg.Done()
			
			taskData := map[string]interface{}{
				"task_index": taskIndex,
				"data":       fmt.Sprintf("task-%d", taskIndex),
			}

			err := producer.PublishTask(ctx, "test_queue", &Task{
				Type:     "concurrent_task",
				Payload:  taskData,
				Priority: 1,
				MaxRetry: 3,
			})
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 等待所有任务处理完成
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'concurrent_task'").Scan(&count)
		return count >= int64(taskCount)
	}, 25*time.Second, 500*time.Millisecond)

	assert.True(t, success, "All concurrent tasks should be completed")

	// 验证没有重复处理
	var completedCount int64
	err := testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'concurrent_task'").Scan(&completedCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(taskCount), completedCount)
}

func testTaskPriorityProcessing(t *testing.T) {
	producer := NewTaskProducer(testRedis)
	worker := NewTestWorker("priority_task", testDB, testRedis)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go worker.Start(ctx)

	// 发布不同优先级的任务
	tasks := []struct {
		priority int
		data     string
	}{
		{1, "low-priority"},
		{5, "high-priority"},
		{3, "medium-priority"},
		{5, "high-priority-2"},
		{1, "low-priority-2"},
	}

	for i, task := range tasks {
		taskData := map[string]interface{}{
			"order": i,
			"data":  task.data,
		}

		err := producer.PublishTask(ctx, "test_queue", &Task{
			Type:     "priority_task",
			Payload:  taskData,
			Priority: task.priority,
			MaxRetry: 3,
		})
		require.NoError(t, err)
	}

	// 等待所有任务处理完成
	success := waitForCondition(func() bool {
		var count int64
		testDB.Raw("SELECT COUNT(*) FROM task_queue WHERE status = 'completed' AND task_type = 'priority_task'").Scan(&count)
		return count >= int64(len(tasks))
	}, 25*time.Second, 500*time.Millisecond)

	assert.True(t, success, "All priority tasks should be completed")

	// 验证高优先级任务先处理（通过完成时间）
	var completedTasks []TaskRecord
	err := testDB.Raw(`
		SELECT payload, completed_at FROM task_queue 
		WHERE status = 'completed' AND task_type = 'priority_task'
		ORDER BY completed_at
	`).Scan(&completedTasks).Error

	require.NoError(t, err)
	assert.Len(t, completedTasks, len(tasks))

	// 验证前两个完成的任务是高优先级的
	for i := 0; i < 2; i++ {
		var payload map[string]interface{}
		err = json.Unmarshal([]byte(completedTasks[i].Payload), &payload)
		require.NoError(t, err)
		
		data := payload["data"].(string)
		assert.Contains(t, data, "high-priority")
	}
}

// Task 任务结构
type Task struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Priority int                    `json:"priority"`
	Retry    int                    `json:"retry"`
	MaxRetry int                    `json:"max_retry"`
	CreateAt time.Time              `json:"create_at"`
}

// TaskRecord 数据库任务记录
type TaskRecord struct {
	ID           string    `db:"id"`
	TaskType     string    `db:"task_type"`
	Payload      string    `db:"payload"`
	Status       string    `db:"status"`
	RetryCount   int       `db:"retry_count"`
	ErrorMessage string    `db:"error_message"`
	CompletedAt  time.Time `db:"completed_at"`
}

// ConfigSyncTask 配置同步任务
type ConfigSyncTask struct {
	Type     string   `json:"type"`
	RuleID   string   `json:"rule_id"`
	Targets  []string `json:"targets"`
	Priority string   `json:"priority"`
}

// TaskProducer 任务生产者
type TaskProducer struct {
	redis *redis.Client
}

func NewTaskProducer(redisClient *redis.Client) *TaskProducer {
	return &TaskProducer{redis: redisClient}
}

func (tp *TaskProducer) PublishTask(ctx context.Context, queueName string, task *Task) error {
	task.ID = uuid.New().String()
	task.CreateAt = time.Now()

	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 保存到数据库
	err = testDB.Exec(`
		INSERT INTO task_queue (id, queue_name, task_type, payload, priority, max_retry, status)
		VALUES (?, ?, ?, ?, ?, ?, 'pending')
	`, task.ID, queueName, task.Type, string(taskData), task.Priority, task.MaxRetry).Error

	if err != nil {
		return err
	}

	// 发送到Redis队列
	return tp.redis.LPush(ctx, queueName, taskData).Err()
}

func (tp *TaskProducer) PublishAIAnalysisTask(ctx context.Context, alertID string, alertData map[string]interface{}) error {
	task := &Task{
		Type:     "ai_analysis",
		Priority: 1,
		MaxRetry: 3,
		Payload: map[string]interface{}{
			"alert_id":       alertID,
			"alert_data":     alertData,
			"analysis_type":  "root_cause",
		},
	}

	return tp.PublishTask(ctx, "ai_analysis", task)
}

func (tp *TaskProducer) PublishNotificationTask(ctx context.Context, alertID string, channels []string, message map[string]interface{}) error {
	task := &Task{
		Type:     "notification",
		Priority: 2,
		MaxRetry: 5,
		Payload: map[string]interface{}{
			"alert_id": alertID,
			"channels": channels,
			"message":  message,
		},
	}

	return tp.PublishTask(ctx, "notification", task)
}

func (tp *TaskProducer) PublishConfigSyncTask(ctx context.Context, syncTask *ConfigSyncTask) error {
	task := &Task{
		Type:     "config_sync",
		Priority: 3,
		MaxRetry: 3,
		Payload: map[string]interface{}{
			"sync_type": syncTask.Type,
			"rule_id":   syncTask.RuleID,
			"targets":   syncTask.Targets,
			"priority":  syncTask.Priority,
		},
	}

	return tp.PublishTask(ctx, "config_sync", task)
}

// TestWorker 测试用Worker
type TestWorker struct {
	taskType string
	db       *gorm.DB
	redis    *redis.Client
}

func NewTestWorker(taskType string, db *gorm.DB, redisClient *redis.Client) *TestWorker {
	return &TestWorker{
		taskType: taskType,
		db:       db,
		redis:    redisClient,
	}
}

func (w *TestWorker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := w.processTask(ctx); err != nil {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (w *TestWorker) processTask(ctx context.Context) error {
	// 从Redis队列获取任务
	result, err := w.redis.BRPop(ctx, time.Second, w.taskType).Result()
	if err != nil {
		return err
	}

	if len(result) < 2 {
		return fmt.Errorf("invalid task data")
	}

	taskData := result[1]
	var task Task
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		return err
	}

	// 更新任务状态为处理中
	err = w.db.Exec(`
		UPDATE task_queue SET status = 'processing', started_at = NOW() 
		WHERE id = ?
	`, task.ID).Error
	if err != nil {
		return err
	}

	// 处理任务
	err = w.handleTask(ctx, &task)
	
	if err != nil {
		// 处理失败，更新重试次数
		w.db.Exec(`
			UPDATE task_queue 
			SET retry_count = retry_count + 1, error_message = ?, status = 'failed'
			WHERE id = ?
		`, err.Error(), task.ID)

		// 检查是否需要重试
		var retryCount, maxRetry int
		w.db.Raw("SELECT retry_count, max_retry FROM task_queue WHERE id = ?", task.ID).Row().Scan(&retryCount, &maxRetry)
		
		if retryCount < maxRetry {
			// 重新放入队列
			task.Retry = retryCount
			retryData, _ := json.Marshal(task)
			w.redis.LPush(ctx, w.taskType, retryData)
			
			w.db.Exec("UPDATE task_queue SET status = 'pending' WHERE id = ?", task.ID)
		}

		return err
	}

	// 处理成功
	err = w.db.Exec(`
		UPDATE task_queue SET status = 'completed', completed_at = NOW() 
		WHERE id = ?
	`, task.ID).Error

	return err
}

func (w *TestWorker) handleTask(ctx context.Context, task *Task) error {
	switch task.Type {
	case "ai_analysis":
		return w.handleAIAnalysis(ctx, task)
	case "notification":
		return w.handleNotification(ctx, task)
	case "config_sync":
		return w.handleConfigSync(ctx, task)
	case "failing_task":
		return w.handleFailingTask(ctx, task)
	case "concurrent_task":
		return w.handleConcurrentTask(ctx, task)
	case "priority_task":
		return w.handlePriorityTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (w *TestWorker) handleAIAnalysis(ctx context.Context, task *Task) error {
	// 模拟AI分析处理
	time.Sleep(100 * time.Millisecond)
	
	alertID := task.Payload["alert_id"].(string)
	if alertID == "" {
		return fmt.Errorf("missing alert_id")
	}

	// 模拟分析结果
	analysisResult := map[string]interface{}{
		"root_cause":    "High CPU usage due to memory leak",
		"suggestion":    "Restart the application service",
		"confidence":    0.85,
		"analysis_time": time.Now(),
	}

	// 这里应该保存分析结果到数据库
	_ = analysisResult

	return nil
}

func (w *TestWorker) handleNotification(ctx context.Context, task *Task) error {
	// 模拟通知发送
	time.Sleep(50 * time.Millisecond)
	
	channels := task.Payload["channels"].([]interface{})
	if len(channels) == 0 {
		return fmt.Errorf("no notification channels specified")
	}

	// 模拟发送到各个渠道
	for _, channel := range channels {
		channelName := channel.(string)
		// 模拟发送延迟
		time.Sleep(10 * time.Millisecond)
		_ = channelName
	}

	return nil
}

func (w *TestWorker) handleConfigSync(ctx context.Context, task *Task) error {
	// 模拟配置同步
	time.Sleep(200 * time.Millisecond)
	
	targets := task.Payload["targets"].([]interface{})
	if len(targets) == 0 {
		return fmt.Errorf("no sync targets specified")
	}

	// 模拟同步到各个目标
	for _, target := range targets {
		targetName := target.(string)
		time.Sleep(50 * time.Millisecond)
		_ = targetName
	}

	return nil
}

func (w *TestWorker) handleFailingTask(ctx context.Context, task *Task) error {
	shouldFail := task.Payload["should_fail"].(bool)
	if !shouldFail {
		return nil
	}

	failCount := int(task.Payload["fail_count"].(float64))
	if task.Retry < failCount {
		return fmt.Errorf("simulated failure %d", task.Retry+1)
	}

	// 第failCount+1次成功
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (w *TestWorker) handleConcurrentTask(ctx context.Context, task *Task) error {
	// 模拟并发任务处理
	taskIndex := int(task.Payload["task_index"].(float64))
	
	// 不同任务有不同的处理时间
	processingTime := time.Duration(50+taskIndex*10) * time.Millisecond
	time.Sleep(processingTime)
	
	return nil
}

func (w *TestWorker) handlePriorityTask(ctx context.Context, task *Task) error {
	// 模拟优先级任务处理
	time.Sleep(100 * time.Millisecond)
	return nil
}