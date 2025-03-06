package config

import (
	"runtime"
	"time"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	*CommonConfig // 继承通用配置

	// 网络相关配置
	Address         string        // 服务监听地址
	MaxConnections  int           // 最大连接数
	ReadTimeout     time.Duration // 读取超时
	WriteTimeout    time.Duration // 写入超时
	ShutdownTimeout time.Duration // 优雅关闭超时
	MaxHeaderBytes  int           // 最大请求头大小

	// 注册中心配置
	RegisterTTL      int64             // 服务注册租约时间（秒）
	RegisterInterval time.Duration     // 服务注册间隔
	EnableRegistry   bool              // 是否启用服务注册
	ServiceName      string            // 服务名称
	ServiceVersion   string            // 服务版本
	ServiceWeight    int               // 服务权重
	Metadata         map[string]string // 服务元数据

	// 处理器配置
	WorkerPoolSize   int           // 工作线程池大小
	MaxRequestSize   int           // 最大请求大小(字节)
	MaxConcurrent    int           // 最大并发请求数
	SlowRequestTime  time.Duration // 慢请求阈值
	EnableAccessLog  bool          // 是否启用访问日志
	EnableMetricsLog bool          // 是否启用指标日志
	MetricsLogPeriod time.Duration // 指标日志周期
}

// DefaultServerConfig 服务器默认配置
var DefaultServerConfig = &ServerConfig{
	CommonConfig: DefaultCommonConfig,

	// 网络相关默认配置
	Address:         ":8000",
	MaxConnections:  1000,
	ReadTimeout:     time.Second * 30,
	WriteTimeout:    time.Second * 30,
	ShutdownTimeout: time.Second * 10,
	MaxHeaderBytes:  1 << 20, // 1MB

	// 注册中心默认配置
	RegisterTTL:      30,
	RegisterInterval: time.Second * 10,
	EnableRegistry:   false,
	ServiceName:      "",
	ServiceVersion:   "1.0.0",
	ServiceWeight:    100,
	Metadata:         make(map[string]string),

	// 处理器默认配置
	WorkerPoolSize:   runtime.NumCPU() * 2,
	MaxRequestSize:   4 << 20, // 4MB
	MaxConcurrent:    100,
	SlowRequestTime:  time.Second * 1,
	EnableAccessLog:  true,
	EnableMetricsLog: false,
	MetricsLogPeriod: time.Minute,
}

// ServerOption 服务器配置选项函数类型
type ServerOption func(*ServerConfig)

// WithAddress 设置服务监听地址
func WithAddress(address string) ServerOption {
	return func(c *ServerConfig) {
		c.Address = address
	}
}

// WithNetworkConfig 设置网络相关配置
func WithNetworkConfig(maxConn int, readTimeout, writeTimeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		if maxConn > 0 {
			c.MaxConnections = maxConn
		}
		if readTimeout > 0 {
			c.ReadTimeout = readTimeout
		}
		if writeTimeout > 0 {
			c.WriteTimeout = writeTimeout
		}
	}
}

// WithShutdownTimeout 设置优雅关闭超时
func WithShutdownTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		if timeout > 0 {
			c.ShutdownTimeout = timeout
		}
	}
}

// WithServiceInfo 设置服务信息
func WithServiceInfo(name, version string, weight int) ServerOption {
	return func(c *ServerConfig) {
		if name != "" {
			c.ServiceName = name
		}
		if version != "" {
			c.ServiceVersion = version
		}
		if weight > 0 {
			c.ServiceWeight = weight
		}
	}
}

// WithServiceMetadata 设置服务元数据
func WithServiceMetadata(metadata map[string]string) ServerOption {
	return func(c *ServerConfig) {
		if len(metadata) > 0 {
			c.Metadata = metadata
		}
	}
}

// WithWorkerPoolSize 设置工作线程池大小
func WithWorkerPoolSize(size int) ServerOption {
	return func(c *ServerConfig) {
		if size > 0 {
			c.WorkerPoolSize = size
		}
	}
}

// WithConcurrencyLimit 设置并发限制
func WithConcurrencyLimit(maxConcurrent, maxRequestSize int) ServerOption {
	return func(c *ServerConfig) {
		if maxConcurrent > 0 {
			c.MaxConcurrent = maxConcurrent
		}
		if maxRequestSize > 0 {
			c.MaxRequestSize = maxRequestSize
		}
	}
}

// WithRegistryConfig 设置注册中心配置
func WithRegistryConfig(enable bool, ttl int64, interval time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.EnableRegistry = enable
		if ttl > 0 {
			c.RegisterTTL = ttl
		}
		if interval > 0 {
			c.RegisterInterval = interval
		}
	}
}

// WithLoggingConfig 设置日志配置
func WithLoggingConfig(enableAccess, enableMetrics bool, metricsPeriod time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.EnableAccessLog = enableAccess
		c.EnableMetricsLog = enableMetrics
		if metricsPeriod > 0 {
			c.MetricsLogPeriod = metricsPeriod
		}
	}
}

// WithSlowRequestTime 设置慢请求阈值
func WithSlowRequestTime(duration time.Duration) ServerOption {
	return func(c *ServerConfig) {
		if duration > 0 {
			c.SlowRequestTime = duration
		}
	}
}

// NewServerConfig 创建服务器配置
func NewServerConfig(options ...ServerOption) *ServerConfig {
	// 创建默认配置的副本
	config := *DefaultServerConfig

	// 应用配置选项
	for _, option := range options {
		option(&config)
	}

	return &config
}

// Apply 应用通用配置选项
func (c *ServerConfig) Apply(options ...CommonOption) *ServerConfig {
	for _, option := range options {
		option(c.CommonConfig)
	}
	return c
}

// Init 初始化配置后的钩子
func (c *ServerConfig) Init() {
	// 初始化通用配置
	c.CommonConfig.Init()

	// 服务器特定初始化逻辑
	if c.ServiceName == "" {
		// 如果没有设置服务名，设置一个默认名称
		c.ServiceName = "fyerrpc-server"
	}
}
