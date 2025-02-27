package selector

import (
	"context"
	"fmt"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// Selector 定义了选择器接口
type Selector interface {
	// Select 从服务实例列表中选择符合条件的实例
	Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
	// Name 返回选择器名称
	Name() string
}

// Chain 选择器链，支持多个选择器串联
type Chain struct {
	selectors []Selector
	name      string
}

// NewChain 创建新的选择器链
func NewChain(name string, selectors ...Selector) *Chain {
	return &Chain{
		selectors: selectors,
		name:      name,
	}
}

// Select 按顺序执行选择器链中的所有选择器
func (c *Chain) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("empty instance list")
	}

	current := instances
	for _, selector := range c.selectors {
		selected, err := selector.Select(ctx, current)
		if err != nil {
			return nil, fmt.Errorf("selector %s failed: %v", selector.Name(), err)
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("selector %s returned empty result", selector.Name())
		}
		current = selected
	}

	return current, nil
}

// Add 添加选择器到链中
func (c *Chain) Add(selector Selector) {
	c.selectors = append(c.selectors, selector)
}

// Remove 从链中移除选择器
func (c *Chain) Remove(name string) bool {
	for i, s := range c.selectors {
		if s.Name() == name {
			c.selectors = append(c.selectors[:i], c.selectors[i+1:]...)
			return true
		}
	}
	return false
}

// Reset 重置选择器链
func (c *Chain) Reset() {
	c.selectors = make([]Selector, 0)
}

// Len 返回选择器链的长度
func (c *Chain) Len() int {
	return len(c.selectors)
}

// Name 返回选择器链名称
func (c *Chain) Name() string {
	return c.name
}

// BaseSelector 选择器基础实现，可以被其他选择器组合使用
type BaseSelector struct {
	name     string
	selectFn func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

// NewBaseSelector 创建基础选择器
func NewBaseSelector(name string, fn func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)) *BaseSelector {
	return &BaseSelector{
		name:     name,
		selectFn: fn,
	}
}

// Select 实现Selector接口
func (s *BaseSelector) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	if s.selectFn == nil {
		return instances, nil
	}
	return s.selectFn(ctx, instances)
}

// Name 返回选择器名称
func (s *BaseSelector) Name() string {
	return s.name
}

// ChainBuilder 选择器链构建器
type ChainBuilder struct {
	selectors []Selector
}

// NewChainBuilder 创建选择器链构建器
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		selectors: make([]Selector, 0),
	}
}

// Add 添加选择器
func (b *ChainBuilder) Add(selector Selector) *ChainBuilder {
	b.selectors = append(b.selectors, selector)
	return b
}

// Build 构建选择器链
func (b *ChainBuilder) Build() *Chain {
	return NewChain("", b.selectors...)
}
