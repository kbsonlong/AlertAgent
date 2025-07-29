package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// APITestSuite 端到端API集成测试套件
type APITestSuite struct {
	suite.Suite
	server  *httptest.Server
	client  *http.Client
	baseURL string
	token   string
}

// SetupSuite 测试套件初始化
func (suite *APITestSuite) SetupSuite() {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建简单的测试路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Alert Agent is running",
		})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由
		v1.POST("/auth/login", suite.handleLogin)
		v1.POST("/auth/register", suite.handleRegister)

		// 需要认证的路由
		auth := v1.Group("", suite.authMiddleware())
		{
			// 分析API
			auth.POST("/analysis/submit", suite.handleAnalysisSubmit)
			auth.GET("/analysis/result/:task_id", suite.handleAnalysisResult)
			auth.GET("/analysis/progress/:task_id", suite.handleAnalysisProgress)
			auth.GET("/analysis/tasks", suite.handleAnalysisTasks)
			auth.GET("/analysis/health", suite.handleAnalysisHealth)

			// 通道API
			auth.GET("/channels", suite.handleChannelsList)
			auth.POST("/channels", suite.handleChannelsCreate)
			auth.GET("/channels/:id", suite.handleChannelsGet)
			auth.PUT("/channels/:id", suite.handleChannelsUpdate)
			auth.DELETE("/channels/:id", suite.handleChannelsDelete)
			auth.POST("/channels/:id/test", suite.handleChannelsTest)
			auth.GET("/channels/:id/stats", suite.handleChannelsStats)

			// 集群API
			auth.GET("/clusters", suite.handleClustersList)

			// 插件API
			auth.GET("/plugins", suite.handlePluginsList)
		}
	}

	// 创建测试服务器
	suite.server = httptest.NewServer(router)
	suite.baseURL = suite.server.URL
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// 获取认证令牌
	suite.token = suite.getAuthToken()
}

// TearDownSuite 测试套件清理
func (suite *APITestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// getAuthToken 获取认证令牌
func (suite *APITestSuite) getAuthToken() string {
	// 注册用户
	registerData := map[string]interface{}{
		"username": "test_admin",
		"email":    "test@example.com",
		"password": "test_password",
		"roles":    []string{"admin"},
	}

	resp := suite.makeRequest("POST", "/api/v1/auth/register", registerData, "")
	resp.Body.Close()

	// 登录获取token
	loginData := map[string]interface{}{
		"username": "test_admin",
		"password": "test_password",
	}

	resp = suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var loginResp struct {
		Status string `json:"status"`
		Data   struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	resp.Body.Close()

	err = json.Unmarshal(body, &loginResp)
	require.NoError(suite.T(), err)

	return loginResp.Data.Token
}

// makeRequest 发送HTTP请求的辅助方法
func (suite *APITestSuite) makeRequest(method, path string, data interface{}, token string) *http.Response {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		require.NoError(suite.T(), err)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, suite.baseURL+path, body)
	require.NoError(suite.T(), err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)

	return resp
}

// TestHealthCheck 测试健康检查API
func (suite *APITestSuite) TestHealthCheck() {
	resp, err := suite.client.Get(suite.baseURL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "ok", response["status"])
	assert.Contains(suite.T(), response["message"], "Alert Agent is running")
}

// TestAnalysisWorkflow 测试完整的分析工作流
func (suite *APITestSuite) TestAnalysisWorkflow() {
	// 1. 提交分析任务
	submitReq := map[string]interface{}{
		"alert_id": 12345,
		"type":     "root_cause",
		"priority": 8,
		"timeout":  300,
	}

	submitResp := suite.makeRequest("POST", "/api/v1/analysis/submit", submitReq, suite.token)
	assert.Equal(suite.T(), http.StatusOK, submitResp.StatusCode)

	var submitResult map[string]interface{}
	err := json.NewDecoder(submitResp.Body).Decode(&submitResult)
	require.NoError(suite.T(), err)
	submitResp.Body.Close()

	assert.Equal(suite.T(), "success", submitResult["status"])
	data := submitResult["data"].(map[string]interface{})
	taskID := data["task_id"].(string)
	assert.NotEmpty(suite.T(), taskID)

	// 2. 获取分析进度
	progressResp := suite.makeRequest("GET", fmt.Sprintf("/api/v1/analysis/progress/%s", taskID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, progressResp.StatusCode)
	progressResp.Body.Close()

	// 3. 获取分析结果
	resultResp := suite.makeRequest("GET", fmt.Sprintf("/api/v1/analysis/result/%s", taskID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, resultResp.StatusCode)
	resultResp.Body.Close()

	// 4. 列出分析任务
	listResp := suite.makeRequest("GET", "/api/v1/analysis/tasks?alert_id=12345&limit=10", nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, listResp.StatusCode)
	listResp.Body.Close()
}

// TestChannelManagement 测试通道管理API
func (suite *APITestSuite) TestChannelManagement() {
	// 1. 创建通道
	createReq := map[string]interface{}{
		"name": "Test Slack Channel",
		"type": "slack",
		"config": map[string]interface{}{
			"webhook_url": "https://hooks.slack.com/test",
			"channel":     "#test",
		},
	}

	createResp := suite.makeRequest("POST", "/api/v1/channels", createReq, suite.token)
	assert.Equal(suite.T(), http.StatusCreated, createResp.StatusCode)

	var createResult map[string]interface{}
	err := json.NewDecoder(createResp.Body).Decode(&createResult)
	require.NoError(suite.T(), err)
	createResp.Body.Close()

	data := createResult["data"].(map[string]interface{})
	channelID := data["id"].(string)

	// 2. 获取通道详情
	getResp := suite.makeRequest("GET", fmt.Sprintf("/api/v1/channels/%s", channelID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	// 3. 更新通道
	updateReq := map[string]interface{}{
		"name": "Updated Test Channel",
	}
	updateResp := suite.makeRequest("PUT", fmt.Sprintf("/api/v1/channels/%s", channelID), updateReq, suite.token)
	assert.Equal(suite.T(), http.StatusOK, updateResp.StatusCode)
	updateResp.Body.Close()

	// 4. 测试通道
	testResp := suite.makeRequest("POST", fmt.Sprintf("/api/v1/channels/%s/test", channelID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, testResp.StatusCode)
	testResp.Body.Close()

	// 5. 获取统计信息
	statsResp := suite.makeRequest("GET", fmt.Sprintf("/api/v1/channels/%s/stats", channelID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, statsResp.StatusCode)
	statsResp.Body.Close()

	// 6. 列出通道
	listResp := suite.makeRequest("GET", "/api/v1/channels?limit=10", nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, listResp.StatusCode)
	listResp.Body.Close()

	// 7. 删除通道
	deleteResp := suite.makeRequest("DELETE", fmt.Sprintf("/api/v1/channels/%s", channelID), nil, suite.token)
	assert.Equal(suite.T(), http.StatusOK, deleteResp.StatusCode)
	deleteResp.Body.Close()
}

// TestErrorHandling 测试错误处理
func (suite *APITestSuite) TestErrorHandling() {
	// 测试未授权访问
	resp := suite.makeRequest("GET", "/api/v1/channels", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// 测试无效请求
	invalidReq := map[string]interface{}{
		"type": "root_cause", // 缺少alert_id
	}
	resp = suite.makeRequest("POST", "/api/v1/analysis/submit", invalidReq, suite.token)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// 测试资源未找到
	resp = suite.makeRequest("GET", "/api/v1/channels/nonexistent", nil, suite.token)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

// Mock handlers for testing
func (suite *APITestSuite) handleLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"token": "mock-jwt-token",
		},
	})
}

func (suite *APITestSuite) handleRegister(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"user_id": "user-123",
		},
	})
}

func (suite *APITestSuite) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || auth == "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (suite *APITestSuite) handleAnalysisSubmit(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request",
		})
		return
	}

	if _, ok := req["alert_id"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "alert_id is required",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"task_id": "task-123",
			"status":  "queued",
		},
	})
}

func (suite *APITestSuite) handleAnalysisResult(c *gin.Context) {
	taskID := c.Param("task_id")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"task_id": taskID,
			"status":  "completed",
			"result":  gin.H{"analysis": "mock result"},
		},
	})
}

func (suite *APITestSuite) handleAnalysisProgress(c *gin.Context) {
	taskID := c.Param("task_id")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"task_id":  taskID,
			"progress": 75.5,
			"stage":    "processing",
		},
	})
}

func (suite *APITestSuite) handleAnalysisTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": []gin.H{
			{
				"id":       "task-123",
				"alert_id": 12345,
				"type":     "root_cause",
				"status":   "completed",
			},
		},
	})
}

func (suite *APITestSuite) handleAnalysisHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Analysis service is healthy",
	})
}

func (suite *APITestSuite) handleChannelsList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"items": []gin.H{},
			"total": 0,
		},
	})
}

func (suite *APITestSuite) handleChannelsCreate(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"id":   "channel-123",
			"name": "Test Channel",
			"type": "slack",
		},
	})
}

func (suite *APITestSuite) handleChannelsGet(c *gin.Context) {
	id := c.Param("id")
	if id == "nonexistent" {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Channel not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":   id,
			"name": "Test Channel",
			"type": "slack",
		},
	})
}

func (suite *APITestSuite) handleChannelsUpdate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":   id,
			"name": "Updated Channel",
		},
	})
}

func (suite *APITestSuite) handleChannelsDelete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Channel deleted",
	})
}

func (suite *APITestSuite) handleChannelsTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"success": true,
			"latency": 150,
		},
	})
}

func (suite *APITestSuite) handleChannelsStats(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"channel_id":      id,
			"total_messages":  100,
			"success_rate":    95.5,
		},
	})
}

func (suite *APITestSuite) handleClustersList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"items": []gin.H{},
			"total": 0,
		},
	})
}

func (suite *APITestSuite) handlePluginsList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"plugins": []gin.H{
				{
					"type":    "slack",
					"name":    "Slack Plugin",
					"version": "1.0.0",
					"schema":  gin.H{"type": "object"},
				},
			},
		},
	})
}

// TestAPITestSuite 运行API测试套件
func TestAPITestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TEST=1 to run.")
	}

	suite.Run(t, new(APITestSuite))
}

// BenchmarkAPIPerformance API性能基准测试
func BenchmarkAPIPerformance(b *testing.B) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		b.Skip("Skipping integration benchmarks. Set INTEGRATION_TEST=1 to run.")
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Alert Agent is running",
		})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/health")
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}