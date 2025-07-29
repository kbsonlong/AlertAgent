# AlertAgent 安全框架

这是一个基于 Go 语言的企业级安全框架，提供了完整的认证、授权、审计和安全防护功能。

## 功能特性

### 🔐 认证 (Authentication)
- JWT Token 认证
- 密码强度验证
- 登录失败锁定
- Token 刷新机制

### 🛡️ 授权 (Authorization)
- 基于角色的访问控制 (RBAC)
- 细粒度权限管理
- 动态权限检查
- 多角色支持

### 📝 审计 (Audit)
- 完整的操作日志记录
- 安全事件追踪
- 结构化日志输出
- 可配置的日志级别

### 🔒 加密 (Encryption)
- AES-GCM 对称加密
- PBKDF2/Scrypt 密钥派生
- 安全的密码哈希
- 配置文件加密

### ✅ 输入验证 (Validation)
- SQL 注入防护
- XSS 攻击防护
- 输入数据清理
- 自定义验证规则

### 🚦 安全中间件
- 安全头设置
- 请求限流
- 输入验证
- 审计日志

## 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Handler  │───▶│   Middleware    │───▶│   Use Case      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Security      │    │   Validation    │    │   Repository    │
│   Components    │    │   & Audit       │    │   Layer         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 快速开始

### 1. 初始化安全容器

```go
package main

import (
    "alert_agent/internal/security/di"
    "alert_agent/internal/security/routes"
    "github.com/gin-gonic/gin"
)

func main() {
    // 初始化依赖注入容器
    container, err := di.NewContainer()
    if err != nil {
        log.Fatalf("Failed to initialize container: %v", err)
    }
    defer container.Cleanup()

    // 创建路由
    router := gin.New()
    
    // 设置安全中间件
    routes.SetupSecurityMiddleware(router)
    
    // 设置认证路由
    routes.SetupAuthRoutes(
        router,
        container.GetAuthHandler(),
        container.GetMiddlewareConfig(),
    )
    
    // 启动服务
    router.Run(":8080")
}
```

### 2. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "SecurePass123!",
    "roles": ["admin"]
  }'
```

### 3. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecurePass123!"
  }'
```

### 4. 访问受保护的资源

```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 配置说明

### 安全配置结构

```go
type SecurityConfig struct {
    JWT        JWTConfig        `yaml:"jwt"`
    Encryption EncryptionConfig `yaml:"encryption"`
    Audit      AuditConfig      `yaml:"audit"`
    RateLimit  RateLimitConfig  `yaml:"rate_limit"`
    Auth       AuthConfig       `yaml:"auth"`
    Session    SessionConfig    `yaml:"session"`
}
```

### JWT 配置

```yaml
jwt:
  secret: "your-secret-key"
  expiration: 24h
  refresh_expiration: 168h  # 7 days
  issuer: "AlertAgent"
  audience: "AlertAgent-Users"
```

### 加密配置

```yaml
encryption:
  key: "your-encryption-key"
  salt: "your-salt"
  iterations: 100000
  key_length: 32
```

### 审计配置

```yaml
audit:
  enabled: true
  log_level: "info"
  log_file: "/var/log/alertagent/audit.log"
  max_size: 100    # MB
  max_backups: 10
  max_age: 30      # days
  compress: true
```

## 权限系统

### 预定义权限

- `user:read` - 读取用户信息
- `user:write` - 创建和更新用户
- `user:delete` - 删除用户
- `system:admin` - 系统管理
- `system:config` - 系统配置
- `system:audit` - 审计查看

### 预定义角色

- **user**: 普通用户，拥有基本的读取权限
- **moderator**: 内容管理员，拥有用户管理权限
- **admin**: 系统管理员，拥有所有权限

### 自定义权限检查

```go
// 在中间件中检查权限
func RequirePermission(permission rbac.Permission) gin.HandlerFunc {
    return middleware.PermissionMiddleware(securityConfig, permission)
}

// 在处理器中检查权限
func (h *Handler) SomeProtectedAction(c *gin.Context) {
    userID := c.GetString("user_id")
    if !h.rbacManager.HasPermission(userID, rbac.PermissionUserWrite) {
        c.JSON(403, gin.H{"error": "Permission denied"})
        return
    }
    // 执行操作...
}
```

## 安全最佳实践

### 1. 密码安全
- 使用强密码策略
- 密码哈希使用 PBKDF2 或 Scrypt
- 每个密码使用唯一的盐值

### 2. Token 安全
- JWT Token 设置合理的过期时间
- 使用 HTTPS 传输 Token
- 实现 Token 刷新机制

### 3. 输入验证
- 所有用户输入都进行验证
- 防止 SQL 注入和 XSS 攻击
- 使用白名单验证

### 4. 审计日志
- 记录所有安全相关操作
- 包含足够的上下文信息
- 定期备份和分析日志

### 5. 权限控制
- 遵循最小权限原则
- 定期审查用户权限
- 实现权限分离

## API 端点

### 公开端点
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /health/*` - 健康检查

### 认证端点 (需要 JWT Token)
- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/auth/profile` - 获取用户信息
- `PUT /api/v1/auth/profile` - 更新用户信息
- `POST /api/v1/auth/change-password` - 修改密码

### 管理员端点 (需要 admin 角色)
- `GET /api/v1/admin/users` - 获取用户列表
- `GET /api/v1/admin/users/:id` - 获取指定用户
- `PUT /api/v1/admin/users/:id` - 更新指定用户
- `DELETE /api/v1/admin/users/:id` - 删除用户

## 错误处理

框架提供统一的错误响应格式：

```json
{
  "code": 400,
  "message": "Validation failed",
  "error": "Password must be at least 8 characters"
}
```

## 测试

运行安全框架演示：

```bash
go run cmd/security-demo/main.go
```

访问 http://localhost:8080/health/ping 验证服务是否正常运行。

## 扩展开发

### 添加新的权限

```go
const (
    PermissionCustomAction Permission = "custom:action"
)
```

### 添加新的验证规则

```go
func (v *Validator) CustomValidation(value string) {
    if !isValid(value) {
        v.AddError("custom", "Custom validation failed")
    }
}
```

### 自定义审计事件

```go
func (al *AuditLogger) LogCustomEvent(userID, action, resource string, details map[string]interface{}) {
    event := AuditEvent{
        Level:      AuditLevelINFO,
        Action:     AuditAction(action),
        UserID:     userID,
        Resource:   resource,
        Details:    details,
        Timestamp:  time.Now(),
    }
    al.LogEvent(event)
}
```

## 许可证

本项目采用 MIT 许可证。