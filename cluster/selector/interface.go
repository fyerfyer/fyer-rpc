package selector

import (
	"context"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// SelectorInterface 定义了服务选择器接口
type SelectorInterface interface {
	// Select 从服务实例列表中选择符合条件的实例
	Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)

	// Name 返回选择器名称
	Name() string
}

// SelectFunc 选择器函数类型，用于快速创建选择器
type SelectFunc func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)

// ContextKey 用于在context中存储选择器相关信息的key类型
type contextKey struct {
	name string
}

var (
	// GroupKey 用于在context中存储分组信息的key
	GroupKey = &contextKey{name: "group"}
	// VersionKey 用于在context中存储版本信息的key
	VersionKey = &contextKey{name: "version"}
	// RegionKey 用于在context中存储地域信息的key
	RegionKey = &contextKey{name: "region"}
)

// SelectorFactory 选择器工厂方法类型
type SelectorFactory func(conf *Config) Selector

// Config 选择器配置
type Config struct {
	// Strategy 选择策略，如group、version、region等
	Strategy string `json:"strategy"`

	// Filter 筛选规则
	Filter map[string]string `json:"filter"`

	// Priority 选择器优先级，数字越小优先级越高
	Priority int `json:"priority"`

	// Required 是否必须满足选择条件
	Required bool `json:"required"`

	// Metadata 额外的元数据
	Metadata map[string]string `json:"metadata"`
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	Strategy: "group",
	Filter:   make(map[string]string),
	Priority: 0,
	Required: true,
	Metadata: make(map[string]string),
}

// SelectorChain 选择器链，用于组合多个选择器
type SelectorChain interface {
	// Add 添加选择器到链中
	Add(selector Selector) SelectorChain

	// Remove 从链中移除选择器
	Remove(name string) bool

	// Select 执行选择器链的选择逻辑
	Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)

	// Reset 重置选择器链
	Reset()

	// Len 返回选择器链的长度
	Len() int
}

// Option 定义选择器配置选项函数类型
type Option func(*Config)

// WithStrategy 设置选择策略
func WithStrategy(strategy string) Option {
	return func(c *Config) {
		c.Strategy = strategy
	}
}

// WithFilter 设置筛选规则
func WithFilter(filter map[string]string) Option {
	return func(c *Config) {
		c.Filter = filter
	}
}

// WithPriority 设置优先级
func WithPriority(priority int) Option {
	return func(c *Config) {
		c.Priority = priority
	}
}

// WithRequired 设置是否必须满足条件
func WithRequired(required bool) Option {
	return func(c *Config) {
		c.Required = required
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata map[string]string) Option {
	return func(c *Config) {
		c.Metadata = metadata
	}
}

// Errors
var (
	ErrNoInstances     = NewError("no instances available")
	ErrSelectorClosed  = NewError("selector is closed")
	ErrInvalidConfig   = NewError("invalid selector config")
	ErrInvalidStrategy = NewError("invalid selector strategy")
)

// Error 选择器错误类型
type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

// NewError 创建新的错误
func NewError(message string) error {
	return &Error{message: message}
}
