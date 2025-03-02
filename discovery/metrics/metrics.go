package metrics

import (
	"context"
	"time"
)

// ResponseMetric 响应时间指标数据
type ResponseMetric struct {
	ServiceName string            // 服务名称
	MethodName  string            // 方法名称
	Instance    string            // 实例地址
	Duration    time.Duration     // 响应时间
	Status      string            // 调用状态 (success/error)
	Timestamp   time.Time         // 记录时间戳
	Tags        map[string]string // 额外的标签信息
}

// Metrics 指标收集接口
type Metrics interface {
	// RecordResponse 记录响应时间
	RecordResponse(ctx context.Context, metric *ResponseMetric) error

	// GetLatency 获取指定服务实例的平均响应时间
	GetLatency(ctx context.Context, service, instance string) (time.Duration, error)

	// GetServiceLatency 获取服务所有实例的平均响应时间
	GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error)

	// RecordFailover 记录故障转移事件
	RecordFailover(ctx context.Context, service, fromInstance, toInstance string) error

	// RecordCircuitBreak 记录熔断事件
	RecordCircuitBreak(ctx context.Context, service, instance string, state string) error

	// RecordRetry 记录重试事件
	RecordRetry(ctx context.Context, service, instance string, attempt int) error

	// GetFailoverRate 获取故障转移率
	GetFailoverRate(ctx context.Context, service string) (float64, error)

	// Close 关闭指标收集器
	Close() error
}

// NoopMetrics 空操作指标收集器实现，用于测试或禁用指标收集
type NoopMetrics struct{}

func (m *NoopMetrics) RecordResponse(ctx context.Context, metric *ResponseMetric) error {
	return nil
}

func (m *NoopMetrics) GetLatency(ctx context.Context, service, instance string) (time.Duration, error) {
	return 0, nil
}

func (m *NoopMetrics) GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error) {
	return make(map[string]time.Duration), nil
}

func (m *NoopMetrics) RecordFailover(ctx context.Context, service, fromInstance, toInstance string) error {
	return nil
}

func (m *NoopMetrics) RecordCircuitBreak(ctx context.Context, service, instance string, state string) error {
	return nil
}

func (m *NoopMetrics) RecordRetry(ctx context.Context, service, instance string, attempt int) error {
	return nil
}

func (m *NoopMetrics) GetFailoverRate(ctx context.Context, service string) (float64, error) {
	return 0, nil
}

func (m *NoopMetrics) Close() error {
	return nil
}

// NewNoopMetrics 创建一个空操作指标收集器
func NewNoopMetrics() Metrics {
	return &NoopMetrics{}
}
