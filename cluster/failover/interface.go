package failover

import (
	"context"
	"errors"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// 错误定义
var (
	ErrNoAvailableInstances = errors.New("no available instances")
	ErrMaxRetriesExceeded   = errors.New("maximum retries exceeded")
	ErrCircuitOpen          = errors.New("circuit breaker is open")
	ErrRequestTimeout       = errors.New("request timeout")
	ErrServiceUnavailable   = errors.New("service is unavailable")
)

// Status 实例状态
type Status int

const (
	StatusHealthy   Status = iota // 健康状态
	StatusUnhealthy               // 不健康状态
	StatusSuspect                 // 可疑状态，可能不健康
	StatusIsolated                // 被隔离状态
)

// FailoverResult 故障转移操作结果
type FailoverResult struct {
	Success     bool             // 是否成功
	Instance    *naming.Instance // 最终使用的实例
	RetryCount  int              // 重试次数
	Duration    time.Duration    // 操作耗时
	Error       error            // 错误信息
	FailedNodes []string         // 失败的节点列表
}

// Detector 故障检测接口
type Detector interface {
	// Detect 检测实例是否健康
	Detect(ctx context.Context, instance *naming.Instance) (Status, error)

	// MarkFailed 标记实例为失败状态
	MarkFailed(ctx context.Context, instance *naming.Instance) error

	// MarkSuccess 标记实例为成功状态
	MarkSuccess(ctx context.Context, instance *naming.Instance) error
}

// RetryPolicy 重试策略接口
type RetryPolicy interface {
	// ShouldRetry 决定是否需要重试
	ShouldRetry(ctx context.Context, attempt int, err error) bool

	// NextBackoff 计算下一次重试的等待时间
	NextBackoff(attempt int) time.Duration

	// MaxAttempts 返回最大重试次数
	MaxAttempts() int
}

// CircuitBreaker 熔断器接口
type CircuitBreaker interface {
	// Allow 判断请求是否允许通过
	Allow(ctx context.Context, instance *naming.Instance) (bool, error)

	// MarkSuccess 标记成功调用
	MarkSuccess(ctx context.Context, instance *naming.Instance) error

	// MarkFailure 标记失败调用
	MarkFailure(ctx context.Context, instance *naming.Instance, err error) error

	// GetState 获取熔断器状态
	GetState(instance *naming.Instance) (State, error)

	// Reset 重置熔断器状态
	Reset(instance *naming.Instance) error
}

// State 熔断器状态
type State int

const (
	StateClosed   State = iota // 关闭状态，允许请求通过
	StateOpen                  // 打开状态，请求被拒绝
	StateHalfOpen              // 半开状态，允许部分请求通过以探测服务是否恢复
)

// RecoveryStrategy 故障恢复策略接口
type RecoveryStrategy interface {
	// CanRecover 判断实例是否可以恢复
	CanRecover(ctx context.Context, instance *naming.Instance) bool

	// Recover 恢复实例
	Recover(ctx context.Context, instance *naming.Instance) error

	// RecoveryDelay 返回恢复尝试间隔
	RecoveryDelay(instance *naming.Instance) time.Duration
}

// InstanceMonitor 实例监控接口
type InstanceMonitor interface {
	// ReportSuccess 报告成功请求
	ReportSuccess(ctx context.Context, instance *naming.Instance, duration time.Duration)

	// ReportFailure 报告失败请求
	ReportFailure(ctx context.Context, instance *naming.Instance, err error)

	// GetStatus 获取实例状态
	GetStatus(instance *naming.Instance) Status

	// GetStats 获取实例统计信息
	GetStats(instance *naming.Instance) (*InstanceStats, error)
}

// InstanceStats 实例统计信息
type InstanceStats struct {
	TotalRequests       int64         // 总请求数
	SuccessRequests     int64         // 成功请求数
	FailureRequests     int64         // 失败请求数
	AvgResponseTime     time.Duration // 平均响应时间
	LastResponseTime    time.Duration // 最近一次响应时间
	LastFailure         time.Time     // 最近一次失败时间
	ConsecutiveFailures int32         // 连续失败次数
}

// FailoverHandler 故障转移处理器接口
type FailoverHandler interface {
	// Execute 执行带故障转移的调用
	Execute(ctx context.Context, instances []*naming.Instance, operation func(context.Context, *naming.Instance) error) (*FailoverResult, error)

	// GetDetector 获取故障检测器
	GetDetector() Detector

	// GetCircuitBreaker 获取熔断器
	GetCircuitBreaker() CircuitBreaker

	// GetRetryPolicy 获取重试策略
	GetRetryPolicy() RetryPolicy

	// GetRecoveryStrategy 获取恢复策略
	GetRecoveryStrategy() RecoveryStrategy

	// GetMonitor 获取实例监控器
	GetMonitor() InstanceMonitor
}
