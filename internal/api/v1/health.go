package v1

import (
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查系统各组件的健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HealthStatus}
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceInfo),
	}

	// 检查数据库连接
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			health.Services["database"] = ServiceInfo{
				Status:  "unhealthy",
				Message: "Database connection failed",
			}
			health.Status = "degraded"
		} else {
			health.Services["database"] = ServiceInfo{
				Status: "healthy",
			}
		}
	} else {
		health.Services["database"] = ServiceInfo{
			Status:  "unhealthy",
			Message: "Database not initialized",
		}
		health.Status = "degraded"
	}

	// 检查Redis连接
	if redis.Client != nil {
		if err := redis.Client.Ping(c.Request.Context()).Err(); err != nil {
			health.Services["redis"] = ServiceInfo{
				Status:  "unhealthy",
				Message: "Redis connection failed",
			}
			health.Status = "degraded"
		} else {
			health.Services["redis"] = ServiceInfo{
				Status: "healthy",
			}
		}
	} else {
		health.Services["redis"] = ServiceInfo{
			Status:  "unhealthy",
			Message: "Redis not initialized",
		}
		health.Status = "degraded"
	}

	response.Success(c, health)
}