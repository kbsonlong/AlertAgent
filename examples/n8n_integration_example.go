package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/infrastructure/di"
	httpHandlers "alert_agent/internal/interfaces/http"
	"alert_agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// N8NIntegrationExample n8n 集成示例
func N8NIntegrationExample() {
	// 初始化日志
	logger := logger.GetLogger()
	defer logger.Sync()

	// 数据库连接
	dsn := "host=localhost user=alertagent password=password dbname=alertagent port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 创建 n8n 容器
	n8nContainer := di.NewN8NContainer(db, logger)

	// 初始化 n8n 组件
	n8nBaseURL := getEnvOrDefault("N8N_BASE_URL", "http://localhost:5678")
	n8nAPIKey := getEnvOrDefault("N8N_API_KEY", "your-n8n-api-key")
	n8nContainer.Initialize(n8nBaseURL, n8nAPIKey)

	// 创建 Gin 路由器
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 设置 CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Alert Agent with n8n integration is running",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")

	// 注册 n8n 分析路由
	n8n := v1.Group("/n8n")
	httpHandlers.RegisterN8NRoutes(n8n, n8nContainer.GetN8NAnalysisService())

	// 注册 n8n 回调路由
	callbacks := v1.Group("/callbacks")
	httpHandlers.RegisterN8NCallbackRoutes(callbacks, n8nContainer.GetWorkflowManager())

	// 创建 HTTP 服务器
	port := getEnvOrDefault("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 清理 n8n 容器资源
	n8nContainer.Cleanup()

	logger.Info("Server exited")
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 示例：如何使用 n8n 分析服务
func exampleUsage() {
	// 这是一个示例函数，展示如何使用 n8n 分析服务
	// 在实际应用中，这些调用会在业务逻辑中进行

	/*
	// 1. 分析单个告警
	ctx := context.Background()
	alertID := uint(123)
	workflowTemplateID := "workflow-template-id"

	execution, err := n8nService.AnalyzeAlert(ctx, alertID, workflowTemplateID)
	if err != nil {
		log.Printf("Failed to analyze alert: %v", err)
		return
	}
	log.Printf("Analysis started: %s", execution.ID)

	// 2. 批量分析告警
	config := analysis.N8NAnalysisConfig{
		DefaultWorkflowTemplateID: "batch-workflow-template-id",
		BatchSize:                 10,
		ProcessInterval:           5 * time.Second,
		MaxRetries:                3,
		Timeout:                   300 * time.Second,
		AutoAnalysisEnabled:       true,
	}

	err = n8nService.BatchAnalyzeAlerts(ctx, config)
	if err != nil {
		log.Printf("Failed to start batch analysis: %v", err)
		return
	}
	log.Println("Batch analysis started")

	// 3. 获取分析状态
	executionID := "execution-id"
	status, err := n8nService.GetAnalysisStatus(ctx, executionID)
	if err != nil {
		log.Printf("Failed to get analysis status: %v", err)
		return
	}
	log.Printf("Analysis status: %s", status.Status)

	// 4. 获取分析历史
	history, err := n8nService.GetAnalysisHistory(ctx, alertID, 10)
	if err != nil {
		log.Printf("Failed to get analysis history: %v", err)
		return
	}
	log.Printf("Found %d analysis records", len(history))

	// 5. 获取分析指标
	timeRange := alert.TimeRange{
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
	}
	metrics, err := n8nService.GetAnalysisMetrics(ctx, timeRange)
	if err != nil {
		log.Printf("Failed to get analysis metrics: %v", err)
		return
	}
	log.Printf("Total executions: %d", metrics.TotalExecutions)
	*/
}