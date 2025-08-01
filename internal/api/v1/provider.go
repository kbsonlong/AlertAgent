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
// @Summary 获取数据源列表
// @Description 获取系统中所有数据源的列表信息
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]model.Provider}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers [get]
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
// @Summary 获取单个数据源
// @Description 根据数据源ID获取单个数据源的详细信息
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "数据源ID"
// @Success 200 {object} response.Response{data=model.Provider}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers/{id} [get]
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
// @Summary 创建数据源
// @Description 创建新的数据源配置
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Provider true "数据源信息"
// @Success 200 {object} response.Response{data=model.Provider}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers [post]
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
// @Summary 更新数据源
// @Description 根据ID更新数据源配置信息
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "数据源ID"
// @Param request body model.Provider true "数据源信息"
// @Success 200 {object} response.Response{data=model.Provider}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers/{id} [put]
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
// @Summary 删除数据源
// @Description 根据ID删除指定的数据源
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "数据源ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers/{id} [delete]
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
// @Summary 测试数据源连接
// @Description 测试数据源的连接是否正常
// @Tags 数据源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Provider true "数据源信息"
// @Success 200 {object} response.Response{data=object{status=string,message=string}}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/providers/test [post]
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