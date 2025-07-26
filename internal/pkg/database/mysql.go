package database

import (
	"fmt"
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

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

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

	// 获取默认数据源ID
	var provider model.Provider
	if err := DB.First(&provider, "name = ?", "本地Prometheus").Error; err != nil {
		return fmt.Errorf("failed to find default provider: %w", err)
	}

	// 创建默认告警规则
	defaultRules := []model.Rule{
		{
			Name:          "CPU使用率告警",
			Description:   "CPU使用率超过90%时触发告警",
			Level:         "warning",
			Enabled:       true,
			ProviderID:    provider.ID,
			QueryExpr:     "100 - (avg by (instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
			ConditionExpr: "cpu_usage > 90",
			NotifyType:    "email",
			NotifyGroup:   "默认通知组",
			Template:      "默认邮件模板",
		},
		{
			Name:          "内存使用率告警",
			Description:   "内存使用率超过90%时触发告警",
			Level:         "warning",
			Enabled:       true,
			ProviderID:    provider.ID,
			QueryExpr:     "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
			ConditionExpr: "memory_usage > 90",
			NotifyType:    "email",
			NotifyGroup:   "默认通知组",
			Template:      "默认邮件模板",
		},
		{
			Name:          "磁盘使用率告警",
			Description:   "磁盘使用率超过90%时触发告警",
			Level:         "critical",
			Enabled:       true,
			ProviderID:    provider.ID,
			QueryExpr:     "(1 - (node_filesystem_avail_bytes{fstype!=\"tmpfs\"} / node_filesystem_size_bytes{fstype!=\"tmpfs\"})) * 100",
			ConditionExpr: "disk_usage > 90",
			NotifyType:    "email",
			NotifyGroup:   "默认通知组",
			Template:      "默认邮件模板",
		},
		{
			Name:          "服务可用性告警",
			Description:   "服务不可用时触发告警",
			Level:         "critical",
			Enabled:       true,
			ProviderID:    provider.ID,
			QueryExpr:     "up",
			ConditionExpr: "up == 0",
			NotifyType:    "email",
			NotifyGroup:   "默认通知组",
			Template:      "默认邮件模板",
		},
	}

	for _, rule := range defaultRules {
		if err := DB.Create(&rule).Error; err != nil {
			return fmt.Errorf("failed to create default rule %s: %w", rule.Name, err)
		}
	}

	return nil
}

// seedAlerts 插入默认告警示例
func seedAlerts() error {
	// 检查是否已存在默认告警示例
	var count int64
	if err := DB.Model(&model.Alert{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有告警，跳过初始化
	if count > 0 {
		return nil
	}

	// 获取默认规则ID
	var rule model.Rule
	if err := DB.First(&rule, "name = ?", "CPU使用率告警").Error; err != nil {
		return fmt.Errorf("failed to find default rule: %w", err)
	}

	// 创建默认告警示例
	defaultAlerts := []model.Alert{
		{
			Name:     "CPU使用率过高",
			Title:    "服务器CPU使用率超过阈值",
			Level:    "warning",
			Status:   "new",
			Source:   "prometheus",
			Content:  "服务器 server-01 的CPU使用率已达到95%，超过了90%的告警阈值。请及时检查系统负载情况。",
			Labels:   `{"instance":"server-01","job":"node-exporter","severity":"warning"}`,
			RuleID:   rule.ID,
			Severity: "medium",
		},
		{
			Name:     "内存使用率告警",
			Title:    "服务器内存使用率异常",
			Level:    "warning",
			Status:   "acknowledged",
			Source:   "prometheus",
			Content:  "服务器 server-02 的内存使用率已达到92%，可能存在内存泄漏或负载过高的情况。",
			Labels:   `{"instance":"server-02","job":"node-exporter","severity":"warning"}`,
			RuleID:   rule.ID,
			Severity: "medium",
			Handler:  "admin",
			HandleNote: "已确认告警，正在排查内存使用情况",
		},
		{
			Name:     "磁盘空间不足",
			Title:    "服务器磁盘空间严重不足",
			Level:    "critical",
			Status:   "new",
			Source:   "prometheus",
			Content:  "服务器 server-03 的根分区磁盘使用率已达到95%，请立即清理磁盘空间或扩容。",
			Labels:   `{"instance":"server-03","device":"/dev/sda1","mountpoint":"/","severity":"critical"}`,
			RuleID:   rule.ID,
			Severity: "high",
		},
	}

	for _, alert := range defaultAlerts {
		if err := DB.Create(&alert).Error; err != nil {
			return fmt.Errorf("failed to create default alert %s: %w", alert.Name, err)
		}
	}

	return nil
}

func autoMigrate() error {
	// 自动迁移数据库表
	if err := DB.AutoMigrate(
		&model.Alert{},
		&model.Rule{},
		&model.NotifyTemplate{},
		&model.NotifyGroup{},
		&model.NotifyRecord{},
		&model.Provider{},
		&model.Knowledge{},
	); err != nil {
		return err
	}

	// 插入初始化数据
	if err := seedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

// seedData 插入初始化数据
func seedData() error {
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

	// 检查并插入默认告警示例
	if err := seedAlerts(); err != nil {
		return err
	}

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
		Members:     "admin@example.com", // 可根据实际需求修改
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
