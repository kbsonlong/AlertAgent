package v1

import (
	"net/http"
	"strconv"

	"alert_agent/internal/model"
	"alert_agent/internal/pkg/database"
	"alert_agent/internal/service"

	"github.com/gin-gonic/gin"
)

// ListKnowledge 获取知识库列表
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
