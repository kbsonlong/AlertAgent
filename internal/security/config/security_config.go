package config

import (
	"fmt"
	"time"
)

// SecurityConfig 安全配置结构
type SecurityConfig struct {
	// JWT配置
	JWT JWTConfig `yaml:"jwt" json:"jwt"`
	
	// 加密配置
	Encryption EncryptionConfig `yaml:"encryption" json:"encryption"`
	
	// 审计配置
	Audit AuditConfig `yaml:"audit" json:"audit"`
	
	// 限流配置
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	
	// 认证配置
	Auth AuthConfig `yaml:"auth" json:"auth"`
	
	// 会话配置
	Session SessionConfig `yaml:"session" json:"session"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string        `yaml:"secret" json:"secret"`
	Expiration       time.Duration `yaml:"expiration" json:"expiration"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration" json:"refresh_expiration"`
	Issuer           string        `yaml:"issuer" json:"issuer"`
	Audience         string        `yaml:"audience" json:"audience"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key        string `yaml:"key" json:"key"`
	Salt       string `yaml:"salt" json:"salt"`
	Iterations int    `yaml:"iterations" json:"iterations"`
	KeyLength  int    `yaml:"key_length" json:"key_length"`
}

// AuditConfig 审计配置
type AuditConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	LogLevel   string `yaml:"log_level" json:"log_level"`
	LogFile    string `yaml:"log_file" json:"log_file"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // 保留的日志文件数
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // 天数
	Compress   bool   `yaml:"compress" json:"compress"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	RequestsPerMinute int           `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size" json:"burst_size"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	RedisEnabled      bool          `yaml:"redis_enabled" json:"redis_enabled"`
	RedisAddr         string        `yaml:"redis_addr" json:"redis_addr"`
	RedisPassword     string        `yaml:"redis_password" json:"redis_password"`
	RedisDB           int           `yaml:"redis_db" json:"redis_db"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	SkipPaths         []string      `yaml:"skip_paths" json:"skip_paths"`
	MaxLoginAttempts  int           `yaml:"max_login_attempts" json:"max_login_attempts"`
	LockoutDuration   time.Duration `yaml:"lockout_duration" json:"lockout_duration"`
	PasswordMinLength int           `yaml:"password_min_length" json:"password_min_length"`
	PasswordRequireSpecial bool      `yaml:"password_require_special" json:"password_require_special"`
	PasswordRequireNumber  bool      `yaml:"password_require_number" json:"password_require_number"`
	PasswordRequireUpper   bool      `yaml:"password_require_upper" json:"password_require_upper"`
	PasswordRequireLower   bool      `yaml:"password_require_lower" json:"password_require_lower"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	Secure         bool          `yaml:"secure" json:"secure"`
	HttpOnly       bool          `yaml:"http_only" json:"http_only"`
	SameSite       string        `yaml:"same_site" json:"same_site"`
	Domain         string        `yaml:"domain" json:"domain"`
	Path           string        `yaml:"path" json:"path"`
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWT: JWTConfig{
			Secret:            "your-secret-key-change-in-production",
			Expiration:        time.Hour * 24,     // 24小时
			RefreshExpiration: time.Hour * 24 * 7, // 7天
			Issuer:            "alert-agent",
			Audience:          "alert-agent-users",
		},
		Encryption: EncryptionConfig{
			Key:        "your-encryption-key-32-bytes-long",
			Salt:       "your-salt-16-bytes",
			Iterations: 100000,
			KeyLength:  32,
		},
		Audit: AuditConfig{
			Enabled:    true,
			LogLevel:   "info",
			LogFile:    "/var/log/alert-agent/audit.log",
			MaxSize:    100, // 100MB
			MaxBackups: 10,
			MaxAge:     30, // 30天
			Compress:   true,
		},
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 100,
			BurstSize:         10,
			CleanupInterval:   time.Minute * 5,
			RedisEnabled:      false,
			RedisAddr:         "localhost:6379",
			RedisPassword:     "",
			RedisDB:           0,
		},
		Auth: AuthConfig{
			SkipPaths: []string{
				"/api/v1/health",
				"/api/v1/metrics",
				"/api/v1/login",
				"/api/v1/register",
				"/swagger",
				"/docs",
			},
			MaxLoginAttempts:       5,
			LockoutDuration:        time.Minute * 15,
			PasswordMinLength:      8,
			PasswordRequireSpecial: true,
			PasswordRequireNumber:  true,
			PasswordRequireUpper:   true,
			PasswordRequireLower:   true,
		},
		Session: SessionConfig{
			Timeout:         time.Hour * 2, // 2小时
			CleanupInterval: time.Minute * 30,
			Secure:          true,
			HttpOnly:        true,
			SameSite:        "Strict",
			Domain:          "",
			Path:            "/",
		},
	}
}

// Validate 验证安全配置
func (c *SecurityConfig) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}
	
	if c.JWT.Expiration <= 0 {
		return fmt.Errorf("JWT expiration must be positive")
	}
	
	if c.Encryption.Key == "" {
		return fmt.Errorf("encryption key cannot be empty")
	}
	
	if len(c.Encryption.Key) < 32 {
		return fmt.Errorf("encryption key must be at least 32 characters")
	}
	
	if c.Auth.PasswordMinLength < 6 {
		return fmt.Errorf("password minimum length must be at least 6")
	}
	
	return nil
}