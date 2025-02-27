package metrics

import (
	"context"
	"time"
)

// ResponseMetric 响应时间指标数据
type ResponseMetric struct {
	ServiceName string        // 服务名称
	MethodName  string        // 方法名称
	Instance    string        // 实例地址
	Duration    time.Duration // 响应时间
	Status      string        // 调用状态 (success/error)
	Timestamp   time.Time     // 记录时间戳
}

// Metrics 指标收集接口
type Metrics interface {
	// RecordResponse 记录响应时间
	RecordResponse(ctx context.Context, metric *ResponseMetric) error

	// GetLatency 获取指定服务实例的平均响应时间
	GetLatency(ctx context.Context, service, instance string) (time.Duration, error)

	// GetServiceLatency 获取服务所有实例的平均响应时间
	GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error)

	// Close 关闭指标收集器
	Close() error
}
