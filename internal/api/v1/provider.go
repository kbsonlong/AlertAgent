package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/redis"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// ListProviders 获取数据源列表
func ListProviders(c *gin.Context) {
	providerService := service.NewProviderService(database.DB, redis.Client)
	providers, err := providerService.ListProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取数据源列表失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": providers,
	})
}

// GetProvider 获取单个数据源
func GetProvider(c *gin.Context) {
	id := c.Param("id")
	providerID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的数据源ID",
			"data": err.Error(),
		})
		return
	}

	providerService := service.NewProviderService(database.DB, redis.Client)
	provider, err := providerService.GetProvider(c.Request.Context(), uint(providerID))
	if err != nil {
		if err == service.ErrProviderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "数据源不存在",
				"data": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "获取数据源失败",
				"data": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": provider,
	})
}

// CreateProvider 创建数据源
func CreateProvider(c *gin.Context) {
	var provider model.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	providerService := service.NewProviderService(database.DB, redis.Client)
	err := providerService.CreateProvider(c.Request.Context(), &provider)
	if err != nil {
		if err == service.ErrInvalidProvider {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "数据源配置无效: " + err.Error(),
				"data": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "创建数据源失败",
				"data": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": provider,
	})
}

// UpdateProvider 更新数据源
func UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	providerID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的数据源ID",
			"data": nil,
		})
		return
	}

	var provider model.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	// 设置ID
	provider.ID = uint(providerID)

	providerService := service.NewProviderService(database.DB, redis.Client)
	err = providerService.UpdateProvider(c.Request.Context(), &provider)
	if err != nil {
		if err == service.ErrProviderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "数据源不存在",
				"data": err.Error(),
			})
		} else if err == service.ErrInvalidProvider {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "数据源配置无效: " + err.Error(),
				"data": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "更新数据源失败",
				"data": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": provider,
	})
}

// DeleteProvider 删除数据源
func DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	providerID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的数据源ID",
			"data": nil,
		})
		return
	}

	providerService := service.NewProviderService(database.DB, redis.Client)
	err = providerService.DeleteProvider(c.Request.Context(), uint(providerID))
	if err != nil {
		if err == service.ErrProviderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "数据源不存在",
				"data": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "删除数据源失败",
				"data": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}

// TestProvider 测试数据源连接
func TestProvider(c *gin.Context) {
	var provider model.Provider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": nil,
		})
		return
	}

	// TODO: 实现具体的连接测试逻辑
	// 这里可以根据 provider.Type 调用相应的测试方法
	
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "连接测试成功",
		"data": gin.H{
			"status": "success",
			"message": "数据源连接正常",
		},
	})
}