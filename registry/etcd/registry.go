package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdRegistry etcd注册中心实现
type EtcdRegistry struct {
	client   *clientv3.Client
	options  *Options
	leases   sync.Map                                      // 用于存储租约和key的对应关系
	watchers map[string]map[string]chan []*naming.Instance // service -> version -> channel
	mu       sync.RWMutex
}

// New 创建etcd注册中心实例
func New(opts ...Option) (*EtcdRegistry, error) {
	options := &Options{
		Options: registry.DefaultOptions,
	}
	for _, opt := range opts {
		opt(options)
	}

	config := clientv3.Config{
		Endpoints:            options.Endpoints,
		DialTimeout:          options.DialTimeout,
		Username:             options.Username,
		Password:             options.Password,
		AutoSyncInterval:     options.AutoSyncInterval,
		DialKeepAliveTime:    options.DialKeepAlive,
		DialKeepAliveTimeout: options.DialTimeout,
	}

	if options.CertFile != "" && options.KeyFile != "" && options.TrustedCAFile != "" {
		tlsConfig, err := loadTLSConfig(options.CertFile, options.KeyFile, options.TrustedCAFile)
		if err != nil {
			return nil, err
		}
		config.TLS = tlsConfig
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	return &EtcdRegistry{
		client:   client,
		options:  options,
		watchers: make(map[string]map[string]chan []*naming.Instance),
	}, nil
}

// Register 注册服务实例
func (r *EtcdRegistry) Register(ctx context.Context, service *naming.Instance) error {
	key := naming.BuildServiceKey(service.Service, service.Version, service.ID)
	value, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 创建租约
	lease, err := r.client.Grant(ctx, r.options.TTL)
	if err != nil {
		return err
	}

	// 注册服务并绑定租约
	_, err = r.client.Put(ctx, key, string(value), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	// 保持租约
	keepAliveCh, err := r.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return err
	}

	// 存储租约信息
	r.leases.Store(key, lease.ID)

	// 启动心跳协程
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-keepAliveCh:
				if resp == nil { // 租约已过期
					r.Deregister(ctx, service)
					return
				}
			}
		}
	}()

	return nil
}

// Deregister 注销服务实例
func (r *EtcdRegistry) Deregister(ctx context.Context, service *naming.Instance) error {
	key := naming.BuildServiceKey(service.Service, service.Version, service.ID)

	// 删除服务实例
	if leaseID, ok := r.leases.Load(key); ok {
		// 撤销租约
		_, err := r.client.Revoke(ctx, leaseID.(clientv3.LeaseID))
		if err != nil {
			return err
		}
		r.leases.Delete(key)
	}

	return nil
}

// Subscribe 订阅服务变更
func (r *EtcdRegistry) Subscribe(ctx context.Context, service, version string) (<-chan []*naming.Instance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 初始化服务映射
	if _, ok := r.watchers[service]; !ok {
		r.watchers[service] = make(map[string]chan []*naming.Instance)
	}

	// 生成唯一键
	key := fmt.Sprintf("%s/%s", service, version)

	// 检查是否已经存在相同服务和版本的订阅
	if ch, ok := r.watchers[service][version]; ok {
		// 已存在订阅，直接返回
		return ch, nil
	}

	// 创建服务版本的订阅通道
	ch := make(chan []*naming.Instance, 10)
	r.watchers[service][version] = ch

	// 获取当前服务列表
	instances, err := r.ListServices(ctx, service, version)
	if err != nil {
		delete(r.watchers[service], version)
		if len(r.watchers[service]) == 0 {
			delete(r.watchers, service)
		}
		return nil, err
	}

	// 创建一个独立的上下文，用于控制监听
	watchCtx, cancel := context.WithCancel(context.Background())

	// 存储取消函数，在Unsubscribe时调用
	r.leases.Store(key, cancel)

	// 启动监听协程前先发送初始数据
	select {
	case ch <- instances:
	default:
		// 通道已满，跳过发送
	}

	// 启动监听协程
	prefix := fmt.Sprintf("/fyerrpc/services/%s/%s/", service, version)
	watchCh := r.client.Watch(watchCtx, prefix, clientv3.WithPrefix())

	go func() {
		// 监听退出时的清理工作
		defer func() {
			// 捕获可能的panic
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in Subscribe goroutine: %v", r)
			}

			// 取消watchCtx避免资源泄露
			cancel()

			// 安全地关闭和清理通道
			r.mu.Lock()
			defer r.mu.Unlock()

			// 再次检查watcher是否还存在，并且是否是我们期望的channel
			if versionMap, serviceExists := r.watchers[service]; serviceExists {
				if existingCh, versionExists := versionMap[version]; versionExists && existingCh == ch {
					// 安全地删除映射并关闭通道
					delete(r.watchers[service], version)
					if len(r.watchers[service]) == 0 {
						delete(r.watchers, service)
					}
					close(ch)
				}
			}
		}()

		for {
			select {
			case <-watchCtx.Done():
				return
			case resp, ok := <-watchCh:
				if !ok {
					return
				}
				if resp.Canceled {
					return
				}
				if resp.Err() != nil {
					log.Printf("watch error: %v", resp.Err())
					continue
				}

				// 有任何事件发生时，重新获取完整的服务列表
				newCtx, cancel := context.WithTimeout(context.Background(), r.options.DialTimeout)
				instances, err := r.ListServices(newCtx, service, version)
				cancel() // 立即释放context资源

				if err != nil {
					log.Printf("list services error: %v", err)
					continue
				}

				// 先检查通道是否还存在
				r.mu.RLock()
				_, serviceExists := r.watchers[service]
				channelExists := false
				if serviceExists {
					_, channelExists = r.watchers[service][version]
				}
				r.mu.RUnlock()

				if !channelExists {
					return // 通道已被删除，停止goroutine
				}

				// 发送更新后的服务列表
				select {
				case ch <- instances:
					// 成功发送
				default:
					// 通道已满，跳过本次更新
				}
			}
		}
	}()

	return ch, nil
}

// Unsubscribe 取消订阅服务变更
func (r *EtcdRegistry) Unsubscribe(ctx context.Context, service, version string) error {
	key := fmt.Sprintf("%s/%s", service, version)

	// 首先取消监听上下文，这会触发监听goroutine结束
	if cancelFunc, ok := r.leases.Load(key); ok {
		cancel := cancelFunc.(context.CancelFunc)
		cancel()
		r.leases.Delete(key)
	}

	// 给监听goroutine一些时间来完成清理工作
	r.mu.RLock()
	hasService := false
	if svcMap, ok := r.watchers[service]; ok {
		_, hasService = svcMap[version]
	}
	r.mu.RUnlock()

	// 如果监听goroutine没有及时清理，我们手动清理
	if hasService {
		time.Sleep(10 * time.Millisecond) // 给goroutine一点时间处理关闭

		r.mu.Lock()
		defer r.mu.Unlock()

		// 再次检查是否已被清理
		if svcMap, ok := r.watchers[service]; ok {
			if ch, exists := svcMap[version]; exists {
				// 安全地删除映射并关闭通道
				delete(r.watchers[service], version)
				if len(r.watchers[service]) == 0 {
					delete(r.watchers, service)
				}
				close(ch)
			}
		}
	}

	return nil
}

// ListServices 获取服务实例列表
func (r *EtcdRegistry) ListServices(ctx context.Context, service, version string) ([]*naming.Instance, error) {
	prefix := fmt.Sprintf("/fyerrpc/services/%s/%s/", service, version)
	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	instances := make([]*naming.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		instance := &naming.Instance{}
		if err := json.Unmarshal(kv.Value, instance); err != nil {
			log.Fatalf("unmarshal instance error: %v", err)
			continue
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

// Heartbeat 服务心跳
func (r *EtcdRegistry) Heartbeat(ctx context.Context, service *naming.Instance) error {
	key := naming.BuildServiceKey(service.Service, service.Version, service.ID)
	value, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 更新服务实例信息，确保服务活跃
	_, err = r.client.Put(ctx, key, string(value))
	return err
}

// Close 关闭注册中心
func (r *EtcdRegistry) Close() error {
	r.mu.Lock()

	// 取消所有监听上下文
	r.leases.Range(func(key, value interface{}) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		return true
	})

	// 关闭所有监听器通道
	for service, versions := range r.watchers {
		for version, ch := range versions {
			close(ch)
			delete(versions, version)
		}
		delete(r.watchers, service)
	}

	r.mu.Unlock()

	// 关闭etcd客户端
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// loadTLSConfig 加载TLS配置
func loadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	caData, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}, nil
}
