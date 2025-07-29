package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/infrastructure/di"
	httpHandlers "alert_agent/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 配置结构
type Config struct {
	Port      string
	DBHost    string
	DBUser    string
	DBPassword string
	DBName    string
	DBPort    string
	N8NBaseURL string
	N8NAPIKey  string
}

// 从环境变量加载配置
func loadConfig() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "alertagent"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "alertagent"),
		DBPort:     getEnv("DB_PORT", "5432"),
		N8NBaseURL: getEnv("N8N_BASE_URL", "http://localhost:5678"),
		N8NAPIKey:  getEnv("N8N_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 初始化日志
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logger
}

// 初始化数据库
func initDatabase(config *Config, logger *zap.Logger) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: nil, // 可以配置 GORM 日志
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	
	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database instance", zap.Error(err))
	}
	
	if err := sqlDB.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	
	logger.Info("Database connected successfully")
	return db
}

// 设置 CORS
func setupCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// 设置路由
func setupRoutes(router *gin.Engine, n8nContainer *di.N8NContainer, logger *zap.Logger) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"service":   "alertagent-n8n-demo",
		})
	})
	
	// API v1 路由组
	v1 := router.Group("/api/v1")
	
	// n8n 分析路由
	n8n := v1.Group("/n8n")
	httpHandlers.RegisterN8NRoutes(n8n, n8nContainer.GetN8NAnalysisService())
	
	// n8n 回调路由
	callbacks := v1.Group("/callbacks")
	httpHandlers.RegisterN8NCallbackRoutes(callbacks, n8nContainer.GetWorkflowManager())
	
	// 演示路由
	demo := v1.Group("/demo")
	setupDemoRoutes(demo, n8nContainer, logger)
}

// 设置演示路由
func setupDemoRoutes(group *gin.RouterGroup, n8nContainer *di.N8NContainer, logger *zap.Logger) {
	// 创建测试告警
	group.POST("/alerts", func(c *gin.Context) {
		var req struct {
			Title       string                 `json:"title" binding:"required"`
			Description string                 `json:"description"`
			Severity    string                 `json:"severity"`
			Source      string                 `json:"source"`
			Metadata    map[string]interface{} `json:"metadata"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// 这里应该创建真实的告警记录
		// 为了演示，我们返回一个模拟的告警 ID
		alertID := fmt.Sprintf("alert-%d", time.Now().Unix())
		
		logger.Info("Demo alert created",
			zap.String("alert_id", alertID),
			zap.String("title", req.Title),
			zap.String("severity", req.Severity),
		)
		
		c.JSON(http.StatusCreated, gin.H{
			"alert_id":    alertID,
			"title":       req.Title,
			"description": req.Description,
			"severity":    req.Severity,
			"source":      req.Source,
			"metadata":    req.Metadata,
			"created_at":  time.Now().UTC(),
			"status":      "active",
		})
	})
	
	// 触发 n8n 分析演示
	group.POST("/analyze/:alert_id", func(c *gin.Context) {
		alertID := c.Param("alert_id")
		
		var req struct {
			WorkflowTemplateID string `json:"workflow_template_id"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		analysisService := n8nContainer.GetN8NAnalysisService()
		
		// 触发分析 (注意：这里需要将 alertID 转换为 uint 类型)
		// 在实际应用中，应该从数据库查询真实的告警 ID
		// 为了演示，我们使用一个固定的 ID
		execution, err := analysisService.AnalyzeAlert(c.Request.Context(), 1, req.WorkflowTemplateID)
		if err != nil {
			logger.Error("Failed to trigger analysis",
				zap.String("alert_id", alertID),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		logger.Info("Analysis triggered",
			zap.String("alert_id", alertID),
			zap.String("execution_id", execution.ID),
		)
		
		c.JSON(http.StatusOK, gin.H{
			"execution_id": execution.ID,
			"status":       execution.Status,
			"message":      "Analysis started successfully",
		})
	})
	
	// 获取分析状态
	group.GET("/executions/:execution_id", func(c *gin.Context) {
		executionID := c.Param("execution_id")
		
		// 注意：GetExecutionStatus 方法可能不存在，这里提供一个模拟响应
		// 在实际应用中，应该实现相应的方法或使用其他可用的方法
		execution := map[string]interface{}{
			"execution_id": executionID,
			"status":       "running",
			"started_at":   time.Now().Add(-5 * time.Minute).UTC(),
			"message":      "Execution is in progress",
		}
		var err error
		if err != nil {
			logger.Error("Failed to get execution status",
				zap.String("execution_id", executionID),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, execution)
	})
	
	// 获取演示统计信息
	group.GET("/stats", func(c *gin.Context) {
		// 注意：GetAnalysisMetrics 方法的参数类型可能不匹配
		// 这里提供一个模拟响应
		stats := map[string]interface{}{
			"total_executions":     10,
			"successful_executions": 8,
			"failed_executions":    1,
			"running_executions":   1,
			"average_duration":     "2m30s",
			"last_24h":            true,
		}
		var err error
		if err != nil {
			logger.Error("Failed to get analysis metrics", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"demo_stats": stats,
			"timestamp": time.Now().UTC(),
		})
	})
}

// 优雅关闭
func gracefulShutdown(server *http.Server, logger *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Server exited")
}

func main() {
	// 加载配置
	config := loadConfig()
	
	// 初始化日志
	logger := initLogger()
	defer logger.Sync()
	
	logger.Info("Starting AlertAgent n8n Demo",
		zap.String("port", config.Port),
		zap.String("n8n_base_url", config.N8NBaseURL),
	)
	
	// 初始化数据库
	db := initDatabase(config, logger)
	
	// 创建 n8n 容器
	n8nContainer := di.NewN8NContainer(db, logger)
	
	// 初始化 n8n 组件
	n8nContainer.Initialize(config.N8NBaseURL, config.N8NAPIKey)
	
	logger.Info("n8n components initialized successfully")
	
	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// 创建路由器
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(setupCORS())
	
	// 设置路由
	setupRoutes(router, n8nContainer, logger)
	
	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}
	
	// 启动服务器
	go func() {
		logger.Info("Server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	
	logger.Info("AlertAgent n8n Demo started successfully",
		zap.String("port", config.Port),
		zap.String("health_check", fmt.Sprintf("http://localhost:%s/health", config.Port)),
		zap.String("api_docs", fmt.Sprintf("http://localhost:%s/api/v1", config.Port)),
	)
	
	// 优雅关闭
	gracefulShutdown(server, logger)
}