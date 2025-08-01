package v1

import (
	"net/http"
	"strconv"
	"strings"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// ListKnowledge 获取知识库列表
// @Summary 获取知识库列表
// @Description 获取知识库列表，支持关键词搜索、分类筛选和分页
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词搜索"
// @Param category query string false "分类筛选"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=object{list=[]model.Knowledge,total=int64,page=int,pageSize=int}}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge [get]
func ListKnowledge(c *gin.Context) {
	// 获取查询参数
	keyword := c.Query("keyword")
	category := c.Query("category")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	// 转换分页参数
	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)
	offset := (pageNum - 1) * pageSizeNum

	// 构建查询
	query := database.DB.Model(&model.Knowledge{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取知识库总数失败",
			"data": nil,
		})
		return
	}

	// 获取数据
	var knowledge []model.Knowledge
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSizeNum).Find(&knowledge)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取知识库列表失败",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":     knowledge,
			"total":    total,
			"page":     pageNum,
			"pageSize": pageSizeNum,
		},
	})
}

// GetKnowledge 获取单个知识库记录
// @Summary 获取单个知识库记录
// @Description 根据ID获取单个知识库记录的详细信息
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "知识库记录ID"
// @Success 200 {object} response.Response{data=model.Knowledge}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/{id} [get]
func GetKnowledge(c *gin.Context) {
	id := c.Param("id")
	knowledgeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的知识库ID",
			"data": err.Error(),
		})
		return
	}

	var knowledge model.Knowledge
	result := database.DB.First(&knowledge, knowledgeID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "知识库记录不存在",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": knowledge,
	})
}

// CreateKnowledge 创建知识库记录
// @Summary 创建知识库记录
// @Description 创建新的知识库记录
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Knowledge true "知识库记录信息"
// @Success 200 {object} response.Response{data=model.Knowledge}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge [post]
func CreateKnowledge(c *gin.Context) {
	var knowledge model.Knowledge
	if err := c.ShouldBindJSON(&knowledge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": err.Error(),
		})
		return
	}

	result := database.DB.Create(&knowledge)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建知识库记录失败",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": knowledge,
	})
}

// UpdateKnowledge 更新知识库记录
// @Summary 更新知识库记录
// @Description 根据ID更新知识库记录信息
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "知识库记录ID"
// @Param request body model.Knowledge true "知识库记录信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/{id} [put]
func UpdateKnowledge(c *gin.Context) {
	id := c.Param("id")
	knowledgeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的知识库ID",
			"data": nil,
		})
		return
	}

	var knowledge model.Knowledge
	if err := c.ShouldBindJSON(&knowledge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的请求参数",
			"data": nil,
		})
		return
	}

	result := database.DB.Model(&model.Knowledge{}).Where("id = ?", knowledgeID).Updates(knowledge)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新知识库记录失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "知识库记录不存在",
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

// DeleteKnowledge 删除知识库记录
// @Summary 删除知识库记录
// @Description 根据ID删除指定的知识库记录
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "知识库记录ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/{id} [delete]
func DeleteKnowledge(c *gin.Context) {
	id := c.Param("id")
	knowledgeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的知识库ID",
			"data": nil,
		})
		return
	}

	result := database.DB.Delete(&model.Knowledge{}, knowledgeID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除知识库记录失败",
			"data": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "知识库记录不存在",
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

// ConvertAlertToKnowledge 将告警转换为知识库记录
// @Summary 将告警转换为知识库记录
// @Description 根据告警ID将告警信息转换为知识库记录
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警ID"
// @Success 200 {object} response.Response{data=model.Knowledge}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/convert/{id} [post]
func ConvertAlertToKnowledge(c *gin.Context) {
	id := c.Param("id")
	alertID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的告警ID",
			"data": nil,
		})
		return
	}

	// 获取告警信息
	var alert model.Alert
	if err := database.DB.First(&alert, alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "告警不存在",
			"data": nil,
		})
		return
	}

	// 创建知识库记录
	knowledge, err := service.CreateKnowledge(&alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "转换知识库记录失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": knowledge,
	})
}

// GetKnowledgeCategories 获取知识库分类列表
// @Summary 获取知识库分类列表
// @Description 获取所有知识库记录的分类列表
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]string}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/categories [get]
func GetKnowledgeCategories(c *gin.Context) {
	var categories []string
	result := database.DB.Model(&model.Knowledge{}).Distinct("category").Pluck("category", &categories)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取分类列表失败",
			"data": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": categories,
	})
}

// GetKnowledgeTags 获取知识库标签列表
// @Summary 获取知识库标签列表
// @Description 获取所有知识库记录的标签列表，自动去重
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]string}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/knowledge/tags [get]
func GetKnowledgeTags(c *gin.Context) {
	// 获取所有标签字符串
	var tagStrings []string
	result := database.DB.Model(&model.Knowledge{}).Where("tags != '' AND tags IS NOT NULL").Pluck("tags", &tagStrings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取标签列表失败",
			"data": result.Error.Error(),
		})
		return
	}

	// 解析标签字符串，去重
	tagSet := make(map[string]bool)
	for _, tagString := range tagStrings {
		if tagString != "" {
			// 按逗号分割标签
			tags := strings.Split(tagString, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagSet[tag] = true
				}
			}
		}
	}

	// 转换为数组
	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": tags,
	})
}
