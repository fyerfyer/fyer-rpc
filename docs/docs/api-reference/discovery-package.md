# `discovery` Package

`discovery`包提供了服务发现功能，支持动态查找和监控服务实例。

### Discovery

Discovery接口定义了服务发现的核心功能：

```go
type Discovery interface {
    // GetService 获取服务实例
    GetService(ctx context.Context, name string, version string) ([]*naming.Instance, error)

    // Watch 监听服务变更
    Watch(ctx context.Context, name string, version string) (Watcher, error)

    // Close 关闭服务发现
    Close() error
}
```

### Watcher

`Watcher`接口用于监听服务实例的变化：

```go
type Watcher interface {
    // Next 获取下一次服务更新
    Next() ([]*naming.Instance, error)

    // Stop 停止监听
    Stop() error
}
```

### Resolver

`Resolver`结构体实现了服务解析器：

```go
type Resolver struct {
    registry  registry.Registry
    service   string
    version   string
    instances []*naming.Instance
    watcher   <-chan []*naming.Instance
}

// NewResolver 创建服务解析器
func NewResolver(reg registry.Registry, service, version string, opts ...ResolverOption) (*Resolver, error)

// Resolve 解析服务地址
func (r *Resolver) Resolve() ([]*naming.Instance, error)
```

### LoadBalancer

`LoadBalancer`结构体实现了客户端负载均衡：

```go
type LoadBalancer struct {
    resolver       *Resolver
    balancer       balancer.Balancer
    failover       *failover.DefaultFailoverHandler
    metrics        metrics.Metrics
    serviceName    string
    version        string
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(serviceName, version string, resolver *Resolver, metrics metrics.Metrics, balancerType balancer.BalancerType, options ...LoadBalancerOption) (*LoadBalancer, error)

// Select 选择一个服务实例
func (lb *LoadBalancer) Select(ctx context.Context) (*naming.Instance, error)
```

### balancer 子包

`balancer`子包提供了不同的负载均衡策略：

```go
type Balancer interface {
    // Initialize 初始化负载均衡器
    Initialize(instances []*naming.Instance) error

    // Select 选择一个服务实例
    Select(ctx context.Context) (*naming.Instance, error)

    // Update 更新服务实例列表
    Update(instances []*naming.Instance) error

    // Feedback 服务调用结果反馈
    Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error)

    // Name 返回负载均衡器名称
    Name() string
}
```

内置策略：
- `Random`: 随机选择
- `RoundRobin`: 轮询选择
- `FastestResponse`: 最快响应时间

### metrics 子包

`metrics`子包提供了性能指标收集功能：

```go
type Metrics interface {
    // RecordResponse 记录响应时间
    RecordResponse(ctx context.Context, metric *ResponseMetric) error

    // GetLatency 获取指定服务实例的平均响应时间
    GetLatency(ctx context.Context, service, instance string) (time.Duration, error)

    // GetServiceLatency 获取服务所有实例的平均响应时间
    GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error)

    // 其它方法...
}
```