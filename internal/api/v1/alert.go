package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/pkg/logger"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	log = logger.L
)

// AlertAPI 告警API
type AlertAPI struct {
	alertService *service.AlertService
}

// NewAlertAPI 创建告警API实例
func NewAlertAPI(alertService *service.AlertService) *AlertAPI {
	return &AlertAPI{
		alertService: alertService,
	}
}

// ListAlerts 获取告警列表
// @Summary 获取告警列表
// @Description 获取系统中的告警列表，支持分页和筛选
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "告警状态筛选"
// @Param severity query string false "严重程度筛选"
// @Param search query string false "搜索关键词"
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts [get]
func ListAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取筛选参数
	status := c.Query("status")
	severity := c.Query("severity")
	search := c.Query("search")

	// 构建查询
	query := database.DB.Model(&model.Alert{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if search != "" {
		query = query.Where("name LIKE ? OR title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取告警总数失败",
			"data": err.Error(),
		})
		return
	}

	// 获取分页数据
	var alerts []model.Alert
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取告警列表失败",
			"data": err.Error(),
		})
		return
	}

	// 转换为响应格式
	var items []*model.AlertResponse
	for i := range alerts {
		items = append(items, alerts[i].ToResponse())
	}

	// 返回分页数据
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"items": items,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}

// CreateAlert 创建告警（同步版本，保持向后兼容）
// @Summary 创建告警
// @Description 创建新的告警记录
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Alert true "告警信息"
// @Success 201 {object} response.Response{data=model.Alert}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts [post]
func CreateAlert(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 暂时跳过 Ollama 分析，保持原有行为
	alert.Analysis = ""
	alert.Status = "active"

	result := database.DB.Create(&alert)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建告警失败: " + result.Error.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// CreateAlertWithService 使用服务创建告警（新版本，支持异步处理）
func (api *AlertAPI) CreateAlertWithService(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 使用服务创建告警（会触发异步分析）
	if err := api.alertService.CreateAlert(c.Request.Context(), &alert); err != nil {
		logger.L.Error("Failed to create alert with service",
			zap.Error(err),
			zap.String("alert_name", alert.Name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建告警失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// GetAlert 获取单个告警
// @Summary 获取告警详情
// @Description 根据告警ID获取告警的详细信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Success 200 {object} response.Response{data=model.Alert}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/{id} [get]
func GetAlert(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	result := database.DB.First(&alert, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// GetAlertWithService 使用服务获取告警（支持缓存）
func (api *AlertAPI) GetAlertWithService(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	alert, err := api.alertService.GetAlert(c.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrAlertNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "告警不存在",
				"data": nil,
			})
			return
		}

		logger.L.Error("Failed to get alert with service",
			zap.Error(err),
			zap.Uint64("alert_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取告警失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": alert.ToResponse(),
	})
}

// UpdateAlert 更新告警
// @Summary 更新告警
// @Description 更新告警信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Param request body model.Alert true "告警信息"
// @Success 200 {object} response.Response{data=model.Alert}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/{id} [put]
func UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	result := database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(alert)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新告警失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}

// HandleAlert 处理告警
// @Summary 处理告警
// @Description 处理指定的告警，更新告警状态
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Param request body object{action=string,comment=string} true "处理操作"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/{id}/handle [post]
func HandleAlert(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Handler string `json:"handler" binding:"required"`
		Note    string `json:"note" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	now := time.Now()
	result := database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"handler":     req.Handler,
		"note":        req.Note,
		"status":      "handled",
		"handle_time": &now,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "处理告警失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": nil,
	})
}

// HandleAlertWithService 使用服务处理告警
func (api *AlertAPI) HandleAlertWithService(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	var req struct {
		Handler string `json:"handler" binding:"required"`
		Note    string `json:"note" binding:"required"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 默认状态为已处理
	if req.Status == "" {
		req.Status = model.AlertStatusResolved
	}

	// 更新告警状态
	if err := api.alertService.UpdateAlertStatus(c.Request.Context(), uint(id), req.Status, req.Handler, req.Note); err != nil {
		if err == service.ErrAlertNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "告警不存在",
				"data": nil,
			})
			return
		}

		logger.L.Error("Failed to handle alert with service",
			zap.Error(err),
			zap.Uint64("alert_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "处理告警失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"alert_id": id,
			"status":   req.Status,
			"handler":  req.Handler,
		},
	})
}

// AnalyzeAlert 分析告警
// @Summary 分析告警
// @Description 对指定告警进行AI分析
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/{id}/analyze [post]
func AnalyzeAlert(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 动态创建ollama服务实例以获取最新配置
	ollamaService := service.NewOllamaService()
	analysis, err := ollamaService.AnalyzeAlert(c.Request.Context(), &alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
	})
}

// FindSimilarAlerts 查找相似告警
// @Summary 查找相似告警
// @Description 查找与指定告警相似的其他告警
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Success 200 {object} response.Response{data=[]model.Alert}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/{id}/similar [get]
func FindSimilarAlerts(c *gin.Context) {
	var alert model.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 动态创建ollama服务实例以获取最新配置
	ollamaService := service.NewOllamaService()
	similarAlerts, err := ollamaService.FindSimilarAlerts(c.Request.Context(), &alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"similar_alerts": similarAlerts,
	})
}

// GetAlertStats 获取告警统计
// @Summary 获取告警统计
// @Description 获取告警数量统计信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alerts/stats [get]
func GetAlertStats(c *gin.Context) {
	var stats struct {
		Total        int64                  `json:"total"`
		Firing       int64                  `json:"firing"`
		Acknowledged int64                  `json:"acknowledged"`
		Resolved     int64                  `json:"resolved"`
		ByLevel      map[string]int64       `json:"by_level"`
		BySource     map[string]int64       `json:"by_source"`
	}

	// 初始化统计数据
	stats.ByLevel = make(map[string]int64)
	stats.BySource = make(map[string]int64)

	// 获取总数
	if err := database.DB.Model(&model.Alert{}).Count(&stats.Total).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}

	// 获取各状态统计
	if err := database.DB.Model(&model.Alert{}).Where("status = ?", "firing").Count(&stats.Firing).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}

	if err := database.DB.Model(&model.Alert{}).Where("status = ?", "acknowledged").Count(&stats.Acknowledged).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}

	if err := database.DB.Model(&model.Alert{}).Where("status = ?", "resolved").Count(&stats.Resolved).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}

	// 获取按级别统计
	var levelStats []struct {
		Level string `json:"level"`
		Count int64  `json:"count"`
	}
	if err := database.DB.Model(&model.Alert{}).Select("level, count(*) as count").Group("level").Scan(&levelStats).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}
	for _, stat := range levelStats {
		stats.ByLevel[stat.Level] = stat.Count
	}

	// 获取按来源统计
	var sourceStats []struct {
		Source string `json:"source"`
		Count  int64  `json:"count"`
	}
	if err := database.DB.Model(&model.Alert{}).Select("source, count(*) as count").Group("source").Scan(&sourceStats).Error; err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"获取告警统计失败","data":"%s"}`, err.Error())))
		return
	}
	for _, stat := range sourceStats {
		stats.BySource[stat.Source] = stat.Count
	}

	data, err := json.Marshal(gin.H{
		"code": 200,
		"msg":  "获取告警统计成功",
		"data": stats,
	})
	if err != nil {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusInternalServerError, "application/json; charset=utf-8", []byte(fmt.Sprintf(`{"code":500,"msg":"序列化响应失败","data":"%s"}`, err.Error())))
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}
