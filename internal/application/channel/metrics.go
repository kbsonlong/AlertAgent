package channel

import (
	"sync"
	"time"

	"alert_agent/internal/domain/channel"
)

// ChannelMetrics 渠道指标
type ChannelMetrics struct {
	mutex           sync.RWMutex
	messagesSent    map[channel.ChannelType]int64
	messagesFailed  map[channel.ChannelType]int64
	latencies       map[channel.ChannelType][]time.Duration
	channelsCreated map[channel.ChannelType]int64
	channelsDeleted map[channel.ChannelType]int64
	lastUpdated     time.Time
}

// NewChannelMetrics 创建渠道指标
func NewChannelMetrics() *ChannelMetrics {
	return &ChannelMetrics{
		messagesSent:    make(map[channel.ChannelType]int64),
		messagesFailed:  make(map[channel.ChannelType]int64),
		latencies:       make(map[channel.ChannelType][]time.Duration),
		channelsCreated: make(map[channel.ChannelType]int64),
		channelsDeleted: make(map[channel.ChannelType]int64),
		lastUpdated:     time.Now(),
	}
}

// IncMessageSent 增加发送成功计数
func (m *ChannelMetrics) IncMessageSent(channelType channel.ChannelType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messagesSent[channelType]++
	m.lastUpdated = time.Now()
}

// IncMessageFailed 增加发送失败计数
func (m *ChannelMetrics) IncMessageFailed(channelType channel.ChannelType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messagesFailed[channelType]++
	m.lastUpdated = time.Now()
}

// RecordLatency 记录延迟
func (m *ChannelMetrics) RecordLatency(channelType channel.ChannelType, latency time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 保持最近100个延迟记录
	latencies := m.latencies[channelType]
	if len(latencies) >= 100 {
		latencies = latencies[1:]
	}
	latencies = append(latencies, latency)
	m.latencies[channelType] = latencies
	m.lastUpdated = time.Now()
}

// IncChannelCreated 增加渠道创建计数
func (m *ChannelMetrics) IncChannelCreated(channelType channel.ChannelType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.channelsCreated[channelType]++
	m.lastUpdated = time.Now()
}

// IncChannelDeleted 增加渠道删除计数
func (m *ChannelMetrics) IncChannelDeleted(channelType channel.ChannelType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.channelsDeleted[channelType]++
	m.lastUpdated = time.Now()
}

// GetMessagesSent 获取发送成功计数
func (m *ChannelMetrics) GetMessagesSent(channelType channel.ChannelType) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.messagesSent[channelType]
}

// GetMessagesFailed 获取发送失败计数
func (m *ChannelMetrics) GetMessagesFailed(channelType channel.ChannelType) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.messagesFailed[channelType]
}

// GetAverageLatency 获取平均延迟
func (m *ChannelMetrics) GetAverageLatency(channelType channel.ChannelType) time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	latencies := m.latencies[channelType]
	if len(latencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	
	return total / time.Duration(len(latencies))
}

// GetSuccessRate 获取成功率
func (m *ChannelMetrics) GetSuccessRate(channelType channel.ChannelType) float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	sent := m.messagesSent[channelType]
	failed := m.messagesFailed[channelType]
	total := sent + failed
	
	if total == 0 {
		return 0
	}
	
	return float64(sent) / float64(total) * 100
}

// GlobalMetrics 全局指标
type GlobalMetrics struct {
	TotalSent   int64
	TotalFailed int64
	AvgLatency  float64
}

// GetGlobalMetrics 获取全局指标
func (m *ChannelMetrics) GetGlobalMetrics() *GlobalMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var totalSent, totalFailed int64
	var totalLatency time.Duration
	var latencyCount int
	
	// 计算总发送数和失败数
	for _, sent := range m.messagesSent {
		totalSent += sent
	}
	for _, failed := range m.messagesFailed {
		totalFailed += failed
	}
	
	// 计算平均延迟
	for _, latencies := range m.latencies {
		for _, latency := range latencies {
			totalLatency += latency
			latencyCount++
		}
	}
	
	var avgLatency float64
	if latencyCount > 0 {
		avgLatency = float64(totalLatency.Nanoseconds()) / float64(latencyCount) / 1e6 // 转换为毫秒
	}
	
	return &GlobalMetrics{
		TotalSent:   totalSent,
		TotalFailed: totalFailed,
		AvgLatency:  avgLatency,
	}
}

// GetMetricsByType 按类型获取指标
func (m *ChannelMetrics) GetMetricsByType() map[channel.ChannelType]*TypeMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[channel.ChannelType]*TypeMetrics)
	
	// 获取所有渠道类型
	types := make(map[channel.ChannelType]bool)
	for channelType := range m.messagesSent {
		types[channelType] = true
	}
	for channelType := range m.messagesFailed {
		types[channelType] = true
	}
	for channelType := range m.latencies {
		types[channelType] = true
	}
	
	// 为每种类型计算指标
	for channelType := range types {
		sent := m.messagesSent[channelType]
		failed := m.messagesFailed[channelType]
		total := sent + failed
		
		var successRate float64
		if total > 0 {
			successRate = float64(sent) / float64(total) * 100
		}
		
		var avgLatency float64
		latencies := m.latencies[channelType]
		if len(latencies) > 0 {
			var totalLatency time.Duration
			for _, latency := range latencies {
				totalLatency += latency
			}
			avgLatency = float64(totalLatency.Nanoseconds()) / float64(len(latencies)) / 1e6
		}
		
		result[channelType] = &TypeMetrics{
			ChannelType:     channelType,
			MessagesSent:    sent,
			MessagesFailed:  failed,
			SuccessRate:     successRate,
			AvgLatency:      avgLatency,
			ChannelsCreated: m.channelsCreated[channelType],
			ChannelsDeleted: m.channelsDeleted[channelType],
		}
	}
	
	return result
}

// TypeMetrics 类型指标
type TypeMetrics struct {
	ChannelType     channel.ChannelType `json:"channel_type"`
	MessagesSent    int64               `json:"messages_sent"`
	MessagesFailed  int64               `json:"messages_failed"`
	SuccessRate     float64             `json:"success_rate"`
	AvgLatency      float64             `json:"avg_latency"`
	ChannelsCreated int64               `json:"channels_created"`
	ChannelsDeleted int64               `json:"channels_deleted"`
}

// Reset 重置指标
func (m *ChannelMetrics) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.messagesSent = make(map[channel.ChannelType]int64)
	m.messagesFailed = make(map[channel.ChannelType]int64)
	m.latencies = make(map[channel.ChannelType][]time.Duration)
	m.channelsCreated = make(map[channel.ChannelType]int64)
	m.channelsDeleted = make(map[channel.ChannelType]int64)
	m.lastUpdated = time.Now()
}

// GetLastUpdated 获取最后更新时间
func (m *ChannelMetrics) GetLastUpdated() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastUpdated
}