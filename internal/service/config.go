package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConfigService 配置服务
type ConfigService struct{}

// NewConfigService 创建配置服务
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// SyncStatusUpdate 同步状态更新
type SyncStatusUpdate struct {
	ClusterID    string
	ConfigType   string
	Status       string
	SyncTime     time.Time
	ErrorMessage string
	ConfigHash   string
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	ClusterID   string                 `json:"cluster_id"`
	ConfigTypes []string               `json:"config_types"`
	LastSync    map[string]time.Time   `json:"last_sync"`
	Status      map[string]string      `json:"status"`
}

// GetConfig 获取指定集群和类型的配置
func (cs *ConfigService) GetConfig(ctx context.Context, clusterID, configType string) (string, error) {
	logger.L.Debug("Getting config",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	switch configType {
	case "prometheus":
		return cs.getPrometheusConfig(ctx, clusterID)
	case "alertmanager":
		return cs.getAlertmanagerConfig(ctx, clusterID)
	case "vmalert":
		return cs.getVMAlertConfig(ctx, clusterID)
	default:
		return "", fmt.Errorf("unsupported config type: %s", configType)
	}
}

// getPrometheusConfig 获取Prometheus配置
func (cs *ConfigService) getPrometheusConfig(ctx context.Context, clusterID string) (string, error) {
	// 获取该集群的所有告警规则
	var rules []model.Rule
	err := database.DB.WithContext(ctx).
		Where("JSON_CONTAINS(targets, ?)", fmt.Sprintf(`"%s"`, clusterID)).
		Where("status = ?", "active").
		Find(&rules).Error
	
	if err != nil {
		return "", fmt.Errorf("failed to get rules: %w", err)
	}

	// 生成Prometheus规则配置
	config := cs.generatePrometheusRules(rules)
	return config, nil
}

// getAlertmanagerConfig 获取Alertmanager配置
func (cs *ConfigService) getAlertmanagerConfig(ctx context.Context, clusterID string) (string, error) {
	// 获取通知配置
	var templates []model.NotifyTemplate
	err := database.DB.WithContext(ctx).Find(&templates).Error
	if err != nil {
		return "", fmt.Errorf("failed to get notify templates: %w", err)
	}

	var groups []model.NotifyGroup
	err = database.DB.WithContext(ctx).Find(&groups).Error
	if err != nil {
		return "", fmt.Errorf("failed to get notify groups: %w", err)
	}

	// 生成Alertmanager配置
	config := cs.generateAlertmanagerConfig(templates, groups, clusterID)
	return config, nil
}

// getVMAlertConfig 获取VMAlert配置
func (cs *ConfigService) getVMAlertConfig(ctx context.Context, clusterID string) (string, error) {
	// VMAlert使用类似Prometheus的规则格式
	return cs.getPrometheusConfig(ctx, clusterID)
}

// generatePrometheusRules 生成Prometheus规则配置
func (cs *ConfigService) generatePrometheusRules(rules []model.Rule) string {
	if len(rules) == 0 {
		return "groups: []"
	}

	var config strings.Builder
	config.WriteString("groups:\n")
	config.WriteString("  - name: alertagent_rules\n")
	config.WriteString("    interval: 30s\n")
	config.WriteString("    rules:\n")

	for _, rule := range rules {
		config.WriteString(fmt.Sprintf("      - alert: %s\n", rule.Name))
		config.WriteString(fmt.Sprintf("        expr: %s\n", rule.Expression))
		
		if rule.Duration != "" {
			config.WriteString(fmt.Sprintf("        for: %s\n", rule.Duration))
		}

		// 添加标签
		labels, err := rule.GetLabelsMap()
		if err == nil && len(labels) > 0 {
			config.WriteString("        labels:\n")
			// 添加默认标签
			config.WriteString(fmt.Sprintf("          severity: %s\n", rule.Severity))
			config.WriteString("          source: alertagent\n")
			for key, value := range labels {
				config.WriteString(fmt.Sprintf("          %s: \"%s\"\n", key, value))
			}
		} else {
			// 如果没有自定义标签，至少添加默认标签
			config.WriteString("        labels:\n")
			config.WriteString(fmt.Sprintf("          severity: %s\n", rule.Severity))
			config.WriteString("          source: alertagent\n")
		}

		// 添加注释
		annotations, err := rule.GetAnnotationsMap()
		if err == nil && len(annotations) > 0 {
			config.WriteString("        annotations:\n")
			// 添加默认注释
			config.WriteString(fmt.Sprintf("          summary: \"Alert: %s\"\n", rule.Name))
			config.WriteString(fmt.Sprintf("          description: \"Rule: %s triggered\"\n", rule.Name))
			for key, value := range annotations {
				config.WriteString(fmt.Sprintf("          %s: \"%s\"\n", key, value))
			}
		} else {
			// 如果没有自定义注释，至少添加默认注释
			config.WriteString("        annotations:\n")
			config.WriteString(fmt.Sprintf("          summary: \"Alert: %s\"\n", rule.Name))
			config.WriteString(fmt.Sprintf("          description: \"Rule: %s triggered\"\n", rule.Name))
		}

		config.WriteString("\n")
	}

	return config.String()
}

// generateAlertmanagerConfig 生成Alertmanager配置
func (cs *ConfigService) generateAlertmanagerConfig(templates []model.NotifyTemplate, groups []model.NotifyGroup, clusterID string) string {
	var config strings.Builder
	
	// 全局配置
	config.WriteString("global:\n")
	config.WriteString("  smtp_smarthost: 'localhost:587'\n")
	config.WriteString("  smtp_from: 'alertmanager@example.com'\n")
	config.WriteString("  resolve_timeout: 5m\n")
	config.WriteString("\n")

	// 模板配置
	if len(templates) > 0 {
		config.WriteString("templates:\n")
		for _, template := range templates {
			config.WriteString(fmt.Sprintf("  - '/etc/alertmanager/templates/%s.tmpl'\n", template.Name))
		}
		config.WriteString("\n")
	}

	// 路由配置
	config.WriteString("route:\n")
	config.WriteString("  group_by: ['alertname', 'cluster', 'service']\n")
	config.WriteString("  group_wait: 10s\n")
	config.WriteString("  group_interval: 10s\n")
	config.WriteString("  repeat_interval: 12h\n")
	config.WriteString("  receiver: 'default'\n")
	
	// 根据通知组生成路由规则
	if len(groups) > 0 {
		config.WriteString("  routes:\n")
		for _, group := range groups {
			config.WriteString(fmt.Sprintf("    - match:\n"))
			config.WriteString(fmt.Sprintf("        severity: %s\n", group.Name)) // 假设group.Name对应severity
			config.WriteString(fmt.Sprintf("      receiver: '%s'\n", group.Name))
		}
	}
	config.WriteString("\n")

	// 接收器配置
	config.WriteString("receivers:\n")
	
	// 默认接收器
	config.WriteString("  - name: 'default'\n")
	config.WriteString("    email_configs:\n")
	config.WriteString("      - to: 'admin@example.com'\n")
	config.WriteString("        subject: '[{{ .Status | toUpper }}{{ if eq .Status \"firing\" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .GroupLabels.alertname }}'\n")
	config.WriteString("        body: |\n")
	config.WriteString("          {{ range .Alerts }}\n")
	config.WriteString("          **Alert:** {{ .Annotations.summary }}\n")
	config.WriteString("          **Description:** {{ .Annotations.description }}\n")
	config.WriteString("          **Severity:** {{ .Labels.severity }}\n")
	config.WriteString("          **Time:** {{ .StartsAt.Format \"2006-01-02 15:04:05\" }}\n")
	config.WriteString("          {{ end }}\n")

	// 根据通知组生成接收器
	for _, group := range groups {
		config.WriteString(fmt.Sprintf("  - name: '%s'\n", group.Name))
		config.WriteString("    email_configs:\n")
		config.WriteString(fmt.Sprintf("      - to: '%s@example.com'\n", group.Name))
		config.WriteString("        subject: '[{{ .Status | toUpper }}] {{ .GroupLabels.alertname }} - {{ .GroupLabels.severity }}'\n")
		config.WriteString("        body: |\n")
		config.WriteString("          {{ range .Alerts }}\n")
		config.WriteString("          **Alert:** {{ .Annotations.summary }}\n")
		config.WriteString("          **Description:** {{ .Annotations.description }}\n")
		config.WriteString("          **Severity:** {{ .Labels.severity }}\n")
		config.WriteString("          **Cluster:** {{ .Labels.cluster }}\n")
		config.WriteString("          **Time:** {{ .StartsAt.Format \"2006-01-02 15:04:05\" }}\n")
		config.WriteString("          {{ end }}\n")
	}

	// 抑制规则
	config.WriteString("\ninhibit_rules:\n")
	config.WriteString("  - source_match:\n")
	config.WriteString("      severity: 'critical'\n")
	config.WriteString("    target_match:\n")
	config.WriteString("      severity: 'warning'\n")
	config.WriteString("    equal: ['alertname', 'cluster', 'service']\n")

	return config.String()
}

// UpdateSyncStatus 更新同步状态
func (cs *ConfigService) UpdateSyncStatus(ctx context.Context, update *SyncStatusUpdate) error {
	// 查找或创建同步状态记录
	var status model.ConfigSyncStatus
	err := database.DB.WithContext(ctx).
		Where("cluster_id = ? AND config_type = ?", update.ClusterID, update.ConfigType).
		First(&status).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		status = model.ConfigSyncStatus{
			ClusterID:  update.ClusterID,
			ConfigType: update.ConfigType,
		}
	} else if err != nil {
		return fmt.Errorf("failed to query sync status: %w", err)
	}

	// 更新字段
	status.SyncStatus = update.Status
	status.SyncTime = &update.SyncTime
	status.ErrorMessage = update.ErrorMessage
	status.ConfigHash = update.ConfigHash

	// 保存到数据库
	if err := database.DB.WithContext(ctx).Save(&status).Error; err != nil {
		return fmt.Errorf("failed to save sync status: %w", err)
	}

	logger.L.Debug("Sync status updated in database",
		zap.String("cluster_id", update.ClusterID),
		zap.String("config_type", update.ConfigType),
		zap.String("status", update.Status),
	)

	return nil
}

// GetSyncStatus 获取同步状态
func (cs *ConfigService) GetSyncStatus(ctx context.Context, clusterID, configType string) ([]model.ConfigSyncStatus, error) {
	var statuses []model.ConfigSyncStatus
	query := database.DB.WithContext(ctx).Where("cluster_id = ?", clusterID)
	
	if configType != "" {
		query = query.Where("config_type = ?", configType)
	}

	err := query.Order("updated_at DESC").Find(&statuses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return statuses, nil
}

// ListClusters 列出所有集群
func (cs *ConfigService) ListClusters(ctx context.Context) ([]ClusterInfo, error) {
	// 从同步状态表获取所有集群信息
	var statuses []model.ConfigSyncStatus
	err := database.DB.WithContext(ctx).Find(&statuses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sync statuses: %w", err)
	}

	// 按集群ID分组
	clusterMap := make(map[string]*ClusterInfo)
	for _, status := range statuses {
		cluster, exists := clusterMap[status.ClusterID]
		if !exists {
			cluster = &ClusterInfo{
				ClusterID:   status.ClusterID,
				ConfigTypes: []string{},
				LastSync:    make(map[string]time.Time),
				Status:      make(map[string]string),
			}
			clusterMap[status.ClusterID] = cluster
		}

		cluster.ConfigTypes = append(cluster.ConfigTypes, status.ConfigType)
		cluster.Status[status.ConfigType] = status.SyncStatus
		if status.SyncTime != nil {
			cluster.LastSync[status.ConfigType] = *status.SyncTime
		}
	}

	// 转换为切片
	clusters := make([]ClusterInfo, 0, len(clusterMap))
	for _, cluster := range clusterMap {
		clusters = append(clusters, *cluster)
	}

	return clusters, nil
}

// TriggerSync 触发配置同步
func (cs *ConfigService) TriggerSync(ctx context.Context, clusterID, configType string) error {
	// 这里可以通过消息队列或其他方式通知Sidecar进行同步
	// 目前只是记录日志，实际实现可能需要Redis发布/订阅或其他机制
	
	logger.L.Info("Sync trigger requested",
		zap.String("cluster_id", clusterID),
		zap.String("config_type", configType),
	)

	// 可以在这里发送消息到Redis队列，让Sidecar监听
	// 或者更新数据库中的触发标志

	return nil
}