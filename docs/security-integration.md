# AlertAgent 安全框架集成指南

## 概述

AlertAgent 项目已成功集成了完整的安全框架，提供了用户认证、授权、审计日志等核心安全功能。

## 集成内容

### 1. 安全配置

在 `internal/infrastructure/config/config.go` 中添加了完整的安全配置结构：

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

### 2. 数据库模型

在 `internal/security/domain/` 目录下创建了以下实体模型：

- **User**: 用户实体，包含用户名、邮箱、密码哈希等信息
- **Role**: 角色实体，支持基于角色的访问控制
- **Permission**: 权限实体，定义具体的操作权限
- **AuditLog**: 审计日志实体，记录用户操作历史

### 3. 依赖注入集成

在主容器 `internal/infrastructure/di/container.go` 中集成了安全容器：

```go
type Container struct {
    // ... 其他字段
    config            *config.Config
    securityContainer *di.Container
}
```

### 4. 路由集成

在 `internal/interfaces/http/routes.go` 中集成了安全路由：

```go
func (r *Router) SetupRoutes(engine *gin.Engine) {
    // 设置全局安全中间件
    routes.SetupSecurityMiddleware(engine, r.securityContainer)
    
    // 设置认证路由
    routes.SetupAuthRoutes(engine, r.securityContainer)
    
    // 设置健康检查路由
    routes.SetupHealthRoutes(engine, r.securityContainer)
    
    // ... 其他路由
}
```

### 5. 数据库迁移

在 `internal/infrastructure/database/database.go` 中添加了安全相关表的自动迁移：

```go
func Migrate(db *gorm.DB) error {
    err := db.AutoMigrate(
        &cluster.Cluster{},
        &channel.Channel{},
        &domain.User{},
        &domain.Role{},
        &domain.Permission{},
        &domain.AuditLog{},
    )
    return err
}
```

## 核心功能

### 1. 用户管理

- 用户注册和登录
- 密码加密存储
- 用户状态管理（激活/禁用）
- 登录失败锁定机制

### 2. JWT 认证

- JWT 令牌生成和验证
- 令牌刷新机制
- 自定义声明支持

### 3. RBAC 权限控制

- 基于角色的访问控制
- 细粒度权限管理
- 动态权限检查

### 4. 审计日志

- 用户操作记录
- 安全事件追踪
- 详细的审计信息

### 5. 数据加密

- 密码哈希加密
- 敏感数据加密
- 安全的密钥管理

### 6. 输入验证

- 请求参数验证
- 数据格式检查
- 安全过滤

## 使用示例

### 启动应用

```bash
# 编译主应用
go build ./cmd/api

# 运行应用
./api
```

### 测试安全功能

```bash
# 编译测试程序
go build ./cmd/security-test

# 运行安全功能测试
./security-test
```

### API 使用示例

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "roles": ["user"]
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

#### 访问受保护的资源

```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <your-jwt-token>"
```

## 配置说明

### 环境变量

```bash
# JWT 配置
SECURITY_JWT_SECRET_KEY=your-secret-key
SECURITY_JWT_EXPIRATION_HOURS=24

# 加密配置
SECURITY_ENCRYPTION_KEY=your-encryption-key

# 审计配置
SECURITY_AUDIT_ENABLED=true
SECURITY_AUDIT_LOG_LEVEL=info

# 认证配置
SECURITY_AUTH_MAX_LOGIN_ATTEMPTS=5
SECURITY_AUTH_LOCKOUT_DURATION_MINUTES=30
```

### 配置文件示例

```yaml
security:
  jwt:
    secret_key: "your-secret-key"
    expiration_hours: 24
    refresh_expiration_hours: 168
  encryption:
    key: "your-encryption-key"
    algorithm: "AES-256-GCM"
  audit:
    enabled: true
    log_level: "info"
    retention_days: 90
  auth:
    max_login_attempts: 5
    lockout_duration_minutes: 30
    password_min_length: 8
```

## 安全最佳实践

1. **密钥管理**: 使用环境变量或安全的密钥管理服务存储敏感信息
2. **HTTPS**: 在生产环境中始终使用 HTTPS
3. **令牌过期**: 设置合理的 JWT 令牌过期时间
4. **审计日志**: 启用审计日志记录重要操作
5. **输入验证**: 对所有用户输入进行严格验证
6. **权限最小化**: 遵循最小权限原则分配用户权限

## 故障排除

### 常见问题

1. **数据库连接失败**: 检查数据库配置和连接字符串
2. **JWT 验证失败**: 确认密钥配置正确
3. **权限拒绝**: 检查用户角色和权限配置
4. **迁移失败**: 确保数据库用户有足够的权限

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看审计日志
tail -f logs/audit.log
```

## 扩展开发

### 添加新的权限

1. 在数据库中添加新的权限记录
2. 在相应的处理器中添加权限检查
3. 更新 RBAC 配置

### 自定义中间件

```go
func CustomSecurityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 自定义安全逻辑
        c.Next()
    }
}
```

### 扩展审计功能

```go
func LogCustomEvent(auditLogger *audit.AuditLogger, userID, action, resource string) {
    auditLogger.Log(context.Background(), &audit.LogEntry{
        UserID:   userID,
        Action:   action,
        Resource: resource,
        // ... 其他字段
    })
}
```

## 总结

AlertAgent 安全框架提供了完整的企业级安全功能，包括用户认证、授权、审计等核心特性。通过模块化设计和依赖注入，框架易于扩展和维护，满足现代应用的安全需求。