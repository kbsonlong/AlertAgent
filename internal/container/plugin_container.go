package container

import (
	"alert_agent/internal/service"
	"sync"
)

var (
	pluginManagerService *service.PluginManagerService
	pluginManagerOnce    sync.Once
)

// GetPluginManagerService 获取插件管理服务单例
func GetPluginManagerService() *service.PluginManagerService {
	pluginManagerOnce.Do(func() {
		pluginManagerService = service.NewPluginManagerService()
	})
	return pluginManagerService
}

// InitializePluginManager 初始化插件管理器
func InitializePluginManager() error {
	service := GetPluginManagerService()
	return service.Start()
}

// ShutdownPluginManager 关闭插件管理器
func ShutdownPluginManager() error {
	if pluginManagerService != nil {
		return pluginManagerService.Stop()
	}
	return nil
}