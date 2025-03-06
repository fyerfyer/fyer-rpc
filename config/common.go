package config

import (
	"io"
	"os"
	"time"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/utils"
)

// SerializationType 序列化类型
type SerializationType uint8

const (
	// SerializationJSON JSON序列化
	SerializationJSON SerializationType = SerializationType(protocol.SerializationTypeJSON)
	// SerializationProtobuf Protobuf序列化
	SerializationProtobuf SerializationType = SerializationType(protocol.SerializationTypeProtobuf)
)

// CompressType 压缩类型
type CompressType uint8

const (
	// CompressNone 不压缩
	CompressNone CompressType = CompressType(protocol.CompressTypeNone)
	// CompressGzip Gzip压缩
	CompressGzip CompressType = CompressType(protocol.CompressTypeGzip)
)

// CommonConfig 通用配置选项
type CommonConfig struct {
	// 日志配置
	LogLevel     utils.LogLevel // 日志级别
	LogOutput    io.Writer      // 日志输出
	EnabledDebug bool           // 是否启用调试日志

	// 协议配置
	SerializationType SerializationType // 序列化类型
	CompressType      CompressType      // 压缩类型
	ProtocolVersion   uint8             // 协议版本

	// 超时配置
	DialTimeout    time.Duration // 连接超时
	RequestTimeout time.Duration // 请求超时

	// 重试配置
	MaxRetries     int           // 最大重试次数
	RetryInterval  time.Duration // 重试间隔
	RetryableError []string      // 可重试的错误类型

	// 注册中心配置
	RegistryType     string   // 注册中心类型（etcd, consul, etc.）
	RegistryEndpoint []string // 注册中心地址

	// 指标配置
	EnableMetrics   bool          // 是否启用指标收集
	MetricsInterval time.Duration // 指标收集间隔

	// 元数据，用于存储额外信息
	Metadata map[string]string
}

// DefaultCommonConfig 默认通用配置
var DefaultCommonConfig = &CommonConfig{
	// 日志配置
	LogLevel:     utils.InfoLevel,
	LogOutput:    os.Stdout,
	EnabledDebug: false,

	// 协议配置
	SerializationType: SerializationJSON,
	CompressType:      CompressNone,
	ProtocolVersion:   1,

	// 超时配置
	DialTimeout:    time.Second * 3,
	RequestTimeout: time.Second * 5,

	// 重试配置
	MaxRetries:     3,
	RetryInterval:  time.Millisecond * 100,
	RetryableError: []string{"timeout", "connection_refused"},

	// 注册中心配置
	RegistryType:     "etcd",
	RegistryEndpoint: []string{"localhost:2379"},

	// 指标配置
	EnableMetrics:   false,
	MetricsInterval: time.Second * 15,

	// 元数据
	Metadata: make(map[string]string),
}

// CommonOption 通用配置选项函数类型
type CommonOption func(*CommonConfig)

// WithLogLevel 设置日志级别
func WithLogLevel(level utils.LogLevel) CommonOption {
	return func(c *CommonConfig) {
		c.LogLevel = level
	}
}

// WithLogOutput 设置日志输出
func WithLogOutput(output io.Writer) CommonOption {
	return func(c *CommonConfig) {
		c.LogOutput = output
	}
}

// WithDebug 设置是否启用调试日志
func WithDebug(enabled bool) CommonOption {
	return func(c *CommonConfig) {
		c.EnabledDebug = enabled
	}
}

// WithSerialization 设置序列化类型
func WithSerialization(serType SerializationType) CommonOption {
	return func(c *CommonConfig) {
		c.SerializationType = serType
	}
}

// WithCompression 设置压缩类型
func WithCompression(compType CompressType) CommonOption {
	return func(c *CommonConfig) {
		c.CompressType = compType
	}
}

// WithTimeouts 设置超时时间
func WithTimeouts(dialTimeout, requestTimeout time.Duration) CommonOption {
	return func(c *CommonConfig) {
		if dialTimeout > 0 {
			c.DialTimeout = dialTimeout
		}
		if requestTimeout > 0 {
			c.RequestTimeout = requestTimeout
		}
	}
}

// WithRetry 设置重试策略
func WithRetry(maxRetries int, interval time.Duration, retryableErrors []string) CommonOption {
	return func(c *CommonConfig) {
		if maxRetries >= 0 {
			c.MaxRetries = maxRetries
		}
		if interval > 0 {
			c.RetryInterval = interval
		}
		if len(retryableErrors) > 0 {
			c.RetryableError = retryableErrors
		}
	}
}

// WithRegistry 设置注册中心
func WithRegistry(regType string, endpoints []string) CommonOption {
	return func(c *CommonConfig) {
		c.RegistryType = regType
		c.RegistryEndpoint = endpoints
	}
}

// WithMetrics 设置指标收集
func WithMetrics(enabled bool, interval time.Duration) CommonOption {
	return func(c *CommonConfig) {
		c.EnableMetrics = enabled
		if interval > 0 {
			c.MetricsInterval = interval
		}
	}
}

// WithMetadata 添加元数据
func WithMetadata(key, value string) CommonOption {
	return func(c *CommonConfig) {
		if c.Metadata == nil {
			c.Metadata = make(map[string]string)
		}
		c.Metadata[key] = value
	}
}

// NewCommonConfig 创建通用配置
func NewCommonConfig(options ...CommonOption) *CommonConfig {
	// 创建默认配置的副本
	config := *DefaultCommonConfig

	// 应用配置选项
	for _, option := range options {
		option(&config)
	}

	return &config
}

// Init 初始化配置后的钩子
func (c *CommonConfig) Init() {
	// 设置全局日志级别
	utils.SetDefaultLogger(utils.NewLogger(c.LogLevel, c.LogOutput))
}
