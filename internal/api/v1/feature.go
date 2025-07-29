package v1

import (
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/pkg/feature"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FeatureHandler 功能开关处理器
type FeatureHandler struct {
	featureService *service.FeatureService
	logger         *zap.Logger
}

// NewFeatureHandler 创建功能开关处理器
func NewFeatureHandler(featureService *service.FeatureService, logger *zap.Logger) *FeatureHandler {
	return &FeatureHandler{
		featureService: featureService,
		logger:         logger,
	}
}

// ListFeatures 列出所有功能
// @Summary 列出所有功能开关
// @Description 获取系统中所有功能开关的配置信息
// @Tags features
// @Accept json
// @Produce json
// @Param phase query string false "阶段过滤 (phase_one, phase_two)"
// @Success 200 {object} map[string]interface{} "功能列表"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features [get]
func (h *FeatureHandler) ListFeatures(c *gin.Context) {
	phase := c.Query("phase")
	
	var features map[feature.FeatureName]*feature.FeatureConfig
	
	if phase != "" {
		switch phase {
		case "phase_one":
			features = h.featureService.GetPhaseFeatures(feature.PhaseOne)
		case "phase_two":
			features = h.featureService.GetPhaseFeatures(feature.PhaseTwo)
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "Invalid phase parameter",
			})
			return
		}
	} else {
		features = h.featureService.ListFeatures()
	}
	
	// 转换为响应格式
	response := make(map[string]interface{})
	for name, config := range features {
		response[string(name)] = map[string]interface{}{
			"name":        config.Name,
			"phase":       config.Phase,
			"state":       config.State,
			"description": config.Description,
			"dependencies": config.Dependencies,
			"created_at":  config.CreatedAt,
			"updated_at":  config.UpdatedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": response,
	})
}

// GetFeature 获取单个功能配置
// @Summary 获取功能开关详情
// @Description 获取指定功能开关的详细配置信息
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Success 200 {object} map[string]interface{} "功能详情"
// @Failure 404 {object} map[string]interface{} "功能不存在"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/{name} [get]
func (h *FeatureHandler) GetFeature(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	config, err := h.featureService.GetFeature(featureName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": config,
	})
}

// UpdateFeatureRequest 更新功能请求
type UpdateFeatureRequest struct {
	State       feature.FeatureState `json:"state" binding:"required"`
	Description string               `json:"description"`
}

// UpdateFeature 更新功能配置
// @Summary 更新功能开关状态
// @Description 更新指定功能开关的状态和配置
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Param request body UpdateFeatureRequest true "更新请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "功能不存在"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/{name} [put]
func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	var req UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid request parameters",
			"data": err.Error(),
		})
		return
	}
	
	// 获取当前配置
	currentConfig, err := h.featureService.GetFeature(featureName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  err.Error(),
		})
		return
	}
	
	// 更新配置
	newConfig := *currentConfig
	newConfig.State = req.State
	if req.Description != "" {
		newConfig.Description = req.Description
	}
	newConfig.UpdatedAt = time.Now()
	
	if err := h.featureService.UpdateFeature(featureName, &newConfig); err != nil {
		h.logger.Error("Failed to update feature",
			zap.String("feature", string(featureName)),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to update feature",
			"data": err.Error(),
		})
		return
	}
	
	h.logger.Info("Feature updated",
		zap.String("feature", string(featureName)),
		zap.String("state", string(req.State)))
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Feature updated successfully",
		"data": newConfig,
	})
}

// CheckFeature 检查功能是否启用
// @Summary 检查功能开关状态
// @Description 检查指定功能开关是否启用，支持用户上下文
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Param user_group query string false "用户组"
// @Param cluster query string false "集群"
// @Success 200 {object} map[string]interface{} "检查结果"
// @Failure 404 {object} map[string]interface{} "功能不存在"
// @Router /api/v1/features/{name}/check [get]
func (h *FeatureHandler) CheckFeature(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	// 构建用户上下文
	userContext := make(map[string]interface{})
	if userGroup := c.Query("user_group"); userGroup != "" {
		userContext["user_group"] = userGroup
	}
	if cluster := c.Query("cluster"); cluster != "" {
		userContext["cluster"] = cluster
	}
	
	enabled := h.featureService.IsEnabled(c.Request.Context(), featureName, userContext)
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": gin.H{
			"feature": featureName,
			"enabled": enabled,
			"context": userContext,
		},
	})
}

// EnablePhase 启用阶段
// @Summary 启用指定阶段的所有功能
// @Description 批量启用指定阶段的所有功能开关
// @Tags features
// @Accept json
// @Produce json
// @Param phase path string true "阶段名称 (phase_one, phase_two)"
// @Success 200 {object} map[string]interface{} "启用成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/phases/{phase}/enable [post]
func (h *FeatureHandler) EnablePhase(c *gin.Context) {
	phaseStr := c.Param("phase")
	
	var phase feature.Phase
	switch phaseStr {
	case "phase_one":
		phase = feature.PhaseOne
	case "phase_two":
		phase = feature.PhaseTwo
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid phase parameter",
		})
		return
	}
	
	if err := h.featureService.EnablePhase(c.Request.Context(), phase); err != nil {
		h.logger.Error("Failed to enable phase",
			zap.String("phase", string(phase)),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to enable phase",
			"data": err.Error(),
		})
		return
	}
	
	h.logger.Info("Phase enabled", zap.String("phase", string(phase)))
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Phase enabled successfully",
		"data": gin.H{
			"phase": phase,
		},
	})
}

// DisablePhase 禁用阶段
// @Summary 禁用指定阶段的所有功能
// @Description 批量禁用指定阶段的所有功能开关
// @Tags features
// @Accept json
// @Produce json
// @Param phase path string true "阶段名称 (phase_one, phase_two)"
// @Success 200 {object} map[string]interface{} "禁用成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/phases/{phase}/disable [post]
func (h *FeatureHandler) DisablePhase(c *gin.Context) {
	phaseStr := c.Param("phase")
	
	var phase feature.Phase
	switch phaseStr {
	case "phase_one":
		phase = feature.PhaseOne
	case "phase_two":
		phase = feature.PhaseTwo
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid phase parameter",
		})
		return
	}
	
	if err := h.featureService.DisablePhase(c.Request.Context(), phase); err != nil {
		h.logger.Error("Failed to disable phase",
			zap.String("phase", string(phase)),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to disable phase",
			"data": err.Error(),
		})
		return
	}
	
	h.logger.Info("Phase disabled", zap.String("phase", string(phase)))
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Phase disabled successfully",
		"data": gin.H{
			"phase": phase,
		},
	})
}

// GetAIMaturity 获取AI成熟度评估
// @Summary 获取AI成熟度评估结果
// @Description 获取指定功能的AI模型成熟度评估结果
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Success 200 {object} map[string]interface{} "评估结果"
// @Failure 404 {object} map[string]interface{} "功能不存在或无评估数据"
// @Router /api/v1/features/{name}/ai-maturity [get]
func (h *FeatureHandler) GetAIMaturity(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	assessment, err := h.featureService.GetAIMaturityAssessment(featureName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": assessment,
	})
}

// RecordAIMetricsRequest 记录AI指标请求
type RecordAIMetricsRequest struct {
	Accuracy    float64 `json:"accuracy" binding:"required,min=0,max=1"`
	Confidence  float64 `json:"confidence" binding:"required,min=0,max=1"`
	Latency     int     `json:"latency" binding:"required,min=0"`
	SuccessRate float64 `json:"success_rate" binding:"required,min=0,max=1"`
	ErrorRate   float64 `json:"error_rate" binding:"required,min=0,max=1"`
	SampleCount int     `json:"sample_count" binding:"required,min=1"`
}

// RecordAIMetrics 记录AI指标
// @Summary 记录AI模型指标
// @Description 记录指定功能的AI模型性能指标
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Param request body RecordAIMetricsRequest true "指标数据"
// @Success 200 {object} map[string]interface{} "记录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Router /api/v1/features/{name}/ai-metrics [post]
func (h *FeatureHandler) RecordAIMetrics(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	var req RecordAIMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid request parameters",
			"data": err.Error(),
		})
		return
	}
	
	metrics := feature.AIMetrics{
		Accuracy:    req.Accuracy,
		Confidence:  req.Confidence,
		Latency:     req.Latency,
		SuccessRate: req.SuccessRate,
		ErrorRate:   req.ErrorRate,
		SampleCount: req.SampleCount,
		Timestamp:   time.Now(),
	}
	
	h.featureService.RecordAIMetrics(featureName, metrics)
	
	h.logger.Debug("AI metrics recorded",
		zap.String("feature", string(featureName)),
		zap.Float64("accuracy", metrics.Accuracy),
		zap.Float64("confidence", metrics.Confidence))
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "AI metrics recorded successfully",
	})
}

// GetMonitoringReport 获取监控报告
// @Summary 获取功能监控报告
// @Description 获取指定功能的监控报告和统计信息
// @Tags features
// @Accept json
// @Produce json
// @Param name path string true "功能名称"
// @Param hours query int false "报告时间范围（小时）" default(24)
// @Success 200 {object} map[string]interface{} "监控报告"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/{name}/monitoring [get]
func (h *FeatureHandler) GetMonitoringReport(c *gin.Context) {
	featureName := feature.FeatureName(c.Param("name"))
	
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid hours parameter",
		})
		return
	}
	
	duration := time.Duration(hours) * time.Hour
	report, err := h.featureService.GetMonitoringReport(featureName, duration)
	if err != nil {
		h.logger.Error("Failed to generate monitoring report",
			zap.String("feature", string(featureName)),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to generate monitoring report",
			"data": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": report,
	})
}

// GetActiveAlerts 获取活跃告警
// @Summary 获取活跃的功能告警
// @Description 获取当前所有活跃的功能相关告警
// @Tags features
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "活跃告警列表"
// @Router /api/v1/features/alerts [get]
func (h *FeatureHandler) GetActiveAlerts(c *gin.Context) {
	alerts := h.featureService.GetActiveAlerts()
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": alerts,
	})
}

// ExportConfig 导出功能配置
// @Summary 导出功能配置
// @Description 导出当前所有功能开关的配置为JSON格式
// @Tags features
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "配置数据"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/export [get]
func (h *FeatureHandler) ExportConfig(c *gin.Context) {
	data, err := h.featureService.ExportConfig()
	if err != nil {
		h.logger.Error("Failed to export feature config", zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to export configuration",
			"data": err.Error(),
		})
		return
	}
	
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=feature-config.json")
	c.Data(http.StatusOK, "application/json", data)
}

// ImportConfigRequest 导入配置请求
type ImportConfigRequest struct {
	Config string `json:"config" binding:"required"`
}

// ImportConfig 导入功能配置
// @Summary 导入功能配置
// @Description 从JSON数据导入功能开关配置
// @Tags features
// @Accept json
// @Produce json
// @Param request body ImportConfigRequest true "配置数据"
// @Success 200 {object} map[string]interface{} "导入成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/features/import [post]
func (h *FeatureHandler) ImportConfig(c *gin.Context) {
	var req ImportConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "Invalid request parameters",
			"data": err.Error(),
		})
		return
	}
	
	if err := h.featureService.ImportConfig([]byte(req.Config)); err != nil {
		h.logger.Error("Failed to import feature config", zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "Failed to import configuration",
			"data": err.Error(),
		})
		return
	}
	
	h.logger.Info("Feature configuration imported successfully")
	
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "Configuration imported successfully",
	})
}