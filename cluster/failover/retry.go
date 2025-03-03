package failover

import (
	"context"
	"math"
	"math/rand"
	"strings"
	"time"
)

// BaseRetryPolicy 基础重试策略实现
type BaseRetryPolicy struct {
	maxAttempts int
}

// ShouldRetry 判断是否应该重试
func (p *BaseRetryPolicy) ShouldRetry(_ context.Context, attempt int, err error) bool {
	if err == nil {
		return false
	}
	return attempt < p.MaxAttempts()
}

// NextBackoff 下一次重试的等待时间
func (p *BaseRetryPolicy) NextBackoff(_ int) time.Duration {
	return 0 // 基础策略不等待，立即重试
}

// MaxAttempts 返回最大重试次数
func (p *BaseRetryPolicy) MaxAttempts() int {
	return p.maxAttempts
}

// 简单重试策略
type SimpleRetryPolicy struct {
	BaseRetryPolicy
	interval   time.Duration
	errorTypes []string
}

// NewSimpleRetryPolicy 创建简单重试策略
func NewSimpleRetryPolicy(maxAttempts int, interval time.Duration, retryableErrors []string) *SimpleRetryPolicy {
	return &SimpleRetryPolicy{
		BaseRetryPolicy: BaseRetryPolicy{maxAttempts: maxAttempts},
		interval:        interval,
		errorTypes:      retryableErrors,
	}
}

// ShouldRetry 判断是否应该重试
func (p *SimpleRetryPolicy) ShouldRetry(ctx context.Context, attempt int, err error) bool {
	// 确保这里返回 true，直到达到最大重试次数
	if err == nil || attempt >= p.MaxAttempts() {
		return false
	}

	// 当有错误且尚未达到最大重试次数时，应返回 true
	return true
}

// NextBackoff 下一次重试的等待时间
func (p *SimpleRetryPolicy) NextBackoff(_ int) time.Duration {
	return p.interval
}

// 指数退避重试策略
type ExponentialBackoffRetryPolicy struct {
	BaseRetryPolicy
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	errorTypes      []string
}

// NewExponentialBackoffRetryPolicy 创建指数退避重试策略
func NewExponentialBackoffRetryPolicy(maxAttempts int, initialInterval, maxInterval time.Duration, multiplier float64, retryableErrors []string) *ExponentialBackoffRetryPolicy {
	return &ExponentialBackoffRetryPolicy{
		BaseRetryPolicy: BaseRetryPolicy{maxAttempts: maxAttempts},
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
		errorTypes:      retryableErrors,
	}
}

// ShouldRetry 判断是否应该重试
func (p *ExponentialBackoffRetryPolicy) ShouldRetry(ctx context.Context, attempt int, err error) bool {
	if err == nil || attempt >= p.maxAttempts {
		return false
	}

	// 如果没有指定可重试的错误类型，则所有错误都可重试
	if len(p.errorTypes) == 0 {
		return true
	}

	// 检查错误类型是否在可重试列表中
	errMsg := err.Error()
	for _, errType := range p.errorTypes {
		if strings.Contains(errMsg, errType) {
			return true
		}
	}

	return false
}

// NextBackoff 下一次重试的等待时间（指数增长）
func (p *ExponentialBackoffRetryPolicy) NextBackoff(attempt int) time.Duration {
	// 计算指数退避时间
	interval := p.initialInterval * time.Duration(math.Pow(p.multiplier, float64(attempt)))

	// 确保不超过最大间隔
	if interval > p.maxInterval {
		interval = p.maxInterval
	}

	return interval
}

// 带抖动的重试策略（避免重试风暴）
type JitteredRetryPolicy struct {
	BaseRetryPolicy
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	jitterFactor    float64
	errorTypes      []string
	rnd             *rand.Rand
}

// NewJitteredRetryPolicy 创建带抖动的重试策略
func NewJitteredRetryPolicy(maxAttempts int, initialInterval, maxInterval time.Duration, multiplier float64, jitterFactor float64, retryableErrors []string) *JitteredRetryPolicy {
	src := rand.NewSource(time.Now().UnixNano())
	return &JitteredRetryPolicy{
		BaseRetryPolicy: BaseRetryPolicy{maxAttempts: maxAttempts},
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
		jitterFactor:    jitterFactor,
		errorTypes:      retryableErrors,
		rnd:             rand.New(src),
	}
}

// ShouldRetry 判断是否应该重试
func (p *JitteredRetryPolicy) ShouldRetry(ctx context.Context, attempt int, err error) bool {
	if err == nil || attempt >= p.maxAttempts {
		return false
	}

	// 如果没有指定可重试的错误类型，则所有错误都可重试
	if len(p.errorTypes) == 0 {
		return true
	}

	// 检查错误类型是否在可重试列表中
	errMsg := err.Error()
	for _, errType := range p.errorTypes {
		if strings.Contains(errMsg, errType) {
			return true
		}
	}

	return false
}

// NextBackoff 下一次重试的等待时间（带随机抖动）
func (p *JitteredRetryPolicy) NextBackoff(attempt int) time.Duration {
	// 计算基础退避时间
	base := p.initialInterval * time.Duration(math.Pow(p.multiplier, float64(attempt)))

	// 确保不超过最大间隔
	if base > p.maxInterval {
		base = p.maxInterval
	}

	// 添加随机抖动，范围是 [base*(1-jitterFactor), base*(1+jitterFactor)]
	jitter := p.rnd.Float64()*2*p.jitterFactor - p.jitterFactor // -jitterFactor 到 +jitterFactor
	interval := time.Duration(float64(base) * (1.0 + jitter))

	return interval
}

// ContextBasedRetryPolicy 基于上下文的重试策略
type ContextBasedRetryPolicy struct {
	BaseRetryPolicy
	delegate RetryPolicy
}

// NewContextBasedRetryPolicy 创建基于上下文的重试策略
func NewContextBasedRetryPolicy(maxAttempts int, delegate RetryPolicy) *ContextBasedRetryPolicy {
	if delegate == nil {
		delegate = &SimpleRetryPolicy{
			BaseRetryPolicy: BaseRetryPolicy{maxAttempts: maxAttempts},
			interval:        time.Millisecond * 100,
		}
	}

	return &ContextBasedRetryPolicy{
		BaseRetryPolicy: BaseRetryPolicy{maxAttempts: maxAttempts},
		delegate:        delegate,
	}
}

// ShouldRetry 判断是否应该重试
func (p *ContextBasedRetryPolicy) ShouldRetry(ctx context.Context, attempt int, err error) bool {
	if err == nil || attempt >= p.maxAttempts {
		return false
	}

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return false
	}

	// 委托给内部策略
	return p.delegate.ShouldRetry(ctx, attempt, err)
}

// NextBackoff 下一次重试的等待时间
func (p *ContextBasedRetryPolicy) NextBackoff(attempt int) time.Duration {
	return p.delegate.NextBackoff(attempt)
}

// MaxAttempts 返回最大重试次数
func (p *ContextBasedRetryPolicy) MaxAttempts() int {
	return p.maxAttempts
}

// 工厂函数，根据配置创建重试策略
func NewRetryPolicy(config *Config) RetryPolicy {
	switch config.RetryStrategy {
	case "simple":
		return NewSimpleRetryPolicy(
			config.MaxRetries,
			config.RetryInterval,
			config.RetryableErrors,
		)
	case "exponential":
		return NewExponentialBackoffRetryPolicy(
			config.MaxRetries,
			config.RetryInterval,
			config.MaxRetryDelay,
			config.RetryBackoff,
			config.RetryableErrors,
		)
	case "jittered":
		return NewJitteredRetryPolicy(
			config.MaxRetries,
			config.RetryInterval,
			config.MaxRetryDelay,
			config.RetryBackoff,
			config.RetryJitter,
			config.RetryableErrors,
		)
	default:
		// 默认使用简单重试策略
		return NewSimpleRetryPolicy(
			config.MaxRetries,
			config.RetryInterval,
			config.RetryableErrors,
		)
	}
}
