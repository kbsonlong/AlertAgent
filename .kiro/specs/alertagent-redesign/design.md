# AlertAgent 重构设计文档

## 概述

本设计文档基于需求文档，详细描述了AlertAgent重构的技术架构、组件设计和实现方案。重构后的系统采用微服务架构，支持水平扩展，提供统一的API接口和插件化的扩展能力。

## 架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AlertAgent 重构架构                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │   Web Frontend  │    │   API Gateway   │    │  Config Syncer  │         │
│  │   React + TS    │    │   Gin + Go      │    │   Sidecar       │         │
│  └─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘         │
│            │                      │                      │                 │
│            └──────────────────────┼──────────────────────┘                 │
│                                   │                                        │
│  ┌─────────────────────────────────▼─────────────────────────────┐          │
│  │                    AlertAgent Core                            │          │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │          │
│  │  │ Rule Manager│  │Alert Gateway│  │Plugin Manager│          │          │
│  │  └─────────────┘  └─────────────┘  └─────────────┘           │          │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │          │
│  │  │Task Producer│  │Config API   │  │Notification │           │          │
│  │  └─────────────┘  └─────────────┘  └─────────────┘           │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                     │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │                 Message Queue                     │          │
│  │              Redis / RabbitMQ                     │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                     │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │                Worker Cluster                     │          │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │          │
│  │  │AI Analysis  │  │Notification │  │Config Sync  │ │          │
│  │  │Worker       │  │Worker       │  │Worker       │ │          │
│  │  └─────────────┘  └─────────────┘  └─────────────┘ │          │
│  └─────────────────────────┬─────────────────────────┘          │
│                           │                                     │
│  ┌─────────────────────────▼─────────────────────────┐          │
│  │              External Integrations                │          │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │          │
│  │  │Alertmanager │  │ Prometheus  │  │   vmalert   │ │          │
│  │  │+ Sidecar    │  │+ Sidecar    │  │+ Sidecar    │ │          │
│  │  └─────────────┘  └─────────────┘  └─────────────┘ │          │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │          │
│  │  │    Dify     │  │    n8n      │  │Notification │ │          │
│  │  │   AI平台    │  │  工作流     │  │  Channels   │ │          │
│  │  └─────────────┘  └─────────────┘  └─────────────┘ │          │
│  └─────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 组件和接口

### 1. API Gateway

**职责：** 统一API入口，请求路由，认证授权，限流熔断

**接口设计：**

```go
// 告警规则管理API
type RuleAPI struct {
    ruleService *service.RuleService
}

// 创建告警规则
// POST /api/v1/rules
func (r *RuleAPI) CreateRule(c *gin.Context) {
    var req CreateRuleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }
    
    rule, err := r.ruleService.CreateRule(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(201, SuccessResponse{Data: rule})
}

// 更新告警规则
// PUT /api/v1/rules/{id}
func (r *RuleAPI) UpdateRule(c *gin.Context) {
    ruleID := c.Param("id")
    var req UpdateRuleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }
    
    rule, err := r.ruleService.UpdateRule(c.Request.Context(), ruleID, &req)
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(200, SuccessResponse{Data: rule})
}

// 获取规则分发状态
// GET /api/v1/rules/{id}/distribution
func (r *RuleAPI) GetRuleDistribution(c *gin.Context) {
    ruleID := c.Param("id")
    status, err := r.ruleService.GetDistributionStatus(c.Request.Context(), ruleID)
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(200, SuccessResponse{Data: status})
}

// 配置同步API
type ConfigAPI struct {
    configService *service.ConfigService
}

// Sidecar配置拉取接口
// GET /api/v1/config/sync
func (c *ConfigAPI) GetSyncConfig(ctx *gin.Context) {
    clusterID := ctx.Query("cluster_id")
    configType := ctx.Query("type") // prometheus, alertmanager, vmalert
    
    if clusterID == "" || configType == "" {
        ctx.JSON(400, ErrorResponse{Error: "cluster_id and type are required"})
        return
    }
    
    config, hash, err := c.configService.GetConfig(ctx.Request.Context(), clusterID, configType)
    if err != nil {
        ctx.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    ctx.Header("X-Config-Hash", hash)
    ctx.Header("Content-Type", "application/yaml")
    ctx.String(200, config)
}
```

### 2. Rule Manager

**职责：** 告警规则的CRUD操作，版本管理，语法验证

**核心接口：**

```go
type RuleManager struct {
    repo      repository.RuleRepository
    validator RuleValidator
    publisher TaskPublisher
}

type Rule struct {
    ID          string            `json:"id" db:"id"`
    Name        string            `json:"name" db:"name"`
    Expression  string            `json:"expression" db:"expression"`
    Duration    string            `json:"duration" db:"duration"`
    Severity    string            `json:"severity" db:"severity"`
    Labels      map[string]string `json:"labels" db:"labels"`
    Annotations map[string]string `json:"annotations" db:"annotations"`
    Targets     []string          `json:"targets" db:"targets"`
    Version     string            `json:"version" db:"version"`
    Status      string            `json:"status" db:"status"`
    CreatedAt   time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

func (rm *RuleManager) CreateRule(ctx context.Context, req *CreateRuleRequest) (*Rule, error) {
    // 1. 验证规则语法
    if err := rm.validator.ValidateRule(req.Expression, req.Duration); err != nil {
        return nil, fmt.Errorf("rule validation failed: %w", err)
    }
    
    // 2. 创建规则对象
    rule := &Rule{
        ID:          uuid.New().String(),
        Name:        req.Name,
        Expression:  req.Expression,
        Duration:    req.Duration,
        Severity:    req.Severity,
        Labels:      req.Labels,
        Annotations: req.Annotations,
        Targets:     req.Targets,
        Version:     "v1.0.0",
        Status:      "pending",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    // 3. 保存到数据库
    if err := rm.repo.Create(ctx, rule); err != nil {
        return nil, fmt.Errorf("failed to create rule: %w", err)
    }
    
    // 4. 发布配置同步任务
    task := &ConfigSyncTask{
        Type:     "rule_create",
        RuleID:   rule.ID,
        Targets:  rule.Targets,
        Priority: "normal",
    }
    
    if err := rm.publisher.PublishTask(ctx, "config_sync", task); err != nil {
        log.Errorf("failed to publish config sync task: %v", err)
    }
    
    return rule, nil
}

func (rm *RuleManager) UpdateRule(ctx context.Context, ruleID string, req *UpdateRuleRequest) (*Rule, error) {
    // 1. 获取现有规则
    rule, err := rm.repo.GetByID(ctx, ruleID)
    if err != nil {
        return nil, fmt.Errorf("rule not found: %w", err)
    }
    
    // 2. 验证新规则语法
    if req.Expression != "" {
        if err := rm.validator.ValidateRule(req.Expression, req.Duration); err != nil {
            return nil, fmt.Errorf("rule validation failed: %w", err)
        }
        rule.Expression = req.Expression
    }
    
    // 3. 更新规则字段
    if req.Name != "" {
        rule.Name = req.Name
    }
    if req.Duration != "" {
        rule.Duration = req.Duration
    }
    if req.Severity != "" {
        rule.Severity = req.Severity
    }
    if req.Labels != nil {
        rule.Labels = req.Labels
    }
    if req.Annotations != nil {
        rule.Annotations = req.Annotations
    }
    if req.Targets != nil {
        rule.Targets = req.Targets
    }
    
    // 4. 版本递增
    rule.Version = rm.incrementVersion(rule.Version)
    rule.UpdatedAt = time.Now()
    rule.Status = "pending"
    
    // 5. 保存更新
    if err := rm.repo.Update(ctx, rule); err != nil {
        return nil, fmt.Errorf("failed to update rule: %w", err)
    }
    
    // 6. 发布配置同步任务
    task := &ConfigSyncTask{
        Type:     "rule_update",
        RuleID:   rule.ID,
        Targets:  rule.Targets,
        Priority: "normal",
    }
    
    if err := rm.publisher.PublishTask(ctx, "config_sync", task); err != nil {
        log.Errorf("failed to publish config sync task: %v", err)
    }
    
    return rule, nil
}
```

### 3. Config Syncer (Sidecar)

**职责：** 作为Sidecar容器与监控系统集成，定时同步配置并触发reload

**实现方案：**

```go
type ConfigSyncer struct {
    AlertAgentEndpoint string
    ClusterID         string
    ConfigType        string // prometheus, alertmanager, vmalert
    ConfigPath        string
    ReloadURL         string
    SyncInterval      time.Duration
    lastConfigHash    string
    httpClient        *http.Client
}

func (cs *ConfigSyncer) Start(ctx context.Context) error {
    ticker := time.NewTicker(cs.SyncInterval)
    defer ticker.Stop()
    
    // 启动时立即同步一次
    if err := cs.syncConfig(ctx); err != nil {
        log.Errorf("Initial config sync failed: %v", err)
    }
    
    for {
        select {
        case <-ctx.Done():
            log.Info("Context cancelled, stopping config syncer")
            return ctx.Err()
        case <-ticker.C:
            if err := cs.syncConfig(ctx); err != nil {
                log.Errorf("Config sync failed: %v", err)
            }
        }
    }
}

func (cs *ConfigSyncer) syncConfig(ctx context.Context) error {
    // 1. 从AlertAgent拉取配置
    config, serverHash, err := cs.fetchConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch config: %w", err)
    }
    
    // 2. 检查配置是否有变化
    if serverHash == cs.lastConfigHash {
        log.Debug("Config unchanged, skipping sync")
        return nil
    }
    
    // 3. 验证配置格式
    if err := cs.validateConfig(config); err != nil {
        return fmt.Errorf("config validation failed: %w", err)
    }
    
    // 4. 原子性写入配置文件
    if err := cs.writeConfigFile(config); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // 5. 触发目标系统reload
    if err := cs.triggerReload(ctx); err != nil {
        return fmt.Errorf("failed to trigger reload: %w", err)
    }
    
    // 6. 更新hash并记录成功
    cs.lastConfigHash = serverHash
    log.Infof("Successfully synced %s config (hash: %s)", cs.ConfigType, serverHash)
    
    // 7. 回调AlertAgent更新同步状态
    if err := cs.reportSyncStatus(ctx, "success", ""); err != nil {
        log.Errorf("Failed to report sync status: %v", err)
    }
    
    return nil
}

func (cs *ConfigSyncer) fetchConfig(ctx context.Context) ([]byte, string, error) {
    endpoint := fmt.Sprintf("%s/api/v1/config/sync?cluster_id=%s&type=%s", 
        cs.AlertAgentEndpoint, cs.ClusterID, cs.ConfigType)
    
    req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
    if err != nil {
        return nil, "", err
    }
    
    resp, err := cs.httpClient.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("API returned status: %s", resp.Status)
    }
    
    config, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, "", err
    }
    
    serverHash := resp.Header.Get("X-Config-Hash")
    return config, serverHash, nil
}

func (cs *ConfigSyncer) writeConfigFile(config []byte) error {
    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(cs.ConfigPath), 0755); err != nil {
        return err
    }
    
    // 原子性写入：先写临时文件，再重命名
    tmpFile := cs.ConfigPath + ".tmp"
    if err := os.WriteFile(tmpFile, config, 0644); err != nil {
        return err
    }
    
    return os.Rename(tmpFile, cs.ConfigPath)
}

func (cs *ConfigSyncer) triggerReload(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, "POST", cs.ReloadURL, nil)
    if err != nil {
        return err
    }
    
    resp, err := cs.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("reload failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    log.Infof("%s reload triggered successfully", cs.ConfigType)
    return nil
}

func (cs *ConfigSyncer) reportSyncStatus(ctx context.Context, status, errorMsg string) error {
    endpoint := fmt.Sprintf("%s/api/v1/config/sync/status", cs.AlertAgentEndpoint)
    
    payload := map[string]interface{}{
        "cluster_id":   cs.ClusterID,
        "config_type":  cs.ConfigType,
        "status":       status,
        "sync_time":    time.Now().Unix(),
        "error_message": errorMsg,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := cs.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### 4. Task Producer & Worker

**职责：** 异步任务生产和消费，支持AI分析、通知发送、配置同步等任务

**Task Producer：**

```go
type TaskProducer struct {
    queue MessageQueue
}

type Task struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`
    Payload  map[string]interface{} `json:"payload"`
    Priority int                    `json:"priority"`
    Retry    int                    `json:"retry"`
    MaxRetry int                    `json:"max_retry"`
    CreateAt time.Time              `json:"create_at"`
}

func (tp *TaskProducer) PublishTask(ctx context.Context, queueName string, task *Task) error {
    task.ID = uuid.New().String()
    task.CreateAt = time.Now()
    
    taskData, err := json.Marshal(task)
    if err != nil {
        return fmt.Errorf("failed to marshal task: %w", err)
    }
    
    return tp.queue.Publish(ctx, queueName, taskData)
}

// 发布AI分析任务
func (tp *TaskProducer) PublishAIAnalysisTask(ctx context.Context, alertID string, alertData map[string]interface{}) error {
    task := &Task{
        Type:     "ai_analysis",
        Priority: 1,
        MaxRetry: 3,
        Payload: map[string]interface{}{
            "alert_id":   alertID,
            "alert_data": alertData,
            "analysis_type": "root_cause",
        },
    }
    
    return tp.PublishTask(ctx, "ai_analysis", task)
}

// 发布通知任务
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
```

**Worker实现：**

```go
type Worker struct {
    id          string
    queueName   string
    queue       MessageQueue
    handlers    map[string]TaskHandler
    concurrency int
    stopCh      chan struct{}
}

type TaskHandler interface {
    Handle(ctx context.Context, task *Task) error
    Type() string
}

func (w *Worker) Start(ctx context.Context) error {
    log.Infof("Starting worker %s for queue %s with concurrency %d", w.id, w.queueName, w.concurrency)
    
    for i := 0; i < w.concurrency; i++ {
        go w.processLoop(ctx, i)
    }
    
    <-ctx.Done()
    close(w.stopCh)
    return nil
}

func (w *Worker) processLoop(ctx context.Context, workerIndex int) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-w.stopCh:
            return
        default:
            if err := w.processTask(ctx); err != nil {
                log.Errorf("Worker %s-%d process task error: %v", w.id, workerIndex, err)
                time.Sleep(time.Second) // 避免错误循环
            }
        }
    }
}

func (w *Worker) processTask(ctx context.Context) error {
    // 1. 从队列获取任务
    taskData, err := w.queue.Consume(ctx, w.queueName)
    if err != nil {
        return fmt.Errorf("failed to consume task: %w", err)
    }
    
    if taskData == nil {
        time.Sleep(100 * time.Millisecond) // 队列为空时短暂等待
        return nil
    }
    
    // 2. 解析任务
    var task Task
    if err := json.Unmarshal(taskData, &task); err != nil {
        return fmt.Errorf("failed to unmarshal task: %w", err)
    }
    
    // 3. 查找处理器
    handler, exists := w.handlers[task.Type]
    if !exists {
        return fmt.Errorf("no handler found for task type: %s", task.Type)
    }
    
    // 4. 执行任务
    startTime := time.Now()
    err = handler.Handle(ctx, &task)
    duration := time.Since(startTime)
    
    // 5. 处理结果
    if err != nil {
        log.Errorf("Task %s failed: %v (duration: %v)", task.ID, err, duration)
        
        // 重试逻辑
        if task.Retry < task.MaxRetry {
            task.Retry++
            if retryErr := w.retryTask(ctx, &task); retryErr != nil {
                log.Errorf("Failed to retry task %s: %v", task.ID, retryErr)
            }
        } else {
            log.Errorf("Task %s exceeded max retries, moving to dead letter queue", task.ID)
            w.moveToDeadLetter(ctx, &task, err)
        }
        
        return err
    }
    
    log.Infof("Task %s completed successfully (duration: %v)", task.ID, duration)
    return nil
}

// AI分析任务处理器
type AIAnalysisHandler struct {
    difyClient *DifyClient
    alertRepo  repository.AlertRepository
}

func (h *AIAnalysisHandler) Type() string {
    return "ai_analysis"
}

func (h *AIAnalysisHandler) Handle(ctx context.Context, task *Task) error {
    alertID := task.Payload["alert_id"].(string)
    alertData := task.Payload["alert_data"].(map[string]interface{})
    analysisType := task.Payload["analysis_type"].(string)
    
    // 1. 更新告警状态为分析中
    if err := h.alertRepo.UpdateStatus(ctx, alertID, "analyzing"); err != nil {
        return fmt.Errorf("failed to update alert status: %w", err)
    }
    
    // 2. 调用Dify AI进行分析
    analysisResult, err := h.difyClient.AnalyzeAlert(ctx, alertData, analysisType)
    if err != nil {
        // 更新状态为分析失败
        h.alertRepo.UpdateStatus(ctx, alertID, "analysis_failed")
        return fmt.Errorf("AI analysis failed: %w", err)
    }
    
    // 3. 保存分析结果
    if err := h.alertRepo.SaveAnalysisResult(ctx, alertID, analysisResult); err != nil {
        return fmt.Errorf("failed to save analysis result: %w", err)
    }
    
    // 4. 更新告警状态为已分析
    if err := h.alertRepo.UpdateStatus(ctx, alertID, "analyzed"); err != nil {
        return fmt.Errorf("failed to update alert status: %w", err)
    }
    
    log.Infof("AI analysis completed for alert %s", alertID)
    return nil
}
```

### 5. Plugin Manager

**职责：** 通知插件的注册、加载、管理和调用

**插件接口定义：**

```go
type NotificationPlugin interface {
    // 插件基本信息
    Name() string
    Version() string
    Description() string
    
    // 配置相关
    ConfigSchema() map[string]interface{} // JSON Schema
    ValidateConfig(config map[string]interface{}) error
    
    // 通知发送
    Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error
    
    // 健康检查
    HealthCheck(ctx context.Context, config map[string]interface{}) error
}

type NotificationMessage struct {
    Title       string                 `json:"title"`
    Content     string                 `json:"content"`
    Severity    string                 `json:"severity"`
    AlertID     string                 `json:"alert_id"`
    Timestamp   time.Time              `json:"timestamp"`
    Labels      map[string]string      `json:"labels"`
    Annotations map[string]string      `json:"annotations"`
    Extra       map[string]interface{} `json:"extra"`
}

type PluginManager struct {
    plugins map[string]NotificationPlugin
    configs map[string]map[string]interface{}
    mutex   sync.RWMutex
}

func (pm *PluginManager) RegisterPlugin(plugin NotificationPlugin) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    name := plugin.Name()
    if _, exists := pm.plugins[name]; exists {
        return fmt.Errorf("plugin %s already registered", name)
    }
    
    pm.plugins[name] = plugin
    log.Infof("Plugin %s v%s registered successfully", name, plugin.Version())
    return nil
}

func (pm *PluginManager) GetAvailablePlugins() []PluginInfo {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    var plugins []PluginInfo
    for name, plugin := range pm.plugins {
        plugins = append(plugins, PluginInfo{
            Name:        name,
            Version:     plugin.Version(),
            Description: plugin.Description(),
            Schema:      plugin.ConfigSchema(),
        })
    }
    
    return plugins
}

func (pm *PluginManager) SendNotification(ctx context.Context, pluginName string, config map[string]interface{}, message *NotificationMessage) error {
    pm.mutex.RLock()
    plugin, exists := pm.plugins[pluginName]
    pm.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("plugin %s not found", pluginName)
    }
    
    // 验证配置
    if err := plugin.ValidateConfig(config); err != nil {
        return fmt.Errorf("invalid config for plugin %s: %w", pluginName, err)
    }
    
    // 发送通知
    return plugin.Send(ctx, config, message)
}
```

**钉钉插件示例：**

```go
type DingTalkPlugin struct{}

func (d *DingTalkPlugin) Name() string {
    return "dingtalk"
}

func (d *DingTalkPlugin) Version() string {
    return "1.0.0"
}

func (d *DingTalkPlugin) Description() string {
    return "钉钉群机器人通知插件"
}

func (d *DingTalkPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "webhook_url": map[string]interface{}{
                "type":        "string",
                "description": "钉钉机器人Webhook URL",
                "required":    true,
            },
            "secret": map[string]interface{}{
                "type":        "string",
                "description": "钉钉机器人密钥",
                "required":    false,
            },
            "at_mobiles": map[string]interface{}{
                "type":        "array",
                "description": "@指定手机号",
                "items": map[string]interface{}{
                    "type": "string",
                },
            },
            "at_all": map[string]interface{}{
                "type":        "boolean",
                "description": "是否@所有人",
                "default":     false,
            },
        },
    }
}

func (d *DingTalkPlugin) ValidateConfig(config map[string]interface{}) error {
    webhookURL, ok := config["webhook_url"].(string)
    if !ok || webhookURL == "" {
        return fmt.Errorf("webhook_url is required")
    }
    
    if !strings.HasPrefix(webhookURL, "https://oapi.dingtalk.com/robot/send") {
        return fmt.Errorf("invalid dingtalk webhook URL")
    }
    
    return nil
}

func (d *DingTalkPlugin) Send(ctx context.Context, config map[string]interface{}, message *NotificationMessage) error {
    webhookURL := config["webhook_url"].(string)
    secret, _ := config["secret"].(string)
    atMobiles, _ := config["at_mobiles"].([]interface{})
    atAll, _ := config["at_all"].(bool)
    
    // 构建钉钉消息格式
    dingMessage := map[string]interface{}{
        "msgtype": "markdown",
        "markdown": map[string]interface{}{
            "title": message.Title,
            "text":  d.formatMessage(message),
        },
    }
    
    // 添加@信息
    if len(atMobiles) > 0 || atAll {
        at := map[string]interface{}{
            "isAtAll": atAll,
        }
        
        if len(atMobiles) > 0 {
            mobiles := make([]string, len(atMobiles))
            for i, mobile := range atMobiles {
                mobiles[i] = mobile.(string)
            }
            at["atMobiles"] = mobiles
        }
        
        dingMessage["at"] = at
    }
    
    // 计算签名
    if secret != "" {
        timestamp := time.Now().UnixMilli()
        sign := d.calculateSign(timestamp, secret)
        webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
    }
    
    // 发送请求
    jsonData, err := json.Marshal(dingMessage)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("dingtalk API returned status %d: %s", resp.StatusCode, string(body))
    }
    
    return nil
}

func (d *DingTalkPlugin) formatMessage(message *NotificationMessage) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("## %s\n\n", message.Title))
    builder.WriteString(fmt.Sprintf("**告警级别**: %s\n\n", message.Severity))
    builder.WriteString(fmt.Sprintf("**告警时间**: %s\n\n", message.Timestamp.Format("2006-01-02 15:04:05")))
    builder.WriteString(fmt.Sprintf("**告警内容**: %s\n\n", message.Content))
    
    if len(message.Labels) > 0 {
        builder.WriteString("**标签信息**:\n\n")
        for key, value := range message.Labels {
            builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
        }
        builder.WriteString("\n")
    }
    
    if len(message.Annotations) > 0 {
        builder.WriteString("**注释信息**:\n\n")
        for key, value := range message.Annotations {
            builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
        }
    }
    
    return builder.String()
}

func (d *DingTalkPlugin) calculateSign(timestamp int64, secret string) string {
    stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(stringToSign))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (d *DingTalkPlugin) HealthCheck(ctx context.Context, config map[string]interface{}) error {
    // 发送测试消息验证配置
    testMessage := &NotificationMessage{
        Title:     "健康检查",
        Content:   "这是一条测试消息，用于验证钉钉通知配置是否正确",
        Severity:  "info",
        Timestamp: time.Now(),
    }
    
    return d.Send(ctx, config, testMessage)
}
```

## 数据模型

### 数据库表设计

```sql
-- 告警规则表
CREATE TABLE alert_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    expression TEXT NOT NULL,
    duration VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    labels JSON,
    annotations JSON,
    targets JSON,
    version VARCHAR(50) NOT NULL DEFAULT 'v1.0.0',
    status ENUM('pending', 'active', 'inactive', 'error') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- 配置同步状态表
CREATE TABLE config_sync_status (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(100) NOT NULL,
    config_type VARCHAR(50) NOT NULL,
    config_hash VARCHAR(64),
    sync_status ENUM('success', 'failed', 'pending') DEFAULT 'pending',
    sync_time TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_cluster_type (cluster_id, config_type),
    INDEX idx_sync_status (sync_status),
    INDEX idx_sync_time (sync_time)
);

-- 任务队列表（用于持久化重要任务）
CREATE TABLE task_queue (
    id VARCHAR(36) PRIMARY KEY,
    queue_name VARCHAR(100) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    payload JSON NOT NULL,
    priority INT DEFAULT 0,
    retry_count INT DEFAULT 0,
    max_retry INT DEFAULT 3,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_queue_status (queue_name, status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_priority (priority)
);

-- 通知插件配置表
CREATE TABLE notification_plugins (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    config_schema JSON,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_enabled (enabled)
);

-- 用户通知配置表
CREATE TABLE user_notification_configs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    plugin_name VARCHAR(100) NOT NULL,
    config JSON NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_name) REFERENCES notification_plugins(name) ON DELETE CASCADE,
    UNIQUE KEY uk_user_plugin (user_id, plugin_name),
    INDEX idx_user_id (user_id),
    INDEX idx_enabled (enabled)
);

-- 扩展现有告警表
ALTER TABLE alerts ADD COLUMN analysis_status ENUM('pending', 'analyzing', 'analyzed', 'failed') DEFAULT 'pending';
ALTER TABLE alerts ADD COLUMN analysis_result JSON;
ALTER TABLE alerts ADD COLUMN task_id VARCHAR(36);
ALTER TABLE alerts ADD COLUMN notification_status ENUM('pending', 'sent', 'failed') DEFAULT 'pending';
```

## 错误处理

### 错误分类和处理策略

```go
// 错误类型定义
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeNotFound     ErrorType = "not_found"
    ErrorTypeConflict     ErrorType = "conflict"
    ErrorTypeInternal     ErrorType = "internal"
    ErrorTypeExternal     ErrorType = "external"
    ErrorTypeTimeout      ErrorType = "timeout"
    ErrorTypeRateLimit    ErrorType = "rate_limit"
)

type AppError struct {
    Type    ErrorType `json:"type"`
    Code    string    `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
    Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var appErr *AppError
            if errors.As(err, &appErr) {
                handleAppError(c, appErr)
            } else {
                handleUnknownError(c, err)
            }
        }
    }
}

func handleAppError(c *gin.Context, err *AppError) {
    var statusCode int
    
    switch err.Type {
    case ErrorTypeValidation:
        statusCode = http.StatusBadRequest
    case ErrorTypeNotFound:
        statusCode = http.StatusNotFound
    case ErrorTypeConflict:
        statusCode = http.StatusConflict
    case ErrorTypeTimeout:
        statusCode = http.StatusRequestTimeout
    case ErrorTypeRateLimit:
        statusCode = http.StatusTooManyRequests
    case ErrorTypeExternal:
        statusCode = http.StatusBadGateway
    default:
        statusCode = http.StatusInternalServerError
    }
    
    c.JSON(statusCode, gin.H{
        "error": gin.H{
            "type":    err.Type,
            "code":    err.Code,
            "message": err.Message,
            "details": err.Details,
        },
        "timestamp": time.Now().Unix(),
        "path":      c.Request.URL.Path,
    })
}

// 重试机制
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
    var lastErr error
    delay := config.BaseDelay
    
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
            
            if attempt == config.MaxAttempts {
                break
            }
            
            // 指数退避
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
                delay = time.Duration(float64(delay) * config.Multiplier)
                if delay > config.MaxDelay {
                    delay = config.MaxDelay
                }
            }
        }
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
```

## 测试策略

### 单元测试

```go
// 规则管理器测试
func TestRuleManager_CreateRule(t *testing.T) {
    tests := []struct {
        name    string
        req     *CreateRuleRequest
        wantErr bool
        errType ErrorType
    }{
        {
            name: "valid rule",
            req: &CreateRuleRequest{
                Name:       "test-rule",
                Expression: "cpu_usage > 80",
                Duration:   "5m",
                Severity:   "warning",
                Labels:     map[string]string{"team": "platform"},
                Targets:    []string{"cluster-1"},
            },
            wantErr: false,
        },
        {
            name: "invalid expression",
            req: &CreateRuleRequest{
                Name:       "test-rule",
                Expression: "invalid expression",
                Duration:   "5m",
                Severity:   "warning",
            },
            wantErr: true,
            errType: ErrorTypeValidation,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := &MockRuleRepository{}
            mockValidator := &MockRuleValidator{}
            mockPublisher := &MockTaskPublisher{}
            
            if tt.wantErr && tt.errType == ErrorTypeValidation {
                mockValidator.On("ValidateRule", tt.req.Expression, tt.req.Duration).
                    Return(fmt.Errorf("invalid expression"))
            } else {
                mockValidator.On("ValidateRule", tt.req.Expression, tt.req.Duration).
                    Return(nil)
                mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
                mockPublisher.On("PublishTask", mock.Anything, mock.Anything, mock.Anything).Return(nil)
            }
            
            rm := &RuleManager{
                repo:      mockRepo,
                validator: mockValidator,
                publisher: mockPublisher,
            }
            
            // Execute
            rule, err := rm.CreateRule(context.Background(), tt.req)
            
            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, rule)
                
                var appErr *AppError
                if errors.As(err, &appErr) {
                    assert.Equal(t, tt.errType, appErr.Type)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, rule)
                assert.Equal(t, tt.req.Name, rule.Name)
                assert.Equal(t, tt.req.Expression, rule.Expression)
            }
            
            // Verify mocks
            mockValidator.AssertExpectations(t)
            if !tt.wantErr {
                mockRepo.AssertExpectations(t)
                mockPublisher.AssertExpectations(t)
            }
        })
    }
}
```

### 集成测试

```go
func TestConfigSyncerIntegration(t *testing.T) {
    // Setup test environment
    testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/api/v1/config/sync":
            w.Header().Set("X-Config-Hash", "test-hash-123")
            w.Header().Set("Content-Type", "application/yaml")
            w.WriteString("test: config\nversion: 1.0")
        case "/api/v1/config/sync/status":
            w.WriteHeader(http.StatusOK)
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer testServer.Close()
    
    // Setup mock reload server
    reloadCalled := false
    reloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" && r.URL.Path == "/-/reload" {
            reloadCalled = true
            w.WriteHeader(http.StatusOK)
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer reloadServer.Close()
    
    // Create temp config file
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "test-config.yml")
    
    // Create config syncer
    syncer := &ConfigSyncer{
        AlertAgentEndpoint: testServer.URL,
        ClusterID:         "test-cluster",
        ConfigType:        "prometheus",
        ConfigPath:        configPath,
        ReloadURL:         reloadServer.URL + "/-/reload",
        SyncInterval:      100 * time.Millisecond,
        httpClient:        &http.Client{Timeout: 5 * time.Second},
    }
    
    // Test sync
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    err := syncer.syncConfig(ctx)
    assert.NoError(t, err)
    
    // Verify config file was written
    configData, err := os.ReadFile(configPath)
    assert.NoError(t, err)
    assert.Contains(t, string(configData), "test: config")
    assert.Contains(t, string(configData), "version: 1.0")
    
    // Verify reload was called
    assert.True(t, reloadCalled, "Reload endpoint should have been called")
    
    // Verify hash was updated
    assert.Equal(t, "test-hash-123", syncer.lastConfigHash)
}
```

这个设计文档详细描述了AlertAgent重构的技术架构和实现方案，涵盖了统一API接口、Sidecar集成、异步任务处理、插件化通知等核心需求。接下来我将创建对应的任务列表。