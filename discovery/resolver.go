package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry"
)

// Resolver 服务解析器
type Resolver struct {
	registry  registry.Registry         // 注册中心接口
	service   string                    // 服务名称
	version   string                    // 服务版本
	instances []*naming.Instance        // 服务实例列表
	watcher   <-chan []*naming.Instance // 服务变更通知通道
	notify    []chan struct{}           // 更新通知通道列表
	done      chan struct{}             // 关闭信号
	mutex     sync.RWMutex
}

// ResolverOption 解析器配置选项
type ResolverOption func(*Resolver)

// NewResolver 创建新的服务解析器
func NewResolver(reg registry.Registry, service, version string, opts ...ResolverOption) (*Resolver, error) {
	r := &Resolver{
		registry: reg,
		service:  service,
		version:  version,
		notify:   make([]chan struct{}, 0),
		done:     make(chan struct{}),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(r)
	}

	// 获取初始服务列表
	instances, err := reg.ListServices(context.Background(), service, version)
	if err != nil {
		return nil, err
	}
	r.instances = instances

	// 订阅服务变更
	watcher, err := reg.Subscribe(context.Background(), service, version)
	if err != nil {
		return nil, err
	}
	r.watcher = watcher

	// 启动更新处理
	go r.watch()

	return r, nil
}

// watch 监听服务变更
func (r *Resolver) watch() {
	for {
		select {
		case <-r.done:
			return
		case instances := <-r.watcher:
			r.mutex.Lock()
			r.instances = instances
			r.mutex.Unlock()

			// 通知所有观察者
			r.notifyUpdate()
		}
	}
}

// Resolve 解析服务地址
func (r *Resolver) Resolve() ([]*naming.Instance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 返回当前实例列表的副本
	instances := make([]*naming.Instance, len(r.instances))
	copy(instances, r.instances)
	return instances, nil
}

// Watch 监听服务变更
func (r *Resolver) Watch() (<-chan struct{}, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ch := make(chan struct{}, 1)
	r.notify = append(r.notify, ch)
	return ch, nil
}

// Close 关闭解析器
func (r *Resolver) Close() error {
	close(r.done)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 关闭所有通知通道
	for _, ch := range r.notify {
		close(ch)
	}
	r.notify = nil

	// 取消服务订阅
	return r.registry.Unsubscribe(context.Background(), r.service, r.version)
}

// notifyUpdate 通知所有观察者服务列表已更新
func (r *Resolver) notifyUpdate() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, ch := range r.notify {
		select {
		case ch <- struct{}{}:
		default:
			// 通道已满，跳过
		}
	}
}

// WithTimeout 设置解析超时时间
func WithTimeout(timeout time.Duration) ResolverOption {
	return func(r *Resolver) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		go func() {
			<-ctx.Done()
			cancel()
		}()
	}
}

// WithRefreshInterval 设置刷新间隔
func WithRefreshInterval(interval time.Duration) ResolverOption {
	return func(r *Resolver) {
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-r.done:
					return
				case <-ticker.C:
					instances, err := r.registry.ListServices(context.Background(), r.service, r.version)
					if err != nil {
						continue
					}
					r.mutex.Lock()
					r.instances = instances
					r.mutex.Unlock()
					r.notifyUpdate()
				}
			}
		}()
	}
}
