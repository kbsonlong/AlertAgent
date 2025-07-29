package compatibility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APIVersionTestSuite API版本兼容性测试套件
type APIVersionTestSuite struct {
	suite.Suite
	serverV1 *httptest.Server
	serverV2 *httptest.Server
	client   *http.Client
	token    string
}

// SetupSuite 测试套件初始化
func (suite *APIVersionTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 创建V1 API服务器
	suite.serverV1 = httptest.NewServer(suite.setupV1Router())

	// 创建V2 API服务器
	suite.serverV2 = httptest.NewServer(suite.setupV2Router())

	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	suite.token = "mock-jwt-token"
}

// TearDownSuite 测试套件清理
func (suite *APIVersionTestSuite) TearDownSuite() {
	if suite.serverV1 != nil {
		suite.serverV1.Close()
	}
	if suite.serverV2 != nil {
		suite.serverV2.Close()
	}
}

// setupV1Router 设置V1 API路由
func (suite *APIVersionTestSuite) setupV1Router() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		// V1版本的分析API
		v1.POST("/analysis/submit", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// V1版本返回格式
			c.JSON(http.StatusOK, gin.H{
				"task_id": "task-v1-123",
				"status":  "submitted",
			})
		})

		v1.GET("/analysis/result/:task_id", func(c *gin.Context) {
			// V1版本返回格式
			c.JSON(http.StatusOK, gin.H{
				"task_id": c.Param("task_id"),
				"status":  "completed",
				"result": gin.H{
					"analysis": "V1 analysis result",
					"score":    85,
				},
			})
		})

		// V1版本的通道API
		v1.GET("/channels", func(c *gin.Context) {
			// V1版本返回格式（简化）
			c.JSON(http.StatusOK, gin.H{
				"channels": []gin.H{
					{
						"id":   "ch-1",
						"name": "Slack Channel",
						"type": "slack",
					},
				},
			})
		})

		v1.POST("/channels", func(c *gin.Context) {
			// V1版本创建通道
			c.JSON(http.StatusCreated, gin.H{
				"id":      "ch-v1-123",
				"name":    "New Channel",
				"created": true,
			})
		})
	}

	return router
}

// setupV2Router 设置V2 API路由
func (suite *APIVersionTestSuite) setupV2Router() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	v2 := router.Group("/api/v2")
	{
		// V2版本的分析API（增强版）
		v2.POST("/analysis/submit", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request",
					"code":    "INVALID_REQUEST",
				})
				return
			}

			// V2版本返回格式（标准化）
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data": gin.H{
					"task_id":    "task-v2-123",
					"status":     "queued",
					"priority":   req["priority"],
					"created_at": time.Now().Format(time.RFC3339),
					"estimated_completion": time.Now().Add(5 * time.Minute).Format(time.RFC3339),
				},
				"meta": gin.H{
					"version": "2.0",
					"api_version": "v2",
				},
			})
		})

		v2.GET("/analysis/result/:task_id", func(c *gin.Context) {
			// V2版本返回格式（增强）
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data": gin.H{
					"task_id": c.Param("task_id"),
					"status":  "completed",
					"result": gin.H{
						"analysis":   "V2 enhanced analysis result",
						"confidence": 0.95,
						"score":      92,
						"details": gin.H{
							"root_cause":    "Database connection timeout",
							"recommendations": []string{"Increase connection pool", "Add retry logic"},
							"severity":      "high",
						},
					},
					"completed_at": time.Now().Format(time.RFC3339),
					"duration_ms":  1500,
				},
				"meta": gin.H{
					"version": "2.0",
					"api_version": "v2",
				},
			})
		})

		// V2版本的通道API（标准化）
		v2.GET("/channels", func(c *gin.Context) {
			// V2版本返回格式（标准化分页）
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data": gin.H{
					"items": []gin.H{
						{
							"id":         "ch-1",
							"name":       "Slack Channel",
							"type":       "slack",
							"status":     "active",
							"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
							"updated_at": time.Now().Format(time.RFC3339),
						},
					},
					"total":      1,
					"page":       1,
					"page_size":  10,
					"total_pages": 1,
				},
				"meta": gin.H{
					"version": "2.0",
					"api_version": "v2",
				},
			})
		})

		v2.POST("/channels", func(c *gin.Context) {
			// V2版本创建通道（标准化）
			c.JSON(http.StatusCreated, gin.H{
				"status": "success",
				"data": gin.H{
					"id":         "ch-v2-123",
					"name":       "New Channel V2",
					"type":       "slack",
					"status":     "active",
					"created_at": time.Now().Format(time.RFC3339),
				},
				"meta": gin.H{
					"version": "2.0",
					"api_version": "v2",
				},
			})
		})
	}

	return router
}

// makeRequest 发送HTTP请求的辅助方法
func (suite *APIVersionTestSuite) makeRequest(server *httptest.Server, method, path string, data interface{}) *http.Response {
	var body *bytes.Buffer
	if data != nil {
		jsonData, err := json.Marshal(data)
		suite.Require().NoError(err)
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, server.URL+path, body)
	suite.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)

	return resp
}

// TestV1AnalysisAPI 测试V1分析API
func (suite *APIVersionTestSuite) TestV1AnalysisAPI() {
	// 提交分析任务
	submitReq := map[string]interface{}{
		"alert_id": 12345,
		"type":     "root_cause",
	}

	resp := suite.makeRequest(suite.serverV1, "POST", "/api/v1/analysis/submit", submitReq)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V1响应格式
	assert.Equal(suite.T(), "task-v1-123", result["task_id"])
	assert.Equal(suite.T(), "submitted", result["status"])
	assert.NotContains(suite.T(), result, "meta") // V1没有meta字段

	// 获取分析结果
	resp = suite.makeRequest(suite.serverV1, "GET", "/api/v1/analysis/result/task-v1-123", nil)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V1结果格式
	assert.Equal(suite.T(), "task-v1-123", result["task_id"])
	assert.Equal(suite.T(), "completed", result["status"])
	assert.Contains(suite.T(), result, "result")
	resultData := result["result"].(map[string]interface{})
	assert.Equal(suite.T(), "V1 analysis result", resultData["analysis"])
	assert.Equal(suite.T(), float64(85), resultData["score"])
}

// TestV2AnalysisAPI 测试V2分析API
func (suite *APIVersionTestSuite) TestV2AnalysisAPI() {
	// 提交分析任务
	submitReq := map[string]interface{}{
		"alert_id": 12345,
		"type":     "root_cause",
		"priority": 8,
	}

	resp := suite.makeRequest(suite.serverV2, "POST", "/api/v2/analysis/submit", submitReq)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V2响应格式（标准化）
	assert.Equal(suite.T(), "success", result["status"])
	assert.Contains(suite.T(), result, "data")
	assert.Contains(suite.T(), result, "meta")

	data := result["data"].(map[string]interface{})
	assert.Equal(suite.T(), "task-v2-123", data["task_id"])
	assert.Equal(suite.T(), "queued", data["status"])
	assert.Equal(suite.T(), float64(8), data["priority"])
	assert.Contains(suite.T(), data, "created_at")
	assert.Contains(suite.T(), data, "estimated_completion")

	meta := result["meta"].(map[string]interface{})
	assert.Equal(suite.T(), "2.0", meta["version"])
	assert.Equal(suite.T(), "v2", meta["api_version"])

	// 获取分析结果
	resp = suite.makeRequest(suite.serverV2, "GET", "/api/v2/analysis/result/task-v2-123", nil)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V2结果格式（增强）
	assert.Equal(suite.T(), "success", result["status"])
	data = result["data"].(map[string]interface{})
	assert.Equal(suite.T(), "task-v2-123", data["task_id"])
	assert.Equal(suite.T(), "completed", data["status"])
	assert.Contains(suite.T(), data, "completed_at")
	assert.Contains(suite.T(), data, "duration_ms")

	resultData := data["result"].(map[string]interface{})
	assert.Equal(suite.T(), "V2 enhanced analysis result", resultData["analysis"])
	assert.Equal(suite.T(), 0.95, resultData["confidence"])
	assert.Equal(suite.T(), float64(92), resultData["score"])
	assert.Contains(suite.T(), resultData, "details")

	details := resultData["details"].(map[string]interface{})
	assert.Equal(suite.T(), "Database connection timeout", details["root_cause"])
	assert.Contains(suite.T(), details, "recommendations")
	assert.Equal(suite.T(), "high", details["severity"])
}

// TestV1ChannelAPI 测试V1通道API
func (suite *APIVersionTestSuite) TestV1ChannelAPI() {
	// 列出通道
	resp := suite.makeRequest(suite.serverV1, "GET", "/api/v1/channels", nil)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V1响应格式（简化）
	assert.Contains(suite.T(), result, "channels")
	assert.NotContains(suite.T(), result, "status") // V1没有标准化状态字段
	assert.NotContains(suite.T(), result, "meta")   // V1没有meta字段

	channels := result["channels"].([]interface{})
	assert.Len(suite.T(), channels, 1)

	channel := channels[0].(map[string]interface{})
	assert.Equal(suite.T(), "ch-1", channel["id"])
	assert.Equal(suite.T(), "Slack Channel", channel["name"])
	assert.Equal(suite.T(), "slack", channel["type"])
	assert.NotContains(suite.T(), channel, "status")     // V1没有状态字段
	assert.NotContains(suite.T(), channel, "created_at") // V1没有时间戳

	// 创建通道
	createReq := map[string]interface{}{
		"name": "Test Channel",
		"type": "slack",
	}

	resp = suite.makeRequest(suite.serverV1, "POST", "/api/v1/channels", createReq)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V1创建响应格式
	assert.Equal(suite.T(), "ch-v1-123", result["id"])
	assert.Equal(suite.T(), "New Channel", result["name"])
	assert.Equal(suite.T(), true, result["created"])
	assert.NotContains(suite.T(), result, "status") // V1没有标准化字段
}

// TestV2ChannelAPI 测试V2通道API
func (suite *APIVersionTestSuite) TestV2ChannelAPI() {
	// 列出通道
	resp := suite.makeRequest(suite.serverV2, "GET", "/api/v2/channels", nil)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V2响应格式（标准化分页）
	assert.Equal(suite.T(), "success", result["status"])
	assert.Contains(suite.T(), result, "data")
	assert.Contains(suite.T(), result, "meta")

	data := result["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "items")
	assert.Equal(suite.T(), float64(1), data["total"])
	assert.Equal(suite.T(), float64(1), data["page"])
	assert.Equal(suite.T(), float64(10), data["page_size"])
	assert.Equal(suite.T(), float64(1), data["total_pages"])

	items := data["items"].([]interface{})
	assert.Len(suite.T(), items, 1)

	channel := items[0].(map[string]interface{})
	assert.Equal(suite.T(), "ch-1", channel["id"])
	assert.Equal(suite.T(), "Slack Channel", channel["name"])
	assert.Equal(suite.T(), "slack", channel["type"])
	assert.Equal(suite.T(), "active", channel["status"])
	assert.Contains(suite.T(), channel, "created_at")
	assert.Contains(suite.T(), channel, "updated_at")

	meta := result["meta"].(map[string]interface{})
	assert.Equal(suite.T(), "2.0", meta["version"])
	assert.Equal(suite.T(), "v2", meta["api_version"])

	// 创建通道
	createReq := map[string]interface{}{
		"name": "Test Channel V2",
		"type": "slack",
	}

	resp = suite.makeRequest(suite.serverV2, "POST", "/api/v2/channels", createReq)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	// 验证V2创建响应格式（标准化）
	assert.Equal(suite.T(), "success", result["status"])
	assert.Contains(suite.T(), result, "data")
	assert.Contains(suite.T(), result, "meta")

	data = result["data"].(map[string]interface{})
	assert.Equal(suite.T(), "ch-v2-123", data["id"])
	assert.Equal(suite.T(), "New Channel V2", data["name"])
	assert.Equal(suite.T(), "slack", data["type"])
	assert.Equal(suite.T(), "active", data["status"])
	assert.Contains(suite.T(), data, "created_at")
}

// TestBackwardCompatibility 测试向后兼容性
func (suite *APIVersionTestSuite) TestBackwardCompatibility() {
	// 确保V1 API仍然可以正常工作
	submitReq := map[string]interface{}{
		"alert_id": 12345,
		"type":     "root_cause",
	}

	// V1 API调用
	resp := suite.makeRequest(suite.serverV1, "POST", "/api/v1/analysis/submit", submitReq)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// V2 API调用（相同功能）
	resp = suite.makeRequest(suite.serverV2, "POST", "/api/v2/analysis/submit", submitReq)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// 两个版本都应该能够处理相同的输入
	// 但返回格式不同
}

// TestAPIVersionHeaders 测试API版本头
func (suite *APIVersionTestSuite) TestAPIVersionHeaders() {
	// 测试V1 API响应头
	resp := suite.makeRequest(suite.serverV1, "GET", "/api/v1/channels", nil)
	defer resp.Body.Close()

	// V1可能没有版本头（向后兼容）
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// 测试V2 API响应头
	resp = suite.makeRequest(suite.serverV2, "GET", "/api/v2/channels", nil)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	// V2应该在响应体中包含版本信息
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)

	meta := result["meta"].(map[string]interface{})
	assert.Equal(suite.T(), "v2", meta["api_version"])
}

// TestErrorFormatCompatibility 测试错误格式兼容性
func (suite *APIVersionTestSuite) TestErrorFormatCompatibility() {
	// V1错误格式
	resp := suite.makeRequest(suite.serverV1, "POST", "/api/v1/analysis/submit", "invalid json")
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var v1Error map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&v1Error)
	suite.Require().NoError(err)

	// V1简单错误格式
	assert.Contains(suite.T(), v1Error, "error")
	assert.NotContains(suite.T(), v1Error, "status")
	assert.NotContains(suite.T(), v1Error, "code")

	// V2错误格式
	resp = suite.makeRequest(suite.serverV2, "POST", "/api/v2/analysis/submit", "invalid json")
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var v2Error map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&v2Error)
	suite.Require().NoError(err)

	// V2标准化错误格式
	assert.Equal(suite.T(), "error", v2Error["status"])
	assert.Contains(suite.T(), v2Error, "message")
	assert.Contains(suite.T(), v2Error, "code")
	assert.Equal(suite.T(), "INVALID_REQUEST", v2Error["code"])
}

// TestAPIVersionTestSuite 运行API版本兼容性测试套件
func TestAPIVersionTestSuite(t *testing.T) {
	if os.Getenv("COMPATIBILITY_TEST") == "" {
		t.Skip("Skipping compatibility tests. Set COMPATIBILITY_TEST=1 to run.")
	}

	suite.Run(t, new(APIVersionTestSuite))
}

// BenchmarkAPIVersionPerformance API版本性能对比基准测试
func BenchmarkAPIVersionPerformance(b *testing.B) {
	if os.Getenv("COMPATIBILITY_TEST") == "" {
		b.Skip("Skipping compatibility benchmarks. Set COMPATIBILITY_TEST=1 to run.")
	}

	gin.SetMode(gin.TestMode)
	testSuite := &APIVersionTestSuite{}
	testSuite.SetupSuite()
	defer testSuite.TearDownSuite()

	submitReq := map[string]interface{}{
		"alert_id": 12345,
		"type":     "root_cause",
	}

	b.Run("V1_API", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp := testSuite.makeRequest(testSuite.serverV1, "POST", "/api/v1/analysis/submit", submitReq)
			resp.Body.Close()
		}
	})

	b.Run("V2_API", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp := testSuite.makeRequest(testSuite.serverV2, "POST", "/api/v2/analysis/submit", submitReq)
			resp.Body.Close()
		}
	})
}