# 设计文档

## 概述

本设计文档详细描述了AlertAgent架构重新设计的技术实现方案。该重新设计将AlertAgent从独立的告警管理系统转变为智能告警管理和分发中心，与Alertmanager、n8n+Dify形成完整的智能告警处理生态系统。

## 架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AlertAgent 智能告警生态系统                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │   AlertAgent    │    │  Alertmanager   │    │   n8n + Dify    │         │
│  │   智能管理中心    │    │    告警引擎      │    │   智能分析层     │         │
│  │                 │    │                 │    │                 │         │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │         │
│  │ │ 渠道管理器   │ │────│ │ 规则执行器   │ │    │ │ 告警分析器   │ │         │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │         │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │         │
│  │ │ 集群管理器   │ │────│ │ 告警路由器   │ │────│ │ 工作流引擎   │ │         │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │         │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │         │
│  │ │ 智能网关     │ │────│ │ 通知分发器   │ │────│ │ 决策引擎     │ │         │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │         │
│  │ ┌─────────────┐ │    │                 │    │ ┌─────────────┐ │         │
│  │ │ API 网关     │ │    │                 │    │ │ 知识库       │ │         │
│  │ └─────────────┘ │    │                 │    │ └─────────────┘ │         │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘         │
│           │                       │                       │                │
│           ▼                       ▼                       ▼                │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │   数据存储层     │    │   监控数据源     │    │   外部通知渠道   │         │
│  │                 │    │                 │    │                 │         │
│  │ • MySQL         │    │ • Prometheus    │    │ • 钉钉           │         │
│  │ • Redis         │    │ • VictoriaMetrics│   │ • 企业微信       │         │
│  │ • InfluxDB      │    │ • Grafana       │    │ • 邮件           │         │
│  │                 │    │                 │    │ • Webhook        │         │
│  │                 │    │                 │    │ • Slack          │         │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 核心架构原则

1. **微服务架构**：各组件松耦合，独立部署和扩展
2. **事件驱动**：基于事件的异步处理模式
3. **插件化设计**：支持渠道和功能的插件扩展
4. **配置驱动**：通过配置文件和API动态调整行为
5. **高可用性**：支持集群部署和故障转移

## 组件和接口

### 1. 告警渠道管理器 (Channel Manager)

#### 核心接口设计

```go
// ChannelManager 渠道管理器接口
type ChannelManager interface {
    // 渠道生命周期管理
    CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error)
    UpdateChannel(ctx context.Context, id string, req *UpdateChannelRequest) (*Channel, error)
    DeleteChannel(ctx context.Context, id string) error
    GetChannel(ctx context.Context, id string) (*Channel, error)
    ListChannels(ctx context.Context, query *ChannelQuery) ([]*Channel, int64, error)
    
    // 消息发送
    SendMessage(ctx context.Context, channelID string, message *Message) error
    BroadcastMessage(ctx context.Context, channelIDs []string, message *Message) error
    
    // 健康检查和测试
    TestChannel(ctx context.Context, id string) (*TestResult, error)
    GetChannelHealth(ctx context.Context, id string) (*HealthStatus, error)
    
    // 插件管理
    RegisterPlugin(plugin ChannelPlugin) error
    GetAvailablePlugins() []PluginInfo
}

// ChannelPlugin 渠道插件接口
type ChannelPlugin interface {
    GetType() string
    GetName() string
    GetConfigSchema() *ConfigSchema
    ValidateConfig(config map[string]interface{}) error
    TestConnection(config map[string]interface{}) error
    SendMessage(config map[string]interface{}, message *Message) error
    GetHealthStatus(config map[string]interface{}) (*HealthStatus, error)
}
```

#### 插件架构实现

```go
// 插件管理器
type PluginManager struct {
    plugins map[string]ChannelPlugin
    mutex   sync.RWMutex
    logger  *zap.Logger
}

func (pm *PluginManager) RegisterPlugin(plugin ChannelPlugin) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pluginType := plugin.GetType()
    if _, exists := pm.plugins[pluginType]; exists {
        return fmt.Errorf("plugin type %s already registered", pluginType)
    }
    
    pm.plugins[pluginType] = plugin
    pm.logger.Info("Plugin registered", zap.String("type", pluginType))
    return nil
}

// 内置插件实现示例
type DingTalkPlugin struct{}

func (p *DingTalkPlugin) GetType() string { return "dingtalk" }
func (p *DingTalkPlugin) GetName() string { return "钉钉" }

func (p *DingTalkPlugin) GetConfigSchema() *ConfigSchema {
    return &ConfigSchema{
        Fields: []ConfigField{
            {
                Name:     "webhook_url",
                Type:     "string",
                Label:    "Webhook URL",
                Required: true,
                Validation: &ValidationRule{
                    Pattern: `^https://oapi\.dingtalk\.com/robot/send\?access_token=.*`,
                },
            },
            {
                Name:     "secret",
                Type:     "password",
                Label:    "签名密钥",
                Required: false,
            },
        },
    }
}
```

### 2. 集群管理器 (Cluster Manager)

#### 核心接口设计

```go
// ClusterManager 集群管理器接口
type ClusterManager interface {
    // 集群管理
    RegisterCluster(ctx context.Context, config *ClusterConfig) error
    UpdateCluster(ctx context.Context, id string, config *ClusterConfig) error
    DeleteCluster(ctx context.Context, id string) error
    GetCluster(ctx context.Context, id string) (*Cluster, error)
    ListClusters(ctx context.Context) ([]*Cluster, error)
    
    // 健康检查
    HealthCheck(ctx context.Context) map[string]*HealthStatus
    GetClusterHealth(ctx context.Context, id string) (*HealthStatus, error)
    
    // 配置分发
    DistributeConfig(ctx context.Context, clusterID string, config *Config) error
    GetSyncStatus(ctx context.Context, clusterID string) (*SyncStatus, error)
}

// ConfigSynchronizer 配置同步器接口
type ConfigSynchronizer interface {
    SyncConfig(ctx context.Context, cluster *Cluster, config *Config) error
    GetSyncStatus(ctx context.Context, clusterID string) (*SyncStatus, error)
}
```

#### Sidecar配置同步实现

```go
// SidecarSynchronizer Sidecar模式配置同步器
type SidecarSynchronizer struct {
    httpClient *http.Client
    logger     *zap.Logger
}

func (s *SidecarSynchronizer) SyncConfig(ctx context.Context, cluster *Cluster, config *Config) error {
    // 1. 渲染配置模板
    renderedConfig, err := s.renderConfig(config, cluster)
    if err != nil {
        return fmt.Errorf("failed to render config: %w", err)
    }
    
    // 2. 计算配置哈希
    configHash := s.calculateHash(renderedConfig)
    
    // 3. 检查是否需要更新
    if s.isConfigUpToDate(cluster, configHash) {
        return nil
    }
    
    // 4. 通过API暴露配置给Sidecar拉取
    s.exposeConfig(cluster.ID, renderedConfig, configHash)
    
    // 5. 通知Sidecar进行同步
    return s.notifySidecar(cluster)
}

// 配置API端点实现
func (s *SidecarSynchronizer) HandleConfigRequest(c *gin.Context) {
    clusterID := c.Query("cluster_id")
    configType := c.Query("type")
    
    config, hash, err := s.getClusterConfig(clusterID, configType)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.Header("X-Config-Hash", hash)
    c.Header("Content-Type", "application/yaml")
    c.String(200, config)
}
```

### 3. 智能网关 (Smart Gateway)

#### 核心接口设计

```go
// SmartGateway 智能网关接口
type SmartGateway interface {
    // 告警接收和处理
    ReceiveAlert(ctx context.Context, alert *Alert) error
    ProcessAlert(ctx context.Context, alert *Alert) (*ProcessingResult, error)
    
    // 告警收敛和抑制
    ConvergeAlerts(ctx context.Context, alerts []*Alert) ([]*Alert, error)
    SuppressAlert(ctx context.Context, alert *Alert) (bool, error)
    
    // 智能路由
    RouteAlert(ctx context.Context, alert *Alert) (*RoutingDecision, error)
}

// AlertProcessor 告警处理器
type AlertProcessor struct {
    convergenceEngine *ConvergenceEngine
    suppressionEngine *SuppressionEngine
    routingEngine     *RoutingEngine
    logger           *zap.Logger
}
```

#### 告警收敛实现

```go
// ConvergenceEngine 告警收敛引擎
type ConvergenceEngine struct {
    redis  *redis.Client
    config *ConvergenceConfig
    logger *zap.Logger
}

func (ce *ConvergenceEngine) ConvergeAlert(ctx context.Context, alert *Alert) (*ConvergenceResult, error) {
    // 1. 生成收敛键
    convergenceKey := ce.generateConvergenceKey(alert)
    
    // 2. 检查收敛窗口
    window := ce.getConvergenceWindow(convergenceKey)
    if window == nil {
        // 创建新的收敛窗口
        window = ce.createConvergenceWindow(convergenceKey, alert)
    }
    
    // 3. 添加告警到收敛窗口
    window.AddAlert(alert)
    
    // 4. 检查是否达到收敛条件
    if window.ShouldTrigger() {
        return &ConvergenceResult{
            Action:           "trigger",
            ConvergedAlerts:  window.GetAlerts(),
            RepresentativeAlert: window.GetRepresentativeAlert(),
        }, nil
    }
    
    return &ConvergenceResult{
        Action: "converged",
    }, nil
}

func (ce *ConvergenceEngine) generateConvergenceKey(alert *Alert) string {
    // 基于告警标签生成收敛键
    labels := []string{
        alert.Labels["alertname"],
        alert.Labels["instance"],
        alert.Labels["job"],
    }
    return strings.Join(labels, ":")
}
```

### 4. AI分析集成

#### n8n工作流集成

```go
// N8NClient n8n客户端
type N8NClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *zap.Logger
}

func (nc *N8NClient) TriggerWorkflow(ctx context.Context, workflowName string, data interface{}) (*WorkflowExecution, error) {
    webhookURL := fmt.Sprintf("%s/webhook/%s", nc.baseURL, workflowName)
    
    payload, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal data: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := nc.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    var execution WorkflowExecution
    if err := json.NewDecoder(resp.Body).Decode(&execution); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &execution, nil
}
```

#### Dify AI分析集成

```go
// DifyClient Dify AI客户端
type DifyClient struct {
    baseURL string
    apiKey  string
    httpClient *http.Client
    logger  *zap.Logger
}

func (dc *DifyClient) AnalyzeAlert(ctx context.Context, alert *Alert, analysisType string) (*AnalysisResult, error) {
    analysisRequest := &DifyAnalysisRequest{
        Inputs: map[string]interface{}{
            "alert_data":     alert,
            "analysis_type":  analysisType,
            "context":        dc.buildContext(alert),
        },
        Query:          "请分析这个告警的根因并提供处理建议",
        User:           "alertagent-system",
        ConversationID: fmt.Sprintf("alert-%d", alert.ID),
    }
    
    payload, err := json.Marshal(analysisRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/v1/chat-messages", dc.baseURL), 
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", dc.apiKey))
    
    resp, err := dc.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    var difyResponse DifyResponse
    if err := json.NewDecoder(resp.Body).Decode(&difyResponse); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return dc.parseAnalysisResult(&difyResponse)
}
```

## 数据模型

### 核心数据结构

```go
// Channel 告警渠道
type Channel struct {
    ID          string                 `json:"id" gorm:"primarykey"`
    Name        string                 `json:"name" gorm:"size:100;not null"`
    Type        string                 `json:"type" gorm:"size:50;not null"`
    Description string                 `json:"description" gorm:"type:text"`
    Config      map[string]interface{} `json:"config" gorm:"type:json"`
    GroupID     string                 `json:"group_id" gorm:"size:36"`
    Tags        []string               `json:"tags" gorm:"type:json"`
    Status      ChannelStatus          `json:"status" gorm:"type:varchar(20);default:'active'"`
    HealthStatus HealthStatus          `json:"health_status"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// Cluster Alertmanager集群
type Cluster struct {
    ID                  string            `json:"id" gorm:"primarykey"`
    Name                string            `json:"name" gorm:"size:100;not null;uniqueIndex"`
    Endpoint            string            `json:"endpoint" gorm:"size:255;not null"`
    ConfigPath          string            `json:"config_path" gorm:"size:255"`
    RulesPath           string            `json:"rules_path" gorm:"size:255"`
    SyncInterval        int               `json:"sync_interval" gorm:"default:30"`
    HealthCheckInterval int               `json:"health_check_interval" gorm:"default:10"`
    Status              ClusterStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
    Labels              map[string]string `json:"labels" gorm:"type:json"`
    CreatedAt           time.Time         `json:"created_at"`
    UpdatedAt           time.Time         `json:"updated_at"`
}

// AlertProcessingRecord 告警处理记录
type AlertProcessingRecord struct {
    ID               string                 `json:"id" gorm:"primarykey"`
    AlertID          string                 `json:"alert_id" gorm:"size:100;not null;index"`
    AlertName        string                 `json:"alert_name" gorm:"size:100"`
    Severity         string                 `json:"severity" gorm:"size:20"`
    ClusterID        string                 `json:"cluster_id" gorm:"size:36;index"`
    ReceivedAt       time.Time              `json:"received_at"`
    ProcessedAt      *time.Time             `json:"processed_at"`
    ProcessingStatus ProcessingStatus       `json:"processing_status" gorm:"type:varchar(20);default:'received'"`
    AnalysisID       string                 `json:"analysis_id" gorm:"size:36"`
    Decision         map[string]interface{} `json:"decision" gorm:"type:json"`
    ActionTaken      string                 `json:"action_taken" gorm:"size:100"`
    ResolutionTime   int                    `json:"resolution_time"` // 秒
    FeedbackScore    float64                `json:"feedback_score" gorm:"type:decimal(3,2)"`
    Labels           map[string]string      `json:"labels" gorm:"type:json"`
    Annotations      map[string]string      `json:"annotations" gorm:"type:json"`
    CreatedAt        time.Time              `json:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at"`
}

// AIAnalysisRecord AI分析记录
type AIAnalysisRecord struct {
    ID              string                 `json:"id" gorm:"primarykey"`
    AlertID         string                 `json:"alert_id" gorm:"size:100;not null;index"`
    AnalysisType    string                 `json:"analysis_type" gorm:"size:50;default:'root_cause_analysis'"`
    RequestData     map[string]interface{} `json:"request_data" gorm:"type:json"`
    ResponseData    map[string]interface{} `json:"response_data" gorm:"type:json"`
    AnalysisResult  map[string]interface{} `json:"analysis_result" gorm:"type:json"`
    ConfidenceScore float64                `json:"confidence_score" gorm:"type:decimal(3,2)"`
    ProcessingTime  int                    `json:"processing_time"` // 毫秒
    Status          AnalysisStatus         `json:"status" gorm:"type:varchar(20);default:'pending'"`
    ErrorMessage    string                 `json:"error_message" gorm:"type:text"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
}
```

### 数据库迁移策略

```go
// Migration 数据库迁移
type Migration struct {
    db     *gorm.DB
    logger *zap.Logger
}

func (m *Migration) MigrateFromV1() error {
    // 1. 创建新表结构
    if err := m.createNewTables(); err != nil {
        return fmt.Errorf("failed to create new tables: %w", err)
    }
    
    // 2. 迁移现有数据
    if err := m.migrateExistingData(); err != nil {
        return fmt.Errorf("failed to migrate existing data: %w", err)
    }
    
    // 3. 创建索引
    if err := m.createIndexes(); err != nil {
        return fmt.Errorf("failed to create indexes: %w", err)
    }
    
    // 4. 验证数据完整性
    if err := m.validateDataIntegrity(); err != nil {
        return fmt.Errorf("data integrity validation failed: %w", err)
    }
    
    return nil
}

func (m *Migration) migrateExistingData() error {
    // 迁移告警数据
    if err := m.migrateAlerts(); err != nil {
        return err
    }
    
    // 迁移规则数据
    if err := m.migrateRules(); err != nil {
        return err
    }
    
    // 迁移通知配置
    if err := m.migrateNotificationConfig(); err != nil {
        return err
    }
    
    return nil
}
```

## 错误处理

### 错误分类和处理策略

```go
// ErrorType 错误类型
type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeNotFound      ErrorType = "not_found"
    ErrorTypeConflict      ErrorType = "conflict"
    ErrorTypeInternal      ErrorType = "internal"
    ErrorTypeExternal      ErrorType = "external"
    ErrorTypeTimeout       ErrorType = "timeout"
    ErrorTypeRateLimit     ErrorType = "rate_limit"
)

// AppError 应用错误
type AppError struct {
    Type    ErrorType `json:"type"`
    Code    string    `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
    Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// ErrorHandler 错误处理器
type ErrorHandler struct {
    logger *zap.Logger
}

func (eh *ErrorHandler) HandleError(c *gin.Context, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        eh.handleAppError(c, appErr)
    } else {
        eh.handleUnknownError(c, err)
    }
}

func (eh *ErrorHandler) handleAppError(c *gin.Context, err *AppError) {
    statusCode := eh.getStatusCode(err.Type)
    
    response := gin.H{
        "code":    statusCode,
        "type":    err.Type,
        "message": err.Message,
    }
    
    if err.Details != "" {
        response["details"] = err.Details
    }
    
    eh.logger.Error("Application error",
        zap.String("type", string(err.Type)),
        zap.String("code", err.Code),
        zap.String("message", err.Message),
        zap.Error(err.Cause),
    )
    
    c.JSON(statusCode, response)
}
```

### 重试和熔断机制

```go
// RetryConfig 重试配置
type RetryConfig struct {
    MaxAttempts int           `json:"max_attempts"`
    BaseDelay   time.Duration `json:"base_delay"`
    MaxDelay    time.Duration `json:"max_delay"`
    Multiplier  float64       `json:"multiplier"`
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
    name           string
    maxFailures    int
    resetTimeout   time.Duration
    state          CircuitState
    failures       int
    lastFailTime   time.Time
    mutex          sync.RWMutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.RLock()
    state := cb.state
    cb.mutex.RUnlock()
    
    if state == CircuitStateOpen {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.setState(CircuitStateHalfOpen)
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.recordSuccess()
    return nil
}
```

## 测试策略

### 单元测试

```go
// ChannelManager单元测试示例
func TestChannelManager_CreateChannel(t *testing.T) {
    tests := []struct {
        name    string
        request *CreateChannelRequest
        want    *Channel
        wantErr bool
    }{
        {
            name: "valid dingtalk channel",
            request: &CreateChannelRequest{
                Name: "测试钉钉群",
                Type: "dingtalk",
                Config: map[string]interface{}{
                    "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=test",
                },
            },
            want: &Channel{
                Name: "测试钉钉群",
                Type: "dingtalk",
                Status: ChannelStatusActive,
            },
            wantErr: false,
        },
        {
            name: "invalid config",
            request: &CreateChannelRequest{
                Name: "无效配置",
                Type: "dingtalk",
                Config: map[string]interface{}{
                    "invalid_field": "value",
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cm := NewChannelManager(mockDB, mockRedis, mockPluginManager)
            got, err := cm.CreateChannel(context.Background(), tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want.Name, got.Name)
            assert.Equal(t, tt.want.Type, got.Type)
            assert.Equal(t, tt.want.Status, got.Status)
        })
    }
}
```

### 集成测试

```go
// 集成测试示例
func TestAlertProcessingFlow(t *testing.T) {
    // 设置测试环境
    testEnv := setupTestEnvironment(t)
    defer testEnv.Cleanup()
    
    // 创建测试告警
    alert := &Alert{
        Name:     "test-alert",
        Severity: "warning",
        Labels: map[string]string{
            "alertname": "HighCPUUsage",
            "instance":  "server-1",
        },
    }
    
    // 发送告警到智能网关
    err := testEnv.SmartGateway.ReceiveAlert(context.Background(), alert)
    assert.NoError(t, err)
    
    // 验证告警处理记录
    time.Sleep(100 * time.Millisecond) // 等待异步处理
    
    records, err := testEnv.DB.GetProcessingRecords(alert.ID)
    assert.NoError(t, err)
    assert.Len(t, records, 1)
    assert.Equal(t, ProcessingStatusProcessed, records[0].ProcessingStatus)
}
```

### 性能测试

```go
// 性能测试示例
func BenchmarkChannelManager_SendMessage(b *testing.B) {
    cm := setupChannelManager()
    channel := createTestChannel()
    message := &Message{
        Title:   "性能测试",
        Content: "这是一条性能测试消息",
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            err := cm.SendMessage(context.Background(), channel.ID, message)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

## 部署和运维

### Docker部署配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  alertagent:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
      - N8N_ENDPOINT=http://n8n:5678
      - DIFY_ENDPOINT=http://dify:5001
    depends_on:
      - mysql
      - redis
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  alertmanager-prod:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/prod:/etc/alertmanager
      - alertmanager-prod-data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.enable-lifecycle'

  config-syncer:
    image: alertagent/config-syncer:latest
    environment:
      - ALERTAGENT_ENDPOINT=http://alertagent:8080
      - CLUSTER_ID=prod-cluster
      - CONFIG_TYPE=alertmanager
      - CONFIG_PATH=/etc/alertmanager/alertmanager.yml
      - RELOAD_URL=http://alertmanager-prod:9093/-/reload
      - SYNC_INTERVAL=30s
    volumes:
      - ./alertmanager/prod:/etc/alertmanager
    depends_on:
      - alertagent
      - alertmanager-prod
```

### Kubernetes部署配置

```yaml
# k8s/alertagent-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertagent
  namespace: alertagent-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: alertagent
  template:
    metadata:
      labels:
        app: alertagent
    spec:
      containers:
      - name: alertagent
        image: alertagent:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "mysql-service"
        - name: REDIS_HOST
          value: "redis-service"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 监控和可观测性

```go
// Metrics 指标收集
type Metrics struct {
    // 告警处理指标
    AlertsReceived    prometheus.Counter
    AlertsProcessed   prometheus.Counter
    ProcessingLatency prometheus.Histogram
    
    // 渠道健康指标
    ChannelHealth     prometheus.GaugeVec
    MessagesSent      prometheus.CounterVec
    MessageLatency    prometheus.HistogramVec
    
    // 集群同步指标
    ConfigSyncs       prometheus.CounterVec
    SyncLatency       prometheus.HistogramVec
    SyncErrors        prometheus.CounterVec
}

func NewMetrics() *Metrics {
    return &Metrics{
        AlertsReceived: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "alertagent_alerts_received_total",
            Help: "Total number of alerts received",
        }),
        AlertsProcessed: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "alertagent_alerts_processed_total",
            Help: "Total number of alerts processed",
        }),
        ProcessingLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "alertagent_processing_duration_seconds",
            Help:    "Alert processing duration in seconds",
            Buckets: prometheus.DefBuckets,
        }),
    }
}
```

### 日志配置

```go
// Logger 日志配置
func InitLogger(level string) (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(parseLogLevel(level))
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    logger, err := config.Build()
    if err != nil {
        return nil, fmt.Errorf("failed to build logger: %w", err)
    }
    
    return logger, nil
}

// StructuredLogging 结构化日志示例
func (cm *ChannelManager) logChannelOperation(operation string, channelID string, err error) {
    fields := []zap.Field{
        zap.String("operation", operation),
        zap.String("channel_id", channelID),
        zap.String("component", "channel_manager"),
    }
    
    if err != nil {
        fields = append(fields, zap.Error(err))
        cm.logger.Error("Channel operation failed", fields...)
    } else {
        cm.logger.Info("Channel operation completed", fields...)
    }
}
```

这个设计文档涵盖了AlertAgent架构重新设计的所有核心组件、接口、数据模型、错误处理、测试策略和部署配置。设计遵循了微服务架构原则，支持高可用性、可扩展性和可维护性。