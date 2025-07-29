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

	"github.com/gin-gonic/gin"

	"alert_agent/internal/config"
	"alert_agent/internal/infrastructure/di"
	"alert_agent/internal/shared/middleware"
)

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化应用程序
	ctx := context.Background()
	app, cleanup, err := di.InitializeApplication(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	// 设置Gin模式
	cfg := config.GetConfig()
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := setupRouter(app)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		app.Logger.Info("Starting server", 
			"port", cfg.Server.Port,
			"mode", cfg.Server.Mode)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown", "error", err)
	}

	app.Logger.Info("Server exited")
}

// setupRouter 设置路由
func setupRouter(app *di.Application) *gin.Engine {
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryHandler())
	router.Use(middleware.ErrorHandler())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 渠道管理路由
		channels := v1.Group("/channels")
		{
			channels.POST("", createChannel(app))
			channels.GET("", listChannels(app))
			channels.GET("/:id", getChannel(app))
			channels.PUT("/:id", updateChannel(app))
			channels.DELETE("/:id", deleteChannel(app))
			channels.POST("/:id/test", testChannel(app))
			channels.GET("/:id/health", getChannelHealth(app))
		}

		// 集群管理路由
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", registerCluster(app))
			clusters.GET("", listClusters(app))
			clusters.GET("/:id", getCluster(app))
			clusters.PUT("/:id", updateCluster(app))
			clusters.DELETE("/:id", deleteCluster(app))
			clusters.GET("/:id/health", getClusterHealth(app))
			clusters.POST("/:id/sync", syncClusterConfig(app))
		}

		// 网关路由
		gateway := v1.Group("/gateway")
		{
			gateway.POST("/alerts", receiveAlert(app))
			gateway.GET("/processing/:alert_id", getProcessingRecord(app))
			gateway.GET("/processing", listProcessingRecords(app))
		}

		// 分析路由
		analysis := v1.Group("/analysis")
		{
			analysis.POST("", createAnalysis(app))
			analysis.GET("/:id", getAnalysis(app))
			analysis.GET("", listAnalysis(app))
			analysis.GET("/stats", getAnalysisStats(app))
		}
	}

	return router
}

// 渠道管理处理器
func createChannel(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func listChannels(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getChannel(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func updateChannel(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func deleteChannel(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func testChannel(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getChannelHealth(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

// 集群管理处理器
func registerCluster(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func listClusters(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getCluster(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func updateCluster(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func deleteCluster(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getClusterHealth(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func syncClusterConfig(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

// 网关处理器
func receiveAlert(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getProcessingRecord(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func listProcessingRecords(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

// 分析处理器
func createAnalysis(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getAnalysis(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func listAnalysis(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}

func getAnalysisStats(app *di.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
	}
}