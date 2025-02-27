package group

import (
	"time"
)

// Config 分组配置
type Config struct {
	// Name 分组名称
	Name string `json:"name"`

	// Type 分组类型，如A/B测试、金丝雀发布等
	Type string `json:"type"`

	// Matcher 分组匹配规则
	Matcher *MatchConfig `json:"matcher"`

	// Weight 分组权重(0-100)
	Weight int `json:"weight"`

	// EnableHealthCheck 是否启用健康检查
	EnableHealthCheck bool `json:"enable_health_check"`

	// HealthCheckInterval 健康检查间隔
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	// Metadata 分组元数据
	Metadata map[string]string `json:"metadata"`
}

// MatchConfig 匹配规则配置
type MatchConfig struct {
	// MatchType 匹配类型：exact(精确匹配)、prefix(前缀匹配)、regex(正则匹配)
	MatchType string `json:"match_type"`

	// MatchKey 匹配的键，如环境变量名、Header名等
	MatchKey string `json:"match_key"`

	// MatchValue 匹配的值
	MatchValue string `json:"match_value"`

	// Labels 标签匹配规则
	Labels map[string]string `json:"labels"`
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	Type:                "default",
	Weight:              100,
	EnableHealthCheck:   true,
	HealthCheckInterval: time.Second * 30,
	Metadata:            make(map[string]string),
	Matcher: &MatchConfig{
		MatchType: "exact",
		Labels:    make(map[string]string),
	},
}

// Option 配置选项函数类型
type Option func(*Config)

// WithName 设置分组名称
func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

// WithType 设置分组类型
func WithType(typ string) Option {
	return func(c *Config) {
		c.Type = typ
	}
}

// WithWeight 设置分组权重
func WithWeight(weight int) Option {
	return func(c *Config) {
		if weight >= 0 && weight <= 100 {
			c.Weight = weight
		}
	}
}

// WithMatcher 设置匹配规则
func WithMatcher(matcher *MatchConfig) Option {
	return func(c *Config) {
		c.Matcher = matcher
	}
}

// WithHealthCheck 设置健康检查
func WithHealthCheck(enable bool, interval time.Duration) Option {
	return func(c *Config) {
		c.EnableHealthCheck = enable
		if interval > 0 {
			c.HealthCheckInterval = interval
		}
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata map[string]string) Option {
	return func(c *Config) {
		c.Metadata = metadata
	}
}
