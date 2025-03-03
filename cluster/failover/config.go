package failover

import (
	"time"
)

// Config 故障转移配置
type Config struct {
	// 重试相关配置
	MaxRetries      int           // 最大重试次数
	RetryInterval   time.Duration // 重试间隔基准时间
	MaxRetryDelay   time.Duration // 最大重试延迟时间
	RetryBackoff    float64       // 重试退避指数
	RetryJitter     float64       // 重试抖动因子(0-1)
	RetryableErrors []string      // 可重试的错误类型列表
	RetryStrategy   string        // 重试策略: simple, exponential, littered

	// 熔断相关配置
	CircuitBreakThreshold    int           // 熔断阈值，连续失败次数
	CircuitBreakTimeout      time.Duration // 熔断超时时间
	HalfOpenMaxCalls         int           // 半开状态下允许的最大调用次数
	HalfOpenSuccessThreshold float64       // 半开状态下成功率阈值(0-1)

	// 故障检测配置
	FailureDetectionTime time.Duration // 故障检测时间窗口
	FailureThreshold     int           // 故障阈值次数
	SuccessThreshold     int           // 成功阈值次数
	ConnectionTimeout    time.Duration // 连接超时时间
	RequestTimeout       time.Duration // 请求超时时间

	// 恢复策略配置
	RecoveryInterval  time.Duration // 恢复检查间隔
	RecoveryTimeout   time.Duration // 恢复超时时间
	RecoveryStrategy  string        // 恢复策略：immediate(立即), gradual(渐进), probing(探测)
	RecoveryThreshold int           // 恢复阈值，成功次数

	// 通用配置
	EnableMetrics    bool   // 是否启用指标收集
	FailoverStrategy string // 故障转移策略：next(下一个), random(随机), best(最优)
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	// 重试相关默认配置
	MaxRetries:      3,
	RetryInterval:   100 * time.Millisecond,
	MaxRetryDelay:   30 * time.Second,
	RetryBackoff:    2.0,
	RetryJitter:     0.2,
	RetryableErrors: []string{"timeout", "connection_refused", "service_unavailable"},
	RetryStrategy:   "jittered", // 默认使用带抖动的重试策略

	// 熔断相关默认配置
	CircuitBreakThreshold:    5,
	CircuitBreakTimeout:      30 * time.Second,
	HalfOpenMaxCalls:         3,
	HalfOpenSuccessThreshold: 0.5,

	// 故障检测默认配置
	FailureDetectionTime: 10 * time.Second,
	FailureThreshold:     3,
	SuccessThreshold:     2,
	ConnectionTimeout:    3 * time.Second,
	RequestTimeout:       5 * time.Second,

	// 恢复策略默认配置
	RecoveryInterval:  5 * time.Second,
	RecoveryTimeout:   60 * time.Second,
	RecoveryStrategy:  "gradual",
	RecoveryThreshold: 2,

	// 通用默认配置
	EnableMetrics:    true,
	FailoverStrategy: "next",
}

// Option 配置选项函数类型
type Option func(*Config)

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		if maxRetries >= 0 {
			c.MaxRetries = maxRetries
		}
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) Option {
	return func(c *Config) {
		if interval > 0 {
			c.RetryInterval = interval
		}
	}
}

// WithRetryBackoff 设置重试退避策略
func WithRetryBackoff(backoff float64, maxDelay time.Duration) Option {
	return func(c *Config) {
		if backoff >= 1.0 {
			c.RetryBackoff = backoff
		}
		if maxDelay > 0 {
			c.MaxRetryDelay = maxDelay
		}
	}
}

// WithRetryJitter 设置重试抖动
func WithRetryJitter(jitter float64) Option {
	return func(c *Config) {
		if jitter >= 0 && jitter <= 1 {
			c.RetryJitter = jitter
		}
	}
}

// WithRetryableErrors 设置可重试的错误类型
func WithRetryableErrors(errors []string) Option {
	return func(c *Config) {
		c.RetryableErrors = errors
	}
}

// WithRetryStrategy 设置重试策略
func WithRetryStrategy(strategy string) Option {
	return func(c *Config) {
		validStrategies := map[string]bool{
			"simple":      true,
			"exponential": true,
			"jittered":    true,
		}
		if validStrategies[strategy] {
			c.RetryStrategy = strategy
		}
	}
}

// WithCircuitBreaker 设置熔断器配置
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *Config) {
		if threshold > 0 {
			c.CircuitBreakThreshold = threshold
		}
		if timeout > 0 {
			c.CircuitBreakTimeout = timeout
		}
	}
}

// WithHalfOpenConfig 设置半开状态配置
func WithHalfOpenConfig(maxCalls int, successThreshold float64) Option {
	return func(c *Config) {
		if maxCalls > 0 {
			c.HalfOpenMaxCalls = maxCalls
		}
		if successThreshold > 0 && successThreshold <= 1 {
			c.HalfOpenSuccessThreshold = successThreshold
		}
	}
}

// WithDetectionConfig 设置故障检测配置
func WithDetectionConfig(detectionTime time.Duration, failureThreshold int, successThreshold int) Option {
	return func(c *Config) {
		if detectionTime > 0 {
			c.FailureDetectionTime = detectionTime
		}
		if failureThreshold > 0 {
			c.FailureThreshold = failureThreshold
		}
		if successThreshold > 0 {
			c.SuccessThreshold = successThreshold
		}
	}
}

// WithTimeouts 设置超时时间
func WithTimeouts(connTimeout, reqTimeout time.Duration) Option {
	return func(c *Config) {
		if connTimeout > 0 {
			c.ConnectionTimeout = connTimeout
		}
		if reqTimeout > 0 {
			c.RequestTimeout = reqTimeout
		}
	}
}

// WithRecoveryStrategy 设置恢复策略
func WithRecoveryStrategy(strategy string, interval time.Duration) Option {
	return func(c *Config) {
		validStrategies := map[string]bool{
			"immediate": true,
			"gradual":   true,
			"probing":   true,
		}
		if validStrategies[strategy] {
			c.RecoveryStrategy = strategy
		}
		if interval > 0 {
			c.RecoveryInterval = interval
		}
	}
}

// WithRecoveryThreshold 设置恢复阈值
func WithRecoveryThreshold(threshold int, timeout time.Duration) Option {
	return func(c *Config) {
		if threshold > 0 {
			c.RecoveryThreshold = threshold
		}
		if timeout > 0 {
			c.RecoveryTimeout = timeout
		}
	}
}

// WithFailoverStrategy 设置故障转移策略
func WithFailoverStrategy(strategy string) Option {
	return func(c *Config) {
		validStrategies := map[string]bool{
			"next":   true,
			"random": true,
			"best":   true,
		}
		if validStrategies[strategy] {
			c.FailoverStrategy = strategy
		}
	}
}

// WithMetricsEnabled 设置是否启用指标收集
func WithMetricsEnabled(enabled bool) Option {
	return func(c *Config) {
		c.EnableMetrics = enabled
	}
}

// NewConfig 创建新的配置
func NewConfig(opts ...Option) *Config {
	config := *DefaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return &config
}
