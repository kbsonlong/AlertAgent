package gateway

import (
	"alert_agent/internal/config"
	"alert_agent/internal/middleware"
	"alert_agent/internal/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Gateway API网关结构
type Gateway struct {
	engine *gin.Engine
	server *http.Server
	config *config.Config
}

// NewGateway 创建新的API网关实例
func NewGateway() *Gateway {
	cfg := config.GetConfig()
	
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)
	
	engine := gin.New()
	
	return &Gateway{
		engine: engine,
		config: &cfg,
	}
}

// SetupMiddleware 设置中间件
func (g *Gateway) SetupMiddleware() {
	// 基础中间件
	g.engine.Use(middleware.Logger())
	g.engine.Use(middleware.Recovery())
	g.engine.Use(middleware.Cors())
	
	// 限流中间件
	g.engine.Use(middleware.RateLimit())
	
	// 请求ID中间件
	g.engine.Use(middleware.RequestID())
	
	// 响应格式化中间件
	g.engine.Use(middleware.ResponseFormatter())
}

// SetupRoutes 设置路由
func (g *Gateway) SetupRoutes(routeSetup func(*gin.Engine)) {
	routeSetup(g.engine)
}

// Start 启动网关
func (g *Gateway) Start() error {
	// 创建HTTP服务器
	g.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", g.config.Server.Port),
		Handler:        g.engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启动服务器
	go func() {
		logger.L.Info("Starting API Gateway", 
			zap.Int("port", g.config.Server.Port),
			zap.String("mode", g.config.Server.Mode),
		)
		
		if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatal("Failed to start API Gateway", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止网关
func (g *Gateway) Stop() error {
	if g.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.L.Info("Shutting down API Gateway...")
	
	if err := g.server.Shutdown(ctx); err != nil {
		logger.L.Error("Failed to shutdown API Gateway gracefully", zap.Error(err))
		return err
	}

	logger.L.Info("API Gateway stopped")
	return nil
}

// WaitForShutdown 等待关闭信号
func (g *Gateway) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	if err := g.Stop(); err != nil {
		logger.L.Error("Error during gateway shutdown", zap.Error(err))
	}
}

// GetEngine 获取Gin引擎实例
func (g *Gateway) GetEngine() *gin.Engine {
	return g.engine
}