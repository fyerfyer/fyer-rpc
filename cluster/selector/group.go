package selector

import (
	"context"
	"errors"
	"github.com/fyerfyer/fyer-rpc/naming"
)

var ErrNoAvailableInstances = errors.New("no available instances")

// GroupSelector 基于分组的选择器实现
type GroupSelector struct {
	name      string
	groupKey  string
	groupName string
}

// NewGroupSelector 创建新的分组选择器
func NewGroupSelector(name, groupKey, groupName string) *GroupSelector {
	return &GroupSelector{
		name:      name,
		groupKey:  groupKey,
		groupName: groupName,
	}
}

// Select 根据分组信息选择实例
func (s *GroupSelector) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstances
	}

	// 从context中获取目标分组
	targetGroup, ok := ctx.Value(s.groupKey).(string)
	if !ok {
		// 如果没有指定分组，使用默认分组
		targetGroup = s.groupName
	}

	// 选择符合分组的实例
	var selected []*naming.Instance
	for _, instance := range instances {
		if group, ok := instance.Metadata[s.groupKey]; ok && group == targetGroup {
			selected = append(selected, instance)
		}
	}

	if len(selected) == 0 {
		// 如果没有找到匹配的实例，返回错误
		return nil, ErrNoAvailableInstances
	}

	return selected, nil
}

// Name 返回选择器名称
func (s *GroupSelector) Name() string {
	return s.name
}

// WithGroup 向context中添加分组信息
func WithGroup(ctx context.Context, groupKey string, groupValue string) context.Context {
	return context.WithValue(ctx, groupKey, groupValue)
}

// ABGroupSelector A/B测试分组选择器
type ABGroupSelector struct {
	name   string
	groupA string
	groupB string
	ratio  float64 // A组流量比例(0-1)
}

// NewABGroupSelector 创建新的A/B测试分组选择器
func NewABGroupSelector(name string, groupA, groupB string, ratio float64) *ABGroupSelector {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return &ABGroupSelector{
		name:   name,
		groupA: groupA,
		groupB: groupB,
		ratio:  ratio,
	}
}

// Select 根据A/B分组策略选择实例
func (s *ABGroupSelector) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
    if len(instances) == 0 {
        return nil, ErrNoInstances
    }

    // 从context中获取是否强制指定分组
    if forcedGroup, ok := ctx.Value("ab_group").(string); ok {
        var selected []*naming.Instance
        for _, instance := range instances {
            if group := instance.Metadata["group"]; group == forcedGroup {
                selected = append(selected, instance)
            }
        }
        if len(selected) > 0 {
            return selected, nil
        }
    }

    // 根据比例分配实例到A/B组
    var groupA, groupB []*naming.Instance
    for _, instance := range instances {
        if group := instance.Metadata["group"]; group == s.groupA {
            groupA = append(groupA, instance)
        } else if group == s.groupB {
            groupB = append(groupB, instance)
        }
    }

    // 根据比例选择分组
    if len(groupA) > 0 && (len(groupB) == 0 || s.ratio > 0.5) {
        return groupA, nil
    } else if len(groupB) > 0 {
        return groupB, nil
    }

    return nil, ErrNoAvailableInstances
}

// Name 返回选择器名称
func (s *ABGroupSelector) Name() string {
	return s.name
}
