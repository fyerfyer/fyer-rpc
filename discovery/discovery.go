package discovery

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry"
)

var (
	ErrNotFound      = errors.New("service not found")
	ErrWatcherClosed = errors.New("watcher closed")
)

// Discovery 服务发现接口
type Discovery interface {
	// GetService 获取服务实例
	GetService(ctx context.Context, name string, version string) ([]*naming.Instance, error)

	// Watch 监听服务变更
	Watch(ctx context.Context, name string, version string) (Watcher, error)

	// Close 关闭服务发现
	Close() error
}

// Watcher 服务变更监听器接口
type Watcher interface {
	// Next 获取下一次服务更新
	Next() ([]*naming.Instance, error)

	// Stop 停止监听
	Stop() error
}

// DefaultDiscovery 默认的服务发现实现
type DefaultDiscovery struct {
	registry registry.Registry     // 注册中心
	cache    map[string]*cacheData // 本地缓存，key为service:version
	cacheTTL time.Duration         // 缓存过期时间
	watchers sync.Map              // 存储所有的监听器
	mu       sync.RWMutex
	close    chan struct{}
	closed   bool // 新增：标记服务是否已关闭
}

// cacheData 缓存数据结构
type cacheData struct {
	instances  []*naming.Instance // 服务实例列表
	timestamp  time.Time          // 最后更新时间
	updateChan chan struct{}      // 更新通知通道
}

// NewDiscovery 创建服务发现实例
func NewDiscovery(reg registry.Registry, cacheTTL time.Duration) Discovery {
	d := &DefaultDiscovery{
		registry: reg,
		cache:    make(map[string]*cacheData),
		cacheTTL: cacheTTL,
		close:    make(chan struct{}),
	}

	// 启动缓存清理协程
	go d.cleanCache()

	return d
}

// GetService 获取服务实例
func (d *DefaultDiscovery) GetService(ctx context.Context, name string, version string) ([]*naming.Instance, error) {
	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		return nil, errors.New("discovery is closed")
	}
	d.mu.RUnlock()

	key := name + ":" + version

	// 先查询缓存
	d.mu.RLock()
	if cd, ok := d.cache[key]; ok {
		if time.Since(cd.timestamp) < d.cacheTTL {
			instances := make([]*naming.Instance, len(cd.instances))
			copy(instances, cd.instances)
			d.mu.RUnlock()
			return instances, nil
		}
	}
	d.mu.RUnlock()

	// 缓存未命中或已过期，从注册中心获取
	instances, err := d.registry.ListServices(ctx, name, version)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	d.mu.Lock()
	// 再次检查是否已关闭
	if d.closed {
		d.mu.Unlock()
		return nil, errors.New("discovery is closed")
	}
	if d.cache != nil {
		d.cache[key] = &cacheData{
			instances:  instances,
			timestamp:  time.Now(),
			updateChan: make(chan struct{}, 1),
		}
	}
	d.mu.Unlock()

	return instances, nil
}

// Watch 监听服务变更
func (d *DefaultDiscovery) Watch(ctx context.Context, name string, version string) (Watcher, error) {
	// 创建监听器
	w := &defaultWatcher{
		discovery: d,
		service:   name,
		version:   version,
		eventChan: make(chan []*naming.Instance, 10),
		stopChan:  make(chan struct{}),
	}

	// 订阅注册中心的变更
	watchChan, err := d.registry.Subscribe(ctx, name, version)
	if err != nil {
		return nil, err
	}

	// 启动监听协程
	go func() {
		defer func() {
			// 确保在协程退出时关闭通道
			w.mu.Lock()
			if !w.stopped {
				close(w.eventChan)
				w.stopped = true
			}
			w.mu.Unlock()
		}()

		for {
			select {
			case <-w.stopChan:
				return
			case instances, ok := <-watchChan:
				if !ok {
					return
				}
				// 检查是否已停止
				w.mu.Lock()
				stopped := w.stopped
				w.mu.Unlock()
				if stopped {
					return
				}

				// 更新本地缓存
				d.updateCache(name, version, instances)

				// 通知观察者
				select {
				case w.eventChan <- instances:
				case <-w.stopChan:
					return
				default:
					// 通道已满，跳过本次更新
				}
			}
		}
	}()

	return w, nil
}

// Close 关闭服务发现
func (d *DefaultDiscovery) Close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	close(d.close)
	d.cache = nil
	d.mu.Unlock()
	return nil
}

// cleanCache 定期清理过期缓存
func (d *DefaultDiscovery) cleanCache() {
	ticker := time.NewTicker(d.cacheTTL)
	defer ticker.Stop()

	for {
		select {
		case <-d.close:
			return
		case <-ticker.C:
			d.mu.Lock()
			for key, cd := range d.cache {
				if time.Since(cd.timestamp) > d.cacheTTL {
					delete(d.cache, key)
				}
			}
			d.mu.Unlock()
		}
	}
}

// updateCache 更新本地缓存
func (d *DefaultDiscovery) updateCache(service, version string, instances []*naming.Instance) {
	key := service + ":" + version
	d.mu.Lock()
	if cd, ok := d.cache[key]; ok {
		cd.instances = instances
		cd.timestamp = time.Now()
		select {
		case cd.updateChan <- struct{}{}:
		default:
		}
	} else {
		d.cache[key] = &cacheData{
			instances:  instances,
			timestamp:  time.Now(),
			updateChan: make(chan struct{}, 1),
		}
	}
	d.mu.Unlock()
}

// defaultWatcher 默认的服务监听器实现
type defaultWatcher struct {
	discovery *DefaultDiscovery
	service   string
	version   string
	eventChan chan []*naming.Instance
	stopChan  chan struct{}
	stopped   bool
	mu        sync.Mutex
}

// Next 获取下一次服务更新
func (w *defaultWatcher) Next() ([]*naming.Instance, error) {
	select {
	case instances, ok := <-w.eventChan:
		if !ok {
			return nil, ErrWatcherClosed
		}
		return instances, nil
	case <-w.stopChan:
		return nil, ErrWatcherClosed
	}
}

// Stop 停止监听
func (w *defaultWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.stopped {
		w.stopped = true
		close(w.stopChan)
		// 不在这里关闭 eventChan，让监听协程来关闭它
	}
	return nil
}
