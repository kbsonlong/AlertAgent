package test

import (
	"testing"

	"alert_agent/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestRuleModelValidation 测试规则模型验证
func TestRuleModelValidation(t *testing.T) {
	t.Run("创建有效规则", func(t *testing.T) {
		rule := &model.Rule{
			Name:        "测试规则",
			Expression:  "up == 0",
			Duration:    "5m",
			Severity:    "critical",
			Labels:      `{"team":"ops","service":"api"}`,
			Annotations: `{"summary":"服务不可用","description":"API服务已停止"}`,
			Targets:     `["http://localhost:9090"]`,
			Version:     1,
			Status:      true,
		}

		assert.NotNil(t, rule)
		assert.Equal(t, "测试规则", rule.Name)
		assert.Equal(t, "up == 0", rule.Expression)
		assert.Equal(t, "critical", rule.Severity)
		assert.True(t, rule.Status)
	})

	t.Run("测试标签操作", func(t *testing.T) {
		rule := &model.Rule{
			Name:       "标签测试规则",
			Expression: "cpu_usage > 80",
			Duration:   "1m",
			Severity:   "warning",
		}

		// 设置标签
		labels := map[string]string{
			"environment": "production",
			"team":        "backend",
			"priority":    "high",
		}
		rule.SetLabelsMap(labels)

		// 获取标签
		retrievedLabels, err := rule.GetLabelsMap()
		assert.NoError(t, err)
		assert.Equal(t, labels, retrievedLabels)
	})

	t.Run("测试注解操作", func(t *testing.T) {
		rule := &model.Rule{
			Name:       "注解测试规则",
			Expression: "memory_usage > 90",
			Severity:   "critical",
		}

		// 设置注解
		annotations := map[string]string{
			"summary":     "内存使用率过高",
			"description": "内存使用率超过90%，需要立即处理",
			"runbook":     "https://wiki.example.com/runbooks/memory",
		}
		rule.SetAnnotationsMap(annotations)

		// 获取注解
		retrievedAnnotations, err := rule.GetAnnotationsMap()
		assert.NoError(t, err)
		assert.Equal(t, annotations, retrievedAnnotations)
	})
}

// TestAlertModelValidation 测试告警模型验证
func TestAlertModelValidation(t *testing.T) {
	t.Run("创建告警", func(t *testing.T) {
		alert := &model.Alert{
			Name:    "CPU使用率过高",
			Level:   model.AlertLevelCritical,
			Status:  model.AlertStatusFiring,
			Content: "CPU使用率超过90%",
			Labels:  `{"instance":"server-01","job":"node-exporter"}`,
			RuleID:  1,
		}

		assert.NotNil(t, alert)
		assert.Equal(t, "CPU使用率过高", alert.Name)
		assert.Equal(t, model.AlertLevelCritical, alert.Level)
		assert.Equal(t, model.AlertStatusFiring, alert.Status)
	})

	t.Run("测试告警级别常量", func(t *testing.T) {
		assert.Equal(t, "info", model.AlertLevelInfo)
		assert.Equal(t, "warning", model.AlertLevelWarning)
		assert.Equal(t, "critical", model.AlertLevelCritical)
	})

	t.Run("测试告警状态常量", func(t *testing.T) {
		assert.Equal(t, "pending", model.AlertStatusPending)
		assert.Equal(t, "firing", model.AlertStatusFiring)
		assert.Equal(t, "resolved", model.AlertStatusResolved)
	})

	t.Run("测试标签操作", func(t *testing.T) {
		alert := &model.Alert{
			Name:   "磁盘空间告警",
			Level:  model.AlertLevelWarning,
			Status: model.AlertStatusFiring,
		}

		// 设置标签
		labels := map[string]string{
			"device":     "/dev/sda1",
			"mountpoint": "/",
			"usage":      "85%",
		}
		alert.SetLabelsMap(labels)

		// 获取标签
		retrievedLabels, err := alert.GetLabelsMap()
		assert.NoError(t, err)
		assert.Equal(t, labels, retrievedLabels)
	})
}

// TestProviderModelValidation 测试提供者模型验证
func TestProviderModelValidation(t *testing.T) {
	t.Run("创建Prometheus提供者", func(t *testing.T) {
		provider := &model.Provider{
			Name:       "Prometheus-01",
			Type:       model.ProviderTypePrometheus,
			Status:     model.ProviderStatusActive,
			Endpoint:   "http://prometheus:9090",
			AuthConfig: `{"type":"basic","username":"admin","password":"secret"}`,
			Labels:     `{"environment":"production","region":"us-west-1"}`,
		}

		assert.NotNil(t, provider)
		assert.Equal(t, "Prometheus-01", provider.Name)
		assert.Equal(t, model.ProviderTypePrometheus, provider.Type)
		assert.Equal(t, model.ProviderStatusActive, provider.Status)
	})

	t.Run("创建VictoriaMetrics提供者", func(t *testing.T) {
		provider := &model.Provider{
			Name:     "VictoriaMetrics-01",
			Type:     model.ProviderTypeVictoriaMetrics,
			Status:   model.ProviderStatusActive,
			Endpoint: "http://victoria-metrics:8428",
		}

		assert.Equal(t, model.ProviderTypeVictoriaMetrics, provider.Type)
	})

	t.Run("测试提供者类型常量", func(t *testing.T) {
		assert.Equal(t, "prometheus", model.ProviderTypePrometheus)
		assert.Equal(t, "victoriametrics", model.ProviderTypeVictoriaMetrics)
	})

	t.Run("测试提供者状态常量", func(t *testing.T) {
		assert.Equal(t, "active", model.ProviderStatusActive)
		assert.Equal(t, "inactive", model.ProviderStatusInactive)
	})

	t.Run("测试认证配置", func(t *testing.T) {
		provider := &model.Provider{
			Name:     "测试提供者",
			Type:     model.ProviderTypePrometheus,
			Endpoint: "http://test:9090",
		}

		// 设置认证配置
		authConfig := map[string]interface{}{
			"type":     "bearer",
			"token":    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			"timeout":  30,
			"insecure": false,
		}
		provider.SetAuthConfigMap(authConfig)

		// 获取认证配置
		retrievedConfig, err := provider.GetAuthConfigMap()
		assert.NoError(t, err)
		assert.Equal(t, authConfig, retrievedConfig)
	})
}

// TestUserModelValidation 测试用户模型验证
func TestUserModelValidation(t *testing.T) {
	t.Run("创建用户", func(t *testing.T) {
		user := &model.User{
			Username: "admin",
			Email:    "admin@example.com",
			FullName: "系统管理员",
			Password: "$2a$10$...", // bcrypt哈希
			Role:     model.UserRoleAdmin,
			Status:   model.UserStatusActive,
		}

		assert.NotNil(t, user)
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, "admin@example.com", user.Email)
		assert.Equal(t, model.UserRoleAdmin, user.Role)
		assert.Equal(t, model.UserStatusActive, user.Status)
	})

	t.Run("测试用户角色常量", func(t *testing.T) {
		assert.Equal(t, "admin", model.UserRoleAdmin)
		assert.Equal(t, "operator", model.UserRoleOperator)
		assert.Equal(t, "viewer", model.UserRoleViewer)
	})

	t.Run("测试用户状态常量", func(t *testing.T) {
		assert.Equal(t, "active", model.UserStatusActive)
		assert.Equal(t, "inactive", model.UserStatusInactive)
	})
}

// TestNotifyModelValidation 测试通知模型验证
func TestNotifyModelValidation(t *testing.T) {
	t.Run("创建通知模板", func(t *testing.T) {
		template := &model.NotifyTemplate{
			Name:        "告警通知模板",
			Type:        model.NotifyTypeEmail,
			Content:     "告警: {{.AlertName}} 状态: {{.Status}}",
			Description: "用于发送告警邮件的模板",
			Variables:   `["AlertName", "Status", "Severity"]`,
			Enabled:     true,
		}

		err := template.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "告警通知模板", template.Name)
		assert.Equal(t, model.NotifyTypeEmail, template.Type)
	})

	t.Run("测试通知类型常量", func(t *testing.T) {
		assert.Equal(t, "email", model.NotifyTypeEmail)
		assert.Equal(t, "sms", model.NotifyTypeSMS)
		assert.Equal(t, "webhook", model.NotifyTypeWebhook)
	})

	t.Run("测试通知状态常量", func(t *testing.T) {
		assert.Equal(t, "pending", model.NotifyStatusPending)
		assert.Equal(t, "sent", model.NotifyStatusSent)
		assert.Equal(t, "failed", model.NotifyStatusFailed)
		assert.Equal(t, "retrying", model.NotifyStatusRetrying)
	})

	t.Run("创建通知组", func(t *testing.T) {
		group := &model.NotifyGroup{
			Name:        "运维团队",
			Description: "负责系统运维的团队",
			Contacts:    `[{"type":"email","value":"ops@example.com"},{"type":"sms","value":"+1234567890"}]`,
			Members:     "admin,operator1,operator2",
			Channels:    "email,sms,webhook",
			Enabled:     true,
		}

		err := group.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "运维团队", group.Name)
	})

	t.Run("创建通知记录", func(t *testing.T) {
		record := &model.NotifyRecord{
			AlertID:    1,
			Type:       model.NotifyTypeEmail,
			Target:     "admin@example.com",
			Content:    "告警: CPU使用率过高",
			Status:     model.NotifyStatusPending,
			RetryCount: 0,
		}

		err := record.Validate()
		assert.NoError(t, err)
		assert.Equal(t, uint(1), record.AlertID)
		assert.Equal(t, model.NotifyTypeEmail, record.Type)
	})
}

// TestTaskQueueModelValidation 测试任务队列模型验证
func TestTaskQueueModelValidation(t *testing.T) {
	t.Run("测试任务状态常量", func(t *testing.T) {
		assert.Equal(t, "pending", model.TaskStatusPending)
		assert.Equal(t, "processing", model.TaskStatusProcessing)
		assert.Equal(t, "completed", model.TaskStatusCompleted)
		assert.Equal(t, "failed", model.TaskStatusFailed)
	})

	t.Run("测试任务类型常量", func(t *testing.T) {
		assert.Equal(t, "ai_analysis", model.TaskTypeAIAnalysis)
		assert.Equal(t, "notification", model.TaskTypeNotification)
		assert.Equal(t, "config_sync", model.TaskTypeConfigSync)
		assert.Equal(t, "rule_distribute", model.TaskTypeRuleDistribute)
	})

	t.Run("创建任务队列", func(t *testing.T) {
		task := &model.TaskQueue{
			QueueName: "default",
			TaskType:  model.TaskTypeNotification,
			Payload:   `{"alert_id":1,"recipients":["admin@example.com"]}`,
			Priority:  1,
			MaxRetry:  3,
			Status:    model.TaskStatusPending,
		}

		assert.NotNil(t, task)
		assert.Equal(t, "default", task.QueueName)
		assert.Equal(t, model.TaskTypeNotification, task.TaskType)
		assert.Equal(t, model.TaskStatusPending, task.Status)
	})

	t.Run("测试任务载荷操作", func(t *testing.T) {
		task := &model.TaskQueue{
			QueueName: "ai_analysis",
			TaskType:  model.TaskTypeAIAnalysis,
		}

		// 设置载荷
		payload := map[string]interface{}{
			"alert_id":    1,
			"content":     "CPU使用率过高",
			"model_name":  "llama2",
			"temperature": 0.7,
		}
		err := task.SetPayloadMap(payload)
		assert.NoError(t, err)

		// 获取载荷
		retrievedPayload, err := task.GetPayloadMap()
		assert.NoError(t, err)
		assert.Equal(t, payload, retrievedPayload)
	})
}

// TestKnowledgeModelValidation 测试知识库模型验证
func TestKnowledgeModelValidation(t *testing.T) {
	t.Run("创建知识库条目", func(t *testing.T) {
		knowledge := &model.Knowledge{
			Title:    "CPU使用率过高处理方案",
			Content:  "当CPU使用率超过80%时，需要检查进程列表...",
			Category: "运维手册",
			Tags:     "cpu,性能,监控",
			Source:   "manual",
			SourceID: 1,
			Summary:  "CPU使用率过高的处理步骤和解决方案",
		}

		assert.NotNil(t, knowledge)
		assert.Equal(t, "CPU使用率过高处理方案", knowledge.Title)
		assert.Equal(t, "运维手册", knowledge.Category)
		assert.Equal(t, "manual", knowledge.Source)
	})

	t.Run("测试知识库响应转换", func(t *testing.T) {
		knowledge := &model.Knowledge{
			ID:       1,
			Title:    "内存泄漏排查指南",
			Content:  "内存泄漏排查的详细步骤...",
			Category: "故障排查",
			Tags:     "内存,泄漏,调试",
			Source:   "wiki",
			SourceID: 2,
			Summary:  "内存泄漏问题的识别和解决方法",
		}

		response := knowledge.ToResponse()
		assert.Equal(t, knowledge.ID, response.ID)
		assert.Equal(t, knowledge.Title, response.Title)
		assert.Equal(t, knowledge.Content, response.Content)
		assert.Equal(t, knowledge.Category, response.Category)
		assert.Equal(t, knowledge.Source, response.Source)
	})
}

// TestSettingsModelValidation 测试设置模型验证
func TestSettingsModelValidation(t *testing.T) {
	t.Run("创建系统设置", func(t *testing.T) {
		settings := &model.Settings{
			OllamaEndpoint: "http://ollama:11434",
			OllamaModel:    "llama2:7b",
		}

		assert.NotNil(t, settings)
		assert.Equal(t, "http://ollama:11434", settings.OllamaEndpoint)
		assert.Equal(t, "llama2:7b", settings.OllamaModel)
	})
}
