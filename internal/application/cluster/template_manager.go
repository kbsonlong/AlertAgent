package cluster

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	clusterDomain "alert_agent/internal/domain/cluster"
	"go.uber.org/zap"
)

// TemplateManager 模板管理器
type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*clusterDomain.ConfigTemplate
	parsed    map[string]*template.Template
	logger    *zap.Logger
}

// NewTemplateManager 创建新的模板管理器
func NewTemplateManager(logger *zap.Logger) *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*clusterDomain.ConfigTemplate),
		parsed:    make(map[string]*template.Template),
		logger:    logger,
	}
}

// CreateConfigTemplate 创建配置模板
func (tm *TemplateManager) CreateConfigTemplate(ctx context.Context, configTemplate *clusterDomain.ConfigTemplate) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if configTemplate.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}
	
	if configTemplate.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	
	if configTemplate.Template == "" {
		return fmt.Errorf("template content cannot be empty")
	}
	
	// 解析模板
	parsedTemplate, err := tm.parseTemplate(configTemplate.ID, configTemplate.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	// 设置时间戳
	now := time.Now()
	if configTemplate.CreatedAt.IsZero() {
		configTemplate.CreatedAt = now
	}
	configTemplate.UpdatedAt = now
	
	// 存储模板
	tm.templates[configTemplate.ID] = configTemplate
	tm.parsed[configTemplate.ID] = parsedTemplate
	
	tm.logger.Info("Config template created", 
		zap.String("template_id", configTemplate.ID),
		zap.String("name", configTemplate.Name),
		zap.String("type", string(configTemplate.Type)))
	
	return nil
}

// UpdateConfigTemplate 更新配置模板
func (tm *TemplateManager) UpdateConfigTemplate(ctx context.Context, templateID string, configTemplate *clusterDomain.ConfigTemplate) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	existingTemplate, exists := tm.templates[templateID]
	if !exists {
		return fmt.Errorf("template not found: %s", templateID)
	}
	
	// 保留原有的创建时间
	configTemplate.ID = templateID
	configTemplate.CreatedAt = existingTemplate.CreatedAt
	configTemplate.UpdatedAt = time.Now()
	
	// 重新解析模板
	parsedTemplate, err := tm.parseTemplate(templateID, configTemplate.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	// 更新模板
	tm.templates[templateID] = configTemplate
	tm.parsed[templateID] = parsedTemplate
	
	tm.logger.Info("Config template updated", 
		zap.String("template_id", templateID),
		zap.String("name", configTemplate.Name))
	
	return nil
}

// DeleteConfigTemplate 删除配置模板
func (tm *TemplateManager) DeleteConfigTemplate(ctx context.Context, templateID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if _, exists := tm.templates[templateID]; !exists {
		return fmt.Errorf("template not found: %s", templateID)
	}
	
	delete(tm.templates, templateID)
	delete(tm.parsed, templateID)
	
	tm.logger.Info("Config template deleted", zap.String("template_id", templateID))
	return nil
}

// GetConfigTemplate 获取配置模板
func (tm *TemplateManager) GetConfigTemplate(ctx context.Context, templateID string) (*clusterDomain.ConfigTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	configTemplate, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	// 返回副本
	return &clusterDomain.ConfigTemplate{
		ID:          configTemplate.ID,
		Name:        configTemplate.Name,
		Description: configTemplate.Description,
		Type:        configTemplate.Type,
		Template:    configTemplate.Template,
		Variables:   configTemplate.Variables,
		Version:     configTemplate.Version,
		CreatedAt:   configTemplate.CreatedAt,
		UpdatedAt:   configTemplate.UpdatedAt,
	}, nil
}

// ListConfigTemplates 列出所有配置模板
func (tm *TemplateManager) ListConfigTemplates(ctx context.Context, clusterType clusterDomain.ClusterType) ([]*clusterDomain.ConfigTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	templates := make([]*clusterDomain.ConfigTemplate, 0)
	
	for _, configTemplate := range tm.templates {
		// 如果指定了集群类型，只返回匹配的模板
		if clusterType != "" && configTemplate.Type != clusterType {
			continue
		}
		
		templates = append(templates, &clusterDomain.ConfigTemplate{
			ID:          configTemplate.ID,
			Name:        configTemplate.Name,
			Description: configTemplate.Description,
			Type:        configTemplate.Type,
			Template:    configTemplate.Template,
			Variables:   configTemplate.Variables,
			Version:     configTemplate.Version,
			CreatedAt:   configTemplate.CreatedAt,
			UpdatedAt:   configTemplate.UpdatedAt,
		})
	}
	
	return templates, nil
}

// RenderConfig 渲染配置模板
func (tm *TemplateManager) RenderConfig(ctx context.Context, templateID string, variables map[string]interface{}) (string, error) {
	tm.mu.RLock()
	parsedTemplate, exists := tm.parsed[templateID]
	configTemplate := tm.templates[templateID]
	tm.mu.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateID)
	}
	
	// 合并默认变量和传入的变量
	allVariables := make(map[string]interface{})
	
	// 首先添加模板的默认变量
	for key, value := range configTemplate.Variables {
		allVariables[key] = value
	}
	
	// 然后添加传入的变量（会覆盖默认变量）
	for key, value := range variables {
		allVariables[key] = value
	}
	
	// 渲染模板
	var buf bytes.Buffer
	err := parsedTemplate.Execute(&buf, allVariables)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	
	renderedConfig := buf.String()
	
	tm.logger.Debug("Template rendered", 
		zap.String("template_id", templateID),
		zap.Int("variables_count", len(allVariables)),
		zap.Int("output_length", len(renderedConfig)))
	
	return renderedConfig, nil
}

// ValidateTemplate 验证模板语法
func (tm *TemplateManager) ValidateTemplate(ctx context.Context, templateContent string) error {
	_, err := tm.parseTemplate("validation", templateContent)
	return err
}

// ValidateVariables 验证模板变量
func (tm *TemplateManager) ValidateVariables(ctx context.Context, templateID string, variables map[string]interface{}) error {
	tm.mu.RLock()
	configTemplate, exists := tm.templates[templateID]
	tm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("template not found: %s", templateID)
	}
	
	// 检查必需的变量
	for key, defaultValue := range configTemplate.Variables {
		if _, provided := variables[key]; !provided {
			// 如果没有提供变量且没有默认值，则报错
			if defaultValue == nil {
				return fmt.Errorf("required variable missing: %s", key)
			}
		}
	}
	
	return nil
}

// GetTemplateVariables 获取模板变量定义
func (tm *TemplateManager) GetTemplateVariables(ctx context.Context, templateID string) (map[string]interface{}, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	configTemplate, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	// 返回变量副本
	variables := make(map[string]interface{})
	for key, value := range configTemplate.Variables {
		variables[key] = value
	}
	
	return variables, nil
}

// CloneTemplate 克隆模板
func (tm *TemplateManager) CloneTemplate(ctx context.Context, sourceTemplateID, newTemplateID, newName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	sourceTemplate, exists := tm.templates[sourceTemplateID]
	if !exists {
		return fmt.Errorf("source template not found: %s", sourceTemplateID)
	}
	
	if _, exists := tm.templates[newTemplateID]; exists {
		return fmt.Errorf("template with ID %s already exists", newTemplateID)
	}
	
	// 创建新模板
	newTemplate := &clusterDomain.ConfigTemplate{
		ID:          newTemplateID,
		Name:        newName,
		Description: fmt.Sprintf("Cloned from %s", sourceTemplate.Name),
		Type:        sourceTemplate.Type,
		Template:    sourceTemplate.Template,
		Variables:   make(map[string]interface{}),
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 复制变量
	for key, value := range sourceTemplate.Variables {
		newTemplate.Variables[key] = value
	}
	
	// 解析模板
	parsedTemplate, err := tm.parseTemplate(newTemplateID, newTemplate.Template)
	if err != nil {
		return fmt.Errorf("failed to parse cloned template: %w", err)
	}
	
	// 存储模板
	tm.templates[newTemplateID] = newTemplate
	tm.parsed[newTemplateID] = parsedTemplate
	
	tm.logger.Info("Template cloned", 
		zap.String("source_template_id", sourceTemplateID),
		zap.String("new_template_id", newTemplateID),
		zap.String("new_name", newName))
	
	return nil
}

// parseTemplate 解析模板
func (tm *TemplateManager) parseTemplate(name, templateContent string) (*template.Template, error) {
	// 创建模板并添加自定义函数
	tmpl := template.New(name).Funcs(template.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"trim":     strings.TrimSpace,
		"replace":  strings.ReplaceAll,
		"contains": strings.Contains,
		"join":     strings.Join,
		"split":    strings.Split,
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
		"now": time.Now,
		"formatTime": func(format string, t time.Time) string {
			return t.Format(format)
		},
	})
	
	// 解析模板内容
	parsedTemplate, err := tmpl.Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}
	
	return parsedTemplate, nil
}

// GetTemplateStats 获取模板统计信息
func (tm *TemplateManager) GetTemplateStats(ctx context.Context) map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_templates"] = len(tm.templates)
	
	// 按类型统计
	typeStats := make(map[string]int)
	for _, configTemplate := range tm.templates {
		typeStats[string(configTemplate.Type)]++
	}
	stats["by_type"] = typeStats
	
	return stats
}

// ExportTemplate 导出模板
func (tm *TemplateManager) ExportTemplate(ctx context.Context, templateID string) (map[string]interface{}, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	configTemplate, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	export := map[string]interface{}{
		"id":          configTemplate.ID,
		"name":        configTemplate.Name,
		"description": configTemplate.Description,
		"type":        configTemplate.Type,
		"template":    configTemplate.Template,
		"variables":   configTemplate.Variables,
		"version":     configTemplate.Version,
		"created_at":  configTemplate.CreatedAt,
		"updated_at":  configTemplate.UpdatedAt,
	}
	
	return export, nil
}

// ImportTemplate 导入模板
func (tm *TemplateManager) ImportTemplate(ctx context.Context, templateData map[string]interface{}) error {
	// 解析导入数据
	configTemplate := &clusterDomain.ConfigTemplate{}
	
	if id, ok := templateData["id"].(string); ok {
		configTemplate.ID = id
	} else {
		return fmt.Errorf("invalid or missing template ID")
	}
	
	if name, ok := templateData["name"].(string); ok {
		configTemplate.Name = name
	} else {
		return fmt.Errorf("invalid or missing template name")
	}
	
	if description, ok := templateData["description"].(string); ok {
		configTemplate.Description = description
	}
	
	if typeStr, ok := templateData["type"].(string); ok {
		configTemplate.Type = clusterDomain.ClusterType(typeStr)
	} else {
		return fmt.Errorf("invalid or missing template type")
	}
	
	if templateContent, ok := templateData["template"].(string); ok {
		configTemplate.Template = templateContent
	} else {
		return fmt.Errorf("invalid or missing template content")
	}
	
	if variables, ok := templateData["variables"].(map[string]interface{}); ok {
		configTemplate.Variables = variables
	} else {
		configTemplate.Variables = make(map[string]interface{})
	}
	
	if version, ok := templateData["version"].(string); ok {
		configTemplate.Version = version
	} else {
		configTemplate.Version = "1.0.0"
	}
	
	// 创建模板
	return tm.CreateConfigTemplate(ctx, configTemplate)
}