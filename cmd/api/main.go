package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alert_agent/internal/infrastructure/config"
	"alert_agent/internal/infrastructure/database"
	"alert_agent/internal/infrastructure/di"
	"alert_agent/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志器
	logger := logger.NewLogger()
	defer logger.Sync()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// 初始化数据库
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// 运行数据库迁移
	if err := database.Migrate(db); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 测试 Redis 连接
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := redisClient.Ping(ctx2).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}

	// 创建依赖注入容器
	container := di.NewContainer(db, redisClient, logger, cfg)
	defer container.Close()

	// 设置Gin模式
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin引擎
	engine := gin.New()

	// 添加中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(requestIDMiddleware())

	// 设置路由
	router := container.GetHTTPRouter()
	router.SetupRoutes(engine)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("starting server", 
			zap.String("address", server.Addr),
			zap.String("environment", cfg.App.Environment),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware 请求ID中间件
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}