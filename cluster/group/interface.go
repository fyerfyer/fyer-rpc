package group

import (
	"context"
	"github.com/fyerfyer/fyer-rpc/naming"
)

// GroupKey 分组标识的类型
type GroupKey string

// Group 表示一个服务分组的接口
type Group interface {
	// Name 返回分组名称
	Name() string

	// Match 检查实例是否匹配该分组
	Match(instance *naming.Instance) bool

	// Select 从实例列表中选择匹配该分组的实例
	Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

// GroupSelector 分组选择器接口
type GroupSelector interface {
	// Select 根据上下文选择合适的分组
	Select(ctx context.Context) (GroupKey, error)
}

// GroupRouter 分组路由接口
type GroupRouter interface {
	// Route 根据上下文路由到指定分组
	Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

// GroupManager 分组管理接口
type GroupManager interface {
	// RegisterGroup 注册一个新的分组
	RegisterGroup(group Group) error

	// GetGroup 获取指定名称的分组
	GetGroup(name string) (Group, error)

	// ListGroups 列出所有已注册的分组
	ListGroups() []Group
}
