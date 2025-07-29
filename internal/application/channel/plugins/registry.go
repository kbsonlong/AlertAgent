package plugins

import (
	"alert_agent/internal/domain/channel"
	"fmt"
)

// PluginRegistry 插件注册表
type PluginRegistry struct {
	plugins map[channel.ChannelType]func() channel.ChannelPlugin
}

// NewPluginRegistry 创建新的插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[channel.ChannelType]func() channel.ChannelPlugin),
	}
}

// RegisterPlugin 注册插件
func (r *PluginRegistry) RegisterPlugin(channelType channel.ChannelType, factory func() channel.ChannelPlugin) {
	r.plugins[channelType] = factory
}

// CreatePlugin 创建插件实例
func (r *PluginRegistry) CreatePlugin(channelType channel.ChannelType) (channel.ChannelPlugin, error) {
	factory, exists := r.plugins[channelType]
	if !exists {
		return nil, fmt.Errorf("plugin for channel type %s not registered", channelType)
	}
	return factory(), nil
}

// GetRegisteredTypes 获取已注册的插件类型
func (r *PluginRegistry) GetRegisteredTypes() []channel.ChannelType {
	types := make([]channel.ChannelType, 0, len(r.plugins))
	for channelType := range r.plugins {
		types = append(types, channelType)
	}
	return types
}

// IsRegistered 检查插件是否已注册
func (r *PluginRegistry) IsRegistered(channelType channel.ChannelType) bool {
	_, exists := r.plugins[channelType]
	return exists
}

// GetDefaultRegistry 获取默认插件注册表（包含所有内置插件）
func GetDefaultRegistry() *PluginRegistry {
	registry := NewPluginRegistry()
	
	// 注册内置插件
	registry.RegisterPlugin(channel.ChannelTypeDingTalk, func() channel.ChannelPlugin {
		return NewDingTalkPlugin()
	})
	
	registry.RegisterPlugin(channel.ChannelTypeWeChat, func() channel.ChannelPlugin {
		return NewWeChatPlugin()
	})
	
	registry.RegisterPlugin(channel.ChannelTypeEmail, func() channel.ChannelPlugin {
		return NewEmailPlugin()
	})
	
	registry.RegisterPlugin(channel.ChannelTypeWebhook, func() channel.ChannelPlugin {
		return NewWebhookPlugin()
	})
	
	registry.RegisterPlugin(channel.ChannelTypeSlack, func() channel.ChannelPlugin {
		return NewSlackPlugin()
	})
	
	return registry
}

// RegisterAllPlugins 向渠道管理器注册所有插件
func RegisterAllPlugins(manager channel.ChannelManager) error {
	registry := GetDefaultRegistry()
	
	for _, channelType := range registry.GetRegisteredTypes() {
		plugin, err := registry.CreatePlugin(channelType)
		if err != nil {
			return fmt.Errorf("failed to create plugin %s: %w", channelType, err)
		}
		
		if err := manager.RegisterPlugin(plugin); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", channelType, err)
		}
	}
	
	return nil
}