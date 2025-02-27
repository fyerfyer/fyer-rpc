package group

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// DefaultGroup 默认分组实现
type DefaultGroup struct {
	name     string
	config   *Config
	matcher  Matcher
	weight   int
	metadata map[string]string
	mu       sync.RWMutex
}

// Matcher 分组匹配器接口
type Matcher interface {
	Match(instance *naming.Instance) bool
}

// NewGroup 创建新的分组
func NewGroup(name string, opts ...Option) (*DefaultGroup, error) {
	// 使用默认配置
	config := *DefaultConfig
	// 应用选项
	for _, opt := range opts {
		opt(&config)
	}

	if name == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	// 创建匹配器
	matcher, err := createMatcher(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create matcher: %v", err)
	}

	return &DefaultGroup{
		name:     name,
		config:   &config,
		matcher:  matcher,
		weight:   config.Weight,
		metadata: config.Metadata,
	}, nil
}

// Name 返回分组名称
func (g *DefaultGroup) Name() string {
	return g.name
}

// Match 检查实例是否匹配该分组
func (g *DefaultGroup) Match(instance *naming.Instance) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.matcher.Match(instance)
}

// Select 从实例列表中选择匹配该分组的实例
func (g *DefaultGroup) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	matched := make([]*naming.Instance, 0)
	for _, instance := range instances {
		if g.matcher.Match(instance) {
			matched = append(matched, instance)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no instances match group %s", g.name)
	}

	return matched, nil
}

// 具体的匹配器实现
type exactMatcher struct {
	key   string
	value string
}

type prefixMatcher struct {
	key    string
	prefix string
}

type regexMatcher struct {
	key    string
	regexp *regexp.Regexp
}

type labelMatcher struct {
	labels map[string]string
}

func (m *exactMatcher) Match(instance *naming.Instance) bool {
	if value, ok := instance.Metadata[m.key]; ok {
		return value == m.value
	}
	return false
}

func (m *prefixMatcher) Match(instance *naming.Instance) bool {
	if value, ok := instance.Metadata[m.key]; ok {
		return strings.HasPrefix(value, m.prefix)
	}
	return false
}

func (m *regexMatcher) Match(instance *naming.Instance) bool {
	if value, ok := instance.Metadata[m.key]; ok {
		return m.regexp.MatchString(value)
	}
	return false
}

func (m *labelMatcher) Match(instance *naming.Instance) bool {
	for k, v := range m.labels {
		if value, ok := instance.Metadata[k]; !ok || value != v {
			return false
		}
	}
	return true
}

// createMatcher 根据配置创建匹配器
func createMatcher(config *Config) (Matcher, error) {
	if config.Matcher == nil {
		return &labelMatcher{labels: make(map[string]string)}, nil
	}

	switch config.Matcher.MatchType {
	case "exact":
		return &exactMatcher{
			key:   config.Matcher.MatchKey,
			value: config.Matcher.MatchValue,
		}, nil
	case "prefix":
		return &prefixMatcher{
			key:    config.Matcher.MatchKey,
			prefix: config.Matcher.MatchValue,
		}, nil
	case "regex":
		reg, err := regexp.Compile(config.Matcher.MatchValue)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
		return &regexMatcher{
			key:    config.Matcher.MatchKey,
			regexp: reg,
		}, nil
	case "label":
		return &labelMatcher{
			labels: config.Matcher.Labels,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported match type: %s", config.Matcher.MatchType)
	}
}

// UpdateConfig 更新分组配置
func (g *DefaultGroup) UpdateConfig(opts ...Option) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 创建新的配置副本
	newConfig := *g.config
	for _, opt := range opts {
		opt(&newConfig)
	}

	// 创建新的匹配器
	matcher, err := createMatcher(&newConfig)
	if err != nil {
		return err
	}

	// 更新配置
	g.config = &newConfig
	g.matcher = matcher
	g.weight = newConfig.Weight
	g.metadata = newConfig.Metadata

	return nil
}

// GetWeight 获取分组权重
func (g *DefaultGroup) GetWeight() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.weight
}

// GetMetadata 获取分组元数据
func (g *DefaultGroup) GetMetadata() map[string]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	metadata := make(map[string]string)
	for k, v := range g.metadata {
		metadata[k] = v
	}
	return metadata
}
