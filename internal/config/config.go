package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var GlobalConfig Config

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
}

// OllamaConfig Ollama配置
type OllamaConfig struct {
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

	// 读取配置文件
	configPath := filepath.Join(workDir, "config", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 解析配置
	if err := yaml.Unmarshal(data, &GlobalConfig); err != nil {
		return err
	}

	return nil
}
