package common

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SimpleMetrics 是一个简单的指标收集实现，用于示例
type SimpleMetrics struct {
	// 计数器
	requestCount  int64 // 请求总数
	successCount  int64 // 成功请求数
	failureCount  int64 // 失败请求数
	retryCount    int64 // 重试次数
	failoverCount int64 // 故障转移次数
	circuitBreaks int64 // 熔断次数

	// 实例级别统计
	instanceStats map[string]*InstanceStat

	// 响应时间
	responseTimes []time.Duration // 响应时间记录
	maxSamples    int             // 最大样本数

	// 事件记录
	events    []string // 事件记录
	maxEvents int      // 最大事件数
	mu        sync.RWMutex
}

// InstanceStat 单个实例的统计信息
type InstanceStat struct {
	Address      string        // 实例地址
	RequestCount int64         // 请求数
	SuccessCount int64         // 成功请求数
	FailureCount int64         // 失败请求数
	AvgLatency   time.Duration // 平均延迟
	CircuitState string        // 熔断器状态
	LastErr      string        // 最后一次错误
	LastSuccess  time.Time     // 最后一次成功时间
	LastFailure  time.Time     // 最后一次失败时间
}

// NewSimpleMetrics 创建一个新的简单指标收集器
func NewSimpleMetrics(maxSamples int, maxEvents int) *SimpleMetrics {
	return &SimpleMetrics{
		instanceStats: make(map[string]*InstanceStat),
		responseTimes: make([]time.Duration, 0, maxSamples),
		events:        make([]string, 0, maxEvents),
		maxSamples:    maxSamples,
		maxEvents:     maxEvents,
	}
}

// RecordRequest 记录请求
func (m *SimpleMetrics) RecordRequest(instance string, duration time.Duration, err error) {
	atomic.AddInt64(&m.requestCount, 1)

	if err == nil {
		atomic.AddInt64(&m.successCount, 1)
	} else {
		atomic.AddInt64(&m.failureCount, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新实例统计信息
	stat, exists := m.instanceStats[instance]
	if !exists {
		stat = &InstanceStat{
			Address: instance,
		}
		m.instanceStats[instance] = stat
	}

	stat.RequestCount++

	if err == nil {
		stat.SuccessCount++
		stat.LastSuccess = time.Now()
	} else {
		stat.FailureCount++
		stat.LastErr = err.Error()
		stat.LastFailure = time.Now()
	}

	// 更新响应时间
	if len(m.responseTimes) >= m.maxSamples {
		m.responseTimes = m.responseTimes[1:]
	}
	m.responseTimes = append(m.responseTimes, duration)

	// 更新实例平均延迟
	// 简单实现，实际应该用更复杂的算法
	stat.AvgLatency = duration
}

// RecordRetry 记录重试
func (m *SimpleMetrics) RecordRetry(instance string) {
	atomic.AddInt64(&m.retryCount, 1)
	m.addEvent(fmt.Sprintf("Retry: %s", instance))
}

// RecordFailover 记录故障转移
func (m *SimpleMetrics) RecordFailover(fromInstance, toInstance string) {
	atomic.AddInt64(&m.failoverCount, 1)
	m.addEvent(fmt.Sprintf("Failover: %s -> %s", fromInstance, toInstance))
}

// RecordCircuitBreak 记录熔断器状态变更
func (m *SimpleMetrics) RecordCircuitBreak(instance string, state string) {
	atomic.AddInt64(&m.circuitBreaks, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新实例熔断器状态
	if stat, exists := m.instanceStats[instance]; exists {
		stat.CircuitState = state
	}

	m.addEvent(fmt.Sprintf("Circuit Break: %s -> %s", instance, state))
}

// GetRequestCount 获取请求数统计
func (m *SimpleMetrics) GetRequestCount() (total, success, failure int64) {
	return atomic.LoadInt64(&m.requestCount),
		atomic.LoadInt64(&m.successCount),
		atomic.LoadInt64(&m.failureCount)
}

// GetRetryCount 获取重试次数
func (m *SimpleMetrics) GetRetryCount() int64 {
	return atomic.LoadInt64(&m.retryCount)
}

// GetFailoverCount 获取故障转移次数
func (m *SimpleMetrics) GetFailoverCount() int64 {
	return atomic.LoadInt64(&m.failoverCount)
}

// GetCircuitBreaks 获取熔断次数
func (m *SimpleMetrics) GetCircuitBreaks() int64 {
	return atomic.LoadInt64(&m.circuitBreaks)
}

// GetInstanceStats 获取所有实例统计信息
func (m *SimpleMetrics) GetInstanceStats() map[string]*InstanceStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免并发修改问题
	result := make(map[string]*InstanceStat, len(m.instanceStats))
	for k, v := range m.instanceStats {
		statCopy := *v
		result[k] = &statCopy
	}
	return result
}

// GetAvgResponseTime 获取平均响应时间
func (m *SimpleMetrics) GetAvgResponseTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range m.responseTimes {
		total += t
	}
	return total / time.Duration(len(m.responseTimes))
}

// GetEvents 获取记录的事件
func (m *SimpleMetrics) GetEvents() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]string, len(m.events))
	copy(events, m.events)
	return events
}

// 添加事件记录
func (m *SimpleMetrics) addEvent(event string) {
	// 锁应该已经被调用方获取
	if len(m.events) >= m.maxEvents {
		m.events = m.events[1:]
	}
	m.events = append(m.events, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), event))
}

// Reset 重置所有指标
func (m *SimpleMetrics) Reset() {
	atomic.StoreInt64(&m.requestCount, 0)
	atomic.StoreInt64(&m.successCount, 0)
	atomic.StoreInt64(&m.failureCount, 0)
	atomic.StoreInt64(&m.retryCount, 0)
	atomic.StoreInt64(&m.failoverCount, 0)
	atomic.StoreInt64(&m.circuitBreaks, 0)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.instanceStats = make(map[string]*InstanceStat)
	m.responseTimes = make([]time.Duration, 0, m.maxSamples)
	m.events = make([]string, 0, m.maxEvents)
}
