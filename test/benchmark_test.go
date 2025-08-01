package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"alert_agent/internal/model"
)

// BenchmarkRuleOperations 基准测试规则操作性能
func BenchmarkRuleOperations(b *testing.B) {
	b.Run("规则创建", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rule := &model.Rule{
				Name:        fmt.Sprintf("规则-%d", i),
				Expression:  "cpu_usage > 80",
				Duration:    "5m",
				Severity:    "warning",
				Labels:      `{"team":"ops","service":"api"}`,
				Annotations: `{"summary":"CPU使用率过高"}`,
				Targets:     `["http://prometheus:9090"]`,
				Version:     "v1.0.0",
				Status:      "active",
			}
			_ = rule
		}
	})

	b.Run("规则标签操作", func(b *testing.B) {
		rule := &model.Rule{}
		labels := map[string]string{
			"environment": "production",
			"team":        "backend",
			"priority":    "high",
			"service":     "api-gateway",
			"region":      "us-west-1",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rule.SetLabelsMap(labels)
			_, _ = rule.GetLabelsMap()
		}
	})

	b.Run("规则注解操作", func(b *testing.B) {
		rule := &model.Rule{}
		annotations := map[string]string{
			"summary":     "系统性能告警",
			"description": "CPU使用率持续超过阈值，需要立即处理",
			"runbook":     "https://wiki.example.com/runbooks/cpu-high",
			"dashboard":   "https://grafana.example.com/d/cpu-dashboard",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rule.SetAnnotationsMap(annotations)
			_, _ = rule.GetAnnotationsMap()
		}
	})

	b.Run("规则目标操作", func(b *testing.B) {
		rule := &model.Rule{}
		targets := []string{
			"http://prometheus-1:9090",
			"http://prometheus-2:9090",
			"http://prometheus-3:9090",
			"http://victoriametrics:8428",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rule.SetTargetsList(targets)
			_, _ = rule.GetTargetsList()
		}
	})
}

// BenchmarkAlertOperations 基准测试告警操作性能
func BenchmarkAlertOperations(b *testing.B) {
	b.Run("告警创建", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			alert := &model.Alert{
				Name:     fmt.Sprintf("告警-%d", i),
				Title:    fmt.Sprintf("系统告警-%d", i),
				Level:    "critical",
				Status:   "new",
				Source:   "prometheus",
				Content:  "CPU使用率超过90%",
				Labels:   `{"instance":"server-01","job":"node-exporter"}`,
				RuleID:   1,
				Severity: "high",
			}
			_ = alert
		}
	})

	b.Run("告警验证", func(b *testing.B) {
		alert := &model.Alert{
			Name:     "CPU告警",
			Title:    "CPU使用率过高",
			Level:    "critical",
			Status:   "new",
			Source:   "prometheus",
			Content:  "CPU使用率达到95%",
			Severity: "high",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = alert.Validate()
		}
	})

	b.Run("告警序列化", func(b *testing.B) {
		alert := &model.Alert{
			Name:     "内存告警",
			Title:    "内存使用率过高",
			Level:    "critical",
			Status:   "new",
			Source:   "prometheus",
			Content:  "内存使用率达到90%",
			Severity: "high",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = alert.MarshalBinary()
		}
	})

	b.Run("告警反序列化", func(b *testing.B) {
		alert := &model.Alert{
			Name:     "磁盘告警",
			Title:    "磁盘空间不足",
			Level:    "critical",
			Status:   "new",
			Source:   "prometheus",
			Content:  "磁盘使用率达到95%",
			Severity: "high",
		}
		data, _ := alert.MarshalBinary()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newAlert := &model.Alert{}
			_ = newAlert.UnmarshalBinary(data)
		}
	})

	b.Run("告警分析结果操作", func(b *testing.B) {
		alert := &model.Alert{}
		analysisResult := map[string]interface{}{
			"confidence":    0.95,
			"category":      "performance",
			"root_cause":    "high_cpu_usage",
			"impact_level":  "high",
			"affected_services": []string{"api-gateway", "user-service"},
			"recommendations": []string{"scale_up", "optimize_queries"},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = alert.SetAnalysisResultMap(analysisResult)
			_, _ = alert.GetAnalysisResultMap()
		}
	})
}

// BenchmarkProviderOperations 基准测试提供者操作性能
func BenchmarkProviderOperations(b *testing.B) {
	b.Run("提供者创建", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider := &model.Provider{
				Name:        fmt.Sprintf("Prometheus-%d", i),
				Type:        "prometheus",
				Status:      "active",
				Description: "生产环境Prometheus实例",
				Endpoint:    "http://prometheus.monitoring.svc.cluster.local:9090",
				AuthType:    "bearer",
				AuthConfig:  `{"token":"eyJhbGciOiJIUzI1NiJ9..."}`,
				Labels:      `{"environment":"production","region":"us-west-1"}`,
			}
			_ = provider
		}
	})

	b.Run("提供者验证", func(b *testing.B) {
		provider := &model.Provider{
			Name:     "TestProvider",
			Type:     "prometheus",
			Status:   "active",
			Endpoint: "http://prometheus:9090",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.Validate()
		}
	})

	b.Run("提供者序列化", func(b *testing.B) {
		provider := &model.Provider{
			Name:        "VictoriaMetrics",
			Type:        "victoriametrics",
			Status:      "active",
			Description: "高性能时序数据库",
			Endpoint:    "http://vmselect:8481",
			AuthType:    "basic",
			AuthConfig:  `{"username":"admin","password":"secret"}`,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = provider.MarshalBinary()
		}
	})
}

// BenchmarkTaskQueueOperations 基准测试任务队列操作性能
func BenchmarkTaskQueueOperations(b *testing.B) {
	b.Run("任务创建", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := &model.TaskQueue{
				QueueName:   "ai_analysis",
				TaskType:    "ai_analysis",
				Payload:     fmt.Sprintf(`{"alert_id":%d,"content":"CPU使用率过高"}`, i),
				Priority:    2,
				RetryCount:  0,
				MaxRetry:    3,
				Status:      "pending",
				ScheduledAt: time.Now(),
			}
			_ = task
		}
	})

	b.Run("任务载荷操作", func(b *testing.B) {
		task := &model.TaskQueue{}
		payload := map[string]interface{}{
			"alert_id":       123,
			"analysis_type":  "anomaly_detection",
			"model":          "llama2",
			"temperature":    0.7,
			"max_tokens":     1000,
			"parameters": map[string]interface{}{
				"threshold": 0.8,
				"window":    "1h",
				"features":  []string{"cpu", "memory", "disk", "network"},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = task.SetPayloadMap(payload)
			_, _ = task.GetPayloadMap()
		}
	})
}

// BenchmarkNotificationOperations 基准测试通知操作性能
func BenchmarkNotificationOperations(b *testing.B) {
	b.Run("通知模板验证", func(b *testing.B) {
		template := &model.NotifyTemplate{
			Name:        "告警邮件模板",
			Type:        "email",
			Content:     "告警: {{.AlertName}} - 级别: {{.Severity}}",
			Description: "标准告警邮件模板",
			Variables:   `["AlertName", "Severity", "Content"]`,
			Enabled:     true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = template.Validate()
		}
	})

	b.Run("通知组验证", func(b *testing.B) {
		group := &model.NotifyGroup{
			Name:        "运维团队",
			Description: "负责系统运维的团队",
			Contacts:    `[{"type":"email","value":"ops@company.com"}]`,
			Members:     "admin,operator1,operator2",
			Channels:    "email,webhook",
			Enabled:     true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = group.Validate()
		}
	})

	b.Run("通知记录验证", func(b *testing.B) {
		record := &model.NotifyRecord{
			AlertID:    1,
			Type:       "email",
			Target:     "admin@company.com",
			Content:    "告警通知: CPU使用率过高",
			Status:     "pending",
			RetryCount: 0,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = record.Validate()
		}
	})
}

// BenchmarkKnowledgeOperations 基准测试知识库操作性能
func BenchmarkKnowledgeOperations(b *testing.B) {
	b.Run("知识条目创建", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			knowledge := &model.Knowledge{
				Title:    fmt.Sprintf("故障排查手册-%d", i),
				Content:  "详细的故障排查步骤和解决方案",
				Category: "故障排查",
				Tags:     "cpu,性能,监控",
				Source:   "manual",
				SourceID: uint(i),
				Summary:  "系统性能问题的排查方法",
			}
			_ = knowledge
		}
	})

	b.Run("知识响应转换", func(b *testing.B) {
		knowledge := &model.Knowledge{
			ID:       1,
			Title:    "内存泄漏检测",
			Content:  "使用各种工具检测内存泄漏的方法",
			Category: "开发调试",
			Tags:     "内存,泄漏,调试",
			Source:   "wiki",
			SourceID: 2,
			Summary:  "内存泄漏检测工具和方法",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = knowledge.ToResponse()
		}
	})
}

// BenchmarkJSONOperations 基准测试JSON操作性能
func BenchmarkJSONOperations(b *testing.B) {
	b.Run("大型JSON序列化", func(b *testing.B) {
		// 创建复杂的数据结构
		data := map[string]interface{}{
			"alerts": []map[string]interface{}{
				{
					"id":       1,
					"name":     "CPU告警",
					"level":    "critical",
					"status":   "firing",
					"labels":   map[string]string{"instance": "server-01", "job": "node-exporter"},
					"metrics": []float64{95.5, 96.2, 94.8, 97.1, 95.9},
				},
				{
					"id":       2,
					"name":     "内存告警",
					"level":    "warning",
					"status":   "pending",
					"labels":   map[string]string{"instance": "server-02", "job": "node-exporter"},
					"metrics": []float64{85.1, 86.3, 84.7, 87.2, 85.8},
				},
			},
			"metadata": map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"version":   "v1.0.0",
				"source":    "prometheus",
				"cluster":   "production",
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})

	b.Run("大型JSON反序列化", func(b *testing.B) {
		jsonData := `{
			"alerts": [
				{
					"id": 1,
					"name": "CPU告警",
					"level": "critical",
					"status": "firing",
					"labels": {"instance": "server-01", "job": "node-exporter"},
					"metrics": [95.5, 96.2, 94.8, 97.1, 95.9]
				},
				{
					"id": 2,
					"name": "内存告警",
					"level": "warning",
					"status": "pending",
					"labels": {"instance": "server-02", "job": "node-exporter"},
					"metrics": [85.1, 86.3, 84.7, 87.2, 85.8]
				}
			],
			"metadata": {
				"timestamp": 1640995200,
				"version": "v1.0.0",
				"source": "prometheus",
				"cluster": "production"
			}
		}`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result map[string]interface{}
			_ = json.Unmarshal([]byte(jsonData), &result)
		}
	})
}