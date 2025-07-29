package container

import (
	"fmt"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	services   map[string]interface{}
	factories  map[string]func() interface{}
	singletons map[string]interface{}
	mutex      sync.RWMutex
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		services:   make(map[string]interface{}),
		factories:  make(map[string]func() interface{}),
		singletons: make(map[string]interface{}),
	}
}

// Register 注册服务
func (c *Container) Register(name string, service interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.services[name] = service
}

// RegisterFactory 注册工厂函数
func (c *Container) RegisterFactory(name string, factory func() interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.factories[name] = factory
}

// RegisterSingleton 注册单例服务
func (c *Container) RegisterSingleton(name string, factory func() interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.factories[name] = factory
	// 标记为单例
	c.singletons[name] = nil
}

// Get 获取服务
func (c *Container) Get(name string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// 检查已注册的服务
	if service, exists := c.services[name]; exists {
		return service, nil
	}
	
	// 检查单例缓存
	if singleton, exists := c.singletons[name]; exists && singleton != nil {
		return singleton, nil
	}
	
	// 检查工厂函数
	if factory, exists := c.factories[name]; exists {
		service := factory()
		// 如果是单例，缓存结果
		if _, isSingleton := c.singletons[name]; isSingleton {
			c.singletons[name] = service
		}
		return service, nil
	}
	
	return nil, fmt.Errorf("service %s not found", name)
}

// GetT 获取指定类型的服务
func GetT[T any](c *Container, name string) (T, error) {
	var zero T
	service, err := c.Get(name)
	if err != nil {
		return zero, err
	}
	
	if typed, ok := service.(T); ok {
		return typed, nil
	}
	
	return zero, fmt.Errorf("service %s is not of type %T", name, zero)
}

// MustGet 获取服务，如果不存在则panic
func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// Has 检查服务是否存在
func (c *Container) Has(name string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	_, hasService := c.services[name]
	_, hasFactory := c.factories[name]
	return hasService || hasFactory
}

// Remove 移除服务
func (c *Container) Remove(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.services, name)
	delete(c.factories, name)
	delete(c.singletons, name)
}

// Clear 清空容器
func (c *Container) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.services = make(map[string]interface{})
	c.factories = make(map[string]func() interface{})
	c.singletons = make(map[string]interface{})
}

// Services 获取所有已注册的服务名称
func (c *Container) Services() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	var names []string
	for name := range c.services {
		names = append(names, name)
	}
	for name := range c.factories {
		names = append(names, name)
	}
	return names
}