package group

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// contextKey 用于在context中存储分组信息
type contextKey struct{}

var (
	// groupKey 用于从context中获取分组信息的key
	groupKey = &contextKey{}
)

// WithGroup 向context中添加分组信息
func WithGroup(ctx context.Context, group GroupKey) context.Context {
	return context.WithValue(ctx, groupKey, group)
}

// GetGroup 从context中获取分组信息
func GetGroup(ctx context.Context) (GroupKey, bool) {
	group, ok := ctx.Value(groupKey).(GroupKey)
	return group, ok
}

// DefaultRouter 默认的分组路由实现
type DefaultRouter struct {
	manager GroupManager
	mu      sync.RWMutex
}

// NewRouter 创建新的路由器
func NewRouter(manager GroupManager) *DefaultRouter {
	return &DefaultRouter{
		manager: manager,
	}
}

// Route 根据上下文路由到指定分组
func (r *DefaultRouter) Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 从context中获取分组信息
	groupKey, ok := GetGroup(ctx)
	if !ok {
		// 如果没有指定分组，返回所有实例
		return instances, nil
	}

	// 获取分组
	group, err := r.manager.GetGroup(string(groupKey))
	if err != nil {
		return nil, fmt.Errorf("group not found: %v", err)
	}

	// 使用分组选择实例
	return group.Select(ctx, instances)
}

// RouterChain 路由链，支持多个路由规则串联
type RouterChain struct {
	routers []GroupRouter
}

// NewRouterChain 创建新的路由链
func NewRouterChain(routers ...GroupRouter) *RouterChain {
	return &RouterChain{
		routers: routers,
	}
}

// Route 按顺序执行路由链中的所有路由规则
func (c *RouterChain) Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	current := instances
	var err error

	// 按顺序执行每个路由规则
	for _, router := range c.routers {
		current, err = router.Route(ctx, current)
		if err != nil {
			return nil, err
		}
		// 如果某个路由规则过滤后没有实例了，直接返回错误
		if len(current) == 0 {
			return nil, fmt.Errorf("no instances available after routing")
		}
	}

	return current, nil
}

// GroupRouter的一些常用实现

// TagRouter 基于标签的路由
type TagRouter struct {
	tagKey   string
	tagValue string
}

// NewTagRouter 创建基于标签的路由
func NewTagRouter(tagKey, tagValue string) *TagRouter {
	return &TagRouter{
		tagKey:   tagKey,
		tagValue: tagValue,
	}
}

// Route 根据标签路由实例
func (r *TagRouter) Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	var filtered []*naming.Instance
	for _, ins := range instances {
		if value, ok := ins.Metadata[r.tagKey]; ok && value == r.tagValue {
			filtered = append(filtered, ins)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no instances match tag %s=%s", r.tagKey, r.tagValue)
	}
	return filtered, nil
}

// WeightRouter 基于权重的路由
type WeightRouter struct {
	weightKey string
	minWeight int
}

// NewWeightRouter 创建基于权重的路由
func NewWeightRouter(weightKey string, minWeight int) *WeightRouter {
	return &WeightRouter{
		weightKey: weightKey,
		minWeight: minWeight,
	}
}

// Route 根据权重路由实例
func (r *WeightRouter) Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	var filtered []*naming.Instance
	for _, ins := range instances {
		if weight, ok := ins.Metadata[r.weightKey]; ok {
			if w, err := strconv.Atoi(weight); err == nil && w >= r.minWeight {
				filtered = append(filtered, ins)
			}
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no instances match minimum weight %d", r.minWeight)
	}
	return filtered, nil
}

// VersionRouter 基于版本的路由
type VersionRouter struct {
	version string
}

// NewVersionRouter 创建基于版本的路由
func NewVersionRouter(version string) *VersionRouter {
	return &VersionRouter{
		version: version,
	}
}

// Route 根据版本路由实例
func (r *VersionRouter) Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	var filtered []*naming.Instance
	for _, ins := range instances {
		if ins.Version == r.version {
			filtered = append(filtered, ins)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no instances match version %s", r.version)
	}
	return filtered, nil
}
