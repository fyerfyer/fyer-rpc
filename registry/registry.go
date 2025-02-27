package registry

import (
	"context"
	"github.com/fyerfyer/fyer-rpc/naming"
)

// Registry 注册中心接口
type Registry interface {
	// Register 注册服务实例
	Register(ctx context.Context, service *naming.Instance) error

	// Deregister 注销服务实例
	Deregister(ctx context.Context, service *naming.Instance) error

	// Subscribe 订阅服务变更
	Subscribe(ctx context.Context, service string, version string) (<-chan []*naming.Instance, error)

	// Unsubscribe 取消订阅服务变更
	Unsubscribe(ctx context.Context, service string, version string) error

	// ListServices 获取服务实例列表
	ListServices(ctx context.Context, service string, version string) ([]*naming.Instance, error)

	// Heartbeat 服务心跳
	Heartbeat(ctx context.Context, service *naming.Instance) error

	// Close 关闭注册中心连接
	Close() error
}
