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
	cfg := config.GlobalConfig.Database

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci&interpolateParams=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)
	// fmt.Println(dsn)
	fmt.Println(cfg.Charset)
	fmt.Println(cfg.Host)
	fmt.Println(cfg.Port)
	fmt.Println(cfg.DBName)
	fmt.Println(cfg.Username)
	fmt.Println(cfg.Password)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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

func autoMigrate() error {
	return DB.AutoMigrate(
		&model.Alert{},
		&model.Rule{},
		&model.NotifyTemplate{},
		&model.NotifyGroup{},
		&model.NotifyRecord{},
	)
}
