package config

import (
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/discovery/balancer"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	*CommonConfig // 继承通用配置

	// 连接相关配置
	PoolSize        int           // 连接池大小
	MaxIdle         int           // 最大空闲连接数
	IdleTimeout     time.Duration // 空闲连接超时时间
	KeepAlive       bool          // 是否保持连接活跃
	KeepAliveTime   time.Duration // 连接保活时间
	KeepAliveCount  int           // 保活探测次数
	KeepAliveIdle   time.Duration // 连接空闲多久开始保活探测
	ConnectionLimit int           // 单个地址最大连接数

	// 负载均衡相关配置
	LoadBalanceType    balancer.BalancerType // 负载均衡类型
	UpdateInterval     time.Duration         // 服务发现更新间隔
	EnableConsistentLB bool                  // 是否启用一致性负载均衡

	// 故障转移配置
	EnableFailover  bool             // 是否启用故障转移
	FailoverConfig  *failover.Config // 故障转移配置
	FailoverTimeout time.Duration    // 故障转移超时时间

	// 限流相关配置
	MaxConcurrentRequests int           // 最大并发请求数
	MaxQPS                int           // 每秒最大请求数
	RequestTimeout        time.Duration // 请求超时时间
}

// DefaultClientConfig 客户端默认配置
var DefaultClientConfig = &ClientConfig{
	CommonConfig: DefaultCommonConfig,

	// 连接相关默认配置
	PoolSize:        10,
	MaxIdle:         5,
	IdleTimeout:     time.Minute * 5,
	KeepAlive:       true,
	KeepAliveTime:   time.Second * 30,
	KeepAliveCount:  3,
	KeepAliveIdle:   time.Second * 60,
	ConnectionLimit: 100,

	// 负载均衡相关默认配置
	LoadBalanceType:    balancer.Random,
	UpdateInterval:     time.Second * 10,
	EnableConsistentLB: false,

	// 故障转移默认配置
	EnableFailover:  true,
	FailoverConfig:  failover.DefaultConfig,
	FailoverTimeout: time.Second * 10,

	// 限流相关默认配置
	MaxConcurrentRequests: 100,
	MaxQPS:                1000,
	RequestTimeout:        time.Second * 5,
}

// ClientOption 客户端配置选项函数类型
type ClientOption func(*ClientConfig)

// WithPoolConfig 设置连接池配置
func WithPoolConfig(poolSize, maxIdle int, idleTimeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		if poolSize > 0 {
			c.PoolSize = poolSize
		}
		if maxIdle > 0 {
			c.MaxIdle = maxIdle
		}
		if idleTimeout > 0 {
			c.IdleTimeout = idleTimeout
		}
	}
}

// WithKeepAlive 设置连接保活配置
func WithKeepAlive(enabled bool, keepAliveTime, keepAliveIdle time.Duration, keepAliveCount int) ClientOption {
	return func(c *ClientConfig) {
		c.KeepAlive = enabled
		if keepAliveTime > 0 {
			c.KeepAliveTime = keepAliveTime
		}
		if keepAliveIdle > 0 {
			c.KeepAliveIdle = keepAliveIdle
		}
		if keepAliveCount > 0 {
			c.KeepAliveCount = keepAliveCount
		}
	}
}

// WithLoadBalancer 设置负载均衡配置
func WithLoadBalancer(balancerType balancer.BalancerType, updateInterval time.Duration, enableConsistent bool) ClientOption {
	return func(c *ClientConfig) {
		c.LoadBalanceType = balancerType
		if updateInterval > 0 {
			c.UpdateInterval = updateInterval
		}
		c.EnableConsistentLB = enableConsistent
	}
}

// WithFailover 设置故障转移配置
func WithFailover(enabled bool, failoverConfig *failover.Config, timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.EnableFailover = enabled
		if failoverConfig != nil {
			c.FailoverConfig = failoverConfig
		}
		if timeout > 0 {
			c.FailoverTimeout = timeout
		}
	}
}

// WithRateLimit 设置请求限流配置
func WithRateLimit(maxConcurrent, maxQPS int, timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		if maxConcurrent > 0 {
			c.MaxConcurrentRequests = maxConcurrent
		}
		if maxQPS > 0 {
			c.MaxQPS = maxQPS
		}
		if timeout > 0 {
			c.RequestTimeout = timeout
		}
	}
}

// WithConnectionLimit 设置连接限制
func WithConnectionLimit(limit int) ClientOption {
	return func(c *ClientConfig) {
		if limit > 0 {
			c.ConnectionLimit = limit
		}
	}
}

// NewClientConfig 创建客户端配置
func NewClientConfig(options ...ClientOption) *ClientConfig {
	// 创建默认配置的副本
	config := *DefaultClientConfig

	// 应用配置选项
	for _, option := range options {
		option(&config)
	}

	return &config
}

// Apply 应用通用配置选项
func (c *ClientConfig) Apply(options ...CommonOption) *ClientConfig {
	for _, option := range options {
		option(c.CommonConfig)
	}
	return c
}

// Init 初始化配置后的钩子
func (c *ClientConfig) Init() {
	// 初始化通用配置
	c.CommonConfig.Init()

	// 客户端特定初始化逻辑
}
