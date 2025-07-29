package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	App      AppConfig      `json:"app"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Logging  LoggingConfig  `json:"logging"`
	Security SecurityConfig `json:"security"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int `json:"port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
	IdleTimeout  int `json:"idle_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Name            string `json:"name"`
	SSLMode         string `json:"ssl_mode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	MaxRetries   int    `json:"max_retries"`
	DialTimeout  int    `json:"dial_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"output_path"`
	ErrorPath  string `json:"error_path"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWT        JWTConfig        `json:"jwt"`
	Encryption EncryptionConfig `json:"encryption"`
	Audit      AuditConfig      `json:"audit"`
	RateLimit  RateLimitConfig  `json:"rate_limit"`
	Auth       AuthConfig       `json:"auth"`
	Session    SessionConfig    `json:"session"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `json:"secret"`
	Expiration int    `json:"expiration"`
	Issuer     string `json:"issuer"`
	Audience   string `json:"audience"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key       string `json:"key"`
	Algorithm string `json:"algorithm"`
	KeySize   int    `json:"key_size"`
}

// AuditConfig 审计配置
type AuditConfig struct {
	Enabled    bool   `json:"enabled"`
	LogPath    string `json:"log_path"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool `json:"enabled"`
	RPS     int  `json:"rps"`
	Burst   int  `json:"burst"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	MaxLoginAttempts int `json:"max_login_attempts"`
	LockoutDuration  int `json:"lockout_duration"`
	PasswordMinLen   int `json:"password_min_len"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	Timeout    int    `json:"timeout"`
	CookieName string `json:"cookie_name"`
	Secure     bool   `json:"secure"`
	HttpOnly   bool   `json:"http_only"`
	SameSite   string `json:"same_site"`
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "alert-agent"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			Debug:       getEnvBool("APP_DEBUG", true),
		},
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "alert_agent"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 10),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnvInt("REDIS_DIAL_TIMEOUT", 5),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", "stdout"),
			ErrorPath:  getEnv("LOG_ERROR_PATH", "stderr"),
		},
		Security: SecurityConfig{
			JWT: JWTConfig{
				Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
				Expiration: getEnvInt("JWT_EXPIRATION", 3600),
				Issuer:     getEnv("JWT_ISSUER", "alert-agent"),
				Audience:   getEnv("JWT_AUDIENCE", "alert-agent-users"),
			},
			Encryption: EncryptionConfig{
				Key:       getEnv("ENCRYPTION_KEY", "your-encryption-key-32-bytes-long"),
				Algorithm: getEnv("ENCRYPTION_ALGORITHM", "AES-GCM"),
				KeySize:   getEnvInt("ENCRYPTION_KEY_SIZE", 32),
			},
			Audit: AuditConfig{
				Enabled:    getEnvBool("AUDIT_ENABLED", true),
				LogPath:    getEnv("AUDIT_LOG_PATH", "./logs/audit.log"),
				MaxSize:    getEnvInt("AUDIT_MAX_SIZE", 100),
				MaxBackups: getEnvInt("AUDIT_MAX_BACKUPS", 3),
				MaxAge:     getEnvInt("AUDIT_MAX_AGE", 28),
				Compress:   getEnvBool("AUDIT_COMPRESS", true),
			},
			RateLimit: RateLimitConfig{
				Enabled: getEnvBool("RATE_LIMIT_ENABLED", true),
				RPS:     getEnvInt("RATE_LIMIT_RPS", 100),
				Burst:   getEnvInt("RATE_LIMIT_BURST", 200),
			},
			Auth: AuthConfig{
				MaxLoginAttempts: getEnvInt("AUTH_MAX_LOGIN_ATTEMPTS", 5),
				LockoutDuration:  getEnvInt("AUTH_LOCKOUT_DURATION", 900),
				PasswordMinLen:   getEnvInt("AUTH_PASSWORD_MIN_LEN", 8),
				RequireUppercase: getEnvBool("AUTH_REQUIRE_UPPERCASE", true),
				RequireLowercase: getEnvBool("AUTH_REQUIRE_LOWERCASE", true),
				RequireNumbers:   getEnvBool("AUTH_REQUIRE_NUMBERS", true),
				RequireSymbols:   getEnvBool("AUTH_REQUIRE_SYMBOLS", false),
			},
			Session: SessionConfig{
				Timeout:    getEnvInt("SESSION_TIMEOUT", 3600),
				CookieName: getEnv("SESSION_COOKIE_NAME", "alert-agent-session"),
				Secure:     getEnvBool("SESSION_SECURE", false),
				HttpOnly:   getEnvBool("SESSION_HTTP_ONLY", true),
				SameSite:   getEnv("SESSION_SAME_SITE", "Lax"),
			},
		},
	}

	return cfg, nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// getEnv 获取环境变量字符串值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool 获取环境变量布尔值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}