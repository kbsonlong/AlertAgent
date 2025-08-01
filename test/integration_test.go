package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"alert_agent/internal/model"
)

// TestRuleIntegration 测试规则模型集成
func TestRuleIntegration(t *testing.T) {
	t.Run("规则创建和验证", func(t *testing.T) {
		rule := &model.Rule{
			Name:        "CPU使用率告警",
			Expression:  "cpu_usage > 80",
			Duration:    "5m",
			Severity:    "warning",
			Labels:      `{"team":"ops","service":"api"}`,
			Annotations: `{"summary":"CPU使用率过高","description":"CPU使用率超过80%"}`,
			Targets:     `["http://prometheus:9090"]`,
			Version:     "v1.0.0",
			Status:      "active",
		}

		// 验证基本字段
		assert.Equal(t, "CPU使用率告警", rule.Name)
		assert.Equal(t, "cpu_usage > 80", rule.Expression)
		assert.Equal(t, "warning", rule.Severity)
		assert.Equal(t, "active", rule.Status)
		assert.Equal(t, "v1.0.0", rule.Version)
	})

	t.Run("规则标签操作", func(t *testing.T) {
		rule := &model.Rule{}
		
		// 测试标签设置和获取
		labels := map[string]string{
			"environment": "production",
			"team":        "backend",
			"priority":    "high",
		}
		rule.SetLabelsMap(labels)
		
		retrievedLabels, err := rule.GetLabelsMap()
		assert.NoError(t, err)
		assert.Equal(t, labels, retrievedLabels)
	})

	t.Run("规则注解操作", func(t *testing.T) {
		rule := &model.Rule{}
		
		// 测试注解设置和获取
		annotations := map[string]string{
			"summary":     "内存使用率过高",
			"description": "内存使用率超过90%",
			"runbook":     "https://wiki.example.com/runbooks/memory",
		}
		rule.SetAnnotationsMap(annotations)
		
		retrievedAnnotations, err := rule.GetAnnotationsMap()
		assert.NoError(t, err)
		assert.Equal(t, annotations, retrievedAnnotations)
	})
}

// TestAlertIntegration 测试告警模型集成
func TestAlertIntegration(t *testing.T) {
	t.Run("告警创建和验证", func(t *testing.T) {
		alert := &model.Alert{
			Name:    "磁盘空间不足",
			Level:   "critical",
			Status:  "firing",
			Content: "磁盘使用率超过90%",
			Labels:  `{"device":"/dev/sda1","mountpoint":"/"}`,
			RuleID:  1,
		}

		// 验证基本字段
		assert.Equal(t, "磁盘空间不足", alert.Name)
		assert.Equal(t, "critical", alert.Level)
		assert.Equal(t, "firing", alert.Status)
		assert.Equal(t, uint(1), alert.RuleID)
	})

	t.Run("告警标签操作", func(t *testing.T) {
		alert := &model.Alert{}
		
		// 测试标签设置（直接设置JSON字符串）
		labelsJSON := `{"instance":"server-01","job":"node-exporter","severity":"high"}`
		alert.Labels = labelsJSON
		
		// 验证标签设置
		assert.Equal(t, labelsJSON, alert.Labels)
		assert.Contains(t, alert.Labels, "server-01")
		assert.Contains(t, alert.Labels, "node-exporter")
	})
}

// TestProviderIntegration 测试提供者模型集成
func TestProviderIntegration(t *testing.T) {
	t.Run("Prometheus提供者创建", func(t *testing.T) {
		provider := &model.Provider{
			Name:       "Prometheus-Production",
			Type:       "prometheus",
			Status:     "active",
			Endpoint:   "http://prometheus.monitoring.svc.cluster.local:9090",
			AuthConfig: `{"type":"bearer","token":"eyJhbGciOiJIUzI1NiJ9..."}`,
			Labels:     `{"environment":"production","region":"us-west-1"}`,
		}

		// 验证基本字段
		assert.Equal(t, "Prometheus-Production", provider.Name)
		assert.Equal(t, "prometheus", provider.Type)
		assert.Equal(t, "active", provider.Status)
		assert.Contains(t, provider.Endpoint, "prometheus")
	})

	t.Run("VictoriaMetrics提供者创建", func(t *testing.T) {
		provider := &model.Provider{
			Name:     "VictoriaMetrics-Cluster",
			Type:     "victoriametrics",
			Status:   "active",
			Endpoint: "http://vmselect.monitoring.svc.cluster.local:8481",
		}

		assert.Equal(t, "victoriametrics", provider.Type)
		assert.Contains(t, provider.Endpoint, "vmselect")
	})

	t.Run("提供者认证配置操作", func(t *testing.T) {
		provider := &model.Provider{}
		
		// 测试认证配置设置（直接设置JSON字符串）
		authConfigJSON := `{"type":"basic","username":"admin","password":"secret123","timeout":30,"verify_ssl":false}`
		provider.AuthConfig = authConfigJSON
		provider.AuthType = "basic"
		
		// 验证认证配置
		assert.Equal(t, "basic", provider.AuthType)
		assert.Equal(t, authConfigJSON, provider.AuthConfig)
		assert.Contains(t, provider.AuthConfig, "admin")
	})
}

// TestUserIntegration 测试用户模型集成
func TestUserIntegration(t *testing.T) {
	t.Run("管理员用户创建", func(t *testing.T) {
		user := &model.User{
			Username: "admin",
			Email:    "admin@alertagent.com",
			FullName: "系统管理员",
			Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye...", // bcrypt哈希
			Role:     "admin",
			Status:   "active",
		}

		// 验证基本字段
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, "admin@alertagent.com", user.Email)
		assert.Equal(t, "admin", user.Role)
		assert.Equal(t, "active", user.Status)
		assert.Contains(t, user.Password, "$2a$10$")
	})

	t.Run("操作员用户创建", func(t *testing.T) {
		user := &model.User{
			Username: "operator",
			Email:    "ops@alertagent.com",
			FullName: "运维操作员",
			Role:     "operator",
			Status:   "active",
		}

		assert.Equal(t, "operator", user.Role)
		assert.Equal(t, "active", user.Status)
	})

	t.Run("查看者用户创建", func(t *testing.T) {
		user := &model.User{
			Username: "viewer",
			Email:    "view@alertagent.com",
			FullName: "只读用户",
			Role:     "viewer",
			Status:   "active",
		}

		assert.Equal(t, "viewer", user.Role)
	})
}

// TestNotificationIntegration 测试通知模型集成
func TestNotificationIntegration(t *testing.T) {
	t.Run("邮件通知模板创建", func(t *testing.T) {
		template := &model.NotifyTemplate{
			Name:        "告警邮件模板",
			Type:        "email",
			Content:     "告警名称: {{.AlertName}}\n严重级别: {{.Severity}}\n告警内容: {{.Content}}",
			Description: "用于发送告警邮件的标准模板",
			Variables:   `["AlertName", "Severity", "Content", "Timestamp"]`,
			Enabled:     true,
		}

		// 验证模板
		err := template.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "email", template.Type)
		assert.True(t, template.Enabled)
	})

	t.Run("Webhook通知模板创建", func(t *testing.T) {
		template := &model.NotifyTemplate{
			Name:    "Slack Webhook模板",
			Type:    "webhook",
			Content: `{"text":"🚨 告警: {{.AlertName}}\n级别: {{.Severity}}\n时间: {{.Timestamp}}"}`,
			Enabled: true,
		}

		err := template.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "webhook", template.Type)
	})

	t.Run("通知组创建", func(t *testing.T) {
		group := &model.NotifyGroup{
			Name:        "运维团队",
			Description: "负责系统运维和故障处理的团队",
			Contacts:    `[{"type":"email","value":"ops@company.com"},{"type":"webhook","value":"https://hooks.slack.com/..."}]`,
			Members:     "admin,operator1,operator2",
			Channels:    "email,webhook",
			Enabled:     true,
		}

		err := group.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "运维团队", group.Name)
		assert.True(t, group.Enabled)
	})

	t.Run("通知记录创建", func(t *testing.T) {
		record := &model.NotifyRecord{
			AlertID:    1,
			Type:       "email",
			Target:     "admin@company.com",
			Content:    "告警: CPU使用率过高 - 当前值: 95%",
			Status:     "pending",
			RetryCount: 0,
		}

		err := record.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "pending", record.Status)
		assert.Equal(t, 0, record.RetryCount)
	})
}

// TestTaskQueueIntegration 测试任务队列模型集成
func TestTaskQueueIntegration(t *testing.T) {
	t.Run("AI分析任务创建", func(t *testing.T) {
		task := &model.TaskQueue{
			QueueName:   "ai_analysis",
			TaskType:    "ai_analysis",
			Payload:     `{"alert_id":1,"content":"CPU使用率持续过高","model":"llama2"}`,
			Priority:    2,
			RetryCount:  0,
			MaxRetry:    3,
			Status:      "pending",
			ScheduledAt: time.Now(),
		}

		// 验证任务字段
		assert.Equal(t, "ai_analysis", task.TaskType)
		assert.Equal(t, "pending", task.Status)
		assert.Equal(t, 3, task.MaxRetry)
		assert.Equal(t, 2, task.Priority)
	})

	t.Run("通知任务创建", func(t *testing.T) {
		task := &model.TaskQueue{
			QueueName: "notification",
			TaskType:  "notification",
			Payload:   `{"alert_id":2,"recipients":["admin@company.com","ops@company.com"],"template":"email_alert"}`,
			Priority:  1,
			Status:    "pending",
		}

		assert.Equal(t, "notification", task.TaskType)
		assert.Equal(t, 1, task.Priority)
	})

	t.Run("任务载荷操作", func(t *testing.T) {
		task := &model.TaskQueue{}
		
		// 设置载荷
		payload := map[string]interface{}{
			"alert_id":     123,
			"analysis_type": "anomaly_detection",
			"parameters": map[string]interface{}{
				"threshold": 0.8,
				"window":    "1h",
			},
		}
		err := task.SetPayloadMap(payload)
		assert.NoError(t, err)
		
		// 获取载荷
		retrievedPayload, err := task.GetPayloadMap()
		assert.NoError(t, err)
		// JSON反序列化会将数字转换为float64，需要进行类型转换比较
		assert.Equal(t, float64(123), retrievedPayload["alert_id"])
		assert.Equal(t, "anomaly_detection", retrievedPayload["analysis_type"])
		params, ok := retrievedPayload["parameters"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0.8, params["threshold"])
		assert.Equal(t, "1h", params["window"])
	})
}

// TestKnowledgeIntegration 测试知识库模型集成
func TestKnowledgeIntegration(t *testing.T) {
	t.Run("运维知识条目创建", func(t *testing.T) {
		knowledge := &model.Knowledge{
			Title:    "CPU使用率过高故障排查手册",
			Content:  "1. 检查top命令输出\n2. 分析进程CPU占用\n3. 检查系统负载\n4. 查看系统日志",
			Category: "故障排查",
			Tags:     "cpu,性能,监控,故障排查",
			Source:   "manual",
			SourceID: 1,
			Summary:  "CPU使用率过高时的系统性排查方法和解决方案",
		}

		// 验证知识条目
		assert.Equal(t, "故障排查", knowledge.Category)
		assert.Equal(t, "manual", knowledge.Source)
		assert.Contains(t, knowledge.Content, "检查top命令")
		assert.Contains(t, knowledge.Tags, "cpu")
	})

	t.Run("知识库响应转换", func(t *testing.T) {
		knowledge := &model.Knowledge{
			ID:       1,
			Title:    "内存泄漏检测方法",
			Content:  "使用valgrind、pprof等工具检测内存泄漏",
			Category: "开发调试",
			Tags:     "内存,泄漏,调试工具",
			Source:   "wiki",
			SourceID: 2,
			Summary:  "内存泄漏的检测工具和方法介绍",
		}

		response := knowledge.ToResponse()
		assert.Equal(t, knowledge.ID, response.ID)
		assert.Equal(t, knowledge.Title, response.Title)
		assert.Equal(t, knowledge.Category, response.Category)
		assert.Equal(t, knowledge.Source, response.Source)
		assert.Equal(t, float32(0), response.Similarity) // 默认值
	})
}

// TestConfigSyncIntegration 测试配置同步模型集成
func TestConfigSyncIntegration(t *testing.T) {
	t.Run("配置同步状态创建", func(t *testing.T) {
		syncStatus := &model.ConfigSyncStatus{
			ClusterID:  "prod-cluster-01",
			ConfigType: "prometheus_rules",
			ConfigHash: "sha256:abc123def456...",
			SyncStatus: "success",
			SyncTime:   &time.Time{},
		}

		// 验证同步状态
		assert.Equal(t, "prod-cluster-01", syncStatus.ClusterID)
		assert.Equal(t, "prometheus_rules", syncStatus.ConfigType)
		assert.Equal(t, "success", syncStatus.SyncStatus)
		assert.Contains(t, syncStatus.ConfigHash, "sha256:")
	})

	t.Run("配置同步触发记录创建", func(t *testing.T) {
		trigger := &model.ConfigSyncTrigger{
			ClusterID:  "prod-cluster-01",
			ConfigType: "alertmanager_config",
			TriggerBy:  "admin",
			Reason:     "规则更新",
			Status:     "pending",
		}

		assert.Equal(t, "admin", trigger.TriggerBy)
		assert.Equal(t, "规则更新", trigger.Reason)
		assert.Equal(t, "pending", trigger.Status)
	})
}

// TestSettingsIntegration 测试设置模型集成
func TestSettingsIntegration(t *testing.T) {
	t.Run("系统设置创建", func(t *testing.T) {
		settings := &model.Settings{
			OllamaEndpoint: "http://ollama.ai.svc.cluster.local:11434",
			OllamaModel:    "llama2:13b-chat",
		}

		// 验证设置
		assert.Contains(t, settings.OllamaEndpoint, "ollama")
		assert.Contains(t, settings.OllamaModel, "llama2")
		assert.Contains(t, settings.OllamaEndpoint, ":11434")
	})
}

// TestModelRelationships 测试模型关系
func TestModelRelationships(t *testing.T) {
	t.Run("规则和告警关系", func(t *testing.T) {
		// 创建规则
		rule := &model.Rule{
			Name:       "内存使用率告警",
			Expression: "memory_usage > 90",
			Severity:   "critical",
		}
		// 假设规则ID为1
		ruleID := uint(1)

		// 创建基于该规则的告警
		alert := &model.Alert{
			Name:    "内存使用率过高",
			Level:   "critical",
			Status:  "firing",
			Content: "内存使用率达到95%",
			RuleID:  ruleID,
		}

		// 验证关系
		assert.Equal(t, ruleID, alert.RuleID)
		assert.Equal(t, rule.Severity, alert.Level)
	})

	t.Run("告警和通知关系", func(t *testing.T) {
		// 创建告警
		alertID := uint(1)

		// 创建通知记录
		notifyRecord := &model.NotifyRecord{
			AlertID: alertID,
			Type:    "email",
			Target:  "admin@company.com",
			Content: "告警通知: 内存使用率过高",
			Status:  "sent",
		}

		// 验证关系
		assert.Equal(t, alertID, notifyRecord.AlertID)
		assert.Equal(t, "sent", notifyRecord.Status)
	})
}