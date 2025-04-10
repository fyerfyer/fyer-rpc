# `registry` Package

`registry`包提供了服务注册与发现的抽象和实现。

### Registry

Registry接口定义了注册中心的核心功能：

```go
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

    // UpdateService 服务心跳
    UpdateService(ctx context.Context, service *naming.Instance) error

    // Close 关闭注册中心连接
    Close() error
}
```

### etcd 子包

`etcd`子包提供了基于 etcd 的注册中心实现：

```go
// New 创建etcd注册中心实例
func New(opts ...Option) (*EtcdRegistry, error)
```

配置选项：

```go
// WithEndpoints 设置etcd endpoints
func WithEndpoints(endpoints []string) Option

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(timeout time.Duration) Option

// WithTTL 设置服务租约时间
func WithTTL(ttl int64) Option

// WithTLSConfig 设置TLS配置
func WithTLSConfig(certFile, keyFile, caFile string) Option

// 其它配置选项...
```