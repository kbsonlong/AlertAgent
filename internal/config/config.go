package config

import (
	"alert_agent/internal/pkg/logger"
	"os"
	"path/filepath"
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
		Port      int    `yaml:"port"`
		Mode      string `yaml:"mode"`
		JWTSecret string `yaml:"jwt_secret"`
	} `yaml:"server"`
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
	Features FeatureConfig `yaml:"features"`
}

// FeatureConfig 功能开关配置
type FeatureConfig struct {
	Enabled                bool                   `yaml:"enabled"`                  // 是否启用功能开关系统
	ConfigPath             string                 `yaml:"config_path"`              // 功能配置文件路径
	MonitoringEnabled      bool                   `yaml:"monitoring_enabled"`       // 是否启用监控
	AIMaturityEnabled      bool                   `yaml:"ai_maturity_enabled"`      // 是否启用AI成熟度评估
	DefaultPhase           string                 `yaml:"default_phase"`            // 默认阶段
	AutoDegradationEnabled bool                   `yaml:"auto_degradation_enabled"` // 是否启用自动降级
	MetricsRetentionHours  int                    `yaml:"metrics_retention_hours"`  // 指标保留时间
	AlertingConfig         FeatureAlertingConfig  `yaml:"alerting"`                 // 告警配置
}

// FeatureAlertingConfig 功能告警配置
type FeatureAlertingConfig struct {
	Enabled           bool                   `yaml:"enabled"`
	WebhookURL        string                 `yaml:"webhook_url"`
	SlackChannel      string                 `yaml:"slack_channel"`
	EmailRecipients   []string               `yaml:"email_recipients"`
	AlertThresholds   map[string]float64     `yaml:"alert_thresholds"`
}

// OllamaConfig Ollama配置
type OllamaConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
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
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 创建临时配置对象
	var tempConfig Config
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return err
	}

	// 使用写锁更新全局配置
	configMutex.Lock()
	GlobalConfig = tempConfig
	configMutex.Unlock()

	logger.L.Debug("Configuration loaded successfully from %s", zap.String("file", configPath))
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
