package test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"alert_agent/internal/model"
)

// MockDataGenerator 模拟数据生成器
type MockDataGenerator struct {
	random *rand.Rand
}

// NewMockDataGenerator 创建新的模拟数据生成器
func NewMockDataGenerator() *MockDataGenerator {
	return &MockDataGenerator{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRule 生成模拟规则数据
func (g *MockDataGenerator) GenerateRule(index int) *model.Rule {
	severities := []string{"critical", "warning", "info"}
	statuses := []string{"active", "inactive", "pending"}
	services := []string{"api-gateway", "user-service", "order-service", "payment-service", "notification-service"}
	teams := []string{"backend", "frontend", "devops", "sre", "platform"}
	environments := []string{"production", "staging", "development"}

	labels := map[string]string{
		"service":     services[g.random.Intn(len(services))],
		"team":        teams[g.random.Intn(len(teams))],
		"environment": environments[g.random.Intn(len(environments))],
		"region":      fmt.Sprintf("us-west-%d", g.random.Intn(3)+1),
	}
	labelsJSON, _ := json.Marshal(labels)

	annotations := map[string]string{
		"summary":     fmt.Sprintf("规则%d告警", index),
		"description": fmt.Sprintf("这是第%d个测试规则的详细描述", index),
		"runbook":     fmt.Sprintf("https://wiki.example.com/runbooks/rule-%d", index),
		"dashboard":   fmt.Sprintf("https://grafana.example.com/d/rule-%d", index),
	}
	annotationsJSON, _ := json.Marshal(annotations)

	targets := []string{
		fmt.Sprintf("http://prometheus-%d:9090", g.random.Intn(3)+1),
		fmt.Sprintf("http://victoriametrics-%d:8428", g.random.Intn(2)+1),
	}
	targetsJSON, _ := json.Marshal(targets)

	return &model.Rule{
		ID:          fmt.Sprintf("rule-%d", index),
		Name:        fmt.Sprintf("规则-%d", index),
		Expression:  fmt.Sprintf("cpu_usage > %d", 70+g.random.Intn(30)),
		Duration:    fmt.Sprintf("%dm", g.random.Intn(10)+1),
		Severity:    severities[g.random.Intn(len(severities))],
		Labels:      string(labelsJSON),
		Annotations: string(annotationsJSON),
		Targets:     string(targetsJSON),
		Version:     fmt.Sprintf("v1.%d.%d", g.random.Intn(10), g.random.Intn(10)),
		Status:      statuses[g.random.Intn(len(statuses))],
		CreatedAt:   time.Now().Add(-time.Duration(g.random.Intn(30)) * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-time.Duration(g.random.Intn(7)) * 24 * time.Hour),
	}
}

// GenerateAlert 生成模拟告警数据
func (g *MockDataGenerator) GenerateAlert(index int, ruleID uint) *model.Alert {
	levels := []string{"critical", "high", "medium", "low"}
	statuses := []string{"new", "acknowledged", "resolved"}
	sources := []string{"prometheus", "victoriametrics", "grafana", "custom"}
	severities := []string{"critical", "warning", "info"}
	instances := []string{"server-01", "server-02", "server-03", "k8s-node-01", "k8s-node-02"}
	jobs := []string{"node-exporter", "cadvisor", "kube-state-metrics", "prometheus", "alertmanager"}

	labels := map[string]string{
		"instance": instances[g.random.Intn(len(instances))],
		"job":      jobs[g.random.Intn(len(jobs))],
		"severity": severities[g.random.Intn(len(severities))],
		"cluster":  fmt.Sprintf("prod-cluster-%d", g.random.Intn(3)+1),
	}
	labelsJSON, _ := json.Marshal(labels)

	analysisResult := map[string]interface{}{
		"confidence":    0.8 + g.random.Float64()*0.2,
		"category":      []string{"performance", "availability", "security", "capacity"}[g.random.Intn(4)],
		"root_cause":    []string{"high_cpu", "memory_leak", "disk_full", "network_issue"}[g.random.Intn(4)],
		"impact_level":  []string{"low", "medium", "high", "critical"}[g.random.Intn(4)],
		"affected_services": []string{"api-gateway", "user-service", "order-service"},
		"recommendations": []string{"scale_up", "optimize_queries", "increase_memory"},
	}
	analysisResultJSON, _ := json.Marshal(analysisResult)

	similarAlerts := []model.SimilarAlert{
		{
			Alert: model.Alert{
				ID:   uint(g.random.Intn(1000) + 1),
				Name: fmt.Sprintf("相似告警-%d", g.random.Intn(100)),
			},
			Similarity: 0.7 + g.random.Float64()*0.3,
		},
	}
	similarAlertsJSON, _ := json.Marshal(similarAlerts)

	return &model.Alert{
		ID:                   uint(index),
		Name:                 fmt.Sprintf("告警-%d", index),
		Title:                fmt.Sprintf("系统告警-%d", index),
		Level:                levels[g.random.Intn(len(levels))],
		Status:               statuses[g.random.Intn(len(statuses))],
		Source:               sources[g.random.Intn(len(sources))],
		Content:              fmt.Sprintf("告警内容: 系统指标异常，当前值: %d%%", 80+g.random.Intn(20)),
		Labels:               string(labelsJSON),
		RuleID:               ruleID,
		TemplateID:           uint(g.random.Intn(10) + 1),
		GroupID:              uint(g.random.Intn(5) + 1),
		Handler:              []string{"admin", "operator1", "operator2", ""}[g.random.Intn(4)],
		HandleNote:           fmt.Sprintf("处理备注-%d", index),
		Analysis:             fmt.Sprintf("AI分析结果: 这是第%d个告警的分析", index),
		AnalysisStatus:       []string{"pending", "processing", "completed", "failed"}[g.random.Intn(4)],
		AnalysisResult:       string(analysisResultJSON),
		AISummary:            fmt.Sprintf("AI摘要: 告警%d的智能分析摘要", index),
		SimilarAlerts:        string(similarAlertsJSON),
		ResolutionSuggestion: fmt.Sprintf("解决建议: 针对告警%d的处理建议", index),
		Fingerprint:          fmt.Sprintf("fp-%d-%d", index, time.Now().Unix()),
		NotifyCount:          g.random.Intn(5),
		Severity:             severities[g.random.Intn(len(severities))],
		CreatedAt:            time.Now().Add(-time.Duration(g.random.Intn(24)) * time.Hour),
		UpdatedAt:            time.Now().Add(-time.Duration(g.random.Intn(6)) * time.Hour),
	}
}

// GenerateProvider 生成模拟提供者数据
func (g *MockDataGenerator) GenerateProvider(index int) *model.Provider {
	types := []string{"prometheus", "victoriametrics"}
	statuses := []string{"active", "inactive"}
	authTypes := []string{"none", "basic", "bearer", "oauth2"}
	environments := []string{"production", "staging", "development"}
	regions := []string{"us-west-1", "us-west-2", "us-east-1", "eu-west-1"}

	providerType := types[g.random.Intn(len(types))]
	authType := authTypes[g.random.Intn(len(authTypes))]

	var endpoint string
	if providerType == "prometheus" {
		endpoint = fmt.Sprintf("http://prometheus-%d.monitoring.svc.cluster.local:9090", index)
	} else {
		endpoint = fmt.Sprintf("http://vmselect-%d.monitoring.svc.cluster.local:8481", index)
	}

	var authConfig string
	switch authType {
	case "basic":
		authConfig = `{"username":"admin","password":"secret123","timeout":30}`
	case "bearer":
		authConfig = `{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...","timeout":30}`
	case "oauth2":
		authConfig = `{"client_id":"client123","client_secret":"secret456","token_url":"https://auth.example.com/token"}`
	default:
		authConfig = ""
	}

	labels := map[string]string{
		"environment": environments[g.random.Intn(len(environments))],
		"region":      regions[g.random.Intn(len(regions))],
		"cluster":     fmt.Sprintf("cluster-%d", g.random.Intn(5)+1),
		"version":     fmt.Sprintf("v2.%d.%d", g.random.Intn(10), g.random.Intn(10)),
	}
	labelsJSON, _ := json.Marshal(labels)

	lastCheck := time.Now().Add(-time.Duration(g.random.Intn(60)) * time.Minute)
	var lastError string
	if g.random.Float64() < 0.1 { // 10%的概率有错误
		lastError = []string{
			"connection timeout",
			"authentication failed",
			"invalid endpoint",
			"rate limit exceeded",
		}[g.random.Intn(4)]
	}

	return &model.Provider{
		ID:          uint(index),
		Name:        fmt.Sprintf("%s-Provider-%d", providerType, index),
		Type:        providerType,
		Status:      statuses[g.random.Intn(len(statuses))],
		Description: fmt.Sprintf("第%d个%s数据源实例", index, providerType),
		Endpoint:    endpoint,
		AuthType:    authType,
		AuthConfig:  authConfig,
		Labels:      string(labelsJSON),
		LastCheck:   &lastCheck,
		LastError:   lastError,
		CreatedAt:   time.Now().Add(-time.Duration(g.random.Intn(90)) * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-time.Duration(g.random.Intn(7)) * 24 * time.Hour),
	}
}

// GenerateUser 生成模拟用户数据
func (g *MockDataGenerator) GenerateUser(index int) *model.User {
	roles := []string{"admin", "operator", "viewer"}
	statuses := []string{"active", "inactive", "pending"}
	usernames := []string{"admin", "operator", "viewer", "developer", "manager", "analyst"}
	domains := []string{"company.com", "example.org", "test.net"}

	username := fmt.Sprintf("%s%d", usernames[g.random.Intn(len(usernames))], index)
	email := fmt.Sprintf("%s@%s", username, domains[g.random.Intn(len(domains))])
	fullName := fmt.Sprintf("用户%d", index)

	return &model.User{
		Username: username,
		Email:    email,
		FullName: fullName,
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye...", // bcrypt哈希
		Role:     roles[g.random.Intn(len(roles))],
		Status:   statuses[g.random.Intn(len(statuses))],
		CreatedAt: time.Now().Add(-time.Duration(g.random.Intn(30)) * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(g.random.Intn(7)) * 24 * time.Hour),
	}
}

// GenerateNotifyTemplate 生成模拟通知模板数据
func (g *MockDataGenerator) GenerateNotifyTemplate(index int) *model.NotifyTemplate {
	types := []string{"email", "webhook", "sms", "slack"}
	templateType := types[g.random.Intn(len(types))]

	var content string
	var variables []string

	switch templateType {
	case "email":
		content = "告警通知\n\n告警名称: {{.AlertName}}\n严重级别: {{.Severity}}\n告警时间: {{.Timestamp}}\n告警内容: {{.Content}}\n\n请及时处理。"
		variables = []string{"AlertName", "Severity", "Timestamp", "Content"}
	case "webhook":
		content = `{"text":"🚨 告警: {{.AlertName}}\n级别: {{.Severity}}\n时间: {{.Timestamp}}","channel":"#alerts"}`
		variables = []string{"AlertName", "Severity", "Timestamp"}
	case "sms":
		content = "告警: {{.AlertName}} - {{.Severity}} - {{.Timestamp}}"
		variables = []string{"AlertName", "Severity", "Timestamp"}
	case "slack":
		content = `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*告警:* {{.AlertName}}\n*级别:* {{.Severity}}"}}]}`
		variables = []string{"AlertName", "Severity"}
	}

	variablesJSON, _ := json.Marshal(variables)

	return &model.NotifyTemplate{
		Name:        fmt.Sprintf("%s通知模板-%d", templateType, index),
		Type:        templateType,
		Content:     content,
		Description: fmt.Sprintf("用于%s通知的标准模板", templateType),
		Variables:   string(variablesJSON),
		Enabled:     g.random.Float64() > 0.2, // 80%概率启用
	}
}

// GenerateNotifyGroup 生成模拟通知组数据
func (g *MockDataGenerator) GenerateNotifyGroup(index int) *model.NotifyGroup {
	teams := []string{"运维团队", "开发团队", "测试团队", "安全团队", "产品团队"}
	channels := [][]string{
		{"email", "webhook"},
		{"email", "sms"},
		{"webhook", "slack"},
		{"email", "webhook", "sms"},
	}

	teamName := teams[g.random.Intn(len(teams))]
	selectedChannels := channels[g.random.Intn(len(channels))]

	contacts := []map[string]string{
		{"type": "email", "value": fmt.Sprintf("team%d@company.com", index)},
		{"type": "webhook", "value": fmt.Sprintf("https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX%d", index)},
	}
	contactsJSON, _ := json.Marshal(contacts)

	members := []string{
		fmt.Sprintf("admin%d", index),
		fmt.Sprintf("operator%d", index),
		fmt.Sprintf("manager%d", index),
	}

	return &model.NotifyGroup{
		Name:        fmt.Sprintf("%s-%d", teamName, index),
		Description: fmt.Sprintf("%s的通知组配置", teamName),
		Contacts:    string(contactsJSON),
		Members:     fmt.Sprintf("%s", members[0:g.random.Intn(len(members))+1]),
		Channels:    fmt.Sprintf("%s", selectedChannels),
		Enabled:     g.random.Float64() > 0.1, // 90%概率启用
	}
}

// GenerateTaskQueue 生成模拟任务队列数据
func (g *MockDataGenerator) GenerateTaskQueue(index int) *model.TaskQueue {
	queueNames := []string{"ai_analysis", "notification", "sync_config", "cleanup", "backup"}
	taskTypes := []string{"ai_analysis", "notification", "config_sync", "data_cleanup", "backup_task"}
	statuses := []string{"pending", "processing", "completed", "failed", "cancelled"}

	queueName := queueNames[g.random.Intn(len(queueNames))]
	taskType := taskTypes[g.random.Intn(len(taskTypes))]
	status := statuses[g.random.Intn(len(statuses))]

	var payload map[string]interface{}
	switch taskType {
	case "ai_analysis":
		payload = map[string]interface{}{
			"alert_id":      g.random.Intn(1000) + 1,
			"content":       fmt.Sprintf("告警内容-%d", index),
			"model":         "llama2",
			"temperature":   0.7,
			"max_tokens":    1000,
			"analysis_type": "anomaly_detection",
		}
	case "notification":
		payload = map[string]interface{}{
			"alert_id":    g.random.Intn(1000) + 1,
			"recipients": []string{"admin@company.com", "ops@company.com"},
			"template":    "email_alert",
			"priority":    g.random.Intn(5) + 1,
		}
	case "config_sync":
		payload = map[string]interface{}{
			"cluster_id":   fmt.Sprintf("cluster-%d", g.random.Intn(5)+1),
			"config_type":  "prometheus_rules",
			"config_hash":  fmt.Sprintf("sha256:abc123def456%d", index),
			"force_sync":   g.random.Float64() > 0.8,
		}
	default:
		payload = map[string]interface{}{
			"task_id":   index,
			"timestamp": time.Now().Unix(),
			"metadata":  map[string]string{"source": "generator"},
		}
	}

	payloadJSON, _ := json.Marshal(payload)

	var startedAt, completedAt *time.Time
	var errorMsg string
	var workerID string

	if status == "processing" || status == "completed" || status == "failed" {
		start := time.Now().Add(-time.Duration(g.random.Intn(60)) * time.Minute)
		startedAt = &start
		workerID = fmt.Sprintf("worker-%d", g.random.Intn(5)+1)
	}

	if status == "completed" || status == "failed" {
		completed := time.Now().Add(-time.Duration(g.random.Intn(30)) * time.Minute)
		completedAt = &completed
	}

	if status == "failed" {
		errorMsg = "Task execution failed: timeout"
	}

	return &model.TaskQueue{
		ID:           fmt.Sprintf("task-%d", index),
		QueueName:    queueName,
		TaskType:     taskType,
		Payload:      string(payloadJSON),
		Priority:     g.random.Intn(5) + 1,
		RetryCount:   g.random.Intn(3),
		MaxRetry:     3,
		Status:       status,
		ScheduledAt:  time.Now().Add(time.Duration(g.random.Intn(60)) * time.Minute),
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		ErrorMessage: errorMsg,
		WorkerID:     workerID,
		CreatedAt:    time.Now().Add(-time.Duration(g.random.Intn(24)) * time.Hour),
		UpdatedAt:    time.Now().Add(-time.Duration(g.random.Intn(6)) * time.Hour),
	}
}

// GenerateKnowledge 生成模拟知识库数据
func (g *MockDataGenerator) GenerateKnowledge(index int) *model.Knowledge {
	categories := []string{"故障排查", "性能优化", "安全防护", "运维指南", "开发规范"}
	sources := []string{"manual", "wiki", "documentation", "experience", "ai_generated"}
	topics := []string{"CPU", "内存", "磁盘", "网络", "数据库", "缓存", "消息队列", "微服务"}

	category := categories[g.random.Intn(len(categories))]
	source := sources[g.random.Intn(len(sources))]
	topic := topics[g.random.Intn(len(topics))]

	title := fmt.Sprintf("%s%s问题处理指南-%d", topic, category, index)
	content := fmt.Sprintf(`# %s

## 问题描述
%s相关的常见问题和解决方案。

## 排查步骤
1. 检查系统状态
2. 分析日志信息
3. 确定问题根因
4. 实施解决方案
5. 验证修复效果

## 解决方案
- 方案一: 重启相关服务
- 方案二: 调整配置参数
- 方案三: 扩容资源

## 预防措施
定期监控和维护，及时发现潜在问题。`, title, topic)

	tags := []string{topic, category, "监控", "告警", "运维"}
	tagsStr := fmt.Sprintf("%s", tags[:g.random.Intn(len(tags))+1])

	summary := fmt.Sprintf("%s%s的详细处理指南，包含问题排查、解决方案和预防措施", topic, category)

	return &model.Knowledge{
		ID:       uint(index),
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     tagsStr,
		Source:   source,
		SourceID: uint(g.random.Intn(1000) + 1),
		Vector:   "", // 向量数据通常由AI模型生成
		Summary:  summary,
		CreatedAt: time.Now().Add(-time.Duration(g.random.Intn(180)) * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(g.random.Intn(30)) * 24 * time.Hour),
	}
}

// GenerateConfigSyncStatus 生成模拟配置同步状态数据
func (g *MockDataGenerator) GenerateConfigSyncStatus(index int) *model.ConfigSyncStatus {
	clusters := []string{"prod-cluster-01", "prod-cluster-02", "staging-cluster-01", "dev-cluster-01"}
	configTypes := []string{"prometheus_rules", "alertmanager_config", "grafana_dashboards", "service_monitors"}
	statuses := []string{"success", "failed", "pending", "in_progress"}

	clusterID := clusters[g.random.Intn(len(clusters))]
	configType := configTypes[g.random.Intn(len(configTypes))]
	status := statuses[g.random.Intn(len(statuses))]

	syncTime := time.Now().Add(-time.Duration(g.random.Intn(60)) * time.Minute)
	var errorMsg string
	if status == "failed" {
		errorMsgs := []string{
			"connection timeout",
			"authentication failed",
			"invalid configuration",
			"resource conflict",
			"permission denied",
		}
		errorMsg = errorMsgs[g.random.Intn(len(errorMsgs))]
	}

	return &model.ConfigSyncStatus{
		ID:           fmt.Sprintf("sync-%d", index),
		ClusterID:    clusterID,
		ConfigType:   configType,
		ConfigHash:   fmt.Sprintf("sha256:abc123def456%d", index),
		SyncStatus:   status,
		SyncTime:     &syncTime,
		ErrorMessage: errorMsg,
		CreatedAt:    time.Now().Add(-time.Duration(g.random.Intn(24)) * time.Hour),
		UpdatedAt:    time.Now().Add(-time.Duration(g.random.Intn(6)) * time.Hour),
	}
}

// GenerateBatchData 批量生成测试数据
func (g *MockDataGenerator) GenerateBatchData(count int) map[string]interface{} {
	data := make(map[string]interface{})

	// 生成规则数据
	rules := make([]*model.Rule, count)
	for i := 0; i < count; i++ {
		rules[i] = g.GenerateRule(i + 1)
	}
	data["rules"] = rules

	// 生成告警数据
	alerts := make([]*model.Alert, count*2) // 每个规则生成2个告警
	for i := 0; i < count*2; i++ {
		ruleID := uint((i % count) + 1)
		alerts[i] = g.GenerateAlert(i+1, ruleID)
	}
	data["alerts"] = alerts

	// 生成提供者数据
	providers := make([]*model.Provider, count/2) // 规则数量的一半
	for i := 0; i < count/2; i++ {
		providers[i] = g.GenerateProvider(i + 1)
	}
	data["providers"] = providers

	// 生成用户数据
	users := make([]*model.User, count/5) // 规则数量的1/5
	for i := 0; i < count/5; i++ {
		users[i] = g.GenerateUser(i + 1)
	}
	data["users"] = users

	// 生成通知模板数据
	templates := make([]*model.NotifyTemplate, count/10) // 规则数量的1/10
	for i := 0; i < count/10; i++ {
		templates[i] = g.GenerateNotifyTemplate(i + 1)
	}
	data["notify_templates"] = templates

	// 生成通知组数据
	groups := make([]*model.NotifyGroup, count/10)
	for i := 0; i < count/10; i++ {
		groups[i] = g.GenerateNotifyGroup(i + 1)
	}
	data["notify_groups"] = groups

	// 生成任务队列数据
	tasks := make([]*model.TaskQueue, count)
	for i := 0; i < count; i++ {
		tasks[i] = g.GenerateTaskQueue(i + 1)
	}
	data["tasks"] = tasks

	// 生成知识库数据
	knowledge := make([]*model.Knowledge, count/5)
	for i := 0; i < count/5; i++ {
		knowledge[i] = g.GenerateKnowledge(i + 1)
	}
	data["knowledge"] = knowledge

	// 生成配置同步状态数据
	syncStatuses := make([]*model.ConfigSyncStatus, count/3)
	for i := 0; i < count/3; i++ {
		syncStatuses[i] = g.GenerateConfigSyncStatus(i + 1)
	}
	data["config_sync_statuses"] = syncStatuses

	return data
}