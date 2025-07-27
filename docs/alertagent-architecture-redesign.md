# AlertAgent 重新架构设计文档

**文档版本**: v1.0  
**创建时间**: 2024-12-19  
**作者**: SRE Team  
**状态**: 设计阶段  

## 1. 架构概述

### 1.1 重新定位

AlertAgent 从原有的独立告警系统重新定位为**智能告警管理和分发中心**，与 Alertmanager、n8n+Dify 形成完整的智能告警处理生态系统。

### 1.2 核心价值

- **告警规则管理中心**: 统一管理、版本控制、分发告警规则
- **智能告警网关**: 告警路由、收敛、抑制、智能分发
- **AI 驱动分析**: 集成 n8n+Dify 实现智能告警分析和处理
- **多集群管理**: 支持多 Alertmanager 集群的统一管理

## 2. 整体架构图

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
│  │ │ 规则管理器   │ │────│ │ 规则执行器   │ │    │ │ 告警分析器   │ │         │
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
│  │ • MySQL         │    │ • Prometheus    │    │ • 邮件           │         │
│  │ • Redis         │    │ • VictoriaMetrics│   │ • 钉钉           │         │
│  │ • InfluxDB      │    │ • Grafana       │    │ • 微信           │         │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 3. 核心组件设计

### 3.1 AlertAgent - 智能告警管理中心

#### 3.1.1 规则管理器 (Rule Manager)

**职责**:
- 告警规则的 CRUD 操作
- 规则版本控制和历史管理
- 规则语法验证和测试
- 规则模板管理

**核心功能**:
```go
type RuleManager struct {
    ruleRepo     repository.RuleRepository
    validator    RuleValidator
    versionCtrl  VersionController
    templateMgr  TemplateManager
}

// 规则分发到 Alertmanager
func (rm *RuleManager) DistributeRule(ruleID string, targets []string) error

// 规则同步状态检查
func (rm *RuleManager) CheckSyncStatus(ruleID string) (*SyncStatus, error)

// 批量规则操作
func (rm *RuleManager) BatchOperation(operation string, ruleIDs []string) error
```

#### 3.1.2 集群管理器 (Cluster Manager)

**职责**:
- 多 Alertmanager 集群注册和管理
- 集群健康状态监控
- 负载均衡和故障转移
- 配置同步管理

**核心功能**:
```go
type ClusterManager struct {
    clusters     map[string]*AlertmanagerCluster
    healthCheck  HealthChecker
    loadBalancer LoadBalancer
    configSync   ConfigSynchronizer
}

// 集群注册
func (cm *ClusterManager) RegisterCluster(config *ClusterConfig) error

// 集群健康检查
func (cm *ClusterManager) HealthCheck() map[string]*HealthStatus

// 配置分发
func (cm *ClusterManager) DistributeConfig(clusterID string, config *Config) error
```

#### 3.1.3 智能网关 (Smart Gateway)

**职责**:
- 告警接收和预处理
- 告警收敛和抑制
- 智能路由决策
- 告警优先级管理

**核心功能**:
```go
type SmartGateway struct {
    receiver     AlertReceiver
    processor    AlertProcessor
    router       SmartRouter
    suppressor   AlertSuppressor
}

// 告警接收处理
func (sg *SmartGateway) ReceiveAlert(alert *Alert) error

// 智能路由
func (sg *SmartGateway) RouteAlert(alert *Alert) (*RoutingDecision, error)

// 告警收敛
func (sg *SmartGateway) ConvergeAlerts(alerts []*Alert) []*Alert
```

### 3.2 Alertmanager 集群管理

#### 3.2.1 配置文件管理

**基于Sidecar的配置分发**（推荐）:

Alertmanager 不支持动态发现和远程API配置，AlertAgent 采用Sidecar模式实现优雅的配置分发：

1. **Sidecar配置拉取方案**（推荐）:
```yaml
# Alertmanager Pod with Sidecar
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  template:
    spec:
      containers:
      # 主容器：Alertmanager
      - name: alertmanager
        image: prom/alertmanager:latest
        ports:
        - containerPort: 9093
        args:
        - --config.file=/etc/alertmanager/alertmanager.yml
        - --storage.path=/alertmanager
        - --web.enable-lifecycle
        volumeMounts:
        - name: config
          mountPath: /etc/alertmanager
        - name: storage
          mountPath: /alertmanager
      
      # Sidecar容器：配置同步器
      - name: config-syncer
        image: alertagent/config-syncer:latest
        env:
        - name: ALERTAGENT_ENDPOINT
          value: "http://alertagent-service:8080"
        - name: CLUSTER_ID
          value: "prod-cluster"
        - name: SYNC_INTERVAL
          value: "30s"
        - name: CONFIG_PATH
          value: "/etc/alertmanager/alertmanager.yml"
        - name: ALERTMANAGER_RELOAD_URL
          value: "http://localhost:9093/-/reload"
        volumeMounts:
        - name: config
          mountPath: /etc/alertmanager
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      
      volumes:
      - name: config
        emptyDir: {}
```

2. **Sidecar配置同步器实现**:
```go
// config-syncer/main.go
package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "text/template"
)

type ConfigSyncer struct {
    AlertAgentEndpoint string
    ClusterID         string
    SyncInterval      time.Duration
    ConfigPath        string
    ReloadURL         string
    LastConfigHash    string
    Template          *template.Template
}

type AlertmanagerConfig struct {
    Global    GlobalConfig    `json:"global"`
    Route     RouteConfig     `json:"route"`
    Receivers []ReceiverConfig `json:"receivers"`
    InhibitRules []InhibitRule `json:"inhibit_rules"`
}

func (cs *ConfigSyncer) Start(ctx context.Context) error {
    ticker := time.NewTicker(cs.SyncInterval)
    defer ticker.Stop()
    
    // 初始同步
    if err := cs.syncConfig(); err != nil {
        return fmt.Errorf("initial sync failed: %v", err)
    }
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := cs.syncConfig(); err != nil {
                fmt.Printf("Config sync failed: %v\n", err)
            }
        }
    }
}

// 从AlertAgent拉取配置
func (cs *ConfigSyncer) fetchConfig() (*AlertmanagerConfig, error) {
    url := fmt.Sprintf("%s/api/v1/clusters/%s/alertmanager/config", 
        cs.AlertAgentEndpoint, cs.ClusterID)
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch config: %d", resp.StatusCode)
    }
    
    var config AlertmanagerConfig
    if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

// 渲染配置文件
func (cs *ConfigSyncer) renderConfig(config *AlertmanagerConfig) (string, error) {
    tmpl := `global:
  smtp_smarthost: '{{ .Global.SMTPHost }}:{{ .Global.SMTPPort }}'
  smtp_from: '{{ .Global.SMTPFrom }}'

route:
  group_by: {{ .Route.GroupBy }}
  group_wait: {{ .Route.GroupWait }}
  group_interval: {{ .Route.GroupInterval }}
  repeat_interval: {{ .Route.RepeatInterval }}
  receiver: '{{ .Route.Receiver }}'
  routes:
  {{- range .Route.Routes }}
  - match:
      severity: {{ .Severity }}
    receiver: {{ .Receiver }}
    continue: {{ .Continue }}
  {{- end }}

receivers:
{{- range .Receivers }}
- name: '{{ .Name }}'
  webhook_configs:
  {{- range .WebhookConfigs }}
  - url: '{{ .URL }}'
    send_resolved: {{ .SendResolved }}
  {{- end }}
{{- end }}

inhibit_rules:
{{- range .InhibitRules }}
- source_match:
    severity: '{{ .SourceMatch.Severity }}'
  target_match:
    severity: '{{ .TargetMatch.Severity }}'
  equal: {{ .Equal }}
{{- end }}`
    
    t, err := template.New("alertmanager").Parse(tmpl)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := t.Execute(&buf, config); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}

// 同步配置
func (cs *ConfigSyncer) syncConfig() error {
    // 1. 拉取最新配置
    config, err := cs.fetchConfig()
    if err != nil {
        return fmt.Errorf("fetch config failed: %v", err)
    }
    
    // 2. 渲染配置文件
    configContent, err := cs.renderConfig(config)
    if err != nil {
        return fmt.Errorf("render config failed: %v", err)
    }
    
    // 3. 检查配置是否变更
    hash := sha256.Sum256([]byte(configContent))
    configHash := hex.EncodeToString(hash[:])
    
    if configHash == cs.LastConfigHash {
        return nil // 配置未变更
    }
    
    // 4. 写入配置文件
    if err := cs.writeConfigFile(configContent); err != nil {
        return fmt.Errorf("write config failed: %v", err)
    }
    
    // 5. 触发Alertmanager重载
    if err := cs.reloadAlertmanager(); err != nil {
        return fmt.Errorf("reload alertmanager failed: %v", err)
    }
    
    cs.LastConfigHash = configHash
    fmt.Printf("Config synced successfully, hash: %s\n", configHash)
    return nil
}

// 写入配置文件
func (cs *ConfigSyncer) writeConfigFile(content string) error {
    // 原子性写入：先写临时文件，再重命名
    tmpFile := cs.ConfigPath + ".tmp"
    
    if err := ioutil.WriteFile(tmpFile, []byte(content), 0644); err != nil {
        return err
    }
    
    return os.Rename(tmpFile, cs.ConfigPath)
}

// 触发Alertmanager重载
func (cs *ConfigSyncer) reloadAlertmanager() error {
    resp, err := http.Post(cs.ReloadURL, "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("reload failed with status: %d", resp.StatusCode)
    }
    
    return nil
}

func main() {
    syncer := &ConfigSyncer{
        AlertAgentEndpoint: os.Getenv("ALERTAGENT_ENDPOINT"),
        ClusterID:         os.Getenv("CLUSTER_ID"),
        SyncInterval:      30 * time.Second,
        ConfigPath:        os.Getenv("CONFIG_PATH"),
        ReloadURL:         os.Getenv("ALERTMANAGER_RELOAD_URL"),
    }
    
    if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
        if d, err := time.ParseDuration(interval); err == nil {
            syncer.SyncInterval = d
        }
    }
    
    ctx := context.Background()
    if err := syncer.Start(ctx); err != nil {
        fmt.Printf("Config syncer failed: %v\n", err)
        os.Exit(1)
    }
}
```

3. **共享存储方案**（备选）:
```yaml
# 使用 Kubernetes ConfigMap 或共享存储
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: monitoring
data:
  alertmanager.yml: |
    global:
      smtp_smarthost: '{{ .SMTPHost }}:{{ .SMTPPort }}'
      smtp_from: '{{ .SMTPFrom }}'
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'alertagent-webhook'
      routes:
      {{- range .Routes }}
      - match:
          severity: {{ .Severity }}
        receiver: {{ .Receiver }}
        continue: {{ .Continue }}
      {{- end }}
    
    receivers:
    - name: 'alertagent-webhook'
      webhook_configs:
      - url: 'http://alertagent:8080/api/v1/alerts/webhook'
        send_resolved: true
```

2. **文件同步机制**:
```go
type ConfigDistributor struct {
    clusters     map[string]*ClusterConfig
    fileManager  FileManager
    syncManager  SyncManager
}

// 配置文件生成和分发
func (cd *ConfigDistributor) DistributeConfig(clusterID string, config *AlertmanagerConfig) error {
    // 1. 生成配置文件
    configContent, err := cd.generateConfig(config)
    if err != nil {
        return err
    }
    
    // 2. 写入共享存储或ConfigMap
    err = cd.writeConfigFile(clusterID, configContent)
    if err != nil {
        return err
    }
    
    // 3. 触发Alertmanager reload
    return cd.triggerReload(clusterID)
}

// 触发Alertmanager重新加载配置
func (cd *ConfigDistributor) triggerReload(clusterID string) error {
    cluster := cd.clusters[clusterID]
    reloadURL := fmt.Sprintf("%s/-/reload", cluster.Endpoint)
    
    resp, err := http.Post(reloadURL, "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("reload failed with status: %d", resp.StatusCode)
    }
    
    return nil
}
```

#### 3.2.2 告警规则分发管理

**基于Sidecar的规则分发策略**:

由于Alertmanager本身不执行告警规则（规则由Prometheus执行），AlertAgent需要将规则分发到对应的Prometheus实例。采用Sidecar模式实现优雅的规则分发：

1. **Prometheus Sidecar规则同步方案**:
```yaml
# Prometheus Pod with Rule Syncer Sidecar
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: prometheus
  namespace: monitoring
spec:
  template:
    spec:
      containers:
      # 主容器：Prometheus
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        args:
        - --config.file=/etc/prometheus/prometheus.yml
        - --storage.tsdb.path=/prometheus
        - --web.enable-lifecycle
        - --web.enable-admin-api
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: rules
          mountPath: /etc/prometheus/rules
        - name: storage
          mountPath: /prometheus
      
      # Sidecar容器：规则同步器
      - name: rule-syncer
        image: alertagent/rule-syncer:latest
        env:
        - name: ALERTAGENT_ENDPOINT
          value: "http://alertagent-service:8080"
        - name: CLUSTER_ID
          value: "prod-cluster"
        - name: SYNC_INTERVAL
          value: "30s"
        - name: RULES_PATH
          value: "/etc/prometheus/rules"
        - name: PROMETHEUS_RELOAD_URL
          value: "http://localhost:9090/-/reload"
        volumeMounts:
        - name: rules
          mountPath: /etc/prometheus/rules
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      
      volumes:
      - name: config
        configMap:
          name: prometheus-config
      - name: rules
        emptyDir: {}
```

2. **规则同步器实现**:
```go
// rule-syncer/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "gopkg.in/yaml.v2"
)

type RuleSyncer struct {
    AlertAgentEndpoint string
    ClusterID         string
    SyncInterval      time.Duration
    RulesPath         string
    ReloadURL         string
    LastRulesHash     string
}

type PrometheusRules struct {
    Groups []RuleGroup `json:"groups" yaml:"groups"`
}

type RuleGroup struct {
    Name     string `json:"name" yaml:"name"`
    Interval string `json:"interval,omitempty" yaml:"interval,omitempty"`
    Rules    []Rule `json:"rules" yaml:"rules"`
}

type Rule struct {
    Alert       string            `json:"alert,omitempty" yaml:"alert,omitempty"`
    Expr        string            `json:"expr" yaml:"expr"`
    For         string            `json:"for,omitempty" yaml:"for,omitempty"`
    Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

func (rs *RuleSyncer) Start(ctx context.Context) error {
    ticker := time.NewTicker(rs.SyncInterval)
    defer ticker.Stop()
    
    // 初始同步
    if err := rs.syncRules(); err != nil {
        return fmt.Errorf("initial sync failed: %v", err)
    }
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := rs.syncRules(); err != nil {
                fmt.Printf("Rules sync failed: %v\n", err)
            }
        }
    }
}

// 从AlertAgent拉取规则
func (rs *RuleSyncer) fetchRules() (*PrometheusRules, error) {
    url := fmt.Sprintf("%s/api/v1/clusters/%s/prometheus/rules", 
        rs.AlertAgentEndpoint, rs.ClusterID)
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch rules: %d", resp.StatusCode)
    }
    
    var rules PrometheusRules
    if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
        return nil, err
    }
    
    return &rules, nil
}

// 同步规则
func (rs *RuleSyncer) syncRules() error {
    // 1. 拉取最新规则
    rules, err := rs.fetchRules()
    if err != nil {
        return fmt.Errorf("fetch rules failed: %v", err)
    }
    
    // 2. 转换为YAML格式
    rulesYAML, err := yaml.Marshal(rules)
    if err != nil {
        return fmt.Errorf("marshal rules failed: %v", err)
    }
    
    // 3. 检查规则是否变更
    hash := sha256.Sum256(rulesYAML)
    rulesHash := hex.EncodeToString(hash[:])
    
    if rulesHash == rs.LastRulesHash {
        return nil // 规则未变更
    }
    
    // 4. 写入规则文件
    if err := rs.writeRulesFile(rulesYAML); err != nil {
        return fmt.Errorf("write rules failed: %v", err)
    }
    
    // 5. 触发Prometheus重载
    if err := rs.reloadPrometheus(); err != nil {
        return fmt.Errorf("reload prometheus failed: %v", err)
    }
    
    rs.LastRulesHash = rulesHash
    fmt.Printf("Rules synced successfully, hash: %s\n", rulesHash)
    return nil
}

// 写入规则文件
func (rs *RuleSyncer) writeRulesFile(content []byte) error {
    rulesFile := filepath.Join(rs.RulesPath, "alertagent-rules.yml")
    tmpFile := rulesFile + ".tmp"
    
    // 确保目录存在
    if err := os.MkdirAll(rs.RulesPath, 0755); err != nil {
        return err
    }
    
    // 原子性写入
    if err := ioutil.WriteFile(tmpFile, content, 0644); err != nil {
        return err
    }
    
    return os.Rename(tmpFile, rulesFile)
}

// 触发Prometheus重载
func (rs *RuleSyncer) reloadPrometheus() error {
    resp, err := http.Post(rs.ReloadURL, "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("reload failed with status: %d", resp.StatusCode)
    }
    
    return nil
}

func main() {
    syncer := &RuleSyncer{
        AlertAgentEndpoint: os.Getenv("ALERTAGENT_ENDPOINT"),
        ClusterID:         os.Getenv("CLUSTER_ID"),
        SyncInterval:      30 * time.Second,
        RulesPath:         os.Getenv("RULES_PATH"),
        ReloadURL:         os.Getenv("PROMETHEUS_RELOAD_URL"),
    }
    
    if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
        if d, err := time.ParseDuration(interval); err == nil {
            syncer.SyncInterval = d
        }
    }
    
    ctx := context.Background()
    if err := syncer.Start(ctx); err != nil {
        fmt.Printf("Rule syncer failed: %v\n", err)
        os.Exit(1)
    }
}
```

3. **Prometheus规则文件生成**（备选方案）:
```yaml
# prometheus-rules.yml
groups:
{{- range .RuleGroups }}
- name: {{ .Name }}
  interval: {{ .Interval }}
  rules:
  {{- range .Rules }}
  - alert: {{ .AlertName }}
    expr: {{ .Expression }}
    for: {{ .Duration }}
    labels:
      severity: {{ .Severity }}
      team: {{ .Team }}
      cluster: {{ .Cluster }}
    annotations:
      summary: {{ .Summary }}
      description: {{ .Description }}
  {{- end }}
{{- end }}
```

2. **规则分发实现**:
```go
type RuleDistributor struct {
    prometheusClients map[string]*PrometheusClient
    configManager     ConfigManager
    ruleValidator     RuleValidator
}

// 分发规则到Prometheus集群
func (rd *RuleDistributor) DistributeRules(rules []*AlertRule, targetClusters []string) error {
    for _, cluster := range targetClusters {
        // 1. 生成规则文件内容
        ruleContent, err := rd.generateRuleFile(rules, cluster)
        if err != nil {
            return err
        }
        
        // 2. 验证规则语法
        if err := rd.ruleValidator.Validate(ruleContent); err != nil {
            return fmt.Errorf("rule validation failed for cluster %s: %v", cluster, err)
        }
        
        // 3. 写入规则文件（ConfigMap或共享存储）
        if err := rd.writeRuleFile(cluster, ruleContent); err != nil {
            return err
        }
        
        // 4. 触发Prometheus重新加载规则
        if err := rd.reloadPrometheusRules(cluster); err != nil {
            return err
        }
    }
    
    return nil
}

// 触发Prometheus重新加载规则
func (rd *RuleDistributor) reloadPrometheusRules(cluster string) error {
    client := rd.prometheusClients[cluster]
    reloadURL := fmt.Sprintf("%s/-/reload", client.Endpoint)
    
    resp, err := http.Post(reloadURL, "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("prometheus reload failed with status: %d", resp.StatusCode)
    }
    
    return nil
}
```

3. **Kubernetes部署配置**:
```yaml
# prometheus-rules-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: monitoring
  labels:
    app: prometheus
data:
  alertagent-rules.yml: |
    groups:
    - name: alertagent-generated
      rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
          team: platform
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is {{ $value }}%"
```

### 3.3 n8n + Dify 智能分析层

#### 3.3.1 告警分析工作流

**n8n 工作流设计**:
```json
{
  "nodes": [
    {
      "name": "Alert Webhook",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "alert-analysis",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Dify Analysis",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "http://dify:5001/v1/chat-messages",
        "method": "POST",
        "body": {
          "inputs": {
            "alert_data": "={{ $json.alert }}",
            "context": "={{ $json.context }}"
          },
          "query": "分析这个告警的根因和建议处理方案",
          "user": "alertagent"
        }
      }
    },
    {
      "name": "Decision Engine",
      "type": "n8n-nodes-base.function",
      "parameters": {
        "functionCode": "// 决策逻辑\nconst analysis = $input.first().json;\nconst severity = analysis.severity_score;\nconst confidence = analysis.confidence;\n\nif (severity > 0.8 && confidence > 0.9) {\n  return [{ action: 'auto_resolve', method: analysis.suggested_action }];\n} else {\n  return [{ action: 'escalate', reason: analysis.analysis }];\n}"
      }
    }
  ]
}
```

#### 3.3.2 Dify Agent 配置

**告警分析 Agent**:
```yaml
name: "Alert Analysis Agent"
description: "智能告警分析和处理建议"
model:
  provider: "openai"
  name: "gpt-4"
  parameters:
    temperature: 0.1
    max_tokens: 2000

prompt_template: |
  你是一个专业的SRE工程师，请分析以下告警信息：
  
  告警数据：
  {{ alert_data }}
  
  
  上下文信息：
  {{ context }}
  
  请提供：
  1. 根因分析
  2. 影响评估
  3. 处理建议
  4. 严重程度评分 (0-1)
  5. 置信度评分 (0-1)
  
  请以JSON格式返回分析结果。

knowledge_base:
  - name: "运维知识库"
    description: "包含常见故障处理方案"
  - name: "历史告警库"
    description: "历史告警处理记录"
```

## 4. 数据流设计

### 4.1 规则分发流程

**基于Sidecar模式的配置分发机制**:

```mermaid
sequenceDiagram
    participant U as User/API
    participant AA as AlertAgent
    participant PS as Prometheus-Sidecar
    participant Prom as Prometheus
    participant AS as Alertmanager-Sidecar
    participant AM as Alertmanager
    
    U->>AA: 创建/更新告警规则
    AA->>AA: 规则验证和模板渲染
    AA->>AA: 暴露配置接口
    
    Note over PS,Prom: Prometheus规则分发
    PS->>AA: 定期拉取规则配置
    AA->>PS: 返回最新规则配置
    PS->>PS: 渲染规则文件
    PS->>Prom: 写入共享卷规则文件
    PS->>Prom: 调用/-/reload接口
    Prom->>Prom: 重新加载规则文件
    Prom->>Prom: 开始执行新规则
    
    Note over AS,AM: Alertmanager配置分发
    AS->>AA: 定期拉取配置
    AA->>AS: 返回最新配置
    AS->>AS: 渲染配置文件
    AS->>AM: 写入共享卷配置文件
    AS->>AM: 调用/-/reload接口
    AM->>AM: 重新加载配置文件
    
    Note over Prom,AA: 告警触发流程
    Prom->>AM: 规则触发，发送告警
    AM->>AA: Webhook回调告警信息
    AA->>AA: 告警预处理和智能路由
    AA->>n8n: 发送到智能分析工作流
```

#### 4.1.1 Sidecar模式配置分发实现

**1. AlertAgent统一配置接口**:
```go
type ConfigAPI struct {
    ruleManager    *RuleManager
    configManager  *ConfigManager
    renderer       *ConfigRenderer
}

// 统一配置接口 - 直接返回渲染好的配置
func (ca *ConfigAPI) GetConfig(c *gin.Context) {
    clusterID := c.Query("cluster_id")
    configType := c.Query("type") // "prometheus-rules" or "alertmanager-config"
    
    if clusterID == "" || configType == "" {
        c.JSON(400, gin.H{"error": "cluster_id and type are required"})
        return
    }
    
    var renderedConfig string
    var err error
    
    switch configType {
    case "prometheus-rules":
        rules, err := ca.ruleManager.GetRulesByCluster(clusterID)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        renderedConfig, err = ca.renderer.RenderPrometheusRules(rules)
        
    case "alertmanager-config":
        config, err := ca.configManager.GetAlertmanagerConfig(clusterID)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        renderedConfig, err = ca.renderer.RenderAlertmanagerConfig(config)
        
    default:
        c.JSON(400, gin.H{"error": "unsupported config type"})
        return
    }
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 设置响应头并返回渲染好的配置
    c.Header("Content-Type", "application/yaml")
    c.Header("X-Config-Hash", ca.calculateHash(renderedConfig))
    c.String(200, renderedConfig)
}

// 配置渲染器
type ConfigRenderer struct {
    prometheusTemplate    *template.Template
    alertmanagerTemplate  *template.Template
}

func (cr *ConfigRenderer) RenderPrometheusRules(rules []*Rule) (string, error) {
    tmpl := `groups:
{{- range .Groups }}
- name: {{ .Name }}
  interval: {{ .Interval }}
  rules:
  {{- range .Rules }}
  - alert: {{ .AlertName }}
    expr: {{ .Expression }}
    for: {{ .Duration }}
    labels:
      severity: {{ .Severity }}
      team: {{ .Team }}
      cluster: {{ .Cluster }}
    annotations:
      summary: {{ .Summary }}
      description: {{ .Description }}
  {{- end }}
{{- end }}`
    
    t, err := template.New("prometheus-rules").Parse(tmpl)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := t.Execute(&buf, map[string]interface{}{"Groups": rules}); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}

func (cr *ConfigRenderer) RenderAlertmanagerConfig(config *AlertmanagerConfig) (string, error) {
    tmpl := `global:
  smtp_smarthost: '{{ .Global.SMTPHost }}:{{ .Global.SMTPPort }}'
  smtp_from: '{{ .Global.SMTPFrom }}'

route:
  group_by: {{ .Route.GroupBy }}
  group_wait: {{ .Route.GroupWait }}
  group_interval: {{ .Route.GroupInterval }}
  repeat_interval: {{ .Route.RepeatInterval }}
  receiver: '{{ .Route.Receiver }}'
  routes:
  {{- range .Route.Routes }}
  - match:
      severity: {{ .Severity }}
    receiver: {{ .Receiver }}
    continue: {{ .Continue }}
  {{- end }}

receivers:
{{- range .Receivers }}
- name: '{{ .Name }}'
  webhook_configs:
  {{- range .WebhookConfigs }}
  - url: '{{ .URL }}'
    send_resolved: {{ .SendResolved }}
  {{- end }}
{{- end }}

inhibit_rules:
{{- range .InhibitRules }}
- source_match:
    severity: '{{ .SourceMatch.Severity }}'
  target_match:
    severity: '{{ .TargetMatch.Severity }}'
  equal: {{ .Equal }}
{{- end }}`
    
    t, err := template.New("alertmanager-config").Parse(tmpl)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := t.Execute(&buf, config); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}

func (ca *ConfigAPI) calculateHash(content string) string {
    h := sha256.Sum256([]byte(content))
    return hex.EncodeToString(h[:])
}

// 注册路由
func (ca *ConfigAPI) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    {
        api.GET("/configs", ca.GetConfig)
    }
}
```

**2. Sidecar容器实现**:
```go
// Sidecar通用配置同步器
type ConfigSyncer struct {
    AlertAgentEndpoint string
    ClusterID         string
    SyncInterval      time.Duration
    ConfigPath        string
    ReloadURL         string
    ConfigType        string // "prometheus" or "alertmanager"
}

func (cs *ConfigSyncer) Start(ctx context.Context) error {
    ticker := time.NewTicker(cs.SyncInterval)
    defer ticker.Stop()
    
    // 启动时立即同步一次
    if err := cs.syncConfig(); err != nil {
        log.Printf("Initial config sync failed: %v", err)
    }
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Context cancelled, stopping config syncer")
            return ctx.Err()
        case <-ticker.C:
            if err := cs.syncConfig(); err != nil {
                log.Printf("Config sync failed: %v", err)
            }
        }
    }
}

func (cs *ConfigSyncer) syncConfig() error {
    // 1. 从AlertAgent拉取配置
    config, err := cs.fetchConfig()
    if err != nil {
        return fmt.Errorf("failed to fetch config: %w", err)
    }
    
    // 2. 检查配置是否有变化
    if !cs.hasConfigChanged(config) {
        logrus.Debug("Config unchanged, skipping sync")
        return nil
    }
    
    // 3. 写入配置文件
    if err := cs.writeConfigFile(config); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // 4. 触发重载
    if err := cs.triggerReload(); err != nil {
        return fmt.Errorf("failed to trigger reload: %w", err)
    }
    
    logrus.Infof("Successfully synced %s config", cs.ConfigType)
    return nil
}

// fetchConfig函数已合并到fetchConfigWithHash中，提供更完整的功能
```

**3. 优化的Sidecar配置同步逻辑**:
```go
// 添加字段存储上次的hash
type ConfigSyncer struct {
    AlertAgentEndpoint string
    ClusterID         string
    SyncInterval      time.Duration
    ConfigPath        string
    ReloadURL         string
    ConfigType        string // "prometheus" or "alertmanager"
    lastConfigHash    string // 存储上次配置的hash
    httpClient        *http.Client
}

func (cs *ConfigSyncer) syncConfig() error {
    // 1. 从AlertAgent拉取渲染好的配置
    config, serverHash, err := cs.fetchConfigWithHash()
    if err != nil {
        return fmt.Errorf("failed to fetch config: %w", err)
    }
    
    // 2. 对比hash，如果没有变化则跳过
     if serverHash != "" && serverHash == cs.lastConfigHash {
         log.Printf("%s config unchanged (server hash match), skipping sync", cs.ConfigType)
         return nil
     }
     
     // 3. 本地hash对比作为备用
     localHash := cs.calculateHash(config)
     if serverHash == "" {
         serverHash = localHash
     }
     
     if serverHash == cs.lastConfigHash {
         log.Printf("%s config unchanged (local hash match), skipping sync", cs.ConfigType)
         return nil
     }
    
    // 4. 写入配置文件
    if err := cs.writeConfigFile(config); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // 5. 触发热加载
    if err := cs.triggerReload(); err != nil {
        return fmt.Errorf("failed to trigger reload: %w", err)
    }
    
    // 6. 更新hash
     cs.lastConfigHash = serverHash
     log.Printf("Successfully synced %s config (hash: %s)", cs.ConfigType, serverHash)
     return nil
}

func (cs *ConfigSyncer) fetchConfigWithHash() ([]byte, string, error) {
    // 使用统一的配置接口
    var configType string
    switch cs.ConfigType {
    case "prometheus":
        configType = "prometheus-rules"
    case "alertmanager":
        configType = "alertmanager-config"
    default:
        return nil, "", fmt.Errorf("unknown config type: %s", cs.ConfigType)
    }
    
    endpoint := fmt.Sprintf("%s/api/v1/configs?cluster_id=%s&type=%s", 
        cs.AlertAgentEndpoint, cs.ClusterID, configType)
    
    resp, err := cs.httpClient.Get(endpoint)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("API returned status: %s", resp.Status)
    }
    
    config, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, "", err
    }
    
    // 获取服务端计算的hash
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
    if err := ioutil.WriteFile(tmpFile, config, 0644); err != nil {
        return err
    }
    
    return os.Rename(tmpFile, cs.ConfigPath)
}

func (cs *ConfigSyncer) triggerReload() error {
    resp, err := cs.httpClient.Post(cs.ReloadURL, "application/json", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("reload failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    log.Printf("%s reload triggered successfully", cs.ConfigType)
     return nil
}

func (cs *ConfigSyncer) calculateHash(data []byte) string {
    h := sha256.Sum256(data)
    return hex.EncodeToString(h[:])
}

// 初始化函数
func NewConfigSyncer(alertAgentEndpoint, clusterID, configType, configPath, reloadURL string, syncInterval time.Duration) *ConfigSyncer {
    return &ConfigSyncer{
        AlertAgentEndpoint: alertAgentEndpoint,
        ClusterID:         clusterID,
        ConfigType:        configType,
        ConfigPath:        configPath,
        ReloadURL:         reloadURL,
        SyncInterval:      syncInterval,
        httpClient:        &http.Client{Timeout: 30 * time.Second},
    }
}
```

**4. Sidecar主函数实现**:
```go
// main.go - 统一的Sidecar主程序
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // 从环境变量获取配置
    alertAgentEndpoint := os.Getenv("ALERTAGENT_ENDPOINT")
    clusterID := os.Getenv("CLUSTER_ID")
    configType := os.Getenv("CONFIG_TYPE") // "prometheus" or "alertmanager"
    configPath := os.Getenv("CONFIG_PATH")
    reloadURL := os.Getenv("RELOAD_URL")
    
    if alertAgentEndpoint == "" || clusterID == "" || configType == "" || configPath == "" || reloadURL == "" {
        log.Fatal("Missing required environment variables")
    }
    
    // 解析同步间隔
    syncInterval := 30 * time.Second
    if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
        if d, err := time.ParseDuration(interval); err == nil {
            syncInterval = d
        }
    }
    
    // 创建配置同步器
    syncer := NewConfigSyncer(alertAgentEndpoint, clusterID, configType, configPath, reloadURL, syncInterval)
    
    // 设置信号处理
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Received shutdown signal, stopping syncer...")
        cancel()
    }()
    
    log.Printf("Starting %s config syncer for cluster %s", configType, clusterID)
    log.Printf("Sync interval: %v", syncInterval)
    log.Printf("Config path: %s", configPath)
    log.Printf("Reload URL: %s", reloadURL)
    
    // 启动同步器
    if err := syncer.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Syncer failed: %v", err)
    }
    
    log.Println("Config syncer stopped")
}
```

**5. Kubernetes部署配置**:
```yaml
# prometheus-with-sidecar.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: prometheus
  labels:
    app: prometheus
spec:
  serviceName: prometheus
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      # Prometheus主容器
      - name: prometheus
        image: prom/prometheus:v2.45.0
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.path=/prometheus'
          - '--web.console.libraries=/etc/prometheus/console_libraries'
          - '--web.console.templates=/etc/prometheus/consoles'
          - '--web.enable-lifecycle'
          - '--web.enable-admin-api'
        ports:
          - name: web
            containerPort: 9090
        volumeMounts:
          - name: config-volume
            mountPath: /etc/prometheus
          - name: rules-volume
            mountPath: /etc/prometheus/rules
          - name: data-volume
            mountPath: /prometheus
      
      # 统一的配置同步Sidecar
      - name: config-syncer
        image: alertagent/config-syncer:latest
        env:
          - name: ALERTAGENT_ENDPOINT
            value: "http://alertagent-service.default.svc.cluster.local"
          - name: CLUSTER_ID
            value: "production-cluster"
          - name: CONFIG_TYPE
            value: "prometheus"
          - name: CONFIG_PATH
            value: "/etc/prometheus/rules/alertagent-rules.yml"
          - name: RELOAD_URL
            value: "http://localhost:9090/-/reload"
          - name: SYNC_INTERVAL
            value: "30s"
        volumeMounts:
          - name: rules-volume
            mountPath: /etc/prometheus/rules
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      
      volumes:
        - name: config-volume
          configMap:
            name: prometheus-config
        - name: rules-volume
          emptyDir: {}
        - name: data-volume
          persistentVolumeClaim:
            claimName: prometheus-data

---
# alertmanager-with-sidecar.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: alertmanager
  labels:
    app: alertmanager
spec:
  serviceName: alertmanager
  replicas: 1
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      # Alertmanager主容器
      - name: alertmanager
        image: prom/alertmanager:v0.27.0
        args:
          - '--config.file=/etc/alertmanager/alertmanager.yml'
          - '--storage.path=/alertmanager'
          - '--web.enable-lifecycle'
        ports:
          - name: web
            containerPort: 9093
        volumeMounts:
          - name: config-volume
            mountPath: /etc/alertmanager
          - name: data-volume
            mountPath: /alertmanager
      
      # 统一的配置同步Sidecar
      - name: config-syncer
        image: alertagent/config-syncer:latest
        env:
          - name: ALERTAGENT_ENDPOINT
            value: "http://alertagent-service.default.svc.cluster.local"
          - name: CLUSTER_ID
            value: "production-cluster"
          - name: CONFIG_TYPE
            value: "alertmanager"
          - name: CONFIG_PATH
            value: "/etc/alertmanager/alertmanager.yml"
          - name: RELOAD_URL
            value: "http://localhost:9093/-/reload"
          - name: SYNC_INTERVAL
            value: "30s"
        volumeMounts:
          - name: config-volume
            mountPath: /etc/alertmanager
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      
      volumes:
        - name: config-volume
          emptyDir: {}
        - name: data-volume
          persistentVolumeClaim:
            claimName: alertmanager-data
```

### 4.2 告警处理流程

```mermaid
sequenceDiagram
    participant AM as Alertmanager
    participant AA as AlertAgent
    participant N8N as n8n
    participant Dify as Dify
    participant NC as NotificationChannel
    
    AM->>AA: Webhook 告警
    AA->>AA: 告警收敛和抑制
    AA->>N8N: 触发分析工作流
    N8N->>Dify: 调用分析 Agent
    Dify->>Dify: AI 分析处理
    Dify-->>N8N: 返回分析结果
    N8N->>N8N: 决策引擎处理
    
    alt 自动处理
        N8N->>AA: 执行自动化操作
        AA->>AA: 记录处理结果
    else 人工介入
        N8N->>NC: 发送告警通知
        NC-->>N8N: 通知状态反馈
    end
    
    N8N->>AA: 更新告警状态
    AA->>AA: 记录处理历史
```

### 4.3 反馈学习流程

```mermaid
sequenceDiagram
    participant AA as AlertAgent
    participant Dify as Dify
    participant KB as KnowledgeBase
    participant ML as MachineLearning
    
    AA->>AA: 收集处理结果
    AA->>Dify: 发送反馈数据
    Dify->>KB: 更新知识库
    Dify->>ML: 训练优化模型
    ML-->>Dify: 返回优化结果
    Dify->>AA: 更新处理策略
    AA->>AA: 应用新策略
```

## 5. API 接口设计

### 5.1 AlertAgent API

#### 5.1.1 规则管理 API

```go
// 创建告警规则
POST /api/v1/rules
{
  "name": "high-cpu-usage",
  "expression": "cpu_usage > 80",
  "duration": "5m",
  "severity": "warning",
  "labels": {
    "team": "platform",
    "service": "web"
  },
  "annotations": {
    "summary": "High CPU usage detected",
    "description": "CPU usage is above 80% for 5 minutes"
  },
  "targets": ["cluster-1", "cluster-2"]
}

// 规则分发状态
GET /api/v1/rules/{id}/distribution
{
  "rule_id": "rule-123",
  "targets": [
    {
      "cluster_id": "cluster-1",
      "status": "synced",
      "last_sync": "2024-12-19T10:00:00Z",
      "version": "v1.2.3"
    }
  ]
}

// 批量操作
POST /api/v1/rules/batch
{
  "operation": "distribute",
  "rule_ids": ["rule-1", "rule-2"],
  "targets": ["cluster-1"]
}
```

#### 5.1.2 集群管理 API

```go
// 注册 Alertmanager 集群
POST /api/v1/clusters
{
  "name": "production-cluster",
  "endpoint": "http://alertmanager:9093",
  "config": {
    "sync_interval": "30s",
    "health_check_interval": "10s",
    "retry_attempts": 3
  },
  "labels": {
    "environment": "production",
    "region": "us-west-1"
  }
}

// 集群健康状态
GET /api/v1/clusters/health
{
  "clusters": [
    {
      "id": "cluster-1",
      "name": "production-cluster",
      "status": "healthy",
      "last_check": "2024-12-19T10:00:00Z",
      "metrics": {
        "response_time": "50ms",
        "success_rate": "99.9%"
      }
    }
  ]
}
```

#### 5.1.3 告警处理 API

```go
// 接收 Alertmanager Webhook
POST /api/v1/alerts/webhook
{
  "receiver": "alertagent",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighCPUUsage",
        "instance": "server-1",
        "severity": "warning"
      },
      "annotations": {
        "summary": "High CPU usage on server-1",
        "description": "CPU usage is 85%"
      },
      "startsAt": "2024-12-19T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=cpu_usage"
    }
  ]
}

// AI 分析接口
POST /api/v1/alerts/analyze
{
  "alert_id": "alert-123",
  "context": {
    "historical_data": true,
    "related_metrics": true,
    "similar_incidents": true
  }
}

// 响应
{
  "analysis_id": "analysis-456",
  "status": "processing",
  "estimated_completion": "2024-12-19T10:05:00Z"
}

// 获取分析结果
GET /api/v1/alerts/analyze/{analysis_id}
{
  "analysis_id": "analysis-456",
  "status": "completed",
  "result": {
    "root_cause": "Memory leak in application",
    "severity_score": 0.7,
    "confidence": 0.85,
    "suggested_actions": [
      "Restart application service",
      "Check memory usage patterns",
      "Review recent deployments"
    ],
    "estimated_resolution_time": "15m",
    "similar_incidents": [
      {
        "incident_id": "inc-789",
        "similarity": 0.92,
        "resolution": "Application restart"
      }
    ]
  }
}
```

### 5.2 n8n Webhook API

```go
// 告警处理工作流触发
POST /webhook/alert-processing
{
  "alert": {
    "id": "alert-123",
    "name": "HighCPUUsage",
    "severity": "warning",
    "labels": {...},
    "annotations": {...}
  },
  "context": {
    "cluster_id": "cluster-1",
    "environment": "production",
    "team": "platform"
  },
  "analysis_request": {
    "include_historical": true,
    "include_related_metrics": true,
    "analysis_depth": "detailed"
  }
}

// 自动化操作执行
POST /webhook/automation-action
{
  "action_type": "restart_service",
  "target": {
    "cluster": "cluster-1",
    "namespace": "default",
    "service": "web-app"
  },
  "parameters": {
    "graceful_shutdown": true,
    "wait_time": "30s"
  }
}
```

### 5.3 Dify API 集成

```go
// 告警分析请求
POST /v1/chat-messages
{
  "inputs": {
    "alert_data": {
      "name": "HighCPUUsage",
      "severity": "warning",
      "metrics": {...},
      "context": {...}
    },
    "historical_context": {
      "similar_alerts": [...],
      "recent_changes": [...]
    }
  },
  "query": "请分析这个告警的根因并提供处理建议",
  "user": "alertagent-system",
  "conversation_id": "conv-123"
}

// 知识库查询
POST /v1/datasets/{dataset_id}/documents/search
{
  "query": "CPU使用率过高的常见原因和解决方案",
  "top_k": 5,
  "score_threshold": 0.7
}
```

## 6. 配置文件设计

### 6.1 AlertAgent 配置

```yaml
# config/alertagent.yml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "production"

database:
  mysql:
    host: "mysql"
    port: 3306
    database: "alertagent"
    username: "alertagent"
    password: "password"
  redis:
    host: "redis"
    port: 6379
    database: 0

alertmanager:
  clusters:
    - name: "production"
      endpoint: "http://alertmanager-prod:9093"
      config_path: "/etc/alertmanager/config"
      rules_path: "/etc/prometheus/rules"
      sync_interval: "30s"
      health_check_interval: "10s"
    - name: "staging"
      endpoint: "http://alertmanager-staging:9093"
      config_path: "/etc/alertmanager/config"
      rules_path: "/etc/prometheus/rules"
      sync_interval: "60s"
      health_check_interval: "30s"

n8n:
  endpoint: "http://n8n:5678"
  webhook_base_url: "http://n8n:5678/webhook"
  workflows:
    alert_analysis: "alert-analysis"
    automation_action: "automation-action"
    notification: "notification"

dify:
  endpoint: "http://dify:5001"
  api_key: "dify-api-key"
  agents:
    alert_analyzer: "alert-analysis-agent"
    decision_maker: "decision-making-agent"
  knowledge_bases:
    - "operations-kb"
    - "historical-incidents-kb"

gateway:
  convergence:
    enabled: true
    window: "5m"
    max_alerts: 10
  suppression:
    enabled: true
    rules:
      - name: "maintenance-window"
        condition: "labels.maintenance == 'true'"
        duration: "2h"
  routing:
    default_workflow: "alert-analysis"
    rules:
      - match:
          severity: "critical"
        workflow: "critical-alert-handling"
      - match:
          team: "platform"
        workflow: "platform-alert-analysis"

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  endpoint: "/metrics"
  namespace: "alertagent"
```

### 6.2 Alertmanager 配置模板

```yaml
# templates/alertmanager.yml.tmpl
global:
  smtp_smarthost: '{{ .Global.SMTPHost }}:{{ .Global.SMTPPort }}'
  smtp_from: '{{ .Global.SMTPFrom }}'
  smtp_auth_username: '{{ .Global.SMTPUsername }}'
  smtp_auth_password: '{{ .Global.SMTPPassword }}'

templates:
  - '/etc/alertmanager/templates/*.tmpl'

route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: {{ .Route.GroupWait }}
  group_interval: {{ .Route.GroupInterval }}
  repeat_interval: {{ .Route.RepeatInterval }}
  receiver: 'alertagent-webhook'
  routes:
  {{- range .Route.Routes }}
  - match:
    {{- range $key, $value := .Match }}
      {{ $key }}: {{ $value }}
    {{- end }}
    receiver: {{ .Receiver }}
    {{- if .Continue }}
    continue: true
    {{- end }}
    {{- if .GroupWait }}
    group_wait: {{ .GroupWait }}
    {{- end }}
    {{- if .RepeatInterval }}
    repeat_interval: {{ .RepeatInterval }}
    {{- end }}
  {{- end }}

receivers:
- name: 'alertagent-webhook'
  webhook_configs:
  - url: '{{ .Receivers.AlertAgent.WebhookURL }}'
    send_resolved: true
    http_config:
      bearer_token: '{{ .Receivers.AlertAgent.Token }}'

{{- range .Receivers.Additional }}
- name: '{{ .Name }}'
  {{- if .EmailConfigs }}
  email_configs:
  {{- range .EmailConfigs }}
  - to: '{{ .To }}'
    subject: '{{ .Subject }}'
    body: '{{ .Body }}'
  {{- end }}
  {{- end }}
  {{- if .WebhookConfigs }}
  webhook_configs:
  {{- range .WebhookConfigs }}
  - url: '{{ .URL }}'
    send_resolved: {{ .SendResolved }}
  {{- end }}
  {{- end }}
{{- end }}

inhibit_rules:
{{- range .InhibitRules }}
- source_match:
  {{- range $key, $value := .SourceMatch }}
    {{ $key }}: {{ $value }}
  {{- end }}
  target_match:
  {{- range $key, $value := .TargetMatch }}
    {{ $key }}: {{ $value }}
  {{- end }}
  equal: {{ .Equal }}
{{- end }}
```

### 6.3 n8n 工作流配置

```json
{
  "name": "Alert Analysis Workflow",
  "nodes": [
    {
      "parameters": {
        "path": "alert-analysis",
        "options": {}
      },
      "name": "Alert Webhook",
      "type": "n8n-nodes-base.webhook",
      "typeVersion": 1,
      "position": [240, 300]
    },
    {
      "parameters": {
        "functionCode": "// 预处理告警数据\nconst alert = $input.first().json.alert;\nconst context = $input.first().json.context;\n\n// 提取关键信息\nconst processedData = {\n  alert_id: alert.id,\n  alert_name: alert.name,\n  severity: alert.severity,\n  labels: alert.labels,\n  annotations: alert.annotations,\n  context: context,\n  timestamp: new Date().toISOString()\n};\n\nreturn { json: processedData };"
      },
      "name": "Preprocess Alert",
      "type": "n8n-nodes-base.function",
      "typeVersion": 1,
      "position": [460, 300]
    },
    {
      "parameters": {
        "url": "http://dify:5001/v1/chat-messages",
        "authentication": "genericCredentialType",
        "genericAuthType": "httpHeaderAuth",
        "httpMethod": "POST",
        "sendBody": true,
        "bodyContentType": "json",
        "jsonBody": "={\n  \"inputs\": {\n    \"alert_data\": {{ JSON.stringify($json) }},\n    \"analysis_type\": \"root_cause_analysis\"\n  },\n  \"query\": \"请分析这个告警的根因并提供处理建议\",\n  \"user\": \"alertagent\",\n  \"conversation_id\": \"{{ $json.alert_id }}\"\n}"
      },
      "name": "Dify Analysis",
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 4,
      "position": [680, 300]
    },
    {
      "parameters": {
        "functionCode": "// 决策引擎\nconst analysis = $input.first().json;\nconst severity = parseFloat(analysis.severity_score || 0);\nconst confidence = parseFloat(analysis.confidence || 0);\n\nlet decision = {\n  action: 'manual_review',\n  reason: 'Default action',\n  priority: 'medium',\n  escalation_required: false\n};\n\n// 自动处理条件\nif (severity < 0.3 && confidence > 0.8) {\n  decision = {\n    action: 'auto_resolve',\n    method: analysis.suggested_actions[0],\n    priority: 'low',\n    escalation_required: false\n  };\n}\n// 立即升级条件\nelse if (severity > 0.8 || confidence < 0.5) {\n  decision = {\n    action: 'immediate_escalation',\n    reason: 'High severity or low confidence',\n    priority: 'critical',\n    escalation_required: true\n  };\n}\n// 智能路由\nelse {\n  decision = {\n    action: 'smart_routing',\n    target_team: analysis.recommended_team || 'platform',\n    priority: severity > 0.6 ? 'high' : 'medium',\n    escalation_required: severity > 0.7\n  };\n}\n\nreturn {\n  json: {\n    ...analysis,\n    decision: decision,\n    processed_at: new Date().toISOString()\n  }\n};"
      },
      "name": "Decision Engine",
      "type": "n8n-nodes-base.function",
      "typeVersion": 1,
      "position": [900, 300]
    },
    {
      "parameters": {
        "conditions": {
          "string": [
            {
              "value1": "={{ $json.decision.action }}",
              "value2": "auto_resolve"
            }
          ]
        }
      },
      "name": "Route Decision",
      "type": "n8n-nodes-base.if",
      "typeVersion": 1,
      "position": [1120, 300]
    },
    {
      "parameters": {
        "url": "http://alertagent:8080/api/v1/automation/execute",
        "httpMethod": "POST",
        "sendBody": true,
        "bodyContentType": "json",
        "jsonBody": "={\n  \"action_type\": \"{{ $json.decision.method }}\",\n  \"alert_id\": \"{{ $json.alert_id }}\",\n  \"parameters\": {{ JSON.stringify($json.decision.parameters || {}) }}\n}"
      },
      "name": "Execute Automation",
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 4,
      "position": [1340, 200]
    },
    {
      "parameters": {
        "url": "http://alertagent:8080/api/v1/alerts/escalate",
        "httpMethod": "POST",
        "sendBody": true,
        "bodyContentType": "json",
        "jsonBody": "={\n  \"alert_id\": \"{{ $json.alert_id }}\",\n  \"escalation_type\": \"{{ $json.decision.action }}\",\n  \"target_team\": \"{{ $json.decision.target_team }}\",\n  \"priority\": \"{{ $json.decision.priority }}\",\n  \"analysis\": {{ JSON.stringify($json) }}\n}"
      },
      "name": "Escalate Alert",
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 4,
      "position": [1340, 400]
    }
  ],
  "connections": {
    "Alert Webhook": {
      "main": [
        [
          {
            "node": "Preprocess Alert",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "Preprocess Alert": {
      "main": [
        [
          {
            "node": "Dify Analysis",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "Dify Analysis": {
      "main": [
        [
          {
            "node": "Decision Engine",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "Decision Engine": {
      "main": [
        [
          {
            "node": "Route Decision",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "Route Decision": {
      "main": [
        [
          {
            "node": "Execute Automation",
            "type": "main",
            "index": 0
          }
        ],
        [
          {
            "node": "Escalate Alert",
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  }
}
```

## 7. 数据库设计

### 7.1 核心表结构

```sql
-- Alertmanager 集群管理表
CREATE TABLE alertmanager_clusters (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    endpoint VARCHAR(255) NOT NULL,
    config_path VARCHAR(255),
    rules_path VARCHAR(255),
    sync_interval INT DEFAULT 30,
    health_check_interval INT DEFAULT 10,
    status ENUM('active', 'inactive', 'error') DEFAULT 'active',
    labels JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_status (status)
);

-- 规则分发记录表
CREATE TABLE rule_distributions (
    id VARCHAR(36) PRIMARY KEY,
    rule_id VARCHAR(36) NOT NULL,
    cluster_id VARCHAR(36) NOT NULL,
    version VARCHAR(50),
    status ENUM('pending', 'synced', 'failed') DEFAULT 'pending',
    sync_time TIMESTAMP NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (rule_id) REFERENCES rules(id) ON DELETE CASCADE,
    FOREIGN KEY (cluster_id) REFERENCES alertmanager_clusters(id) ON DELETE CASCADE,
    UNIQUE KEY uk_rule_cluster (rule_id, cluster_id),
    INDEX idx_status (status),
    INDEX idx_sync_time (sync_time)
);

-- 告警处理记录表
CREATE TABLE alert_processing_records (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(100) NOT NULL,
    alert_name VARCHAR(100),
    severity VARCHAR(20),
    cluster_id VARCHAR(36),
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    processing_status ENUM('received', 'analyzing', 'processed', 'failed') DEFAULT 'received',
    analysis_id VARCHAR(36),
    decision JSON,
    action_taken VARCHAR(100),
    resolution_time INT, -- 处理时间（秒）
    feedback_score DECIMAL(3,2), -- 处理效果评分
    labels JSON,
    annotations JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_alert_id (alert_id),
    INDEX idx_status (processing_status),
    INDEX idx_received_at (received_at),
    INDEX idx_cluster_id (cluster_id)
);

-- AI 分析记录表
CREATE TABLE ai_analysis_records (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(100) NOT NULL,
    analysis_type VARCHAR(50) DEFAULT 'root_cause_analysis',
    request_data JSON,
    response_data JSON,
    analysis_result JSON,
    confidence_score DECIMAL(3,2),
    processing_time INT, -- 分析耗时（毫秒）
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_alert_id (alert_id),
    INDEX idx_status (status),
    INDEX idx_analysis_type (analysis_type)
);

-- 自动化操作记录表
CREATE TABLE automation_actions (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(100) NOT NULL,
    action_type VARCHAR(100) NOT NULL,
    target_info JSON,
    parameters JSON,
    execution_status ENUM('pending', 'executing', 'completed', 'failed') DEFAULT 'pending',
    execution_result JSON,
    execution_time INT, -- 执行耗时（毫秒）
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_alert_id (alert_id),
    INDEX idx_action_type (action_type),
    INDEX idx_status (execution_status)
);

-- 告警收敛记录表
CREATE TABLE alert_convergence_records (
    id VARCHAR(36) PRIMARY KEY,
    convergence_key VARCHAR(255) NOT NULL, -- 收敛键（基于标签生成）
    alert_count INT DEFAULT 1,
    first_alert_time TIMESTAMP,
    last_alert_time TIMESTAMP,
    convergence_window INT, -- 收敛窗口（秒）
    status ENUM('active', 'expired', 'processed') DEFAULT 'active',
    representative_alert_id VARCHAR(100), -- 代表性告警ID
    converged_alerts JSON, -- 被收敛的告警列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_convergence_key (convergence_key),
    INDEX idx_status (status),
    INDEX idx_last_alert_time (last_alert_time)
);

-- 集群健康状态表
CREATE TABLE cluster_health_status (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    status ENUM('healthy', 'warning', 'critical', 'unknown') DEFAULT 'unknown',
    response_time INT, -- 响应时间（毫秒）
    success_rate DECIMAL(5,2), -- 成功率
    last_check_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    error_message TEXT,
    metrics JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES alertmanager_clusters(id) ON DELETE CASCADE,
    INDEX idx_cluster_id (cluster_id),
    INDEX idx_status (status),
    INDEX idx_last_check_time (last_check_time)
);
```

### 7.2 扩展现有表结构

```sql
-- 扩展 rules 表
ALTER TABLE rules ADD COLUMN target_clusters JSON AFTER annotations;
ALTER TABLE rules ADD COLUMN distribution_status ENUM('pending', 'partial', 'complete', 'failed') DEFAULT 'pending';
ALTER TABLE rules ADD COLUMN last_distribution_time TIMESTAMP NULL;
ALTER TABLE rules ADD COLUMN version VARCHAR(50) DEFAULT 'v1.0.0';

-- 扩展 alerts 表
ALTER TABLE alerts ADD COLUMN cluster_id VARCHAR(36);
ALTER TABLE alerts ADD COLUMN processing_record_id VARCHAR(36);
ALTER TABLE alerts ADD COLUMN ai_analysis_id VARCHAR(36);
ALTER TABLE alerts ADD COLUMN convergence_id VARCHAR(36);
ALTER TABLE alerts ADD COLUMN auto_resolved BOOLEAN DEFAULT FALSE;
ALTER TABLE alerts ADD COLUMN resolution_method VARCHAR(100);

-- 添加外键约束
ALTER TABLE alerts ADD FOREIGN KEY (cluster_id) REFERENCES alertmanager_clusters(id);
ALTER TABLE alerts ADD FOREIGN KEY (processing_record_id) REFERENCES alert_processing_records(id);
ALTER TABLE alerts ADD FOREIGN KEY (ai_analysis_id) REFERENCES ai_analysis_records(id);
ALTER TABLE alerts ADD FOREIGN KEY (convergence_id) REFERENCES alert_convergence_records(id);
```

## 8. 部署架构

### 8.1 Docker Compose 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  # AlertAgent 服务
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
    networks:
      - alertagent-network

  # Alertmanager 集群
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
      - '--web.external-url=http://localhost:9093'
      - '--cluster.listen-address=0.0.0.0:9094'
    networks:
      - alertagent-network

  alertmanager-staging:
    image: prom/alertmanager:latest
    ports:
      - "9094:9093"
    volumes:
      - ./alertmanager/staging:/etc/alertmanager
      - alertmanager-staging-data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9094'
    networks:
      - alertagent-network

  # Prometheus (用于规则执行)
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus:/etc/prometheus
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - alertagent-network

  # n8n 工作流引擎
  n8n:
    image: n8nio/n8n:latest
    ports:
      - "5678:5678"
    environment:
      - N8N_BASIC_AUTH_ACTIVE=true
      - N8N_BASIC_AUTH_USER=admin
      - N8N_BASIC_AUTH_PASSWORD=password
      - WEBHOOK_URL=http://localhost:5678
    volumes:
      - n8n-data:/home/node/.n8n
      - ./n8n/workflows:/home/node/.n8n/workflows
    networks:
      - alertagent-network

  # Dify AI 平台
  dify-api:
    image: langgenius/dify-api:latest
    ports:
      - "5001:5001"
    environment:
      - MODE=api
      - LOG_LEVEL=INFO
      - SECRET_KEY=sk-xxx
      - DATABASE_URL=postgresql://postgres:difyai123456@postgres:5432/dify
      - REDIS_URL=redis://:difyai123456@redis:6379/0
      - CELERY_BROKER_URL=redis://:difyai123456@redis:6379/1
    depends_on:
      - postgres
      - redis
    volumes:
      - dify-data:/app/storage
    networks:
      - alertagent-network

  dify-worker:
    image: langgenius/dify-api:latest
    environment:
      - MODE=worker
      - LOG_LEVEL=INFO
      - SECRET_KEY=sk-xxx
      - DATABASE_URL=postgresql://postgres:difyai123456@postgres:5432/dify
      - REDIS_URL=redis://:difyai123456@redis:6379/0
      - CELERY_BROKER_URL=redis://:difyai123456@redis:6379/1
    depends_on:
      - postgres
      - redis
    volumes:
      - dify-data:/app/storage
    networks:
      - alertagent-network

  # 数据库服务
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=rootpassword
      - MYSQL_DATABASE=alertagent
      - MYSQL_USER=alertagent
      - MYSQL_PASSWORD=password
    volumes:
      - mysql-data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    networks:
      - alertagent-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=difyai123456
      - POSTGRES_DB=dify
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - alertagent-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass difyai123456
    volumes:
      - redis-data:/data
    networks:
      - alertagent-network

  # 监控和可视化
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - alertagent-network

volumes:
  mysql-data:
  postgres-data:
  redis-data:
  prometheus-data:
  alertmanager-prod-data:
  alertmanager-staging-data:
  grafana-data:
  n8n-data:
  dify-data:

networks:
  alertagent-network:
    driver: bridge
```

### 8.2 Kubernetes 部署

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: alertagent-system
  labels:
    name: alertagent-system

---
# k8s/alertagent-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertagent
  namespace: alertagent-system
  labels:
    app: alertagent
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
        - name: N8N_ENDPOINT
          value: "http://n8n-service:5678"
        - name: DIFY_ENDPOINT
          value: "http://dify-api-service:5001"
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: logs
          mountPath: /app/logs
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
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: alertagent-config
      - name: logs
        emptyDir: {}

---
# k8s/alertagent-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: alertagent-service
  namespace: alertagent-system
  labels:
    app: alertagent
spec:
  selector:
    app: alertagent
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  type: ClusterIP

---
# k8s/alertagent-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: alertagent-ingress
  namespace: alertagent-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - alertagent.example.com
    secretName: alertagent-tls
  rules:
  - host: alertagent.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: alertagent-service
            port:
              number: 8080

---
# k8s/alertmanager-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: alertmanager
  namespace: alertagent-system
spec:
  serviceName: alertmanager-headless
  replicas: 3
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      - name: alertmanager
        image: prom/alertmanager:latest
        ports:
        - containerPort: 9093
        - containerPort: 9094
        args:
        - --config.file=/etc/alertmanager/alertmanager.yml
        - --storage.path=/alertmanager
        - --web.external-url=http://alertmanager.example.com
        - --cluster.listen-address=0.0.0.0:9094
        - --cluster.peer=alertmanager-0.alertmanager-headless:9094
        - --cluster.peer=alertmanager-1.alertmanager-headless:9094
        - --cluster.peer=alertmanager-2.alertmanager-headless:9094
        volumeMounts:
        - name: config
          mountPath: /etc/alertmanager
        - name: storage
          mountPath: /alertmanager
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
      volumes:
      - name: config
        configMap:
          name: alertmanager-config
  volumeClaimTemplates:
  - metadata:
      name: storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
# k8s/prometheus-rules-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: alertagent-system
  labels:
    app: prometheus
    component: rules
data:
  alertagent-rules.yml: |
    groups:
    - name: alertagent-generated
      interval: 30s
      rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
          team: platform
          cluster: "{{ $labels.cluster }}"
        annotations:
          summary: "High CPU usage detected on {{ $labels.instance }}"
          description: "CPU usage is {{ $value }}% for more than 5 minutes"
      - alert: HighMemoryUsage
        expr: memory_usage > 85
        for: 3m
        labels:
          severity: critical
          team: platform
          cluster: "{{ $labels.cluster }}"
        annotations:
          summary: "High memory usage detected on {{ $labels.instance }}"
          description: "Memory usage is {{ $value }}% for more than 3 minutes"

---
# k8s/alertmanager-config-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: alertagent-system
  labels:
    app: alertmanager
    component: config
data:
  alertmanager.yml: |
    global:
      smtp_smarthost: 'smtp.example.com:587'
      smtp_from: 'alerts@example.com'
      smtp_auth_username: 'alerts@example.com'
      smtp_auth_password: 'password'
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'alertagent-webhook'
      routes:
      - match:
          severity: critical
        receiver: 'critical-alerts'
        continue: true
      - match:
          severity: warning
        receiver: 'warning-alerts'
        continue: true
    
    receivers:
    - name: 'alertagent-webhook'
      webhook_configs:
      - url: 'http://alertagent-service:8080/api/v1/alerts/webhook'
        send_resolved: true
        http_config:
          bearer_token: 'your-webhook-token'
    - name: 'critical-alerts'
      webhook_configs:
      - url: 'http://alertagent-service:8080/api/v1/alerts/critical'
        send_resolved: true
    - name: 'warning-alerts'
      webhook_configs:
      - url: 'http://alertagent-service:8080/api/v1/alerts/warning'
        send_resolved: true

    inhibit_rules:
    - source_match:
        severity: 'critical'
      target_match:
        severity: 'warning'
      equal: ['alertname', 'cluster', 'service']
```

### 8.3 配置分发部署注意事项

#### 8.3.1 Kubernetes 环境配置分发

**1. RBAC 权限配置**:
```yaml
# k8s/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: alertagent
  namespace: alertagent-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: alertagent-configmap-manager
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: alertagent-configmap-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: alertagent-configmap-manager
subjects:
- kind: ServiceAccount
  name: alertagent
  namespace: alertagent-system
```

**2. 配置分发流程**:
- AlertAgent 通过 Kubernetes API 更新 ConfigMap
- Prometheus/Alertmanager 通过 Volume Mount 自动获取配置更新
- AlertAgent 调用 `/-/reload` 接口触发配置重载
- 监控配置分发状态和错误处理

**3. 部署要求**:
- AlertAgent 需要 ConfigMap 读写权限
- Prometheus/Alertmanager 需要配置 ConfigMap Volume Mount
- 网络策略允许 AlertAgent 访问 Prometheus/Alertmanager 的 reload 接口
- 配置备份和回滚机制

#### 8.3.2 Docker Compose 环境配置分发

**1. 共享存储配置**:
```yaml
# docker-compose.yml 配置分发部分
services:
  alertagent:
    volumes:
      - ./config/prometheus:/shared/prometheus
      - ./config/alertmanager:/shared/alertmanager
  
  prometheus:
    volumes:
      - ./config/prometheus:/etc/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--web.enable-lifecycle'
  
  alertmanager:
    volumes:
      - ./config/alertmanager:/etc/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--web.enable-lifecycle'
```

**2. 配置分发流程**:
- AlertAgent 生成配置文件到共享目录
- 通过 HTTP POST 调用 `/-/reload` 接口
- 监控文件变更和重载状态

**3. 注意事项**:
- 确保共享目录权限正确
- 配置文件语法验证
- 原子性写入（先写临时文件再重命名）
- 配置变更日志记录

#### 8.3.3 配置分发最佳实践

**1. 配置验证**:
```go
// 配置语法验证
func ValidatePrometheusRules(content string) error {
    // 使用 Prometheus 规则解析器验证
    _, err := promconfig.LoadFile(content)
    return err
}

func ValidateAlertmanagerConfig(content string) error {
    // 使用 Alertmanager 配置解析器验证
    _, err := amconfig.LoadFile(content)
    return err
}
```

**2. 配置备份和回滚**:
```go
type ConfigBackup struct {
    Timestamp   time.Time
    ConfigType  string // "prometheus" or "alertmanager"
    ClusterID   string
    Content     string
    Checksum    string
}

// 配置回滚
func (cd *ConfigDistributor) RollbackConfig(clusterID, configType string) error {
    backup := cd.getLastValidBackup(clusterID, configType)
    if backup == nil {
        return errors.New("no valid backup found")
    }
    
    return cd.applyConfig(clusterID, configType, backup.Content)
}
```

**3. 监控和告警**:
```yaml
# 配置分发监控规则
groups:
- name: alertagent-config-distribution
  rules:
  - alert: ConfigDistributionFailed
    expr: increase(alertagent_rule_distribution_total{status="failed"}[5m]) > 0
    for: 0m
    labels:
      severity: critical
    annotations:
      summary: "AlertAgent configuration distribution failed"
      description: "Failed to distribute configuration to cluster {{ $labels.cluster }}"
  
  - alert: ConfigReloadFailed
    expr: increase(alertagent_config_reload_total{status="failed"}[5m]) > 0
    for: 0m
    labels:
      severity: warning
    annotations:
      summary: "Configuration reload failed"
      description: "Failed to reload configuration on {{ $labels.target }}"
```

**4. 安全考虑**:
- 配置文件敏感信息加密
- API 访问认证和授权
- 网络访问控制
- 审计日志记录

## 9. 监控和运维

### 9.1 监控指标

```go
// metrics/alertagent_metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 规则分发指标
    RuleDistributionTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alertagent_rule_distribution_total",
            Help: "Total number of rule distributions",
        },
        []string{"cluster", "status"},
    )

    RuleDistributionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "alertagent_rule_distribution_duration_seconds",
            Help: "Time spent distributing rules",
            Buckets: prometheus.DefBuckets,
        },
        []string{"cluster"},
    )

    // 告警处理指标
    AlertProcessingTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alertagent_alert_processing_total",
            Help: "Total number of alerts processed",
        },
        []string{"severity", "status"},
    )

    AlertProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "alertagent_alert_processing_duration_seconds",
            Help: "Time spent processing alerts",
            Buckets: prometheus.DefBuckets,
        },
        []string{"severity"},
    )

    // AI 分析指标
    AIAnalysisTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alertagent_ai_analysis_total",
            Help: "Total number of AI analyses",
        },
        []string{"type", "status"},
    )

    AIAnalysisDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "alertagent_ai_analysis_duration_seconds",
            Help: "Time spent on AI analysis",
            Buckets: prometheus.DefBuckets,
        },
        []string{"type"},
    )

    // 集群健康指标
    ClusterHealthStatus = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "alertagent_cluster_health_status",
            Help: "Health status of Alertmanager clusters",
        },
        []string{"cluster", "status"},
    )

    ClusterResponseTime = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "alertagent_cluster_response_time_seconds",
            Help: "Response time of Alertmanager clusters",
        },
        []string{"cluster"},
    )

    // 告警收敛指标
    AlertConvergenceTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alertagent_alert_convergence_total",
            Help: "Total number of alert convergences",
        },
        []string{"convergence_key"},
    )

    ConvergedAlertsCount = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "alertagent_converged_alerts_count",
            Help: "Number of alerts in convergence groups",
        },
        []string{"convergence_key"},
    )
)

### 9.2 健康检查

```go
// health/health_checker.go
package health

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck interface {
    Check(ctx context.Context) error
    Name() string
}

type HealthStatus struct {
    Status string            `json:"status"`
    Checks map[string]string `json:"checks"`
    Timestamp time.Time      `json:"timestamp"`
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
    status := &HealthStatus{
        Status: "healthy",
        Checks: make(map[string]string),
        Timestamp: time.Now(),
    }

    for name, check := range hc.checks {
        if err := check.Check(ctx); err != nil {
            status.Checks[name] = fmt.Sprintf("unhealthy: %v", err)
            status.Status = "unhealthy"
        } else {
            status.Checks[name] = "healthy"
        }
    }

    return status
}

// 数据库健康检查
type DatabaseHealthCheck struct {
    db interface{ Ping() error }
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) error {
    return d.db.Ping()
}

func (d *DatabaseHealthCheck) Name() string {
    return "database"
}

// Alertmanager 集群健康检查
type AlertmanagerHealthCheck struct {
    endpoint string
    client   *http.Client
}

func (a *AlertmanagerHealthCheck) Check(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, "GET", a.endpoint+"/-/healthy", nil)
    if err != nil {
        return err
    }

    resp, err := a.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
    }

    return nil
}

func (a *AlertmanagerHealthCheck) Name() string {
    return "alertmanager"
}
```

### 9.3 日志配置

```yaml
# config/logging.yml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  
  # 结构化日志字段
  fields:
    service: "alertagent"
    version: "v1.0.0"
    environment: "production"
  
  # 日志轮转配置
  rotation:
    enabled: true
    max_size: "100MB"
    max_age: "7d"
    max_backups: 10
    compress: true
  
  # 不同组件的日志级别
  loggers:
    "alertagent.rule": "debug"
    "alertagent.gateway": "info"
    "alertagent.cluster": "warn"
    "alertagent.ai": "debug"
```

## 10. 实施路线图

### 10.1 第一阶段：基础架构搭建（4-6周）

**目标**: 建立 AlertAgent 与 Alertmanager 的基础集成

**主要任务**:
1. **重构 AlertAgent 核心架构**
   - 实现规则管理器
   - 开发集群管理器
   - 构建智能网关基础框架

2. **Alertmanager 集成**
   - 实现配置文件动态生成
   - 开发规则分发机制
   - 建立健康检查系统

3. **数据库扩展**
   - 新增集群管理表
   - 扩展规则分发记录
   - 实现数据迁移脚本

4. **基础 API 开发**
   - 规则管理 API
   - 集群管理 API
   - 健康检查 API

**交付物**:
- 可运行的 AlertAgent + Alertmanager 集成系统
- 基础的规则分发功能
- 集群管理界面
- 部署文档

### 10.2 第二阶段：智能分析集成（6-8周）

**目标**: 集成 n8n + Dify 实现智能告警分析

**主要任务**:
1. **n8n 工作流开发**
   - 设计告警分析工作流
   - 实现决策引擎逻辑
   - 开发自动化操作流程

2. **Dify AI Agent 配置**
   - 创建告警分析 Agent
   - 构建运维知识库
   - 训练决策模型

3. **智能网关增强**
   - 实现告警收敛算法
   - 开发智能路由逻辑
   - 集成 AI 分析接口

4. **反馈学习机制**
   - 收集处理结果数据
   - 实现模型优化循环
   - 建立效果评估体系

**交付物**:
- 完整的智能告警分析系统
- AI 驱动的决策引擎
- 自动化处理工作流
- 知识库管理系统

### 10.3 第三阶段：高级功能开发（4-6周）

**目标**: 实现高级告警管理功能

**主要任务**:
1. **告警收敛和抑制**
   - 多维度收敛策略
   - 动态抑制规则
   - 智能去重算法

2. **多渠道通知增强**
   - 渠道智能选择
   - 通知模板管理
   - 升级策略配置

3. **监控和可观测性**
   - 完善监控指标
   - 构建运维仪表板
   - 实现链路追踪

4. **性能优化**
   - 异步处理优化
   - 缓存策略改进
   - 数据库性能调优

**交付物**:
- 高性能告警处理系统
- 完善的监控体系
- 运维管理界面
- 性能测试报告

### 10.4 第四阶段：生产优化（2-4周）

**目标**: 生产环境部署和优化

**主要任务**:
1. **生产部署**
   - Kubernetes 部署配置
   - 高可用架构设计
   - 灾备方案实施

2. **安全加固**
   - 认证授权机制
   - 数据加密传输
   - 安全审计日志

3. **运维工具**
   - 自动化部署脚本
   - 备份恢复方案
   - 故障排查工具

4. **文档和培训**
   - 用户操作手册
   - 运维指南
   - 团队培训材料

**交付物**:
- 生产就绪的系统
- 完整的运维文档
- 安全合规报告
- 团队培训计划

## 11. 技术优势总结

### 11.1 架构优势

1. **职责分离**: AlertAgent 专注规则管理，Alertmanager 专注告警执行，各司其职
2. **技术解耦**: 通过标准接口集成，降低技术栈耦合度
3. **水平扩展**: 支持多集群管理，可根据需求灵活扩展
4. **高可用性**: 分布式架构设计，单点故障不影响整体服务

### 11.2 智能化优势

1. **AI 驱动**: 集成 Dify 实现智能分析和决策
2. **自动化处理**: 通过 n8n 实现复杂的自动化工作流
3. **持续学习**: 基于处理结果的反馈学习机制
4. **知识积累**: 构建运维知识库，提升处理效率

### 11.3 运维优势

1. **统一管理**: 集中管理多个 Alertmanager 集群
2. **可视化操作**: 提供友好的 Web 界面
3. **监控完善**: 全面的监控指标和健康检查
4. **易于维护**: 标准化的部署和运维流程

### 11.4 扩展性优势

1. **插件化设计**: 支持自定义分析插件和处理器
2. **开放接口**: 标准的 REST API 和 Webhook 接口
3. **多云支持**: 支持不同云平台和环境部署
4. **生态兼容**: 与现有监控生态系统无缝集成

---

**文档状态**: 设计完成，待实施  
**下一步**: 开始第一阶段开发工作  
**联系人**: SRE Team  
**更新时间**: 2024-12-19