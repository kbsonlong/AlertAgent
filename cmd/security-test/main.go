package main

import (
	"context"
	"fmt"
	"log"

	"alert_agent/internal/infrastructure/config"
	"alert_agent/internal/infrastructure/database"
	"alert_agent/internal/security/di"
	"alert_agent/internal/security/user"
)

func main() {
	fmt.Println("=== AlertAgent 安全框架集成测试 ===")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 执行数据库迁移
	fmt.Println("执行数据库迁移...")
	if err := database.Migrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("✓ 数据库迁移完成")

	// 初始化安全容器
	fmt.Println("初始化安全框架...")
	securityContainer, err := di.NewContainer()
	if err != nil {
		log.Fatalf("安全容器初始化失败: %v", err)
	}
	fmt.Println("✓ 安全框架初始化完成")

	// 测试用户服务
	fmt.Println("\n=== 测试用户管理功能 ===")
	testUserService(securityContainer)

	// 测试JWT功能
	fmt.Println("\n=== 测试JWT功能 ===")
	testJWTService(securityContainer)

	fmt.Println("\n=== 所有测试完成 ===")
}

func testUserService(container *di.Container) {
	userService := container.GetUserService()

	// 创建测试用户
	req := &user.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Roles:    []string{"user"},
	}

	createdUser, err := userService.CreateUser(context.Background(), req, "system")
	if err != nil {
		fmt.Printf("✗ 创建用户失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 创建用户成功: %s (ID: %s)\n", createdUser.Username, createdUser.ID)

	// 测试登录
	loginReq := &user.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	loginResp, err := userService.Login(context.Background(), loginReq, "127.0.0.1", "test-agent")
	if err != nil {
		fmt.Printf("✗ 用户登录失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 用户登录成功, Token: %s...\n", loginResp.AccessToken[:20])
}



func testJWTService(container *di.Container) {
	jwtManager := container.GetJWTManager()

	// 生成JWT令牌
	userID := "test-user-123"
	username := "testuser"
	roles := []string{"user"}
	permissions := []string{"read:profile"}

	token, err := jwtManager.GenerateToken(userID, username, roles, permissions)
	if err != nil {
		fmt.Printf("✗ 生成JWT令牌失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 生成JWT令牌成功: %s...\n", token[:30])

	// 验证JWT令牌
	claims, err := jwtManager.VerifyToken(token)
	if err != nil {
		fmt.Printf("✗ 验证JWT令牌失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 验证JWT令牌成功, 用户ID: %s\n", claims.UserID)
}