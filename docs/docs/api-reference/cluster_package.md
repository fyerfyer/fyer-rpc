# `cluster` Package

cluster包提供了集群管理功能，包括分组路由、故障转移和服务选择。

### failover 子包

`failover`子包实现了故障转移功能：

```go
// DefaultFailoverHandler 默认故障转移处理器
type DefaultFailoverHandler struct {
    detector        Detector
    circuitBreaker  CircuitBreaker
    retryPolicy     RetryPolicy
    recovery        RecoveryStrategy
    monitor         InstanceMonitor
}

// NewFailoverHandler 创建故障转移处理器
func NewFailoverHandler(config *Config) (*DefaultFailoverHandler, error)

// Execute 执行带故障转移的调用
func (h *DefaultFailoverHandler) Execute(ctx context.Context, instances []*naming.Instance, operation func(context.Context, *naming.Instance) error) (*FailoverResult, error)
```

故障检测：

```go
type Detector interface {
    // Detect 检测实例是否健康
    Detect(ctx context.Context, instance *naming.Instance) (Status, error)

    // MarkFailed 标记实例为失败状态
    MarkFailed(ctx context.Context, instance *naming.Instance) error

    // MarkSuccess 标记实例为成功状态
    MarkSuccess(ctx context.Context, instance *naming.Instance) error
}
```

熔断器：

```go
type CircuitBreaker interface {
    // Allow 判断请求是否允许通过
    Allow(ctx context.Context, instance *naming.Instance) (bool, error)

    // MarkSuccess 标记成功调用
    MarkSuccess(ctx context.Context, instance *naming.Instance) error

    // MarkFailure 标记失败调用
    MarkFailure(ctx context.Context, instance *naming.Instance, err error) error

    // GetState 获取熔断器状态
    GetState(instance *naming.Instance) (State, error)

    // Reset 重置熔断器状态
    Reset(instance *naming.Instance) error
}
```

重试策略：

```go
type RetryPolicy interface {
    // ShouldRetry 决定是否需要重试
    ShouldRetry(ctx context.Context, attempt int, err error) bool

    // NextBackoff 计算下一次重试的等待时间
    NextBackoff(attempt int) time.Duration

    // MaxAttempts 返回最大重试次数
    MaxAttempts() int
}
```

### selector 子包

`selector`子包实现了服务选择功能：

```go
// Selector 定义了选择器接口
type Selector interface {
    // Select 从服务实例列表中选择符合条件的实例
    Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
    
    // Name 返回选择器名称
    Name() string
}
```

选择器链：

```go
// Chain 选择器链，支持多个选择器串联
type Chain struct {
    selectors []Selector
    name      string
}

// NewChain 创建新的选择器链
func NewChain(name string, selectors ...Selector) *Chain

// Select 按顺序执行选择器链中的所有选择器
func (c *Chain) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
```

### group 子包

`group`子包实现了服务分组功能：

```go
// Group 表示一个服务分组的接口
type Group interface {
    // Name 返回分组名称
    Name() string

    // Match 检查实例是否匹配该分组
    Match(instance *naming.Instance) bool

    // Select 从实例列表中选择匹配该分组的实例
    Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

// NewGroup 创建新的分组
func NewGroup(name string, opts ...Option) (*DefaultGroup, error)
```

分组路由：

```go
// GroupRouter 分组路由接口
type GroupRouter interface {
    // Route 根据上下文路由到指定分组
    Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

// NewRouter 创建新的路由器
func NewRouter(manager GroupManager) *DefaultRouter

// NewRouterChain 创建新的路由链
func NewRouterChain(routers ...GroupRouter) *RouterChain
```