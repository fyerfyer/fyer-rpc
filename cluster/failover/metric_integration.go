package failover

import (
	"context"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/naming"
)

// MetricsIntegration 故障转移指标集成
type MetricsIntegration struct {
	collector       MetricsCollector  // 内部故障转移指标收集器
	globalCollector metrics.Metrics   // 全局指标收集器
	serviceName     string            // 服务名称
	instanceStats   map[string]*Stats // 实例统计信息缓存
	config          *Config           // 故障转移配置
	mu              sync.RWMutex
}

// Stats 实例统计信息
type Stats struct {
	FailoverCount   int64         // 故障转移次数
	RetryCount      int64         // 重试次数
	LastFailure     time.Time     // 最后一次失败时间
	SuccessRate     float64       // 成功率
	AvgLatency      time.Duration // 平均延迟
	BreakerState    State         // 熔断器状态
	RecoveryAttempt int           // 恢复尝试次数
	LastError       string        // 最后一次错误信息
	UpdatedAt       time.Time     // 更新时间
}

// NewMetricsIntegration 创建新的指标集成
func NewMetricsIntegration(serviceName string, collector MetricsCollector, globalCollector metrics.Metrics, config *Config) *MetricsIntegration {
	return &MetricsIntegration{
		collector:       collector,
		globalCollector: globalCollector,
		serviceName:     serviceName,
		instanceStats:   make(map[string]*Stats),
		config:          config,
	}
}

// RecordFailover 记录故障转移事件
func (m *MetricsIntegration) RecordFailover(ctx context.Context, result *FailoverResult) {
	if result == nil || result.Instance == nil {
		return
	}

	// 更新内部统计信息
	m.mu.Lock()
	instanceID := result.Instance.ID
	stats, ok := m.instanceStats[instanceID]
	if !ok {
		stats = &Stats{UpdatedAt: time.Now()}
		m.instanceStats[instanceID] = stats
	}

	stats.FailoverCount++
	stats.RetryCount += int64(result.RetryCount)
	if result.Error != nil {
		stats.LastFailure = time.Now()
		stats.LastError = result.Error.Error()
	}
	stats.AvgLatency = result.Duration
	stats.UpdatedAt = time.Now()
	m.mu.Unlock()

	// 集成到全局指标系统
	if m.globalCollector != nil && m.config.EnableMetrics {
		// 记录故障转移次数
		m.collector.IncrementFailovers(m.serviceName, "", result.Instance.Address)

		// 记录重试次数
		if result.RetryCount > 0 {
			m.collector.IncrementRetries(m.serviceName)
		}

		// 记录延迟
		m.collector.RecordLatency(result.Instance, result.Duration)

		// 记录响应指标
		status := "success"
		if !result.Success {
			status = "failure"
		}

		m.globalCollector.RecordResponse(ctx, &metrics.ResponseMetric{
			ServiceName: m.serviceName,
			MethodName:  "failover",
			Instance:    result.Instance.Address,
			Duration:    result.Duration,
			Status:      status,
			Timestamp:   time.Now(),
		})
	}
}

// RecordBreakerStateChange 记录熔断器状态变更
func (m *MetricsIntegration) RecordBreakerStateChange(instance *naming.Instance, oldState, newState State) {
	if instance == nil || !m.config.EnableMetrics {
		return
	}

	// 更新内部统计信息
	m.mu.Lock()
	stats, ok := m.instanceStats[instance.ID]
	if !ok {
		stats = &Stats{UpdatedAt: time.Now()}
		m.instanceStats[instance.ID] = stats
	}
	stats.BreakerState = newState
	stats.UpdatedAt = time.Now()
	m.mu.Unlock()

	// 记录状态变更
	m.collector.RecordBreaker(instance, newState)
}

// RecordDetectionResult 记录故障检测结果
func (m *MetricsIntegration) RecordDetectionResult(instance *naming.Instance, status Status, latency time.Duration) {
	if instance == nil || !m.config.EnableMetrics {
		return
	}

	// 更新内部统计信息
	m.mu.Lock()
	stats, ok := m.instanceStats[instance.ID]
	if !ok {
		stats = &Stats{UpdatedAt: time.Now()}
		m.instanceStats[instance.ID] = stats
	}
	m.mu.Unlock()

	// 记录检测结果
	m.collector.RecordDetection(instance, status)
	m.collector.RecordLatency(instance, latency)
}

// RecordRecoveryAttempt 记录恢复尝试
func (m *MetricsIntegration) RecordRecoveryAttempt(instance *naming.Instance, success bool) {
	if instance == nil || !m.config.EnableMetrics {
		return
	}

	// 更新内部统计信息
	m.mu.Lock()
	stats, ok := m.instanceStats[instance.ID]
	if !ok {
		stats = &Stats{UpdatedAt: time.Now()}
		m.instanceStats[instance.ID] = stats
	}
	stats.RecoveryAttempt++
	if success {
		// 重置失败相关指标
		stats.LastError = ""
	}
	stats.UpdatedAt = time.Now()
	m.mu.Unlock()
}

// GetInstanceStats 获取实例统计信息
func (m *MetricsIntegration) GetInstanceStats(instanceID string) *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, ok := m.instanceStats[instanceID]; ok {
		statsCopy := *stats
		return &statsCopy
	}
	return nil
}

// GetAllStats 获取所有实例的统计信息
func (m *MetricsIntegration) GetAllStats() map[string]*Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Stats)
	for id, stats := range m.instanceStats {
		statsCopy := *stats
		result[id] = &statsCopy
	}
	return result
}

// Reset 重置统计信息
func (m *MetricsIntegration) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instanceStats = make(map[string]*Stats)
}
