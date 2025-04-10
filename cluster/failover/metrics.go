package failover

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// SimpleInstanceMonitor 简单实例监控器实现
type SimpleInstanceMonitor struct {
	stats map[string]*InstanceStats
	mu    sync.RWMutex
}

// NewInstanceMonitor 创建实例监控器
func NewInstanceMonitor(config *Config) InstanceMonitor {
	return &SimpleInstanceMonitor{
		stats: make(map[string]*InstanceStats),
	}
}

// ReportSuccess 报告成功请求
func (m *SimpleInstanceMonitor) ReportSuccess(ctx context.Context, instance *naming.Instance, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, ok := m.stats[instance.ID]
	if !ok {
		stats = &InstanceStats{
			LastResponseTime: duration,
		}
		m.stats[instance.ID] = stats
	}

	atomic.AddInt64(&stats.TotalRequests, 1)
	atomic.AddInt64(&stats.SuccessRequests, 1)
	stats.LastResponseTime = duration

	// 更新平均响应时间
	if stats.AvgResponseTime == 0 {
		stats.AvgResponseTime = duration
	} else {
		oldAvg := stats.AvgResponseTime.Nanoseconds()
		newAvg := oldAvg + (duration.Nanoseconds()-oldAvg)/stats.TotalRequests
		stats.AvgResponseTime = time.Duration(newAvg)
	}

	atomic.StoreInt32(&stats.ConsecutiveFailures, 0)
}

// ReportFailure 报告失败请求
func (m *SimpleInstanceMonitor) ReportFailure(ctx context.Context, instance *naming.Instance, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, ok := m.stats[instance.ID]
	if !ok {
		stats = &InstanceStats{
			LastFailure: time.Now(),
		}
		m.stats[instance.ID] = stats
	}

	atomic.AddInt64(&stats.TotalRequests, 1)
	atomic.AddInt64(&stats.FailureRequests, 1)
	stats.LastFailure = time.Now()
	atomic.AddInt32((*int32)(&stats.ConsecutiveFailures), 1)
}

// GetStatus 获取实例状态
func (m *SimpleInstanceMonitor) GetStatus(instance *naming.Instance) Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, ok := m.stats[instance.ID]
	if !ok {
		return StatusHealthy // 默认为健康状态
	}

	// 如果连续失败次数过多，标记为不健康
	if stats.ConsecutiveFailures >= 3 {
		return StatusUnhealthy
	}

	// 如果最近有失败但未达到阈值，标记为可疑
	if stats.ConsecutiveFailures > 0 {
		// 如果最近失败时间在30秒内，则判定为可疑
		if time.Since(stats.LastFailure) < time.Second*30 && !stats.LastFailure.IsZero() {
			return StatusSuspect
		}
	}

	return StatusHealthy
}

// GetStats 获取实例统计信息
func (m *SimpleInstanceMonitor) GetStats(instance *naming.Instance) (*InstanceStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, ok := m.stats[instance.ID]
	if !ok {
		return &InstanceStats{}, nil
	}

	// 返回副本，避免并发修改问题
	statsCopy := *stats
	return &statsCopy, nil
}

// MetricsCollector 指标收集器接口，用于收集和输出故障转移相关指标
type MetricsCollector interface {
	// IncrementRetries 增加重试计数
	IncrementRetries(service string)

	// IncrementFailovers 增加故障转移计数
	IncrementFailovers(service string, fromInstance, toInstance string)

	// RecordDetection 记录故障检测事件
	RecordDetection(instance *naming.Instance, status Status)

	// RecordBreaker 记录熔断器状态变更
	RecordBreaker(instance *naming.Instance, state State)

	// RecordLatency 记录请求延迟
	RecordLatency(instance *naming.Instance, latency time.Duration)
}

// NoopMetricsCollector 空操作指标收集器
type NoopMetricsCollector struct{}

// IncrementRetries 增加重试计数
func (c *NoopMetricsCollector) IncrementRetries(service string) {}

// IncrementFailovers 增加故障转移计数
func (c *NoopMetricsCollector) IncrementFailovers(service string, fromInstance, toInstance string) {}

// RecordDetection 记录故障检测事件
func (c *NoopMetricsCollector) RecordDetection(instance *naming.Instance, status Status) {}

// RecordBreaker 记录熔断器状态变更
func (c *NoopMetricsCollector) RecordBreaker(instance *naming.Instance, state State) {}

// RecordLatency 记录请求延迟
func (c *NoopMetricsCollector) RecordLatency(instance *naming.Instance, latency time.Duration) {}

// InMemoryMetricsCollector 内存指标收集器实现
type InMemoryMetricsCollector struct {
	retryCount        map[string]int64
	failoverCount     map[string]int64
	detections        map[string]map[Status]int64
	breakerChanges    map[string]map[State]int64
	latencies         map[string][]time.Duration
	maxLatencySamples int
	mu                sync.RWMutex
}

// NewInMemoryMetricsCollector 创建内存指标收集器
func NewInMemoryMetricsCollector(maxLatencySamples int) *InMemoryMetricsCollector {
	if maxLatencySamples <= 0 {
		maxLatencySamples = 100
	}

	return &InMemoryMetricsCollector{
		retryCount:        make(map[string]int64),
		failoverCount:     make(map[string]int64),
		detections:        make(map[string]map[Status]int64),
		breakerChanges:    make(map[string]map[State]int64),
		latencies:         make(map[string][]time.Duration),
		maxLatencySamples: maxLatencySamples,
	}
}

// IncrementRetries 增加重试计数
func (c *InMemoryMetricsCollector) IncrementRetries(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.retryCount[service]++
}

// IncrementFailovers 增加故障转移计数
func (c *InMemoryMetricsCollector) IncrementFailovers(service string, fromInstance, toInstance string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := service + ":" + fromInstance + "->" + toInstance
	c.failoverCount[key]++
}

// RecordDetection 记录故障检测事件
func (c *InMemoryMetricsCollector) RecordDetection(instance *naming.Instance, status Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.detections[instance.ID]; !ok {
		c.detections[instance.ID] = make(map[Status]int64)
	}
	c.detections[instance.ID][status]++
}

// RecordBreaker 记录熔断器状态变更
func (c *InMemoryMetricsCollector) RecordBreaker(instance *naming.Instance, state State) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.breakerChanges[instance.ID]; !ok {
		c.breakerChanges[instance.ID] = make(map[State]int64)
	}
	c.breakerChanges[instance.ID][state]++
}

// RecordLatency 记录请求延迟
func (c *InMemoryMetricsCollector) RecordLatency(instance *naming.Instance, latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	latencies, ok := c.latencies[instance.ID]
	if !ok {
		latencies = make([]time.Duration, 0, c.maxLatencySamples)
	}

	if len(latencies) >= c.maxLatencySamples {
		// 移除最早的样本
		latencies = latencies[1:]
	}

	latencies = append(latencies, latency)
	c.latencies[instance.ID] = latencies
}

// GetRetryCount 获取服务的重试次数
func (c *InMemoryMetricsCollector) GetRetryCount(service string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.retryCount[service]
}

// GetFailoverCount 获取故障转移次数
func (c *InMemoryMetricsCollector) GetFailoverCount(service string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total int64
	prefix := service + ":"
	for key, count := range c.failoverCount {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			total += count
		}
	}

	return total
}

// GetAverageLatency 获取实例的平均响应时间
func (c *InMemoryMetricsCollector) GetAverageLatency(instanceID string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	latencies, ok := c.latencies[instanceID]
	if !ok || len(latencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, latency := range latencies {
		sum += latency
	}

	return sum / time.Duration(len(latencies))
}

// GetDetectionStats 获取故障检测统计信息
func (c *InMemoryMetricsCollector) GetDetectionStats(instanceID string) map[Status]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats, ok := c.detections[instanceID]
	if !ok {
		return make(map[Status]int64)
	}

	// 返回副本，避免并发修改问题
	result := make(map[Status]int64)
	for status, count := range stats {
		result[status] = count
	}

	return result
}

// GetBreakerStats 获取熔断器统计信息
func (c *InMemoryMetricsCollector) GetBreakerStats(instanceID string) map[State]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats, ok := c.breakerChanges[instanceID]
	if !ok {
		return make(map[State]int64)
	}

	// 返回副本，避免并发修改问题
	result := make(map[State]int64)
	for state, count := range stats {
		result[state] = count
	}

	return result
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(config *Config) MetricsCollector {
	if !config.EnableMetrics {
		return &NoopMetricsCollector{}
	}

	return NewInMemoryMetricsCollector(100)
}
