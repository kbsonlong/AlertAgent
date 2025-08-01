package database

import (
	"fmt"
	"strings"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	cfg := config.GetConfig().Database

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci&interpolateParams=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 设置会话字符集
	if err := DB.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return fmt.Errorf("failed to set character set: %w", err)
	}

	// 强制设置客户端字符集
	if err := DB.Exec("SET character_set_client = utf8mb4").Error; err != nil {
		return fmt.Errorf("failed to set character_set_client: %w", err)
	}
	if err := DB.Exec("SET character_set_connection = utf8mb4").Error; err != nil {
		return fmt.Errorf("failed to set character_set_connection: %w", err)
	}
	if err := DB.Exec("SET character_set_results = utf8mb4").Error; err != nil {
		return fmt.Errorf("failed to set character_set_results: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// 设置MySQL会话级别的优化参数
	optimizationQueries := []string{
		"SET SESSION sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO'",
		"SET SESSION innodb_lock_wait_timeout = 50",
		"SET SESSION lock_wait_timeout = 30",
		"SET SESSION autocommit = 1",
		"SET SESSION innodb_adaptive_hash_index = ON",
	}

	for _, query := range optimizationQueries {
		if err := DB.Exec(query).Error; err != nil {
			// 记录警告但不中断初始化
			fmt.Printf("Warning: Failed to execute optimization query: %s, error: %v\n", query, err)
		}
	}

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	return nil
}

// seedRules 插入默认告警规则
func seedRules() error {
	// 检查是否已存在默认告警规则
	var count int64
	if err := DB.Model(&model.Rule{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有规则，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建默认告警规则（使用新的Rule模型）
	defaultRules := []*model.Rule{
		{
			ID:         "rule-cpu-usage",
			Name:       "CPU使用率告警",
			Expression: "100 - (avg by (instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 90",
			Duration:   "5m",
			Severity:   "warning",
			Version:    "v1.0.0",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "rule-memory-usage",
			Name:       "内存使用率告警",
			Expression: "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 90",
			Duration:   "5m",
			Severity:   "warning",
			Version:    "v1.0.0",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "rule-disk-usage",
			Name:       "磁盘使用率告警",
			Expression: "(1 - (node_filesystem_avail_bytes{fstype!=\"tmpfs\"} / node_filesystem_size_bytes{fstype!=\"tmpfs\"})) * 100 > 90",
			Duration:   "5m",
			Severity:   "critical",
			Version:    "v1.0.0",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "rule-service-down",
			Name:       "服务可用性告警",
			Expression: "up == 0",
			Duration:   "1m",
			Severity:   "critical",
			Version:    "v1.0.0",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	// 设置默认标签和注释
	for _, rule := range defaultRules {
		// 设置默认标签
		labels := map[string]string{
			"team":        "ops",
			"environment": "production",
		}
		if err := rule.SetLabelsMap(labels); err != nil {
			return fmt.Errorf("failed to set labels for rule %s: %w", rule.Name, err)
		}

		// 设置默认注释
		annotations := map[string]string{
			"summary":     rule.Name,
			"description": "告警规则: " + rule.Name,
		}
		if err := rule.SetAnnotationsMap(annotations); err != nil {
			return fmt.Errorf("failed to set annotations for rule %s: %w", rule.Name, err)
		}

		// 设置默认目标
		targets := []string{"prometheus", "alertmanager"}
		if err := rule.SetTargetsList(targets); err != nil {
			return fmt.Errorf("failed to set targets for rule %s: %w", rule.Name, err)
		}

		// 创建规则
		if err := DB.Create(rule).Error; err != nil {
			return fmt.Errorf("failed to create default rule %s: %w", rule.Name, err)
		}
	}

	return nil
}

// seedAlerts 插入默认告警示例
// TODO: 需要更新Alert模型以支持新的Rule ID格式（string类型）
func seedAlerts() error {
	// 暂时注释掉，等Alert模型更新后再启用
	return nil
}

func autoMigrate() error {
	// 自动迁移数据库表
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Alert{},
		&model.Rule{},
		&model.RuleDistributionRecord{},
		&model.NotifyTemplate{},
		&model.NotifyGroup{},
		&model.NotifyRecord{},
		&model.Provider{},
		&model.Knowledge{},
		&model.Permission{},
		&model.Role{},
	); err != nil {
		return err
	}

	// 创建新的规则表索引
	if err := createRuleIndexes(); err != nil {
		return fmt.Errorf("failed to create rule indexes: %w", err)
	}

	// 插入初始化数据
	if err := seedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

// createRuleIndexes 创建规则表索引
func createRuleIndexes() error {
	// 为规则表创建必要的索引，使用MySQL兼容的语法
	indexes := map[string]string{
		"idx_alert_rules_name":       "CREATE INDEX idx_alert_rules_name ON alert_rules(name)",
		"idx_alert_rules_status":     "CREATE INDEX idx_alert_rules_status ON alert_rules(status)",
		"idx_alert_rules_created_at": "CREATE INDEX idx_alert_rules_created_at ON alert_rules(created_at)",
		"idx_alert_rules_severity":   "CREATE INDEX idx_alert_rules_severity ON alert_rules(severity)",
	}

	for indexName, indexSQL := range indexes {
		// 检查索引是否已存在
		var count int64
		checkSQL := `SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
					WHERE table_name = 'alert_rules' 
					AND table_schema = DATABASE() 
					AND index_name = ?`
		
		if err := DB.Raw(checkSQL, indexName).Scan(&count).Error; err != nil {
			return fmt.Errorf("failed to check index %s: %w", indexName, err)
		}
		
		// 如果索引不存在，则创建
		if count == 0 {
			if err := DB.Exec(indexSQL).Error; err != nil {
				// 忽略索引已存在的错误
				if !isIndexExistsError(err) {
					return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
				}
			}
		}
	}

	return nil
}

// isIndexExistsError 检查是否是索引已存在的错误
func isIndexExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate key name") || strings.Contains(errStr, "already exists")
}

// seedUsers 插入默认用户
func seedUsers() error {
	// 检查是否已存在用户
	var count int64
	if err := DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有用户，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建默认用户
	users := []model.User{
		{
			ID:       "admin-001",
			Username: "admin",
			Email:    "admin@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:     "admin",
			Status:   "active",
		},
		{
			ID:       "operator-001",
			Username: "operator",
			Email:    "operator@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:     "operator",
			Status:   "active",
		},
		{
			ID:       "viewer-001",
			Username: "viewer",
			Email:    "viewer@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:     "viewer",
			Status:   "inactive",
		},
		{
			ID:       "test-001",
			Username: "test_user",
			Email:    "test@example.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:     "viewer",
			Status:   "locked",
		},
	}

	for _, user := range users {
		if err := DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}

	return nil
}

// seedData 插入初始化数据
func seedData() error {
	// 检查并插入默认权限
	if err := seedPermissions(); err != nil {
		return err
	}

	// 检查并插入默认角色
	if err := seedRoles(); err != nil {
		return err
	}

	// 检查并插入默认用户
	if err := seedUsers(); err != nil {
		return err
	}

	// 检查并插入默认通知模板
	if err := seedNotifyTemplates(); err != nil {
		return err
	}

	// 检查并插入默认通知组
	if err := seedNotifyGroups(); err != nil {
		return err
	}

	// 检查并插入默认数据源
	if err := seedProviders(); err != nil {
		return err
	}

	// 检查并插入默认告警规则
	if err := seedRules(); err != nil {
		return err
	}

	// 初始化用户角色关联
	if err := seedUserRoles(); err != nil {
		return err
	}

	// TODO: 检查并插入默认告警示例 (需要更新Alert模型以支持新的Rule ID格式)
	// if err := seedAlerts(); err != nil {
	//     return err
	// }

	return nil
}

// seedNotifyTemplates 插入默认通知模板
func seedNotifyTemplates() error {
	// 检查是否已存在默认模板
	var count int64
	if err := DB.Model(&model.NotifyTemplate{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有模板，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建默认通知模板
	defaultTemplates := []model.NotifyTemplate{
		{
			Name:        "默认邮件模板",
			Type:        "email",
			Content:     "告警通知: {{.AlertName}}\n\n告警名称: {{.AlertName}}\n告警级别: {{.Level}}\n告警时间: {{.Time}}\n告警描述: {{.Description}}",
			Description: "系统默认邮件通知模板",
			Enabled:     true,
		},
		{
			Name:        "默认钉钉模板",
			Type:        "webhook",
			Content:     "## 告警通知\n\n**告警名称:** {{.AlertName}}\n\n**告警级别:** {{.Level}}\n\n**告警时间:** {{.Time}}\n\n**告警描述:** {{.Description}}",
			Description: "系统默认钉钉通知模板",
			Enabled:     true,
		},
	}

	for _, template := range defaultTemplates {
		if err := DB.Create(&template).Error; err != nil {
			return fmt.Errorf("failed to create default template %s: %w", template.Name, err)
		}
	}

	return nil
}

// seedNotifyGroups 插入默认通知组
func seedNotifyGroups() error {
	// 检查是否已存在默认通知组
	var count int64
	if err := DB.Model(&model.NotifyGroup{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有通知组，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建默认通知组
	defaultGroup := model.NotifyGroup{
		Name:        "默认通知组",
		Description: "系统默认通知组",
		Contacts:    "[]", // JSON数组格式
		Members:     "admin@example.com", // 可根据实际需求修改
		Channels:    "",
	}

	if err := DB.Create(&defaultGroup).Error; err != nil {
		return fmt.Errorf("failed to create default notify group: %w", err)
	}

	return nil
}

// seedProviders 插入默认数据源
func seedProviders() error {
	// 检查是否已存在默认数据源
	var count int64
	if err := DB.Model(&model.Provider{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有数据源，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建默认数据源示例（可根据实际需求修改）
	defaultProviders := []model.Provider{
		{
			Name:        "本地Prometheus",
			Type:        "prometheus",
			Endpoint:    "http://localhost:9090",
			Description: "本地Prometheus监控数据源",
		},
	}

	for _, provider := range defaultProviders {
		if err := DB.Create(&provider).Error; err != nil {
			return fmt.Errorf("failed to create default provider %s: %w", provider.Name, err)
		}
	}

	return nil
}

// seedPermissions 插入默认权限
func seedPermissions() error {
	// 检查是否已存在权限
	var count int64
	if err := DB.Model(&model.Permission{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有权限，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建系统默认权限
	defaultPermissions := []model.Permission{
		// 用户管理权限
		{Code: "user:create", Name: "创建用户", Description: "创建新用户的权限", Resource: "user", Action: "create", Category: "user", Type: "system", IsSystem: true},
		{Code: "user:read", Name: "查看用户", Description: "查看用户信息的权限", Resource: "user", Action: "read", Category: "user", Type: "system", IsSystem: true},
		{Code: "user:update", Name: "更新用户", Description: "更新用户信息的权限", Resource: "user", Action: "update", Category: "user", Type: "system", IsSystem: true},
		{Code: "user:delete", Name: "删除用户", Description: "删除用户的权限", Resource: "user", Action: "delete", Category: "user", Type: "system", IsSystem: true},
		{Code: "user:list", Name: "用户列表", Description: "查看用户列表的权限", Resource: "user", Action: "list", Category: "user", Type: "system", IsSystem: true},
		
		// 角色管理权限
		{Code: "role:create", Name: "创建角色", Description: "创建新角色的权限", Resource: "role", Action: "create", Category: "role", Type: "system", IsSystem: true},
		{Code: "role:read", Name: "查看角色", Description: "查看角色信息的权限", Resource: "role", Action: "read", Category: "role", Type: "system", IsSystem: true},
		{Code: "role:update", Name: "更新角色", Description: "更新角色信息的权限", Resource: "role", Action: "update", Category: "role", Type: "system", IsSystem: true},
		{Code: "role:delete", Name: "删除角色", Description: "删除角色的权限", Resource: "role", Action: "delete", Category: "role", Type: "system", IsSystem: true},
		{Code: "role:list", Name: "角色列表", Description: "查看角色列表的权限", Resource: "role", Action: "list", Category: "role", Type: "system", IsSystem: true},
		{Code: "role:assign", Name: "分配角色", Description: "为用户分配角色的权限", Resource: "role", Action: "manage", Category: "role", Type: "system", IsSystem: true},
		
		// 权限管理权限
		{Code: "permission:create", Name: "创建权限", Description: "创建新权限的权限", Resource: "permission", Action: "create", Category: "permission", Type: "system", IsSystem: true},
		{Code: "permission:read", Name: "查看权限", Description: "查看权限信息的权限", Resource: "permission", Action: "read", Category: "permission", Type: "system", IsSystem: true},
		{Code: "permission:update", Name: "更新权限", Description: "更新权限信息的权限", Resource: "permission", Action: "update", Category: "permission", Type: "system", IsSystem: true},
		{Code: "permission:delete", Name: "删除权限", Description: "删除权限的权限", Resource: "permission", Action: "delete", Category: "permission", Type: "system", IsSystem: true},
		{Code: "permission:list", Name: "权限列表", Description: "查看权限列表的权限", Resource: "permission", Action: "list", Category: "permission", Type: "system", IsSystem: true},
		
		// 告警管理权限
		{Code: "alert:create", Name: "创建告警", Description: "创建告警的权限", Resource: "alert", Action: "create", Category: "alert", Type: "system", IsSystem: true},
		{Code: "alert:read", Name: "查看告警", Description: "查看告警信息的权限", Resource: "alert", Action: "read", Category: "alert", Type: "system", IsSystem: true},
		{Code: "alert:update", Name: "更新告警", Description: "更新告警状态的权限", Resource: "alert", Action: "update", Category: "alert", Type: "system", IsSystem: true},
		{Code: "alert:delete", Name: "删除告警", Description: "删除告警的权限", Resource: "alert", Action: "delete", Category: "alert", Type: "system", IsSystem: true},
		{Code: "alert:list", Name: "告警列表", Description: "查看告警列表的权限", Resource: "alert", Action: "list", Category: "alert", Type: "system", IsSystem: true},
		
		// 规则管理权限
		{Code: "rule:create", Name: "创建规则", Description: "创建告警规则的权限", Resource: "rule", Action: "create", Category: "rule", Type: "system", IsSystem: true},
		{Code: "rule:read", Name: "查看规则", Description: "查看告警规则的权限", Resource: "rule", Action: "read", Category: "rule", Type: "system", IsSystem: true},
		{Code: "rule:update", Name: "更新规则", Description: "更新告警规则的权限", Resource: "rule", Action: "update", Category: "rule", Type: "system", IsSystem: true},
		{Code: "rule:delete", Name: "删除规则", Description: "删除告警规则的权限", Resource: "rule", Action: "delete", Category: "rule", Type: "system", IsSystem: true},
		{Code: "rule:list", Name: "规则列表", Description: "查看告警规则列表的权限", Resource: "rule", Action: "list", Category: "rule", Type: "system", IsSystem: true},
		
		// 数据源管理权限
		{Code: "provider:create", Name: "创建数据源", Description: "创建数据源的权限", Resource: "provider", Action: "create", Category: "provider", Type: "system", IsSystem: true},
		{Code: "provider:read", Name: "查看数据源", Description: "查看数据源信息的权限", Resource: "provider", Action: "read", Category: "provider", Type: "system", IsSystem: true},
		{Code: "provider:update", Name: "更新数据源", Description: "更新数据源配置的权限", Resource: "provider", Action: "update", Category: "provider", Type: "system", IsSystem: true},
		{Code: "provider:delete", Name: "删除数据源", Description: "删除数据源的权限", Resource: "provider", Action: "delete", Category: "provider", Type: "system", IsSystem: true},
		{Code: "provider:list", Name: "数据源列表", Description: "查看数据源列表的权限", Resource: "provider", Action: "list", Category: "provider", Type: "system", IsSystem: true},
		
		// 系统配置权限
		{Code: "config:read", Name: "查看配置", Description: "查看系统配置的权限", Resource: "config", Action: "read", Category: "config", Type: "system", IsSystem: true},
		{Code: "config:update", Name: "更新配置", Description: "更新系统配置的权限", Resource: "config", Action: "update", Category: "config", Type: "system", IsSystem: true},
		
		// 系统管理权限
		{Code: "system:manage", Name: "系统管理", Description: "系统管理的权限", Resource: "system", Action: "manage", Category: "system", Type: "system", IsSystem: true},
		{Code: "system:monitor", Name: "系统监控", Description: "系统监控的权限", Resource: "system", Action: "read", Category: "system", Type: "system", IsSystem: true},
	}

	for _, permission := range defaultPermissions {
		if err := DB.Create(&permission).Error; err != nil {
			return fmt.Errorf("failed to create default permission %s: %w", permission.Code, err)
		}
	}

	return nil
}

// seedRoles 插入默认角色
func seedRoles() error {
	// 检查是否已存在角色
	var count int64
	if err := DB.Model(&model.Role{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有角色，跳过初始化
	if count > 0 {
		return nil
	}

	// 创建系统默认角色
	defaultRoles := []model.Role{
		{
			Code:        "super_admin",
			Name:        "超级管理员",
			Description: "拥有系统所有权限的超级管理员角色",
			Type:        "system",
			IsSystem:    true,
			Status:      "active",
		},
		{
			Code:        "admin",
			Name:        "管理员",
			Description: "拥有大部分管理权限的管理员角色",
			Type:        "system",
			IsSystem:    true,
			Status:      "active",
		},
		{
			Code:        "operator",
			Name:        "操作员",
			Description: "拥有操作权限的操作员角色",
			Type:        "system",
			IsSystem:    true,
			Status:      "active",
		},
		{
			Code:        "viewer",
			Name:        "查看者",
			Description: "只拥有查看权限的查看者角色",
			Type:        "system",
			IsSystem:    true,
			Status:      "active",
		},
	}

	for _, role := range defaultRoles {
		if err := DB.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create default role %s: %w", role.Code, err)
		}
	}

	// 为角色分配权限
	if err := assignRolePermissions(); err != nil {
		return fmt.Errorf("failed to assign role permissions: %w", err)
	}

	return nil
}

// assignRolePermissions 为角色分配权限
func assignRolePermissions() error {
	// 获取所有权限
	var permissions []model.Permission
	if err := DB.Find(&permissions).Error; err != nil {
		return err
	}

	// 获取所有角色
	var roles []model.Role
	if err := DB.Find(&roles).Error; err != nil {
		return err
	}

	// 为每个角色分配相应权限
	for _, role := range roles {
		var rolePermissions []model.Permission
		
		switch role.Code {
		case "super_admin":
			// 超级管理员拥有所有权限
			rolePermissions = permissions
		case "admin":
			// 管理员拥有除系统管理外的所有权限
			for _, perm := range permissions {
				if perm.Category != "system" || perm.Action == "read" {
					rolePermissions = append(rolePermissions, perm)
				}
			}
		case "operator":
			// 操作员拥有告警、规则、数据源的操作权限，用户和角色的查看权限
			for _, perm := range permissions {
				if (perm.Category == "alert" || perm.Category == "rule" || perm.Category == "provider") ||
					(perm.Category == "user" || perm.Category == "role" || perm.Category == "permission") && perm.Action == "read" ||
					(perm.Category == "config" && perm.Action == "read") ||
					(perm.Category == "system" && perm.Action == "read") {
					rolePermissions = append(rolePermissions, perm)
				}
			}
		case "viewer":
			// 查看者只拥有查看权限
			for _, perm := range permissions {
				if perm.Action == "read" || perm.Action == "list" {
					rolePermissions = append(rolePermissions, perm)
				}
			}
		}

		// 为角色分配权限
		if len(rolePermissions) > 0 {
			if err := DB.Model(&role).Association("Permissions").Replace(rolePermissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", role.Code, err)
			}
		}
	}

	return nil
}

// seedUserRoles 初始化用户角色关联
func seedUserRoles() error {
	// 获取admin用户
	var adminUser model.User
	if err := DB.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有admin用户，跳过
			return nil
		}
		return err
	}

	// 获取超级管理员角色
	var superAdminRole model.Role
	if err := DB.Where("code = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有超级管理员角色，跳过
			return nil
		}
		return err
	}

	// 为admin用户分配超级管理员角色
	if err := DB.Model(&adminUser).Association("Roles").Append(&superAdminRole); err != nil {
		return fmt.Errorf("failed to assign super_admin role to admin user: %w", err)
	}

	return nil
}

// GetConnectionStats 获取连接池统计信息
func GetConnectionStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// OptimizeForHighLoad 为高负载场景优化数据库连接
func OptimizeForHighLoad() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 高负载场景的连接池配置
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// 高负载优化查询
	optimizationQueries := []string{
		"SET SESSION innodb_buffer_pool_dump_at_shutdown = ON",
		"SET SESSION innodb_buffer_pool_load_at_startup = ON",
		"SET SESSION innodb_adaptive_hash_index = ON",
		"SET SESSION innodb_change_buffering = all",
		"SET SESSION innodb_flush_log_at_trx_commit = 2", // 提高写性能，但可能丢失1秒数据
	}

	for _, query := range optimizationQueries {
		if err := DB.Exec(query).Error; err != nil {
			fmt.Printf("Warning: Failed to execute high-load optimization query: %s, error: %v\n", query, err)
		}
	}

	fmt.Println("Database optimized for high-load scenarios")
	return nil
}

// OptimizeForReadHeavy 为读密集型场景优化数据库连接
func OptimizeForReadHeavy() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 读密集型场景的连接池配置
	sqlDB.SetMaxIdleConns(30)
	sqlDB.SetMaxOpenConns(150)
	sqlDB.SetConnMaxLifetime(2 * time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	// 读优化查询
	optimizationQueries := []string{
		"SET SESSION innodb_read_ahead_threshold = 0",
		"SET SESSION read_buffer_size = 2097152",    // 2MB
		"SET SESSION read_rnd_buffer_size = 8388608", // 8MB
		"SET SESSION sort_buffer_size = 4194304",     // 4MB
	}

	for _, query := range optimizationQueries {
		if err := DB.Exec(query).Error; err != nil {
			fmt.Printf("Warning: Failed to execute read-heavy optimization query: %s, error: %v\n", query, err)
		}
	}

	fmt.Println("Database optimized for read-heavy scenarios")
	return nil
}

// AnalyzeTables 分析表以更新统计信息
func AnalyzeTables() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tables := []string{
		"alerts", "alert_rules", "rule_versions", "rule_distribution_records",
		"config_sync_status", "config_sync_history", "config_sync_exceptions",
		"task_queue", "worker_instances", "task_execution_history",
		"user_notification_configs", "notification_records", "notification_plugins",
		"knowledges", "providers", "notify_templates", "notify_groups",
	}

	for _, table := range tables {
		if err := DB.Exec(fmt.Sprintf("ANALYZE TABLE %s", table)).Error; err != nil {
			fmt.Printf("Warning: Failed to analyze table %s: %v\n", table, err)
		}
	}

	fmt.Println("Table analysis completed")
	return nil
}

// OptimizeTables 优化表以回收空间和重建索引
func OptimizeTables() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tables := []string{
		"alerts", "alert_rules", "config_sync_history", 
		"task_execution_history", "notification_records",
	}

	for _, table := range tables {
		if err := DB.Exec(fmt.Sprintf("OPTIMIZE TABLE %s", table)).Error; err != nil {
			fmt.Printf("Warning: Failed to optimize table %s: %v\n", table, err)
		}
	}

	fmt.Println("Table optimization completed")
	return nil
}

// GetTableSizes 获取表大小统计
func GetTableSizes() ([]map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var results []map[string]interface{}
	
	query := `
		SELECT 
			table_name,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb,
			table_rows,
			ROUND((index_length / 1024 / 1024), 2) as index_size_mb,
			ROUND((data_length / 1024 / 1024), 2) as data_size_mb
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		AND table_type = 'BASE TABLE'
		ORDER BY (data_length + index_length) DESC
	`

	rows, err := DB.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var sizeMB, indexSizeMB, dataSizeMB float64
		var tableRows int64

		if err := rows.Scan(&tableName, &sizeMB, &tableRows, &indexSizeMB, &dataSizeMB); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"table_name":     tableName,
			"size_mb":        sizeMB,
			"table_rows":     tableRows,
			"index_size_mb":  indexSizeMB,
			"data_size_mb":   dataSizeMB,
		})
	}

	return results, nil
}