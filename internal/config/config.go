package config

import (
	"alert_agent/internal/pkg/logger"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	GlobalConfig    Config
	configMutex     sync.RWMutex
	configPath      string
	watcher         *fsnotify.Watcher
	reloadCallbacks []func(Config)
)

// Config 配置结构
type Config struct {
	Server struct {
		Port         int    `yaml:"port"`
		Mode         string `yaml:"mode"`
		JWTSecret    string `yaml:"jwt_secret"`
		ReadTimeout  int    `yaml:"read_timeout"`
		WriteTimeout int    `yaml:"write_timeout"`
		IdleTimeout  int    `yaml:"idle_timeout"`
	} `yaml:"server"`
	Gateway struct {
		RateLimit struct {
			Enabled bool `yaml:"enabled"`
			RPS     int  `yaml:"rps"`     // requests per second
			Burst   int  `yaml:"burst"`   // burst size
		} `yaml:"rate_limit"`
		Auth struct {
			Enabled       bool     `yaml:"enabled"`
			SkipPaths     []string `yaml:"skip_paths"`
			TokenExpiry   int      `yaml:"token_expiry"` // hours
			RefreshExpiry int      `yaml:"refresh_expiry"` // hours
		} `yaml:"auth"`
		CORS struct {
			Enabled          bool     `yaml:"enabled"`
			AllowOrigins     []string `yaml:"allow_origins"`
			AllowMethods     []string `yaml:"allow_methods"`
			AllowHeaders     []string `yaml:"allow_headers"`
			ExposeHeaders    []string `yaml:"expose_headers"`
			AllowCredentials bool     `yaml:"allow_credentials"`
		} `yaml:"cors"`
	} `yaml:"gateway"`
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		DBName       string `yaml:"dbname"`
		Charset      string `yaml:"charset"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxOpenConns int    `yaml:"max_open_conns"`
	} `yaml:"database"`
	Ollama OllamaConfig `yaml:"ollama"`
	Dify   DifyConfig   `yaml:"dify"`
	Redis  struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Password     string `yaml:"password"`
		DB           int    `yaml:"db"`
		PoolSize     int    `yaml:"pool_size"`
		MinIdleConns int    `yaml:"min_idle_conns"`
		MaxRetries   int    `yaml:"max_retries"`
		DialTimeout  int    `yaml:"dial_timeout"`
	} `yaml:"redis"`
	Log struct {
		Level      string `yaml:"level"`
		Filename   string `yaml:"filename"`
		MaxSize    int    `yaml:"max_size"`
		MaxAge     int    `yaml:"max_age"`
		MaxBackups int    `yaml:"max_backups"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
	Worker struct {
		Enabled     bool `yaml:"enabled"`
		Concurrency int  `yaml:"concurrency"`
	} `yaml:"worker"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Server: struct {
			Port         int    `yaml:"port"`
			Mode         string `yaml:"mode"`
			JWTSecret    string `yaml:"jwt_secret"`
			ReadTimeout  int    `yaml:"read_timeout"`
			WriteTimeout int    `yaml:"write_timeout"`
			IdleTimeout  int    `yaml:"idle_timeout"`
		}{
			Port:         8080,
			Mode:         "release",
			JWTSecret:    "change-me-in-production",
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  60,
		},
		Database: struct {
			Host         string `yaml:"host"`
			Port         int    `yaml:"port"`
			Username     string `yaml:"username"`
			Password     string `yaml:"password"`
			DBName       string `yaml:"dbname"`
			Charset      string `yaml:"charset"`
			MaxIdleConns int    `yaml:"max_idle_conns"`
			MaxOpenConns int    `yaml:"max_open_conns"`
		}{
			Host:         "localhost",
			Port:         3306,
			Username:     "root",
			Password:     "",
			DBName:       "alert_agent",
			Charset:      "utf8mb4",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		},
		Ollama: OllamaConfig{
			Enabled:     false,
			APIEndpoint: "http://localhost:11434",
			Model:       "llama3:latest",
			Timeout:     30,
			MaxRetries:  3,
		},
		Redis: struct {
			Host         string `yaml:"host"`
			Port         int    `yaml:"port"`
			Password     string `yaml:"password"`
			DB           int    `yaml:"db"`
			PoolSize     int    `yaml:"pool_size"`
			MinIdleConns int    `yaml:"min_idle_conns"`
			MaxRetries   int    `yaml:"max_retries"`
			DialTimeout  int    `yaml:"dial_timeout"`
		}{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           0,
			PoolSize:     100,
			MinIdleConns: 10,
			MaxRetries:   3,
			DialTimeout:  5,
		},
		Gateway: struct {
			RateLimit struct {
				Enabled bool `yaml:"enabled"`
				RPS     int  `yaml:"rps"`
				Burst   int  `yaml:"burst"`
			} `yaml:"rate_limit"`
			Auth struct {
				Enabled       bool     `yaml:"enabled"`
				SkipPaths     []string `yaml:"skip_paths"`
				TokenExpiry   int      `yaml:"token_expiry"`
				RefreshExpiry int      `yaml:"refresh_expiry"`
			} `yaml:"auth"`
			CORS struct {
				Enabled          bool     `yaml:"enabled"`
				AllowOrigins     []string `yaml:"allow_origins"`
				AllowMethods     []string `yaml:"allow_methods"`
				AllowHeaders     []string `yaml:"allow_headers"`
				ExposeHeaders    []string `yaml:"expose_headers"`
				AllowCredentials bool     `yaml:"allow_credentials"`
			} `yaml:"cors"`
		}{
			RateLimit: struct {
				Enabled bool `yaml:"enabled"`
				RPS     int  `yaml:"rps"`
				Burst   int  `yaml:"burst"`
			}{
				Enabled: true,
				RPS:     100,
				Burst:   200,
			},
			Auth: struct {
				Enabled       bool     `yaml:"enabled"`
				SkipPaths     []string `yaml:"skip_paths"`
				TokenExpiry   int      `yaml:"token_expiry"`
				RefreshExpiry int      `yaml:"refresh_expiry"`
			}{
				Enabled:       true,
				SkipPaths:     []string{"/api/v1/health", "/api/v1/auth/login"},
				TokenExpiry:   24,
				RefreshExpiry: 168,
			},
			CORS: struct {
				Enabled          bool     `yaml:"enabled"`
				AllowOrigins     []string `yaml:"allow_origins"`
				AllowMethods     []string `yaml:"allow_methods"`
				AllowHeaders     []string `yaml:"allow_headers"`
				ExposeHeaders    []string `yaml:"expose_headers"`
				AllowCredentials bool     `yaml:"allow_credentials"`
			}{
				Enabled:          true,
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
				ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
				AllowCredentials: true,
			},
		},
		Log: struct {
			Level      string `yaml:"level"`
			Filename   string `yaml:"filename"`
			MaxSize    int    `yaml:"max_size"`
			MaxAge     int    `yaml:"max_age"`
			MaxBackups int    `yaml:"max_backups"`
			Compress   bool   `yaml:"compress"`
		}{
			Level:      "info",
			Filename:   "logs/alert_agent.log",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
			Compress:   true,
		},
		Worker: struct {
			Enabled     bool `yaml:"enabled"`
			Concurrency int  `yaml:"concurrency"`
		}{
			Enabled:     true,
			Concurrency: 2,
		},
	}
}

// OllamaConfig Ollama配置
type OllamaConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
	Model       string `yaml:"model"`
	Timeout     int    `yaml:"timeout"`
	MaxRetries  int    `yaml:"max_retries"`
}

// DifyConfig Dify配置
type DifyConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
	APIKey      string `yaml:"api_key"`
	Model       string `yaml:"model"`
	Timeout     int    `yaml:"timeout"`
	MaxRetries  int    `yaml:"max_retries"`
}

// Load 加载配置
func Load() error {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 设置配置文件路径
	configPath = filepath.Join(workDir, "config", "config.yaml")

	// 加载配置
	if err := loadConfig(); err != nil {
		return err
	}

	// 启动文件监听
	if err := startWatcher(); err != nil {
		logger.L.Warn("Failed to start config file watcher", zap.Error(err))
	}

	return nil
}

// loadConfig 从文件加载配置
func loadConfig() error {
	// 从默认配置开始
	tempConfig := DefaultConfig()
	
	// 如果配置文件存在，则加载并合并
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &tempConfig); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else {
		logger.L.Warn("Config file not found, using default configuration", zap.String("file", configPath))
	}

	// 应用环境变量覆盖
	if err := applyEnvOverrides(&tempConfig); err != nil {
		return fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// 验证配置
	if err := validateConfig(&tempConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 使用写锁更新全局配置
	configMutex.Lock()
	GlobalConfig = tempConfig
	configMutex.Unlock()

	logger.L.Info("Configuration loaded successfully", zap.String("file", configPath))
	return nil
}

// applyEnvOverrides 应用环境变量覆盖
func applyEnvOverrides(config *Config) error {
	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Server.JWTSecret = secret
	}

	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if username := os.Getenv("DB_USERNAME"); username != "" {
		config.Database.Username = username
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		config.Database.DBName = dbname
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = d
		}
	}

	// Ollama配置
	if enabled := os.Getenv("OLLAMA_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Ollama.Enabled = e
		}
	}
	if endpoint := os.Getenv("OLLAMA_ENDPOINT"); endpoint != "" {
		config.Ollama.APIEndpoint = endpoint
	}
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		config.Ollama.Model = model
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Log.Level = level
	}
	if filename := os.Getenv("LOG_FILENAME"); filename != "" {
		config.Log.Filename = filename
	}

	// Worker配置
	if enabled := os.Getenv("WORKER_ENABLED"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			config.Worker.Enabled = e
		}
	}
	if concurrency := os.Getenv("WORKER_CONCURRENCY"); concurrency != "" {
		if c, err := strconv.Atoi(concurrency); err == nil {
			config.Worker.Concurrency = c
		}
	}

	return nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	var errors []string

	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		errors = append(errors, "server.port must be between 1 and 65535")
	}
	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		errors = append(errors, "server.mode must be one of: debug, release, test")
	}
	if config.Server.JWTSecret == "" || config.Server.JWTSecret == "change-me-in-production" {
		if config.Server.Mode == "release" {
			errors = append(errors, "server.jwt_secret must be set in production mode")
		}
	}

	// 验证数据库配置
	if config.Database.Host == "" {
		errors = append(errors, "database.host is required")
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		errors = append(errors, "database.port must be between 1 and 65535")
	}
	if config.Database.Username == "" {
		errors = append(errors, "database.username is required")
	}
	if config.Database.DBName == "" {
		errors = append(errors, "database.dbname is required")
	}

	// 验证Redis配置
	if config.Redis.Host == "" {
		errors = append(errors, "redis.host is required")
	}
	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		errors = append(errors, "redis.port must be between 1 and 65535")
	}

	// 验证Ollama配置
	if config.Ollama.Enabled {
		if config.Ollama.APIEndpoint == "" {
			errors = append(errors, "ollama.api_endpoint is required when ollama is enabled")
		}
		if config.Ollama.Model == "" {
			errors = append(errors, "ollama.model is required when ollama is enabled")
		}
	}

	// 验证日志配置
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	isValidLevel := false
	for _, level := range validLogLevels {
		if config.Log.Level == level {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		errors = append(errors, fmt.Sprintf("log.level must be one of: %s", strings.Join(validLogLevels, ", ")))
	}

	// 验证网关配置
	if config.Gateway.RateLimit.Enabled {
		if config.Gateway.RateLimit.RPS <= 0 {
			errors = append(errors, "gateway.rate_limit.rps must be greater than 0")
		}
		if config.Gateway.RateLimit.Burst <= 0 {
			errors = append(errors, "gateway.rate_limit.burst must be greater than 0")
		}
	}

	if config.Gateway.Auth.Enabled {
		if config.Gateway.Auth.TokenExpiry <= 0 {
			errors = append(errors, "gateway.auth.token_expiry must be greater than 0")
		}
		if config.Gateway.Auth.RefreshExpiry <= 0 {
			errors = append(errors, "gateway.auth.refresh_expiry must be greater than 0")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// startWatcher 启动配置文件监听
func startWatcher() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 添加配置文件到监听列表
	if err := watcher.Add(configPath); err != nil {
		return err
	}

	// 启动监听协程
	go func() {
		defer watcher.Close()

		// 防抖动：避免短时间内多次重载
		var lastReload time.Time
		const debounceInterval = 1 * time.Second

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 只处理写入和创建事件
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// 防抖动检查
					if time.Since(lastReload) < debounceInterval {
						continue
					}

					logger.L.Info("Config file changed", zap.String("file", event.Name))

					// 重新加载配置
					if err := reloadConfig(); err != nil {
						logger.L.Error("Failed to reload config", zap.Error(err))
					} else {
						lastReload = time.Now()
						logger.L.Info("Configuration reloaded successfully")
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.L.Error("Config watcher error", zap.Error(err))
			}
		}
	}()

	logger.L.Info("Started watching config file", zap.String("file", configPath))
	return nil
}

// reloadConfig 重新加载配置并触发回调
func reloadConfig() error {
	// 重新加载配置
	if err := loadConfig(); err != nil {
		return err
	}

	// 获取当前配置的副本用于回调
	configMutex.RLock()
	currentConfig := GlobalConfig
	configMutex.RUnlock()

	// 触发所有注册的回调函数
	for _, callback := range reloadCallbacks {
		go func(cb func(Config)) {
			defer func() {
				if r := recover(); r != nil {
					logger.L.Error("Config reload callback panic", zap.Any("panic", r))
				}
			}()
			cb(currentConfig)
		}(callback)
	}

	return nil
}

// GetConfig 安全地获取配置
func GetConfig() Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return GlobalConfig
}

// RegisterReloadCallback 注册配置重载回调函数
func RegisterReloadCallback(callback func(Config)) {
	reloadCallbacks = append(reloadCallbacks, callback)
}

// StopWatcher 停止配置文件监听
func StopWatcher() {
	if watcher != nil {
		watcher.Close()
	}
}

// GetConfigAsYAML 获取配置的YAML格式
func GetConfigAsYAML() ([]byte, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	
	return yaml.Marshal(GlobalConfig)
}

// UpdateConfig 更新配置
func UpdateConfig(newConfig Config) error {
	// 验证新配置
	if err := validateConfig(&newConfig); err != nil {
		return fmt.Errorf("new config validation failed: %w", err)
	}

	// 更新全局配置
	configMutex.Lock()
	GlobalConfig = newConfig
	configMutex.Unlock()

	// 触发回调
	for _, callback := range reloadCallbacks {
		go func(cb func(Config)) {
			defer func() {
				if r := recover(); r != nil {
					logger.L.Error("Config update callback panic", zap.Any("panic", r))
				}
			}()
			cb(newConfig)
		}(callback)
	}

	logger.L.Info("Configuration updated successfully")
	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	configMutex.RLock()
	config := GlobalConfig
	configMutex.RUnlock()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 原子性写入
	tmpFile := configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	if err := os.Rename(tmpFile, configPath); err != nil {
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}

	logger.L.Info("Configuration saved successfully", zap.String("file", configPath))
	return nil
}

// GetConfigValue 获取配置值（支持点分隔路径）
func GetConfigValue(path string) (interface{}, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	parts := strings.Split(path, ".")
	value := reflect.ValueOf(GlobalConfig)

	for _, part := range parts {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() != reflect.Struct {
			return nil, fmt.Errorf("invalid path: %s", path)
		}

		field := value.FieldByName(strings.Title(part))
		if !field.IsValid() {
			return nil, fmt.Errorf("field not found: %s", part)
		}

		value = field
	}

	return value.Interface(), nil
}

// SetConfigValue 设置配置值（支持点分隔路径）
func SetConfigValue(path string, newValue interface{}) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	parts := strings.Split(path, ".")
	value := reflect.ValueOf(&GlobalConfig).Elem()

	for i, part := range parts {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() != reflect.Struct {
			return fmt.Errorf("invalid path: %s", path)
		}

		field := value.FieldByName(strings.Title(part))
		if !field.IsValid() {
			return fmt.Errorf("field not found: %s", part)
		}

		if i == len(parts)-1 {
			// 最后一个字段，设置值
			if !field.CanSet() {
				return fmt.Errorf("field cannot be set: %s", part)
			}

			newVal := reflect.ValueOf(newValue)
			if field.Type() != newVal.Type() {
				return fmt.Errorf("type mismatch for field %s: expected %s, got %s", 
					part, field.Type(), newVal.Type())
			}

			field.Set(newVal)
		} else {
			value = field
		}
	}

	return nil
}

// ResetToDefaults 重置为默认配置
func ResetToDefaults() error {
	defaultConfig := DefaultConfig()
	return UpdateConfig(defaultConfig)
}
