# Load Balance

负载均衡是分布式系统中的关键组件，它可以将客户端请求均匀地分发到多个服务实例上，以提高系统的整体性能、可用性和可扩展性。

## 负载均衡基础

### 负载均衡器接口

fyerrpc通过`balancer.Balancer`接口定义了负载均衡器的标准行为：

```go
type Balancer interface {
    // Initialize 初始化负载均衡器
    Initialize(instances []*naming.Instance) error

    // Select 选择一个服务实例
    Select(ctx context.Context) (*naming.Instance, error)

    // Update 更新服务实例列表
    Update(instances []*naming.Instance) error

    // Feedback 服务调用结果反馈，用于更新实例状态
    Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error)

    // Name 返回负载均衡器名称
    Name() string
}
```

### 支持的负载均衡策略

fyerrpc目前支持以下几种负载均衡策略：

```go
type BalancerType string

const (
    FastestResponse BalancerType = "fastest_response" // 最快响应时间
    Random          BalancerType = "random"           // 随机，暂未实现
    RoundRobin      BalancerType = "round_robin"      // 轮询，暂未实现
)
```

### 负载均衡配置

负载均衡器的配置通过`balancer.Config`结构体定义：

```go
type Config struct {
    Type           BalancerType    // 负载均衡类型
    MetricsClient  metrics.Metrics // 指标收集客户端
    UpdateInterval int64           // 更新间隔(秒)
    RetryTimes     int             // 重试次数
}
```

## 内置负载均衡策略

### 最快响应策略 (Fastest Response)

最快响应负载均衡器基于实例的历史响应时间进行选择，优先选择响应时间最短的实例。这种策略特别适用于服务实例的性能不均衡或者网络延迟差异较大的情况。

**工作原理：**

1. 收集并记录每个实例的响应时间
2. 根据历史响应时间对实例进行排序
3. 优先选择响应时间最短的实例
4. 定期更新响应时间统计数据

**内部实现：**

```go
// 负载均衡器结构
type FastestBalancer struct {
    instances    []*instanceWrapper // 服务实例包装列表
    metrics      metrics.Metrics    // 指标收集客户端
    updateTicker *time.Ticker       // 更新定时器
    retryTimes   int                // 重试次数
    mu           sync.RWMutex
}

// 实例包装，包含实例信息和性能指标
type instanceWrapper struct {
    *naming.Instance               // 服务实例
    latency          time.Duration // 平均响应时间
    weight           float64       // 权重分数
    lastUpdate       time.Time     // 最后更新时间
}

// Select 选择一个服务实例
func (b *FastestBalancer) Select(ctx context.Context) (*naming.Instance, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    if len(b.instances) == 0 {
        return nil, ErrNoAvailableInstances
    }

    // 按响应时间排序
    instances := make([]*instanceWrapper, len(b.instances))
    copy(instances, b.instances)
    sort.Slice(instances, func(i, j int) bool {
        return instances[i].latency < instances[j].latency
    })

    // 选择响应最快的实例
    for i := 0; i < b.retryTimes && i < len(instances); i++ {
        if instances[i].Status == naming.StatusEnabled {
            return instances[i].Instance, nil
        }
    }

    return nil, ErrNoAvailableInstances
}
```

**使用方式：**

要使用最快响应负载均衡器，您需要提供一个指标收集器：

```go
import (
    "github.com/fyerfyer/fyer-rpc/discovery/balancer"
    "github.com/fyerfyer/fyer-rpc/discovery/metrics"
)

// 创建Prometheus指标收集器
metricsClient, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
    PushGatewayURL: "http://localhost:9091",
    QueryURL:       "http://localhost:9090",
    JobName:        "fyerrpc",
    PushInterval:   time.Second * 10,
})
if err != nil {
    log.Fatalf("Failed to create metrics client: %v", err)
}

// 创建配置
config := &balancer.Config{
    Type:           balancer.FastestResponse,
    MetricsClient:  metricsClient,
    UpdateInterval: 30,  // 30秒更新一次性能数据
    RetryTimes:     3,
}

// 创建负载均衡器
lb, err := balancer.Build(config)
// 使用方式同上...
```

## 使用负载均衡器

### 创建负载均衡器

fyerrpc提供了统一的工厂方法来创建各种类型的负载均衡器：

```go
import (
    "github.com/fyerfyer/fyer-rpc/discovery/balancer"
)

// 创建负载均衡器配置
config := &balancer.Config{
    Type:           balancer.RoundRobin, // 可以是Random、RoundRobin或FastestResponse
    MetricsClient:  metricsClient,       // 指标收集客户端
    UpdateInterval: 10,                  // 10秒更新一次
    RetryTimes:     3,                   // 最多重试3次
}

// 使用工厂方法创建负载均衡器
lb, err := balancer.Build(config)
if err != nil {
    log.Fatalf("Failed to create balancer: %v", err)
}
```

### 与服务发现集成

fyerrpc的`discovery.LoadBalancer`类将负载均衡器与服务发现功能结合起来：

```go
import (
    "github.com/fyerfyer/fyer-rpc/discovery"
    "github.com/fyerfyer/fyer-rpc/discovery/balancer"
    "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

// 创建注册中心
registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379"}),
)
if err != nil {
    log.Fatalf("Failed to create registry: %v", err)
}

// 创建服务解析器
resolver, err := discovery.NewResolver(registry, "user-service", "1.0.0")
if err != nil {
    log.Fatalf("Failed to create resolver: %v", err)
}

// 创建指标收集器
metricsClient := metrics.NewNoopMetrics()

// 创建负载均衡器
lb, err := discovery.NewLoadBalancer(
    "user-service",
    "1.0.0",
    resolver,
    metricsClient,
    balancer.RoundRobin,
)
if err != nil {
    log.Fatalf("Failed to create load balancer: %v", err)
}
defer lb.Close()

// 使用负载均衡器选择实例
instance, err := lb.Select(context.Background())
if err != nil {
    log.Fatalf("Failed to select instance: %v", err)
}

// 使用所选实例进行调用
// ...
```

### 使用故障转移功能

结合负载均衡器和故障转移机制，可以提高系统的可靠性：

```go
import (
    "github.com/fyerfyer/fyer-rpc/cluster/failover"
    "github.com/fyerfyer/fyer-rpc/discovery"
)

// 创建故障转移配置
failoverConfig := &failover.Config{
    MaxAttempts:      3,
    RetryInterval:    time.Millisecond * 100,
    RetryableErrors:  []string{"connection refused", "timeout"},
    FailureThreshold: 5,
    SuccessThreshold: 2,
    ResetInterval:    time.Minute,
}

// 创建带故障转移功能的负载均衡器
lb, err := discovery.NewLoadBalancer(
    "user-service",
    "1.0.0",
    resolver,
    metricsClient,
    balancer.RoundRobin,
    discovery.WithFailover(failoverConfig),
)
if err != nil {
    log.Fatalf("Failed to create load balancer: %v", err)
}

// 使用故障转移功能执行操作
err = lb.SelectWithFailover(context.Background(), func(ctx context.Context, instance *naming.Instance) error {
    // 使用所选实例进行调用
    // 如果调用失败，会自动重试其他实例
    return callService(ctx, instance.Address)
})
```

### 更新和反馈

为了使负载均衡器能够根据实际情况做出更好的决策，应该提供调用结果反馈：

```go
// 调用服务并提供反馈
instance, err := lb.Select(context.Background())
if err != nil {
    log.Fatalf("Failed to select instance: %v", err)
}

// 记录开始时间
startTime := time.Now()

// 调用服务
result, callErr := callService(instance.Address)

// 计算调用时间
duration := time.Since(startTime)

// 提供反馈
lb.Feedback(context.Background(), instance, duration.Milliseconds(), callErr)
```

## 自定义负载均衡策略

### 实现自定义负载均衡器

您可以通过实现`balancer.Balancer`接口来创建自定义的负载均衡策略：

```go
// 自定义负载均衡器
type WeightedBalancer struct {
    instances []*naming.Instance
    weights   map[string]int // 实例ID到权重的映射
    mu        sync.RWMutex
}

// Initialize 初始化负载均衡器
func (b *WeightedBalancer) Initialize(instances []*naming.Instance) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.instances = instances
    b.weights = make(map[string]int)
    
    // 初始化权重，可以从实例元数据中读取
    for _, ins := range instances {
        if weight, ok := ins.Metadata["weight"]; ok {
            if w, err := strconv.Atoi(weight); err == nil {
                b.weights[ins.ID] = w
            } else {
                b.weights[ins.ID] = 100 // 默认权重
            }
        } else {
            b.weights[ins.ID] = 100 // 默认权重
        }
    }
    
    return nil
}

// Select 根据权重选择实例
func (b *WeightedBalancer) Select(ctx context.Context) (*naming.Instance, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    if len(b.instances) == 0 {
        return nil, balancer.ErrNoAvailableInstances
    }
    
    // 计算总权重
    totalWeight := 0
    for _, ins := range b.instances {
        if ins.Status == naming.StatusEnabled {
            totalWeight += b.weights[ins.ID]
        }
    }
    
    if totalWeight == 0 {
        return nil, balancer.ErrNoAvailableInstances
    }
    
    // 随机选择一个点
    target := rand.Intn(totalWeight)
    current := 0
    
    // 查找对应的实例
    for _, ins := range b.instances {
        if ins.Status == naming.StatusEnabled {
            current += b.weights[ins.ID]
            if current > target {
                return ins, nil
            }
        }
    }
    
    // 不应该到达这里，但为了安全返回最后一个实例
    for i := len(b.instances) - 1; i >= 0; i-- {
        if b.instances[i].Status == naming.StatusEnabled {
            return b.instances[i], nil
        }
    }
    
    return nil, balancer.ErrNoAvailableInstances
}

// Update 更新实例列表
func (b *WeightedBalancer) Update(instances []*naming.Instance) error {
    return b.Initialize(instances) // 简单实现，直接重新初始化
}

// Feedback 处理调用结果反馈
func (b *WeightedBalancer) Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error) {
    // 可以根据调用结果动态调整权重
}

// Name 返回负载均衡器名称
func (b *WeightedBalancer) Name() string {
    return "weighted"
}
```

### 注册自定义负载均衡器

实现自定义负载均衡器后，需要将其注册到框架中：

```go
// 定义自定义负载均衡器类型
const (
    Weighted balancer.BalancerType = "weighted"
)

// 创建工厂函数
func NewWeightedBalancer(conf *balancer.Config) balancer.Balancer {
    return &WeightedBalancer{}
}

// 注册自定义负载均衡器
func init() {
    balancer.Register(Weighted, NewWeightedBalancer)
}
```

### 使用自定义负载均衡器

注册后，可以像使用内置负载均衡器一样使用自定义负载均衡器：

```go
// 创建配置
config := &balancer.Config{
    Type: Weighted,
    // 其他配置...
}

// 创建负载均衡器
lb, err := balancer.Build(config)
if err != nil {
    log.Fatalf("Failed to create balancer: %v", err)
}

// 使用方式与内置负载均衡器相同
```

## 指标收集与监控

负载均衡器的性能依赖于准确的服务实例指标，fyerrpc提供了指标收集功能：

### Metrics接口

```go
type Metrics interface {
    // RecordResponse 记录响应时间
    RecordResponse(ctx context.Context, metric *ResponseMetric) error

    // GetLatency 获取指定服务实例的平均响应时间
    GetLatency(ctx context.Context, service, instance string) (time.Duration, error)

    // GetServiceLatency 获取服务所有实例的平均响应时间
    GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error)

    // RecordFailover 记录故障转移事件
    RecordFailover(ctx context.Context, service, fromInstance, toInstance string) error

    // RecordCircuitBreak 记录熔断事件
    RecordCircuitBreak(ctx context.Context, service, instance string, state string) error

    // RecordRetry 记录重试事件
    RecordRetry(ctx context.Context, service, instance string, attempt int) error

    // GetFailoverRate 获取故障转移率
    GetFailoverRate(ctx context.Context, service string) (float64, error)

    // Close 关闭指标收集器
    Close() error
}
```

### Prometheus集成

fyerrpc提供了与Prometheus集成的指标收集实现：

```go
import (
    "github.com/fyerfyer/fyer-rpc/discovery/metrics"
)

// 创建Prometheus指标收集器
metricsClient, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
    PushGatewayURL: "http://localhost:9091",
    QueryURL:       "http://localhost:9090",
    JobName:        "fyerrpc",
    PushInterval:   time.Second * 10,
})
if err != nil {
    log.Fatalf("Failed to create metrics client: %v", err)
}
defer metricsClient.Close()

// 使用指标收集器创建负载均衡器
config := &balancer.Config{
    Type:           balancer.FastestResponse,
    MetricsClient:  metricsClient,
    UpdateInterval: 30,
    RetryTimes:     3,
}

// 创建负载均衡器
lb, err := balancer.Build(config)
// 使用负载均衡器...
```

