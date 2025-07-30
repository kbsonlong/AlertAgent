package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/pkg/queue"
	"alert_agent/internal/service"

	"go.uber.org/zap"
)

// ConfigSyncHandler 配置同步任务处理器
type ConfigSyncHandler struct {
	ruleService *service.RuleService
}

// ConfigSyncTask 配置同步任务
type ConfigSyncTask struct {
	Type       string                 `json:"type"`        // rule_create, rule_update, rule_delete
	RuleID     string                 `json:"rule_id"`
	Targets    []string               `json:"targets"`
	ConfigData map[string]interface{} `json:"config_data,omitempty"`
	Priority   string                 `json:"priority"`
}

// ConfigSyncResult 配置同步结果
type ConfigSyncResult struct {
	TaskID      string                 `json:"task_id"`
	RuleID      string                 `json:"rule_id"`
	Type        string                 `json:"type"`
	Targets     []string               `json:"targets"`
	Results     map[string]*SyncResult `json:"results"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// SyncResult 单个目标的同步结果
type SyncResult struct {
	Target    string        `json:"target"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	ConfigHash string       `json:"config_hash,omitempty"`
}

// NewConfigSyncHandler 创建配置同步处理器
func NewConfigSyncHandler() *ConfigSyncHandler {
	return &ConfigSyncHandler{
		ruleService: service.NewRuleService(),
	}
}

// Type 返回处理器类型
func (h *ConfigSyncHandler) Type() queue.TaskType {
	return queue.TaskTypeConfigSync
}

// Handle 处理配置同步任务
func (h *ConfigSyncHandler) Handle(ctx context.Context, task *queue.Task) error {
	startTime := time.Now()
	
	// 解析任务载荷
	syncType, ok := task.Payload["type"].(string)
	if !ok {
		return fmt.Errorf("invalid type in task payload")
	}

	ruleID, ok := task.Payload["rule_id"].(string)
	if !ok {
		return fmt.Errorf("invalid rule_id in task payload")
	}

	targetsInterface, ok := task.Payload["targets"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid targets in task payload")
	}

	// 转换targets
	targets := make([]string, len(targetsInterface))
	for i, target := range targetsInterface {
		targets[i] = target.(string)
	}

	logger.L.Info("Processing config sync task",
		zap.String("task_id", task.ID),
		zap.String("type", syncType),
		zap.String("rule_id", ruleID),
		zap.Strings("targets", targets),
	)

	// 创建同步结果
	syncResult := &ConfigSyncResult{
		TaskID:    task.ID,
		RuleID:    ruleID,
		Type:      syncType,
		Targets:   targets,
		Results:   make(map[string]*SyncResult),
		StartTime: startTime,
	}

	// 根据同步类型执行相应操作
	var err error
	switch syncType {
	case "rule_create":
		err = h.handleRuleCreate(ctx, ruleID, targets, syncResult)
	case "rule_update":
		err = h.handleRuleUpdate(ctx, ruleID, targets, syncResult)
	case "rule_delete":
		err = h.handleRuleDelete(ctx, ruleID, targets, syncResult)
	case "full_sync":
		err = h.handleFullSync(ctx, targets, syncResult)
	default:
		err = fmt.Errorf("unknown sync type: %s", syncType)
	}

	// 完成同步结果
	syncResult.EndTime = time.Now()
	syncResult.Duration = syncResult.EndTime.Sub(syncResult.StartTime)
	syncResult.Success = err == nil

	if err != nil {
		syncResult.ErrorMsg = err.Error()
	}

	// 保存同步结果
	if saveErr := h.saveSyncResult(ctx, syncResult); saveErr != nil {
		logger.L.Error("Failed to save sync result", zap.Error(saveErr))
	}

	// 更新规则分发状态
	if updateErr := h.updateRuleDistributionStatus(ctx, ruleID, syncResult); updateErr != nil {
		logger.L.Error("Failed to update rule distribution status", zap.Error(updateErr))
	}

	if err != nil {
		logger.L.Error("Config sync task failed",
			zap.String("task_id", task.ID),
			zap.String("rule_id", ruleID),
			zap.Duration("duration", syncResult.Duration),
			zap.Error(err),
		)
		return err
	}

	logger.L.Info("Config sync task completed successfully",
		zap.String("task_id", task.ID),
		zap.String("rule_id", ruleID),
		zap.Duration("duration", syncResult.Duration),
		zap.Int("success_count", h.countSuccessfulSyncs(syncResult)),
		zap.Int("total_targets", len(targets)),
	)

	return nil
}

// handleRuleCreate 处理规则创建同步
func (h *ConfigSyncHandler) handleRuleCreate(ctx context.Context, ruleID string, targets []string, result *ConfigSyncResult) error {
	// 获取规则信息
	rule, err := h.getRuleByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get rule %s: %w", ruleID, err)
	}

	// 为每个目标生成配置并同步
	for _, target := range targets {
		syncResult := h.syncRuleToTarget(ctx, rule, target, "create")
		result.Results[target] = syncResult
	}

	return h.checkSyncResults(result)
}

// handleRuleUpdate 处理规则更新同步
func (h *ConfigSyncHandler) handleRuleUpdate(ctx context.Context, ruleID string, targets []string, result *ConfigSyncResult) error {
	// 获取规则信息
	rule, err := h.getRuleByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get rule %s: %w", ruleID, err)
	}

	// 为每个目标生成配置并同步
	for _, target := range targets {
		syncResult := h.syncRuleToTarget(ctx, rule, target, "update")
		result.Results[target] = syncResult
	}

	return h.checkSyncResults(result)
}

// handleRuleDelete 处理规则删除同步
func (h *ConfigSyncHandler) handleRuleDelete(ctx context.Context, ruleID string, targets []string, result *ConfigSyncResult) error {
	// 为每个目标删除规则配置
	for _, target := range targets {
		syncResult := h.deleteRuleFromTarget(ctx, ruleID, target)
		result.Results[target] = syncResult
	}

	return h.checkSyncResults(result)
}

// handleFullSync 处理全量同步
func (h *ConfigSyncHandler) handleFullSync(ctx context.Context, targets []string, result *ConfigSyncResult) error {
	// 获取所有活跃规则
	rules, err := h.getAllActiveRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active rules: %w", err)
	}

	// 为每个目标执行全量同步
	for _, target := range targets {
		syncResult := h.fullSyncToTarget(ctx, rules, target)
		result.Results[target] = syncResult
	}

	return h.checkSyncResults(result)
}

// syncRuleToTarget 同步规则到目标系统
func (h *ConfigSyncHandler) syncRuleToTarget(ctx context.Context, rule *model.Rule, target, operation string) *SyncResult {
	startTime := time.Now()
	result := &SyncResult{
		Target:    target,
		Timestamp: startTime,
	}

	// 根据目标类型生成配置
	config, configHash, err := h.generateConfigForTarget(rule, target)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to generate config: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.ConfigHash = configHash

	// 调用目标系统的配置更新接口
	if err := h.callTargetConfigAPI(ctx, target, operation, config); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to sync to target: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	logger.L.Debug("Rule synced to target",
		zap.String("rule_id", fmt.Sprintf("%d", rule.ID)),
		zap.String("target", target),
		zap.String("operation", operation),
		zap.String("config_hash", configHash),
		zap.Duration("duration", result.Duration),
	)

	return result
}

// deleteRuleFromTarget 从目标系统删除规则
func (h *ConfigSyncHandler) deleteRuleFromTarget(ctx context.Context, ruleID, target string) *SyncResult {
	startTime := time.Now()
	result := &SyncResult{
		Target:    target,
		Timestamp: startTime,
	}

	// 调用目标系统的配置删除接口
	if err := h.callTargetConfigAPI(ctx, target, "delete", map[string]interface{}{
		"rule_id": ruleID,
	}); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to delete from target: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	logger.L.Debug("Rule deleted from target",
		zap.String("rule_id", ruleID),
		zap.String("target", target),
		zap.Duration("duration", result.Duration),
	)

	return result
}

// fullSyncToTarget 全量同步到目标系统
func (h *ConfigSyncHandler) fullSyncToTarget(ctx context.Context, rules []*model.Rule, target string) *SyncResult {
	startTime := time.Now()
	result := &SyncResult{
		Target:    target,
		Timestamp: startTime,
	}

	// 生成完整的配置文件
	fullConfig, configHash, err := h.generateFullConfigForTarget(rules, target)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to generate full config: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.ConfigHash = configHash

	// 调用目标系统的全量配置更新接口
	if err := h.callTargetConfigAPI(ctx, target, "full_sync", fullConfig); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to full sync to target: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	logger.L.Info("Full sync completed for target",
		zap.String("target", target),
		zap.Int("rule_count", len(rules)),
		zap.String("config_hash", configHash),
		zap.Duration("duration", result.Duration),
	)

	return result
}

// generateConfigForTarget 为目标系统生成配置
func (h *ConfigSyncHandler) generateConfigForTarget(rule *model.Rule, target string) (map[string]interface{}, string, error) {
	switch target {
	case "prometheus":
		return h.generatePrometheusConfig(rule)
	case "alertmanager":
		return h.generateAlertmanagerConfig(rule)
	case "vmalert":
		return h.generateVMAlertConfig(rule)
	default:
		return nil, "", fmt.Errorf("unsupported target: %s", target)
	}
}

// generateFullConfigForTarget 为目标系统生成完整配置
func (h *ConfigSyncHandler) generateFullConfigForTarget(rules []*model.Rule, target string) (map[string]interface{}, string, error) {
	switch target {
	case "prometheus":
		return h.generateFullPrometheusConfig(rules)
	case "alertmanager":
		return h.generateFullAlertmanagerConfig(rules)
	case "vmalert":
		return h.generateFullVMAlertConfig(rules)
	default:
		return nil, "", fmt.Errorf("unsupported target: %s", target)
	}
}

// generatePrometheusConfig 生成Prometheus配置
func (h *ConfigSyncHandler) generatePrometheusConfig(rule *model.Rule) (map[string]interface{}, string, error) {
	config := map[string]interface{}{
		"groups": []map[string]interface{}{
			{
				"name": fmt.Sprintf("alertagent_rule_%d", rule.ID),
				"rules": []map[string]interface{}{
					{
						"alert":       rule.Name,
						"expr":        rule.Expression,
						"for":         rule.Duration,
						"labels":      h.parseLabels(rule.Labels),
						"annotations": h.parseAnnotations(rule.Annotations),
					},
				},
			},
		},
	}

	// 计算配置哈希
	configJSON, _ := json.Marshal(config)
	configHash := fmt.Sprintf("%x", configJSON)[:16]

	return config, configHash, nil
}

// generateAlertmanagerConfig 生成Alertmanager配置
func (h *ConfigSyncHandler) generateAlertmanagerConfig(rule *model.Rule) (map[string]interface{}, string, error) {
	// Alertmanager主要处理路由和通知，这里生成路由配置
	config := map[string]interface{}{
		"route": map[string]interface{}{
			"group_by":        []string{"alertname"},
			"group_wait":      "10s",
			"group_interval":  "10s",
			"repeat_interval": "1h",
			"receiver":        "default",
			"routes": []map[string]interface{}{
				{
					"match": map[string]string{
						"alertname": rule.Name,
					},
					"receiver": fmt.Sprintf("rule_%d_receiver", rule.ID),
				},
			},
		},
		"receivers": []map[string]interface{}{
			{
				"name": fmt.Sprintf("rule_%d_receiver", rule.ID),
				"webhook_configs": []map[string]interface{}{
					{
						"url": "http://alertagent:8080/api/v1/alerts/webhook",
					},
				},
			},
		},
	}

	configJSON, _ := json.Marshal(config)
	configHash := fmt.Sprintf("%x", configJSON)[:16]

	return config, configHash, nil
}

// generateVMAlertConfig 生成VMAlert配置
func (h *ConfigSyncHandler) generateVMAlertConfig(rule *model.Rule) (map[string]interface{}, string, error) {
	config := map[string]interface{}{
		"groups": []map[string]interface{}{
			{
				"name": fmt.Sprintf("alertagent_rule_%d", rule.ID),
				"rules": []map[string]interface{}{
					{
						"alert":       rule.Name,
						"expr":        rule.Expression,
						"for":         rule.Duration,
						"labels":      h.parseLabels(rule.Labels),
						"annotations": h.parseAnnotations(rule.Annotations),
					},
				},
			},
		},
	}

	configJSON, _ := json.Marshal(config)
	configHash := fmt.Sprintf("%x", configJSON)[:16]

	return config, configHash, nil
}

// generateFullPrometheusConfig 生成完整的Prometheus配置
func (h *ConfigSyncHandler) generateFullPrometheusConfig(rules []*model.Rule) (map[string]interface{}, string, error) {
	groups := make([]map[string]interface{}, 0, len(rules))
	
	for _, rule := range rules {
		group := map[string]interface{}{
			"name": fmt.Sprintf("alertagent_rule_%d", rule.ID),
			"rules": []map[string]interface{}{
				{
					"alert":       rule.Name,
					"expr":        rule.Expression,
					"for":         rule.Duration,
					"labels":      h.parseLabels(rule.Labels),
					"annotations": h.parseAnnotations(rule.Annotations),
				},
			},
		}
		groups = append(groups, group)
	}

	config := map[string]interface{}{
		"groups": groups,
	}

	configJSON, _ := json.Marshal(config)
	configHash := fmt.Sprintf("%x", configJSON)[:16]

	return config, configHash, nil
}

// generateFullAlertmanagerConfig 生成完整的Alertmanager配置
func (h *ConfigSyncHandler) generateFullAlertmanagerConfig(rules []*model.Rule) (map[string]interface{}, string, error) {
	routes := make([]map[string]interface{}, 0, len(rules))
	receivers := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		route := map[string]interface{}{
			"match": map[string]string{
				"alertname": rule.Name,
			},
			"receiver": fmt.Sprintf("rule_%d_receiver", rule.ID),
		}
		routes = append(routes, route)

		receiver := map[string]interface{}{
			"name": fmt.Sprintf("rule_%d_receiver", rule.ID),
			"webhook_configs": []map[string]interface{}{
				{
					"url": "http://alertagent:8080/api/v1/alerts/webhook",
				},
			},
		}
		receivers = append(receivers, receiver)
	}

	config := map[string]interface{}{
		"route": map[string]interface{}{
			"group_by":        []string{"alertname"},
			"group_wait":      "10s",
			"group_interval":  "10s",
			"repeat_interval": "1h",
			"receiver":        "default",
			"routes":          routes,
		},
		"receivers": receivers,
	}

	configJSON, _ := json.Marshal(config)
	configHash := fmt.Sprintf("%x", configJSON)[:16]

	return config, configHash, nil
}

// generateFullVMAlertConfig 生成完整的VMAlert配置
func (h *ConfigSyncHandler) generateFullVMAlertConfig(rules []*model.Rule) (map[string]interface{}, string, error) {
	return h.generateFullPrometheusConfig(rules) // VMAlert使用与Prometheus相同的格式
}

// callTargetConfigAPI 调用目标系统的配置API
func (h *ConfigSyncHandler) callTargetConfigAPI(ctx context.Context, target, operation string, config map[string]interface{}) error {
	// 这里应该调用实际的目标系统API
	// 暂时模拟成功
	logger.L.Debug("Calling target config API (mock)",
		zap.String("target", target),
		zap.String("operation", operation),
		zap.Any("config", config),
	)

	// 模拟网络延迟
	time.Sleep(100 * time.Millisecond)

	return nil
}

// parseLabels 解析标签
func (h *ConfigSyncHandler) parseLabels(labelsJSON string) map[string]string {
	if labelsJSON == "" {
		return make(map[string]string)
	}

	var labels map[string]string
	if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
		logger.L.Warn("Failed to parse labels", zap.String("labels", labelsJSON), zap.Error(err))
		return make(map[string]string)
	}

	return labels
}

// parseAnnotations 解析注释
func (h *ConfigSyncHandler) parseAnnotations(annotationsJSON string) map[string]string {
	if annotationsJSON == "" {
		return make(map[string]string)
	}

	var annotations map[string]string
	if err := json.Unmarshal([]byte(annotationsJSON), &annotations); err != nil {
		logger.L.Warn("Failed to parse annotations", zap.String("annotations", annotationsJSON), zap.Error(err))
		return make(map[string]string)
	}

	return annotations
}

// getRuleByID 根据ID获取规则
func (h *ConfigSyncHandler) getRuleByID(ctx context.Context, ruleID string) (*model.Rule, error) {
	var rule model.Rule
	if err := database.DB.Where("id = ?", ruleID).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// getAllActiveRules 获取所有活跃规则
func (h *ConfigSyncHandler) getAllActiveRules(ctx context.Context) ([]*model.Rule, error) {
	var rules []*model.Rule
	if err := database.DB.Where("status = ?", "active").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// checkSyncResults 检查同步结果
func (h *ConfigSyncHandler) checkSyncResults(result *ConfigSyncResult) error {
	failedTargets := make([]string, 0)
	
	for target, syncResult := range result.Results {
		if !syncResult.Success {
			failedTargets = append(failedTargets, target)
		}
	}

	if len(failedTargets) > 0 {
		return fmt.Errorf("sync failed for targets: %v", failedTargets)
	}

	return nil
}

// countSuccessfulSyncs 统计成功的同步数量
func (h *ConfigSyncHandler) countSuccessfulSyncs(result *ConfigSyncResult) int {
	count := 0
	for _, syncResult := range result.Results {
		if syncResult.Success {
			count++
		}
	}
	return count
}

// saveSyncResult 保存同步结果
func (h *ConfigSyncHandler) saveSyncResult(ctx context.Context, result *ConfigSyncResult) error {
	// 这里可以将同步结果保存到数据库中
	// 暂时只记录到日志
	logger.L.Info("Config sync result",
		zap.String("task_id", result.TaskID),
		zap.String("rule_id", result.RuleID),
		zap.String("type", result.Type),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration),
		zap.Int("target_count", len(result.Targets)),
		zap.Int("success_count", h.countSuccessfulSyncs(result)),
	)

	return nil
}

// updateRuleDistributionStatus 更新规则分发状态
func (h *ConfigSyncHandler) updateRuleDistributionStatus(ctx context.Context, ruleID string, result *ConfigSyncResult) error {
	// 这里可以更新规则的分发状态
	// 暂时跳过数据库操作
	logger.L.Debug("Rule distribution status updated",
		zap.String("rule_id", ruleID),
		zap.Bool("success", result.Success),
	)

	return nil
}