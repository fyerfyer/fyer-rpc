package common

import (
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	// 基本配置
	ID      string // 服务器ID
	Address string // 服务地址
	Port    int    // 服务端口

	// 故障模拟相关
	FailAfter    int           // 在处理这么多个请求后故障
	FailDuration time.Duration // 故障持续时间
	FailRate     float64       // 随机故障概率 (0-1)
}

// DefaultServerConfig 默认服务器配置
var DefaultServerConfig = &ServerConfig{
	Port:         8000,
	FailAfter:    0, // 默认不故障
	FailDuration: 0, // 默认不故障
	FailRate:     0, // 默认不故障
}

// ClientConfig 客户端配置
type ClientConfig struct {
	// 基本配置
	ServerAddresses []string      // 服务器地址列表
	Timeout         time.Duration // 请求超时时间

	// 故障转移配置
	FailoverConfig *failover.Config // 故障转移配置
	EnableFailover bool             // 是否启用故障转移
}

// DefaultClientConfig 默认客户端配置
var DefaultClientConfig = &ClientConfig{
	Timeout:        5 * time.Second,
	EnableFailover: false,
	FailoverConfig: &failover.Config{
		// 重试相关配置
		MaxRetries:      3,
		RetryInterval:   100 * time.Millisecond,
		RetryBackoff:    1.5,
		RetryJitter:     0.2,
		RetryableErrors: []string{"timeout", "connection_refused", "connection_reset"},
		RetryStrategy:   "jittered",

		// 熔断相关配置
		CircuitBreakThreshold:    3,
		CircuitBreakTimeout:      10 * time.Second,
		HalfOpenMaxCalls:         2,
		HalfOpenSuccessThreshold: 0.5,

		// 故障检测配置
		FailureDetectionTime: 5 * time.Second,
		FailureThreshold:     2,
		SuccessThreshold:     2,
		ConnectionTimeout:    2 * time.Second,
		RequestTimeout:       3 * time.Second,

		// 故障转移策略
		FailoverStrategy: "next",
	},
}

// NewClientConfigWithFailover 创建启用故障转移的客户端配置
func NewClientConfigWithFailover(servers []string) *ClientConfig {
	config := *DefaultClientConfig
	config.ServerAddresses = servers
	config.EnableFailover = true
	return &config
}

// NewClusterServerConfigs 创建集群服务器配置
func NewClusterServerConfigs(basePort int, count int) []*ServerConfig {
	configs := make([]*ServerConfig, count)
	for i := 0; i < count; i++ {
		configs[i] = &ServerConfig{
			ID:       "server-" + string(rune('A'+i)),
			Address:  "localhost",
			Port:     basePort + i,
			FailRate: 0,
		}
	}
	return configs
}
