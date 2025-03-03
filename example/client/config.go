package main

import (
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/example/common"
)

// ClientConfigManager 管理客户端的配置
type ClientConfigManager struct {
	// 基本配置
	config *common.ClientConfig
}

// NewClientConfigManager 创建新的客户端配置管理器
func NewClientConfigManager(config *common.ClientConfig) *ClientConfigManager {
	return &ClientConfigManager{
		config: config,
	}
}

// GetConfig 获取当前配置
func (m *ClientConfigManager) GetConfig() *common.ClientConfig {
	return m.config
}

// EnableFailover 启用故障转移功能
func (m *ClientConfigManager) EnableFailover() {
	m.config.EnableFailover = true
}

// DisableFailover 禁用故障转移功能
func (m *ClientConfigManager) DisableFailover() {
	m.config.EnableFailover = false
}

// SetTimeout 设置请求超时时间
func (m *ClientConfigManager) SetTimeout(timeout time.Duration) {
	m.config.Timeout = timeout
}

// SetFailoverConfig 设置故障转移配置
func (m *ClientConfigManager) SetFailoverConfig(config *failover.Config) {
	m.config.FailoverConfig = config
}

// ConfigureRetry 配置重试策略
func (m *ClientConfigManager) ConfigureRetry(maxRetries int, retryInterval time.Duration, backoff float64, jitter float64) {
	if m.config.FailoverConfig == nil {
		// 创建一个新的配置实例并复制 DefaultConfig 的值
		config := *failover.DefaultConfig
		m.config.FailoverConfig = &config
	}

	m.config.FailoverConfig.MaxRetries = maxRetries
	m.config.FailoverConfig.RetryInterval = retryInterval
	m.config.FailoverConfig.RetryBackoff = backoff
	m.config.FailoverConfig.RetryJitter = jitter
}

// ConfigureCircuitBreaker 配置熔断器
func (m *ClientConfigManager) ConfigureCircuitBreaker(threshold int, timeout time.Duration, halfOpenMaxCalls int, halfOpenSuccessThreshold float64) {
	if m.config.FailoverConfig == nil {
		// 创建一个新的配置实例并复制 DefaultConfig 的值
		config := *failover.DefaultConfig
		m.config.FailoverConfig = &config
	}

	m.config.FailoverConfig.CircuitBreakThreshold = threshold
	m.config.FailoverConfig.CircuitBreakTimeout = timeout
	m.config.FailoverConfig.HalfOpenMaxCalls = halfOpenMaxCalls
	m.config.FailoverConfig.HalfOpenSuccessThreshold = halfOpenSuccessThreshold
}

// GetDefaultFailoverConfig 获取适合演示的默认故障转移配置
func GetDefaultFailoverConfig() *failover.Config {
	return &failover.Config{
		// 重试相关配置
		MaxRetries:      3,
		RetryInterval:   200 * time.Millisecond,
		MaxRetryDelay:   5 * time.Second,
		RetryBackoff:    1.5,
		RetryJitter:     0.2,
		RetryableErrors: []string{"timeout", "connection_refused", "connection_reset", "service_unavailable"},
		RetryStrategy:   "jittered",

		// 熔断相关配置
		CircuitBreakThreshold:    3,
		CircuitBreakTimeout:      5 * time.Second,
		HalfOpenMaxCalls:         2,
		HalfOpenSuccessThreshold: 0.5,

		// 故障检测配置
		FailureDetectionTime: 3 * time.Second,
		FailureThreshold:     2,
		SuccessThreshold:     2,
		ConnectionTimeout:    1 * time.Second,
		RequestTimeout:       2 * time.Second,

		// 恢复策略配置
		RecoveryInterval:  3 * time.Second,
		RecoveryTimeout:   30 * time.Second,
		RecoveryStrategy:  "gradual",
		RecoveryThreshold: 3,

		// 通用配置
		EnableMetrics:    true,
		FailoverStrategy: "next",
	}
}

// CreateDefaultClientConfig 创建带有默认故障转移设置的客户端配置
func CreateDefaultClientConfig(servers []string) *common.ClientConfig {
	config := common.NewClientConfigWithFailover(servers)
	config.FailoverConfig = GetDefaultFailoverConfig()
	return config
}

// CreateClusterClientConfig 创建用于连接服务器集群的客户端配置
func CreateClusterClientConfig(basePort int, count int) *common.ClientConfig {
	// 构建服务器地址列表
	servers := make([]string, count)
	for i := 0; i < count; i++ {
		servers[i] = "localhost:" + string('0'+basePort+i)
	}

	return CreateDefaultClientConfig(servers)
}
