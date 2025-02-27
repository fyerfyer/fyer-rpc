package registry

import "time"

// Options 注册中心配置选项
type Options struct {
	Endpoints     []string      // 注册中心节点地址
	DialTimeout   time.Duration // 连接超时时间
	TTL           int64         // 服务租约时间
	MaxRetry      int           // 最大重试次数
	RetryInterval time.Duration // 重试间隔
	Namespace     string        // 服务命名空间
	LogLevel      string        // 日志级别
	EnableCache   bool          // 是否启用本地缓存
	CacheTTL      time.Duration // 缓存过期时间
}

// Option 定义配置选项函数类型
type Option func(*Options)

// DefaultOptions 默认配置
var DefaultOptions = &Options{
	DialTimeout:   time.Second * 3,
	TTL:           15,
	MaxRetry:      3,
	RetryInterval: time.Second,
	Namespace:     "fyerrpc",
	LogLevel:      "info",
	EnableCache:   true,
	CacheTTL:      time.Minute * 5,
}

// WithEndpoints 设置endpoints
func WithEndpoints(endpoints []string) Option {
	return func(o *Options) {
		o.Endpoints = endpoints
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = timeout
	}
}

// WithTTL 设置服务租约时间
func WithTTL(ttl int64) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

// WithMaxRetry 设置最大重试次数
func WithMaxRetry(maxRetry int) Option {
	return func(o *Options) {
		o.MaxRetry = maxRetry
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.RetryInterval = interval
	}
}

// WithNamespace 设置服务命名空间
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level string) Option {
	return func(o *Options) {
		o.LogLevel = level
	}
}

// WithCache 设置缓存配置
func WithCache(enable bool, ttl time.Duration) Option {
	return func(o *Options) {
		o.EnableCache = enable
		if ttl > 0 {
			o.CacheTTL = ttl
		}
	}
}

// Apply 应用配置选项
func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}
