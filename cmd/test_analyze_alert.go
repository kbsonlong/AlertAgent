package main

import (
	"context"
	"fmt"
	"time"

	"alert_agent/internal/config"
	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 初始化日志
	logger.InitLogger("debug", "console")
	log := zap.L()
	log.Info("Starting alert analysis test")

	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GlobalConfig.Database.Username,
		config.GlobalConfig.Database.Password,
		config.GlobalConfig.Database.Host,
		config.GlobalConfig.Database.Port,
		config.GlobalConfig.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	database.DB = db
	log.Info("Database connected successfully")

	// 创建OpenAI服务实例
	openAIService := service.NewOpenAIService(&service.OpenAIConfig{
		Endpoint:   config.GlobalConfig.OpenAI.Endpoint,
		Model:      config.GlobalConfig.OpenAI.Model,
		Timeout:    config.GlobalConfig.OpenAI.Timeout,
		MaxRetries: config.GlobalConfig.OpenAI.MaxRetries,
	})

	// 创建示例告警
	alert := &model.Alert{
		Name:      "CPU使用率过高",
		Level:     "critical",
		Status:    "active",
		Source:    "node-exporter",
		Content:   "服务器192.168.1.100的CPU使用率已经超过95%持续5分钟，当前使用率为98%。该服务器运行着多个关键业务应用，包括订单处理系统和支付网关。",
		RuleID:    1,
		Title:     "高CPU使用率告警",
		CreatedAt: time.Now(),
	}

	// 调用AI分析
	ctx := context.Background()
	log.Info("Analyzing alert with AI",
		zap.String("alert_name", alert.Name),
		zap.String("alert_level", alert.Level),
	)

	analysis, err := openAIService.AnalyzeAlert(ctx, alert)
	if err != nil {
		log.Error("Failed to analyze alert", zap.Error(err))
		return
	}

	// 打印分析结果
	log.Info("Alert analysis completed")
	fmt.Println("\n===== 告警分析结果 =====")
	fmt.Println(analysis)
	fmt.Println("=======================\n")

	// 更新告警分析结果
	alert.Analysis = analysis
	if database.DB != nil {
		result := database.DB.Create(alert)
		if result.Error != nil {
			log.Error("Failed to save alert", zap.Error(result.Error))
		} else {
			log.Info("Alert saved to database", zap.Uint("alert_id", alert.ID))
		}
	}
}
